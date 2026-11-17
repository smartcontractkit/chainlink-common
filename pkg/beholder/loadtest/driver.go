//go:build loadtest

package loadtest

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand/v2"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/batch"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

const (
	timelineBucket = 250 * time.Millisecond
	sampleEvery    = 250 * time.Millisecond
	settleTimeout  = 10 * time.Second
	// maxGeneratorDebt bounds how far a generator goroutine may fall behind its schedule before
	// the missed events are skipped (and counted) instead of being emitted as one huge burst.
	maxGeneratorDebt = time.Second
)

// PayloadSpec describes event payload sizes in bytes: one entry is a fixed size, several are a
// weighted distribution (Weights parallel to Sizes; nil means equal weights).
type PayloadSpec struct {
	Sizes   []int
	Weights []int
}

// Fixed is a PayloadSpec with one size.
func Fixed(n int) PayloadSpec { return PayloadSpec{Sizes: []int{n}} }

func (p PayloadSpec) pick(r *rand.Rand) int {
	if len(p.Sizes) == 1 {
		return p.Sizes[0]
	}
	total := 0
	for i := range p.Sizes {
		total += p.weight(i)
	}
	x := r.IntN(total)
	for i, s := range p.Sizes {
		x -= p.weight(i)
		if x < 0 {
			return s
		}
	}
	return p.Sizes[len(p.Sizes)-1]
}

func (p PayloadSpec) weight(i int) int {
	if i < len(p.Weights) {
		return p.Weights[i]
	}
	return 1
}

// Mean is the expected payload size in bytes.
func (p PayloadSpec) Mean() float64 {
	var sum, tot float64
	for i, s := range p.Sizes {
		sum += float64(s) * float64(p.weight(i))
		tot += float64(p.weight(i))
	}
	if tot == 0 {
		return 0
	}
	return sum / tot
}

// DomainLoad is the traffic offered to one domain emitter.
type DomainLoad struct {
	Domain  beholder.Domain
	Rate    float64 // target events/s (before LOADTEST_RATE_SCALE)
	Payload PayloadSpec
	// OversizedEvery, when > 0, makes every Nth event OversizedSize bytes (meant to exceed
	// ChipIngressMaxGRPCRequestSize).
	OversizedEvery int
	OversizedSize  int
}

// Scenario is one named, table-driven load scenario.
type Scenario struct {
	Name        string
	Description string
	Loads       []DomainLoad
	Duration    time.Duration // load phase; LOADTEST_DURATION overrides
	Server      ServerProfile
	// Tune adjusts the beholder.Config after the production defaults are applied.
	Tune func(*beholder.Config)
	// Check returns pass-bar violations (plan section 5); enforced only with LOADTEST_STRICT=1.
	Check func(*Result, RunOptions) []string
	// Invariants returns scenario-specific correctness violations that are always enforced.
	Invariants func(*Result) []string
	// Notes returns lines always added to the result (e.g. plan bars known to be unmet); never failures.
	Notes func(*Result) []string
	// Extended scenarios are skipped unless LOADTEST_EXTENDED=1 or LOADTEST_SCENARIO matches them.
	Extended bool
}

// RunOptions are the run-wide overrides (see OptionsFromEnv).
type RunOptions struct {
	Duration    time.Duration // 0 = scenario default
	RateScale   float64       // 0 = 1
	MemLimitMiB float64       // 0 = no RSS bar
}

// ---------- results ----------

// DomainResult is the accounting for one domain.
type DomainResult struct {
	Domain       string  `json:"domain"`
	TargetRate   float64 `json:"target_rate"`
	AchievedRate float64 `json:"achieved_rate"` // emitted / load duration
	Emitted      int64   `json:"emitted"`
	// SyncErrors are EmitEvent calls that returned an error (callback not invoked), by code.
	SyncErrors map[string]int64 `json:"sync_errors"`
	Accepted   int64            `json:"accepted"` // callback(nil)
	// CallbackErrors are callbacks with an error, by batch.ErrorCodeFor. Buffer-full drops
	// appear here as "message buffer is full".
	CallbackErrors   map[string]int64  `json:"callback_errors"`
	Pending          int64             `json:"pending"` // queued, callback not yet fired at end
	Dropped          int64             `json:"dropped"` // sync + callback errors
	BuildErrors      int64             `json:"build_errors"`
	GeneratorSkipped int64             `json:"generator_skipped"` // events the generator could not keep up with
	CallbackAnomaly  int64             `json:"callback_anomalies"`
	P50Ms            float64           `json:"p50_ms"` // emit -> callback(nil)
	P99Ms            float64           `json:"p99_ms"`
	PeakPending      int64             `json:"peak_pending_events"`
	PeakPendingMiB   float64           `json:"peak_pending_payload_mib"`
	PredictedMaxRate float64           `json:"predicted_max_rate"` // MaxConcurrentSends*MaxBatchSize/RTT; 0 if no latency
	OversizedEmitted int64             `json:"oversized_emitted"`
	OversizedOK      int64             `json:"oversized_accepted"`
	OversizedErrors  map[string]int64  `json:"oversized_errors"`
	OversizedSample  string            `json:"oversized_error_sample,omitempty"`
	Server           ServerDomainStats `json:"server"`
}

// Bucket is one timeline bucket (time of emit for Emitted, of callback for the others).
type Bucket struct {
	TSec     float64 `json:"t_sec"`
	Emitted  int64   `json:"emitted"`
	Accepted int64   `json:"accepted"`
	Errors   int64   `json:"errors"`
}

// Result is one scenario's outcome.
type Result struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	LoadDurationSec  float64        `json:"load_duration_sec"`
	CloseDurationSec float64        `json:"close_duration_sec"`
	Config           ConfigSnapshot `json:"config"`

	Domains []*DomainResult `json:"domains"`

	// Totals across domains.
	Emitted, Accepted, Dropped, Pending int64
	DroppedByCode                       map[string]int64 `json:"dropped_by_code"`
	ServerAccepted                      int64            `json:"server_accepted"`
	ServerReceived                      int64            `json:"server_received"`
	AchievedRate                        float64          `json:"achieved_rate"`
	P50Ms                               float64          `json:"p50_ms"`
	P99Ms                               float64          `json:"p99_ms"`

	// Resources, sampled every 250ms during load and close. The fake server and the generator run
	// in this process, so these are upper bounds for the beholder client alone.
	BaselineHeapMiB float64 `json:"baseline_heap_inuse_mib"`
	PeakHeapMiB     float64 `json:"peak_heap_inuse_mib"`
	PeakRSSProxyMiB float64 `json:"peak_rss_proxy_mib"` // MemStats Sys - HeapReleased
	PeakGoroutines  int     `json:"peak_goroutines"`
	CPUCores        float64 `json:"cpu_cores"` // process CPU seconds / wall seconds

	// Timing of errors, seconds since load start (-1 = none).
	FirstErrorAtSec float64 `json:"first_error_at_sec"`
	LastErrorAtSec  float64 `json:"last_error_at_sec"`
	// Outage scenarios: seconds from the end of the last outage window until the first accepted
	// event, and until the last error callback (-1 = n/a).
	RecoveryFirstAcceptSec float64 `json:"recovery_first_accept_sec"`
	RecoveryLastErrorSec   float64 `json:"recovery_last_error_sec"`

	// Close phase (restart scenarios, and every scenario's final Close).
	PendingAtStop        int64 `json:"pending_at_stop"` // queued or in flight when the load stopped
	DeliveredDuringClose int64 `json:"delivered_during_close"`
	FailedDuringClose    int64 `json:"failed_during_close"`

	Timeline []Bucket `json:"timeline"`

	InvariantViolations []string `json:"invariant_violations,omitempty"`
	StrictViolations    []string `json:"strict_violations,omitempty"`
	Notes               []string `json:"notes,omitempty"`
}

// ConfigSnapshot records the batching settings a scenario ran with.
type ConfigSnapshot struct {
	BufferSize         uint    `json:"buffer_size"`
	MaxBatchSize       uint    `json:"max_batch_size"`
	SendIntervalMs     float64 `json:"send_interval_ms"`
	SendTimeoutMs      float64 `json:"send_timeout_ms"`
	DrainTimeoutMs     float64 `json:"drain_timeout_ms"`
	MaxConcurrentSends int     `json:"max_concurrent_sends"`
	MaxGRPCRequestSize int     `json:"max_grpc_request_size"`
	ServerLatencyMs    float64 `json:"server_latency_ms"`
	ServerJitterMs     float64 `json:"server_jitter_ms"`
	ServerErrorRate    float64 `json:"server_error_rate"`
	RateScale          float64 `json:"rate_scale"`
}

// ---------- helpers ----------

type counterMap struct{ m sync.Map }

func (c *counterMap) add(k string) {
	v, ok := c.m.Load(k)
	if !ok {
		v, _ = c.m.LoadOrStore(k, new(atomic.Int64))
	}
	v.(*atomic.Int64).Add(1)
}

func (c *counterMap) snapshot() map[string]int64 {
	out := map[string]int64{}
	c.m.Range(func(k, v any) bool {
		out[k.(string)] = v.(*atomic.Int64).Load()
		return true
	})
	return out
}

func sumMap(m map[string]int64) (n int64) {
	for _, v := range m {
		n += v
	}
	return n
}

// latHist is a lock-free log-scale histogram (4 buckets per octave, microsecond resolution).
type latHist struct {
	b [160]atomic.Int64
	n atomic.Int64
}

func (h *latHist) add(d time.Duration) {
	us := float64(d.Microseconds() + 1)
	i := int(math.Log2(us) * 4)
	if i >= len(h.b) {
		i = len(h.b) - 1
	}
	h.b[i].Add(1)
	h.n.Add(1)
}

func mergeQuantile(hs []*latHist, q float64) time.Duration {
	var total int64
	for _, h := range hs {
		total += h.n.Load()
	}
	if total == 0 {
		return 0
	}
	target := int64(math.Ceil(q * float64(total)))
	var cum int64
	for i := range hs[0].b {
		for _, h := range hs {
			cum += h.b[i].Load()
		}
		if cum >= target {
			// upper edge of the bucket
			return time.Duration(math.Pow(2, float64(i+1)/4)) * time.Microsecond
		}
	}
	return 0
}

// domainRun is the live state of one domain's load.
type domainRun struct {
	load DomainLoad
	rate float64 // after scale

	emitted, accepted, syncErr, cbErr, inflight atomic.Int64
	buildErr, skipped, anomalies                atomic.Int64
	seq                                         atomic.Uint64
	syncCodes, cbCodes, overCodes               counterMap
	overEmitted, overOK                         atomic.Int64
	overSample                                  atomic.Pointer[string]
	hist                                        latHist
	peakPending                                 atomic.Int64
}

type runState struct {
	t0       time.Time
	tl       []tlBucket
	firstErr atomic.Int64 // ns since t0 + 1; 0 = none
	lastErr  atomic.Int64
	// outageEnd is the end of the last outage window as ns since t0 (0 = none).
	outageEnd        int64
	firstAcceptAfter atomic.Int64 // ns since t0 + 1
	domains          []*domainRun
}

type tlBucket struct{ emitted, accepted, errs atomic.Int64 }

func (rs *runState) bucket(t time.Time) *tlBucket {
	i := int(t.Sub(rs.t0) / timelineBucket)
	if i < 0 {
		i = 0
	}
	if i >= len(rs.tl) {
		i = len(rs.tl) - 1
	}
	return &rs.tl[i]
}

func storeFirst(a *atomic.Int64, v int64) { a.CompareAndSwap(0, v) }

func (rs *runState) noteError(now time.Time) {
	ns := int64(now.Sub(rs.t0)) + 1
	storeFirst(&rs.firstErr, ns)
	for {
		cur := rs.lastErr.Load()
		if ns <= cur || rs.lastErr.CompareAndSwap(cur, ns) {
			return
		}
	}
}

type noopLogExporter struct{}

func (noopLogExporter) Export(context.Context, []sdklog.Record) error { return nil }
func (noopLogExporter) Shutdown(context.Context) error                { return nil }
func (noopLogExporter) ForceFlush(context.Context) error              { return nil }

func cpuSeconds() float64 {
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		return 0
	}
	return float64(ru.Utime.Sec) + float64(ru.Utime.Usec)/1e6 + float64(ru.Stime.Sec) + float64(ru.Stime.Usec)/1e6
}

func mib(b uint64) float64 { return float64(b) / (1 << 20) }

// ---------- run ----------

// buildConfig returns the production batching defaults pointed at addr, with domain emitters on.
func buildConfig(addr string, sc Scenario) beholder.Config {
	cfg := beholder.DefaultConfig()
	cfg.OtelExporterGRPCEndpoint = addr // OTLP metrics/traces are swallowed by the fake server; logs use a noop exporter
	cfg.LogStreamingEnabled = false
	cfg.EmitterBatchProcessor = false
	cfg.MetricReaderInterval = time.Hour // keep the OTLP metric exporter quiet
	cfg.TraceSampleRatio = 0
	cfg.InsecureConnection = true
	cfg.ChipIngressEmitterEnabled = true
	cfg.ChipIngressEmitterGRPCEndpoint = addr
	cfg.ChipIngressInsecureConnection = true
	cfg.ChipIngressBatchEmitterEnabled = true
	cfg.ChipDomainEmittersEnabled = true
	cfg.ChipIngressLogger = logger.Nop()
	if sc.Tune != nil {
		sc.Tune(&cfg)
	}
	return cfg
}

// RunScenario runs sc against a fresh fake server and a fresh beholder client.
func RunScenario(ctx context.Context, sc Scenario, opts RunOptions) (*Result, error) {
	scale := opts.RateScale
	if scale <= 0 {
		scale = 1
	}
	dur := sc.Duration
	if opts.Duration > 0 {
		dur = opts.Duration
	}

	server, err := NewFakeServer(sc.Server)
	if err != nil {
		return nil, err
	}
	defer server.Close()

	cfg := buildConfig(server.Addr(), sc)
	client, err := beholder.NewGRPCClient(cfg, func(...otlploggrpc.Option) (sdklog.Exporter, error) {
		return noopLogExporter{}, nil
	})
	if err != nil {
		return nil, fmt.Errorf("beholder client: %w", err)
	}
	if err := client.Start(ctx); err != nil {
		return nil, fmt.Errorf("start beholder client: %w", err)
	}
	closed := false
	defer func() {
		if !closed {
			_ = client.Close()
		}
	}()

	rs := &runState{tl: make([]tlBucket, int((dur+3*time.Minute)/timelineBucket)+1)}
	var emitters []beholder.DomainEmitter
	for _, l := range sc.Loads {
		em, err := client.ForDomain(l.Domain)
		if err != nil {
			return nil, err
		}
		emitters = append(emitters, em)
		rs.domains = append(rs.domains, &domainRun{load: l, rate: l.Rate * scale})
	}
	for _, w := range sc.Server.Outages {
		if int64(w.End) > rs.outageEnd {
			rs.outageEnd = int64(w.End)
		}
	}

	var base runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&base)

	res := &Result{Name: sc.Name, Description: sc.Description, FirstErrorAtSec: -1, LastErrorAtSec: -1,
		RecoveryFirstAcceptSec: -1, RecoveryLastErrorSec: -1}
	res.BaselineHeapMiB = mib(base.HeapInuse)

	// Start the clock for load, outage windows and timeline together.
	rs.t0 = time.Now()
	server.MarkStart()
	cpu0 := cpuSeconds()

	// Resource sampler.
	var peakHeap, peakRSS uint64
	var peakG int
	sampCtx, stopSamp := context.WithCancel(context.Background())
	var sampWG sync.WaitGroup
	sampWG.Add(1)
	go func() {
		defer sampWG.Done()
		t := time.NewTicker(sampleEvery)
		defer t.Stop()
		sample := func() {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			peakHeap = max(peakHeap, m.HeapInuse)
			peakRSS = max(peakRSS, m.Sys-m.HeapReleased)
			peakG = max(peakG, runtime.NumGoroutine())
			for _, d := range rs.domains {
				p := d.inflight.Load()
				if p > d.peakPending.Load() {
					d.peakPending.Store(p)
				}
			}
		}
		sample()
		for {
			select {
			case <-sampCtx.Done():
				sample()
				return
			case <-t.C:
				sample()
			}
		}
	}()

	// Generators.
	genCtx, stopGen := context.WithTimeout(ctx, dur)
	var genWG sync.WaitGroup
	for i, d := range rs.domains {
		n := int(math.Ceil(d.rate / 2000))
		n = min(max(n, 1), 16)
		for g := 0; g < n; g++ {
			genWG.Add(1)
			go func(d *domainRun, em beholder.DomainEmitter, share float64, seed uint64) {
				defer genWG.Done()
				generate(genCtx, rs, d, em, share, seed)
			}(d, emitters[i], d.rate/float64(n), uint64(i*100+g+1))
		}
	}
	genWG.Wait()
	stopGen()
	loadDur := time.Since(rs.t0)

	// Close the client: drains per ChipIngressDrainTimeout.
	atStop := snapshotTotals(rs)
	closeStart := time.Now()
	_ = client.Close()
	closed = true
	closeDur := time.Since(closeStart)

	// Callbacks may fire after Close returns; give them time, then report what is still pending.
	deadline := time.Now().Add(settleTimeout)
	for time.Now().Before(deadline) && totalInflight(rs) > 0 {
		time.Sleep(50 * time.Millisecond)
	}
	// Fate of everything that was queued or in flight when the load stopped, after Close and settle.
	afterSettle := snapshotTotals(rs)
	cpuUsed := cpuSeconds() - cpu0
	wall := time.Since(rs.t0)
	stopSamp()
	sampWG.Wait()

	// ---- assemble ----
	res.LoadDurationSec = loadDur.Seconds()
	res.CloseDurationSec = closeDur.Seconds()
	res.Config = ConfigSnapshot{
		BufferSize: cfg.ChipIngressBufferSize, MaxBatchSize: cfg.ChipIngressMaxBatchSize,
		SendIntervalMs: ms(cfg.ChipIngressSendInterval), SendTimeoutMs: ms(cfg.ChipIngressSendTimeout),
		DrainTimeoutMs: ms(cfg.ChipIngressDrainTimeout), MaxConcurrentSends: cfg.ChipIngressMaxConcurrentSends,
		MaxGRPCRequestSize: cfg.ChipIngressMaxGRPCRequestSize,
		ServerLatencyMs:    ms(sc.Server.Latency), ServerJitterMs: ms(sc.Server.Jitter),
		ServerErrorRate: sc.Server.ErrorRate, RateScale: scale,
	}
	res.PeakHeapMiB = mib(peakHeap)
	res.PeakRSSProxyMiB = mib(peakRSS)
	res.PeakGoroutines = peakG
	res.CPUCores = cpuUsed / wall.Seconds()
	res.PendingAtStop = atStop.inflight
	res.DeliveredDuringClose = afterSettle.accepted - atStop.accepted
	res.FailedDuringClose = afterSettle.errs - atStop.errs

	srv := server.Snapshot()
	var hists []*latHist
	rtt := sc.Server.Latency + sc.Server.Jitter/2
	for _, d := range rs.domains {
		dr := &DomainResult{
			Domain:           string(d.load.Domain),
			TargetRate:       d.rate,
			Emitted:          d.emitted.Load(),
			SyncErrors:       d.syncCodes.snapshot(),
			Accepted:         d.accepted.Load(),
			CallbackErrors:   d.cbCodes.snapshot(),
			Pending:          d.inflight.Load(),
			BuildErrors:      d.buildErr.Load(),
			GeneratorSkipped: d.skipped.Load(),
			CallbackAnomaly:  d.anomalies.Load(),
			PeakPending:      d.peakPending.Load(),
			OversizedEmitted: d.overEmitted.Load(),
			OversizedOK:      d.overOK.Load(),
			OversizedErrors:  d.overCodes.snapshot(),
			Server:           srv[string(d.load.Domain)],
		}
		if p := d.overSample.Load(); p != nil {
			dr.OversizedSample = *p
		}
		dr.Dropped = sumMap(dr.SyncErrors) + sumMap(dr.CallbackErrors)
		dr.AchievedRate = float64(dr.Emitted) / loadDur.Seconds()
		dr.PeakPendingMiB = float64(dr.PeakPending) * d.load.Payload.Mean() / (1 << 20)
		dr.P50Ms = ms(mergeQuantile([]*latHist{&d.hist}, 0.50))
		dr.P99Ms = ms(mergeQuantile([]*latHist{&d.hist}, 0.99))
		if rtt > 0 {
			dr.PredictedMaxRate = float64(cfg.ChipIngressMaxConcurrentSends) * float64(cfg.ChipIngressMaxBatchSize) / rtt.Seconds()
			if rtt > cfg.ChipIngressSendTimeout {
				dr.PredictedMaxRate = 0
			}
		}
		hists = append(hists, &d.hist)
		res.Domains = append(res.Domains, dr)
		res.Emitted += dr.Emitted
		res.Accepted += dr.Accepted
		res.Dropped += dr.Dropped
		res.Pending += dr.Pending
		for k, v := range dr.SyncErrors {
			addTo(&res.DroppedByCode, k, v)
		}
		for k, v := range dr.CallbackErrors {
			addTo(&res.DroppedByCode, k, v)
		}
	}
	for _, s := range srv {
		res.ServerAccepted += s.EventsAccepted
		res.ServerReceived += s.EventsReceived
	}
	res.AchievedRate = float64(res.Emitted) / loadDur.Seconds()
	res.P50Ms = ms(mergeQuantile(hists, 0.50))
	res.P99Ms = ms(mergeQuantile(hists, 0.99))

	if v := rs.firstErr.Load(); v > 0 {
		res.FirstErrorAtSec = time.Duration(v - 1).Seconds()
	}
	if v := rs.lastErr.Load(); v > 0 {
		res.LastErrorAtSec = time.Duration(v - 1).Seconds()
	}
	if rs.outageEnd > 0 {
		end := time.Duration(rs.outageEnd).Seconds()
		if v := rs.firstAcceptAfter.Load(); v > 0 {
			res.RecoveryFirstAcceptSec = time.Duration(v-1).Seconds() - end
		}
		if res.LastErrorAtSec >= 0 {
			res.RecoveryLastErrorSec = max(0, res.LastErrorAtSec-end)
		}
	}
	last := 0
	for i := range rs.tl {
		if rs.tl[i].emitted.Load()+rs.tl[i].accepted.Load()+rs.tl[i].errs.Load() > 0 {
			last = i
		}
	}
	for i := 0; i <= last; i++ {
		res.Timeline = append(res.Timeline, Bucket{
			TSec: float64(i) * timelineBucket.Seconds(), Emitted: rs.tl[i].emitted.Load(),
			Accepted: rs.tl[i].accepted.Load(), Errors: rs.tl[i].errs.Load(),
		})
	}

	res.InvariantViolations = accountingViolations(res)
	if sc.Invariants != nil {
		res.InvariantViolations = append(res.InvariantViolations, sc.Invariants(res)...)
	}
	if sc.Notes != nil {
		res.Notes = append(res.Notes, sc.Notes(res)...)
	}
	if sc.Check != nil {
		res.StrictViolations = sc.Check(res, opts)
	}
	if res.AchievedRate < 0.95*sumRates(rs) {
		res.Notes = append(res.Notes, fmt.Sprintf("generator achieved %.0f/s of %.0f/s target (CPU-bound load generator?)", res.AchievedRate, sumRates(rs)))
	}
	return res, nil
}

func sumRates(rs *runState) (s float64) {
	for _, d := range rs.domains {
		s += d.rate
	}
	return s
}

func addTo(m *map[string]int64, k string, v int64) {
	if *m == nil {
		*m = map[string]int64{}
	}
	(*m)[k] += v
}

func ms(d time.Duration) float64 { return float64(d) / float64(time.Millisecond) }

type totals struct{ accepted, errs, inflight int64 }

func snapshotTotals(rs *runState) (t totals) {
	for _, d := range rs.domains {
		t.accepted += d.accepted.Load()
		t.errs += d.cbErr.Load()
		t.inflight += d.inflight.Load()
	}
	return t
}

func totalInflight(rs *runState) int64 { return snapshotTotals(rs).inflight }

// generate paces one goroutine's share of a domain's rate. It wakes every millisecond and emits
// however many events the schedule says are due, so pacing does not depend on timer resolution.
func generate(ctx context.Context, rs *runState, d *domainRun, em beholder.DomainEmitter, rate float64, seed uint64) {
	r := rand.New(rand.NewPCG(seed, seed*7919+1))
	source := string(d.load.Domain) + ".telemetry.loadtest"
	start := time.Now()
	var sent int64
	maxDebt := int64(rate * maxGeneratorDebt.Seconds())
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
		due := int64(time.Since(start).Seconds()*rate) - sent
		if due > maxDebt && maxDebt > 0 {
			d.skipped.Add(due - maxDebt)
			sent += due - maxDebt
			due = maxDebt
		}
		for i := int64(0); i < due; i++ {
			emitOne(rs, d, em, source, r)
		}
		sent += due
	}
}

func emitOne(rs *runState, d *domainRun, em beholder.DomainEmitter, source string, r *rand.Rand) {
	seq := d.seq.Add(1)
	oversized := d.load.OversizedEvery > 0 && seq%uint64(d.load.OversizedEvery) == 0
	size := d.load.Payload.pick(r)
	if oversized {
		size = d.load.OversizedSize
	}
	// A fresh allocation per event, as a real caller has: buffered memory must not alias one slice.
	payload := make([]byte, max(size, 8))[:size]
	if size >= 8 {
		binary.BigEndian.PutUint64(payload, seq)
	}
	ev, err := chipingress.NewEvent(source, "loadtest.event", payload, nil)
	if err != nil {
		d.buildErr.Add(1)
		return
	}

	d.emitted.Add(1)
	if oversized {
		d.overEmitted.Add(1)
	}
	rs.bucket(time.Now()).emitted.Add(1)
	d.inflight.Add(1) // assume queued; undone on a synchronous error
	began := time.Now()
	var fired atomic.Int32
	cb := func(err error) {
		now := time.Now()
		if fired.Swap(1) != 0 { // double fire, or fire after EmitEvent returned an error
			d.anomalies.Add(1)
			return
		}
		b := rs.bucket(now)
		if err == nil {
			d.accepted.Add(1)
			d.hist.add(now.Sub(began))
			b.accepted.Add(1)
			if oversized {
				d.overOK.Add(1)
			}
			if rs.outageEnd > 0 && int64(now.Sub(rs.t0)) >= rs.outageEnd {
				storeFirst(&rs.firstAcceptAfter, int64(now.Sub(rs.t0))+1)
			}
		} else {
			code := batch.ErrorCodeFor(err)
			d.cbCodes.add(code)
			d.cbErr.Add(1)
			b.errs.Add(1)
			rs.noteError(now)
			if oversized {
				d.overCodes.add(code)
				s := err.Error()
				d.overSample.CompareAndSwap(nil, &s)
			}
		}
		d.inflight.Add(-1)
	}
	if err := em.EmitEvent(context.Background(), ev, cb); err != nil {
		if fired.Swap(2) == 1 { // callback ran although an error was returned
			d.anomalies.Add(1)
		}
		d.inflight.Add(-1)
		d.syncErr.Add(1)
		code := batch.ErrorCodeFor(err)
		d.syncCodes.add(code)
		rs.noteError(time.Now())
		rs.bucket(time.Now()).errs.Add(1)
	}
}

// accountingViolations checks the invariants that must hold for every scenario.
func accountingViolations(r *Result) []string {
	var v []string
	for _, d := range r.Domains {
		syncN, cbN := sumMap(d.SyncErrors), sumMap(d.CallbackErrors)
		if d.Emitted != syncN+d.Accepted+cbN+d.Pending {
			v = append(v, fmt.Sprintf("%s: emitted %d != sync %d + accepted %d + callback errors %d + pending %d",
				d.Domain, d.Emitted, syncN, d.Accepted, cbN, d.Pending))
		}
		if d.CallbackAnomaly != 0 {
			v = append(v, fmt.Sprintf("%s: %d callback anomalies (double fire, or fire after Emit returned an error)", d.Domain, d.CallbackAnomaly))
		}
		queued := d.Emitted - syncN
		// No-ack guarantee: a nil callback means the server returned a success result, which it
		// only does after counting the batch, so server accepted >= client accepted. The server
		// can accept more (response lost to a client timeout or cancel), but never more than the
		// events that were queued.
		if d.Server.EventsAccepted < d.Accepted {
			v = append(v, fmt.Sprintf("%s: server accepted %d < client accepted %d", d.Domain, d.Server.EventsAccepted, d.Accepted))
		}
		if d.Server.EventsAccepted > queued {
			v = append(v, fmt.Sprintf("%s: server accepted %d > events queued %d", d.Domain, d.Server.EventsAccepted, queued))
		}
		if d.Server.SourceMismatch != 0 {
			v = append(v, fmt.Sprintf("%s: %d events reached the server on a route for another domain", d.Domain, d.Server.SourceMismatch))
		}
		if d.Server.AckBatches != 0 {
			v = append(v, fmt.Sprintf("%s: %d batches used PublishBatch; domain emitters must use PublishBatchNoAck", d.Domain, d.Server.AckBatches))
		}
	}
	for _, dr := range r.Domains {
		if dr.Pending != 0 {
			v = append(v, fmt.Sprintf("%s: %d events still pending (callback never fired) after Close + %s settle", dr.Domain, dr.Pending, settleTimeout))
		}
	}
	return v
}

// Summary is a one-line human summary.
func (r *Result) Summary() string {
	codes := make([]string, 0, len(r.DroppedByCode))
	for k, v := range r.DroppedByCode {
		codes = append(codes, fmt.Sprintf("%s=%d", k, v))
	}
	sort.Strings(codes)
	dropped := "0"
	if len(codes) > 0 {
		dropped = fmt.Sprintf("%d [%s]", r.Dropped, strings.Join(codes, ", "))
	}
	return fmt.Sprintf("emitted=%d accepted=%d dropped=%s pending=%d server_accepted=%d rate=%.0f/s p50=%.0fms p99=%.0fms heap=%.0fMiB rss~%.0fMiB goroutines=%d cpu=%.1f",
		r.Emitted, r.Accepted, dropped, r.Pending, r.ServerAccepted, r.AchievedRate, r.P50Ms, r.P99Ms,
		r.PeakHeapMiB, r.PeakRSSProxyMiB, r.PeakGoroutines, r.CPUCores)
}

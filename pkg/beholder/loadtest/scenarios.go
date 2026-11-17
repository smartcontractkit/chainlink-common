//go:build loadtest

package loadtest

import (
	"fmt"
	"strings"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

const (
	// peakRate is the top of the OTI range for one Data Streams node (plan section 1).
	peakRate = 16000
	// feedsRate is the median Data Feeds node.
	feedsRate = 40
	mib10     = 10 << 20
)

// Synthetic payload distributions, until real per-domain distributions are collected.
var (
	typicalPayload = PayloadSpec{Sizes: []int{256, 1024, 4096}, Weights: []int{60, 35, 5}}            // mean ~0.7 KiB
	largePayload   = PayloadSpec{Sizes: []int{512, 2048, 8192, 32768}, Weights: []int{50, 35, 13, 2}} // mean ~2.7 KiB
)

func streamsAndFeeds(streams float64) []DomainLoad {
	return []DomainLoad{
		{Domain: beholder.DomainDataStreams, Rate: streams, Payload: typicalPayload},
		{Domain: beholder.DomainDataFeeds, Rate: feedsRate * streams / peakRate, Payload: typicalPayload},
	}
}

func rtt(lat, jitter time.Duration) ServerProfile { return ServerProfile{Latency: lat, Jitter: jitter} }

// zeroDrops is the plan section 5 node pass bar: nothing dropped, nothing left pending, and the
// generator actually offered the load.
func zeroDrops(r *Result, _ RunOptions) []string {
	var v []string
	if r.Dropped != 0 {
		v = append(v, fmt.Sprintf("%d events dropped, want 0 (%v)", r.Dropped, r.DroppedByCode))
	}
	if r.Pending != 0 {
		v = append(v, fmt.Sprintf("%d events pending at end, want 0", r.Pending))
	}
	for _, d := range r.Domains {
		if d.AchievedRate < 0.95*d.TargetRate {
			v = append(v, fmt.Sprintf("%s: generator achieved %.0f/s of %.0f/s target; result is not valid for the pass bar", d.Domain, d.AchievedRate, d.TargetRate))
		}
	}
	return v
}

func memoryBar(r *Result, o RunOptions) []string {
	if o.MemLimitMiB <= 0 {
		return nil
	}
	if r.PeakRSSProxyMiB > o.MemLimitMiB {
		return []string{fmt.Sprintf("peak RSS proxy %.0f MiB > LOADTEST_MEM_LIMIT_MIB %.0f", r.PeakRSSProxyMiB, o.MemLimitMiB)}
	}
	return nil
}

// oversizedInvariants always applies: oversized events must fail locally, alone.
func oversizedInvariants(r *Result) []string {
	var v []string
	var over, overErr int64
	for _, d := range r.Domains {
		over += d.OversizedEmitted
		overErr += sumMap(d.OversizedErrors)
		if d.OversizedOK != 0 {
			v = append(v, fmt.Sprintf("%s: %d oversized events were accepted, want 0", d.Domain, d.OversizedOK))
		}
		if d.OversizedEmitted > 0 && d.OversizedSample != "" && !strings.Contains(d.OversizedSample, "exceeds max gRPC request size") {
			v = append(v, fmt.Sprintf("%s: oversized error is %q, want the local max gRPC request size error", d.Domain, d.OversizedSample))
		}
		if d.Server.MaxBatchBytes > int64(r.Config.MaxGRPCRequestSize) {
			v = append(v, fmt.Sprintf("%s: server saw a %d byte request, over the %d cap", d.Domain, d.Server.MaxBatchBytes, r.Config.MaxGRPCRequestSize))
		}
	}
	if over == 0 {
		v = append(v, "no oversized events were emitted")
	}
	if over != overErr {
		v = append(v, fmt.Sprintf("oversized emitted %d != failed %d", over, overErr))
	}
	// No effect on other messages: every drop is an oversized event.
	if r.Dropped != overErr {
		v = append(v, fmt.Sprintf("%d events dropped but only %d were oversized: oversized events affected others", r.Dropped, overErr))
	}
	return v
}

// outagePlanBar always reports the plan section 5 outage bar. This path has no retry, so every
// event emitted while chip-ingress is down is dropped whatever the buffer size.
func outagePlanBar(r *Result) []string {
	var unavailable int64
	for k, v := range r.DroppedByCode {
		if k == "Unavailable" {
			unavailable += v
		}
	}
	if r.Dropped == 0 {
		return []string{"plan bar (no drops in a 5 s outage) MET: 0 dropped"}
	}
	return []string{fmt.Sprintf("plan bar (no drops in a 5 s outage) NOT MET: %d dropped (%d Unavailable), no retry on this path", r.Dropped, unavailable)}
}

// Scenarios is the stable, ordered scenario table (plan section 3). Durations are short so the
// whole suite runs in a few minutes; LOADTEST_DURATION overrides them.
func Scenarios() []Scenario {
	slow := func(name string, lat time.Duration) Scenario {
		return Scenario{
			Name:        name,
			Description: fmt.Sprintf("peak load, fake RTT %s (per-domain max ~ concurrency x batch / RTT)", lat),
			Loads:       streamsAndFeeds(peakRate),
			Duration:    6 * time.Second,
			Server:      rtt(lat, 0),
		}
	}
	s250 := slow("slow_rtt_250ms", 250*time.Millisecond)
	s250.Check = zeroDrops // plan section 5: zero drops at RTT <= 250 ms

	return []Scenario{
		{
			Name:        "steady_peak",
			Description: "16k msg/s on data-streams plus 40 msg/s on data-feeds, fake RTT 50-100ms",
			Loads:       streamsAndFeeds(peakRate),
			Duration:    8 * time.Second,
			Server:      rtt(50*time.Millisecond, 50*time.Millisecond),
			Check:       zeroDrops,
		},
		{
			Name:        "headroom_1_5x",
			Description: "1.5x peak (24k msg/s data-streams), RTT 100ms; pass bar: zero drops",
			Loads:       streamsAndFeeds(1.5 * peakRate),
			Duration:    6 * time.Second,
			Server:      rtt(100*time.Millisecond, 0),
			Check:       zeroDrops,
		},
		{
			Name:        "headroom_2x",
			Description: "2x peak (32k msg/s data-streams), RTT 100ms; drop rate recorded, not pass/fail",
			Loads:       streamsAndFeeds(2 * peakRate),
			Duration:    6 * time.Second,
			Server:      rtt(100*time.Millisecond, 0),
		},
		s250,
		slow("slow_rtt_500ms", 500*time.Millisecond),
		slow("slow_rtt_1s", time.Second),
		slow("slow_rtt_5s", 5*time.Second),
		{
			Name:        "outage_1s",
			Description: "peak load; chip-ingress returns Unavailable for [2s,3s), then recovers",
			Loads:       streamsAndFeeds(peakRate),
			Duration:    8 * time.Second,
			Server:      ServerProfile{Latency: 50 * time.Millisecond, Jitter: 50 * time.Millisecond, Outages: []Window{{2 * time.Second, 3 * time.Second}}},
		},
		{
			Name:        "outage_5s",
			Description: "peak load; chip-ingress returns Unavailable for [2s,7s), then recovers",
			Loads:       streamsAndFeeds(peakRate),
			Duration:    12 * time.Second,
			Server:      ServerProfile{Latency: 50 * time.Millisecond, Jitter: 50 * time.Millisecond, Outages: []Window{{2 * time.Second, 7 * time.Second}}},
			Notes:       outagePlanBar,
		},
		{
			Name:        "headroom_1_5x_rtt250",
			Description: "1.5x peak (24k msg/s data-streams) at RTT 250ms: the plan section 5 pass-bar condition; zero drops",
			Loads:       streamsAndFeeds(1.5 * peakRate),
			Duration:    6 * time.Second,
			Server:      rtt(250*time.Millisecond, 0),
			Check:       zeroDrops,
		},
		// Extended scenarios: LOADTEST_EXTENDED=1 or select by name with LOADTEST_SCENARIO.
		{
			Name:        "outage_30s",
			Description: "extended: peak load; Unavailable for [2s,32s), then recovers",
			Loads:       streamsAndFeeds(peakRate),
			Duration:    40 * time.Second,
			Server:      ServerProfile{Latency: 50 * time.Millisecond, Jitter: 50 * time.Millisecond, Outages: []Window{{2 * time.Second, 32 * time.Second}}},
			Extended:    true,
		},
		{
			Name:        "slow_rtt_9s",
			Description: "extended: peak load, RTT 9s (just under the 10s SendTimeout)",
			Loads:       streamsAndFeeds(peakRate),
			Duration:    6 * time.Second,
			Server:      rtt(9*time.Second, 0),
			Extended:    true,
		},
		{
			Name:        "peak_split_domains",
			Description: "extended: 16k msg/s total split evenly across data-streams, data-feeds and ccip, RTT 50-100ms",
			Loads: []DomainLoad{
				{Domain: beholder.DomainDataStreams, Rate: peakRate / 3, Payload: typicalPayload},
				{Domain: beholder.DomainDataFeeds, Rate: peakRate / 3, Payload: typicalPayload},
				{Domain: beholder.DomainCCIP, Rate: peakRate / 3, Payload: typicalPayload},
			},
			Duration: 8 * time.Second,
			Server:   rtt(50*time.Millisecond, 50*time.Millisecond),
			Check:    zeroDrops,
			Extended: true,
		},
		{
			Name:        "restart_full_buffer",
			Description: "peak load against RTT 1s so the buffer is full; stop load, Close with DrainTimeout 5s (shortened from 30s)",
			Loads:       streamsAndFeeds(peakRate),
			Duration:    6 * time.Second,
			Server:      rtt(time.Second, 0),
			Tune:        func(c *beholder.Config) { c.ChipIngressDrainTimeout = 5 * time.Second },
		},
		{
			Name:        "restart_full_buffer_short_drain",
			Description: "as restart_full_buffer but RTT 2s and DrainTimeout 1s, to show what a too-short drain drops",
			Loads:       streamsAndFeeds(peakRate),
			Duration:    6 * time.Second,
			Server:      rtt(2*time.Second, 0),
			Tune:        func(c *beholder.Config) { c.ChipIngressDrainTimeout = time.Second },
		},
		{
			Name:        "memory_all_domains",
			Description: "all 3 domains at 16k msg/s, ~2.7 KiB mean payloads (up to 32 KiB), RTT 1s so every buffer stays full",
			Loads: []DomainLoad{
				{Domain: beholder.DomainDataStreams, Rate: peakRate, Payload: largePayload},
				{Domain: beholder.DomainDataFeeds, Rate: peakRate, Payload: largePayload},
				{Domain: beholder.DomainCCIP, Rate: peakRate, Payload: largePayload},
			},
			Duration: 8 * time.Second,
			Server:   rtt(time.Second, 0),
			Check:    memoryBar,
		},
		{
			Name:        "oversized_messages",
			Description: "2k msg/s data-streams + 40 msg/s data-feeds, every 500th data-streams event is 11 MiB (> 10 MiB cap); they must fail locally, alone",
			Loads: []DomainLoad{
				{Domain: beholder.DomainDataStreams, Rate: 2000, Payload: typicalPayload, OversizedEvery: 500, OversizedSize: mib10 + 1<<20},
				{Domain: beholder.DomainDataFeeds, Rate: feedsRate, Payload: typicalPayload},
			},
			Duration:   5 * time.Second,
			Server:     rtt(50*time.Millisecond, 0),
			Invariants: oversizedInvariants,
		},
	}
}

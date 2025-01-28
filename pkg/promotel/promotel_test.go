package promotel_test

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/model/exemplar"
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/metadata"
	"github.com/prometheus/prometheus/scrape"
	"github.com/prometheus/prometheus/storage"
	"github.com/stretchr/testify/require"
)

// TestScrapeLoopScrapeAndReport exercises scrapeAndReport with various scenarios
// (successful scrape, failed scrape, forced error, empty body leading to staleness, etc.).
func TestScrapeLoopScrapeAndReport(t *testing.T) {
	appendable := &collectResultAppendable{&testAppender{}}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reg := prometheus.NewRegistry()
	sl, err := scrape.NewGathererLoop(ctx, nil, appendable, reg, reg, 10*time.Millisecond)
	require.NoError(t, err)

	start := time.Now()
	sl.ScrapeAndReport(time.Time{}, start, nil)
	// The collectResultAppender holds all appended samples. Check the last appended
	// for staleness or actual data, depending on if the scrape was declared OK.
	allSamples := appendable.resultFloats
	// We expect at least one normal sample plus the reported samples.
	require.NotEmpty(t, allSamples, "Expected to see appended samples.")

	// reset the appender
	appendable.testAppender = &testAppender{}
	// create counter metric
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "metric_a",
		Help: "metric_a help",
	}, []string{"label_a"})
	reg.MustRegister(counter)
	counter.WithLabelValues("value_a").Add(42)

	mfs, err := reg.Gather()
	require.NoError(t, err)
	// verify that metric_a is present in Gatherer results
	var foundMetric bool
	for _, mf := range mfs {
		if mf.GetName() == "metric_a" {
			// verify metrics value
			require.Len(t, mf.GetMetric(), 1)
			require.Equal(t, "value_a", mf.GetMetric()[0].GetLabel()[0].GetValue())
			require.Equal(t, 42.0, mf.GetMetric()[0].GetCounter().GetValue()) // nolint
			foundMetric = true
			break
		}
	}
	require.True(t, foundMetric, "Expected to see the 'metric_a' counter metric.")

	sl.ScrapeAndReport(time.Time{}, start, nil)
	// Get all appended samples
	allSamples = appendable.resultFloats
	// verify that the counter metric 'metric_a' was reported
	var found bool
	for _, s := range allSamples {
		if s.metric.Get("__name__") == "metric_a" && s.metric.Get("label_a") == "value_a" {
			found = true
			require.Equal(t, 42.0, s.f) // nolint
		}
	}
	require.True(t, found, "Expected to see the 'metric_a' counter metric.")
}

type floatSample struct {
	metric labels.Labels
	t      int64
	f      float64
}

type histogramSample struct {
	t  int64
	h  *histogram.Histogram
	fh *histogram.FloatHistogram
}

type collectResultAppendable struct {
	*testAppender
}

func (a *collectResultAppendable) Appender(_ context.Context) storage.Appender {
	return a
}

// testAppender records all samples that were added through the appender.
// It can be used as its zero value or be backed by another appender it writes samples through.
type testAppender struct {
	mtx sync.Mutex

	next                 storage.Appender
	resultFloats         []floatSample
	pendingFloats        []floatSample
	rolledbackFloats     []floatSample
	resultHistograms     []histogramSample
	pendingHistograms    []histogramSample
	rolledbackHistograms []histogramSample
	resultExemplars      []exemplar.Exemplar
	pendingExemplars     []exemplar.Exemplar
	resultMetadata       []metadata.Metadata
	pendingMetadata      []metadata.Metadata
}

func (a *testAppender) Append(ref storage.SeriesRef, lset labels.Labels, t int64, v float64) (storage.SeriesRef, error) {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.pendingFloats = append(a.pendingFloats, floatSample{
		metric: lset,
		t:      t,
		f:      v,
	})

	if ref == 0 {
		ref = storage.SeriesRef(rand.Uint64())
	}
	if a.next == nil {
		return ref, nil
	}

	ref, err := a.next.Append(ref, lset, t, v)
	if err != nil {
		return 0, err
	}
	return ref, err
}

func (a *testAppender) AppendExemplar(ref storage.SeriesRef, l labels.Labels, e exemplar.Exemplar) (storage.SeriesRef, error) {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.pendingExemplars = append(a.pendingExemplars, e)
	if a.next == nil {
		return 0, nil
	}

	return a.next.AppendExemplar(ref, l, e)
}

func (a *testAppender) AppendHistogram(ref storage.SeriesRef, l labels.Labels, t int64, h *histogram.Histogram, fh *histogram.FloatHistogram) (storage.SeriesRef, error) {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.pendingHistograms = append(a.pendingHistograms, histogramSample{h: h, fh: fh, t: t})
	if a.next == nil {
		return 0, nil
	}

	return a.next.AppendHistogram(ref, l, t, h, fh)
}

func (a *testAppender) UpdateMetadata(ref storage.SeriesRef, l labels.Labels, m metadata.Metadata) (storage.SeriesRef, error) {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.pendingMetadata = append(a.pendingMetadata, m)
	if ref == 0 {
		ref = storage.SeriesRef(rand.Uint64())
	}
	if a.next == nil {
		return ref, nil
	}

	return a.next.UpdateMetadata(ref, l, m)
}

func (a *testAppender) AppendCTZeroSample(ref storage.SeriesRef, l labels.Labels, _, ct int64) (storage.SeriesRef, error) {
	return a.Append(ref, l, ct, 0.0)
}

func (a *testAppender) Commit() error {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.resultFloats = append(a.resultFloats, a.pendingFloats...)
	a.resultExemplars = append(a.resultExemplars, a.pendingExemplars...)
	a.resultHistograms = append(a.resultHistograms, a.pendingHistograms...)
	a.resultMetadata = append(a.resultMetadata, a.pendingMetadata...)
	a.pendingFloats = nil
	a.pendingExemplars = nil
	a.pendingHistograms = nil
	a.pendingMetadata = nil
	if a.next == nil {
		return nil
	}
	return a.next.Commit()
}

func (a *testAppender) Rollback() error {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.rolledbackFloats = a.pendingFloats
	a.rolledbackHistograms = a.pendingHistograms
	a.pendingFloats = nil
	a.pendingHistograms = nil
	if a.next == nil {
		return nil
	}
	return a.next.Rollback()
}

func (a *testAppender) String() string {
	var sb strings.Builder
	for _, s := range a.resultFloats {
		sb.WriteString(fmt.Sprintf("committed: %s %f %d\n", s.metric, s.f, s.t))
	}
	for _, s := range a.pendingFloats {
		sb.WriteString(fmt.Sprintf("pending: %s %f %d\n", s.metric, s.f, s.t))
	}
	for _, s := range a.rolledbackFloats {
		sb.WriteString(fmt.Sprintf("rolledback: %s %f %d\n", s.metric, s.f, s.t))
	}
	return sb.String()
}

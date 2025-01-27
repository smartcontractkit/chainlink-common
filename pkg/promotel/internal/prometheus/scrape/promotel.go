package scrape

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/textparse"
	"github.com/prometheus/prometheus/storage"

	internaltextparse "github.com/smartcontractkit/chainlink-common/pkg/promotel/internal/prometheus/textparse"
)

type GathereLoop struct {
	*scrapeLoop
	g prometheus.Gatherer
}

func (gl *GathereLoop) newParser() (textparse.Parser, error) {
	mfs, err := gl.g.Gather()
	if err != nil {
		_ = gl.l.Log("msg", "Error while gathering metrics", "err", err)
		return nil, err
	}
	return internaltextparse.NewProtobufParserShim(gl.scrapeClassicHistograms, gl.symbolTable, mfs), err

}

func (gl *GathereLoop) Run(errc chan<- error) {
	gl.scrapeLoop.run(errc)
}

func (gl *GathereLoop) Stop() {
	gl.scrapeLoop.stop()
}

// UnregisterMetrics
func (gl *GathereLoop) UnregisterMetrics() {
	if gl.scrapeLoop.metrics != nil {
		gl.scrapeLoop.metrics.Unregister()
	}
}

func (gl *GathereLoop) ScrapeAndReport(
	last, appendTime time.Time, errc chan<- error,
) time.Time {
	return gl.scrapeAndReport(last, appendTime, errc)
}

func noopScrapeFunc(context.Context, io.Writer) error { return nil }

func newNoopTarget(lbls labels.Labels) *Target {
	return &Target{labels: lbls}
}

func NewGathererLoop(ctx context.Context, logger log.Logger, app storage.Appendable, reg prometheus.Registerer, g prometheus.Gatherer, interval time.Duration) (*GathereLoop, error) {
	nopMutator := func(l labels.Labels) labels.Labels { return l }
	metrics, err := newScrapeMetrics(reg)
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = log.NewNopLogger()
	}
	target := newNoopTarget([]labels.Label{
		{Name: model.JobLabel, Value: "promotel"},      // required label
		{Name: model.InstanceLabel, Value: "promotel"}, // required label
		{Name: model.ScrapeIntervalLabel, Value: interval.String()},
		{Name: model.MetricsPathLabel, Value: config.DefaultScrapeConfig.MetricsPath},
		{Name: model.SchemeLabel, Value: config.DefaultScrapeConfig.Scheme},
	})
	loop := &GathereLoop{
		newScrapeLoop(
			ctx,
			&scraperShim{scrapeFunc: noopScrapeFunc},
			logger,
			nil,
			nopMutator,
			nopMutator,
			func(ctx context.Context) storage.Appender { return app.Appender(ctx) },
			nil,
			labels.NewSymbolTable(),
			0,
			true,
			false,
			true,
			0,
			0,
			histogram.ExponentialSchemaMax,
			nil,
			interval,
			time.Hour,
			false,
			false,
			false,
			false, // todo: pass this from the opts
			false,
			target,
			true,
			metrics,
			true,
		),
		g,
	}
	// Override the newParser function to use the gatherer.
	loop.scrapeLoop.newParserFunc = loop.newParser
	return loop, nil
}

// scraperShim implements the scraper interface and allows setting values
// returned by its methods. It also allows setting a custom scrape function.
type scraperShim struct {
	offsetDur time.Duration

	lastStart    time.Time
	lastDuration time.Duration
	lastError    error

	scrapeErr  error
	scrapeFunc func(context.Context, io.Writer) error
}

func (ts *scraperShim) offset(time.Duration, uint64) time.Duration {
	return ts.offsetDur
}

func (ts *scraperShim) Report(start time.Time, duration time.Duration, err error) {
	ts.lastStart = start
	ts.lastDuration = duration
	ts.lastError = err
}

func (ts *scraperShim) scrape(ctx context.Context) (*http.Response, error) {
	return nil, ts.scrapeErr
}

func (ts *scraperShim) readResponse(ctx context.Context, resp *http.Response, w io.Writer) (string, error) {
	if ts.scrapeFunc != nil {
		return "", ts.scrapeFunc(ctx, w)
	}
	return "", ts.scrapeErr
}

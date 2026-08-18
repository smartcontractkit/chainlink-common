package pg

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/stretchr/testify/mock"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// testDbStater implements mocks for the function signatures
// needed by the stat reporte wrapper for statFn
type testDbStater struct {
	mock.Mock
	t         *testing.T
	name      string
	testGauge prometheus.Gauge
}

func newtestDbStater(t *testing.T, name string) *testDbStater {
	return &testDbStater{
		t:    t,
		name: name,
		testGauge: promauto.NewGauge(prometheus.GaugeOpts{
			Name: strings.ReplaceAll(name, " ", "_"),
		}),
	}
}

func (s *testDbStater) Stats() sql.DBStats {
	s.Called()
	return sql.DBStats{}
}

func (s *testDbStater) Report(stats sql.DBStats) {
	s.Called()
	s.testGauge.Set(float64(stats.MaxOpenConnections))
}

type statScenario struct {
	name   string
	testFn func(*testing.T, *StatsReporter, time.Duration, int)
}

func TestStatReporter(t *testing.T) {
	interval := 2 * time.Millisecond
	expectedIntervals := 4

	lggr := logger.Test(t)

	for _, scenario := range []statScenario{
		{name: "normal_collect_and_stop", testFn: testCollectAndStop},
		{name: "mutli_start", testFn: testMultiStart},
		{name: "multi_stop", testFn: testMultiStop},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			d := newtestDbStater(t, scenario.name)
			d.Mock.On("Stats").Return(sql.DBStats{})
			d.Mock.On("Report").Return()
			reporter := NewStatsReporter(d.Stats,
				lggr,
				StatsInterval(interval),
				StatsCustomReporterFn(d.Report),
			)

			scenario.testFn(
				t,
				reporter,
				interval,
				expectedIntervals,
			)

			d.AssertCalled(t, "Stats")
			d.AssertCalled(t, "Report")
		})
	}
}

// test normal stop
func testCollectAndStop(t *testing.T, r *StatsReporter, interval time.Duration, n int) {
	r.Start()
	time.Sleep(time.Duration(n) * interval)
	r.Stop()
}

// test multiple start calls are idempotent
func testMultiStart(t *testing.T, r *StatsReporter, interval time.Duration, n int) {
	ticker := time.NewTicker(time.Duration(n) * interval)
	defer ticker.Stop()

	r.Start()
	r.Start()
	<-ticker.C
	r.Stop()
}

// test multiple stop calls are idempotent
func testMultiStop(t *testing.T, r *StatsReporter, interval time.Duration, n int) {
	ticker := time.NewTicker(time.Duration(n) * interval)
	defer ticker.Stop()

	r.Start()
	<-ticker.C
	r.Stop()
	r.Stop()
}

//go:build loadtest

package loadtest

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/batch"
)

// batchSource says which copy of pkg/chipingress/batch is compiled in. pkg/chipingress is a separate
// module pinned in the root go.mod, so unless a go.work or replace points at the worktree this is
// the module cache copy, not the worktree's code.
func batchSource() string {
	pc := reflect.ValueOf(batch.ErrorCodeFor).Pointer()
	file, _ := runtime.FuncForPC(pc).FileLine(pc)
	info := ""
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range bi.Deps {
			if dep.Path == "github.com/smartcontractkit/chainlink-common/pkg/chipingress" {
				info = " module " + dep.Path + "@" + dep.Version
				if dep.Replace != nil {
					info += " => " + dep.Replace.Path
				}
			}
		}
	}
	return filepath.Dir(file) + info
}

// OptionsFromEnv reads LOADTEST_DURATION, LOADTEST_RATE_SCALE and LOADTEST_MEM_LIMIT_MIB.
func optionsFromEnv(t *testing.T) RunOptions {
	t.Helper()
	var o RunOptions
	if s := os.Getenv("LOADTEST_DURATION"); s != "" {
		d, err := time.ParseDuration(s)
		if err != nil || d <= 0 {
			t.Fatalf("bad LOADTEST_DURATION %q: %v", s, err)
		}
		o.Duration = d
	}
	if s := os.Getenv("LOADTEST_RATE_SCALE"); s != "" {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil || f <= 0 {
			t.Fatalf("bad LOADTEST_RATE_SCALE %q: %v", s, err)
		}
		o.RateScale = f
	}
	if s := os.Getenv("LOADTEST_MEM_LIMIT_MIB"); s != "" {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil || f <= 0 {
			t.Fatalf("bad LOADTEST_MEM_LIMIT_MIB %q: %v", s, err)
		}
		o.MemLimitMiB = f
	}
	return o
}

// TestScenarios runs every scenario (or those matching LOADTEST_SCENARIO) sequentially. The
// accounting invariants always fail the test; pass-bar checks fail it only with LOADTEST_STRICT=1.
func TestScenarios(t *testing.T) {
	opts := optionsFromEnv(t)
	strict := os.Getenv("LOADTEST_STRICT") == "1"
	var filter *regexp.Regexp
	if s := os.Getenv("LOADTEST_SCENARIO"); s != "" {
		var err error
		if filter, err = regexp.Compile(s); err != nil {
			t.Fatalf("bad LOADTEST_SCENARIO %q: %v", s, err)
		}
	}

	t.Logf("chipingress/batch under test: %s", batchSource())
	var results []*Result
	if out := os.Getenv("LOADTEST_OUT"); out != "" {
		t.Cleanup(func() {
			if len(results) == 0 {
				return
			}
			if err := WriteReport(out, opts, results); err != nil {
				t.Errorf("write LOADTEST_OUT: %v", err)
			} else {
				t.Logf("results written to %s (+ .md)", out)
			}
		})
	}

	ran := 0
	for _, sc := range Scenarios() {
		if filter != nil && !filter.MatchString(sc.Name) {
			continue
		}
		if filter == nil && sc.Extended && os.Getenv("LOADTEST_EXTENDED") != "1" {
			continue
		}
		ran++
		t.Run(sc.Name, func(t *testing.T) {
			res, err := RunScenario(context.Background(), sc, opts)
			if err != nil {
				t.Fatalf("run: %v", err)
			}
			results = append(results, res)
			t.Logf("%s", res.Summary())
			t.Logf("close=%.2fs pending_at_stop=%d delivered_during_close=%d failed_during_close=%d first_err=%.2fs last_err=%.2fs recovery_first_accept=%.2fs",
				res.CloseDurationSec, res.PendingAtStop, res.DeliveredDuringClose, res.FailedDuringClose,
				res.FirstErrorAtSec, res.LastErrorAtSec, res.RecoveryFirstAcceptSec)
			for _, d := range res.Domains {
				t.Logf("  %-12s target=%.0f/s achieved=%.0f/s emitted=%d accepted=%d dropped=%d peak_pending=%d predicted_max=%.0f/s server{batches=%d accepted=%d mean_batch=%.0f}",
					d.Domain, d.TargetRate, d.AchievedRate, d.Emitted, d.Accepted, d.Dropped, d.PeakPending, d.PredictedMaxRate,
					d.Server.Batches, d.Server.EventsAccepted, float64(d.Server.EventsReceived)/max(1, float64(d.Server.Batches)))
			}
			for _, n := range res.Notes {
				t.Logf("note: %s", n)
			}
			for _, v := range res.InvariantViolations {
				t.Errorf("invariant: %s", v)
			}
			for _, v := range res.StrictViolations {
				if strict {
					t.Errorf("pass bar: %s", v)
				} else {
					t.Logf("pass bar not met (report only; set LOADTEST_STRICT=1 to enforce): %s", v)
				}
			}
		})
	}
	if ran == 0 {
		t.Skipf("no scenario matches LOADTEST_SCENARIO=%q", os.Getenv("LOADTEST_SCENARIO"))
	}
}

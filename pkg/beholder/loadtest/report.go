//go:build loadtest

package loadtest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// Report is the JSON document written to LOADTEST_OUT.
type Report struct {
	GeneratedAt time.Time  `json:"generated_at"`
	GoVersion   string     `json:"go_version"`
	GOOS        string     `json:"goos"`
	GOARCH      string     `json:"goarch"`
	NumCPU      int        `json:"num_cpu"`
	Options     RunOptions `json:"options"`
	Results     []*Result  `json:"scenarios"`
}

// WriteReport writes path as JSON and, next to it, the same results as a markdown table
// (path with its extension replaced by .md).
func WriteReport(path string, opts RunOptions, results []*Result) error {
	rep := Report{GeneratedAt: time.Now().UTC(), GoVersion: runtime.Version(), GOOS: runtime.GOOS,
		GOARCH: runtime.GOARCH, NumCPU: runtime.NumCPU(), Options: opts, Results: results}
	b, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return err
	}
	md := strings.TrimSuffix(path, filepath.Ext(path)) + ".md"
	return os.WriteFile(md, []byte(Markdown(rep)), 0o644)
}

func codeList(m map[string]int64) string {
	if len(m) == 0 {
		return "0"
	}
	ks := make([]string, 0, len(m))
	for k, v := range m {
		ks = append(ks, fmt.Sprintf("%s=%d", k, v))
	}
	sort.Strings(ks)
	return strings.Join(ks, "<br>")
}

// Markdown renders the results as markdown tables (aggregates only, no per-event data).
func Markdown(rep Report) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# Beholder node-side load test results\n\n%s, %s/%s, %d CPUs, %s. Rate scale %.2g.\n\n",
		rep.GeneratedAt.Format(time.RFC3339), rep.GOOS, rep.GOARCH, rep.NumCPU, rep.GoVersion, max(rep.Options.RateScale, 1))
	sb.WriteString("| scenario | load s | emitted | accepted | dropped (by code) | pending | server accepted | rate msg/s | p50 ms | p99 ms | peak heap MiB | rss~ MiB | goroutines | cpu cores |\n")
	sb.WriteString("|---|---|---|---|---|---|---|---|---|---|---|---|---|---|\n")
	for _, r := range rep.Results {
		fmt.Fprintf(&sb, "| %s | %.1f | %d | %d | %s | %d | %d | %.0f | %.0f | %.0f | %.0f | %.0f | %d | %.1f |\n",
			r.Name, r.LoadDurationSec, r.Emitted, r.Accepted, dropCell(r), r.Pending, r.ServerAccepted,
			r.AchievedRate, r.P50Ms, r.P99Ms, r.PeakHeapMiB, r.PeakRSSProxyMiB, r.PeakGoroutines, r.CPUCores)
	}
	sb.WriteString("\n## Per domain\n\n| scenario | domain | target/s | achieved/s | emitted | accepted | dropped (by code) | peak pending events | predicted max/s | server batches | mean batch |\n|---|---|---|---|---|---|---|---|---|---|---|\n")
	for _, r := range rep.Results {
		for _, d := range r.Domains {
			all := map[string]int64{}
			for k, v := range d.SyncErrors {
				all[k] += v
			}
			for k, v := range d.CallbackErrors {
				all[k] += v
			}
			mean := 0.0
			if d.Server.Batches > 0 {
				mean = float64(d.Server.EventsReceived) / float64(d.Server.Batches)
			}
			pred := "n/a"
			if d.PredictedMaxRate > 0 {
				pred = fmt.Sprintf("%.0f", d.PredictedMaxRate)
			}
			fmt.Fprintf(&sb, "| %s | %s | %.0f | %.0f | %d | %d | %s | %d | %s | %d | %.0f |\n",
				r.Name, d.Domain, d.TargetRate, d.AchievedRate, d.Emitted, d.Accepted, codeList(all), d.PeakPending, pred, d.Server.Batches, mean)
		}
	}
	sb.WriteString("\n## Timing and close behaviour\n\n| scenario | first error s | last error s | recovery: first accept after outage s | pending at stop | delivered during Close | failed during Close | Close s | drain timeout ms |\n|---|---|---|---|---|---|---|---|---|\n")
	for _, r := range rep.Results {
		fmt.Fprintf(&sb, "| %s | %s | %s | %s | %d | %d | %d | %.2f | %.0f |\n", r.Name, secOrDash(r.FirstErrorAtSec), secOrDash(r.LastErrorAtSec),
			secOrDash(r.RecoveryFirstAcceptSec), r.PendingAtStop, r.DeliveredDuringClose, r.FailedDuringClose, r.CloseDurationSec, r.Config.DrainTimeoutMs)
	}
	var issues []string
	for _, r := range rep.Results {
		for _, v := range r.InvariantViolations {
			issues = append(issues, fmt.Sprintf("- %s: INVARIANT: %s", r.Name, v))
		}
		for _, v := range r.StrictViolations {
			issues = append(issues, fmt.Sprintf("- %s: pass bar: %s", r.Name, v))
		}
		for _, v := range r.Notes {
			issues = append(issues, fmt.Sprintf("- %s: note: %s", r.Name, v))
		}
	}
	if len(issues) > 0 {
		sb.WriteString("\n## Violations and notes\n\n" + strings.Join(issues, "\n") + "\n")
	}
	return sb.String()
}

func dropCell(r *Result) string {
	if r.Dropped == 0 {
		return "0"
	}
	return fmt.Sprintf("%d<br>%s", r.Dropped, codeList(r.DroppedByCode))
}

func secOrDash(v float64) string {
	if v < 0 {
		return "-"
	}
	return fmt.Sprintf("%.2f", v)
}

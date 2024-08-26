package crontrigger

import (
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type CronMonitor struct{}

var (
	PromRunningServices = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "capability_trigger_cron_running_services",
			Help: "Metric representing the number of cron trigger services that are actively scheduling triggers",
		},
	)
	PromTotalTriggersCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "capability_trigger_cron_total_triggers_count",
			Help: "Metric representing the number of currently active triggers",
		},
	)
	PromExecutionTimeMS = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "capability_trigger_cron_execution_time_ms",
			Help: "Metric representing the execution time in milliseconds, by TriggerID",
		},
		[]string{"trigger_id"},
	)
	PromTaskRunsCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "capability_trigger_cron_runs_count",
			Help: "Metric representing the number of runs completed with status, by TriggerID",
		},
		[]string{"trigger_id", "status"},
	)
)

func NewCronMonitor() gocron.Monitor {
	return &CronMonitor{}
}

// Hooks into gocron after the job has finished a run. Proceeds RecordJobTiming.
func (cm *CronMonitor) IncrementJob(_ uuid.UUID, name string, _ []string, status gocron.JobStatus) {
	PromTaskRunsCount.WithLabelValues(name, string(status)).Inc()
}

// Hooks into gocron after the job has finished a run
func (cm *CronMonitor) RecordJobTiming(startTime, endTime time.Time, _ uuid.UUID, name string, _ []string) {
	PromExecutionTimeMS.WithLabelValues(name).Observe(float64(endTime.Sub(startTime).Milliseconds()))
}

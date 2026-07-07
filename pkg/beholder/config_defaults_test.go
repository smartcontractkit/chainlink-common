package beholder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetricReaderInterval_zeroUsesDefault(t *testing.T) {
	t.Parallel()

	assert.Equal(t, time.Second, metricReaderInterval(Config{}))
	assert.Equal(t, 5*time.Second, metricReaderInterval(Config{MetricReaderInterval: 5 * time.Second}))
}

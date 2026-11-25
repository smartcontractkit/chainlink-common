package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvVarName(t *testing.T) {
	path := []string{"cron", "fastestScheduleIntervalSeconds"}
	assert.Equal(t, "CL_CRON_FASTEST_SCHEDULE_INTERVAL_SECONDS", EnvVarName("CL", path...))
	assert.Equal(t, "CRON_FASTEST_SCHEDULE_INTERVAL_SECONDS", EnvVarName("", path...))
	assert.Equal(t, "CL_CRON_FASTEST_SCHEDULE_INTERVAL_SECONDS", EnvVarName("cl_", path...))
}

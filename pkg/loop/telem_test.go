package loop

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SetupTracing(t *testing.T) {
	err := SetupTracing(TracingConfig{
		Enabled: true,
		CollectorTarget: "some:target:",
		TLSCertPath: "",
		OnDialError: func(err error) {},
	})

	require.NoError(t, err)
}
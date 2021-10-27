package test

import (
	"testing"
)

var (
	MockTestEnv = "http://testEnvVar.link"
)

func MockSetRequiredConfigs(t *testing.T, vars []string) {
	for _, key := range vars {
		// set env (clears after test is complete)
		t.Setenv(key, MockTestEnv) // this cannot be used if run in parallel (released in 1.17)
	}
}

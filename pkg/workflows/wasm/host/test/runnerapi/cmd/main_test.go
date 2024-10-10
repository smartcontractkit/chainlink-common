package main

import (
	"testing"
)

func TestSingleFileCanTestBuildWorkflow(t *testing.T) {
	// No assertions, we're just checking that we don't get
	// `BuildWorkflow` not found.
	_ = BuildWorkflow([]byte(""))
}

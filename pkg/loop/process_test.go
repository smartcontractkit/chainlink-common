package loop_test

import (
	"os/exec"

	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal/test"
)

func HelperProcess(command string, opts test.HelperProcessOptions) *exec.Cmd {
	return test.HelperProcess("./internal/test/cmd/main.go", command, opts)
}

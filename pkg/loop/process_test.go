package loop_test

import (
	"os/exec"

	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal/test"
)

func HelperProcess(command string, opts ...func(o *test.HelperProcessCommand)) *exec.Cmd {
	return test.NewHelperProcess("./internal/test/cmd/main.go", command, opts...)
}

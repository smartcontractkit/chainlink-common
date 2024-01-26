package loop_test

import (
	"os/exec"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
)

type HelperProcessCommand test.HelperProcessCommand

func (h HelperProcessCommand) New() *exec.Cmd {
	h.CommandLocation = "./internal/test/cmd/main.go"
	return (test.HelperProcessCommand)(h).New()
}

func NewHelperProcessCommand(command string) *exec.Cmd {
	return HelperProcessCommand{
		Command: command,
	}.New()
}

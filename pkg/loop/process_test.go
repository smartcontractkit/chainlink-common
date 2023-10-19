package loop_test

import (
	"os/exec"

	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal/test"
)

type HelperProcessCommand struct {
	test.HelperProcessCommand
}

func (h HelperProcessCommand) New() *exec.Cmd {
	h.CommandLocation = "./internal/test/cmd/main.go"
	return h.HelperProcessCommand.New()
}

func NewHelperProcessCommand(command string) *exec.Cmd {
	h := HelperProcessCommand{
		HelperProcessCommand: test.HelperProcessCommand{
			Command: command,
		},
	}
	return h.New()
}

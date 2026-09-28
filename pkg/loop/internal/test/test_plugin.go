package test

import (
	"fmt"
	"os"
	"os/exec"
)

type HelperProcessCommand struct {
	Limit           int
	CommandLocation string
	Command         string
	StaticChecks    bool
	ThrowErrorType  int
}

func (h HelperProcessCommand) New() *exec.Cmd {
	cmdArgs := []string{
		"go", "run", h.CommandLocation, "-cmd=" + h.Command,
	}

	if h.Limit != 0 {
		cmdArgs = append(cmdArgs, fmt.Sprintf("-limit=%d", h.Limit))
	}

	if h.StaticChecks {
		cmdArgs = append(cmdArgs, "-static-checks")
	}

	if h.ThrowErrorType != 0 {
		cmdArgs = append(cmdArgs, fmt.Sprintf("-throw-error-type=%d", h.ThrowErrorType))
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...) // #nosec
	cmd.Env = os.Environ()

	return cmd
}

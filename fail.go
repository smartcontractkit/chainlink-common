package common

import (
	"errors"
	"os"
	"testing"
)

func FailLintOutOfTestFile(t *testing.T) {
	const UnusedVar = 1 // lint should complain for unused variable
	const ALL_CAPS = 10 // should be camel cased
	err := os.ErrNotExist
	if err == os.ErrNotExist { // should use errors.Is
		err := errors.New("fake error") // shadowed variable
		t.Log(err)
	}
}

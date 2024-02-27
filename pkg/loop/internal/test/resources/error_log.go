package resources_test

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var ErrorLogImpl = StaticErrorLog{errMsg: "an error"}

var _ types.ErrorLog = (*StaticErrorLog)(nil)

type StaticErrorLog struct {
	errMsg string
}

func (s *StaticErrorLog) SaveError(ctx context.Context, msg string) error {
	if msg != s.errMsg {
		return fmt.Errorf("expected %q but got %q", s.errMsg, msg)
	}
	return nil
}

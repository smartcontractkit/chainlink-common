package pipeline

import (
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/utils"
)

type Spec struct {
	ID                int32
	DotDagSource      string         `json:"dotDagSource"`
	CreatedAt         time.Time      `json:"-"`
	MaxTaskDuration   utils.Interval `json:"-"`
	GasLimit          *uint32        `json:"-"`
	ForwardingAllowed bool           `json:"-"`

	JobID   int32  `json:"-"`
	JobName string `json:"-"`
	JobType string `json:"-"`
}

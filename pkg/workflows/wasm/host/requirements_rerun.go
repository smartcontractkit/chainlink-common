package host

import (
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

type RequirementsRerun sdk.Requirements

func (r *RequirementsRerun) Error() string {
	str, _ := protojson.Marshal((*sdk.Requirements)(r))
	return string(str)
}

var _ error = (*RequirementsRerun)(nil)

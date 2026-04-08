package host

import sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"

type TeeProvider sdkpb.TeeType

func (t TeeProvider) Provides(tee *sdkpb.Tee) bool {
	switch teet := tee.Type.(type) {
	case *sdkpb.Tee_Any:
		return true
	case *sdkpb.Tee_TypeSelection:
		for _, selection := range teet.TypeSelection.Types {
			if selection == sdkpb.TeeType(t) {
				return true
			}
		}
	}

	return false
}

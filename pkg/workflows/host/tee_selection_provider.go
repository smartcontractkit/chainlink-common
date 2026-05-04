package host

import (
	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func NewProviderFromSelection(types []*sdkpb.TeeTypeAndRegions) func(tee *sdkpb.Tee) bool {
	if len(types) == 1 {
		return NewTeeProvider(types[0].Type, types[0].Regions)
	}

	supplies := make(map[sdkpb.TeeType][]string)
	for _, t := range types {
		supplies[t.Type] = append(supplies[t.Type], t.Regions...)
	}

	providers := make(map[sdkpb.TeeType]func(tee *sdkpb.Tee) bool)
	for k, v := range supplies {
		providers[k] = NewTeeProvider(k, v)
	}

	return func(tee *sdkpb.Tee) bool {
		switch teet := tee.Item.(type) {
		case *sdkpb.Tee_AnyRegions:
			for _, provider := range providers {
				if provider(tee) {
					return true
				}
			}

			return false
		case *sdkpb.Tee_TeeTypesAndRegions:
			for _, requestedType := range teet.TeeTypesAndRegions.TeeTypeAndRegions {
				provider, ok := providers[requestedType.Type]
				if !ok {
					continue
				}

				if provider(tee) {
					return true
				}
			}

			return false
		default:
			return false
		}
	}
}

package host

import sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"

type teeProvider struct {
	sdkpb.TeeType
	regions map[string]bool
}

func NewTeeProvider(tpe sdkpb.TeeType, regions []string) func(tee *sdkpb.Tee) bool {
	supportedRegions := map[string]bool{}
	for _, region := range regions {
		supportedRegions[region] = true
	}
	return (&teeProvider{TeeType: tpe, regions: supportedRegions}).Provides
}

func (t *teeProvider) Provides(tee *sdkpb.Tee) bool {
	switch teet := tee.Type.(type) {
	case *sdkpb.Tee_Any:
		return true
	case *sdkpb.Tee_TypeSelection:
		for _, selection := range teet.TypeSelection.Types {
			if selection.Type == t.TeeType {
				if len(selection.Regions) == 0 {
					return true
				}

				for _, region := range selection.Regions {
					if t.regions[region] {
						return true
					}
				}
			}
		}
	}

	return false
}

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
	var regions []string
	switch teet := tee.Item.(type) {
	case *sdkpb.Tee_AnyRegions:
		regions = teet.AnyRegions.Regions
	case *sdkpb.Tee_TeeTypesAndRegions:
		if teet.TeeTypesAndRegions == nil {
			return false
		}

		found := false
		for _, tr := range teet.TeeTypesAndRegions.TeeTypeAndRegions {
			if tr.Type == t.TeeType {
				found = true
				regions = tr.Regions
				break
			}
		}

		if !found {
			return false
		}
	}

	if len(regions) == 0 {
		return true
	}

	for _, region := range regions {
		if t.regions[region] {
			return true
		}
	}

	return false
}

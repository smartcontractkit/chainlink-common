package host

import (
	"context"
	"sync"

	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

type teeProvider struct {
	sdkpb.TeeType
	regionsFn func(ctx context.Context) map[string]bool
	once      sync.Once
}

func NewTeeProvider(tpe sdkpb.TeeType, regionsFn func(ctx context.Context) []string) func(context.Context, *sdkpb.Tee) bool {
	p := &teeProvider{
		TeeType: tpe,
		regionsFn: func(ctx context.Context) map[string]bool {
			regions := regionsFn(ctx)
			rMap := make(map[string]bool, len(regions))
			for _, region := range regions {
				rMap[region] = true
			}
			return rMap
		},
	}
	return p.Provides
}

func (t *teeProvider) Provides(ctx context.Context, tee *sdkpb.Tee) bool {
	switch teet := tee.Type.(type) {
	case *sdkpb.Tee_Any:
		return true
	case *sdkpb.Tee_TypeSelection:
		for _, selection := range teet.TypeSelection.Types {
			if selection.Type == t.TeeType {
				if len(selection.Regions) == 0 {
					return true
				}

				regions := t.regionsFn(ctx)
				for _, region := range selection.Regions {
					if regions[region] {
						return true
					}
				}
			}
		}
	}

	return false
}

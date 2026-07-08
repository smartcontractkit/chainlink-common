package sothealth

import (
	"context"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/services"

	pbHealthInfo "github.com/smartcontractkit/chainlink-protos/node-platform/common/v1"
)

const (
	beholderDomain             = "node-platform"
	beholderProtobufEntity     = "common.v1.NodeHealthInfo"
	beholderProtobufDataSchema = "/node-platform/common/v1"
)

func NewChecker(ver, sha string) (*services.HealthChecker, error) {
	cfg, err := ConfigureHooks(services.HealthCheckerConfig{Ver: ver, Sha: sha})
	if err != nil {
		return nil, err
	}
	return cfg.New(), nil
}

func ConfigureHooks(orig services.HealthCheckerConfig) (services.HealthCheckerConfig, error) {
	cfg := orig // copy
	cfg.EmitSoTHealthData = func(ctx context.Context, isHealthy bool, svcHealth map[string]error) {
		emitter := beholder.GetEmitter()
		svcHealthErrs := make(map[string]string)
		if len(svcHealth) > 0 {
			for svcName, err := range svcHealth {
				svcHealthErrs[svcName] = err.Error()
			}
		}
		payloadBytes, err := proto.Marshal(&pbHealthInfo.NodeHealthInfo{
			Healthy:      isHealthy,
			HealthErrors: svcHealthErrs,
		})
		if err != nil {
			cfg.Logger.Errorw("failed to marshal node-platform health info", "err", err)
			return
		}
		err = emitter.Emit(ctx, payloadBytes,
			beholder.AttrKeyDomain, beholderDomain,
			beholder.AttrKeyEntity, beholderProtobufEntity,
			beholder.AttrKeyDataSchema, beholderProtobufDataSchema,
		)
		if err != nil {
			cfg.Logger.Errorw("failed to emit node-platform health info", "err", err)
		}
	}
	return cfg, nil
}

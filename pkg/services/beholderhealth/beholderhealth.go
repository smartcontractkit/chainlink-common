package beholderhealth

import (
	"context"
	"errors"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"

	pbHealthInfo "github.com/smartcontractkit/chainlink-protos/node-platform/common/v1"
)

const (
	beholderDomain             = "node-platform"
	beholderProtobufEntity     = "common.v1.HealthInfo"
	beholderProtobufDataSchema = "/node-platform/common/v1"
)

func NewChecker(ver, sha string, lggr logger.Logger, emitter beholder.Emitter) (*services.HealthChecker, error) {
	cfg, err := ConfigureHooks(services.HealthCheckerConfig{Ver: ver, Sha: sha}, lggr, emitter)
	if err != nil {
		return nil, err
	}
	return cfg.New(), nil
}

func ConfigureHooks(orig services.HealthCheckerConfig, lggr logger.Logger, emitter beholder.Emitter) (services.HealthCheckerConfig, error) {
	if lggr == nil {
		return orig, errors.New("logger can't be nil")
	}
	if emitter == nil {
		return orig, errors.New("emitter can't be nil")
	}
	cfg := orig // copy
	
	cfg.SetHealth = func(ctx context.Context, svcHealth map[string]error) {
		if orig.SetHealth != nil {
			orig.SetHealth(ctx, svcHealth)
		}
		allSvcHealthy := true
		svcHealthErrs := make(map[string]string)

		for svcName, svcErr := range svcHealth {
			if svcErr != nil {
				svcHealthErrs[svcName] = svcErr.Error()
				allSvcHealthy = false
			}
		}

		payloadBytes, err := proto.Marshal(&pbHealthInfo.HealthInfo{
			Healthy:      allSvcHealthy,
			HealthErrors: svcHealthErrs,
		})
		if err != nil {
			lggr.Errorw("failed to marshal node-platform health info", "err", err)
			return
		}

		err = emitter.Emit(ctx, payloadBytes,
			beholder.AttrKeyDomain, beholderDomain,
			beholder.AttrKeyEntity, beholderProtobufEntity,
			beholder.AttrKeyDataSchema, beholderProtobufDataSchema,
		)
		if err != nil {
			lggr.Errorw("failed to emit node-platform health info", "err", err)
		}
	}
	return cfg, nil
}

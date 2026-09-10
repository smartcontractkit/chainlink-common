package pb

import (
	"encoding/hex"
	"fmt"
	"math"

	"google.golang.org/protobuf/types/known/durationpb"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
	valuespb "github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"
)

// CapabilityConfigFromProto decodes a wire CapabilityConfig into the Go type.
//
// This is the single decoder for capability configuration. Both registry
// transports carry the same message — the go-plugin registry returns it inline,
// the plain-gRPC registry carries it as opaque bytes — and the on-chain registry
// stores exactly these bytes per DON. A second decoder anywhere could only drift
// from this one, so callers should route through here rather than reimplement it.
func CapabilityConfigFromProto(cfg *CapabilityConfig) (capabilities.CapabilityConfiguration, error) {
	if cfg == nil {
		return capabilities.CapabilityConfiguration{}, nil
	}

	defaultConfig, err := values.FromMapValueProto(cfg.DefaultConfig)
	if err != nil {
		return capabilities.CapabilityConfiguration{}, fmt.Errorf("could not convert map valueproto to map: %w", err)
	}

	// RestrictedConfig takes precedence over anything a user supplies, so dropping it silently would
	// hand the user's own value authority it should never have.
	restrictedConfig, err := values.FromMapValueProto(cfg.RestrictedConfig)
	if err != nil {
		return capabilities.CapabilityConfiguration{}, fmt.Errorf("could not convert restricted config valueproto to map: %w", err)
	}

	var (
		remoteTriggerConfig    *capabilities.RemoteTriggerConfig
		remoteTargetConfig     *capabilities.RemoteTargetConfig
		remoteExecutableConfig *capabilities.RemoteExecutableConfig
	)
	switch cfg.RemoteConfig.(type) {
	case *CapabilityConfig_RemoteTriggerConfig:
		remoteTriggerConfig = decodeRemoteTriggerConfig(cfg.GetRemoteTriggerConfig())
	case *CapabilityConfig_RemoteTargetConfig:
		remoteTargetConfig = &capabilities.RemoteTargetConfig{
			RequestHashExcludedAttributes: cfg.GetRemoteTargetConfig().RequestHashExcludedAttributes,
		}
	case *CapabilityConfig_RemoteExecutableConfig:
		remoteExecutableConfig = decodeRemoteExecutableConfig(cfg.GetRemoteExecutableConfig())
	}

	var methodConfig map[string]capabilities.CapabilityMethodConfig
	if cfg.MethodConfigs != nil {
		methodConfig = make(map[string]capabilities.CapabilityMethodConfig, len(cfg.MethodConfigs))
		for name, mCfg := range cfg.MethodConfigs {
			decoded := capabilities.CapabilityMethodConfig{}
			switch mCfg.RemoteConfig.(type) {
			case *CapabilityMethodConfig_RemoteTriggerConfig:
				decoded.RemoteTriggerConfig = decodeRemoteTriggerConfig(mCfg.GetRemoteTriggerConfig())
			case *CapabilityMethodConfig_RemoteExecutableConfig:
				decoded.RemoteExecutableConfig = decodeRemoteExecutableConfig(mCfg.GetRemoteExecutableConfig())
			default:
				// Failing loudly beats handing back a method config with no remote settings at all,
				// which reads as "callable with defaults" rather than "not understood".
				return capabilities.CapabilityConfiguration{}, fmt.Errorf("unknown method config type for method %s", name)
			}
			if mCfg.AggregatorConfig != nil {
				decoded.AggregatorConfig = &capabilities.AggregatorConfig{
					AggregatorType: capabilities.AggregatorType(mCfg.AggregatorConfig.AggregatorType),
				}
			}
			methodConfig[name] = decoded
		}
	}

	var ocr3Configs map[string]ocrtypes.ContractConfig
	if cfg.Ocr3Configs != nil {
		ocr3Configs = make(map[string]ocrtypes.ContractConfig, len(cfg.Ocr3Configs))
		for key, pbCfg := range cfg.Ocr3Configs {
			decodedOcr3, err := decodeOcr3Config(pbCfg)
			if err != nil {
				return capabilities.CapabilityConfiguration{}, fmt.Errorf("could not decode OCR3 config %q: %w", key, err)
			}
			ocr3Configs[key] = decodedOcr3
		}
	}

	var oracleFactoryConfigs map[string]values.Map
	if cfg.OracleFactoryConfigs != nil {
		oracleFactoryConfigs = make(map[string]values.Map, len(cfg.OracleFactoryConfigs))
		for key, pbMap := range cfg.OracleFactoryConfigs {
			m, err := values.FromMapValueProto(pbMap)
			if err != nil {
				return capabilities.CapabilityConfiguration{}, fmt.Errorf("could not decode oracle factory config for key %s: %w", key, err)
			}
			if m != nil {
				oracleFactoryConfigs[key] = *m
			}
		}
	}

	specConfig, err := values.FromMapValueProto(cfg.SpecConfig)
	if err != nil {
		return capabilities.CapabilityConfiguration{}, fmt.Errorf("could not decode spec config: %w", err)
	}

	return capabilities.CapabilityConfiguration{
		DefaultConfig:          defaultConfig,
		RestrictedKeys:         cfg.RestrictedKeys,
		RestrictedConfig:       restrictedConfig,
		RemoteTriggerConfig:    remoteTriggerConfig,
		RemoteTargetConfig:     remoteTargetConfig,
		RemoteExecutableConfig: remoteExecutableConfig,
		CapabilityMethodConfig: methodConfig,
		LocalOnly:              cfg.LocalOnly,
		Ocr3Configs:            ocr3Configs,
		OracleFactoryConfigs:   oracleFactoryConfigs,
		SpecConfig:             specConfig,
	}, nil
}

func decodeRemoteTriggerConfig(prtc *RemoteTriggerConfig) *capabilities.RemoteTriggerConfig {
	if prtc == nil {
		return nil
	}
	return &capabilities.RemoteTriggerConfig{
		RegistrationRefresh:     prtc.RegistrationRefresh.AsDuration(),
		RegistrationExpiry:      prtc.RegistrationExpiry.AsDuration(),
		MinResponsesToAggregate: prtc.MinResponsesToAggregate,
		MessageExpiry:           prtc.MessageExpiry.AsDuration(),
		MaxBatchSize:            prtc.MaxBatchSize,
		BatchCollectionPeriod:   prtc.BatchCollectionPeriod.AsDuration(),
	}
}

func decodeRemoteExecutableConfig(prec *RemoteExecutableConfig) *capabilities.RemoteExecutableConfig {
	if prec == nil {
		return nil
	}
	return &capabilities.RemoteExecutableConfig{
		RequestHashExcludedAttributes: prec.RequestHashExcludedAttributes,
		TransmissionSchedule:          capabilities.TransmissionSchedule(prec.TransmissionSchedule),
		DeltaStage:                    prec.DeltaStage.AsDuration(),
		RequestTimeout:                prec.RequestTimeout.AsDuration(),
		ServerMaxParallelRequests:     prec.ServerMaxParallelRequests,
		RequestHasherType:             capabilities.RequestHasherType(prec.RequestHasherType),
		MinResponsesToAggregate:       prec.MinResponsesToAggregate,
	}
}

// OCR3ConfigFromProto decodes an OCR3 configuration and stamps it with the
// digest that identifies it.
//
// The digest is supplied rather than decoded because the wire form does not
// carry one: it covers the configuration together with the chain and address
// the registry was read from, so only whoever read that contract can compute
// it. See core.OCRConfigRegistry.
func OCR3ConfigFromProto(pbCfg *OCR3Config, digest ocrtypes.ConfigDigest) (ocrtypes.ContractConfig, error) {
	cfg, err := decodeOcr3Config(pbCfg)
	if err != nil {
		return ocrtypes.ContractConfig{}, err
	}
	cfg.ConfigDigest = digest
	return cfg, nil
}

func decodeOcr3Config(pbCfg *OCR3Config) (ocrtypes.ContractConfig, error) {
	signers := make([]ocrtypes.OnchainPublicKey, len(pbCfg.Signers))
	for i, s := range pbCfg.Signers {
		signers[i] = ocrtypes.OnchainPublicKey(s)
	}
	transmitters := make([]ocrtypes.Account, len(pbCfg.Transmitters))
	for i, t := range pbCfg.Transmitters {
		transmitters[i] = ocrtypes.Account(hex.EncodeToString(t))
	}
	// F is the fault tolerance the protocol is run at, so a value that does not fit has to be an
	// error: truncating it would silently run OCR at a completely different F than configured.
	if pbCfg.F > math.MaxUint8 {
		return ocrtypes.ContractConfig{}, fmt.Errorf("F value %d exceeds uint8 max", pbCfg.F)
	}
	return ocrtypes.ContractConfig{
		ConfigCount:           pbCfg.ConfigCount,
		Signers:               signers,
		Transmitters:          transmitters,
		F:                     uint8(pbCfg.F), //#nosec G115 - bounds checked above
		OnchainConfig:         pbCfg.OnchainConfig,
		OffchainConfigVersion: pbCfg.OffchainConfigVersion,
		OffchainConfig:        pbCfg.OffchainConfig,
		// NOTE: ConfigDigest is appended later by ContractConfigTracker.
	}, nil
}

// CapabilityConfigToProto encodes a capability configuration to its wire form.
//
// The counterpart of [CapabilityConfigFromProto] and the single encoder: both
// registry transports carry the same message, so a second encoder anywhere
// could only drift from this one.
func CapabilityConfigToProto(cc capabilities.CapabilityConfiguration) (*CapabilityConfig, error) {
	ccp := &CapabilityConfig{
		DefaultConfig: values.Proto(cc.DefaultConfig).GetMapValue(),
		LocalOnly:     cc.LocalOnly,
	}

	if cc.RestrictedConfig != nil {
		ccp.RestrictedConfig = values.ProtoMap(cc.RestrictedConfig)
	}
	ccp.RestrictedKeys = cc.RestrictedKeys

	if cc.RemoteTriggerConfig != nil {
		ccp.RemoteConfig = &CapabilityConfig_RemoteTriggerConfig{
			RemoteTriggerConfig: encodeRemoteTriggerConfig(cc.RemoteTriggerConfig),
		}
	}

	if cc.RemoteTargetConfig != nil {
		ccp.RemoteConfig = &CapabilityConfig_RemoteTargetConfig{
			RemoteTargetConfig: &RemoteTargetConfig{
				RequestHashExcludedAttributes: cc.RemoteTargetConfig.RequestHashExcludedAttributes,
			},
		}
	}

	if cc.RemoteExecutableConfig != nil {
		ccp.RemoteConfig = &CapabilityConfig_RemoteExecutableConfig{
			RemoteExecutableConfig: encodeRemoteExecutableConfig(cc.RemoteExecutableConfig),
		}
	}

	if cc.CapabilityMethodConfig != nil {
		ccp.MethodConfigs = make(map[string]*CapabilityMethodConfig, len(cc.CapabilityMethodConfig))
		for name, mCfg := range cc.CapabilityMethodConfig {
			pbMethodConfig := &CapabilityMethodConfig{}
			if mCfg.RemoteTriggerConfig != nil {
				pbMethodConfig.RemoteConfig = &CapabilityMethodConfig_RemoteTriggerConfig{
					RemoteTriggerConfig: encodeRemoteTriggerConfig(mCfg.RemoteTriggerConfig),
				}
			}
			if mCfg.RemoteExecutableConfig != nil {
				pbMethodConfig.RemoteConfig = &CapabilityMethodConfig_RemoteExecutableConfig{
					RemoteExecutableConfig: encodeRemoteExecutableConfig(mCfg.RemoteExecutableConfig),
				}
			}
			if mCfg.AggregatorConfig != nil {
				pbMethodConfig.AggregatorConfig = &AggregatorConfig{
					AggregatorType: AggregatorType(mCfg.AggregatorConfig.AggregatorType),
				}
			}
			ccp.MethodConfigs[name] = pbMethodConfig
		}
	}

	if cc.Ocr3Configs != nil {
		ccp.Ocr3Configs = make(map[string]*OCR3Config, len(cc.Ocr3Configs))
		for key, cfg := range cc.Ocr3Configs {
			signers := make([][]byte, len(cfg.Signers))
			for i, s := range cfg.Signers {
				signers[i] = []byte(s)
			}
			transmitters := make([][]byte, len(cfg.Transmitters))
			for i, t := range cfg.Transmitters {
				decoded, err := hex.DecodeString(string(t))
				if err != nil {
					return nil, fmt.Errorf("failed to decode transmitter: %w", err)
				}
				transmitters[i] = decoded
			}
			ccp.Ocr3Configs[key] = &OCR3Config{
				ConfigCount:           cfg.ConfigCount,
				Signers:               signers,
				Transmitters:          transmitters,
				F:                     uint32(cfg.F), //#nosec G115 - ContractConfig.F is a uint8
				OnchainConfig:         cfg.OnchainConfig,
				OffchainConfigVersion: cfg.OffchainConfigVersion,
				OffchainConfig:        cfg.OffchainConfig,
				// NOTE: ConfigDigest is not passed in the proto, nor stored directly onchain.
			}
		}
	}

	if cc.OracleFactoryConfigs != nil {
		ccp.OracleFactoryConfigs = make(map[string]*valuespb.Map, len(cc.OracleFactoryConfigs))
		for key, m := range cc.OracleFactoryConfigs {
			ccp.OracleFactoryConfigs[key] = values.Proto(&m).GetMapValue()
		}
	}

	if cc.SpecConfig != nil {
		ccp.SpecConfig = values.Proto(cc.SpecConfig).GetMapValue()
	}

	return ccp, nil
}

func encodeRemoteTriggerConfig(cfg *capabilities.RemoteTriggerConfig) *RemoteTriggerConfig {
	return &RemoteTriggerConfig{
		RegistrationRefresh:     durationpb.New(cfg.RegistrationRefresh),
		RegistrationExpiry:      durationpb.New(cfg.RegistrationExpiry),
		MinResponsesToAggregate: cfg.MinResponsesToAggregate,
		MessageExpiry:           durationpb.New(cfg.MessageExpiry),
		MaxBatchSize:            cfg.MaxBatchSize,
		BatchCollectionPeriod:   durationpb.New(cfg.BatchCollectionPeriod),
	}
}

func encodeRemoteExecutableConfig(cfg *capabilities.RemoteExecutableConfig) *RemoteExecutableConfig {
	return &RemoteExecutableConfig{
		RequestHashExcludedAttributes: cfg.RequestHashExcludedAttributes,
		TransmissionSchedule:          TransmissionSchedule(cfg.TransmissionSchedule),
		DeltaStage:                    durationpb.New(cfg.DeltaStage),
		RequestTimeout:                durationpb.New(cfg.RequestTimeout),
		ServerMaxParallelRequests:     cfg.ServerMaxParallelRequests,
		RequestHasherType:             RequestHasherType(cfg.RequestHasherType),
		MinResponsesToAggregate:       cfg.MinResponsesToAggregate,
	}
}

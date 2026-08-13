package pb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
	valuespb "github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"
)

func TestCapabilityConfigFromProto_Nil(t *testing.T) {
	got, err := CapabilityConfigFromProto(nil)
	require.NoError(t, err)
	assert.Equal(t, capabilities.CapabilityConfiguration{}, got)
}

func TestCapabilityConfigFromProto_RemoteTrigger(t *testing.T) {
	defaultCfg, err := values.NewMap(map[string]any{"a": int64(1)})
	require.NoError(t, err)
	restrictedCfg, err := values.NewMap(map[string]any{"b": "two"})
	require.NoError(t, err)

	in := &CapabilityConfig{
		DefaultConfig:    values.ProtoMap(defaultCfg),
		RestrictedConfig: values.ProtoMap(restrictedCfg),
		RestrictedKeys:   []string{"b"},
		RemoteConfig: &CapabilityConfig_RemoteTriggerConfig{
			RemoteTriggerConfig: &RemoteTriggerConfig{
				RegistrationRefresh:     durationpb.New(30 * time.Second),
				RegistrationExpiry:      durationpb.New(2 * time.Minute),
				MinResponsesToAggregate: 3,
				MessageExpiry:           durationpb.New(90 * time.Second),
				MaxBatchSize:            7,
				BatchCollectionPeriod:   durationpb.New(time.Second),
			},
		},
		LocalOnly: true,
	}

	got, err := CapabilityConfigFromProto(in)
	require.NoError(t, err)

	assert.Equal(t, defaultCfg, got.DefaultConfig)
	assert.Equal(t, []string{"b"}, got.RestrictedKeys)
	assert.True(t, got.LocalOnly)

	require.NotNil(t, got.RemoteTriggerConfig)
	assert.Equal(t, 30*time.Second, got.RemoteTriggerConfig.RegistrationRefresh)
	assert.Equal(t, 2*time.Minute, got.RemoteTriggerConfig.RegistrationExpiry)
	assert.Equal(t, uint32(3), got.RemoteTriggerConfig.MinResponsesToAggregate)
	assert.Equal(t, 90*time.Second, got.RemoteTriggerConfig.MessageExpiry)
	assert.Equal(t, uint32(7), got.RemoteTriggerConfig.MaxBatchSize)
	assert.Equal(t, time.Second, got.RemoteTriggerConfig.BatchCollectionPeriod)

	// Only the oneof arm that was set should be populated.
	assert.Nil(t, got.RemoteTargetConfig)
	assert.Nil(t, got.RemoteExecutableConfig)
}

func TestCapabilityConfigFromProto_RemoteExecutableAndMethodConfigs(t *testing.T) {
	in := &CapabilityConfig{
		RemoteConfig: &CapabilityConfig_RemoteExecutableConfig{
			RemoteExecutableConfig: &RemoteExecutableConfig{
				RequestHashExcludedAttributes: []string{"x"},
				DeltaStage:                    durationpb.New(5 * time.Second),
				RequestTimeout:                durationpb.New(20 * time.Second),
				ServerMaxParallelRequests:     4,
				MinResponsesToAggregate:       2,
			},
		},
		MethodConfigs: map[string]*CapabilityMethodConfig{
			"vault.secrets.get": {
				RemoteConfig: &CapabilityMethodConfig_RemoteExecutableConfig{
					RemoteExecutableConfig: &RemoteExecutableConfig{
						RequestTimeout:          durationpb.New(11 * time.Second),
						MinResponsesToAggregate: 5,
					},
				},
				AggregatorConfig: &AggregatorConfig{AggregatorType: AggregatorType(1)},
			},
			"some.trigger": {
				RemoteConfig: &CapabilityMethodConfig_RemoteTriggerConfig{
					RemoteTriggerConfig: &RemoteTriggerConfig{
						MaxBatchSize: 9,
					},
				},
			},
		},
	}

	got, err := CapabilityConfigFromProto(in)
	require.NoError(t, err)

	require.NotNil(t, got.RemoteExecutableConfig)
	assert.Equal(t, []string{"x"}, got.RemoteExecutableConfig.RequestHashExcludedAttributes)
	assert.Equal(t, 5*time.Second, got.RemoteExecutableConfig.DeltaStage)
	assert.Equal(t, uint32(2), got.RemoteExecutableConfig.MinResponsesToAggregate)

	require.Len(t, got.CapabilityMethodConfig, 2)

	exec := got.CapabilityMethodConfig["vault.secrets.get"]
	require.NotNil(t, exec.RemoteExecutableConfig)
	assert.Equal(t, 11*time.Second, exec.RemoteExecutableConfig.RequestTimeout)
	assert.Equal(t, uint32(5), exec.RemoteExecutableConfig.MinResponsesToAggregate)
	require.NotNil(t, exec.AggregatorConfig)
	assert.Equal(t, capabilities.AggregatorType(1), exec.AggregatorConfig.AggregatorType)
	assert.Nil(t, exec.RemoteTriggerConfig)

	trig := got.CapabilityMethodConfig["some.trigger"]
	require.NotNil(t, trig.RemoteTriggerConfig)
	assert.Equal(t, uint32(9), trig.RemoteTriggerConfig.MaxBatchSize)
	assert.Nil(t, trig.RemoteExecutableConfig)
	assert.Nil(t, trig.AggregatorConfig)
}

func TestCapabilityConfigFromProto_OCR3AndOracleFactoryConfigs(t *testing.T) {
	oracleCfg, err := values.NewMap(map[string]any{"k": "v"})
	require.NoError(t, err)

	in := &CapabilityConfig{
		Ocr3Configs: map[string]*OCR3Config{
			"__default__": {
				ConfigCount:           4,
				Signers:               [][]byte{{0x01}, {0x02}},
				Transmitters:          [][]byte{{0xaa, 0xbb}},
				F:                     1,
				OnchainConfig:         []byte("onchain"),
				OffchainConfigVersion: 3,
				OffchainConfig:        []byte("offchain"),
			},
		},
		OracleFactoryConfigs: map[string]*valuespb.Map{
			"__default__": values.ProtoMap(oracleCfg),
		},
	}

	got, err := CapabilityConfigFromProto(in)
	require.NoError(t, err)

	require.Len(t, got.Ocr3Configs, 1)
	ocr := got.Ocr3Configs["__default__"]
	assert.Equal(t, uint64(4), ocr.ConfigCount)
	assert.Len(t, ocr.Signers, 2)
	// Transmitters are hex-encoded accounts, not raw bytes.
	require.Len(t, ocr.Transmitters, 1)
	assert.Equal(t, "aabb", string(ocr.Transmitters[0]))
	assert.Equal(t, uint8(1), ocr.F)
	assert.Equal(t, uint64(3), ocr.OffchainConfigVersion)
	// ConfigDigest is filled in later by ContractConfigTracker, not here.
	assert.Zero(t, ocr.ConfigDigest)

	require.Len(t, got.OracleFactoryConfigs, 1)
	assert.Equal(t, *oracleCfg, got.OracleFactoryConfigs["__default__"])
}

// RestrictedConfig is the config only we may set, and it takes precedence over anything a user
// supplies. Dropping it on decode would silently hand the user's own value that precedence, so it
// has to survive even though nothing about the message shape forces it to.
func TestCapabilityConfigFromProto_KeepsRestrictedConfig(t *testing.T) {
	defaultConfig, err := values.NewMap(map[string]any{"userMayChange": "yes"})
	require.NoError(t, err)
	restricted, err := values.NewMap(map[string]any{"onlyWeSetThis": "secret"})
	require.NoError(t, err)

	got, err := CapabilityConfigFromProto(&CapabilityConfig{
		DefaultConfig:    values.ProtoMap(defaultConfig),
		RestrictedConfig: values.ProtoMap(restricted),
		RestrictedKeys:   []string{"onlyWeSetThis"},
	})
	require.NoError(t, err)
	assert.Equal(t, defaultConfig, got.DefaultConfig)
	assert.Equal(t, restricted, got.RestrictedConfig)
	assert.Equal(t, []string{"onlyWeSetThis"}, got.RestrictedKeys)
}

// F is the fault tolerance OCR runs at, so a value that does not fit a uint8 has to be refused
// rather than truncated into a completely different protocol parameter.
func TestCapabilityConfigFromProto_RejectsOutOfRangeOCR3F(t *testing.T) {
	_, err := CapabilityConfigFromProto(&CapabilityConfig{
		Ocr3Configs: map[string]*OCR3Config{"__default__": {F: 256}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds uint8 max")
}

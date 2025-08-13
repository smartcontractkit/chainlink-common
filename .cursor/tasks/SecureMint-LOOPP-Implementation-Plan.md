# Secure Mint LOOPP Implementation Plan

## Overview

This document outlines the plan to convert the Secure Mint plugin from an in-process aggregator to a LOOPP (Local Out of Process Plugin) that follows the established patterns in chainlink-common.

## Current State Analysis

### Secure Mint Plugin (Target for LOOPP Conversion)
- **Location**: `/Users/ggerritsen/dev/cll/por_mock_ocr3plugin/por/`
- **Type**: OCR3 Reporting Plugin (production-ready, standalone)
- **Current State**: External plugin that needs to be converted to LOOPP
- **Registration**: Already registered as `SecureMint` in `pkg/types/plugin.go`

### External Plugin Architecture
- **Core Types**:
  - `ChainSelector uint64`: Chain identifier (matches chain-selectors package)
  - `PorReportingPluginFactory`: Main factory implementing `ocr3types.ReportingPluginFactory[ChainSelector]`
  - `porReportingPlugin`: Main plugin implementing `ocr3types.ReportingPlugin[ChainSelector]`
  - `PorReport`: Report structure with ConfigDigest, SeqNr, Block, Mintable
  - `ExternalAdapterPayload`: Contains Mintables, ReserveInfo, LatestBlocks

- **Key Files**:
  - `porplugin_simple.go`: Main plugin implementation
  - `types.go`: Core data structures and types
  - `external_adapter_interface.go`: External adapter interface
  - `contract_reader_interface.go`: Contract reading interface
  - `report_marshaller_interface.go`: Report serialization interface

- **Key Interfaces**:
  - `ExternalAdapter`: Provides mintable amounts and latest blocks per chain
  - `ContractReader`: Reads latest transmitted report details from contracts
  - `ReportMarshaler`: Serializes reports for transmission

- **Core Functionality**:
  - Processes observations containing mintable amounts and latest blocks
  - Validates observations and calculates honest blocks
  - Generates reports with mintable amounts for specific chains
  - Handles multi-chain support with configurable max chains
  - Uses external adapter for PoR (Proof of Reserve) calculations

### Downstream Components (Not in Scope)
- **Location**: `pkg/capabilities/consensus/ocr3/datafeeds/securemint_aggregator.go`
- **Type**: OCR3 Aggregator (in-process)
- **Functionality**: Processes Secure Mint reports, validates chain selectors and sequence numbers, packs data for DF Cache contract
- **Note**: This is a downstream aggregator that processes the output of the Secure Mint plugin, not the plugin itself

## Target Architecture

### LOOPP Structure
Following the established patterns from other plugins (Median, Mercury, CCIP), the external Secure Mint plugin will be converted to:

1. **Plugin Interface**: `types.PluginSecureMint`
2. **Plugin Factory**: `types.SecureMintPluginFactory`
3. **Service Wrapper**: `loop.SecureMintService`
4. **GRPC Plugin**: `loop.GRPCPluginSecureMint`
5. **Provider Interface**: `types.SecureMintProvider`

### Conversion Strategy
- **External Plugin Integration**: Import and adapt the existing `/por` package
- **Interface Mapping**: Convert external plugin interfaces to LOOPP interfaces
- **Relayer Integration**: Ensure all blockchain operations go through the Relayer interface
- **Preserve Logic**: Maintain the core Secure Mint functionality while adapting to LOOPP architecture

## Implementation Plan

### Phase 1: Define Core Interfaces

#### 1.1 Create SecureMint Provider Interface
**File**: `pkg/types/provider_securemint.go`

```go
package types

import (
    "context"
    "math/big"
    "time"
    "github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
    "github.com/smartcontractkit/chainlink-common/pkg/services"
)

// SecureMintProvider provides components needed for a SecureMint OCR3 plugin
type SecureMintProvider interface {
    PluginProvider
    
    // ExternalAdapter provides mintable amounts and latest blocks per chain
    ExternalAdapter() ExternalAdapter
    
    // ContractReader reads latest transmitted report details from contracts
    ContractReader() ContractReader
    
    // ReportMarshaler serializes reports for transmission
    ReportMarshaler() ReportMarshaler
}

// ExternalAdapter interface for PoR calculations
type ExternalAdapter interface {
    // GetPayload returns mintable amounts and latest blocks for queried blocks
    GetPayload(ctx context.Context, blocks map[uint64]uint64) (ExternalAdapterPayload, error)
}

// ExternalAdapterPayload contains mintable amounts, reserve info, and latest blocks
type ExternalAdapterPayload struct {
    Mintables   map[uint64]BlockMintablePair // ChainSelector -> BlockMintablePair
    ReserveInfo ReserveInfo
    LatestBlocks map[uint64]uint64 // ChainSelector -> BlockNumber
}

// BlockMintablePair contains block number and mintable amount
type BlockMintablePair struct {
    Block    uint64
    Mintable *big.Int
}

// ReserveInfo contains reserve amount and timestamp
type ReserveInfo struct {
    ReserveAmount *big.Int
    Timestamp     time.Time
}

// ContractReader interface for reading contract state
type ContractReader interface {
    // GetLatestTransmittedReportDetails retrieves latest transmission details
    GetLatestTransmittedReportDetails(ctx context.Context, chain uint64) (TransmittedReportDetails, error)
}

// TransmittedReportDetails contains transmission information
type TransmittedReportDetails struct {
    ConfigDigest    ocr2types.ConfigDigest
    SeqNr           uint64
    LatestTimestamp time.Time
}

// ReportMarshaler interface for report serialization
type ReportMarshaler interface {
    // Serialize serializes a report for a specific chain
    Serialize(ctx context.Context, chain uint64, report PorReport) ([]byte, error)
    
    // MaxReportSize returns maximum serialized report size
    MaxReportSize(ctx context.Context) int
}

// PorReport represents a Secure Mint report
type PorReport struct {
    ConfigDigest ocr2types.ConfigDigest
    SeqNr        uint64
    Block        uint64
    Mintable     *big.Int
}

// PluginSecureMint interface for the LOOPP plugin
type PluginSecureMint interface {
    services.Service
    NewSecureMintFactory(ctx context.Context, provider SecureMintProvider, config SecureMintConfig) (SecureMintPluginFactory, error)
}

// SecureMintPluginFactory interface
type SecureMintPluginFactory interface {
    Service
    ocr3types.ReportingPluginFactory
}

// SecureMintConfig holds configuration for the SecureMint plugin
// Configuration comes from offchainConfig, not job specification
type SecureMintConfig struct {
    MaxChains uint32 `json:"maxChains"` // Maximum number of chains to track
}

// Validate validates the SecureMintConfig
func (c SecureMintConfig) Validate() error {
    if c.MaxChains == 0 {
        return fmt.Errorf("maxChains must be greater than 0")
    }
    return nil
}
```

#### 1.2 Add Provider Type to Relayer Interface
**File**: `pkg/types/relayer.go`

Add to the `Relayer` interface:
```go
NewSecureMintProvider(ctx context.Context, rargs RelayArgs, pargs PluginArgs) (SecureMintProvider, error)
```

### Phase 2: Create LOOPP Infrastructure

#### 2.1 Create GRPC Plugin
**File**: `pkg/loop/plugin_securemint.go`

```go
package loop

import (
    "context"
    "github.com/hashicorp/go-plugin"
    "google.golang.org/grpc"
    "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/reportingplugin/securemint"
    "github.com/smartcontractkit/chainlink-common/pkg/types"
)

const PluginSecureMintName = "securemint"

func PluginSecureMintHandshakeConfig() plugin.HandshakeConfig {
    return plugin.HandshakeConfig{
        MagicCookieKey:   "CL_PLUGIN_SECUREMINT_MAGIC_COOKIE",
        MagicCookieValue: "secure-mint-magic-cookie-value", // Generate unique value
    }
}

type GRPCPluginSecureMint struct {
    plugin.NetRPCUnsupportedPlugin
    BrokerConfig
    PluginServer types.PluginSecureMint
    pluginClient *securemint.PluginSecureMintClient
}

// Implement GRPCServer, GRPCClient, and ClientConfig methods
```

#### 2.2 Create Service Wrapper
**File**: `pkg/loop/securemint_service.go`

```go
package loop

import (
    "context"
    "fmt"
    "os/exec"
    "github.com/smartcontractkit/chainlink-common/pkg/logger"
    "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
    "github.com/smartcontractkit/chainlink-common/pkg/services"
    "github.com/smartcontractkit/chainlink-common/pkg/types"
    ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

var _ ocrtypes.ReportingPluginFactory = (*SecureMintService)(nil)

type SecureMintService struct {
    goplugin.PluginService[*GRPCPluginSecureMint, types.ReportingPluginFactory]
}

func NewSecureMintService(
    lggr logger.Logger,
    grpcOpts GRPCOpts,
    cmd func() *exec.Cmd,
    provider types.SecureMintProvider,
    config types.SecureMintConfig,
) *SecureMintService {
    newService := func(ctx context.Context, instance any) (types.ReportingPluginFactory, services.HealthReporter, error) {
        plug, ok := instance.(types.PluginSecureMint)
        if !ok {
            return nil, nil, fmt.Errorf("expected PluginSecureMint but got %T", instance)
        }
        factory, err := plug.NewSecureMintFactory(ctx, provider, config)
        if err != nil {
            return nil, nil, err
        }
        return factory, plug, nil
    }
    stopCh := make(chan struct{})
    lggr = logger.Named(lggr, "SecureMintService")
    var ss SecureMintService
    broker := BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
    ss.Init(PluginSecureMintName, &GRPCPluginSecureMint{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
    return &ss
}

func (s *SecureMintService) NewReportingPlugin(ctx context.Context, config ocrtypes.ReportingPluginConfig) (ocrtypes.ReportingPlugin, ocrtypes.ReportingPluginInfo, error) {
    if err := s.WaitCtx(ctx); err != nil {
        return nil, ocrtypes.ReportingPluginInfo{}, err
    }
    return s.Service.NewReportingPlugin(ctx, config)
}
```

### Phase 3: Create Internal Reporting Plugin Infrastructure

#### 3.1 Create GRPC Client/Server
**File**: `pkg/loop/internal/reportingplugin/securemint/securemint.go`

```go
package securemint

import (
    "context"
    "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
    "github.com/smartcontractkit/chainlink-common/pkg/types"
    "google.golang.org/grpc"
)

type PluginSecureMintClient struct {
    net.BrokerExt
    pluginSecureMint types.PluginSecureMint
}

func NewPluginSecureMintClient(brokerCfg net.BrokerConfig) *PluginSecureMintClient {
    return &PluginSecureMintClient{BrokerExt: net.NewBrokerExt(brokerCfg)}
}

func RegisterPluginSecureMintServer(server *grpc.Server, broker net.Broker, brokerCfg net.BrokerConfig, impl types.PluginSecureMint) error {
    // Register the GRPC server
    return nil
}
```

#### 3.2 Create Protocol Buffers
**File**: `pkg/loop/internal/pb/securemint/securemint_plugin.proto`

```protobuf
syntax = "proto3";

option go_package = "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/securemint";

package securemint;

service PluginSecureMint {
  rpc NewSecureMintFactory (NewSecureMintFactoryRequest) returns (NewSecureMintFactoryReply) {}
}

message NewSecureMintFactoryRequest {
  uint32 providerID = 1;
  SecureMintConfig config = 2;
}

message NewSecureMintFactoryReply {
  uint32 factoryID = 1;
}

message SecureMintConfig {
  int64 targetChainSelector = 1;
}
```

### Phase 4: Update Relayer Integration

#### 4.1 Add Provider Type to Internal Types
**File**: `pkg/loop/internal/types/types.go`

Add:
```go
type SecureMintProvider interface {
    NewSecureMintProvider(context.Context, types.RelayArgs, types.PluginArgs) (types.SecureMintProvider, error)
}
```

#### 4.2 Update Relayer Server
**File**: `pkg/loop/internal/relayer/relayer.go`

Add to the switch statement in `NewPluginProvider`:
```go
case string(types.SecureMint):
    id, err := r.newSecureMintProvider(ctx, relayArgs, pluginArgs)
    if err != nil {
        return nil, err
    }
    return &pb.NewPluginProviderReply{PluginProviderID: id}, nil
```

Add the `newSecureMintProvider` method:
```go
func (r *relayerServer) newSecureMintProvider(ctx context.Context, relayArgs types.RelayArgs, pluginArgs types.PluginArgs) (uint32, error) {
    i, ok := r.impl.(looptypes.SecureMintProvider)
    if !ok {
        return 0, status.Error(codes.Unimplemented, "securemint not supported")
    }

    provider, err := i.NewSecureMintProvider(ctx, relayArgs, pluginArgs)
    if err != nil {
        return 0, err
    }
    err = provider.Start(ctx)
    if err != nil {
        return 0, err
    }
    const name = "SecureMintProvider"
    providerRes := net.Resource{Name: name, Closer: provider}

    id, _, err := r.ServeNew(name, func(s *grpc.Server) {
        securemint.RegisterProviderServices(s, provider)
    }, providerRes)
    if err != nil {
        return 0, err
    }

    return id, err
}
```

### Phase 5: External Plugin Integration

#### 5.1 Create External Plugin Binary
**File**: `cmd/securemint/main.go`

**Note**: This follows the established pattern from the [Chainlink repository](https://github.com/smartcontractkit/chainlink/blob/develop/core/services/ocr2/plugins/median/services.go#L151-L180) where:
- External plugins are imported directly (e.g., `chainlink-feeds/median`, `chainlink-data-streams/mercury`)
- LOOPP services wrap external plugins: `loop.NewMedianService(...)`
- Fallback to in-process: `median.NewPlugin(lggr).NewMedianFactory(...)`
- Configuration comes from offchainConfig, not job specification

```go
package main

import (
    "os"
    "github.com/hashicorp/go-plugin"
    "github.com/smartcontractkit/chainlink-common/pkg/logger"
    "github.com/smartcontractkit/chainlink-common/pkg/loop"
    "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/reportingplugin/securemint"
    "github.com/smartcontractkit/chainlink-common/pkg/types"
    por "github.com/smartcontractkit/por_mock_ocr3plugin/por"
)

func main() {
    lggr := logger.New(os.Stdout)
    
    // Create the plugin server implementation
    pluginServer := &SecureMintPluginServer{
        Logger: lggr,
    }
    
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: loop.PluginSecureMintHandshakeConfig(),
        Plugins: map[string]plugin.Plugin{
            loop.PluginSecureMintName: &loop.GRPCPluginSecureMint{
                PluginServer: pluginServer,
                BrokerConfig: loop.BrokerConfig{Logger: lggr},
            },
        },
        GRPCServer: loop.NewGRPCServer(),
    })
}

type SecureMintPluginServer struct {
    logger.Logger
}

func (s *SecureMintPluginServer) NewSecureMintFactory(ctx context.Context, provider types.SecureMintProvider, config types.SecureMintConfig) (types.SecureMintPluginFactory, error) {
    // Create external adapter implementation using Relayer
    externalAdapter := &RelayerExternalAdapter{
        provider: provider,
        logger:   s.Logger,
    }
    
    // Create contract reader implementation using Relayer
    contractReader := &RelayerContractReader{
        provider: provider,
        logger:   s.Logger,
    }
    
    // Create report marshaler implementation
    reportMarshaler := &ChainlinkReportMarshaler{
        logger: s.Logger,
    }
    
    // Create the external plugin factory using the imported por package
    porFactory := &por.PorReportingPluginFactory{
        Logger:          s.Logger,
        ExternalAdapter: externalAdapter,
        ContractReader:  contractReader,
        ReportMarshaler: reportMarshaler,
    }
    
    // Wrap the external factory in our LOOPP interface
    return &SecureMintFactory{
        porFactory: porFactory,
        config:     config,
        logger:     s.Logger,
    }, nil
}

// RelayerExternalAdapter implements por.ExternalAdapter using Relayer
type RelayerExternalAdapter struct {
    provider types.SecureMintProvider
    logger   logger.Logger
}

func (r *RelayerExternalAdapter) GetPayload(ctx context.Context, blocks por.Blocks) (por.ExternalAdapterPayload, error) {
    // Convert por.Blocks to our format and use Relayer
    // Implementation will use provider.ExternalAdapter() and Relayer contract reading
    return por.ExternalAdapterPayload{}, nil
}

// RelayerContractReader implements por.ContractReader using Relayer
type RelayerContractReader struct {
    provider types.SecureMintProvider
    logger   logger.Logger
}

func (r *RelayerContractReader) GetLatestTransmittedReportDetails(ctx context.Context, chain por.ChainSelector) (por.TransmittedReportDetails, error) {
    // Use Relayer to read contract state
    // Implementation will use provider.ContractReader() and Relayer
    return por.TransmittedReportDetails{}, nil
}

// ChainlinkReportMarshaler implements por.ReportMarshaler
type ChainlinkReportMarshaler struct {
    logger logger.Logger
}

func (c *ChainlinkReportMarshaler) Serialize(ctx context.Context, chain por.ChainSelector, report por.PorReport) ([]byte, error) {
    // Serialize report using chainlink-common utilities
    return nil, nil
}

func (c *ChainlinkReportMarshaler) MaxReportSize(ctx context.Context) int {
    // Return maximum report size
    return 1024
}
```

#### 5.2 Integrate External Plugin Logic
The external plugin integration will need to:

1. **Direct External Repository Import**: 
   - Import the external `por_mock_ocr3plugin` repository directly
   - Follow the same pattern as Median (`chainlink-feeds/median`) and Mercury (`chainlink-data-streams/mercury`)
   - Users will need the external repository as a dependency
2. **Adapter Pattern Implementation**: Based on established LOOPP patterns, create thin wrapper adapters:
   - `RelayerExternalAdapter`: Thin wrapper that delegates to Relayer interface for external adapter operations
   - `RelayerContractReader`: Thin wrapper that delegates to Relayer interface for contract reading
   - `ChainlinkReportMarshaler`: Thin wrapper that handles report serialization using chainlink-common utilities
3. **Type Conversion Strategy**: Follow established patterns for type conversions:
   - Convert between external plugin types (`por.ChainSelector`) and LOOPP types (`uint64`)
   - Use existing chain-selectors package for consistency
   - Maintain type safety through explicit conversions
4. **Handle chain interactions**: Ensure all blockchain interactions go through the Relayer interface:
   - Use Relayer for contract reading operations
   - Implement ExternalAdapter using Relayer's contract reading capabilities
   - Ensure all chain-specific operations use the Relayer interface
5. **Maintain functionality**: Preserve the existing Secure Mint logic:
   - Multi-chain support with configurable max chains (from offchainConfig)
   - Observation validation and honest block calculation
   - Report generation with mintable amounts
   - PoR (Proof of Reserve) calculations through external adapter
6. **Key Integration Points**:
   - `PorReportingPluginFactory` → `SecureMintPluginFactory`
   - `porReportingPlugin` → LOOPP plugin implementation
   - `ExternalAdapter` → Relayer-based implementation
   - `ContractReader` → Relayer contract reading
   - `ReportMarshaler` → Chainlink-common serialization
7. **Follow Established Pattern**: Based on the [Chainlink repository](https://github.com/smartcontractkit/chainlink/blob/develop/core/services/ocr2/plugins/median/services.go#L151-L180), the pattern is:
   - Import external plugin: `"github.com/smartcontractkit/por_mock_ocr3plugin/por"`
   - Use LOOPP service: `loop.NewSecureMintService(...)`
   - Fallback to in-process: `securemint.NewPlugin(lggr).NewSecureMintFactory(...)`

### Phase 6: Testing Infrastructure

#### 6.1 Create Test Files
Following established LOOPP testing patterns:
- `pkg/loop/securemint_service_test.go`: Test the LOOPP service wrapper
- `pkg/loop/internal/reportingplugin/securemint/securemint_test.go`: Test GRPC client/server
- `cmd/securemint/main_test.go`: Test the external plugin binary
- `pkg/loop/internal/test/cmd/main.go`: Add Secure Mint to test helper process

#### 6.2 Integration Tests
Based on established patterns from other LOOPP plugins:
- **LOOPP Lifecycle Tests**: Test plugin startup, shutdown, and recovery (following `TestLOOPPService` pattern)
- **Relayer Integration Tests**: Verify all chain interactions go through Relayer interface
- **Configuration Tests**: Test offchainConfig parsing and validation
- **Error Handling Tests**: Test error propagation and recovery scenarios
- **Mock Testing**: Use mock implementations for external plugin interfaces (following `nettest.MockConn` pattern)
- **Adapter Layer Tests**: Test the adapter layer separately from LOOPP infrastructure

### Phase 7: Documentation and Configuration

#### 7.1 Update Documentation
- Add Secure Mint to LOOPP documentation
- Document configuration options
- Add usage examples

#### 7.2 Feature Flag Handling
- Add feature flag for Secure Mint LOOPP
- Document feature flag usage
- Add comments for future maintainers

## Implementation Details and Reasoning

### Why LOOPP?
1. **Isolation**: Secure Mint plugin runs in separate process, improving stability
2. **Consistency**: Follows established patterns used by other plugins
3. **Maintainability**: Clear separation of concerns and interfaces
4. **Reliability**: Process isolation prevents plugin crashes from affecting the main node

### External Plugin Integration Strategy
- **Direct Import**: Import the external `por_mock_ocr3plugin` repository directly
- **Thin Adapter Wrappers**: Create thin wrapper adapters that delegate to Relayer interface (following established patterns)
- **Type Conversions**: Handle conversions between external plugin types and LOOPP types with explicit type safety
- **Relayer Integration**: All blockchain operations go through the Relayer interface via adapter implementations
- **Preserve Logic**: Maintain the core Secure Mint logic while adapting the interfaces
- **External Dependencies**: Users will need the external repository as a dependency (following established pattern)
- **Configuration Source**: Configuration comes from offchainConfig, not job specification

### Relayer Interface Integration
- **Chain Interactions**: All blockchain operations must go through the Relayer interface
- **Provider Pattern**: Uses the established provider pattern for chain-specific operations
- **Consistency**: Maintains consistency with other LOOPP implementations
- **Adapter Pattern**: Create `RelayerExternalAdapter` and `RelayerContractReader` to bridge external plugin with Relayer

### Configuration Management
- **Type Safety**: Strongly typed configuration structures
- **Validation**: Built-in configuration validation (following established patterns)
- **Flexibility**: Support for plugin-specific configuration options
- **Multi-Chain Support**: Configurable max chains parameter from offchainConfig
- **Source**: Configuration comes from offchainConfig, not job specification

### Error Handling
- **Graceful Degradation**: Plugin failures don't crash the main node
- **Health Reporting**: Comprehensive health reporting for monitoring
- **Logging**: Structured logging for debugging and monitoring
- **Error Propagation**: Follow established patterns for error propagation from external plugin to LOOPP layer
- **Error Wrapping**: Use standard error wrapping patterns without additional context unless established patterns require it
- **Recovery**: Automatic plugin recovery following established LOOPP patterns

## Risk Assessment

### Technical Risks
- **Interface Compatibility**: Ensuring external plugin integrates correctly
- **Performance Impact**: LOOPP overhead vs in-process execution
- **Configuration Complexity**: Managing plugin configuration

### Mitigation Strategies
- **Comprehensive Testing**: Extensive unit and integration tests
- **Feature Flags**: Ability to rollback to in-process implementation
- **Monitoring**: Enhanced monitoring and alerting for LOOPP version

## Success Criteria

1. **Functional Parity**: LOOPP version provides same functionality as external Secure Mint plugin
2. **Performance**: Acceptable performance compared to external plugin version
3. **Stability**: Improved stability through process isolation
4. **Maintainability**: Clear, well-documented code following established patterns
5. **Integration**: Seamless integration with existing Chainlink infrastructure
6. **Relayer Integration**: All blockchain operations go through the Relayer interface

## Timeline Estimate

- **Phase 1-2**: 1-2 weeks (Core interfaces and LOOPP infrastructure)
- **Phase 3-4**: 1-2 weeks (Internal infrastructure and relayer integration)
- **Phase 5**: 2-3 weeks (External plugin integration)
- **Phase 6**: 1 week (Testing)
- **Phase 7**: 1 week (Documentation and configuration)
- **Total**: 6-9 weeks

## Dependencies

1. **External Plugin Repository**: Direct import of `por_mock_ocr3plugin` repository (following pattern of `chainlink-feeds/median` and `chainlink-data-streams/mercury`)
2. **Protocol Buffers**: Generated protobuf files for GRPC communication
3. **Testing Infrastructure**: Test utilities and mock implementations
4. **Documentation**: Existing LOOPP documentation for reference
5. **Chain-Selectors**: Integration with chain-selectors package for ChainSelector type
6. **Relayer Implementation**: EVM Relayer implementation for blockchain interactions

## Next Steps

1. **Approval**: Get approval for this implementation plan
2. **Repository Access**: Ensure access to external Secure Mint plugin
3. **Development Environment**: Set up development environment with all dependencies
4. **Implementation**: Begin implementation following the outlined phases
5. **Testing**: Continuous testing throughout implementation
6. **Review**: Regular code reviews and architecture reviews 
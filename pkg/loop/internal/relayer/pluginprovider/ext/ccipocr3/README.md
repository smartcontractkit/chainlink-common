# CCIP OCR3 Provider gRPC Implementation

This package implements the gRPC-based adapter for the CCIPProvider interface, following the pattern established by the median provider implementation.

## Overview

The CCIPProvider interface requires three main components:
1. **ChainAccessor** - For all direct chain access operations
2. **ContractTransmitter** - For OCR3 contract transmission (reuses existing OCR3 implementation)
3. **Codec** - For encoding/decoding various types of CCIP data

## Files

- `ccip_provider.go` - Main CCIPProvider client/server implementation
- `chainaccessor.go` - ChainAccessor gRPC client/server
- `codec.go` - Codec interfaces gRPC client/server implementations
- `helpers.go` - Helper functions and type conversions
- `chainaccessor.proto` - Protocol buffer definitions for ChainAccessor
- `codec.proto` - Protocol buffer definitions for Codec interfaces

## Architecture

The implementation follows the established gRPC adapter pattern:

1. **Client Side**: Wraps gRPC clients to implement Go interfaces
2. **Server Side**: Wraps Go interface implementations to serve gRPC requests  
3. **Protocol Buffers**: Define the wire format for all method calls
4. **Type Conversion**: Convert between Go types and protobuf types

## Components

### ChainAccessor
Implements the complete ChainAccessor interface with methods for:
- **AllAccessors**: Common functionality (GetContractAddress, GetAllConfigsLegacy, etc.)
- **DestinationAccessor**: Destination chain operations (CommitReports, ExecutedMessages, etc.)
- **SourceAccessor**: Source chain operations (MsgsBetweenSeqNums, TokenPriceUSD, etc.)
- **RMNAccessor**: RMN operations (GetRMNCurseInfo)

### Codec
Implements all codec interfaces:
- **ChainSpecificAddressCodec**: Address encoding/decoding
- **CommitPluginCodec**: Commit plugin report encoding/decoding (reuses existing)
- **ExecutePluginCodec**: Execute plugin report encoding/decoding
- **TokenDataEncoder**: Token data encoding (USDC/CCTP)
- **SourceChainExtraDataCodec**: Source chain extra data decoding

### ContractTransmitter
Reuses the existing OCR3 ContractTransmitter gRPC implementation from `pkg/loop/internal/relayer/pluginprovider/ocr3/`.

## Usage

```go
// Server side - register all services
func RegisterProviderServices(s *grpc.Server, provider types.CCIPProvider) {
    ccipocr3.RegisterProviderServices(s, provider)
}

// Client side - create provider client
func ConnToProvider(conn grpc.ClientConnInterface, broker net.Broker, brokerCfg net.BrokerConfig) types.CCIPProvider {
    be := &net.BrokerExt{Broker: broker, BrokerConfig: brokerCfg}
    return ccipocr3.NewProviderClient(be, conn)
}
```

## Development Notes

- Type conversions between Go and protobuf types are handled in helper functions
- Complex types like maps and nested structures require careful conversion
- Error handling follows gRPC best practices
- The implementation is designed to be compatible with the existing LOOP infrastructure

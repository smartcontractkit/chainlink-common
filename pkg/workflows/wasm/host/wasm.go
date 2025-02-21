package host

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/andybalholm/brotli"
	"github.com/google/uuid"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

func GetWorkflowSpec(ctx context.Context, modCfg *ModuleConfig, config []byte,
	wasmtimeModuleFactory WasmtimeModuleFactoryFn) (*sdk.WorkflowSpec, error) {
	m, err := NewModule(modCfg, wasmtimeModuleFactory, WithDeterminism())
	if err != nil {
		return nil, fmt.Errorf("could not instantiate module: %w", err)
	}

	m.Start()

	rid := uuid.New().String()
	req := &wasmpb.Request{
		Id:     rid,
		Config: config,
		Message: &wasmpb.Request_SpecRequest{
			SpecRequest: &emptypb.Empty{},
		},
	}
	resp, err := m.Run(ctx, req)
	if err != nil {
		return nil, err
	}

	sr := resp.GetSpecResponse()
	if sr == nil {
		return nil, errors.New("unexpected response from WASM binary: got nil spec response")
	}

	m.Close()

	return wasmpb.ProtoToWorkflowSpec(sr)
}

func ValidateAndDecompressBinary(binary []byte, isUncompressed bool, maxCompressedBinarySize uint64, maxDecompressedBinarySize uint64) ([]byte, error) {

	if maxCompressedBinarySize == 0 {
		maxCompressedBinarySize = uint64(defaultMaxCompressedBinarySize)
	}

	if maxDecompressedBinarySize == 0 {
		maxDecompressedBinarySize = uint64(defaultMaxDecompressedBinarySize)
	}

	if !isUncompressed {
		// validate the binary size before decompressing
		// this is to prevent decompression bombs
		if uint64(len(binary)) > maxCompressedBinarySize {
			return nil, fmt.Errorf("compressed binary size exceeds the maximum allowed size of %d bytes", maxCompressedBinarySize)
		}

		rdr := io.LimitReader(brotli.NewReader(bytes.NewBuffer(binary)), int64(maxDecompressedBinarySize+1))
		decompedBinary, err := io.ReadAll(rdr)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress binary: %w", err)
		}

		binary = decompedBinary
	}

	// Validate the decompressed binary size.
	// io.LimitReader prevents decompression bombs by reading up to a set limit, but it will not return an error if the limit is reached.
	// The Read() method will return io.EOF, and ReadAll will gracefully handle it and return nil.
	if uint64(len(binary)) > maxDecompressedBinarySize {
		return nil, fmt.Errorf("decompressed binary size reached the maximum allowed size of %d bytes", maxDecompressedBinarySize)
	}
	return binary, nil
}

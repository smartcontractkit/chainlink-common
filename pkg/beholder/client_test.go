package beholder_test

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/internal/mocks"
)

type MockExporter struct {
	mock.Mock
	sdklog.Exporter
}

func (m *MockExporter) Export(ctx context.Context, records []sdklog.Record) error {
	args := m.Called(ctx, records)
	return args.Error(0)
}

func (m *MockExporter) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockExporter) ForceFlush(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestClient(t *testing.T) {
	defaultCustomAttributes := func() map[string]any {
		return map[string]any{
			"int_key_1":                123,
			"int64_key_1":              int64(123),
			"int32_key_1":              int32(123),
			"str_key_1":                "str_val_1",
			"bool_key_1":               true,
			"float_key_1":              123.456,
			"byte_key_1":               []byte("byte_val_1"),
			"str_slice_key_1":          []string{"str_val_1", "str_val_2"},
			"nil_key_1":                nil,
			beholder.AttrKeyDomain:     "TestDomain",        // Required field
			beholder.AttrKeyEntity:     "TestEntity",        // Required field
			beholder.AttrKeyDataSchema: "/schemas/ids/1001", // Required field, URI
		}
	}
	defaultMessageBody := []byte("body bytes")

	mustNewGRPCClient := func(t *testing.T, exporterMock *mocks.OTLPExporter) *beholder.Client {
		// Override exporter factory which is used by Client
		exporterFactory := func(...otlploggrpc.Option) (sdklog.Exporter, error) {
			return exporterMock, nil
		}
		client, err := beholder.NewGRPCClient(beholder.TestDefaultConfig(), exporterFactory)
		if err != nil {
			t.Fatalf("Error creating beholder client: %v", err)
		}
		return client
	}

	mustNewHTTPClient := func(t *testing.T, exporterMock *mocks.OTLPExporter) *beholder.Client {
		// Override exporter factory which is used by Client
		exporterFactory := func(...otlploghttp.Option) (sdklog.Exporter, error) {
			return exporterMock, nil
		}
		client, err := beholder.NewHTTPClient(beholder.TestDefaultConfigHTTPClient(), exporterFactory)
		if err != nil {
			t.Fatalf("Error creating beholder client: %v", err)
		}
		return client
	}

	testCases := []struct {
		name                   string
		makeCustomAttributes   func() map[string]any
		messageBody            []byte
		messageCount           int
		exporterMockErrorCount int
		exporterOutputExpected bool
		messageGenerator       func(t *testing.T, client *beholder.Client, messageBody []byte, customAttributes map[string]any)
		mustNewGrpcClient      func(*testing.T, *mocks.OTLPExporter) *beholder.Client
	}{
		{
			name:                   "Test Emit (GRPC Client)",
			makeCustomAttributes:   defaultCustomAttributes,
			messageBody:            defaultMessageBody,
			messageCount:           10,
			exporterMockErrorCount: 0,
			exporterOutputExpected: true,
			messageGenerator: func(t *testing.T, client *beholder.Client, messageBody []byte, customAttributes map[string]any) {
				err := client.Emitter.Emit(t.Context(), messageBody, customAttributes)
				assert.NoError(t, err)
			},
			mustNewGrpcClient: mustNewGRPCClient,
		},

		{
			name:                   "Test Emit (HTTP Client)",
			makeCustomAttributes:   defaultCustomAttributes,
			messageBody:            defaultMessageBody,
			messageCount:           10,
			exporterMockErrorCount: 0,
			exporterOutputExpected: true,
			messageGenerator: func(t *testing.T, client *beholder.Client, messageBody []byte, customAttributes map[string]any) {
				err := client.Emitter.Emit(t.Context(), messageBody, customAttributes)
				assert.NoError(t, err)
			},
			mustNewGrpcClient: mustNewHTTPClient,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exporterMock := mocks.NewOTLPExporter(t)
			defer exporterMock.AssertExpectations(t)

			client := tc.mustNewGrpcClient(t, exporterMock)

			// Set error handler and ensure it's reset after test
			originalHandler := otel.GetErrorHandler()
			otel.SetErrorHandler(otelMustNotErr(t))
			defer otel.SetErrorHandler(originalHandler)
			// Number of exported messages
			exportedMessageCount := 0

			// Simulate exporter error if configured
			if tc.exporterMockErrorCount > 0 {
				exporterMock.On("Export", mock.Anything, mock.Anything).Return(errors.New("an error occurred")).Times(tc.exporterMockErrorCount)
			}
			customAttributes := tc.makeCustomAttributes()
			if tc.exporterOutputExpected {
				exporterMock.On("Export", mock.Anything, mock.Anything).Return(nil).Times(tc.messageCount).
					Run(func(args mock.Arguments) {
						assert.IsType(t, []sdklog.Record{}, args.Get(1), "Record type mismatch")
						records := args.Get(1).([]sdklog.Record)
						assert.Len(t, records, 1, "batching is disabled, expecte 1 record")
						record := records[0]
						assert.Equal(t, tc.messageBody, record.Body().AsBytes(), "Record body mismatch")
						actualAttributeKeys := map[string]struct{}{}
						record.WalkAttributes(func(kv otellog.KeyValue) bool {
							key := kv.Key
							actualAttributeKeys[key] = struct{}{}
							expectedValue, ok := customAttributes[key]
							if !ok {
								t.Fatalf("Record attribute key not found: %s", key)
							}
							expectedKv := beholder.OtelAttr(key, expectedValue)
							equal := kv.Value.Equal(expectedKv.Value)
							assert.True(t, equal, "Record attributes mismatch for key %v", key)
							return true
						})
						for key := range customAttributes {
							if _, ok := actualAttributeKeys[key]; !ok {
								t.Fatalf("Record attribute key not found: %s", key)
							}
						}
						exportedMessageCount += len(records)
					})
			}
			for i := 0; i < tc.messageCount; i++ {
				tc.messageGenerator(t, client, tc.messageBody, customAttributes)
			}
			assert.Equal(t, tc.messageCount, exportedMessageCount, "Expect all emitted messages to be exported")
		})
	}
}

func TestClient_Close(t *testing.T) {
	exporterMock := mocks.NewOTLPExporter(t)
	defer exporterMock.AssertExpectations(t)

	client, err := beholder.NewStdoutClient()
	require.NoError(t, err)

	err = client.Close()
	require.NoError(t, err)

	exporterMock.AssertExpectations(t)
}

func TestClient_ForPackage(t *testing.T) {
	exporterMock := mocks.NewOTLPExporter(t)
	defer exporterMock.AssertExpectations(t)
	var b strings.Builder
	client, err := beholder.NewWriterClient(&b)
	require.NoError(t, err)
	clientForTest := client.ForPackage("TestClient_ForPackage")

	// Log
	clientForTest.Logger.Emit(t.Context(), otellog.Record{})
	assert.Contains(t, b.String(), `"Name":"TestClient_ForPackage"`)
	b.Reset()

	// Trace
	_, span := clientForTest.Tracer.Start(t.Context(), "testSpan")
	span.End()
	assert.Contains(t, b.String(), `"Name":"TestClient_ForPackage"`)
	assert.Contains(t, b.String(), "testSpan")
	b.Reset()

	// Meter
	counter, _ := clientForTest.Meter.Int64Counter("testMetric")
	counter.Add(t.Context(), 1)
	clientForTest.Close()
	assert.Contains(t, b.String(), `"Name":"TestClient_ForPackage"`)
	assert.Contains(t, b.String(), "testMetric")
}

func otelMustNotErr(t *testing.T) otel.ErrorHandlerFunc {
	// Create a context that will be canceled when the test completes
	ctx, cancel := context.WithCancel(context.Background())

	// Use t.Cleanup to cancel the context when test completes
	t.Cleanup(cancel)

	return func(err error) {
		// Check if test context is still valid
		select {
		case <-ctx.Done():
			// Test has completed, just log the error instead of failing
			fmt.Printf("otel error after test completion: %v\n", err)
			return
		default:
			// Test is still active, safe to call t.Fatalf
			t.Fatalf("otel error: %v", err)
		}
	}
}

func TestNewClient(t *testing.T) {
	t.Run("both endpoints set", func(t *testing.T) {
		client, err := beholder.NewClient(beholder.Config{
			OtelExporterGRPCEndpoint: "grpc-endpoint",
			OtelExporterHTTPEndpoint: "http-endpoint",
		})
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Equal(t, "only one exporter endpoint should be set", err.Error())
	})

	t.Run("no endpoints set", func(t *testing.T) {
		client, err := beholder.NewClient(beholder.Config{})
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Equal(t, "at least one exporter endpoint should be set", err.Error())
	})

	t.Run("GRPC endpoint set", func(t *testing.T) {
		client, err := beholder.NewClient(beholder.Config{
			OtelExporterGRPCEndpoint: "grpc-endpoint",
		})
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &beholder.Client{}, client)
	})

	t.Run("HTTP endpoint set", func(t *testing.T) {
		client, err := beholder.NewClient(beholder.Config{
			OtelExporterHTTPEndpoint: "http-endpoint",
		})
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &beholder.Client{}, client)
	})

	t.Run("emitter is dual source when ChipIngress is enabled", func(t *testing.T) {
		client, err := beholder.NewClient(beholder.Config{
			OtelExporterGRPCEndpoint:       "grpc-endpoint",
			ChipIngressEmitterEnabled:      true,
			ChipIngressEmitterGRPCEndpoint: "chip-ingress-endpoint:9090",
		})
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &beholder.DualSourceEmitter{}, client.Emitter)
	})

	t.Run("errors when ChipIngress is enabled but no endpoint is set", func(t *testing.T) {
		client, err := beholder.NewClient(beholder.Config{
			OtelExporterGRPCEndpoint:  "grpc-endpoint",
			ChipIngressEmitterEnabled: true,
		})
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Equal(t, "invalid address format: missing port in address", err.Error())
	})
}

func TestNewClientWithChipIngressConfig(t *testing.T) {
	t.Run("creates client with ChipIngress TLS endpoint", func(t *testing.T) {
		client, err := beholder.NewClient(beholder.Config{
			OtelExporterGRPCEndpoint:       "grpc-endpoint",
			ChipIngressEmitterEnabled:      true,
			ChipIngressEmitterGRPCEndpoint: "chip-ingress.example.com:9090",
			ChipIngressInsecureConnection:  false,
		})
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &beholder.DualSourceEmitter{}, client.Emitter)
	})

	t.Run("LogStreamingEnabled true creates logger", func(t *testing.T) {
		cfg := beholder.Config{
			OtelExporterGRPCEndpoint: "grpc-endpoint",
			LogStreamingEnabled:      true,
		}
		client, err := beholder.NewClient(cfg)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.LoggerProvider)
		assert.NotNil(t, client.Logger)
	})

	t.Run("LogStreamingEnabled false disables logger", func(t *testing.T) {
		cfg := beholder.Config{
			OtelExporterGRPCEndpoint: "grpc-endpoint",
			LogStreamingEnabled:      false,
		}
		client, err := beholder.NewClient(cfg)
		require.NoError(t, err)
		// LoggerProvider and Logger should NOT be nil, but should be no-op implementations
		assert.NotNil(t, client.LoggerProvider)
		assert.NotNil(t, client.Logger)

		// Optionally, check that using the logger does not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Logger panicked when LogStreamingEnabled is false: %v", r)
			}
		}()
		client.Logger.Emit(t.Context(), otellog.Record{})
	})

	t.Run("creates client with ChipIngress insecure endpoint", func(t *testing.T) {
		client, err := beholder.NewClient(beholder.Config{
			OtelExporterGRPCEndpoint:       "grpc-endpoint",
			ChipIngressEmitterEnabled:      true,
			ChipIngressEmitterGRPCEndpoint: "chip-ingress.example.com:9090",
			ChipIngressInsecureConnection:  true,
		})
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &beholder.DualSourceEmitter{}, client.Emitter)
	})

	t.Run("creates client with IPv4 ChipIngress endpoint", func(t *testing.T) {
		client, err := beholder.NewClient(beholder.Config{
			OtelExporterGRPCEndpoint:       "grpc-endpoint",
			ChipIngressEmitterEnabled:      true,
			ChipIngressEmitterGRPCEndpoint: "192.168.1.100:9090",
			ChipIngressInsecureConnection:  true,
		})
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &beholder.DualSourceEmitter{}, client.Emitter)
	})

	t.Run("creates client with IPv6 ChipIngress endpoint", func(t *testing.T) {
		client, err := beholder.NewClient(beholder.Config{
			OtelExporterGRPCEndpoint:       "grpc-endpoint",
			ChipIngressEmitterEnabled:      true,
			ChipIngressEmitterGRPCEndpoint: "[::1]:9090",
			ChipIngressInsecureConnection:  true,
		})
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &beholder.DualSourceEmitter{}, client.Emitter)
	})
}

// Update the existing function name to match its actual purpose
func TestNewClientWithInvalidChipIngressConfig(t *testing.T) {
	t.Run("errors with ChipIngress endpoint without port", func(t *testing.T) {
		client, err := beholder.NewClient(beholder.Config{
			OtelExporterGRPCEndpoint:       "grpc-endpoint",
			ChipIngressEmitterEnabled:      true,
			ChipIngressEmitterGRPCEndpoint: "chip-ingress.example.com",
			ChipIngressInsecureConnection:  false,
		})
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "missing port")
	})

	t.Run("errors with malformed ChipIngress endpoint", func(t *testing.T) {
		client, err := beholder.NewClient(beholder.Config{
			OtelExporterGRPCEndpoint:       "grpc-endpoint",
			ChipIngressEmitterEnabled:      true,
			ChipIngressEmitterGRPCEndpoint: "chip-ingress.example.com:invalid:port",
			ChipIngressInsecureConnection:  false,
		})
		require.Error(t, err)
		assert.Nil(t, client)
	})

	t.Run("errors when ChipIngress enabled with empty endpoint", func(t *testing.T) {
		client, err := beholder.NewClient(beholder.Config{
			OtelExporterGRPCEndpoint:       "grpc-endpoint",
			ChipIngressEmitterEnabled:      true,
			ChipIngressEmitterGRPCEndpoint: "",
		})
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "invalid address format: missing port in address")
	})

	t.Run("errors when ChipIngress enabled with whitespace-only endpoint", func(t *testing.T) {
		client, err := beholder.NewClient(beholder.Config{
			OtelExporterGRPCEndpoint:       "grpc-endpoint",
			ChipIngressEmitterEnabled:      true,
			ChipIngressEmitterGRPCEndpoint: "   ",
		})
		require.Error(t, err)
		assert.Nil(t, client)
		// The whitespace is preserved in the address, so the error includes the spaces
		assert.Contains(t, err.Error(), "invalid address format: address")
		assert.Contains(t, err.Error(), "missing port in address")
	})
}

func TestNewGRPCClient_ChipIngressEmitter(t *testing.T) {
	t.Run("chip ingress emitter enabled with insecure connection and auth headers", func(t *testing.T) {
		cfg := beholder.Config{
			OtelExporterGRPCEndpoint:       "localhost:4317",
			ChipIngressEmitterEnabled:      true,
			ChipIngressEmitterGRPCEndpoint: "localhost:8080",
			ChipIngressInsecureConnection:  true,
			AuthHeaders: map[string]string{
				"Authorization": "Bearer my-secret-token",
			},
		}

		// Mock the otlploggrpc.New function to avoid creating a real exporter
		otlploggrpcNew := func(options ...otlploggrpc.Option) (sdklog.Exporter, error) {
			return &mockLogExporter{}, nil
		}

		client, err := beholder.NewGRPCClient(cfg, otlploggrpcNew)
		require.NoError(t, err)
		require.NotNil(t, client)

		assert.NotNil(t, client.Emitter)
		// Check that the emitter is a dualSourceEmitter
		_, ok := client.Emitter.(*beholder.DualSourceEmitter)
		assert.True(t, ok, "Expected Emitter to be a DualSourceEmitter")
	})

	t.Run("chip ingress emitter enabled with tls connection", func(t *testing.T) {
		cfg := beholder.Config{
			OtelExporterGRPCEndpoint:       "localhost:4317",
			ChipIngressEmitterEnabled:      true,
			ChipIngressEmitterGRPCEndpoint: "localhost:8080",
			ChipIngressInsecureConnection:  false, // Use TLS
		}

		// Mock the otlploggrpc.New function
		otlploggrpcNew := func(options ...otlploggrpc.Option) (sdklog.Exporter, error) {
			return &mockLogExporter{}, nil
		}

		client, err := beholder.NewGRPCClient(cfg, otlploggrpcNew)
		require.NoError(t, err)
		require.NotNil(t, client)
		assert.NotNil(t, client.Emitter)
	})
}

func TestNewClient_Chip(t *testing.T) {
	t.Run("chip interface available with chip-ingress endpoint provided", func(t *testing.T) {
		client, err := beholder.NewClient(beholder.Config{
			OtelExporterGRPCEndpoint:       "grpc-endpoint",
			ChipIngressEmitterEnabled:      true,
			ChipIngressEmitterGRPCEndpoint: "chip-ingress.example.com:9090",
			ChipIngressInsecureConnection:  false,
		})
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.Chip)
	})

	t.Run("chip interface can be enabled when chip ingress dual emitter is not enabled ", func(t *testing.T) {
		client, err := beholder.NewClient(beholder.Config{
			OtelExporterGRPCEndpoint:       "grpc-endpoint",
			ChipIngressEmitterEnabled:      false,
			ChipIngressEmitterGRPCEndpoint: "chip-ingress.example.com:9090",
			ChipIngressInsecureConnection:  false,
		})
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.Chip)
		assert.NotNil(t, client.Emitter)
	})

	t.Run("chip interface is nil when chip ingress config is missing", func(t *testing.T) {
		client, err := beholder.NewClient(beholder.Config{
			OtelExporterGRPCEndpoint:  "grpc-endpoint",
			ChipIngressEmitterEnabled: true,
		})
		require.Error(t, err)
		assert.Nil(t, client)
	})
}

// mockLogExporter is a no-op exporter for testing purposes.
type mockLogExporter struct{}

func (m *mockLogExporter) Export(ctx context.Context, logs []sdklog.Record) error {
	return nil
}

func (m *mockLogExporter) Shutdown(ctx context.Context) error {
	return nil
}

func (m *mockLogExporter) ForceFlush(ctx context.Context) error {
	return nil
}

func TestNewGRPCClientRotatingAuth(t *testing.T) {

	pubKey, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	pubKeyHex := hex.EncodeToString(pubKey)

	t.Run("successful rotating auth setup", func(t *testing.T) {

		mockSigner := &MockSigner{}

		cfg := beholder.Config{
			OtelExporterGRPCEndpoint: "localhost:4317",
			AuthPublicKeyHex:         pubKeyHex,
			AuthKeySigner:            mockSigner,
			AuthHeadersTTL:           10 * time.Minute,
			InsecureConnection:       true,
		}

		// Mock the otlploggrpc.New function
		otlploggrpcNew := func(options ...otlploggrpc.Option) (sdklog.Exporter, error) {
			return &mockLogExporter{}, nil
		}

		client, err := beholder.NewGRPCClient(cfg, otlploggrpcNew)
		require.NoError(t, err)
		require.NotNil(t, client)
	})

	t.Run("error when public key hex is empty but signer is set", func(t *testing.T) {

		mockSigner := &MockSigner{}

		cfg := beholder.Config{
			OtelExporterGRPCEndpoint: "localhost:4317",
			AuthPublicKeyHex:         "", // Empty public key hex
			AuthKeySigner:            mockSigner,
			AuthHeadersTTL:           10 * time.Minute,
			InsecureConnection:       true,
		}

		otlploggrpcNew := func(options ...otlploggrpc.Option) (sdklog.Exporter, error) {
			return &mockLogExporter{}, nil
		}

		client, err := beholder.NewGRPCClient(cfg, otlploggrpcNew)
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "auth: public key hex required when signer is set")
	})

	t.Run("error when TTL is too short", func(t *testing.T) {

		mockSigner := &MockSigner{}

		cfg := beholder.Config{
			OtelExporterGRPCEndpoint: "localhost:4317",
			AuthPublicKeyHex:         pubKeyHex,
			AuthKeySigner:            mockSigner,
			AuthHeadersTTL:           5 * time.Minute,
			InsecureConnection:       true,
		}

		otlploggrpcNew := func(options ...otlploggrpc.Option) (sdklog.Exporter, error) {
			return &mockLogExporter{}, nil
		}

		client, err := beholder.NewGRPCClient(cfg, otlploggrpcNew)
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "auth: headers TTL must be at least 10 minutes")
	})

	t.Run("error when public key hex is invalid", func(t *testing.T) {

		mockSigner := &MockSigner{}

		cfg := beholder.Config{
			OtelExporterGRPCEndpoint: "localhost:4317",
			AuthPublicKeyHex:         "invalid-hex", // Invalid hex
			AuthKeySigner:            mockSigner,
			AuthHeadersTTL:           10 * time.Minute,
			InsecureConnection:       true,
		}

		otlploggrpcNew := func(options ...otlploggrpc.Option) (sdklog.Exporter, error) {
			return &mockLogExporter{}, nil
		}

		client, err := beholder.NewGRPCClient(cfg, otlploggrpcNew)
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "auth: failed to decode public key hex")
	})
}

func TestNewGRPCClientStaticAuthFallback(t *testing.T) {
	t.Run("uses static auth when no rotating auth is configured", func(t *testing.T) {

		cfg := beholder.Config{
			OtelExporterGRPCEndpoint: "localhost:4317",
			AuthHeaders: map[string]string{
				"Authorization": "Bearer test-token",
			},
			InsecureConnection: true,
		}

		otlploggrpcNew := func(options ...otlploggrpc.Option) (sdklog.Exporter, error) {
			return &mockLogExporter{}, nil
		}

		client, err := beholder.NewGRPCClient(cfg, otlploggrpcNew)
		require.NoError(t, err)
		require.NotNil(t, client)
	})

	t.Run("no auth when neither rotating nor static auth is configured", func(t *testing.T) {

		cfg := beholder.Config{
			OtelExporterGRPCEndpoint: "localhost:4317",
			InsecureConnection:       true,
		}

		otlploggrpcNew := func(options ...otlploggrpc.Option) (sdklog.Exporter, error) {
			return &mockLogExporter{}, nil
		}

		client, err := beholder.NewGRPCClient(cfg, otlploggrpcNew)
		require.NoError(t, err)
		require.NotNil(t, client)
	})
}

func TestNewGRPCClientChipIngressAuth(t *testing.T) {

	pubKey, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	pubKeyHex := hex.EncodeToString(pubKey)

	t.Run("chip ingress with rotating auth", func(t *testing.T) {

		mockSigner := &MockSigner{}

		cfg := beholder.Config{
			OtelExporterGRPCEndpoint:       "localhost:4317",
			ChipIngressEmitterEnabled:      true,
			ChipIngressEmitterGRPCEndpoint: "localhost:8080",
			ChipIngressInsecureConnection:  true,
			AuthPublicKeyHex:               pubKeyHex,
			AuthKeySigner:                  mockSigner,
			AuthHeadersTTL:                 10 * time.Minute,
			InsecureConnection:             true,
		}

		otlploggrpcNew := func(options ...otlploggrpc.Option) (sdklog.Exporter, error) {
			return &mockLogExporter{}, nil
		}

		client, err := beholder.NewGRPCClient(cfg, otlploggrpcNew)
		require.NoError(t, err)
		require.NotNil(t, client)

		_, ok := client.Emitter.(*beholder.DualSourceEmitter)
		assert.True(t, ok, "Expected Emitter to be a DualSourceEmitter")
	})

	t.Run("chip ingress with static auth", func(t *testing.T) {
		cfg := beholder.Config{
			OtelExporterGRPCEndpoint:       "localhost:4317",
			ChipIngressEmitterEnabled:      true,
			ChipIngressEmitterGRPCEndpoint: "localhost:8080",
			ChipIngressInsecureConnection:  true,
			AuthHeaders: map[string]string{
				"Authorization": "Bearer test-token",
			},
			InsecureConnection: true,
		}

		otlploggrpcNew := func(options ...otlploggrpc.Option) (sdklog.Exporter, error) {
			return &mockLogExporter{}, nil
		}

		client, err := beholder.NewGRPCClient(cfg, otlploggrpcNew)
		require.NoError(t, err)
		require.NotNil(t, client)

		_, ok := client.Emitter.(*beholder.DualSourceEmitter)
		assert.True(t, ok, "Expected Emitter to be a DualSourceEmitter")
	})

	t.Run("chip ingress with no auth", func(t *testing.T) {

		cfg := beholder.Config{
			OtelExporterGRPCEndpoint:       "localhost:4317",
			ChipIngressEmitterEnabled:      true,
			ChipIngressEmitterGRPCEndpoint: "localhost:8080",
			ChipIngressInsecureConnection:  true,
			InsecureConnection:             true,
		}

		otlploggrpcNew := func(options ...otlploggrpc.Option) (sdklog.Exporter, error) {
			return &mockLogExporter{}, nil
		}

		client, err := beholder.NewGRPCClient(cfg, otlploggrpcNew)
		require.NoError(t, err)
		require.NotNil(t, client)

		_, ok := client.Emitter.(*beholder.DualSourceEmitter)
		assert.True(t, ok, "Expected Emitter to be a DualSourceEmitter")
	})
}

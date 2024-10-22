package beholder

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/internal/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
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
			"int_key_1":            123,
			"int64_key_1":          int64(123),
			"int32_key_1":          int32(123),
			"str_key_1":            "str_val_1",
			"bool_key_1":           true,
			"float_key_1":          123.456,
			"byte_key_1":           []byte("byte_val_1"),
			"str_slice_key_1":      []string{"str_val_1", "str_val_2"},
			"nil_key_1":            nil,
			"beholder_domain":      "TestDomain",        // Required field
			"beholder_entity":      "TestEntity",        // Required field
			"beholder_data_schema": "/schemas/ids/1001", // Required field, URI
      "node_csa_key":         "node_csa_val",           // Required
			"node_csa_signature":   "mode_csa_signature_val", // Required
		}
	}
	defaultMessageBody := []byte("body bytes")

	mustNewGRPCClient := func(t *testing.T, exporterMock *mocks.OTLPExporter) *Client {
		// Override exporter factory which is used by Client
		exporterFactory := func(...otlploggrpc.Option) (sdklog.Exporter, error) {
			return exporterMock, nil
		}
		client, err := newGRPCClient(TestDefaultConfig(), exporterFactory)
		if err != nil {
			t.Fatalf("Error creating beholder client: %v", err)
		}
		return client
	}

	mustNewHTTPClient := func(t *testing.T, exporterMock *mocks.OTLPExporter) *Client {
		// Override exporter factory which is used by Client
		exporterFactory := func(...otlploghttp.Option) (sdklog.Exporter, error) {
			return exporterMock, nil
		}
		client, err := newHTTPClient(TestDefaultConfigHTTPClient(), exporterFactory)
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
		messageGenerator       func(client *Client, messageBody []byte, customAttributes map[string]any)
		mustNewGrpcClient      func(*testing.T, *mocks.OTLPExporter) *Client
	}{
		{
			name:                   "Test Emit (GRPC Client)",
			makeCustomAttributes:   defaultCustomAttributes,
			messageBody:            defaultMessageBody,
			messageCount:           10,
			exporterMockErrorCount: 0,
			exporterOutputExpected: true,
			messageGenerator: func(client *Client, messageBody []byte, customAttributes map[string]any) {
				err := client.Emitter.Emit(tests.Context(t), messageBody, customAttributes)
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
			messageGenerator: func(client *Client, messageBody []byte, customAttributes map[string]any) {
				err := client.Emitter.Emit(tests.Context(t), messageBody, customAttributes)
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

			otel.SetErrorHandler(otelMustNotErr(t))
			// Number of exported messages
			exportedMessageCount := 0

			// Simulate exporter error if configured
			if tc.exporterMockErrorCount > 0 {
				exporterMock.On("Export", mock.Anything, mock.Anything).Return(fmt.Errorf("an error occurred")).Times(tc.exporterMockErrorCount)
			}
			customAttributes := tc.makeCustomAttributes()
			if tc.exporterOutputExpected {
				exporterMock.On("Export", mock.Anything, mock.Anything).Return(nil).Times(tc.messageCount).
					Run(func(args mock.Arguments) {
						assert.IsType(t, args.Get(1), []sdklog.Record{}, "Record type mismatch")
						records := args.Get(1).([]sdklog.Record)
						assert.Equal(t, 1, len(records), "batching is disabled, expecte 1 record")
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
							expectedKv := OtelAttr(key, expectedValue)
							equal := kv.Value.Equal(expectedKv.Value)
							assert.True(t, equal, fmt.Sprintf("Record attributes mismatch for key %v", key))
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
				tc.messageGenerator(client, tc.messageBody, customAttributes)
			}
			assert.Equal(t, tc.messageCount, exportedMessageCount, "Expect all emitted messages to be exported")
		})
	}
}

func TestEmitterMessageValidation(t *testing.T) {
	getEmitter := func(exporterMock *mocks.OTLPExporter) Emitter {
		client, err := newGRPCClient(
			TestDefaultConfig(),
			// Override exporter factory which is used by Client
			func(...otlploggrpc.Option) (sdklog.Exporter, error) {
				return exporterMock, nil
			},
		)
		otel.SetErrorHandler(otelMustNotErr(t))
		assert.NoError(t, err)
		return client.Emitter
	}

	for _, tc := range []struct {
		name                string
		attrs               Attributes
		exporterCalledTimes int
		expectedError       string
	}{
		{
			name: "Missing required attribute",
			attrs: Attributes{
				"key": "value",
			},
			exporterCalledTimes: 0,
			expectedError:       "'Metadata.BeholderDataSchema' Error:Field validation for 'BeholderDataSchema' failed on the 'required' tag",
		},
		{
			name: "Invalid URI",
			attrs: Attributes{
				"beholder_domain":      "TestDomain",
				"beholder_entity":      "TestEntity",
				"beholder_data_schema": "example-schema",
				"node_csa_key":         "node_csa_val",           // Required
				"node_csa_signature":   "mode_csa_signature_val", // Required
			},
			exporterCalledTimes: 0,
			expectedError:       "'Metadata.BeholderDataSchema' Error:Field validation for 'BeholderDataSchema' failed on the 'uri' tag",
		},
		{
			name: "Invalid Beholder domain (double underscore)",
			attrs: Attributes{
				"beholder_data_schema": "/example-schema/versions/1",
				"beholder_entity":      "TestEntity",
				"beholder_domain":      "Test__Domain",
			},
			exporterCalledTimes: 0,
			expectedError:       "'Metadata.BeholderDomain' Error:Field validation for 'BeholderDomain' failed on the 'domain_entity' tag",
		},
		{
			name: "Invalid Beholder domain (special characters)",
			attrs: Attributes{
				"beholder_data_schema": "/example-schema/versions/1",
				"beholder_entity":      "TestEntity",
				"beholder_domain":      "TestDomain*$",
			},
			exporterCalledTimes: 0,
			expectedError:       "'Metadata.BeholderDomain' Error:Field validation for 'BeholderDomain' failed on the 'domain_entity' tag",
		},
		{
			name: "Invalid Beholder entity (double underscore)",
			attrs: Attributes{
				"beholder_data_schema": "/example-schema/versions/1",
				"beholder_entity":      "Test__Entity",
				"beholder_domain":      "TestDomain",
			},
			exporterCalledTimes: 0,
			expectedError:       "'Metadata.BeholderEntity' Error:Field validation for 'BeholderEntity' failed on the 'domain_entity' tag",
		},
		{
			name: "Invalid Beholder entity (special characters)",
			attrs: Attributes{
				"beholder_data_schema": "/example-schema/versions/1",
				"beholder_entity":      "TestEntity*$",
				"beholder_domain":      "TestDomain",
			},
			exporterCalledTimes: 0,
			expectedError:       "'Metadata.BeholderEntity' Error:Field validation for 'BeholderEntity' failed on the 'domain_entity' tag",
		},
		{
			name:                "Valid Attributes",
			exporterCalledTimes: 1,
			attrs: Attributes{
				"beholder_domain":      "TestDomain",
				"beholder_entity":      "TestEntity",
				"beholder_data_schema": "/example-schema/versions/1",
			},
			expectedError: "",
		},
		{
			name:                "Valid Attributes (special characters)",
			exporterCalledTimes: 1,
			attrs: Attributes{
				"beholder_domain":      "Test.Domain_42-1",
				"beholder_entity":      "Test.Entity_42-1",
				"beholder_data_schema": "/example-schema/versions/1",
				"node_csa_key":         "node_csa_val",           // Required
				"node_csa_signature":   "mode_csa_signature_val", // Required
			},
			expectedError: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("Emitter.Emit", func(t *testing.T) {
				// Setup
				exporterMock := mocks.NewOTLPExporter(t)
				if tc.exporterCalledTimes > 0 {
					exporterMock.On("Export", mock.Anything, mock.Anything).Return(nil).Times(tc.exporterCalledTimes)
				}
				emitter := getEmitter(exporterMock)
				message := NewMessage([]byte("test"), tc.attrs)
				// Emit
				err := emitter.Emit(tests.Context(t), message.Body, tc.attrs)
				// Assert expectations
				if tc.expectedError != "" {
					assert.ErrorContains(t, err, tc.expectedError)
				} else {
					assert.NoError(t, err)
				}
				if tc.exporterCalledTimes > 0 {
					exporterMock.AssertExpectations(t)
				} else {
					exporterMock.AssertNotCalled(t, "Export")
				}
			})
		})
	}
}

func TestClient_Close(t *testing.T) {
	exporterMock := mocks.NewOTLPExporter(t)
	defer exporterMock.AssertExpectations(t)

	client, err := NewStdoutClient()
	assert.NoError(t, err)

	err = client.Close()
	assert.NoError(t, err)

	exporterMock.AssertExpectations(t)
}

func TestClient_ForPackage(t *testing.T) {
	exporterMock := mocks.NewOTLPExporter(t)
	defer exporterMock.AssertExpectations(t)
	var b strings.Builder
	client, err := NewWriterClient(&b)
	assert.NoError(t, err)
	clientForTest := client.ForPackage("TestClient_ForPackage")

	// Log
	clientForTest.Logger.Emit(tests.Context(t), otellog.Record{})
	assert.Contains(t, b.String(), `"Name":"TestClient_ForPackage"`)
	b.Reset()

	// Trace
	_, span := clientForTest.Tracer.Start(tests.Context(t), "testSpan")
	span.End()
	assert.Contains(t, b.String(), `"Name":"TestClient_ForPackage"`)
	assert.Contains(t, b.String(), "testSpan")
	b.Reset()

	// Meter
	counter, _ := clientForTest.Meter.Int64Counter("testMetric")
	counter.Add(tests.Context(t), 1)
	clientForTest.Close()
	assert.Contains(t, b.String(), `"Name":"TestClient_ForPackage"`)
	assert.Contains(t, b.String(), "testMetric")
}

func otelMustNotErr(t *testing.T) otel.ErrorHandlerFunc {
	return func(err error) { t.Fatalf("otel error: %v", err) }
}

func TestNewClient(t *testing.T) {
	t.Run("both endpoints set", func(t *testing.T) {
		client, err := NewClient(Config{
			OtelExporterGRPCEndpoint: "grpc-endpoint",
			OtelExporterHTTPEndpoint: "http-endpoint",
		})
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Equal(t, "only one exporter endpoint should be set", err.Error())
	})

	t.Run("no endpoints set", func(t *testing.T) {
		client, err := NewClient(Config{})
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Equal(t, "at least one exporter endpoint should be set", err.Error())
	})

	t.Run("GRPC endpoint set", func(t *testing.T) {
		client, err := NewClient(Config{
			OtelExporterGRPCEndpoint: "grpc-endpoint",
		})
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &Client{}, client)
	})

	t.Run("HTTP endpoint set", func(t *testing.T) {
		client, err := NewClient(Config{
			OtelExporterHTTPEndpoint: "http-endpoint",
		})
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.IsType(t, &Client{}, client)
	})
}

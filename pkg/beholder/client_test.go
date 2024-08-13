package beholder

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	otellog "go.opentelemetry.io/otel/log"
	otelglobal "go.opentelemetry.io/otel/log/global"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	oteltrace "go.opentelemetry.io/otel/trace"

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
			"int_key_1":            123,
			"int64_key_1":          int64(123),
			"int32_key_1":          int32(123),
			"str_key_1":            "str_val_1",
			"bool_key_1":           true,
			"float_key_1":          123.456,
			"byte_key_1":           []byte("byte_val_1"),
			"str_slice_key_1":      []string{"str_val_1", "str_val_2"},
			"nil_key_1":            nil,
			"beholder_data_schema": "/schemas/ids/1001", // Required field, URI
		}
	}
	defaultMessageBody := []byte("body bytes")

	tests := []struct {
		name                   string
		makeCustomAttributes   func() map[string]any
		messageBody            []byte
		messageCount           int
		exporterMockErrorCount int
		exporterOutputExpected bool
		messageGenerator       func(client OtelClient, messageBody []byte, customAttributes map[string]any)
	}{
		{
			name:                   "Test Emit",
			makeCustomAttributes:   defaultCustomAttributes,
			messageBody:            defaultMessageBody,
			messageCount:           10,
			exporterMockErrorCount: 0,
			exporterOutputExpected: true,
			messageGenerator: func(client OtelClient, messageBody []byte, customAttributes map[string]any) {
				err := client.Emitter.Emit(context.Background(), messageBody, customAttributes)
				assert.NoError(t, err)
			},
		}, {
			name:                   "Test EmitMessage",
			makeCustomAttributes:   defaultCustomAttributes,
			messageBody:            defaultMessageBody,
			messageCount:           10,
			exporterMockErrorCount: 0,
			exporterOutputExpected: true,
			messageGenerator: func(client OtelClient, messageBody []byte, customAttributes map[string]any) {
				message := NewMessage(messageBody, customAttributes)
				err := client.Emitter.EmitMessage(context.Background(), message)
				assert.NoError(t, err)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exporterMock := mocks.NewOTLPExporter(t)
			defer exporterMock.AssertExpectations(t)

			otelErrorHandler := func(err error) {
				t.Fatalf("otel error: %v", err)
			}
			// Override exporter factory which is used by Client
			exporterFactory := func(context.Context, ...otlploggrpc.Option) (sdklog.Exporter, error) {
				return exporterMock, nil
			}
			client, err := newOtelClient(DefaultConfig(), otelErrorHandler, exporterFactory)
			if err != nil {
				t.Fatalf("Error creating beholder client: %v", err)
			}
			// Number of messages to emit
			done := make(chan struct{}, 1)

			// Simulate exporter error if configured
			if tc.exporterMockErrorCount > 0 {
				exporterMock.On("Export", mock.Anything, mock.Anything).Return(fmt.Errorf("an error occurred")).Times(tc.exporterMockErrorCount)
			}

			customAttributes := tc.makeCustomAttributes()

			if tc.exporterOutputExpected {
				exporterMock.On("Export", mock.Anything, mock.Anything).Return(nil).Once().
					Run(func(args mock.Arguments) {
						assert.IsType(t, args.Get(1), []sdklog.Record{}, "Record type mismatch")
						records := args.Get(1).([]sdklog.Record)
						assert.Equal(t, tc.messageCount, len(records), "Record count mismatch")
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
						done <- struct{}{}
					})
			}
			for i := 0; i < tc.messageCount; i++ {
				go tc.messageGenerator(client, tc.messageBody, customAttributes)
			}

			select {
			case <-done:
			case <-time.After(10 * time.Second):
				t.Fatalf("Timed out waiting for messages to be emitted")
			}
		})
	}
}

func TestEmitterMessageValidation(t *testing.T) {
	getEmitter := func(exporterMock *mocks.OTLPExporter) Emitter {
		client, err := newOtelClient(
			DefaultConfig(),
			func(err error) { t.Fatalf("otel error: %v", err) },
			// Override exporter factory which is used by Client
			func(context.Context, ...otlploggrpc.Option) (sdklog.Exporter, error) {
				return exporterMock, nil
			},
		)
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
				"beholder_data_schema": "example-schema",
			},
			exporterCalledTimes: 0,
			expectedError:       "'Metadata.BeholderDataSchema' Error:Field validation for 'BeholderDataSchema' failed on the 'uri' tag",
		},
		{
			name:                "Valid URI",
			exporterCalledTimes: 1,
			attrs: Attributes{
				"beholder_data_schema": "/example-schema/versions/1",
			},
			expectedError: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			setupMock := func(exporterMock *mocks.OTLPExporter) (*mocks.OTLPExporter, <-chan struct{}) {
				done := make(chan struct{}, tc.exporterCalledTimes)
				if tc.exporterCalledTimes > 0 {
					exporterMock.On("Export", mock.Anything, mock.Anything).Return(nil).Times(tc.exporterCalledTimes).
						Run(func(args mock.Arguments) {
							done <- struct{}{}
						})
				}
				return exporterMock, done
			}

			assertError := func(err error, expected string) {
				if tc.expectedError != "" {
					assert.ErrorContains(t, err, expected)
				} else {
					assert.NoError(t, err)
				}
			}

			assertMock := func(exporterMock *mocks.OTLPExporter) {
				if tc.exporterCalledTimes > 0 {
					exporterMock.AssertExpectations(t)
				} else {
					exporterMock.AssertNotCalled(t, "Export")
				}
			}

			waitUntilSent := func(done <-chan struct{}) {
				for range tc.exporterCalledTimes {
					select {
					case <-done:
					case <-time.After(10 * time.Second):
						t.Fatalf("Timed out waiting for messages to be emitted")
					}
				}
			}

			setupTest := func() (emitter Emitter, message Message, assertExpectations func(err error)) {
				exporterMock, done := setupMock(mocks.NewOTLPExporter(t))
				emitter = getEmitter(exporterMock)
				message = NewMessage([]byte("test"), tc.attrs)

				assertExpectations = func(err error) {
					assertError(err, tc.expectedError)
					if err == nil {
						waitUntilSent(done)
					}
					assertMock(exporterMock)
				}
				return
			}

			t.Run("Emitter.EmitMessage", func(t *testing.T) {
				emitter, message, assertExpectations := setupTest()

				err := emitter.EmitMessage(context.Background(), message)

				assertExpectations(err)
			})

			t.Run("Emitter.Emit", func(t *testing.T) {
				emitter, message, assertExpectations := setupTest()

				err := emitter.Emit(context.Background(), message.Body, tc.attrs)

				assertExpectations(err)
			})
		})
	}
}

func TestClient_Close(t *testing.T) {
	exporterMock := mocks.NewOTLPExporter(t)
	defer exporterMock.AssertExpectations(t)

	otelErrorHandler := func(err error) {
		t.Fatalf("otel error: %v", err)
	}
	// Override exporter factory which is used by Client
	exporterFactory := func(context.Context, ...otlploggrpc.Option) (sdklog.Exporter, error) {
		return exporterMock, nil
	}
	client, err := newOtelClient(DefaultConfig(), otelErrorHandler, exporterFactory)
	if err != nil {
		t.Fatalf("Error creating beholder client: %v", err)
	}

	exporterMock.On("Shutdown", mock.Anything).Return(nil).Once()

	err = client.Close()
	assert.NoError(t, err)

	exporterMock.AssertExpectations(t)
}

func TestClient_SetGlobas(t *testing.T) {
	exporterMock := mocks.NewOTLPExporter(t)
	defer exporterMock.AssertExpectations(t)

	otelErrorHandler := func(err error) {
		t.Fatalf("otel error: %v", err)
	}
	// Override exporter factory which is used by Client
	exporterFactory := func(context.Context, ...otlploggrpc.Option) (sdklog.Exporter, error) {
		return exporterMock, nil
	}
	client, err := newOtelClient(DefaultConfig(), otelErrorHandler, exporterFactory)
	if err != nil {
		t.Fatalf("Error creating beholder client: %v", err)
	}
	assert.NoError(t, err)

	globals := getGlobals()
	defer restoreGlobas(t, globals)

	client.SetGlobals()

	assert.Equal(t, client.LoggerProvider, otelglobal.GetLoggerProvider())
	assert.Equal(t, client.TracerProvider, otel.GetTracerProvider())
	assert.Equal(t, client.MeterProvider, otel.GetMeterProvider())
}

type globals struct {
	loggerProvider    otellog.LoggerProvider
	tracerProvider    oteltrace.TracerProvider
	textMapPropagator propagation.TextMapPropagator
	meterProvider     otelmetric.MeterProvider
}

func getGlobals() globals {
	return globals{
		otelglobal.GetLoggerProvider(),
		otel.GetTracerProvider(),
		otel.GetTextMapPropagator(),
		otel.GetMeterProvider(),
	}
}

func restoreGlobas(t *testing.T, g globals) {
	otelglobal.SetLoggerProvider(g.loggerProvider)
	otel.SetTracerProvider(g.tracerProvider)
	otel.SetTextMapPropagator(g.textMapPropagator)
	otel.SetMeterProvider(g.meterProvider)

	assert.Equal(t, g.loggerProvider, otelglobal.GetLoggerProvider())
	assert.Equal(t, g.tracerProvider, otel.GetTracerProvider())
	assert.Equal(t, g.textMapPropagator, otel.GetTextMapPropagator())
	assert.Equal(t, g.meterProvider, otel.GetMeterProvider())
}

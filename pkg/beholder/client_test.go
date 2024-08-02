package beholder

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/mocks"
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
	defaultEventBody := []byte("body bytes")

	tests := []struct {
		name                 string
		makeCustomAttributes func() map[string]any
		// NOTE: skipping these attributes is necessary due to a limitation in sdklog.Record
		// see INFOPLAT-811
		skipAttributes         []string
		eventBody              []byte
		eventCount             int
		exporterMockErrorCount int
		exporterOutputExpected bool
		eventGenerator         func(client *BeholderClient, eventBody []byte, customAttributes map[string]any)
	}{
		{
			name:                   "Test Emit",
			makeCustomAttributes:   defaultCustomAttributes,
			skipAttributes:         []string{},
			eventBody:              defaultEventBody,
			eventCount:             10,
			exporterMockErrorCount: 0,
			exporterOutputExpected: true,
			eventGenerator: func(client *BeholderClient, eventBody []byte, customAttributes map[string]any) {
				err := client.EventEmitter().Emit(context.Background(), eventBody, customAttributes)
				assert.NoError(t, err)
			},
		}, {
			name:                   "Test EmitEvent",
			makeCustomAttributes:   defaultCustomAttributes,
			skipAttributes:         []string{},
			eventBody:              defaultEventBody,
			eventCount:             10,
			exporterMockErrorCount: 0,
			exporterOutputExpected: true,
			eventGenerator: func(client *BeholderClient, eventBody []byte, customAttributes map[string]any) {
				event := NewEvent(eventBody, customAttributes)
				err := client.EventEmitter().EmitEvent(context.Background(), event)
				assert.NoError(t, err)
			},
		},
		{
			name:                 "Test Send",
			makeCustomAttributes: defaultCustomAttributes,
			// NOTE: removing these attributes is necessary due to a limitation in sdklog.Record
			// see INFOPLAT-811
			skipAttributes:         []string{"str_key_1", "str_slice_key_1", "int32_key_1", "beholder_data_schema"},
			eventBody:              defaultEventBody,
			eventCount:             1,
			exporterMockErrorCount: 1, // Simulate exporter error to test for retries
			exporterOutputExpected: true,
			eventGenerator: func(client *BeholderClient, eventBody []byte, customAttributes map[string]any) {
				err := client.EventEmitter().Send(context.Background(), eventBody, customAttributes)
				assert.NoError(t, err)
			},
		},
		{
			name:                 "Test SendEvent",
			makeCustomAttributes: defaultCustomAttributes,
			// NOTE: removing these attributes is necessary due to a limitation in sdklog.Record
			// see INFOPLAT-811
			skipAttributes:         []string{"str_key_1", "str_slice_key_1", "int32_key_1", "beholder_data_schema"},
			eventBody:              defaultEventBody,
			eventCount:             1,
			exporterMockErrorCount: 1, // Simulate exporter error to test for retries
			exporterOutputExpected: true,
			eventGenerator: func(client *BeholderClient, eventBody []byte, customAttributes map[string]any) {
				event := NewEvent(eventBody, customAttributes)
				err := client.EventEmitter().SendEvent(context.Background(), event)
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
			// Override exporter factory which is used by BeholderClient
			exporterFactory := func(context.Context, ...otlploggrpc.Option) (sdklog.Exporter, error) {
				return exporterMock, nil
			}
			client, err := newOtelClient(DefaultBeholderConfig(), otelErrorHandler, exporterFactory)
			if err != nil {
				t.Fatalf("Error creating beholder client: %v", err)
			}
			// Number of events to emit
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
						assert.Equal(t, tc.eventCount, len(records), "Record count mismatch")
						record := records[0]
						assert.Equal(t, tc.eventBody, record.Body().AsBytes(), "Record body mismatch")
						actualAttributeKeys := map[string]struct{}{}
						record.WalkAttributes(func(kv otellog.KeyValue) bool {
							key := kv.Key
							if slices.Contains(tc.skipAttributes, key) {
								// NOTE: skipping these attributes is necessary due to a limitation in sdklog.Record
								// see INFOPLAT-811
								t.Logf("Skipping attribute key: %s. See INFOPLAT-811", key)
								return true
							}
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

							if slices.Contains(tc.skipAttributes, key) {
								// NOTE: skipping these attributes is necessary due to a limitation in sdklog.Record
								// see INFOPLAT-811
								t.Logf("Skipping attribute key: %s. See INFOPLAT-811", key)
								continue
							}
							if _, ok := actualAttributeKeys[key]; !ok {
								t.Fatalf("Record attribute key not found: %s", key)
							}
						}
						done <- struct{}{}
					})
			}
			for i := 0; i < tc.eventCount; i++ {
				go tc.eventGenerator(client, tc.eventBody, customAttributes)
			}

			select {
			case <-done:
			case <-time.After(10 * time.Second):
				t.Fatalf("Timed out waiting for events to be emitted")
			}
		})
	}
}

func TestEmitterEventValidation(t *testing.T) {
	getEmitter := func(exporterMock *mocks.OTLPExporter) EventEmitter {
		client, err := newOtelClient(
			DefaultBeholderConfig(),
			func(err error) { t.Fatalf("otel error: %v", err) },
			// Override exporter factory which is used by BeholderClient
			func(context.Context, ...otlploggrpc.Option) (sdklog.Exporter, error) {
				return exporterMock, nil
			},
		)
		assert.NoError(t, err)
		return client.EventEmitter()
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
				"beholder_data_schema": "beholder/pb/example.proto",
			},
			exporterCalledTimes: 0,
			expectedError:       "'Metadata.BeholderDataSchema' Error:Field validation for 'BeholderDataSchema' failed on the 'uri' tag",
		},
		{
			name:                "Valid URI",
			exporterCalledTimes: 1,
			attrs: Attributes{
				"beholder_data_schema": "https://example.com/example.proto",
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
				for i := 0; i < tc.exporterCalledTimes; i++ {
					select {
					case <-done:
					case <-time.After(10 * time.Second):
						t.Fatalf("Timed out waiting for events to be emitted")
					}
				}
			}

			setupTest := func() (emitter EventEmitter, event Event, assertExpectations func(err error)) {
				exporterMock, done := setupMock(mocks.NewOTLPExporter(t))
				emitter = getEmitter(exporterMock)
				event = NewEvent([]byte("test"), tc.attrs)

				assertExpectations = func(err error) {
					assertError(err, tc.expectedError)
					if err == nil {
						waitUntilSent(done)
					}
					assertMock(exporterMock)
				}
				return
			}

			t.Run("Emitter.SendEvent", func(t *testing.T) {
				emitter, event, assertExpectations := setupTest()

				err := emitter.SendEvent(context.Background(), event)

				assertExpectations(err)
			})

			t.Run("Emitter.Send", func(t *testing.T) {
				emitter, event, assertExpectations := setupTest()

				err := emitter.Send(context.Background(), event.Body, tc.attrs)

				assertExpectations(err)
			})

			t.Run("Emitter.EmitEvent", func(t *testing.T) {
				emitter, event, assertExpectations := setupTest()

				err := emitter.EmitEvent(context.Background(), event)

				assertExpectations(err)
			})

			t.Run("Emitter.Emit", func(t *testing.T) {
				emitter, event, assertExpectations := setupTest()

				err := emitter.Emit(context.Background(), event.Body, tc.attrs)

				assertExpectations(err)
			})
		})
	}
}

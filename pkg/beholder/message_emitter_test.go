package beholder_test

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

func TestEmitterMessageValidation(t *testing.T) {
	getEmitter := func(exporterMock *mocks.OTLPExporter) beholder.Emitter {
		client, err := beholder.NewGRPCClient(
			beholder.TestDefaultConfig(),
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
		attrs               beholder.Attributes
		exporterCalledTimes int
		expectedError       string
	}{
		{
			name: "Missing required attribute",
			attrs: beholder.Attributes{
				"key": "value",
			},
			exporterCalledTimes: 0,
			expectedError:       "'Metadata.BeholderDataSchema' Error:Field validation for 'BeholderDataSchema' failed on the 'required' tag",
		},
		{
			name: "Invalid URI",
			attrs: beholder.Attributes{
				beholder.AttrKeyDomain:     "TestDomain",
				beholder.AttrKeyEntity:     "TestEntity",
				beholder.AttrKeyDataSchema: "example-schema",
			},
			exporterCalledTimes: 0,
			expectedError:       "'Metadata.BeholderDataSchema' Error:Field validation for 'BeholderDataSchema' failed on the 'uri' tag",
		},
		{
			name: "Invalid Beholder domain (double underscore)",
			attrs: beholder.Attributes{
				beholder.AttrKeyDataSchema: "/example-schema/versions/1",
				beholder.AttrKeyEntity:     "TestEntity",
				beholder.AttrKeyDomain:     "Test__Domain",
			},
			exporterCalledTimes: 0,
			expectedError:       "'Metadata.BeholderDomain' Error:Field validation for 'BeholderDomain' failed on the 'domain_entity' tag",
		},
		{
			name: "Invalid Beholder domain (special characters)",
			attrs: beholder.Attributes{
				beholder.AttrKeyDataSchema: "/example-schema/versions/1",
				beholder.AttrKeyEntity:     "TestEntity",
				beholder.AttrKeyDomain:     "TestDomain*$",
			},
			exporterCalledTimes: 0,
			expectedError:       "'Metadata.BeholderDomain' Error:Field validation for 'BeholderDomain' failed on the 'domain_entity' tag",
		},
		{
			name: "Invalid Beholder entity (double underscore)",
			attrs: beholder.Attributes{
				beholder.AttrKeyDataSchema: "/example-schema/versions/1",
				beholder.AttrKeyEntity:     "Test__Entity",
				beholder.AttrKeyDomain:     "TestDomain",
			},
			exporterCalledTimes: 0,
			expectedError:       "'Metadata.BeholderEntity' Error:Field validation for 'BeholderEntity' failed on the 'domain_entity' tag",
		},
		{
			name: "Invalid Beholder entity (special characters)",
			attrs: beholder.Attributes{
				beholder.AttrKeyDataSchema: "/example-schema/versions/1",
				beholder.AttrKeyEntity:     "TestEntity*$",
				beholder.AttrKeyDomain:     "TestDomain",
			},
			exporterCalledTimes: 0,
			expectedError:       "'Metadata.BeholderEntity' Error:Field validation for 'BeholderEntity' failed on the 'domain_entity' tag",
		},
		{
			name:                "Valid Attributes",
			exporterCalledTimes: 1,
			attrs: beholder.Attributes{
				beholder.AttrKeyDomain:     "TestDomain",
				beholder.AttrKeyEntity:     "TestEntity",
				beholder.AttrKeyDataSchema: "/example-schema/versions/1",
			},
			expectedError: "",
		},
		{
			name:                "Valid Attributes (special characters)",
			exporterCalledTimes: 1,
			attrs: beholder.Attributes{
				beholder.AttrKeyDomain:     "Test.Domain_42-1",
				beholder.AttrKeyEntity:     "Test.Entity_42-1",
				beholder.AttrKeyDataSchema: "/example-schema/versions/1",
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
				message := beholder.NewMessage([]byte("test"), tc.attrs)
				// Emit
				err := emitter.Emit(t.Context(), message.Body, tc.attrs)
				// Assert expectations
				if tc.expectedError != "" {
					require.ErrorContains(t, err, tc.expectedError)
				} else {
					require.NoError(t, err)
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

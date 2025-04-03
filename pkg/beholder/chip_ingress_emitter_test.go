package beholder_test

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewChipIngressEmitter(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		clientMock := &mocks.ChipIngressClient{}
		emitter, err := beholder.NewChipIngressEmitter(clientMock)
		require.NoError(t, err)
		assert.NotNil(t, emitter)
	})

	t.Run("returns error when client is nil", func(t *testing.T) {
		emitter, err := beholder.NewChipIngressEmitter(nil)
		assert.Error(t, err)
		assert.Nil(t, emitter)
	})
}

func TestNewChipIngressClient(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		client, err := beholder.NewChipIngressClient(beholder.Config{
			ChipIngressEmitterGRPCEndpoint: "localhost:8080",
		})
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("returns error when endpoint is empty", func(t *testing.T) {
		client, err := beholder.NewChipIngressClient(beholder.Config{
			ChipIngressEmitterGRPCEndpoint: "",
		})
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestChipIngressEmit(t *testing.T) {

	body := []byte("test body")
	domain := "test-domain"
	entity := "test-entity"

	t.Run("happy path", func(t *testing.T) {

		clientMock := &mocks.ChipIngressClient{}

		clientMock.
			On("Publish", mock.Anything, mock.Anything).
			Return(nil, nil)

		emitter, err := beholder.NewChipIngressEmitter(clientMock)
		require.NoError(t, err)

		err = emitter.Emit(t.Context(), body, "beholder_domain", domain, "beholder_entity", entity)
		require.NoError(t, err)

		clientMock.AssertExpectations(t)
	})

	t.Run("returns error when extractSourceAndType fails", func(t *testing.T) {

		clientMock := &mocks.ChipIngressClient{}

		clientMock.
			On("Publish", mock.Anything, mock.Anything).
			Return(nil, nil)

		emitter, err := beholder.NewChipIngressEmitter(clientMock)
		require.NoError(t, err)

		err = emitter.Emit(t.Context(), body, "beholder_domain", domain)
		assert.Error(t, err)
	})

	t.Run("returns error when Publish fails", func(t *testing.T) {
		clientMock := &mocks.ChipIngressClient{}

		clientMock.
			On("Publish", mock.Anything, mock.Anything).
			Return(nil, assert.AnError)

		emitter, err := beholder.NewChipIngressEmitter(clientMock)
		require.NoError(t, err)

		err = emitter.Emit(t.Context(), body, "beholder_domain", domain, "beholder_entity", entity)
		require.Error(t, err)

		clientMock.AssertExpectations(t)
	})
}

func TestExtractSourceAndType(t *testing.T) {
	tests := []struct {
		name          string
		attrs         []any
		wantDomain    string
		wantEntity    string
		wantErr       bool
		expectedError string
	}{
		{
			name:       "happy path - domain and entity exist",
			attrs:      []any{"beholder_domain", "test-domain", "beholder_entity", "test-entity"},
			wantDomain: "test-domain",
			wantEntity: "test-entity",
			wantErr:    false,
		},
		{
			name:          "missing domain",
			attrs:         []any{"beholder_entity", "test-entity"},
			wantDomain:    "",
			wantEntity:    "",
			wantErr:       true,
			expectedError: "beholder_domain not found in provided key/value attributes",
		},
		{
			name:          "missing entity",
			attrs:         []any{"beholder_domain", "test-domain"},
			wantDomain:    "",
			wantEntity:    "",
			wantErr:       true,
			expectedError: "beholder_entity not found in provided key/value attributes",
		},
		{
			name:       "domain and entity with additional attributes",
			attrs:      []any{"other_key", "other_value", "beholder_domain", "test-domain", "beholder_entity", "test-entity", "something_else", 123},
			wantDomain: "test-domain",
			wantEntity: "test-entity",
			wantErr:    false,
		},
		{
			name:       "non-string keys ignored",
			attrs:      []any{123, "value", "beholder_domain", "test-domain", "beholder_entity", "test-entity"},
			wantDomain: "test-domain",
			wantEntity: "test-entity",
			wantErr:    false,
		},
		{
			name:       "non-string values handled",
			attrs:      []any{"other_key", 123, "beholder_domain", "test-domain", "beholder_entity", "test-entity"},
			wantDomain: "test-domain",
			wantEntity: "test-entity",
			wantErr:    false,
		},
		{
			name:          "empty attribute list",
			attrs:         []any{},
			wantDomain:    "",
			wantEntity:    "",
			wantErr:       true,
			expectedError: "beholder_domain not found in provided key/value attributes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain, entity, err := beholder.ExtractSourceAndType(tt.attrs...)

			if tt.wantErr {
				if err == nil {
					t.Errorf("extractSourceAndType() error = nil, want error")
					return
				}
				if tt.expectedError != "" && tt.expectedError != err.Error() {
					t.Errorf("extractSourceAndType() error = %v, want %v", err, tt.expectedError)
				}
				return
			}

			if err != nil {
				t.Errorf("extractSourceAndType() unexpected error = %v", err)
				return
			}

			if domain != tt.wantDomain {
				t.Errorf("extractSourceAndType() domain = %v, want %v", domain, tt.wantDomain)
			}

			if entity != tt.wantEntity {
				t.Errorf("extractSourceAndType() entity = %v, want %v", entity, tt.wantEntity)
			}
		})
	}
}

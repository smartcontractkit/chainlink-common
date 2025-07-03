package beholder_test

import (
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewChipIngressEmitter(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		clientMock := mocks.NewChipIngressClient(t)
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

func TestChipIngressEmit(t *testing.T) {

	body := []byte("test body")
	domain := "test-domain"
	entity := "test-entity"
	attributes := map[string]any{
		"datacontenttype": "application/protobuf",
		"dataschema":      "/schemas/ids/1001",
		"subject":         "example-subject",
		"time":            time.Now(),
	}

	t.Run("happy path", func(t *testing.T) {

		clientMock := &mocks.ChipIngressClient{}

		clientMock.
			On("Publish", mock.Anything, mock.Anything).
			Return(nil, nil)

		emitter, err := beholder.NewChipIngressEmitter(clientMock)
		require.NoError(t, err)

		err = emitter.Emit(t.Context(), body, beholder.AttrKeyDomain, domain, beholder.AttrKeyEntity, entity, attributes)
		require.NoError(t, err)

		clientMock.AssertExpectations(t)
	})

	t.Run("returns error when ExtractSourceAndType fails", func(t *testing.T) {

		clientMock := &mocks.ChipIngressClient{}

		clientMock.
			On("Publish", mock.Anything, mock.Anything).
			Return(nil, nil)

		emitter, err := beholder.NewChipIngressEmitter(clientMock)
		require.NoError(t, err)

		err = emitter.Emit(t.Context(), body, beholder.AttrKeyDomain, domain)
		assert.Error(t, err)
	})

	t.Run("returns error when Publish fails", func(t *testing.T) {

		clientMock := &mocks.ChipIngressClient{}

		clientMock.
			On("Publish", mock.Anything, mock.Anything).
			Return(nil, assert.AnError)

		emitter, err := beholder.NewChipIngressEmitter(clientMock)
		require.NoError(t, err)

		err = emitter.Emit(t.Context(), body, beholder.AttrKeyDomain, domain, beholder.AttrKeyEntity, entity)
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
			attrs:      []any{map[string]any{beholder.AttrKeyDomain: "test-domain", beholder.AttrKeyEntity: "test-entity"}},
			wantDomain: "test-domain",
			wantEntity: "test-entity",
			wantErr:    false,
		},
		{
			name:       "happy path - domain and entity exist - source/type naming",
			attrs:      []any{map[string]any{"source": "test-domain", "type": "test-entity"}},
			wantDomain: "test-domain",
			wantEntity: "test-entity",
			wantErr:    false,
		},
		{
			name:       "happy path - domain and entity exist - uses source/type",
			attrs:      []any{map[string]any{"source": "other-domain", beholder.AttrKeyDomain: "test-domain", beholder.AttrKeyEntity: "test-entity", "type": "other-entity"}},
			wantDomain: "other-domain",
			wantEntity: "other-entity",
			wantErr:    false,
		},
		{
			name:          "missing domain/source",
			attrs:         []any{map[string]any{beholder.AttrKeyEntity: "test-entity"}},
			wantDomain:    "",
			wantEntity:    "",
			wantErr:       true,
			expectedError: "source/beholder_domain not found in provided key/value attributes",
		},
		{
			name:          "missing entity/type",
			attrs:         []any{map[string]any{beholder.AttrKeyDomain: "test-domain"}},
			wantDomain:    "",
			wantEntity:    "",
			wantErr:       true,
			expectedError: "type/beholder_entity not found in provided key/value attributes",
		},
		{
			name:          "missing domain/source",
			attrs:         []any{"type", "test-entity"},
			wantDomain:    "",
			wantEntity:    "",
			wantErr:       true,
			expectedError: "source/beholder_domain not found in provided key/value attributes",
		},
		{
			name:          "missing entity/type",
			attrs:         []any{"source", "test-domain"},
			wantDomain:    "",
			wantEntity:    "",
			wantErr:       true,
			expectedError: "type/beholder_entity not found in provided key/value attributes",
		},
		{
			name: "domain and entity with additional attributes",
			attrs: []any{map[string]any{
				"other_key":            "other_value",
				beholder.AttrKeyDomain: "test-domain",
				beholder.AttrKeyEntity: "test-entity",
				"something_else":       123,
			}},
			wantDomain: "test-domain",
			wantEntity: "test-entity",
			wantErr:    false,
		},
		{
			name: "non-string keys ignored",
			attrs: []any{map[string]any{
				"other_value":          "value",
				beholder.AttrKeyDomain: "test-domain",
				beholder.AttrKeyEntity: "test-entity",
			}, 123, "other_key"},
			wantDomain: "test-domain",
			wantEntity: "test-entity",
			wantErr:    false,
		},
		{
			name: "non-string values handled",
			attrs: []any{map[string]any{
				"other_key":            123,
				beholder.AttrKeyDomain: "test-domain",
				beholder.AttrKeyEntity: "test-entity",
			}},
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
			expectedError: "source/beholder_domain not found in provided key/value attributes",
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

func TestExtractAttributes(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name           string
		attrs          []any
		wantAttributes map[string]any
		wantErr        bool
		expectedError  string
	}{
		{
			name: "valid attributes with specific keys",
			attrs: []any{map[string]any{
				"datacontenttype": "application/protobuf",
				"dataschema":      "/schemas/ids/1001",
				"subject":         "example-subject",
				"time":            now,
				"recordedtime":    now,
			}},
			wantAttributes: map[string]any{
				"datacontenttype": "application/protobuf",
				"dataschema":      "/schemas/ids/1001",
				"subject":         "example-subject",
				"time":            now,
				"recordedtime":    now,
			},
			wantErr: false,
		},
		{
			name:           "happy path - empty attributes",
			attrs:          []any{},
			wantAttributes: map[string]any{},
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAttributes := beholder.ExtractAttributes(tt.attrs...)

			assert.Equal(t, tt.wantAttributes, gotAttributes)
		})
	}
}

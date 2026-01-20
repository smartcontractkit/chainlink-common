package errors_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	caperrors "github.com/smartcontractkit/chainlink-common/pkg/capabilities/errors"
)

func TestErrorProtoSerialization(t *testing.T) {
	// Assuming you have types: Visibility, Origin, and a custom Error type with serialization logic.
	visibilities := []caperrors.Visibility{caperrors.VisibilityPublic, caperrors.VisibilityPrivate}
	origins := []caperrors.Origin{caperrors.OriginUser, caperrors.OriginSystem}
	errorCodes := []caperrors.ErrorCode{caperrors.Unknown, caperrors.ConsensusFailed, caperrors.InvalidArgument}

	for _, v := range visibilities {
		for _, o := range origins {
			for _, c := range errorCodes {

				errMsg := "test error"
				originalErr := caperrors.NewError(errors.New(errMsg), v, o, c)
				protoErr := originalErr.ToProto()
				deserializedErr := caperrors.FromProto(protoErr)

				if v == caperrors.VisibilityPrivate {
					errMsg = "error whilst executing capability - the error message is not publicly reportable"
				}

				expectedErr := caperrors.NewError(errors.New(errMsg), v, o, c)

				if !expectedErr.Equals(deserializedErr) {
					t.Errorf("expected %v, got %v", expectedErr, deserializedErr)
				}
			}
		}
	}
}

func TestErrorSerializationAndDeserialization(t *testing.T) {
	// Assuming you have types: Visibility, Origin, and a custom Error type with serialization logic.
	visibilities := []caperrors.Visibility{caperrors.VisibilityPublic, caperrors.VisibilityPrivate}
	origins := []caperrors.Origin{caperrors.OriginUser, caperrors.OriginSystem}
	errorCodes := []caperrors.ErrorCode{caperrors.Unknown, caperrors.ConsensusFailed, caperrors.InvalidArgument}

	for _, v := range visibilities {
		for _, o := range origins {
			for _, c := range errorCodes {
				originalErr := caperrors.NewError(errors.New("test error"), v, o, c)
				serialized := originalErr.SerializeToString()
				deserializedErr := caperrors.DeserializeErrorFromString(serialized)
				if !originalErr.Equals(deserializedErr) {
					t.Errorf("expected %v, got %v", originalErr, deserializedErr)
				}
			}
		}
	}
}

func TestRemoteErrorSerializationAndDeserialization(t *testing.T) {
	// Assuming you have types: Visibility, Origin, and a custom Error type with serialization logic.
	visibilities := []caperrors.Visibility{caperrors.VisibilityPublic, caperrors.VisibilityPrivate}
	origins := []caperrors.Origin{caperrors.OriginUser, caperrors.OriginSystem}
	errorCodes := []caperrors.ErrorCode{caperrors.Unknown, caperrors.ConsensusFailed, caperrors.InvalidArgument}

	for _, v := range visibilities {
		for _, o := range origins {
			for _, c := range errorCodes {
				originalErr := caperrors.NewError(errors.New("test error"), v, o, c)
				serialized := originalErr.SerializeToRemoteString()
				deserializedErr := caperrors.DeserializeErrorFromString(serialized)
				if v == caperrors.VisibilityPrivate {
					require.Equal(t, deserializedErr.Visibility(), originalErr.Visibility())
					require.Equal(t, deserializedErr.Origin(), originalErr.Origin())
					require.Equal(t, deserializedErr.Code(), originalErr.Code())
					require.True(t, strings.Contains(deserializedErr.Error(), "error whilst executing capability - the error message is not publicly reportable"))
				} else {
					if !originalErr.Equals(deserializedErr) {
						t.Errorf("expected %v, got %v", originalErr, deserializedErr)
					}
				}
			}
		}
	}
}

// Check behaviour when deserializing the old error format to ensure backwards compatibility
func TestParsingOldErrorFormat(t *testing.T) {
	// Test deserialization of an old error format

	oldErrorMsgString := "failed to execute capability: some error occurred"
	deserializedErr := caperrors.DeserializeErrorFromString(oldErrorMsgString)

	expectedErr := caperrors.NewError(errors.New(oldErrorMsgString), caperrors.VisibilityPrivate, caperrors.OriginSystem, caperrors.Unknown)
	if !deserializedErr.Equals(expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, deserializedErr)
	}
}

// Check behaviour when deserializing messages with invalid visibility, origin, or error codes, also ensures that additional
// errors codes, visibility types and origin types added in future versions do not break deserialization.
func TestParsingWithInvalidVisibilityOriginAndErrorCodesAndBackwardsCompatibility(t *testing.T) {
	// Individually test each invalid message
	t.Run("InvalidVisibility", func(t *testing.T) {
		msg := "some error occurred"
		serializedError := "InvalidVisibility:User:Unknown:some error occurred"
		deserializedErr := caperrors.DeserializeErrorFromString(serializedError)
		expectedErr := caperrors.NewError(errors.New(msg), caperrors.Visibility(-1), caperrors.OriginUser, caperrors.Unknown)
		if !deserializedErr.Equals(expectedErr) {
			t.Errorf("expected %v, got %v", expectedErr, deserializedErr)
		}
	})

	t.Run("InvalidOrigin", func(t *testing.T) {
		msg := "some error occurred"
		serializedError := "Public:InvalidOrigin:Unknown:some error occurred"
		deserializedErr := caperrors.DeserializeErrorFromString(serializedError)
		expectedErr := caperrors.NewError(errors.New(msg), caperrors.VisibilityPublic, caperrors.Origin(-1), caperrors.Unknown)
		if !deserializedErr.Equals(expectedErr) {
			t.Errorf("expected %v, got %v", expectedErr, deserializedErr)
		}
	})

	t.Run("InvalidErrorCode", func(t *testing.T) {
		msg := "some error occurred"
		serializedError := "Public:System:InvalidErrorCode:some error occurred"
		deserializedErr := caperrors.DeserializeErrorFromString(serializedError)
		expectedErr := caperrors.NewError(errors.New(msg), caperrors.VisibilityPublic, caperrors.OriginSystem, caperrors.Unknown)
		if !deserializedErr.Equals(expectedErr) {
			t.Errorf("expected %v, got %v", expectedErr, deserializedErr)
		}
	})

	t.Run("InvalidMessageInsufficientFields", func(t *testing.T) {
		msg := "Public:System:Unknown"
		deserializedErr := caperrors.DeserializeErrorFromString(msg)

		expectedErr := caperrors.NewError(errors.New(msg), caperrors.VisibilityPrivate, caperrors.OriginSystem, caperrors.Unknown)
		if !deserializedErr.Equals(expectedErr) {
			t.Errorf("expected %v, got %v", expectedErr, deserializedErr)
		}
	})
}

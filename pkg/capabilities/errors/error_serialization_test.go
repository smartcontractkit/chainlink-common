package errors_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	caperrors "github.com/smartcontractkit/chainlink-common/pkg/capabilities/errors"
)

func Test_DeserializeFromString(t *testing.T) {
	testDeserialization := func(t *testing.T, serialisedRepresentation, expectedError string, expectedCode caperrors.ErrorCode, expectedReportType caperrors.ReportType) {
		err := caperrors.DeserializeErrorFromString(serialisedRepresentation)
		require.Equal(t, expectedError, err.Error())
		require.Equal(t, expectedCode, err.Code())
		require.Equal(t, expectedReportType, err.ReportType())
	}

	// Remote reportable errors
	testDeserialization(t,
		"RemoteReportableError:"+"some remote reportable error occurred",
		"[0]Uncategorized: some remote reportable error occurred",
		caperrors.Uncategorized,
		caperrors.RemoteReportable,
	)

	testDeserialization(t,
		"RemoteReportableError:ErrorCode=3:"+"some remote reportable error occurred",
		"[3]DeadlineExceeded: some remote reportable error occurred",
		caperrors.DeadlineExceeded,
		caperrors.RemoteReportable,
	)

	testDeserialization(t,
		"RemoteReportableError:ErrorCode=45:"+"some remote reportable error occurred",
		"[0]Uncategorized: some remote reportable error occurred",
		caperrors.Uncategorized,
		caperrors.RemoteReportable,
	)

	// User reportable errors
	testDeserialization(t,
		"RemoteReportableError:UserError:"+"some user reportable error occurred",
		"[0]Uncategorized: some user reportable error occurred",
		caperrors.Uncategorized,
		caperrors.ReportableUser,
	)

	testDeserialization(t,
		"RemoteReportableError:UserError:ErrorCode=4:"+"some user reportable error occurred",
		"[4]NotFound: some user reportable error occurred",
		caperrors.NotFound,
		caperrors.ReportableUser,
	)

	testDeserialization(t,
		"RemoteReportableError:UserError:ErrorCode=50:"+"some user reportable error occurred",
		"[0]Uncategorized: some user reportable error occurred",
		caperrors.Uncategorized,
		caperrors.ReportableUser,
	)

	// Local reportable errors
	testDeserialization(t,
		"LocalReportableError:"+"some local reportable error occurred",
		"[0]Uncategorized: some local reportable error occurred",
		caperrors.Uncategorized,
		caperrors.LocalOnly,
	)

	testDeserialization(t,
		"LocalReportableError:ErrorCode=5:"+"some local reportable error occurred",
		"[5]AlreadyExists: some local reportable error occurred",
		caperrors.AlreadyExists,
		caperrors.LocalOnly,
	)

	testDeserialization(t,
		"LocalReportableError:ErrorCode=-4:"+"some local reportable error occurred",
		"[0]Uncategorized: some local reportable error occurred",
		caperrors.Uncategorized,
		caperrors.LocalOnly,
	)

	// No identifier error - to ensure backwards compatibility with older versions that do not use the reporting type identifiers
	testDeserialization(t,
		"failed to execute capability",
		"[0]Uncategorized: failed to execute capability",
		caperrors.Uncategorized,
		caperrors.LocalOnly,
	)
}

func Test_SerializeToString(t *testing.T) {
	serializeAndAssert := func(t *testing.T, err caperrors.Error, expectedSerializedForm string) {
		serialized := err.SerializeToString()
		require.Equal(t, expectedSerializedForm, serialized)
	}

	serializeAndAssert(t,
		caperrors.NewRemoteReportableError(
			errors.New("some remote reportable error occurred"),
			caperrors.DeadlineExceeded,
		),
		"RemoteReportableError:ErrorCode=3:some remote reportable error occurred",
	)

	serializeAndAssert(t,
		caperrors.NewReportableUserError(
			errors.New("some user reportable error occurred"),
			caperrors.NotFound,
		),
		"RemoteReportableError:UserError:ErrorCode=4:some user reportable error occurred",
	)

	serializeAndAssert(t,
		caperrors.NewLocalReportableError(
			errors.New("some local reportable error occurred"),
			caperrors.AlreadyExists,
		),
		"LocalReportableError:ErrorCode=5:some local reportable error occurred",
	)
}

func Test_SerializeToRemoteReportableString(t *testing.T) {
	serializeToRemoteReportableStringAndAssert := func(t *testing.T, err caperrors.Error, expectedSerializedForm string) {
		serialized := err.SerializeToRemoteReportableString()
		require.Equal(t, expectedSerializedForm, serialized)
	}

	// Remote reportable error
	remoteReportableError := caperrors.NewRemoteReportableError(
		errors.New("some remote reportable error occurred"),
		caperrors.DeadlineExceeded,
	)
	serializeToRemoteReportableStringAndAssert(t, remoteReportableError, "RemoteReportableError:ErrorCode=3:some remote reportable error occurred")

	// User reportable error
	userReportableError := caperrors.NewReportableUserError(
		errors.New("some user reportable error occurred"),
		caperrors.NotFound,
	)
	serializeToRemoteReportableStringAndAssert(t, userReportableError, "RemoteReportableError:UserError:ErrorCode=4:some user reportable error occurred")

	// Local reportable error
	localReportableError := caperrors.NewLocalReportableError(
		errors.New("some local reportable error occurred"),
		caperrors.AlreadyExists,
	)
	serializeToRemoteReportableStringAndAssert(t, localReportableError, "LocalReportableError:ErrorCode=5: failed to execute capability - error message is not remotely reportable")
}

// Legacy format used before ReportableUser, LocalOnly types and Error Codes were introduced
func Test_DeserializeFromLegacyRemoteReportableString(t *testing.T) {
	assertDeserialization(t,
		fmt.Errorf("failed to execute capability: %w", errors.New("some error occurred")).Error(),
		"[0]Uncategorized: failed to execute capability: some error occurred",
		caperrors.Uncategorized,
		caperrors.LocalOnly,
	)
}

func assertDeserialization(t *testing.T, serialized string, expectedError string, expectedCode caperrors.ErrorCode, expectedReportType caperrors.ReportType) {
	err := caperrors.DeserializeErrorFromString(serialized)
	require.Equal(t, err.Error(), expectedError)
	require.Equal(t, err.Code(), expectedCode)
	require.Equal(t, err.ReportType(), expectedReportType)
}

func Test_DeserializeFromRemoteReportableString(t *testing.T) {
	// Remote reportable error
	assertDeserialization(t,
		"RemoteReportableError: failed to execute capability",
		"[0]Uncategorized:  failed to execute capability",
		caperrors.Uncategorized,
		caperrors.RemoteReportable,
	)

	// User reportable error
	assertDeserialization(t,
		"RemoteReportableError:UserError: failed to execute capability",
		"[0]Uncategorized:  failed to execute capability",
		caperrors.Uncategorized,
		caperrors.ReportableUser,
	)

	// Local reportable error
	assertDeserialization(t,
		"LocalReportableError:ErrorCode=5: failed to execute capability",
		"[5]AlreadyExists:  failed to execute capability",
		caperrors.AlreadyExists,
		caperrors.LocalOnly,
	)
}

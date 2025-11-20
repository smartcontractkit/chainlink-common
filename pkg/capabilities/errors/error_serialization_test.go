package errors_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	caperrors "github.com/smartcontractkit/chainlink-common/pkg/capabilities/errors"
)

func Test_DeserializeFromString(t *testing.T) {
	// Remote reportable errors
	remoteReportableErrorWithoutErrorCode := "RemoteReportableError:" + "some remote reportable error occurred"
	err := caperrors.DeserializeErrorFromString(remoteReportableErrorWithoutErrorCode)
	require.Equal(t, err.Error(), "[0]Uncategorized: some remote reportable error occurred")
	require.Equal(t, err.Code(), caperrors.Uncategorized)
	require.Equal(t, err.ReportType(), caperrors.ErrorReportTypeRemote)

	remoteReportableErrorWithErrorCode := "RemoteReportableError:ErrorCode=3:" + "some remote reportable error occurred"
	err = caperrors.DeserializeErrorFromString(remoteReportableErrorWithErrorCode)
	require.Equal(t, err.Error(), "[3]DeadlineExceeded: some remote reportable error occurred")
	require.Equal(t, err.Code(), caperrors.DeadlineExceeded)
	require.Equal(t, err.ReportType(), caperrors.ErrorReportTypeRemote)

	remoteReportableErrorWithErrorCodeThatDoesNotExistLocally := "RemoteReportableError:ErrorCode=45:" + "some remote reportable error occurred"
	err = caperrors.DeserializeErrorFromString(remoteReportableErrorWithErrorCodeThatDoesNotExistLocally)
	require.Equal(t, err.Error(), "[0]Uncategorized: some remote reportable error occurred")
	require.Equal(t, err.Code(), caperrors.Uncategorized)
	require.Equal(t, err.ReportType(), caperrors.ErrorReportTypeRemote)

	// User reportable errors
	userReportableErrorWithoutErrorCode := "RemoteReportableError:UserError:" + "some user reportable error occurred"
	err = caperrors.DeserializeErrorFromString(userReportableErrorWithoutErrorCode)
	require.Equal(t, err.Error(), "[0]Uncategorized: some user reportable error occurred")
	require.Equal(t, err.Code(), caperrors.Uncategorized)
	require.Equal(t, err.ReportType(), caperrors.ErrorReportTypeUser)

	userReportableErrorWithErrorCode := "RemoteReportableError:UserError:ErrorCode=4:" + "some user reportable error occurred"
	err = caperrors.DeserializeErrorFromString(userReportableErrorWithErrorCode)
	require.Equal(t, err.Error(), "[4]NotFound: some user reportable error occurred")
	require.Equal(t, err.Code(), caperrors.NotFound)
	require.Equal(t, err.ReportType(), caperrors.ErrorReportTypeUser)

	userReportableErrorWithErrorCodeThatDoesNotExistLocally := "RemoteReportableError:UserError:ErrorCode=50:" + "some user reportable error occurred"
	err = caperrors.DeserializeErrorFromString(userReportableErrorWithErrorCodeThatDoesNotExistLocally)
	require.Equal(t, err.Error(), "[0]Uncategorized: some user reportable error occurred")
	require.Equal(t, err.Code(), caperrors.Uncategorized)
	require.Equal(t, err.ReportType(), caperrors.ErrorReportTypeUser)

	// Local reportable errors
	localReportableErrorWithoutErrorCode := "LocalReportableError:" + "some local reportable error occurred"
	err = caperrors.DeserializeErrorFromString(localReportableErrorWithoutErrorCode)
	require.Equal(t, err.Error(), "[0]Uncategorized: some local reportable error occurred")
	require.Equal(t, err.Code(), caperrors.Uncategorized)
	require.Equal(t, err.ReportType(), caperrors.ErrorReportTypeLocal)

	localReportableErrorWithErrorCode := "LocalReportableError:ErrorCode=5:" + "some local reportable error occurred"
	err = caperrors.DeserializeErrorFromString(localReportableErrorWithErrorCode)
	require.Equal(t, err.Error(), "[5]AlreadyExists: some local reportable error occurred")
	require.Equal(t, err.Code(), caperrors.AlreadyExists)
	require.Equal(t, err.ReportType(), caperrors.ErrorReportTypeLocal)

	localReportableErrorWithErrorCodeThatDoesNotExistLocally := "LocalReportableError:ErrorCode=-4:" + "some local reportable error occurred"
	err = caperrors.DeserializeErrorFromString(localReportableErrorWithErrorCodeThatDoesNotExistLocally)
	require.Equal(t, err.Error(), "[0]Uncategorized: some local reportable error occurred")
	require.Equal(t, err.Code(), caperrors.Uncategorized)
	require.Equal(t, err.ReportType(), caperrors.ErrorReportTypeLocal)

	// No identifier error - to ensure backwards compatibility with older versions that do not use the reporting type identifiers
	noIdentifierError := "failed to execute capability"
	err = caperrors.DeserializeErrorFromString(noIdentifierError)
	require.Equal(t, err.Error(), "[0]Uncategorized: failed to execute capability")
	require.Equal(t, err.Code(), caperrors.Uncategorized)
	require.Equal(t, err.ReportType(), caperrors.ErrorReportTypeLocal)

}

func Test_SerializeToString(t *testing.T) {
	// Remote reportable error
	remoteReportableError := caperrors.NewRemoteReportableError(
		errors.New("some remote reportable error occurred"),
		caperrors.DeadlineExceeded,
	)
	serialized := remoteReportableError.SerializeToString()
	expectedSerialized := "RemoteReportableError:ErrorCode=3:some remote reportable error occurred"
	require.Equal(t, expectedSerialized, serialized)

	// User reportable error
	userReportableError := caperrors.NewReportableUserError(
		errors.New("some user reportable error occurred"),
		caperrors.NotFound,
	)
	serialized = userReportableError.SerializeToString()
	expectedSerialized = "RemoteReportableError:UserError:ErrorCode=4:some user reportable error occurred"
	require.Equal(t, expectedSerialized, serialized)

	// Local reportable error
	localReportableError := caperrors.NewLocalReportableError(
		errors.New("some local reportable error occurred"),
		caperrors.AlreadyExists,
	)
	serialized = localReportableError.SerializeToString()
	expectedSerialized = "LocalReportableError:ErrorCode=5:some local reportable error occurred"
	require.Equal(t, expectedSerialized, serialized)
}

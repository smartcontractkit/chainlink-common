package capabilities_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

func Test_DeserializeFromString(t *testing.T) {
	// Remote reportable errors
	remoteReportableErrorWithoutErrorCode := "RemoteReportableError:" + "some remote reportable error occurred"
	err := capabilities.DeserializeErrorFromString(remoteReportableErrorWithoutErrorCode)
	require.Equal(t, err.Error(), "[0]Uncategorized: some remote reportable error occurred")
	require.Equal(t, err.Code(), capabilities.Uncategorized)
	require.Equal(t, err.ReportType(), capabilities.ErrorReportTypeRemote)

	remoteReportableErrorWithErrorCode := "RemoteReportableError:ErrorCode=3:" + "some remote reportable error occurred"
	err = capabilities.DeserializeErrorFromString(remoteReportableErrorWithErrorCode)
	require.Equal(t, err.Error(), "[3]DeadlineExceeded: some remote reportable error occurred")
	require.Equal(t, err.Code(), capabilities.DeadlineExceeded)
	require.Equal(t, err.ReportType(), capabilities.ErrorReportTypeRemote)

	remoteReportableErrorWithErrorCodeThatDoesNotExistLocally := "RemoteReportableError:ErrorCode=45:" + "some remote reportable error occurred"
	err = capabilities.DeserializeErrorFromString(remoteReportableErrorWithErrorCodeThatDoesNotExistLocally)
	require.Equal(t, err.Error(), "[0]Uncategorized: some remote reportable error occurred")
	require.Equal(t, err.Code(), capabilities.Uncategorized)
	require.Equal(t, err.ReportType(), capabilities.ErrorReportTypeRemote)

	// User reportable errors
	userReportableErrorWithoutErrorCode := "RemoteReportableError:UserError:" + "some user reportable error occurred"
	err = capabilities.DeserializeErrorFromString(userReportableErrorWithoutErrorCode)
	require.Equal(t, err.Error(), "[0]Uncategorized: some user reportable error occurred")
	require.Equal(t, err.Code(), capabilities.Uncategorized)
	require.Equal(t, err.ReportType(), capabilities.ErrorReportTypeUser)

	userReportableErrorWithErrorCode := "RemoteReportableError:UserError:ErrorCode=4:" + "some user reportable error occurred"
	err = capabilities.DeserializeErrorFromString(userReportableErrorWithErrorCode)
	require.Equal(t, err.Error(), "[4]NotFound: some user reportable error occurred")
	require.Equal(t, err.Code(), capabilities.NotFound)
	require.Equal(t, err.ReportType(), capabilities.ErrorReportTypeUser)

	userReportableErrorWithErrorCodeThatDoesNotExistLocally := "RemoteReportableError:UserError:ErrorCode=50:" + "some user reportable error occurred"
	err = capabilities.DeserializeErrorFromString(userReportableErrorWithErrorCodeThatDoesNotExistLocally)
	require.Equal(t, err.Error(), "[0]Uncategorized: some user reportable error occurred")
	require.Equal(t, err.Code(), capabilities.Uncategorized)
	require.Equal(t, err.ReportType(), capabilities.ErrorReportTypeUser)

	// Local reportable errors
	localReportableErrorWithoutErrorCode := "LocalReportableError:" + "some local reportable error occurred"
	err = capabilities.DeserializeErrorFromString(localReportableErrorWithoutErrorCode)
	require.Equal(t, err.Error(), "[0]Uncategorized: some local reportable error occurred")
	require.Equal(t, err.Code(), capabilities.Uncategorized)
	require.Equal(t, err.ReportType(), capabilities.ErrorReportTypeLocal)

	localReportableErrorWithErrorCode := "LocalReportableError:ErrorCode=5:" + "some local reportable error occurred"
	err = capabilities.DeserializeErrorFromString(localReportableErrorWithErrorCode)
	require.Equal(t, err.Error(), "[5]AlreadyExists: some local reportable error occurred")
	require.Equal(t, err.Code(), capabilities.AlreadyExists)
	require.Equal(t, err.ReportType(), capabilities.ErrorReportTypeLocal)

	localReportableErrorWithErrorCodeThatDoesNotExistLocally := "LocalReportableError:ErrorCode=-4:" + "some local reportable error occurred"
	err = capabilities.DeserializeErrorFromString(localReportableErrorWithErrorCodeThatDoesNotExistLocally)
	require.Equal(t, err.Error(), "[0]Uncategorized: some local reportable error occurred")
	require.Equal(t, err.Code(), capabilities.Uncategorized)
	require.Equal(t, err.ReportType(), capabilities.ErrorReportTypeLocal)

	// No identifier error - to ensure backwards compatibility with older versions that do not use the reporting type identifiers
	noIdentifierError := "failed to execute capability"
	err = capabilities.DeserializeErrorFromString(noIdentifierError)
	require.Equal(t, err.Error(), "[0]Uncategorized: failed to execute capability")
	require.Equal(t, err.Code(), capabilities.Uncategorized)
	require.Equal(t, err.ReportType(), capabilities.ErrorReportTypeLocal)

}

func Test_SerializeToString(t *testing.T) {
	// Remote reportable error
	remoteReportableError := capabilities.NewRemoteReportableError(
		errors.New("some remote reportable error occurred"),
		capabilities.DeadlineExceeded,
	)
	serialized := remoteReportableError.SerializeToString()
	expectedSerialized := "RemoteReportableError:ErrorCode=3:some remote reportable error occurred"
	require.Equal(t, expectedSerialized, serialized)

	// User reportable error
	userReportableError := capabilities.NewReportableUserError(
		errors.New("some user reportable error occurred"),
		capabilities.NotFound,
	)
	serialized = userReportableError.SerializeToString()
	expectedSerialized = "RemoteReportableError:UserError:ErrorCode=4:some user reportable error occurred"
	require.Equal(t, expectedSerialized, serialized)

	// Local reportable error
	localReportableError := capabilities.NewLocalReportableError(
		errors.New("some local reportable error occurred"),
		capabilities.AlreadyExists,
	)
	serialized = localReportableError.SerializeToString()
	expectedSerialized = "LocalReportableError:ErrorCode=5:some local reportable error occurred"
	require.Equal(t, expectedSerialized, serialized)
}

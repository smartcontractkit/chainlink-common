package capabilities_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

func Test_Don2DonToError(t *testing.T) {
	// Remote reportable errors
	remoteReportableErrorWithoutErrorCode := "RemoteReportableError:" + "some remote reportable error occurred"
	err := capabilities.ToCapabilityError(remoteReportableErrorWithoutErrorCode)
	require.Equal(t, err.Error(), "[0]Uncategorized: some remote reportable error occurred")
	require.Equal(t, err.Code(), capabilities.Uncategorized)
	require.Equal(t, err.ReportType(), capabilities.ErrorReportTypeRemote)

	remoteReportableErrorWithErrorCode := "RemoteReportableError:ErrorCode=3:" + "some remote reportable error occurred"
	err = capabilities.ToCapabilityError(remoteReportableErrorWithErrorCode)
	require.Equal(t, err.Error(), "[3]DeadlineExceeded: some remote reportable error occurred")
	require.Equal(t, err.Code(), capabilities.DeadlineExceeded)
	require.Equal(t, err.ReportType(), capabilities.ErrorReportTypeRemote)

	remoteReportableErrorWithErrorCodeThatDoesNotExistLocally := "RemoteReportableError:ErrorCode=45:" + "some remote reportable error occurred"
	err = capabilities.ToCapabilityError(remoteReportableErrorWithErrorCodeThatDoesNotExistLocally)
	require.Equal(t, err.Error(), "[0]Uncategorized: some remote reportable error occurred")
	require.Equal(t, err.Code(), capabilities.Uncategorized)
	require.Equal(t, err.ReportType(), capabilities.ErrorReportTypeRemote)

	// User reportable errors
	userReportableErrorWithoutErrorCode := "RemoteReportableError:UserError:" + "some user reportable error occurred"
	err = capabilities.ToCapabilityError(userReportableErrorWithoutErrorCode)
	require.Equal(t, err.Error(), "[0]Uncategorized: some user reportable error occurred")
	require.Equal(t, err.Code(), capabilities.Uncategorized)
	require.Equal(t, err.ReportType(), capabilities.ErrorReportTypeUser)

	userReportableErrorWithErrorCode := "RemoteReportableError:UserError:ErrorCode=4:" + "some user reportable error occurred"
	err = capabilities.ToCapabilityError(userReportableErrorWithErrorCode)
	require.Equal(t, err.Error(), "[4]NotFound: some user reportable error occurred")
	require.Equal(t, err.Code(), capabilities.NotFound)
	require.Equal(t, err.ReportType(), capabilities.ErrorReportTypeUser)

	userReportableErrorWithErrorCodeThatDoesNotExistLocally := "RemoteReportableError:UserError:ErrorCode=50:" + "some user reportable error occurred"
	err = capabilities.ToCapabilityError(userReportableErrorWithErrorCodeThatDoesNotExistLocally)
	require.Equal(t, err.Error(), "[0]Uncategorized: some user reportable error occurred")
	require.Equal(t, err.Code(), capabilities.Uncategorized)
	require.Equal(t, err.ReportType(), capabilities.ErrorReportTypeUser)

	// Local reportable errors
	localReportableErrorWithoutErrorCode := "LocalReportableError:" + "some local reportable error occurred"
	err = capabilities.ToCapabilityError(localReportableErrorWithoutErrorCode)
	require.Equal(t, err.Error(), "[0]Uncategorized: some local reportable error occurred")
	require.Equal(t, err.Code(), capabilities.Uncategorized)
	require.Equal(t, err.ReportType(), capabilities.ErrorReportTypeLocal)

	localReportableErrorWithErrorCode := "LocalReportableError:ErrorCode=5:" + "some local reportable error occurred"
	err = capabilities.ToCapabilityError(localReportableErrorWithErrorCode)
	require.Equal(t, err.Error(), "[5]AlreadyExists: some local reportable error occurred")
	require.Equal(t, err.Code(), capabilities.AlreadyExists)
	require.Equal(t, err.ReportType(), capabilities.ErrorReportTypeLocal)

	localReportableErrorWithErrorCodeThatDoesNotExistLocally := "LocalReportableError:ErrorCode=-4:" + "some local reportable error occurred"
	err = capabilities.ToCapabilityError(localReportableErrorWithErrorCodeThatDoesNotExistLocally)
	require.Equal(t, err.Error(), "[0]Uncategorized: some local reportable error occurred")
	require.Equal(t, err.Code(), capabilities.Uncategorized)
	require.Equal(t, err.ReportType(), capabilities.ErrorReportTypeLocal)

	// No identifier error - to ensure backwards compatibility with older versions that do not use the reporting type identifiers
	noIdentifierError := "failed to execute capability"
	err = capabilities.ToCapabilityError(noIdentifierError)
	require.Equal(t, err.Error(), "[0]Uncategorized: failed to execute capability")
	require.Equal(t, err.Code(), capabilities.Uncategorized)
	require.Equal(t, err.ReportType(), capabilities.ErrorReportTypeLocal)

}

func Test_ErrorToDon2Don(t *testing.T) {
	// Remote reportable error
	remoteReportableError := capabilities.NewRemoteReportableError(
		capabilities.NewErrorf("some remote reportable error occurred"),
		capabilities.DeadlineExceeded,
	)
	remoteReportableErrorStr := capabilities.ErrorToDon2Don(remoteReportableError)
	require.Equal(t,
		"RemoteReportableError:ErrorCode=3:some remote reportable error occurred",
		remoteReportableErrorStr,
	)

	// User reportable error
	userReportableError := capabilities.NewUserError(
		capabilities.NewErrorf("some user reportable error occurred"),
		capabilities.NotFound,
	)
	userReportableErrorStr := capabilities.ErrorToDon2Don(userReportableError)
	require.Equal(t,
		"RemoteReportableError:UserError:ErrorCode=4:some user reportable error occurred",
		userReportableErrorStr,
	)

	// Local reportable error
	localReportableError := capabilities.NewLocalReportableError(
		capabilities.NewErrorf("some local reportable error occurred"),
		capabilities.AlreadyExists,
	)
	localReportableErrorStr := capabilities.ErrorToDon2Don(localReportableError)
	require.Equal(t,
		"LocalReportableError:ErrorCode=5:some local reportable error occurred",
		localReportableErrorStr,
	)

}

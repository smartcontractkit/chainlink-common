package servicetest

import (
	"github.com/stretchr/testify/mock"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// Mock is implemented by mock generated [services.Service], due to the embedded [mock.Mock].
type Mock interface {
	services.Service
	On(methodName string, arguments ...interface{}) *mock.Call
	String() string
}

// SetupNoOpMock registers possible no-op methods calls on a mocked [services.Service].
func SetupNoOpMock(srv Mock) {
	srv.On("Start", mock.Anything).Return(nil).Maybe()
	srv.On("Close").Return(nil).Maybe()
	srv.On("Name").Return(srv.String()).Maybe()
	srv.On("Ready").Return(nil).Maybe()
	srv.On("HealthReport").Return(map[string]error{srv.String(): nil}, nil).Maybe()
}

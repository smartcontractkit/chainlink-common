package test

import (
	"context"

	testreportingplugin "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/ocr2/reporting_plugin"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var ExecFactoryServer = execFactoryServer{
	execFactoryServerConfig: execFactoryServerConfig{
		provider: ExecProvider,
	},
}

type execFactoryServerConfig struct {
	provider ExecProviderEvaluator
}

var _ types.CCIPExecutionFactoryGenerator = execFactoryServer{}

type execFactoryServer struct {
	execFactoryServerConfig
}

// NewExecutionFactory implements types.CCIPExecFactoryGenerator.
// func (e execFactoryServer) NewExecutionFactory(ctx context.Context, provider types.CCIPExecProvider, config types.CCIPExecFactoryGeneratorConfig) (types.ReportingPluginFactory, error) {
func (e execFactoryServer) NewExecutionFactory(ctx context.Context, provider types.CCIPExecProvider) (types.ReportingPluginFactory, error) {

	err := e.provider.Evaluate(ctx, provider)
	if err != nil {
		return nil, err
	}
	return testreportingplugin.Factory, nil
}

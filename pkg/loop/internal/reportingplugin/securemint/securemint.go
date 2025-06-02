package securemint

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/smartcontractkit/grpc-proxy/proxy"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/errorlog"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/reportingplugin/ocr2"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

// TODO(gg): maybe create a separate one for secure mint plugin here?

var _ core.PluginSecureMint = (*PluginSecureMintClient)(nil)

type PluginSecureMintClient struct {
	*goplugin.PluginClient
	*goplugin.ServiceClient

	secureMint pb.PluginSecureMintClient
}

func NewPluginSecureMintClient(brokerCfg net.BrokerConfig) *PluginSecureMintClient {
	brokerCfg.Logger = logger.Named(brokerCfg.Logger, "PluginSecureMintClient")
	pc := goplugin.NewPluginClient(brokerCfg)
	return &PluginSecureMintClient{PluginClient: pc, secureMint: pb.NewPluginSecureMintClient(pc), ServiceClient: goplugin.NewServiceClient(pc.BrokerExt, pc)}
}

// TODO(gg): update this signature, no need for all params probably
func (m *PluginSecureMintClient) NewSecureMintFactory(ctx context.Context, provider types.MedianProvider, contractID string, dataSource, juelsPerFeeCoin, gasPriceSubunits median.DataSource, errorLog core.ErrorLog, deviationFuncDefinition map[string]any) (types.ReportingPluginFactory, error) {
	cc := m.NewClientConn("SecureMintPluginFactory", func(ctx context.Context) (id uint32, deps net.Resources, err error) {

		// TODO(gg): not sure atm about data source usage, todo

		// dataSourceID, dsRes, err := m.ServeNew("DataSource", func(s *grpc.Server) {
		// 	pb.RegisterDataSourceServer(s, newDataSourceServer(dataSource))
		// })
		// if err != nil {
		// 	return 0, nil, err
		// }
		// deps.Add(dsRes)

		// juelsPerFeeCoinDataSourceID, juelsPerFeeCoinDataSourceRes, err := m.ServeNew("JuelsPerFeeCoinDataSource", func(s *grpc.Server) {
		// 	pb.RegisterDataSourceServer(s, newDataSourceServer(juelsPerFeeCoin))
		// })
		// if err != nil {
		// 	return 0, nil, err
		// }
		// deps.Add(juelsPerFeeCoinDataSourceRes)

		// gasPriceSubunitsDataSourceID, gasPriceSubunitsDataSourceRes, err := m.ServeNew("GasPriceSubunitsDataSource", func(s *grpc.Server) {
		// 	pb.RegisterDataSourceServer(s, newDataSourceServer(gasPriceSubunits))
		// })
		// if err != nil {
		// 	return 0, nil, err
		// }
		// deps.Add(gasPriceSubunitsDataSourceRes)

		var (
			providerID  uint32
			providerRes net.Resource
		)
		if grpcProvider, ok := provider.(goplugin.GRPCClientConn); ok {
			providerID, providerRes, err = m.Serve("SecureMintProvider", proxy.NewProxy(grpcProvider.ClientConn()))
		} else {
			err = fmt.Errorf("secureMintProvider is not a goplugin.GRPCClientConn, this part is not implemented")
		}
		if err != nil {
			return 0, nil, err
		}
		deps.Add(providerRes)

		errorLogID, errorLogRes, err := m.ServeNew("ErrorLog", func(s *grpc.Server) {
			pb.RegisterErrorLogServer(s, errorlog.NewServer(errorLog))
		})
		if err != nil {
			return 0, nil, err
		}
		deps.Add(errorLogRes)

		reply, err := m.secureMint.NewSecureMintFactory(ctx, &pb.NewSecureMintFactoryRequest{
			SecureMintProviderID: providerID,
			ErrorLogID:           errorLogID,

			// TODO(gg): add more params here when needed
			// ContractID:                   contractID,
			// DataSourceID:                 dataSourceID,
			// JuelsPerFeeCoinDataSourceID:  juelsPerFeeCoinDataSourceID,
			// GasPriceSubunitsDataSourceID: gasPriceSubunitsDataSourceID,
			// DeviationFuncDefinition:      deviationFuncDefinitionJSON,
		})
		if err != nil {
			return 0, nil, err
		}
		return reply.ReportingPluginFactoryID, nil, nil
	})
	// TODO(gg): or should we use a securemint-specific plugin factory client?
	return ocr2.NewReportingPluginFactoryClient(m.PluginClient.BrokerExt, cc), nil
}

var _ pb.PluginSecureMintServer = (*pluginSecureMintServer)(nil)

type pluginSecureMintServer struct {
	pb.UnimplementedPluginSecureMintServer

	*net.BrokerExt
	impl core.PluginSecureMint
}

func RegisterPluginSecureMintServer(server *grpc.Server, broker net.Broker, brokerCfg net.BrokerConfig, impl core.PluginSecureMint) error {
	pb.RegisterServiceServer(server, &goplugin.ServiceServer{Srv: impl})
	pb.RegisterPluginSecureMintServer(server, newPluginSecureMintServer(&net.BrokerExt{Broker: broker, BrokerConfig: brokerCfg}, impl))
	return nil
}

func newPluginSecureMintServer(b *net.BrokerExt, mp core.PluginSecureMint) *pluginSecureMintServer {
	return &pluginSecureMintServer{BrokerExt: b.WithName("PluginSecureMint"), impl: mp}
}

func (m *pluginSecureMintServer) NewSecureMintFactory(ctx context.Context, request *pb.NewSecureMintFactoryRequest) (*pb.NewSecureMintFactoryReply, error) {
	// dsConn, err := m.Dial(request.DataSourceID)
	// if err != nil {
	// 	return nil, net.ErrConnDial{Name: "DataSource", ID: request.DataSourceID, Err: err}
	// }
	// dsRes := net.Resource{Closer: dsConn, Name: "DataSource"}
	// dataSource := newDataSourceClient(dsConn)

	providerConn, err := m.Dial(request.SecureMintProviderID)
	if err != nil {
		// m.CloseAll(dsRes, juelsRes, gasPriceSubunitsRes)
		return nil, net.ErrConnDial{Name: "SecureMintProvider", ID: request.SecureMintProviderID, Err: err}
	}
	providerRes := net.Resource{Closer: providerConn, Name: "SecureMintProvider"}
	provider := securemint.NewProviderClient(m.BrokerExt, providerConn)
	provider.RmUnimplemented(ctx)

	errorLogConn, err := m.Dial(request.ErrorLogID)
	if err != nil {
		m.CloseAll(dsRes, juelsRes, gasPriceSubunitsRes, providerRes)
		return nil, net.ErrConnDial{Name: "ErrorLog", ID: request.ErrorLogID, Err: err}
	}
	errorLogRes := net.Resource{Closer: errorLogConn, Name: "ErrorLog"}
	errorLog := errorlog.NewClient(errorLogConn)

	var deviationFuncDefinition map[string]any
	if len(request.DeviationFuncDefinition) > 0 {
		if err = json.Unmarshal(request.DeviationFuncDefinition, deviationFuncDefinition); err != nil {
			m.CloseAll(dsRes, juelsRes, gasPriceSubunitsRes, providerRes, errorLogRes)
			return nil, fmt.Errorf("failed to unmarshal deviationFuncDefinition: %w", err)
		}
	}

	factory, err := m.impl.NewMedianFactory(ctx, provider, request.ContractID, dataSource, juelsPerFeeCoin, gasPriceSubunits, errorLog, deviationFuncDefinition)
	if err != nil {
		m.CloseAll(dsRes, juelsRes, gasPriceSubunitsRes, providerRes, errorLogRes)
		return nil, err
	}

	id, _, err := m.ServeNew("ReportingPluginProvider", func(s *grpc.Server) {
		pb.RegisterServiceServer(s, &goplugin.ServiceServer{Srv: factory})
		pb.RegisterReportingPluginFactoryServer(s, ocr2.NewReportingPluginFactoryServer(factory, m.BrokerExt))
	}, dsRes, juelsRes, gasPriceSubunitsRes, providerRes, errorLogRes)
	if err != nil {
		return nil, err
	}

	return &pb.NewMedianFactoryReply{ReportingPluginFactoryID: id}, nil
}

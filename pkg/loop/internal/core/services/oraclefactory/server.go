package oraclefactory

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	oraclesrv "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/oracle"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/reportingplugin/ocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	oraclepb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/oracle"
	oraclefactorypb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/oraclefactory"
	ocr3relayer "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

var _ oraclefactorypb.OracleFactoryServer = (*server)(nil)

type server struct {
	oraclefactorypb.UnimplementedOracleFactoryServer

	broker    *net.BrokerExt
	impl      core.OracleFactory
	log       logger.Logger
	resources net.Resources

	Name string
}

func NewServer(log logger.Logger, impl core.OracleFactory, broker *net.BrokerExt) (*server, net.Resource) {
	name := "OracleFactoryServer"
	newServer := &server{
		log:       log,
		impl:      impl,
		broker:    broker.WithName(name),
		resources: make(net.Resources, 0),
	}

	return newServer, net.Resource{
		Name:   name,
		Closer: newServer,
	}
}

func (s *server) Close() error {
	s.broker.CloseAll(s.resources...)
	return nil
}

func (s *server) NewOracle(ctx context.Context, req *oraclefactorypb.NewOracleRequest) (*oraclefactorypb.NewOracleReply, error) {
	var resources []net.Resource

	serviceName := "ReportingPluginFactory"
	reportingPluginFactoryServiceConn, err := s.broker.Dial(req.ReportingPluginFactoryServiceId)
	if err != nil {
		return nil, fmt.Errorf("error dialing %s service: %w", serviceName, err)
	}
	resources = append(resources, net.Resource{
		Closer: reportingPluginFactoryServiceConn,
		Name:   serviceName,
	})

	serviceName = "ContractTransmitter"
	contractTransmitterConn, err := s.broker.Dial(req.ContractTransmitterId)
	if err != nil {
		return nil, fmt.Errorf("error dialing %s service: %w", serviceName, err)
	}
	resources = append(resources, net.Resource{
		Closer: contractTransmitterConn,
		Name:   serviceName,
	})

	args := core.OracleArgs{
		LocalConfig: types.LocalConfig{
			BlockchainTimeout:                  req.LocalConfig.BlockchainTimeout.AsDuration(),
			ContractConfigConfirmations:        uint16(req.LocalConfig.ContractConfigConfirmations), // #nosec G115
			SkipContractConfigConfirmations:    req.LocalConfig.SkipContractConfigConfirmations,
			ContractConfigTrackerPollInterval:  req.LocalConfig.ContractConfigTrackerPollInterval.AsDuration(),
			ContractTransmitterTransmitTimeout: req.LocalConfig.ContractTransmitterTransmitTimeout.AsDuration(),
			DatabaseTimeout:                    req.LocalConfig.DatabaseTimeout.AsDuration(),
			MinOCR2MaxDurationQuery:            req.LocalConfig.MinOcr2MaxDurationQuery.AsDuration(),
			DevelopmentMode:                    req.LocalConfig.DevelopmentMode,
		},
		ReportingPluginFactoryService: ocr3.NewReportingPluginFactoryClient(
			s.broker,
			reportingPluginFactoryServiceConn,
		),
		ContractTransmitter: ocr3relayer.NewContractTransmitterClient(s.broker, contractTransmitterConn),
	}

	oracle, err := s.impl.NewOracle(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("NewOracle call failed: %w", err)
	}

	oracleServer, oracleServerRes := oraclesrv.NewServer(s.log, oracle, s.broker)
	resources = append(resources, oracleServerRes)
	oracleID, oracleRes, err := s.broker.ServeNew("Oracle", func(gs *grpc.Server) {
		oraclepb.RegisterOracleServer(gs, oracleServer)
	})
	if err != nil {
		s.broker.CloseAll(resources...)
		return nil, fmt.Errorf("failed to serve new oracle: %w", err)
	}
	resources = append(resources, oracleRes)

	s.resources = append(s.resources, resources...)
	return &oraclefactorypb.NewOracleReply{OracleId: oracleID}, nil
}

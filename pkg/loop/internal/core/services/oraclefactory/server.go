package oraclefactory

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	oracle "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/oracle"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/reportingplugin/ocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	oraclepb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/oracle"
	oraclefactorypb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/oraclefactory"
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

	reportingPluginFactoryServiceConn, err := s.broker.Dial(req.ReportingPluginFactoryServiceId)
	if err != nil {
		return nil, fmt.Errorf("error dialing reporting plugin factory service: %w", err)
	}
	resources = append(resources, net.Resource{
		Closer: reportingPluginFactoryServiceConn,
		Name:   "ReportingPluginFactoryServiceConn",
	})

	// From capabiliteis binary:
	// oracleFactory.Neworacle(ctx, args)
	// Args will be actual implementations (functions)

	// My NewOracle function on the client:
	// Mounts the objects using ServeNew and returns the connection ID
	// ConnectionID is sent over the wire.

	args := core.OracleArgs{
		LocalConfig: types.LocalConfig{
			BlockchainTimeout:                  req.LocalConfig.BlockchainTimeout.AsDuration(),
			ContractConfigConfirmations:        uint16(req.LocalConfig.ContractConfigConfirmations),
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
		// ContractConfigTracker:         req.ContractConfigTrackerId,
		// ContractTransmitter:           req.ContractTransmitter,
		// OffchainConfigDigester:        req.OffchainConfigDigester,
	}

	// I will need to implement the NewOracle function on the Core Node
	// In generics package this should be a lightweight wrapper generic.NewOracle(ctx, args)
	// This implementation will live in chainlink repo
	// This will return an oracle. I need to wrap that in a Oracle Client and Server. Reverse principle.
	// In this case, the client is the capabilities binary and the server is the core node.
	oracleImpl, err := s.impl.NewOracle(ctx, args)
	oracleServer, oracleRes := oracle.NewServer(s.log, oracleImpl, s.broker)
	resources = append(resources, oracleRes)
	oracleID, oracleRes, err := s.broker.ServeNew("Oracle", func(gs *grpc.Server) {
		oraclepb.RegisterOracleServer(gs, oracleServer)
	})
	if err != nil {
		s.broker.CloseAll(resources...)
		return nil, fmt.Errorf("failed to serve new oracle: %w", err)
	}

	s.resources = append(s.resources, resources...)
	return &oraclefactorypb.NewOracleReply{OracleId: oracleID}, nil
}

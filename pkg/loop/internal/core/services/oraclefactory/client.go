package oraclefactory

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/oracle"
	reportingplugin "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/reportingplugin/ocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	ocr3pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ocr3"
	oraclefactorypb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/oraclefactory"
	ocr3relayer "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

var _ core.OracleFactory = (*client)(nil)

type client struct {
	broker        *net.BrokerExt
	grpc          oraclefactorypb.OracleFactoryClient
	log           logger.Logger
	resources     []net.Resource
	serviceClient *goplugin.ServiceClient
}

func NewClient(log logger.Logger, b *net.BrokerExt, conn grpc.ClientConnInterface) *client {
	b = b.WithName("OracleFactoryClient")
	return &client{
		log:           log,
		broker:        b,
		serviceClient: goplugin.NewServiceClient(b, conn),
		grpc:          oraclefactorypb.NewOracleFactoryClient(conn)}
}

func (c *client) NewOracle(ctx context.Context, oracleArgs core.OracleArgs) (core.Oracle, error) {
	var resources []net.Resource

	serviceName := "ReportingPluginFactoryServer"
	reportingPluginFactoryServerID, reportingPluginFactoryServerRes, err := c.broker.ServeNew(
		serviceName, func(gs *grpc.Server) {
			ocr3pb.RegisterReportingPluginFactoryServer(gs, reportingplugin.NewReportingPluginFactoryServer(
				oracleArgs.ReportingPluginFactoryService,
				c.broker,
			))
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to serve new %s: %w", serviceName, err)
	}
	resources = append(resources, reportingPluginFactoryServerRes)

	serviceName = "ContractTransmitterServer"
	contractTransmitterServerID, contractTransmitterServerRes, err := c.broker.ServeNew(
		serviceName, func(gs *grpc.Server) {
			ocr3pb.RegisterContractTransmitterServer(gs, ocr3relayer.NewContractTransmitterServer(
				oracleArgs.ContractTransmitter,
			))
		},
	)
	if err != nil {
		c.broker.CloseAll(resources...)
		return nil, fmt.Errorf("failed to serve new %s: %w", serviceName, err)
	}
	resources = append(resources, contractTransmitterServerRes)

	newOracleRequest := oraclefactorypb.NewOracleRequest{
		LocalConfig: &oraclefactorypb.LocalConfig{
			BlockchainTimeout:                  durationpb.New(oracleArgs.LocalConfig.BlockchainTimeout),
			ContractConfigConfirmations:        uint32(oracleArgs.LocalConfig.ContractConfigConfirmations),
			SkipContractConfigConfirmations:    oracleArgs.LocalConfig.SkipContractConfigConfirmations,
			ContractConfigTrackerPollInterval:  durationpb.New(oracleArgs.LocalConfig.ContractConfigTrackerPollInterval),
			ContractTransmitterTransmitTimeout: durationpb.New(oracleArgs.LocalConfig.ContractTransmitterTransmitTimeout),
			DatabaseTimeout:                    durationpb.New(oracleArgs.LocalConfig.DatabaseTimeout),
			MinOcr2MaxDurationQuery:            durationpb.New(oracleArgs.LocalConfig.MinOCR2MaxDurationQuery),
			DevelopmentMode:                    oracleArgs.LocalConfig.DevelopmentMode,
		},
		ReportingPluginFactoryServiceId: reportingPluginFactoryServerID,
		ContractTransmitterId:           contractTransmitterServerID,
	}

	newOracleReply, err := c.grpc.NewOracle(ctx, &newOracleRequest)
	if err != nil {
		c.broker.CloseAll(resources...)
		return nil, fmt.Errorf("error getting new oracle: %w", err)
	}

	oracleClientConn, err := c.broker.Dial(newOracleReply.OracleId)
	if err != nil {
		c.broker.CloseAll(resources...)
		return nil, fmt.Errorf("error dialing reporting plugin factory service: %w", err)
	}
	resources = append(resources, net.Resource{
		Closer: oracleClientConn,
		Name:   "OracleClientConn",
	})

	c.resources = append(c.resources, resources...)
	return oracle.NewClient(oracleClientConn), nil
}

package chipingress

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc/credentials/insecure"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	ceformat "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	cepb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	ce "github.com/cloudevents/sdk-go/v2"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
)

type ChipIngressClient interface {
	Ping(ctx context.Context) (string, error)
	Publish(ctx context.Context, event ce.Event) (*pb.PublishResponse, error)
	PublishBatch(ctx context.Context, events []ce.Event) (*pb.PublishResponse, error)
	Close() error
}

type chipIngressClient struct {
	Address string

	client pb.ChipIngressClient
	conn   *grpc.ClientConn
	log    *zap.Logger
}

// Opt is a function that configures a ChipIngressClient
type Opt func(*chipIngressClientConfig)

// chipIngressClientConfig is the configuration for the ChipIngressClient.
type chipIngressClientConfig struct {
	log                  *zap.Logger
	transportCredentials credentials.TransportCredentials
	headers              map[string]string
}

// NewChipIngressClient creates a new client for the Chip Ingress service with optional configuration.
func NewChipIngressClient(address string, opts ...Opt) (ChipIngressClient, error) {

	if address == "" {
		return nil, fmt.Errorf("address for chip ingress service is empty")
	}

	cfg := defaultConfig()

	for _, opt := range opts {
		opt(&cfg)
	}

	grpcOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(cfg.transportCredentials),
	}

	// Add headers as a unary interceptor
	if len(cfg.headers) > 0 {
		grpcOpts = append(grpcOpts, grpc.WithUnaryInterceptor(
			func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				for k, v := range cfg.headers {
					ctx = metadata.AppendToOutgoingContext(ctx, k, v)
				}
				return invoker(ctx, method, req, reply, cc, opts...)
			}))
	}

	conn, err := grpc.NewClient(
		address,
		grpcOpts...,
	)
	if err != nil {
		return nil, err
	}

	cfg.log.Info("Connected to ChipIngress service", zap.String("address", address))

	client := &chipIngressClient{
		Address: address,
		client:  pb.NewChipIngressClient(conn),
		conn:    conn,
		log:     cfg.log,
	}

	return client, nil
}

// Ping sends a request to the ChipIngress service to check if it is alive.
func (c *chipIngressClient) Ping(ctx context.Context) (string, error) {

	c.log.Debug("Pinging ChipIngress service")
	resp, err := c.client.Ping(ctx, &pb.EmptyRequest{})
	if err != nil {
		return "", err
	}

	c.log.Debug("Received ping response from ChipIngress service", zap.String("message", resp.Message))
	return resp.Message, nil
}

func (c *chipIngressClient) Publish(ctx context.Context, event ce.Event) (*pb.PublishResponse, error) {
	c.log.Debug("Calling ChipIngress service Publish")

	if err := validateEvents([]ce.Event{event}); err != nil {
		return nil, err
	}

	eventProto, err := ceformat.ToProto(&event)
	if err != nil {
		return nil, fmt.Errorf("failed to convert CloudEvent with ID %v to protobuf: %v", event.ID(), err)
	}

	resp, err := c.client.Publish(ctx, eventProto)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *chipIngressClient) PublishBatch(ctx context.Context, events []ce.Event) (*pb.PublishResponse, error) {
	c.log.Debug("Calling ChipIngress service PublishBatch")

	if err := validateEvents(events); err != nil {
		return nil, err
	}

	eventProtobufs := make([]*cepb.CloudEvent, len(events))
	for index, event := range events {
		eventProto, err := ceformat.ToProto(&event)
		if err != nil {
			return nil, fmt.Errorf("failed to convert CloudEvent with ID %v to protobuf: %v", event.ID(), err)
		}
		eventProtobufs[index] = eventProto
	}

	resp, err := c.client.PublishBatch(ctx, &pb.CloudEventBatch{Events: eventProtobufs})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Close closes the connection to the ChipIngress service.
func (c *chipIngressClient) Close() error {

	err := c.conn.Close()
	if err != nil {
		c.log.Error("Failed to close connection to ChipIngress service", zap.Error(err))
		return err
	}

	c.log.Info("Closed connection to ChipIngress service")
	return nil
}

// defaultConfig returns the default configuration for the ChipIngress service.
func defaultConfig() chipIngressClientConfig {
	return chipIngressClientConfig{
		log:                  zap.NewNop(),
		transportCredentials: insecure.NewCredentials(),
		headers:              make(map[string]string),
	}
}

// WithLogger sets the logger for the ChipIngress service.
func WithLogger(logger *zap.Logger) Opt {
	return func(c *chipIngressClientConfig) {
		c.log = logger
	}
}

// WithTransportCredentials sets the transport credentials for the ChipIngress service.
func WithTransportCredentials(credentials credentials.TransportCredentials) Opt {
	return func(c *chipIngressClientConfig) {
		c.transportCredentials = credentials
	}
}

// WithHeaders sets the headers for requests to the ChipIngress service.
func WithHeaders(headers map[string]string) Opt {
	return func(c *chipIngressClientConfig) {
		c.headers = headers
	}
}

func validateEvents(events []ce.Event) error {
	var errorMessages []string

	for index, event := range events {
		err := event.Validate()
		if err != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("Event ID %s (index %v): %v", event.ID(), index, err))
		}
	}

	if len(errorMessages) > 0 {
		return fmt.Errorf("validation failed for %d of %d events:\n%s", len(errorMessages), len(events), strings.Join(errorMessages, "\n"))
	}
	return nil
}

func NewEvent(domain, entity string, payload []byte) (ce.Event, error) {

	event := ce.NewEvent()
	event.SetSource(domain)
	event.SetType(entity)
	event.SetID(uuid.New().String())

	err := event.SetData(ceformat.ContentTypeProtobuf, payload)
	if err != nil {
		return ce.Event{}, fmt.Errorf("could not set data on event: %w", err)
	}

	return event, nil
}

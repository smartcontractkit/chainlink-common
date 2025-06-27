package chipingress

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	ceformat "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	cepb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	ce "github.com/cloudevents/sdk-go/v2"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
)

// HeaderProvider defines an interface for providing headers
type HeaderProvider interface {
	GetHeaders() map[string]string
}

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
	headerProvider       HeaderProvider
	authority            string
}

// newHeaderInterceptor creates a unary interceptor that adds headers from a HeaderProvider
func newHeaderInterceptor(provider HeaderProvider) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Add dynamic headers from provider if available
		if provider != nil {
			for k, v := range provider.GetHeaders() {
				ctx = metadata.AppendToOutgoingContext(ctx, k, v)
			}
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// NewChipIngressClient creates a new client for the Chip Ingress service with optional configuration.
func NewChipIngressClient(address string, opts ...Opt) (ChipIngressClient, error) {

	if address == "" {
		return nil, fmt.Errorf("address for chip ingress service is empty")
	}
	if !strings.Contains(address, ":") {
		return nil, fmt.Errorf("address is invalid, it must contain a port")
	}

	cfg := defaultConfig()

	// Apply configuration options
	for _, opt := range opts {
		opt(&cfg)
	}

	grpcOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(cfg.transportCredentials),
		grpc.WithAuthority(cfg.authority),
	}

	// Add headers as a unary interceptor
	if cfg.headerProvider != nil {
		grpcOpts = append(grpcOpts, grpc.WithUnaryInterceptor(newHeaderInterceptor(cfg.headerProvider)))
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
		headerProvider:       nil,
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

// WithHeaderProvider sets a dynamic header provider for requests
func WithHeaderProvider(provider HeaderProvider) Opt {
	return func(c *chipIngressClientConfig) {
		c.headerProvider = provider
	}
}

func WithInsecureConnection() Opt {
	return func(config *chipIngressClientConfig) {
		config.transportCredentials = insecure.NewCredentials()
	}
}

// Add a new option function for TLS with HTTP/2
func WithTLSAndHTTP2(serverName string) Opt {
	return func(config *chipIngressClientConfig) {
		tlsCfg := &tls.Config{
			ServerName: serverName,     // must match your server's host (SNI + cert SAN)
			NextProtos: []string{"h2"}, // force HTTP/2
		}
		config.transportCredentials = credentials.NewTLS(tlsCfg)
	}
}

func WithAuthority(authority string) Opt {
	return func(c *chipIngressClientConfig) {
		c.authority = authority
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

// NewEvent creates a new CloudEvent with the specified domain, entity, and payload.
// DE
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

func NewEventWithAttributes(domain, entity string, payload []byte, attributes map[string]any) (ce.Event, error) {

	event := ce.NewEvent()
	event.SetSource(domain)
	event.SetType(entity)
	event.SetID(uuid.New().String())

	// Set optional attributes if provided
	if attributes == nil {
		attributes = make(map[string]any)
	}

	if val, ok := attributes["recordedtime"].(time.Time); ok {
		event.SetExtension("recordedtime", ce.Timestamp{Time: val.UTC()})
	} else {
		event.SetExtension("recordedtime", ce.Timestamp{Time: time.Now().UTC()})
	}

	if val, ok := attributes["time"].(time.Time); ok {
		event.SetTime(val.UTC())
	}
	if val, ok := attributes["datacontenttype"].(string); ok {
		event.SetDataContentType(val)
	}
	if val, ok := attributes["dataschema"].(string); ok {
		event.SetDataSchema(val)
	}
	if val, ok := attributes["subject"].(string); ok {
		event.SetSubject(val)
	}

	err := event.SetData(ceformat.ContentTypeProtobuf, payload)
	if err != nil {
		return ce.Event{}, fmt.Errorf("could not set data on event: %w", err)
	}

	return event, nil
}


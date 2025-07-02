package chipingress

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	ceformat "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	ce "github.com/cloudevents/sdk-go/v2"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
)

// HeaderProvider defines an interface for providing headers
type HeaderProvider interface {
	GetHeaders() map[string]string
}

type Client interface {
	pb.ChipIngressClient
	Close() error
}

type client struct {
	client pb.ChipIngressClient
	conn   *grpc.ClientConn
}

// Opt defines a function type for configuring the ChipIngressClient.
type Opt func(*chipIngressClientConfig)

// chipIngressClientConfig is the configuration for the ChipIngressClient.
type chipIngressClientConfig struct {
	transportCredentials credentials.TransportCredentials
	perRPCCredentials    credentials.PerRPCCredentials
	headerProvider       HeaderProvider
	insecureConnection   bool
	host                 string
}

func newChipIngressConfig(host string) *chipIngressClientConfig {
	cfg := &chipIngressClientConfig{
		headerProvider:    nil,
		perRPCCredentials: nil,
		host:              host,
	}
	WithInsecureConnection()(cfg) // Default to insecure connection
	return cfg
}

// NewChipIngressClient creates a new client for the Chip Ingress service with optional configuration.
func NewChipIngressClient(address string, opts ...Opt) (Client, error) {
	// Validate address
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address format: %v", err)
	}
	cfg := newChipIngressConfig(host)

	// Apply configuration options
	for _, opt := range opts {
		opt(cfg)
	}
	grpcOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(cfg.transportCredentials),
	}
	// Auth
	if cfg.perRPCCredentials != nil {
		grpcOpts = append(grpcOpts, grpc.WithPerRPCCredentials(cfg.perRPCCredentials))
	}
	// Add headers as a unary interceptor, use for non-auth headers
	if cfg.headerProvider != nil {
		grpcOpts = append(grpcOpts, grpc.WithUnaryInterceptor(newHeaderInterceptor(cfg.headerProvider)))
		// NOTE: not supporting streaming interceptors
	}

	conn, err := grpc.NewClient(address, grpcOpts...)
	if err != nil {
		return nil, err
	}
	return &client{pb.NewChipIngressClient(conn), conn}, nil
}

func (c *client) Ping(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	return c.client.Ping(ctx, in, opts...)
}

func (c *client) Publish(ctx context.Context, in *CloudEventPb, opts ...grpc.CallOption) (*PublishResponse, error) {
	return c.client.Publish(ctx, in, opts...)
}

func (c *client) PublishBatch(ctx context.Context, in *CloudEventBatch, opts ...grpc.CallOption) (*PublishResponse, error) {
	return c.client.PublishBatch(ctx, in, opts...)
}

// StreamEvents - Experimental, this API is subject to change.
func (c *client) StreamEvents(_ context.Context, _ ...grpc.CallOption) (grpc.BidiStreamingClient[StreamEventsRequest, StreamEventsResponse], error) {
	return nil, fmt.Errorf("not implemented: StreamEvents is experimental and not supported yet")
}

func (c *client) Close() error {
	return c.conn.Close()
}

// WithBasicAuth sets the basic-auth credentials for the ChipIngress service.
// Default is to require TLS for security.
func WithBasicAuth(user, pass string) Opt {
	return func(c *chipIngressClientConfig) {
		requireTLS := !c.insecureConnection
		c.perRPCCredentials = newBasicAuthCredentials(user, pass, requireTLS)
	}
}

// WithTokenAuth sets the token-based credentials for the ChipIngress service.
// Use for CSA-Key based authentication.
func WithTokenAuth(tokenProvider HeaderProvider) Opt {
	return func(c *chipIngressClientConfig) {
		requireTLS := !c.insecureConnection
		c.perRPCCredentials = newTokenAuthCredentials(tokenProvider, requireTLS)
	}
}

// WithTransportCredentials sets the transport custom credentials for the ChipIngress service.
func WithTransportCredentials(creds credentials.TransportCredentials) Opt {
	return func(c *chipIngressClientConfig) { c.transportCredentials = creds }
}

// WithHeaderProvider sets a dynamic header provider for requests
// NOTE: for CSA-Key based authentication, use WithTokenAuth instead.
func WithHeaderProvider(provider HeaderProvider) Opt {
	return func(c *chipIngressClientConfig) { c.headerProvider = provider }
}

// WithInsecureConnection configures the client to use an insecure connection (no TLS).
func WithInsecureConnection() Opt {
	return func(config *chipIngressClientConfig) {
		config.insecureConnection = true
		config.transportCredentials = insecure.NewCredentials() // Use insecure credentials
	}
}

// Add a new option function for TLS with HTTP/2
func WithTLS() Opt {
	return func(config *chipIngressClientConfig) {
		config.insecureConnection = false
		tlsCfg := &tls.Config{
			ServerName: config.host,    // must match your server's host (SNI + cert SAN)
			NextProtos: []string{"h2"}, // force HTTP/2
		}
		config.transportCredentials = credentials.NewTLS(tlsCfg) // Use TLS
	}
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

// NewEvent creates a new CloudEvent with the specified domain, entity, payload, and optional attributes.
func NewEvent(domain, entity string, payload []byte, attributes map[string]any) (CloudEvent, error) {

	event := ce.NewEvent()
	event.SetSource(domain)
	event.SetType(entity)
	event.SetID(uuid.New().String())

	// Set optional attributes if provided
	if attributes == nil {
		attributes = make(map[string]any)
	}

	if val, ok := attributes["recordedtime"].(time.Time); ok {
		event.SetExtension("recordedtime", val.UTC())
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

func EventToProto(event CloudEvent) (*CloudEventPb, error) {
	eventPb, err := ceformat.ToProto(&event)
	if err != nil {
		return nil, fmt.Errorf("could not convert event to proto: %w", err)
	}
	return eventPb, nil
}

func ProtoToEvent(eventPb *CloudEventPb) (CloudEvent, error) {
	if eventPb == nil {
		return CloudEvent{}, fmt.Errorf("could not convert proto to event: eventPb is nil")
	}
	event, err := ceformat.FromProto(eventPb)
	if err != nil {
		return CloudEvent{}, fmt.Errorf("could not convert proto to event: %w", err)
	}
	return *event, nil
}

func EventsToBatch(events []CloudEvent) (*CloudEventBatch, error) {
	batch := &CloudEventBatch{
		Events: make([]*CloudEventPb, 0, len(events)),
	}
	for _, event := range events {
		eventPb, err := EventToProto(event)
		if err != nil {
			return nil, fmt.Errorf("could not convert event to proto: %w", err)
		}
		batch.Events = append(batch.Events, eventPb)
	}
	return batch, nil
}

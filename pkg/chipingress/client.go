package chipingress

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	ceformat "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	cepb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	ce "github.com/cloudevents/sdk-go/v2"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
)

const maxMessageSize = 16 * 1024 * 1024 // 16MB

// HeaderProvider defines an interface for providing headers

type Client interface {
	pb.ChipIngressClient
	Close() error
	RegisterSchemas(ctx context.Context, schemas ...*pb.Schema) (map[string]int, error)
}

type client struct {
	client pb.ChipIngressClient
	conn   *grpc.ClientConn
}

// Opt defines a function type for configuring the ChipIngressClient.
type Opt func(*clientConfig)

// clientConfig is the configuration for the ChipIngressClient.
type clientConfig struct {
	transportCredentials  credentials.TransportCredentials
	perRPCCredentials     credentials.PerRPCCredentials
	headerProvider        HeaderProvider
	insecureConnection    bool
	host                  string
	meterProvider         metric.MeterProvider
	tracerProvider        trace.TracerProvider
	nopInfoHeaderProvider HeaderProvider
}

func newClientConfig(host string) *clientConfig {
	cfg := &clientConfig{
		headerProvider:    nil,
		perRPCCredentials: nil,
		host:              host,
		// Default to insecure connection
		insecureConnection:    true,
		transportCredentials:  insecure.NewCredentials(),
		nopInfoHeaderProvider: nil,
	}
	return cfg
}

// NewClient creates a new client for the Chip Ingress service with optional configuration.
func NewClient(address string, opts ...Opt) (Client, error) {
	// Validate address
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address format: %v", err)
	}
	cfg := newClientConfig(host)

	// Apply configuration options
	for _, opt := range opts {
		opt(cfg)
	}
	// Build otelgrpc handler options
	var otelOpts []otelgrpc.Option
	if cfg.meterProvider != nil {
		otelOpts = append(otelOpts, otelgrpc.WithMeterProvider(cfg.meterProvider))
	}
	if cfg.tracerProvider != nil {
		otelOpts = append(otelOpts, otelgrpc.WithTracerProvider(cfg.tracerProvider))
	}

	grpcOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(cfg.transportCredentials),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(otelOpts...)),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             1 * time.Second,
			PermitWithoutStream: true,
		}),
	}
	// Retry policy
	retryPolicy := `{
		"maxAttempts": 3,
		"initialBackoff": "100ms",
		"maxBackoff": "1s",
		"backoffMultiplier": 2,
		"retryableStatusCodes": ["UNAVAILABLE", "RESOURCE_EXHAUSTED"]
	}`
	grpcOpts = append(grpcOpts, grpc.WithDefaultServiceConfig(retryPolicy))
	// Auth
	if cfg.perRPCCredentials != nil {
		grpcOpts = append(grpcOpts, grpc.WithPerRPCCredentials(cfg.perRPCCredentials))
	}
	// Add headers as a unary interceptor, use for non-auth headers
	if cfg.headerProvider != nil {
		grpcOpts = append(grpcOpts, grpc.WithUnaryInterceptor(newHeaderInterceptor(cfg.headerProvider)))
		// NOTE: not supporting streaming interceptors
	}

	if cfg.nopInfoHeaderProvider != nil {
		grpcOpts = append(grpcOpts, grpc.WithUnaryInterceptor(newHeaderInterceptor(cfg.nopInfoHeaderProvider)))
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

func (c *client) RegisterSchema(ctx context.Context, in *pb.RegisterSchemaRequest, opts ...grpc.CallOption) (*pb.RegisterSchemaResponse, error) {
	return c.client.RegisterSchema(ctx, in, opts...)
}

func (c *client) Close() error {
	return c.conn.Close()
}

// RegisterSchemas registers one or more schemas with the Chip Ingress service.
func (c *client) RegisterSchemas(ctx context.Context, schemas ...*pb.Schema) (map[string]int, error) {
	request := &pb.RegisterSchemaRequest{Schemas: schemas}

	resp, err := c.client.RegisterSchema(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to register schema: %w", err)
	}

	registeredMap := make(map[string]int)
	for _, schema := range resp.Registered {
		registeredMap[schema.Subject] = int(schema.Version)
	}

	return registeredMap, nil
}

// WithBasicAuth sets the basic-auth credentials for the ChipIngress service.
// Default is to require TLS for security.
func WithBasicAuth(user, pass string) Opt {
	return func(c *clientConfig) {
		requireTLS := !c.insecureConnection
		c.perRPCCredentials = newBasicAuthCredentials(user, pass, requireTLS)
	}
}

// WithTokenAuth sets the token-based credentials for the ChipIngress service.
// Use for CSA-Key based authentication.
func WithTokenAuth(tokenProvider HeaderProvider) Opt {
	return func(c *clientConfig) {
		requireTLS := !c.insecureConnection
		c.perRPCCredentials = newTokenAuthCredentials(tokenProvider, requireTLS)
	}
}

// WithTransportCredentials sets the transport custom credentials for the ChipIngress service.
func WithTransportCredentials(creds credentials.TransportCredentials) Opt {
	return func(c *clientConfig) { c.transportCredentials = creds }
}

// WithHeaderProvider sets a dynamic header provider for requests
// NOTE: for CSA-Key based authentication, use WithTokenAuth instead.
func WithHeaderProvider(provider HeaderProvider) Opt {
	return func(c *clientConfig) { c.headerProvider = provider }
}

// WithInsecureConnection configures the client to use an insecure connection (no TLS).
func WithInsecureConnection() Opt {
	return func(config *clientConfig) {
		config.insecureConnection = true
		config.transportCredentials = insecure.NewCredentials() // Use insecure credentials
	}
}

// Add a new option function for TLS with HTTP/2
func WithTLS() Opt {
	return func(config *clientConfig) {
		config.insecureConnection = false
		tlsCfg := &tls.Config{
			ServerName: config.host,    // must match your server's host (SNI + cert SAN)
			NextProtos: []string{"h2"}, // force HTTP/2
		}
		config.transportCredentials = credentials.NewTLS(tlsCfg) // Use TLS
	}
}

// WithMeterProvider sets a custom OpenTelemetry MeterProvider for metrics collection.
// If not set, the global meter provider will be used.
func WithMeterProvider(provider metric.MeterProvider) Opt {
	return func(c *clientConfig) { c.meterProvider = provider }
}

// WithTracerProvider sets a custom OpenTelemetry TracerProvider for distributed tracing.
// If not set, the global tracer provider will be used.
func WithTracerProvider(provider trace.TracerProvider) Opt {
	return func(c *clientConfig) { c.tracerProvider = provider }
}

func WithNOPLookup() Opt {
	return func(c *clientConfig) {
		c.nopInfoHeaderProvider = headerProviderFunc(func(ctx context.Context) (map[string]string, error) {
			return map[string]string{
				"x-include-nop-info": "true",
			}, nil
		})
	}
}

// newHeaderInterceptor creates a unary interceptor that adds headers from a HeaderProvider
func newHeaderInterceptor(provider HeaderProvider) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Add dynamic headers from provider if available
		if provider != nil {

			headers, err := provider.Headers(ctx)
			if err != nil {
				return fmt.Errorf("failed to get headers: %w", err)
			}

			for k, v := range headers {
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

	recordedTime := time.Now()
	if val, ok := attributes["recordedtime"].(time.Time); ok && !val.IsZero() {
		recordedTime = val
	}
	recordedTime = recordedTime.UTC().Truncate(time.Millisecond)
	event.SetExtension("recordedtime", ce.Timestamp{Time: recordedTime})

	if val, ok := attributes["time"].(time.Time); ok && !val.IsZero() {
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

var _ Client = (*NoopClient)(nil)

// NoopClient is a no-op implementation of the Client interface.
// All methods return successfully without performing any actual operations.
type NoopClient struct{}

// Close is a no-op
func (NoopClient) Close() error {
	return nil
}

// Ping is a no-op
func (NoopClient) Ping(ctx context.Context, in *pb.EmptyRequest, opts ...grpc.CallOption) (*pb.PingResponse, error) {
	return &pb.PingResponse{Message: "pong"}, nil
}

// Publish is a no-op
func (NoopClient) Publish(ctx context.Context, in *cepb.CloudEvent, opts ...grpc.CallOption) (*pb.PublishResponse, error) {
	return &pb.PublishResponse{}, nil
}

// PublishBatch is a no-op
func (NoopClient) PublishBatch(ctx context.Context, in *pb.CloudEventBatch, opts ...grpc.CallOption) (*pb.PublishResponse, error) {
	return &pb.PublishResponse{}, nil
}

// StreamEvents is a no-op
func (NoopClient) StreamEvents(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[pb.StreamEventsRequest, pb.StreamEventsResponse], error) {
	return nil, nil
}

// RegisterSchema is a no-op
func (NoopClient) RegisterSchema(ctx context.Context, in *pb.RegisterSchemaRequest, opts ...grpc.CallOption) (*pb.RegisterSchemaResponse, error) {
	return &pb.RegisterSchemaResponse{}, nil
}

// RegisterSchemas is a no-op
func (NoopClient) RegisterSchemas(ctx context.Context, schemas ...*pb.Schema) (map[string]int, error) {
	return make(map[string]int), nil
}

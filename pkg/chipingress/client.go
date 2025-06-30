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
	cepb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	ce "github.com/cloudevents/sdk-go/v2"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
)

// HeaderProvider defines an interface for providing headers
type HeaderProvider interface {
	GetHeaders() map[string]string
}

// Opt defines a function type for configuring the ChipIngressClient.
type Opt func(*chipIngressClientConfig)

// chipIngressClientConfig is the configuration for the ChipIngressClient.
type chipIngressClientConfig struct {
	transportCredentials credentials.TransportCredentials
	headerProvider       HeaderProvider
	authority            string
}

// NewChipIngressClient creates a new client for the Chip Ingress service with optional configuration.
func NewChipIngressClient(address string, opts ...Opt) (pb.ChipIngressClient, error) {
	// Validate address
	_, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address format: %v", err)
	}
	// Defaults
	cfg := chipIngressClientConfig{
		transportCredentials: insecure.NewCredentials(),
		headerProvider:       nil,
	}
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

	conn, err := grpc.NewClient(address, grpcOpts...)
	if err != nil {
		return nil, err
	}
	return pb.NewChipIngressClient(conn), nil
}

// WithTransportCredentials sets the transport credentials for the ChipIngress service.
func WithTransportCredentials(creds credentials.TransportCredentials) Opt {
	return func(c *chipIngressClientConfig) { c.transportCredentials = creds }
}

// WithHeaderProvider sets a dynamic header provider for requests
func WithHeaderProvider(provider HeaderProvider) Opt {
	return func(c *chipIngressClientConfig) { c.headerProvider = provider }
}

// WithInsecureConnection configures the client to use an insecure connection (no TLS).
func WithInsecureConnection() Opt {
	return func(config *chipIngressClientConfig) { config.transportCredentials = insecure.NewCredentials() }
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

// WithAuthority sets the authority for the gRPC connection.
func WithAuthority(authority string) Opt {
	return func(c *chipIngressClientConfig) { c.authority = authority }
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

// NewEvent creates a new CloudEvent with the specified domain, entity, and payload.
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

// NewEventWithAttributes creates a new CloudEvent with the specified domain, entity, payload, and optional attributes.
func NewEventWithAttributes(domain, entity string, payload []byte, attributes map[string]any) (*cepb.CloudEvent, error) {

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
		return nil, fmt.Errorf("could not set data on event: %w", err)
	}

	return ceformat.ToProto(&event)
}

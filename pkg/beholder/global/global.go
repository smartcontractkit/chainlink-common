package global

import (
	"context"
	"sync/atomic"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	otellog "go.opentelemetry.io/otel/log"
	otelmetric "go.opentelemetry.io/otel/metric"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/logger"
)

var log = logger.New()

// Pointer to the global BeholderClient
var globalBeholderClient = defaultBeholderClient()

// SetClient sets the global BeholderClient
func SetClient(client *beholder.BeholderClient) {
	globalBeholderClient.Store(client)
}

// Returns the global BeholderClient
// This allows to access the BeholderClient from anywhere in the code
// Its thread-safe and can be used in concurrent environment
func GetClient() *beholder.BeholderClient {
	return globalBeholderClient.Load()
}

func Logger() otellog.Logger {
	return GetClient().Logger()
}

func Tracer() oteltrace.Tracer {
	return GetClient().Tracer()
}

func Meter() otelmetric.Meter {
	return GetClient().Meter()
}

func Emitter() beholder.Emitter {
	return GetClient().Emitter()
}

func SpanFromContext(ctx context.Context) oteltrace.Span {
	return oteltrace.SpanFromContext(ctx)
}

func defaultBeholderClient() *atomic.Pointer[beholder.BeholderClient] {
	ptr := &atomic.Pointer[beholder.BeholderClient]{}
	ptr.Store(beholder.NewNoopClient())
	return ptr
}

// TODO: rename to EmitMessage
func EmitMessage(ctx context.Context, message beholder.Message) error {
	return Emitter().EmitMessage(ctx, message)
}

// TODO: rename to EmitMessage
func Emit(ctx context.Context, body []byte, attrs beholder.Attributes) error {
	return Emitter().Emit(ctx, body, attrs)
}

func Bootstrap(cfg beholder.Config) error {
	// Initialize beholder client
	client, err := beholder.NewOtelClient(cfg, func(err error) {
		log.Infof("OTel error %s", err)
	})
	if err != nil {
		return err
	}
	// Set global client so it will be accessible from anywhere through beholder/global functions
	SetClient(client)
	return nil
}

func NewConfig() beholder.Config {
	return beholder.DefaultBeholderConfig()
}

// Creates logger based on zap logger which writes to stdout
func NewSimpleLogger() logger.Logger {
	return logger.New()
}

// Returns a new logger based on otelzap logger
// The logger is able to write to stdout and send logs to otel collector
func NewLogger() *otelzap.Logger {
	return logger.NewOtelzapLogger()
}

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

// Pointer to the global Beholder Client
var globalBeholderClient = defaultBeholderClient()

// SetClient sets the global Beholder Client
func SetClient(client *beholder.Client) {
	globalBeholderClient.Store(client)
}

// Returns the global Beholder Client
// Its thread-safe and can be used concurrently
func GetClient() beholder.Client {
	ptr := globalBeholderClient.Load()
	return *ptr
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

func defaultBeholderClient() *atomic.Pointer[beholder.Client] {
	ptr := &atomic.Pointer[beholder.Client]{}
	client := beholder.NewNoopClient()
	ptr.Store(&client)
	return ptr
}

func EmitMessage(ctx context.Context, message beholder.Message) error {
	return Emitter().EmitMessage(ctx, message)
}

func Emit(ctx context.Context, body []byte, attrs beholder.Attributes) error {
	return Emitter().Emit(ctx, body, attrs)
}

func Bootstrap(cfg beholder.Config) error {
	// Initialize beholder client
	c, err := beholder.NewOtelClient(cfg, func(err error) {
		log.Infof("OTel error %s", err)
	})
	if err != nil {
		return err
	}
	var client beholder.Client = c
	// Set global client so it will be accessible from anywhere through beholder/global functions
	SetClient(&client)
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

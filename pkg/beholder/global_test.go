package beholder_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/otel"
	otelattribute "go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	otellogglobal "go.opentelemetry.io/otel/log/global"
	otellognoop "go.opentelemetry.io/otel/log/noop"
	otelmetric "go.opentelemetry.io/otel/metric"
	otelmetricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
	oteltracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/internal/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestGlobal(t *testing.T) {
	// Get global logger, tracer, meter, messageEmitter
	// If not initialized with beholder.SetClient will return noop client
	logger, tracer, meter, messageEmitter := beholder.Logger(), beholder.Tracer(), beholder.Meter(), beholder.MessageEmitter()
	noopClient := beholder.NewNoopClient()
	assert.IsType(t, otellognoop.Logger{}, logger)
	assert.IsType(t, oteltracenoop.Tracer{}, tracer)
	assert.IsType(t, otelmetricnoop.Meter{}, meter)
	expectedMessageEmitter := beholder.NewNoopClient().Emitter
	assert.IsType(t, expectedMessageEmitter, messageEmitter)

	var noopClientPtr *beholder.Client = noopClient
	assert.IsType(t, noopClientPtr, beholder.GetClient())
	assert.NotSame(t, noopClientPtr, beholder.GetClient())

	// Set beholder client so it will be accessible from anywhere through beholder functions
	beholder.SetClient(noopClientPtr)
	assert.Same(t, noopClientPtr, beholder.GetClient())

	// After that use beholder functions to get logger, tracer, meter, messageEmitter
	logger, tracer, meter, messageEmitter = beholder.Logger(), beholder.Tracer(), beholder.Meter(), beholder.MessageEmitter()

	// Emit otel log record
	logger.Emit(tests.Context(t), otellog.Record{})

	// Create trace span
	ctx, span := tracer.Start(tests.Context(t), "ExampleGlobalClient", oteltrace.WithAttributes(otelattribute.String("key", "value")))
	defer span.End()

	// Create metric counter
	counter, _ := meter.Int64Counter("global_counter")
	counter.Add(tests.Context(t), 1)

	// Emit custom message
	err := messageEmitter.Emit(ctx, []byte("test"), beholder.Attributes{"key": "value"})
	if err != nil {
		t.Fatalf("Error emitting message: %v", err)
	}
}

func TestClient_SetGlobalOtelProviders(t *testing.T) {
	exporterMock := mocks.NewOTLPExporter(t)
	defer exporterMock.AssertExpectations(t)

	// Restore global providers after test
	defer restoreProviders(t, providers{
		otellogglobal.GetLoggerProvider(),
		otel.GetTracerProvider(),
		otel.GetTextMapPropagator(),
		otel.GetMeterProvider(),
	})

	var b strings.Builder
	client, err := beholder.NewStdoutClient(beholder.WithWriter(&b))
	assert.NoError(t, err)
	// Set global Otel Client
	beholder.SetClient(client)

	// Set global otel tracer, meter, logger providers from global beholder otel client
	beholder.SetGlobalOtelProviders()

	assert.Equal(t, client.LoggerProvider, otellogglobal.GetLoggerProvider())
	assert.Equal(t, client.TracerProvider, otel.GetTracerProvider())
	assert.Equal(t, client.MeterProvider, otel.GetMeterProvider())
}

type providers struct {
	loggerProvider    otellog.LoggerProvider
	tracerProvider    oteltrace.TracerProvider
	textMapPropagator propagation.TextMapPropagator
	meterProvider     otelmetric.MeterProvider
}

func restoreProviders(t *testing.T, p providers) {
	otellogglobal.SetLoggerProvider(p.loggerProvider)
	otel.SetTracerProvider(p.tracerProvider)
	otel.SetTextMapPropagator(p.textMapPropagator)
	otel.SetMeterProvider(p.meterProvider)

	assert.Equal(t, p.loggerProvider, otellogglobal.GetLoggerProvider())
	assert.Equal(t, p.tracerProvider, otel.GetTracerProvider())
	assert.Equal(t, p.textMapPropagator, otel.GetTextMapPropagator())
	assert.Equal(t, p.meterProvider, otel.GetMeterProvider())
}

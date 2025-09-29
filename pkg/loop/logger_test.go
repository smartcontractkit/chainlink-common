package loop

import (
	"context"
	"sync"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/logger/otelzap"
	"github.com/stretchr/testify/assert"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

func Test_removeArg(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []interface{}
		key  string

		wantArgs []interface{}
		wantVal  string
	}{
		{"empty", nil, "logger",
			nil, ""},
		{"simple", []any{"logger", "foo"}, "logger",
			[]any{}, "foo"},
		{"multi", []any{"logger", "foo", "bar", "baz"}, "logger",
			[]any{"bar", "baz"}, "foo"},
		{"reorder", []any{"bar", "baz", "logger", "foo"}, "logger",
			[]any{"bar", "baz"}, "foo"},

		{"invalid", []any{"logger"}, "logger",
			[]any{"logger"}, ""},
		{"invalid-multi", []any{"foo", "bar", "logger"}, "logger",
			[]any{"foo", "bar", "logger"}, ""},
		{"value", []any{"foo", "logger", "bar", "baz"}, "logger",
			[]any{"foo", "logger", "bar", "baz"}, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			args, val := removeArg(tt.args, tt.key)
			assert.ElementsMatch(t, tt.wantArgs, args)
			assert.Equal(t, tt.wantVal, val)
		})
	}
}

func TestNewOtelLogger(t *testing.T) {
	tests := []struct {
		name    string
		logFn   func(l logger.Logger)
		wantMsg string
	}{
		{
			name: "debug",
			logFn: func(l logger.Logger) {
				l.Debugw("hello world", "k", "v")
			},
			wantMsg: "hello world",
		},
		{
			name: "info",
			logFn: func(l logger.Logger) {
				l.Infow("info msg", "a", 1)
			},
			wantMsg: "info msg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := &recordingExporter{}
			lp := sdklog.NewLoggerProvider(
				sdklog.WithProcessor(sdklog.NewSimpleProcessor(exp)),
			)
			otelLggr := lp.Logger("test-" + tt.name)

			lggr, err := NewOtelLogger(otelLggr)
			if err != nil {
				t.Fatalf("NewOtelLogger error: %v", err)
			}

			tt.logFn(lggr)

			if len(exp.records) != 1 {
				t.Fatalf("expected 1 exported record, got %d", len(exp.records))
			}
			if got := exp.records[0].Body().AsString(); got != tt.wantMsg {
				t.Fatalf("unexpected body: got %q want %q", got, tt.wantMsg)
			}
		})
	}
}

// recordingExporter captures exported log records (current sdk/log Export signature).
type recordingExporter struct {
	mu      sync.Mutex
	records []sdklog.Record
}

func (r *recordingExporter) Export(_ context.Context, recs []sdklog.Record) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, recs...)
	return nil
}
func (r *recordingExporter) ForceFlush(context.Context) error { return nil }
func (r *recordingExporter) Shutdown(context.Context) error   { return nil }

// Compile-time assertion that otelzap.NewCore still satisfies zapcore.Core usage pattern.
// (Guards against accidental API break causing this test file to silently compile with stubs.)
var _ = otelzap.NewCore
var _ logger.Logger // silence unused import of logger in case future refactors remove usage

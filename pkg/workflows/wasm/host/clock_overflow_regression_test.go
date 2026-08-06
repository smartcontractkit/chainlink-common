package host

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/host/mocks"
	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func TestRegressionClockTimeGetResultPointerOverflowDoesNotPanic(t *testing.T) {
	t.Parallel()
	binary := clockTimeGetOverflowWasm()

	m, err := NewModule(t.Context(), &ModuleConfig{
		Logger:         logger.Test(t),
		IsUncompressed: true,
	}, binary)
	require.NoError(t, err)
	require.False(t, m.IsLegacyDAG())
	m.Start()
	defer m.Close()

	helper := mocks.NewMockExecutionHelper(t)
	helper.EXPECT().GetWorkflowExecutionID().Return("id")
	helper.EXPECT().GetNodeTime().Return(time.Unix(1, 0)).Maybe()

	require.NotPanics(t, func() {
		_, _ = m.Execute(t.Context(), &sdkpb.ExecuteRequest{
			Request: &sdkpb.ExecuteRequest_Trigger{},
		}, helper)
	})
}

func clockTimeGetOverflowWasm() []byte {
	var out []byte
	out = append(out, 0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00)

	appendSection := func(id byte, payload []byte) {
		out = append(out, id)
		out = appendULEB(out, uint32(len(payload)))
		out = append(out, payload...)
	}

	var types []byte
	types = appendULEB(types, 2)
	types = append(types, 0x60, 0x00, 0x00)
	types = append(types, 0x60, 0x03, 0x7f, 0x7e, 0x7f, 0x01, 0x7f)
	appendSection(1, types)

	var imports []byte
	imports = appendULEB(imports, 2)
	imports = appendName(imports, "env")
	imports = appendName(imports, "version_v2")
	imports = append(imports, 0x00)
	imports = appendULEB(imports, 0)
	imports = appendName(imports, "wasi_snapshot_preview1")
	imports = appendName(imports, "clock_time_get")
	imports = append(imports, 0x00)
	imports = appendULEB(imports, 1)
	appendSection(2, imports)

	var functions []byte
	functions = appendULEB(functions, 1)
	functions = appendULEB(functions, 0)
	appendSection(3, functions)

	var memory []byte
	memory = appendULEB(memory, 1)
	memory = append(memory, 0x00)
	memory = appendULEB(memory, 1)
	appendSection(5, memory)

	var exports []byte
	exports = appendULEB(exports, 2)
	exports = appendName(exports, "memory")
	exports = append(exports, 0x02)
	exports = appendULEB(exports, 0)
	exports = appendName(exports, "_start")
	exports = append(exports, 0x00)
	exports = appendULEB(exports, 2)
	appendSection(7, exports)

	var body []byte
	body = appendULEB(body, 0)
	body = append(body, 0x41, 0x00)
	body = append(body, 0x42, 0x00)
	body = append(body, 0x41)
	body = appendSLEB32(body, 0x7ffffffc)
	body = append(body, 0x10)
	body = appendULEB(body, 1)
	body = append(body, 0x1a, 0x0b)

	var code []byte
	code = appendULEB(code, 1)
	code = appendULEB(code, uint32(len(body)))
	code = append(code, body...)
	appendSection(10, code)

	return out
}

func appendName(dst []byte, name string) []byte {
	dst = appendULEB(dst, uint32(len(name)))
	return append(dst, name...)
}

func appendULEB(dst []byte, value uint32) []byte {
	for {
		b := byte(value & 0x7f)
		value >>= 7
		if value != 0 {
			b |= 0x80
		}
		dst = append(dst, b)
		if value == 0 {
			return dst
		}
	}
}

func appendSLEB32(dst []byte, value int32) []byte {
	for {
		b := byte(value & 0x7f)
		value >>= 7
		done := (value == 0 && b&0x40 == 0) || (value == -1 && b&0x40 != 0)
		if !done {
			b |= 0x80
		}
		dst = append(dst, b)
		if done {
			return dst
		}
	}
}

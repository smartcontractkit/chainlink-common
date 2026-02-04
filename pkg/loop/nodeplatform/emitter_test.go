package nodeplatform

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/beholdertest"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	commonv1 "github.com/smartcontractkit/chainlink-protos/node-platform/common/v1"
)

func TestNormalizeEndpoint(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "https with userinfo path and port",
			in:   "https://user:pass@host:8545/path?x=y",
			want: "https://host",
		},
		{
			name: "wss with path",
			in:   "wss://host/ws",
			want: "wss://host",
		},
		{
			name: "host with port no scheme",
			in:   "host:8545",
			want: "host",
		},
		{
			name: "userinfo host path no scheme",
			in:   "user:pass@host:8545/path",
			want: "host",
		},
		{
			name: "invalid",
			in:   "://",
			want: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, NormalizeEndpoint(tt.in))
		})
	}
}

func TestNormalizeEndpointsDedup(t *testing.T) {
	got := NormalizeEndpoints([]string{
		"",
		"   ",
		"https://user:pass@host:8545/path",
		"host:8545",
		"https://user:pass@host:8545/path", // duplicate
		"host:8545",                        // duplicate
	})
	require.Equal(t, []string{"https://host", "host"}, got)
}

func TestParseOriginURL(t *testing.T) {
	scheme, host, err := parseOriginURL("https://user:pass@host:8545/path")
	require.NoError(t, err)
	require.Equal(t, "https", scheme)
	require.Equal(t, "host", host)

	scheme, host, err = parseOriginURL("user:pass@host:8545/path")
	require.NoError(t, err)
	require.Equal(t, "", scheme)
	require.Equal(t, "host", host)

	scheme, host, err = parseOriginURL("/just/path")
	require.NoError(t, err)
	require.Equal(t, "", scheme)
	require.Equal(t, "/just/path", host)

	_, _, err = parseOriginURL("http://\x7f")
	require.Error(t, err)
}

func TestNewChainPluginConfigEmitterWithIntervalDefaults(t *testing.T) {
	prev := beholder.GetClient()
	client := beholder.NewNoopClient()
	client.Config.AuthPublicKeyHex = "from-beholder"
	beholder.SetClient(client)
	t.Cleanup(func() { beholder.SetClient(prev) })

	emitter := NewChainPluginConfigEmitterWithInterval(
		logger.Test(t),
		"",
		"",
		[]string{"host:8545"},
		0,
	)

	require.Equal(t, DefaultEmitInterval, emitter.interval)
	require.Equal(t, "from-beholder", emitter.csaPublicKey)
	require.Equal(t, "", emitter.chainID)
	require.Equal(t, []string{"host"}, emitter.urls)
}

func TestEmitterEmit(t *testing.T) {
	obs := beholdertest.NewObserver(t)
	lggr := logger.Test(t)

	emitter := NewChainPluginConfigEmitterWithInterval(
		lggr,
		"csa-123",
		"chain-1",
		[]string{
			"https://user:pass@host:8545/path",
			"host:8545",
			"https://user:pass@host:8545/path", // duplicate
		},
		DefaultEmitInterval,
	)

	emitter.emit(context.Background())

	msgs := obs.Messages(t, beholder.AttrKeyDomain, BeholderDomain)
	require.Len(t, msgs, 1)

	msg := msgs[0]
	require.Equal(t, BeholderDomain, msg.Attrs[beholder.AttrKeyDomain])
	require.Equal(t, BeholderEntity, msg.Attrs[beholder.AttrKeyEntity])
	require.Equal(t, BeholderDataSchema, msg.Attrs[beholder.AttrKeyDataSchema])

	var got commonv1.ChainPluginConfig
	require.NoError(t, proto.Unmarshal(msg.Body, &got))
	require.Equal(t, "csa-123", got.CsaPublicKey)
	require.Equal(t, "chain-1", got.ChainId)
	require.Equal(t, []string{"https://host", "host"}, got.Urls)
}

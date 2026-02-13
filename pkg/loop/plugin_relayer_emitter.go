package loop

import (
	"context"
	"net/url"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	commonv1 "github.com/smartcontractkit/chainlink-protos/node-platform/common/v1"
)

const (
	beholderDomain     = "node-platform"
	beholderEntity     = "common.v1.ChainPluginConfig"
	beholderDataSchema = "/node-platform/common/v1"

	defaultEmitInterval = time.Minute * 3
	serviceName         = "PluginRelayerConfigEmitter"
)

type pluginRelayerConfigEmitter struct {
	services.Service
	eng *services.Engine

	csaPublicKey string
	chainID      string
	nodes        []*commonv1.Node
	interval     time.Duration
}

// NewPluginRelayerConfigEmitter constructs a service that emits ChainPluginConfig to Beholder.
func NewPluginRelayerConfigEmitter(lggr logger.Logger, csaPublicKey, chainID string, rawNodes []map[string]string) *pluginRelayerConfigEmitter {
	if csaPublicKey == "" {
		csaPublicKey = beholder.GetClient().Config.AuthPublicKeyHex
		if csaPublicKey == "" {
			lggr.Warn("csa_public_key not configured for plugin relayer config emitter")
		}
	}
	if chainID == "" {
		lggr.Warn("chain_id not configured for plugin relayer config emitter")
	}

	emitter := &pluginRelayerConfigEmitter{
		csaPublicKey: csaPublicKey,
		chainID:      chainID,
		nodes:        normalizeNodes(rawNodes),
		interval:     defaultEmitInterval,
	}

	emitter.Service, emitter.eng = services.Config{
		Name:  serviceName,
		Start: emitter.start,
	}.NewServiceEngine(lggr)

	return emitter
}

func (e *pluginRelayerConfigEmitter) start(ctx context.Context) error {
	if e.interval <= 0 {
		e.interval = defaultEmitInterval
	}
	e.eng.Infow(
		"Starting plugin relayer config emitter",
		"interval", e.interval,
		"chainID", e.chainID,
		"csaPublicKeyPresent", e.csaPublicKey != "",
		"nodes", len(e.nodes),
	)
	e.eng.GoTick(services.NewTicker(e.interval), e.emit)
	return nil
}

func (e *pluginRelayerConfigEmitter) emit(ctx context.Context) {
	payload := e.buildConfig()
	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		e.eng.Errorw(
			"failed to marshal ChainPluginConfig",
			"err", err,
			"chainID", e.chainID,
			"nodes", len(payload.Nodes),
		)
		return
	}

	e.eng.Debugw(
		"Emitting ChainPluginConfig",
		"payloadBytes", len(payloadBytes),
		"chainID", e.chainID,
		"nodes", len(payload.Nodes),
	)

	err = beholder.GetEmitter().Emit(ctx, payloadBytes,
		beholder.AttrKeyDomain, beholderDomain,
		beholder.AttrKeyEntity, beholderEntity,
		beholder.AttrKeyDataSchema, beholderDataSchema,
	)
	if err != nil {
		e.eng.Errorw(
			"failed to emit ChainPluginConfig",
			"err", err,
			"payloadBytes", len(payloadBytes),
			"chainID", e.chainID,
			"nodes", len(payload.Nodes),
		)
		return
	}

	e.eng.Debugw(
		"Emitted ChainPluginConfig",
		"payloadBytes", len(payloadBytes),
		"chainID", e.chainID,
		"nodes", len(payload.Nodes),
	)
}

func (e *pluginRelayerConfigEmitter) buildConfig() *commonv1.ChainPluginConfig {
	return &commonv1.ChainPluginConfig{
		CsaPublicKey: e.csaPublicKey,
		ChainId:      e.chainID,
		Nodes:        e.nodes,
	}
}

// normalizeNodes sanitizes and filters URL map values for node entries.
func normalizeNodes(rawNodes []map[string]string) []*commonv1.Node {
	if len(rawNodes) == 0 {
		return nil
	}

	out := make([]*commonv1.Node, 0, len(rawNodes))
	for _, raw := range rawNodes {
		normalized := normalizeEndpoints(raw)
		if len(normalized) == 0 {
			continue
		}
		out = append(out, &commonv1.Node{Urls: normalized})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// normalizeEndpoints sanitizes and filters URL map values.
func normalizeEndpoints(raw map[string]string) map[string]string {
	if len(raw) == 0 {
		return nil
	}

	out := make(map[string]string, len(raw))
	for key, item := range raw {
		if strings.TrimSpace(key) == "" {
			continue
		}
		normalized := normalizeEndpoint(item)
		if normalized == "" {
			continue
		}
		out[key] = normalized
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// normalizeEndpoint returns only scheme://host (no port/userinfo/path/query/fragment).
// If the input has no scheme, the host-only string is returned.
func normalizeEndpoint(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}

	scheme, host, err := parseOriginURL(s)
	if err != nil {
		return ""
	}
	if host == "" {
		return ""
	}
	if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	if scheme == "" {
		return host
	}
	return scheme + "://" + host
}

// parseOriginURL is based on go-ethereum's parseOriginURL, adapted for our needs:
// - returns only scheme and hostname (port is discarded)
// - handles schemeless inputs with userinfo/port/path
func parseOriginURL(origin string) (string, string, error) {
	parsedURL, err := url.Parse(strings.ToLower(origin))
	if err != nil {
		return "", "", err
	}
	if strings.Contains(origin, "://") {
		return parsedURL.Scheme, parsedURL.Hostname(), nil
	}

	if hostURL, err := url.Parse("//" + origin); err == nil {
		if host := hostURL.Hostname(); host != "" {
			return "", host, nil
		}
	}

	hostname := parsedURL.Scheme
	if hostname == "" {
		hostname = origin
	}
	return "", hostname, nil
}

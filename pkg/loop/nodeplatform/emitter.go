package nodeplatform

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
	BeholderDomain     = "node-platform"
	BeholderEntity     = "common.v1.ChainPluginConfig"
	BeholderDataSchema = "common/v1/chain_plugin_config.proto"

	DefaultEmitInterval = time.Hour
	serviceName         = "NodePlatformEmitter"
)

type ChainPluginConfigEmitter struct {
	services.Service
	eng *services.Engine

	lggr logger.Logger

	csaPublicKey string
	chainID      string
	urls         []string
	interval     time.Duration
}

func NewChainPluginConfigEmitter(lggr logger.Logger, csaPublicKey, chainID string, rawURLs []string) *ChainPluginConfigEmitter {
	return NewChainPluginConfigEmitterWithInterval(lggr, csaPublicKey, chainID, rawURLs, DefaultEmitInterval)
}

func NewChainPluginConfigEmitterWithInterval(lggr logger.Logger, csaPublicKey, chainID string, rawURLs []string, interval time.Duration) *ChainPluginConfigEmitter {
	if interval <= 0 {
		interval = DefaultEmitInterval
	}

	if csaPublicKey == "" {
		csaPublicKey = beholder.GetClient().Config.AuthPublicKeyHex
		if csaPublicKey == "" {
			lggr.Warn("csa_public_key not configured for node-platform emitter")
		}
	}
	if chainID == "" {
		lggr.Warn("chain_id not configured for node-platform emitter")
	}

	emitter := &ChainPluginConfigEmitter{
		lggr:         lggr,
		csaPublicKey: csaPublicKey,
		chainID:      chainID,
		urls:         NormalizeEndpoints(rawURLs),
		interval:     interval,
	}

	emitter.Service, emitter.eng = services.Config{
		Name:  serviceName,
		Start: emitter.start,
	}.NewServiceEngine(lggr)

	return emitter
}

func (e *ChainPluginConfigEmitter) start(ctx context.Context) error {
	e.eng.GoTick(services.NewTicker(e.interval), e.emit)
	return nil
}

func (e *ChainPluginConfigEmitter) emit(ctx context.Context) {
	payload := e.buildConfig()
	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		e.lggr.Errorw("failed to marshal ChainPluginConfig", "err", err)
		return
	}

	err = beholder.GetEmitter().Emit(ctx, payloadBytes,
		beholder.AttrKeyDomain, BeholderDomain,
		beholder.AttrKeyEntity, BeholderEntity,
		beholder.AttrKeyDataSchema, BeholderDataSchema,
	)
	if err != nil {
		e.lggr.Errorw("failed to emit ChainPluginConfig", "err", err)
	}
}

func (e *ChainPluginConfigEmitter) buildConfig() *commonv1.ChainPluginConfig {
	return &commonv1.ChainPluginConfig{
		CsaPublicKey: e.csaPublicKey,
		ChainId:      e.chainID,
		Urls:         e.urls,
	}
}

// NormalizeEndpoints sanitizes and de-duplicates URLs, preserving order.
func NormalizeEndpoints(raw []string) []string {
	if len(raw) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		normalized := NormalizeEndpoint(item)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

// NormalizeEndpoint returns only scheme://host (no port/userinfo/path/query/fragment).
// If the input has no scheme, the host-only string is returned.
func NormalizeEndpoint(raw string) string {
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

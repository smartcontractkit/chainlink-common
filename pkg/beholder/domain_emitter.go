package beholder

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/batch"
	pkglogger "github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// Domain is a top-level product domain with its own chip ingress deployment.
type Domain string

// Domains are fixed in code, not configured: each one needs a matching load balancer rule and
// chip ingress deployment, so adding one is a reviewed change here plus infra.
const (
	DomainDataFeeds   Domain = "data-feeds"
	DomainDataStreams Domain = "data-streams"
	DomainCCIP        Domain = "ccip"
)

// allDomains lists every Domain; Config.ChipIngressDomainEmittersEnabled builds an emitter for
// each. A func, not a var, so the list cannot be modified.
func allDomains() []Domain {
	return []Domain{DomainDataFeeds, DomainDataStreams, DomainCCIP}
}

var (
	// ErrDomainEmittersDisabled is returned by ForDomain when the Client has no domain emitters:
	// Config.ChipIngressDomainEmittersEnabled is off, or the Client is a noop or HTTP client.
	ErrDomainEmittersDisabled = errors.New("beholder: chip ingress domain emitters are disabled")
	// ErrUnknownDomain is returned by ForDomain for a Domain the Client has no emitter for.
	ErrUnknownDomain = errors.New("beholder: unknown chip ingress domain")
)

// DomainHeaderKey is the gRPC metadata key carrying a domain emitter's Domain. Load balancers route
// on it to the chip ingress deployment that serves the domain; chip ingress itself ignores it.
const DomainHeaderKey = "chainlink-domain"

// DomainEmitter sends one domain's events to chip ingress only, with PublishBatchNoAck, on its own
// connection and buffer. Obtain one with Client.ForDomain.
//
// Both methods reject an event whose source (the CloudEvent source, or the beholder_domain
// attribute for Emit) is not the emitter's Domain or a subdomain of it. The match is
// case-sensitive and a subdomain must follow a "." separator: for DomainDataFeeds,
// "data-feeds" and "data-feeds.telemetry.ocr2-median" are accepted; "ccip.telemetry.ocr3",
// "data-feeds-extra" and "Data-Feeds" are not.
type DomainEmitter interface {
	// Emitter builds the event from attributes, like the Client's Emitter. Close is a no-op: the
	// Client starts and closes its domain emitters.
	Emitter
	// EmitEvent queues a fully built CloudEvent, keeping every extension the caller set. If it
	// returns a non-nil error the callback is not invoked; otherwise a non-nil callback fires
	// exactly once with the event's fate.
	EmitEvent(ctx context.Context, event chipingress.CloudEvent, callback func(error)) error
}

// ForDomain returns the Client's emitter for domain. It creates nothing, so repeated calls return
// the same emitter. The error wraps ErrDomainEmittersDisabled if the Client has no domain emitters,
// or ErrUnknownDomain if domain is not a known Domain.
func (c *Client) ForDomain(domain Domain) (DomainEmitter, error) {
	if c.domainEmitters == nil {
		return nil, fmt.Errorf("%w: no emitter for domain %q", ErrDomainEmittersDisabled, domain)
	}
	em, ok := c.domainEmitters[domain]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownDomain, domain)
	}
	return em, nil
}

// domainEmitter is a domain's batch emitter as handed out by ForDomain. Close is a no-op so holders
// cannot stop an emitter the Client owns.
type domainEmitter struct {
	*ChipIngressBatchEmitterService
	// chipClient is the emitter's own connection. Once started, the batch client closes it on
	// Stop; it is closed here only when construction fails before the emitter ever starts.
	chipClient chipingress.Client
}

func (*domainEmitter) Close() error { return nil }

// newDomainEmitters builds one emitter per Domain. opts are the Client's chip ingress options
// (transport, auth, OTel providers) without its header provider, which each domain emitter
// replaces with one carrying the resource attribute headers plus its DomainHeaderKey.
func newDomainEmitters(cfg Config, opts []chipingress.Opt, resourceAttrs map[string]string, lggr pkglogger.Logger) (_ map[Domain]*domainEmitter, err error) {
	domains := allDomains()
	emitters := make(map[Domain]*domainEmitter, len(domains))
	defer func() {
		if err != nil {
			for _, em := range emitters {
				_ = em.chipClient.Close()
			}
		}
	}()
	for _, domain := range domains {
		headers := chipingress.SanitizeMetadataHeaders(resourceAttrs)
		headers[DomainHeaderKey] = string(domain)
		clientName := batch.ClientNameBeholder + "_" + string(domain)

		domainOpts := slices.Concat(opts, []chipingress.Opt{
			chipingress.WithHeaderProvider(chipingress.NewStaticHeaderProvider(headers)),
			chipingress.WithNOPLookup(),
			chipingress.WithClientName(clientName),
		})
		chipClient, err := chipingress.NewClient(cfg.ChipIngressEmitterGRPCEndpoint, domainOpts...)
		if err != nil {
			return nil, fmt.Errorf("chip ingress client for domain %q: %w", domain, err)
		}
		svc, err := newChipIngressBatchEmitterService(chipClient, cfg, pkglogger.Named(lggr, string(domain)), clientName, domain, batch.WithNoAck())
		if err != nil {
			_ = chipClient.Close()
			return nil, fmt.Errorf("chip ingress emitter for domain %q: %w", domain, err)
		}
		emitters[domain] = &domainEmitter{ChipIngressBatchEmitterService: svc, chipClient: chipClient}
	}
	return emitters, nil
}

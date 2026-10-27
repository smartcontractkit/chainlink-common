package beholder

import (
	"context"
	"fmt"

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

// domains lists every Domain; Config.ChipIngressDomainEmittersEnabled builds an emitter for each.
var domains = []Domain{DomainDataFeeds, DomainDataStreams, DomainCCIP}

// DomainHeaderKey is the gRPC metadata key carrying a domain emitter's Domain. Load balancers route
// on it to the chip ingress deployment that serves the domain; chip ingress itself ignores it.
const DomainHeaderKey = "chainlink-domain"

// DomainEmitter sends one domain's events to chip ingress only, with PublishBatchNoAck, on its own
// connection and buffer. Obtain one with Client.ForDomain.
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
// the same emitter. It is an error unless Config.ChipIngressDomainEmittersEnabled is set and domain
// is a known Domain.
func (c *Client) ForDomain(domain Domain) (DomainEmitter, error) {
	em, ok := c.domainEmitters[domain]
	if !ok {
		return nil, fmt.Errorf("beholder: no chip ingress emitter for domain %q (ChipIngressDomainEmittersEnabled=%t)",
			domain, c.Config.ChipIngressDomainEmittersEnabled)
	}
	return em, nil
}

// domainEmitter is a domain's batch emitter as handed out by ForDomain. Close is a no-op so holders
// cannot stop an emitter the Client owns.
type domainEmitter struct {
	*ChipIngressBatchEmitterService
}

func (*domainEmitter) Close() error { return nil }

// newDomainEmitters builds one emitter per Domain. opts are the Client's chip ingress options
// (transport, auth, OTel providers) without its header provider, which each domain emitter
// replaces with one carrying the resource attribute headers plus its DomainHeaderKey.
func newDomainEmitters(cfg Config, opts []chipingress.Opt, resourceAttrs map[string]string, lggr pkglogger.Logger) (_ map[Domain]*domainEmitter, err error) {
	emitters := make(map[Domain]*domainEmitter, len(domains))
	defer func() {
		if err != nil {
			for _, em := range emitters {
				_ = em.ownedClient.Close()
			}
		}
	}()
	for _, domain := range domains {
		headers := chipingress.SanitizeMetadataHeaders(resourceAttrs)
		headers[DomainHeaderKey] = string(domain)
		clientName := batch.ClientNameBeholder + "_" + string(domain)

		domainOpts := append(append([]chipingress.Opt{}, opts...),
			chipingress.WithHeaderProvider(chipingress.NewStaticHeaderProvider(headers)),
			chipingress.WithNOPLookup(),
			chipingress.WithClientName(clientName),
		)
		chipClient, err := chipingress.NewClient(cfg.ChipIngressEmitterGRPCEndpoint, domainOpts...)
		if err != nil {
			return nil, fmt.Errorf("chip ingress client for domain %q: %w", domain, err)
		}
		svc, err := newChipIngressBatchEmitterService(chipClient, cfg, pkglogger.Named(lggr, string(domain)), clientName, batch.WithNoAck())
		if err != nil {
			_ = chipClient.Close()
			return nil, fmt.Errorf("chip ingress emitter for domain %q: %w", domain, err)
		}
		svc.ownedClient = chipClient
		emitters[domain] = &domainEmitter{svc}
	}
	return emitters, nil
}

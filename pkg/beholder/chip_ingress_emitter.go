package beholder

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"sync/atomic"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// resourceAttrExtensions holds resource attributes to stamp as CloudEvent extensions on every
// emitted event. It is stored behind a pointer on ChipIngressEmitter (rather than as a bare map
// field) so the struct itself stays a comparable type — a map field would make it incomparable,
// which is an exported-API-breaking change per apidiff. A nil *resourceAttrExtensions means no
// resource attributes are configured.
type resourceAttrExtensions struct {
	attrs map[string]string
}

// ChipIngressEmitter wraps a synchronous chipingress.Client.Publish call
// in a fire-and-forget goroutine so callers are never blocked.
type ChipIngressEmitter struct {
	client        chipingress.Client
	lggr          logger.Logger
	resourceAttrs *resourceAttrExtensions
	stopCh        services.StopChan
	wg            services.WaitGroup
	closed        atomic.Bool
}

func NewChipIngressEmitter(client chipingress.Client) (Emitter, error) {
	return ChipIngressEmitterConfig{}.New(client)
}

// ChipIngressEmitterConfig holds configuration for creating a ChipIngressEmitter.
type ChipIngressEmitterConfig struct {
	Lggr logger.Logger
}

// New creates a ChipIngressEmitter from the config, with no resource attributes configured.
func (c ChipIngressEmitterConfig) New(client chipingress.Client) (Emitter, error) {
	return c.NewWithResourceAttributes(client, nil)
}

// NewWithResourceAttributes creates a ChipIngressEmitter from the config, additionally stamping
// attrs as CloudEvent extensions (via chipingress.WithResourceAttributeExtensions) on every
// emitted event.
func (c ChipIngressEmitterConfig) NewWithResourceAttributes(client chipingress.Client, attrs map[string]string) (Emitter, error) {
	if client == nil {
		return nil, errors.New("chip ingress client is nil")
	}
	lggr := c.Lggr
	if lggr == nil {
		lggr = logger.Nop()
	}

	var resourceAttrs *resourceAttrExtensions
	if len(attrs) > 0 {
		resourceAttrs = &resourceAttrExtensions{attrs: attrs}
	}

	return &ChipIngressEmitter{
		client:        client,
		lggr:          lggr,
		resourceAttrs: resourceAttrs,
		stopCh:        make(services.StopChan),
	}, nil
}

func (c *ChipIngressEmitter) Close() error {
	if wasClosed := c.closed.Swap(true); wasClosed {
		return errors.New("already closed")
	}
	close(c.stopCh)
	c.wg.Wait()
	return c.client.Close()
}

// Emit fires a synchronous gRPC Publish call in a background goroutine
// so the caller is never blocked.
func (c *ChipIngressEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	sourceDomain, entityType, err := ExtractSourceAndType(attrKVs...)
	if err != nil {
		return err
	}

	var event chipingress.CloudEvent
	if c.resourceAttrs != nil {
		event, err = chipingress.NewEventWithOpts(sourceDomain, entityType, body, newAttributes(attrKVs...), chipingress.WithResourceAttributeExtensions(c.resourceAttrs.attrs))
	} else {
		event, err = chipingress.NewEvent(sourceDomain, entityType, body, newAttributes(attrKVs...))
	}
	if err != nil {
		return err
	}

	eventPb, err := chipingress.EventToProto(event)
	if err != nil {
		return fmt.Errorf("failed to convert event to proto: %w", err)
	}

	if err := c.wg.TryAdd(1); err != nil {
		return err
	}
	// Legacy ChipIngressEmitter.Emit is a synchronous gRPC call;
	// fire-and-forget via goroutine to avoid blocking the caller.
	go func(ctx context.Context) {
		defer c.wg.Done()
		var cancel context.CancelFunc
		ctx, cancel = c.stopCh.Ctx(ctx)
		defer cancel()

		if _, err := c.client.Publish(ctx, eventPb); err != nil {
			c.lggr.Infof("failed to emit to chip ingress: %v", err)
		}
	}(context.WithoutCancel(ctx))

	return nil
}

// ExtractSourceAndType extracts source domain and entity from the attributes
func ExtractSourceAndType(attrKVs ...any) (string, string, error) {
	attributes := newAttributes(attrKVs...)

	var sourceDomain string
	var entityType string

	for key, value := range attributes {
		// Retrieve source and type using either ChIP or legacy attribute names, prioritizing source/type
		if key == "source" || (key == AttrKeyDomain && sourceDomain == "") {
			if val, ok := value.(string); ok {
				sourceDomain = val
			}
		}
		if key == "type" || (key == AttrKeyEntity && entityType == "") {
			if val, ok := value.(string); ok {
				entityType = val
			}
		}
	}

	if sourceDomain == "" {
		return "", "", errors.New("source/beholder_domain not found in provided key/value attributes")
	}

	if entityType == "" {
		return "", "", errors.New("type/beholder_entity not found in provided key/value attributes")
	}

	return sourceDomain, entityType, nil
}

func ExtractAttributes(attrKVs ...any) map[string]any {
	attributes := newAttributes(attrKVs...)

	attributesMap := make(map[string]any)
	maps.Copy(attributesMap, attributes)

	return attributesMap
}

package beholder

import (
	"context"
	"fmt"
	"time"

	ch "github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	chpb "github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
)

const (
	defaultTimeout = 3 * time.Second
)

type ChipIngressEmitter struct {
	client  chpb.ChipIngressClient
	timeout time.Duration
}

type Opt func(*ChipIngressEmitter)

func WithTimeout(timeout time.Duration) Opt {
	return func(e *ChipIngressEmitter) {
		e.timeout = timeout
	}
}

func NewChipIngressEmitter(client chpb.ChipIngressClient, opts ...Opt) (Emitter, error) {

	if client == nil {
		return nil, fmt.Errorf("chip ingress client is nil")
	}
	e := &ChipIngressEmitter{client: client, timeout: defaultTimeout}
	for _, opt := range opts {
		opt(e)
	}
	return e, nil
}

func (c *ChipIngressEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	// newCtx returns a ctx with a timeout min(c.timeout, ctx.Deadline())
	newCtx := func(ctx context.Context) (context.Context, context.CancelFunc) {
		// check if ctx has a deadline and it's less than the emitter timeout,
		// then we need a new ctx with the emitter timeout
		if dl, ok := ctx.Deadline(); ok && time.Until(dl) < c.timeout {
			return context.WithTimeout(context.Background(), c.timeout)
		}
		return context.WithTimeout(ctx, c.timeout)
	}
	ctx, cancel := newCtx(ctx)
	defer cancel()

	sourceDomain, entityType, err := ExtractSourceAndType(attrKVs...)
	if err != nil {
		return err
	}

	attributes := ExtractAttributes(attrKVs...)

	event, err := ch.NewEventWithAttributes(sourceDomain, entityType, body, attributes)
	if err != nil {
		return err
	}
	_, err = c.client.Publish(ctx, event)
	if err != nil {
		return err
	}

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
		return "", "", fmt.Errorf("source/beholder_domain not found in provided key/value attributes")
	}

	if entityType == "" {
		return "", "", fmt.Errorf("type/beholder_entity not found in provided key/value attributes")
	}

	return sourceDomain, entityType, nil
}

func ExtractAttributes(attrKVs ...any) map[string]any {
	attributes := newAttributes(attrKVs...)

	attributesMap := make(map[string]any)
	for key, value := range attributes {
		attributesMap[key] = value
	}

	return attributesMap
}

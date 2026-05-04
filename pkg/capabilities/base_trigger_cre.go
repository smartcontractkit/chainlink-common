package capabilities

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/cresettings"
)

// ValidateBaseTriggerRetryInterval returns an error if the configured retry interval is not positive.
// Retransmit enablement is evaluated dynamically at runtime via BaseTriggerRetransmitEnabled.
func ValidateBaseTriggerRetryInterval(ctx context.Context, g settings.Getter) error {
	if g == nil {
		return errors.New("base trigger CRE settings getter is nil")
	}
	iv, err := cresettings.Default.BaseTriggerRetryInterval.GetOrDefault(ctx, g)
	if err != nil {
		return fmt.Errorf("base trigger retry interval: %w", err)
	}
	if iv <= 0 {
		return fmt.Errorf("BaseTriggerRetryInterval must be positive (got %v)", iv)
	}
	return nil
}

// NewBaseTriggerCapabilityWithCRESettings builds a [BaseTriggerCapability] that reads
// [cresettings.Default.BaseTriggerRetransmitEnabled] and [cresettings.Default.BaseTriggerRetryInterval]
// on each delivery, resend, and scan so changes apply without restarting the node.
func NewBaseTriggerCapabilityWithCRESettings[T proto.Message](
	ctx context.Context,
	store EventStore,
	newMsg func() T,
	lggr logger.Logger,
	capabilityID string,
	getter settings.Getter,
) (*BaseTriggerCapability[T], error) {
	if err := ValidateBaseTriggerRetryInterval(ctx, getter); err != nil {
		return nil, err
	}
	return NewBaseTriggerCapability(store, newMsg, lggr, capabilityID, 0, 0, 0, getter), nil
}

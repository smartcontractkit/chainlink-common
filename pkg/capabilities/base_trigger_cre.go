package capabilities

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/cresettings"
)

// ResolveBaseTriggerRetryInterval returns the retransmit ticker interval for [BaseTriggerCapability].
// When [cresettings.Default.BaseTriggerRetransmitEnabled] is false, it returns (0, nil) so the base
// trigger delivers fire-and-forget without persistence or ACK tracking.
// When enabled, [cresettings.Default.BaseTriggerRetryInterval] must be positive.
func ResolveBaseTriggerRetryInterval(ctx context.Context, g settings.Getter, lggr logger.Logger) (retryInterval time.Duration, err error) {
	enabled, gerr := cresettings.Default.BaseTriggerRetransmitEnabled.GetOrDefault(ctx, g)
	if gerr != nil {
		lggr.Errorw("CRE settings read failed for base trigger retransmit flag; using default", "err", gerr)
	}
	if !enabled {
		return 0, nil
	}
	retryInterval, gerr = cresettings.Default.BaseTriggerRetryInterval.GetOrDefault(ctx, g)
	if gerr != nil {
		lggr.Errorw("CRE settings read failed for base trigger retry interval; using default", "err", gerr)
	}
	if retryInterval <= 0 {
		return 0, fmt.Errorf(
			"BaseTriggerRetransmitEnabled is true but BaseTriggerRetryInterval must be positive (got %s)",
			retryInterval,
		)
	}
	return retryInterval, nil
}

// NewBaseTriggerCapabilityWithCRESettings builds a [BaseTriggerCapability] using global CRE settings
// for retransmit enablement and interval. Undelivered warning/critical thresholds are derived from
// the resolved interval when retransmit is enabled.
func NewBaseTriggerCapabilityWithCRESettings[T proto.Message](
	ctx context.Context,
	store EventStore,
	newMsg func() T,
	lggr logger.Logger,
	capabilityID string,
	getter settings.Getter,
) (*BaseTriggerCapability[T], error) {
	retry, err := ResolveBaseTriggerRetryInterval(ctx, getter, lggr)
	if err != nil {
		return nil, err
	}
	var undeliveredWarning, undeliveredCritical time.Duration
	if retry > 0 {
		undeliveredWarning = 5 * retry
		undeliveredCritical = 20 * retry
	}
	return NewBaseTriggerCapability(store, newMsg, lggr, capabilityID, retry, undeliveredWarning, undeliveredCritical), nil
}

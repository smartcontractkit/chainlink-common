package types

import "context"

// HeadTrackable is implemented by the core txm,
// to be able to receive head events from any chain.
// Chain implementations should notify head events to the core txm via this interface.
//
//go:generate mockery --quiet --name HeadTrackable --output ./mocks/ --case=underscore
type HeadTrackable[H Head[BLOCK_HASH], BLOCK_HASH Hashable] interface {
	// OnNewLongestChain sends a new head when it becomes available. Subscribers can recursively trace the parent
	// of the head to the finality depth back. If this is not possible (e.g. due to recent boot, backfill not complete
	// etc), users may get a shorter linked list. If there is a re-org, older blocks won't be sent to this function again.
	// But the new blocks from the re-org will be available in later blocks' parent linked list.
	OnNewLongestChain(ctx context.Context, head H)
}

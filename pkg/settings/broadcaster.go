package settings

import "context"

// Broadcaster provides a means of subscribing to a channel of [Snapshot]s.
// If an instance also implements [LocalBroadcaster], then Unsubscribe should be called before abandoning a channel.
type Broadcaster interface {
	Subscribe(context.Context) (<-chan Snapshot, error)
}

// LocalBroadcaster is implemented by Broadcaster instances which are local and need an explicit signal
// to unsubscribe from updates. Non-local instances automatically unsubscribe when the client/server pair are closed.
type LocalBroadcaster interface {
	Unsubscribe(<-chan Snapshot)
}

// Snapshot is a complete set of settings.
type Snapshot struct {
	Settings string // default format TOML
	Hash     string // default sha256
}

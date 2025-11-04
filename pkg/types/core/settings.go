package core

import "context"

// SettingsBroadcaster provides a means of subscribing to a channel of [SettingsUpdate]s.
// If an instance also implements SettingsBroacasterLocal, then Unsubscribe should be called before abandoning a channel.
type SettingsBroadcaster interface {
	Subscribe(context.Context) (<-chan SettingsUpdate, error)
}

// SettingsBroadcasterLocal is implemented by SettingsBroadcaster instances which are local and need an explicit signal
// to unsubscribe from updates. Non-local instances automatically unsubscribe when the client/server pair are closed.
type SettingsBroadcasterLocal interface {
	Unsubscribe(<-chan SettingsUpdate)
}

type SettingsUpdate struct {
	Settings string // default format TOML
	Hash     string // default sha256
}

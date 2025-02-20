package internal

import "go.opentelemetry.io/collector/component"

type nopHost struct{}

func NewNopHost() component.Host {
	return &nopHost{}
}

func (nh *nopHost) GetFactory(component.Kind, component.Type) component.Factory {
	return nil
}

func (nh *nopHost) GetExtensions() map[component.ID]component.Component {
	return nil
}

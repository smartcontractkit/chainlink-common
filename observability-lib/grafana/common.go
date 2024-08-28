package grafana

import "github.com/grafana/grafana-foundation-sdk/go/dashboard"

func datasourceRef(uid string) dashboard.DataSourceRef {
	return dashboard.DataSourceRef{Uid: &uid}
}

func Inc(p *uint32) uint32 {
	*p++
	return *p
}

func Pointer[T any](d T) *T {
	return &d
}

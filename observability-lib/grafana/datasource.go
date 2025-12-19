package grafana

import (
	"errors"

	"github.com/smartcontractkit/chainlink-common/observability-lib/api"
)

type DataSource struct {
	ID   uint
	Name string
	UID  string
	Type string
}

func NewDataSource(name, uid string) *DataSource {
	return &DataSource{
		Name: name,
		UID:  uid,
	}
}

func GetDataSourceFromGrafana(name string, grafanaURL string, grafanaToken string) (*DataSource, error) {
	grafanaClient := api.NewClient(
		grafanaURL,
		grafanaToken,
	)

	datasource, _, err := grafanaClient.GetDataSourceByName(name)
	if err != nil {
		return nil, err
	}
	if datasource.Name == "" {
		return nil, errors.New("unexpected empty response. please check connection or vpn settings")
	}

	return &DataSource{ID: datasource.ID, Name: datasource.Name, UID: datasource.UID, Type: datasource.Type}, nil
}

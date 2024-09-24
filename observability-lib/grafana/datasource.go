package grafana

import "github.com/smartcontractkit/chainlink-common/observability-lib/api"

type DataSource struct {
	Name string
	UID  string
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

	return &DataSource{Name: datasource.Name, UID: datasource.UID}, nil
}

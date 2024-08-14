package grafana

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

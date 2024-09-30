# Observability-lib

This library enables creating Grafana dashboards and alerts with go code.

It provides abstractions to create grafana resources :
- [Dashboards](https://grafana.com/docs/grafana/latest/dashboards/)
- [Alerts](https://grafana.com/docs/grafana/latest/alerting/)
- [Contact Points](https://grafana.com/docs/grafana/latest/alerting/fundamentals/notifications/contact-points/)
- [Notification Policies](https://grafana.com/docs/grafana/latest/alerting/configure-notifications/create-notification-policy/)
- [Notification Templates](https://grafana.com/docs/grafana/latest/alerting/configure-notifications/template-notifications/create-notification-templates/)

## Folder Structure

The observability-lib is structured as follows:
```shell
observability-lib/
    api/ # Grafana HTTP API Client to interact with resources
    cmd/ # CLI to interact deploy or generateJSON from dashboards defined in folder below
    dashboards/ # Dashboards definitions
    grafana/ # grafana-foundations-sdk abstraction to manipulate grafana resources
```

## Quickstart

### Creating a dashboard

```go
package main

import "github.com/smartcontractkit/chainlink-common/observability-lib/grafana"

builder := grafana.NewBuilder(&grafana.BuilderOptions{
    Name:       "Dashboard Name",
    Tags:       []string{"tags1", "tags2"},
    Refresh:    "30s",
    TimeFrom:   "now-30m",
    TimeTo:     "now",
})

builder.AddVars(grafana.NewQueryVariable(&grafana.QueryVariableOptions{
    VariableOption: &grafana.VariableOption{
        Label: "Environment",
        Name:  "env",
    },
    Datasource: "Prometheus",
    Query:      `label_values(up, env)`,
}))

builder.AddRow("Summary")

builder.AddPanel(grafana.NewStatPanel(&grafana.StatPanelOptions{
    PanelOptions: &grafana.PanelOptions{
        Datasource:  "Prometheus",
        Title:       "Uptime",
        Description: "instance uptime",
        Span:        12,
        Height:      4,
        Decimals:    2,
        Unit:        "s",
        Query: []grafana.Query{
            {
                Expr:   `uptime_seconds`,
                Legend: `{{ pod }}`,
            },
        },
    },
    ColorMode:   common.BigValueColorModeNone,
    TextMode:    common.BigValueTextModeValueAndName,
    Orientation: common.VizOrientationHorizontal,
}))

db, err := builder.Build()
if err != nil {
    return nil, err
}
json, err := db.GenerateJSON()
if err != nil {
    return nil, err
}
fmt.Println(string(json))
```

More advanced examples can be found in the `dashboards` folder.

## Cmd Usage

The CLI can be used to :
- Deploy dashboards and alerts to grafana
- Generate JSON from dashboards defined in the `dashboards` folder

Example to deploy a dashboard to grafana instance using URL and token:
```shell
make build
./observability-lib deploy \
  --dashboard-name DashboardName \
  --dashboard-folder FolderName \
  --grafana-url $GRAFANA_URL \
  --grafana-token $GRAFANA_TOKEN \
  --type core-node \
  --platform kubernetes \
  --metrics-datasource Prometheus
```
To see how to get a grafana token you can check this [page](https://grafana.com/docs/grafana/latest/administration/service-accounts/)

Example to generate JSON from a dashboard defined in the `dashboards` folder:
```shell
make build
./observability-lib generate \
  --dashboard-name DashboardName \
  --type core-node-components \
  --platform kubernetes
```

## Makefile Usage


To build the observability library, run the following command:

```bash
make build
```

To run the tests, run the following command:

```bash
make test
```

To run the linter, run the following command:

```bash
make lint
```

To run the CLI
```bash
make run
```

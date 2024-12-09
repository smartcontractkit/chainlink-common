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
    cmd/ # CLI
    grafana/ # grafana-foundations-sdk abstraction to manipulate grafana resources
```

## Documentation

Godoc generated documentation is available [here](https://pkg.go.dev/github.com/smartcontractkit/chainlink-common/observability-lib)

## Quickstart

### Creating a dashboard

<details><summary>main.go</summary>

```go
package main

import "github.com/smartcontractkit/chainlink-common/observability-lib/grafana"

func main() {
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
}
```
</details>

## Cmd Usage

CLI to manipulate grafana resources

### Contact Point

#### List

```shell
./observability-lib api contact-point list \
  --grafana-url http://localhost:3000 \
  --grafana-token <token>
```

#### Delete

```shell
./observability-lib api contact-point delete <name> \
  --grafana-url http://localhost:3000 \
  --grafana-token <token>
```

### Dashboard

#### Delete

```shell
./observability-lib api dashboard delete <name> \
  --grafana-url http://localhost:3000 \
  --grafana-token <token>
```

### Notification Policy

#### List

```shell
./observability-lib api notification-policy list \
  --grafana-url http://localhost:3000 \
  --grafana-token <token>
```

#### Delete

```shell
./observability-lib api notification-policy delete <receiverName> \ 
  --grafana-url http://localhost:3000 \
  --grafana-token <token> \
  --matchers key,=,value \
  --matchers key2,=,value2
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

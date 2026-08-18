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

There are two ways to add panels to a dashboard:

- **`AddPanel`**: adds a panel directly to the dashboard as a top-level element. Panels appear in the order they are added.
- **`AddPanelToRow`**: adds a panel inside a row. Rows with panels are automatically **collapsed** in Grafana, meaning their panels are nested and hidden until the user expands the row.

Rows without any panels added via `AddPanelToRow` remain open (not collapsed).

You can freely interleave `AddRow`, `AddPanel`, and `AddPanelToRow` calls — the dashboard will preserve the insertion order.

#### Basic dashboard with top-level panels

```go
package main

import (
	"fmt"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

func main() {
	builder := grafana.NewBuilder(&grafana.BuilderOptions{
		Name:     "Dashboard Name",
		Tags:     []string{"tags1", "tags2"},
		Refresh:  "30s",
		TimeFrom: "now-30m",
		TimeTo:   "now",
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
			Title:       grafana.Pointer("Uptime"),
			Description: "instance uptime",
			Span:        12,
			Height:      4,
			Decimals:    grafana.Pointer(2.),
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
		return
	}
	json, err := db.GenerateJSON()
	if err != nil {
		return
	}
	fmt.Println(string(json))
}
```

#### Dashboard with collapsed rows

Use `AddPanelToRow` to nest panels inside a row. The row will be automatically collapsed in Grafana, so users can expand it to see the panels inside.

```go
builder := grafana.NewBuilder(&grafana.BuilderOptions{
	Name:    "Dashboard With Collapsed Rows",
	Tags:    []string{"example"},
	Refresh: "30s",
})

// A collapsed row with multiple panels nested inside
builder.AddRow("Resource Usage")
builder.AddPanelToRow("Resource Usage",
	grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: "Prometheus",
			Title:      grafana.Pointer("CPU Usage"),
			Span:       12,
			Height:     8,
			Query: []grafana.Query{
				{
					Expr:   `rate(cpu_usage_seconds_total[5m])`,
					Legend: `{{ pod }}`,
				},
			},
		},
	}),
	grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: "Prometheus",
			Title:      grafana.Pointer("Memory Usage"),
			Span:       12,
			Height:     4,
			Query: []grafana.Query{
				{
					Expr:   `memory_usage_bytes`,
					Legend: `{{ pod }}`,
				},
			},
		},
	}),
)

// Another collapsed row
builder.AddRow("Network")
builder.AddPanelToRow("Network",
	grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: "Prometheus",
			Title:      grafana.Pointer("Network I/O"),
			Span:       24,
			Height:     8,
			Query: []grafana.Query{
				{
					Expr:   `rate(network_bytes_total[5m])`,
					Legend: `{{ interface }}`,
				},
			},
		},
	}),
)
```

#### Mixing top-level panels, open rows, and collapsed rows

You can combine all three patterns in a single dashboard. The order of calls determines the layout.

```go
builder := grafana.NewBuilder(&grafana.BuilderOptions{
	Name:    "Mixed Layout Dashboard",
	Refresh: "30s",
})

// Top-level panel (not inside any row)
builder.AddPanel(grafana.NewStatPanel(&grafana.StatPanelOptions{
	PanelOptions: &grafana.PanelOptions{
		Title: grafana.Pointer("Health Status"),
		Span:  24,
	},
}))

// Open row (no panels added via AddPanelToRow, so it stays expanded)
builder.AddRow("Overview")

// Top-level panel after the open row
builder.AddPanel(grafana.NewStatPanel(&grafana.StatPanelOptions{
	PanelOptions: &grafana.PanelOptions{
		Title: grafana.Pointer("Request Rate"),
		Span:  12,
	},
}))

// Collapsed row with panels inside
builder.AddRow("Details")
builder.AddPanelToRow("Details",
	grafana.NewTablePanel(&grafana.TablePanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Title: grafana.Pointer("Request Log"),
			Span:  24,
		},
	}),
	grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Title: grafana.Pointer("Latency Over Time"),
			Span:  24,
		},
	}),
)

// Another top-level panel after the collapsed row
builder.AddPanel(grafana.NewStatPanel(&grafana.StatPanelOptions{
	PanelOptions: &grafana.PanelOptions{
		Title: grafana.Pointer("Error Rate"),
		Span:  12,
	},
}))

// Resulting layout:
// 1. Health Status       (top-level panel)
// 2. Overview            (open row, expanded)
// 3. Request Rate        (top-level panel)
// 4. Details             (collapsed row, click to expand)
//    ├─ Request Log      (nested inside row)
//    └─ Latency Over Time(nested inside row)
// 5. Error Rate          (top-level panel)
```

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

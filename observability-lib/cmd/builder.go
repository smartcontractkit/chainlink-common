package cmd

import (
	"errors"
	"github.com/smartcontractkit/chainlink-common/observability-lib/capabilities"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"

	atlasdon "github.com/smartcontractkit/chainlink-common/observability-lib/atlas-don"
	corenode "github.com/smartcontractkit/chainlink-common/observability-lib/core-node"
	corenodecomponents "github.com/smartcontractkit/chainlink-common/observability-lib/core-node-components"
	k8sresources "github.com/smartcontractkit/chainlink-common/observability-lib/k8s-resources"
	nopocr "github.com/smartcontractkit/chainlink-common/observability-lib/nop-ocr"
)

type TypeDashboard string

const (
	TypeDashboardCoreNode           TypeDashboard = "core-node"
	TypeDashboardCoreNodeComponents TypeDashboard = "core-node-components"
	TypeDashboardCoreNodeResources  TypeDashboard = "core-node-resources"
	TypeDashboardDONOCR             TypeDashboard = "don-ocr"
	TypeDashboardDONOCR2            TypeDashboard = "don-ocr2"
	TypeDashboardDONOCR3            TypeDashboard = "don-ocr3"
	TypeDashboardNOPOCR2            TypeDashboard = "nop-ocr2"
	TypeDashboardNOPOCR3            TypeDashboard = "nop-ocr3"
	TypeDashboardCapabilities       TypeDashboard = "capabilities"
)

type OCRVersion string

const (
	OCRVersionOCR  OCRVersion = "ocr"
	OCRVersionOCR2 OCRVersion = "ocr2"
	OCRVersionOCR3 OCRVersion = "ocr3"
)

type BuildOptions struct {
	Name              string
	FolderUID         string
	Platform          grafana.TypePlatform
	TypeDashboard     TypeDashboard
	MetricsDataSource *grafana.DataSource
	LogsDataSource    *grafana.DataSource
	AlertsTags        map[string]string
	SlackWebhookURL   string
	SlackToken        string
	SlackChannel      string
}

func Build(options *BuildOptions) (*grafana.Dashboard, error) {
	var db *grafana.Dashboard
	var err error

	dashboardOptions := &grafana.DashboardOptions{
		Name:              options.Name,
		MetricsDataSource: options.MetricsDataSource,
		FolderUID:         options.FolderUID,
		Platform:          options.Platform,
		AlertsTags:        options.AlertsTags,
		SlackWebhookURL:   options.SlackWebhookURL,
		SlackToken:        options.SlackToken,
		SlackChannel:      options.SlackChannel,
	}

	switch options.TypeDashboard {
	case TypeDashboardCoreNode:
		db, err = corenode.NewDashboard(dashboardOptions)
	case TypeDashboardCoreNodeComponents:
		db, err = corenodecomponents.NewDashboard(dashboardOptions)
	case TypeDashboardCoreNodeResources:
		if options.Platform != grafana.TypePlatformKubernetes {
			return nil, errors.New("core-node-resources dashboard is only available for kubernetes")
		}
		db, err = k8sresources.NewDashboard(dashboardOptions)
	case TypeDashboardDONOCR:
		dashboardOptions.OCRVersion = string(OCRVersionOCR)
		db, err = atlasdon.NewDashboard(dashboardOptions)
	case TypeDashboardDONOCR2:
		dashboardOptions.OCRVersion = string(OCRVersionOCR2)
		db, err = atlasdon.NewDashboard(dashboardOptions)
	case TypeDashboardDONOCR3:
		dashboardOptions.OCRVersion = string(OCRVersionOCR3)
		db, err = atlasdon.NewDashboard(dashboardOptions)
	case TypeDashboardNOPOCR2:
		dashboardOptions.OCRVersion = string(OCRVersionOCR2)
		db, err = nopocr.NewDashboard(dashboardOptions)
	case TypeDashboardNOPOCR3:
		dashboardOptions.OCRVersion = string(OCRVersionOCR3)
		db, err = nopocr.NewDashboard(dashboardOptions)
	case TypeDashboardCapabilities:
		dashboardOptions.OCRVersion = string(OCRVersionOCR3)
		db, err = capabilities.NewDashboard(dashboardOptions)
	default:
		return nil, errors.New("invalid dashboard type")
	}

	return db, err
}

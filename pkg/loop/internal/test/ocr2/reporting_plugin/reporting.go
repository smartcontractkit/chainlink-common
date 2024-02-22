package reportingplugin_test

import (
	"bytes"
	"context"
	"fmt"

	"github.com/stretchr/testify/assert"

	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

type ReportingPluginTestConfig struct {
	ReportContext          libocr.ReportContext
	Query                  libocr.Query
	Observation            libocr.Observation
	AttributedObservations []libocr.AttributedObservation
	Report                 libocr.Report
	ShouldReport           bool
	ShouldAccept           bool
	ShouldTransmit         bool
}

var _ libocr.ReportingPlugin = StaticReportingPlugin{}

type StaticReportingPlugin struct {
	ReportingPluginTestConfig
}

func (s StaticReportingPlugin) Query(ctx context.Context, timestamp libocr.ReportTimestamp) (libocr.Query, error) {
	if timestamp != s.ReportingPluginTestConfig.ReportContext.ReportTimestamp {
		return nil, errExpected(s.ReportingPluginTestConfig.ReportContext.ReportTimestamp, timestamp)
	}
	return s.ReportingPluginTestConfig.Query, nil
}

func (s StaticReportingPlugin) Observation(ctx context.Context, timestamp libocr.ReportTimestamp, q libocr.Query) (libocr.Observation, error) {
	if timestamp != s.ReportingPluginTestConfig.ReportContext.ReportTimestamp {
		return nil, errExpected(s.ReportingPluginTestConfig.ReportContext.ReportTimestamp, timestamp)
	}
	if !bytes.Equal(q, s.ReportingPluginTestConfig.Query) {
		return nil, errExpected(s.ReportingPluginTestConfig.Query, q)
	}
	return s.ReportingPluginTestConfig.Observation, nil
}

func (s StaticReportingPlugin) Report(ctx context.Context, timestamp libocr.ReportTimestamp, q libocr.Query, observations []libocr.AttributedObservation) (bool, libocr.Report, error) {
	if timestamp != s.ReportingPluginTestConfig.ReportContext.ReportTimestamp {
		return false, nil, errExpected(s.ReportingPluginTestConfig.ReportContext.ReportTimestamp, timestamp)
	}
	if !bytes.Equal(q, s.ReportingPluginTestConfig.Query) {
		return false, nil, errExpected(s.ReportingPluginTestConfig.Query, q)
	}
	if !assert.ObjectsAreEqual(s.ReportingPluginTestConfig.AttributedObservations, observations) {
		return false, nil, errExpected(s.ReportingPluginTestConfig.AttributedObservations, observations)
	}
	return s.ReportingPluginTestConfig.ShouldReport, s.ReportingPluginTestConfig.Report, nil
}

func (s StaticReportingPlugin) ShouldAcceptFinalizedReport(ctx context.Context, timestamp libocr.ReportTimestamp, r libocr.Report) (bool, error) {
	if timestamp != s.ReportingPluginTestConfig.ReportContext.ReportTimestamp {
		return false, errExpected(s.ReportingPluginTestConfig.ReportContext.ReportTimestamp, timestamp)
	}
	if !bytes.Equal(r, s.ReportingPluginTestConfig.Report) {
		return false, errExpected(s.ReportingPluginTestConfig.Report, r)
	}
	return shouldAccept, nil
}

func (s StaticReportingPlugin) ShouldTransmitAcceptedReport(ctx context.Context, timestamp libocr.ReportTimestamp, r libocr.Report) (bool, error) {
	if timestamp != s.ReportingPluginTestConfig.ReportContext.ReportTimestamp {
		return false, errExpected(s.ReportingPluginTestConfig.ReportContext.ReportTimestamp, timestamp)
	}
	if !bytes.Equal(r, s.ReportingPluginTestConfig.Report) {
		return false, errExpected(s.ReportingPluginTestConfig.Report, r)
	}
	return shouldTransmit, nil
}

func (s StaticReportingPlugin) Close() error { return nil }

func errExpected(expected, got any) error {
	return fmt.Errorf("expected %v but got %v", expected, got)
}

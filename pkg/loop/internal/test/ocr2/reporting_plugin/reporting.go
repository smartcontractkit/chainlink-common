package reportingplugin_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func (s StaticReportingPlugin) AssertEqual(t *testing.T, ctx context.Context, rp libocr.ReportingPlugin) {
	gotQuery, err := rp.Query(ctx, reportContext.ReportTimestamp)
	require.NoError(t, err)
	assert.Equal(t, query, []byte(gotQuery))
	gotObs, err := rp.Observation(ctx, reportContext.ReportTimestamp, query)
	require.NoError(t, err)
	assert.Equal(t, observation, gotObs)
	gotOk, gotReport, err := rp.Report(ctx, reportContext.ReportTimestamp, query, obs)
	require.NoError(t, err)
	assert.True(t, gotOk)
	assert.Equal(t, report, gotReport)
	gotShouldAccept, err := rp.ShouldAcceptFinalizedReport(ctx, reportContext.ReportTimestamp, report)
	require.NoError(t, err)
	assert.True(t, gotShouldAccept)
	gotShouldTransmit, err := rp.ShouldTransmitAcceptedReport(ctx, reportContext.ReportTimestamp, report)
	require.NoError(t, err)
	assert.True(t, gotShouldTransmit)
}

func errExpected(expected, got any) error {
	return fmt.Errorf("expected %v but got %v", expected, got)
}

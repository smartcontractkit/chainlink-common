package textparse

import (
	"bytes"
	"fmt"
	"io"

	"github.com/gogo/protobuf/proto"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/prometheus/model/labels"
	promtextparse "github.com/prometheus/prometheus/model/textparse"
	dto "github.com/prometheus/prometheus/prompb/io/prometheus/client"
)

func convertMetricFamilyPb(srcMf *io_prometheus_client.MetricFamily, dst *dto.MetricFamily) (n int, err error) {
	protoBuf, err := proto.Marshal(srcMf)
	if err != nil {
		return 0, err
	}
	dst.Reset()
	err = dst.Unmarshal(protoBuf)
	if err != nil {
		return 0, err
	}
	return len(protoBuf), nil
}

type ProtobufParserShim struct {
	*ProtobufParser
	mfs   []*io_prometheus_client.MetricFamily
	index int
}

// Used to override readDelimited method of the ProtobufParser.
func (p *ProtobufParserShim) readDelimited(b []byte, mf *dto.MetricFamily) (n int, err error) {
	if p == nil || p.index >= len(p.mfs) {
		return 0, io.EOF
	}
	// Copies proto message from io_prometheus_client.MetricFamily to dto.MetricFamily
	_, err = convertMetricFamilyPb(p.mfs[p.index], mf)
	if err != nil {
		// todo: test this
		return 0, fmt.Errorf("failed to convert io_prometheus_client.MetricFamily to dto.MetricFamily: %w", err)
	}
	p.index++
	return 0, nil
}

func NewProtobufParserShim(parseClassicHistograms bool, st *labels.SymbolTable, mfs []*io_prometheus_client.MetricFamily) promtextparse.Parser {
	p := &ProtobufParserShim{&ProtobufParser{
		in:                     []byte{},
		state:                  promtextparse.EntryInvalid,
		mf:                     &dto.MetricFamily{},
		metricBytes:            &bytes.Buffer{},
		parseClassicHistograms: parseClassicHistograms,
		builder:                labels.NewScratchBuilderWithSymbolTable(st, 16),
	}, mfs, 0}
	// Overrides readDelimited method of the ProtobufParser
	p.ProtobufParser.readDelimitedFunc = p.readDelimited
	return p
}

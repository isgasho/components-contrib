// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package zipkin

import (
	"encoding/json"
	"strconv"

	"contrib.go.opencensus.io/exporter/zipkin"
	"github.com/dapr/components-contrib/exporters"
	"github.com/dapr/dapr/pkg/logger"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
	"go.opencensus.io/trace"
)

// Metadata is the zipkin config
type zipkinMetadata struct {
	ExporterAddress string `json:"exporterAddress"`
	Enabled         string `json:"enabled"`
}

// NewZipkinExporter returns a new zipkin exporter instance
func NewZipkinExporter(logger logger.Logger) *Exporter {
	return &Exporter{logger: logger}
}

// Exporter is an OpenCensus zipkin exporter
type Exporter struct {
	logger        logger.Logger
	traceExporter trace.Exporter
}

// Init creates a new zipkin endpoint and reporter
func (z *Exporter) Init(daprID string, hostAddress string, metadata exporters.Metadata) error {
	meta, err := z.getZipkinMetadata(metadata)
	if err != nil {
		return err
	}

	enabled, _ := strconv.ParseBool(meta.Enabled)
	if !enabled {
		return nil
	}

	localEndpoint, err := openzipkin.NewEndpoint(daprID, hostAddress)
	if err != nil {
		return err
	}
	reporter := zipkinHTTP.NewReporter(meta.ExporterAddress)
	z.traceExporter = zipkin.NewExporter(reporter, localEndpoint)
	trace.RegisterExporter(z.traceExporter)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	return nil
}

func (z *Exporter) getZipkinMetadata(metadata exporters.Metadata) (*zipkinMetadata, error) {
	b, err := json.Marshal(metadata.Properties)
	if err != nil {
		return nil, err
	}

	var zipkinMeta zipkinMetadata
	err = json.Unmarshal(b, &zipkinMeta)
	if err != nil {
		return nil, err
	}

	return &zipkinMeta, nil
}

// Unregister removes the exporter
func (z *Exporter) Unregister() {
	trace.UnregisterExporter(z.traceExporter)
}

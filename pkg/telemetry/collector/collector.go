package collector

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"git.backbone/corpix/unregistry/pkg/meta"
)

const Delimiter = "_"

var (
	NewCounter      = prometheus.NewCounter
	NewCounterVec   = prometheus.NewCounterVec
	NewGauge        = prometheus.NewGauge
	NewGaugeVec     = prometheus.NewGaugeVec
	NewHistogram    = prometheus.NewHistogram
	NewHistogramVec = prometheus.NewHistogramVec
)

type (
	Counter     = prometheus.Counter
	CounterVec  = prometheus.CounterVec
	CounterOpts = prometheus.CounterOpts

	Gauge     = prometheus.Gauge
	GaugeVec  = prometheus.GaugeVec
	GaugeOpts = prometheus.GaugeOpts

	Histogram     = prometheus.Histogram
	HistogramVec  = prometheus.HistogramVec
	HistogramOpts = prometheus.HistogramOpts

	Labels = prometheus.Labels
)

func NamePart(xs ...string) string {
	xxs := []string{}

	for _, x := range xs {
		if x != "" {
			xxs = append(xxs, x)
		}
	}

	return strings.Join(xxs, Delimiter)
}

func Name(subsystem string, name string, rest ...string) string {
	return NamePart(append(
		[]string{meta.TelemetryNamespace, subsystem, name},
		rest...,
	)...)
}

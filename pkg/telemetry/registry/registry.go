package registry

import (
	"github.com/prometheus/client_golang/prometheus"

	"git.backbone/corpix/unregistry/pkg/meta"
	"git.backbone/corpix/unregistry/pkg/telemetry/collector"
)

type Registry = prometheus.Registry

var DefaultRegistry = NewRegistry()

func init() {
	DefaultRegistry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	DefaultRegistry.MustRegister(prometheus.NewGoCollector())

	DefaultRegistry.MustRegister(
		&collector.Self{
			Self: prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					"go_build_info",
					"Meta information about application",
					nil, prometheus.Labels{
						"name":    meta.Name,
						"version": meta.Version,
					},
				),
				prometheus.GaugeValue, 1,
			),
		},
	)

}

func NewRegistry() *Registry {
	return prometheus.NewRegistry()
}

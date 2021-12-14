package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Self implements Collector for a single Metric so that the Metric
// collects itself. Add it as an anonymous field to a struct that implements
// Metric, and call init with the Metric itself as an argument.
type Self struct {
	Self prometheus.Metric
}

// Describe implements Collector.
func (c *Self) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Self.Desc()
}

// Collect implements Collector.
func (c *Self) Collect(ch chan<- prometheus.Metric) {
	ch <- c.Self
}

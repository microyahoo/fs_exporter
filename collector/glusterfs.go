package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// GlusterfsCollector defines structure of glusterfs stats
type GlusterfsCollector struct {
	logger *zap.Logger
}

// Update implements Collector.Update
func (c *GlusterfsCollector) Update(ch chan<- prometheus.Metric) error {
	return nil
}

// NewGlusterfsCollector returns a new Collector exposing glusterfs stats.
func NewGlusterfsCollector(logger *zap.Logger) (Collector, error) {
	return &ZfsCollector{
		logger: logger,
	}, nil
}

func init() {
	registerCollector("glusterfs", NewGlusterfsCollector)
}

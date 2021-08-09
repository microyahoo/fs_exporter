package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// ZfsCollector defines structure of zfs stats
type ZfsCollector struct {
	logger *zap.Logger
}

// Update implements Collector.Update
func (c *ZfsCollector) Update(ch chan<- prometheus.Metric) error {
	return nil
}

// NewZfsCollector returns a new Collector exposing zfs stats.
func NewZfsCollector(logger *zap.Logger) (Collector, error) {
	return &ZfsCollector{
		logger: logger,
	}, nil
}

func init() {
	registerCollector("zfs", NewZfsCollector)
}

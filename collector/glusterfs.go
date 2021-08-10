package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// glusterfs parameters
var (
	GlusterExecPath string
	GlusterVolumes  []string
	GlusterProfile  bool
	GlusterQuota    bool
)

// GlusterfsCollector defines structure of glusterfs stats
type GlusterfsCollector struct {
	logger *zap.Logger
}

// Update implements Collector.Update
func (c *GlusterfsCollector) Update(ch chan<- prometheus.Metric) error {
	c.logger.Debug("gluster options", zap.String("gluster.executable-path", GlusterExecPath),
		zap.Any("gluster.volumes", GlusterVolumes),
		zap.Bool("gluster.profile", GlusterProfile),
		zap.Bool("gluster.quota", GlusterQuota))
	return nil
}

// NewGlusterfsCollector returns a new Collector exposing glusterfs stats.
func NewGlusterfsCollector(logger *zap.Logger) (Collector, error) {
	return &GlusterfsCollector{
		logger: logger,
	}, nil
}

func init() {
	registerCollector("glusterfs", NewGlusterfsCollector)
}

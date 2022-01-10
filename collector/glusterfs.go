package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

const (
	// glusterCmd is the default path to gluster binary
	glusterCmd = "/usr/sbin/gluster"
)

// glusterfs parameters
var (
	glusterExecPath string
	glusterVolumes  []string
	glusterProfile  bool
	glusterQuota    bool
)

// GlusterfsCollector defines structure of glusterfs stats
type GlusterfsCollector struct {
	logger *zap.Logger
}

// Update implements Collector.Update
func (c *GlusterfsCollector) Update(ch chan<- prometheus.Metric) error {
	c.logger.Debug("gluster options", zap.String("gluster.executable-path", glusterExecPath),
		zap.Any("gluster.volumes", glusterVolumes),
		zap.Bool("gluster.profile", glusterProfile),
		zap.Bool("gluster.quota", glusterQuota))
	return nil
}

// NewGlusterfsCollector returns a new Collector exposing glusterfs stats.
func NewGlusterfsCollector(logger *zap.Logger) (Collector, error) {
	return &GlusterfsCollector{
		logger: logger,
	}, nil
}

func AddGlusterFlags(flags *pflag.FlagSet) {
	flags.StringVar(&glusterExecPath, "gluster.executable-path", glusterCmd, "Path to glusterfs executable")
	flags.StringSliceVar(&glusterVolumes, "gluster.volumes", []string{"_all"}, fmt.Sprintf("Comma separated volume names: vol1,vol2,vol3. Default is '%s' to scrape all metrics", "_all"))
	flags.BoolVar(&glusterProfile, "gluster.profile", false, "Enable gluster profiling reports")
	flags.BoolVar(&glusterQuota, "gluster.quota", false, "Enable gluster quota reports")
}

func init() {
	registerCollector("glusterfs", NewGlusterfsCollector)
}

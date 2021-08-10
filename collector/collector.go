package collector

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// Namespace defines the common namespace to be used by all metrics.
const namespace = "fs"

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"node_exporter: Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"node_exporter: Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)
)

var (
	collectorMutex        sync.Mutex
	factoryMutex          sync.Mutex
	initializedCollectors = make(map[string]Collector)
	collectorFactories    = make(map[string]func(*zap.Logger) (Collector, error))
)

func registerCollector(name string, factory func(*zap.Logger) (Collector, error)) {
	factoryMutex.Lock()
	defer factoryMutex.Unlock()
	if _, ok := collectorFactories[name]; ok {
		panic(fmt.Sprintf("The collector of %s has already been registered", name))
	}
	collectorFactories[name] = factory
}

// Collector is the interface a collector has to implement.
type Collector interface {
	// Get new metrics and expose them via prometheus registry.
	Update(ch chan<- prometheus.Metric) error
}

// FSCollector implements the prometheus.Collector interface.
type FSCollector struct {
	Collectors map[string]Collector
	logger     *zap.Logger
}

// NewFSCollector creates a new fs collector
func NewFSCollector(logger *zap.Logger) (*FSCollector, error) {
	collectors := make(map[string]Collector)

	collectorMutex.Lock()
	defer collectorMutex.Unlock()
	for name, factory := range collectorFactories {
		if c, ok := initializedCollectors[name]; ok {
			collectors[name] = c
		} else {
			c, err := factory(logger)
			if err != nil {
				return nil, err
			}
			collectors[name] = c
			initializedCollectors[name] = c
		}
	}
	return &FSCollector{
		Collectors: collectors,
		logger:     logger,
	}, nil
}

// Describe implements the prometheus.Collector interface.
func (n *FSCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}

// Collect implements the prometheus.Collector interface.
func (n *FSCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(n.Collectors))
	for name, c := range n.Collectors {
		go func(name string, c Collector) {
			defer wg.Done()
			n.execute(name, c, ch)
		}(name, c)
	}
	wg.Wait()
}

func (n *FSCollector) execute(name string, c Collector, ch chan<- prometheus.Metric) {
	begin := time.Now()
	err := c.Update(ch)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		if IsNoDataError(err) {
			n.logger.Debug("collector returned no data", zap.String("name", name),
				zap.Float64("duration_seconds", duration.Seconds()), zap.Error(err))
		} else {
			n.logger.Error("collector failed", zap.String("name", name), zap.Float64("duration_seconds", duration.Seconds()), zap.Error(err))
		}
		success = 0
	} else {
		n.logger.Debug("collector succeeded", zap.String("name", name), zap.Float64("duration_seconds", duration.Seconds()))
		success = 1
	}
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}

type typedDesc struct {
	desc      *prometheus.Desc
	valueType prometheus.ValueType
}

func (d *typedDesc) mustNewConstMetric(value float64, labels ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(d.desc, d.valueType, value, labels...)
}

// ErrNoData indicates the collector found no data to collect, but had no other error.
var ErrNoData = errors.New("collector returned no data")

// IsNoDataError defines the error of no data to collect
func IsNoDataError(err error) bool {
	return err == ErrNoData
}

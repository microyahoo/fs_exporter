package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/microyahoo/fs_exporter/collector"
	"github.com/microyahoo/fs_exporter/pkg/logutil"
)

var (
	cfgFile string
)

type fsExporterOptions struct {
	maxRequests   int64
	logConfig     *logutil.LogConfig
	listenAddress string
	metricsPath   string
}

func newFSExporterOptions() *fsExporterOptions {
	return &fsExporterOptions{
		maxRequests: 64,
		logConfig:   logutil.NewConfig(),
	}
}

// NewFSExporterCommand returns a fs exporter command, it is a root command.
func NewFSExporterCommand() *cobra.Command {
	o := newFSExporterOptions()
	cmds := &cobra.Command{
		Use:   "fs-exporter",
		Short: "An exporter for file systems",

		PersistentPreRunE: func(*cobra.Command, []string) error {
			return initProfiling()
		},
		PersistentPostRunE: func(*cobra.Command, []string) error {
			if err := flushProfiling(); err != nil {
				return err
			}
			return nil
		},

		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(o.Run())
		},
	}
	flags := cmds.PersistentFlags()
	addProfilingFlags(flags)

	cmds.Flags().Int64Var(&o.maxRequests, "web.max-requests", 40, "Maximum number of parallel scrape requests. Use 0 to disable.")
	cmds.Flags().StringVar(&o.logConfig.LogLevel, "log.level", "info", "log level")
	cmds.Flags().StringVar(&o.listenAddress, "web.listen-address", ":9097", "Address to listen on for telemetry")
	cmds.Flags().StringVar(&o.metricsPath, "web.telemetry-path", "/metrics", "Path under which to expose metrics")
	cmds.Flags().StringSliceVar(&o.logConfig.LogOutputs, "log.outputs", []string{"stderr"},
		"log outputs is a list of URLs or file paths to write logging output to.(default|stdout|stderr|file paths)")

	cmds.AddCommand(versionCmd)

	return cmds
}

func (o *fsExporterOptions) Run() error {
	cobra.CheckErr(o.logConfig.Validate())
	logger := o.logConfig.GetLogger()
	logger.Info("Starting fs exporter")
	logger.Debug("fs exporter options", zap.String("listen-address", o.listenAddress),
		zap.String("metric-path", o.metricsPath), zap.Int64("max-requests", o.maxRequests))
	logger.Debug("fs exporter logger options", zap.String("log-level", o.logConfig.LogLevel),
		zap.Any("log-outputs", o.logConfig.LogOutputs))

	http.Handle(o.metricsPath, newHandler(o.maxRequests, logger))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>File system Exporter</title></head>
			<body>
			<h1>File system Exporter</h1>
			<p><a href="` + o.metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})
	logger.Info("Listening on address", zap.String("listen-address", o.listenAddress))
	if err := http.ListenAndServe(o.listenAddress, nil); err != nil {
		logger.Error("error", zap.Error(err))
	}

	return nil
}

func newHandler(maxRequests int64, logger *zap.Logger) *handler {
	return &handler{
		// exporterMetricsRegistry: prometheus.NewRegistry(),
		maxRequests: maxRequests,
		logger:      logger,
	}
}

type handler struct {
	// unfilteredHandler http.Handler

	// exporterMetricsRegistry is a separate registry for the metrics about
	// the exporter itself.
	// exporterMetricsRegistry *prometheus.Registry

	// includeExporterMetrics  bool

	maxRequests int64
	logger      *zap.Logger
}

// ServeHTTP implements http.Handler.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, err := h.innerHandler()
	if err != nil {
		h.logger.Error("Couldn't create metrics handler:", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Couldn't create metrics handler: %s", err)))
		return
	}
	handler.ServeHTTP(w, r)
}

func (h *handler) innerHandler() (http.Handler, error) {
	fsc, err := collector.NewFSCollector(h.logger)
	if err != nil {
		return nil, fmt.Errorf("Failed to create collector: %s", err)
	}

	for n := range fsc.Collectors {
		h.logger.Info("Collector", zap.String("collector", n))
	}

	rgst := prometheus.NewRegistry()
	if err := rgst.Register(fsc); err != nil {
		h.logger.Error("Couldn't register collector:", zap.Error(err))
		return nil, err
	}
	handler := promhttp.HandlerFor(
		prometheus.Gatherers{rgst},
		promhttp.HandlerOpts{
			// ErrorLog:            stdlog.New(log.NewStdlibAdapter(level.Error(h.logger)), "", 0),
			ErrorHandling:       promhttp.ContinueOnError,
			MaxRequestsInFlight: int(h.maxRequests),
			// Registry:            h.exporterMetricsRegistry,
		},
	)
	return handler, nil
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".fs_exporter" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".fs_exporter")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

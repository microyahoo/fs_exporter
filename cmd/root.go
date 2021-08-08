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
	// rootCmd represents the base command when called without any subcommands
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
	cmds.Flags().StringVar(&o.logConfig.LogLevel, "log-level", "info", "log level")
	cmds.Flags().StringVar(&o.listenAddress, "web.listen-address", ":9097", "Address to listen on for telemetry")
	cmds.Flags().StringVar(&o.metricsPath, "web.telemetry-path", "/metrics", "Path under which to expose metrics")
	cmds.Flags().StringSliceVar(&o.logConfig.LogOutputs, "log-outputs", []string{"stderr"}, "log outputs is a list of URLs or file paths to write logging output to.(default|stdout|stderr|file paths)")

	cobra.CheckErr(o.logConfig.Validate())

	cmds.AddCommand(versionCmd)

	return cmds
}

func (o *fsExporterOptions) Run() error {
	logger := o.logConfig.GetLogger()
	logger.Info("Starting fs exporter")
	fmt.Printf("%#v\n", o)
	fmt.Printf("%#v\n", o.logConfig)

	http.Handle(o.metricsPath, newHandler(o.maxRequests, logger))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>File Exporters</title></head>
			<body>
			<h1>File Exporters</h1>
			<p><a href="` + o.metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})
	logger.Info("Listening on address", zap.String("", o.listenAddress))
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
	fsc := collector.NewFSCollector(h.logger)

	rgst := prometheus.NewRegistry()
	// rgst.MustRegister(version.NewCollector("node_exporter"))
	if err := rgst.Register(fsc); err != nil {
		h.logger.Error("Couldn't create filtered metrics handler:", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Couldn't create filtered metrics handler: %s", err)))
		return
	}
	handler := promhttp.HandlerFor(
		prometheus.Gatherers{h.exporterMetricsRegistry, rgst},
		promhttp.HandlerOpts{
			// ErrorLog:            stdlog.New(log.NewStdlibAdapter(level.Error(h.logger)), "", 0),
			ErrorHandling:       promhttp.ContinueOnError,
			MaxRequestsInFlight: int(h.maxRequests),
			Registry:            h.exporterMetricsRegistry,
		},
	)
	handler.ServeHTTP(w, r)
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

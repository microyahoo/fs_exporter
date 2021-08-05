package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/microyahoo/fs_exporter/cmd"
	"github.com/microyahoo/fs_exporter/pkg"
)

type ExporterCli struct {
	config *pkg.Config
}

type handler struct {
	unfilteredHandler http.Handler
	// exporterMetricsRegistry is a separate registry for the metrics about
	// the exporter itself.
	exporterMetricsRegistry *prometheus.Registry
	includeExporterMetrics  bool
	maxRequests             int
	logger                  log.Logger
}

func newCommand() (*cobra.Command, error) {
	// opts := newCommandOptions(pkg.NewConfig())

	cmd := &cobra.Command{
		Use:   "",
		Short: "",
	}
	return cmd, nil
}

func main() {
	var metricsPath *string
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to create a loggger: %s", err)
	}
	logger.Info("Start file exporters")

	// cmd, err := newCommand()
	// if err != nil {
	// 	logger.Fatal("command", zap.Any("error", err))
	// }
	// cmd.SetOut(os.Stdout)
	cmd.Execute()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>File Exporters</title></head>
			<body>
			<h1>File Exporters</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	// closeC := pkg.NewCloseNotifier()
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt)
	// go func() {
	// 	for sig := range c {
	// 		logger.Warn("fs expoerter received signal: ", zap.Any("Signal", sig))
	// 		if os.Interrupt == sig {
	// 			closeC.Close()
	// 			os.Exit(1)
	// 		}
	// 	}
	// }()

	// <-closeC.CloseNotify()
}

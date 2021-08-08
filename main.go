package main

import (
	"math/rand"
	_ "net/http/pprof"
	"time"

	"github.com/spf13/cobra"

	"github.com/microyahoo/fs_exporter/cmd"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	rootCmd := cmd.NewFSExporterCommand()
	cobra.CheckErr(rootCmd.Execute())

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

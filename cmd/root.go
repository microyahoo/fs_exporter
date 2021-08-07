package cmd

import (
	"fmt"
	"os"

	"github.com/microyahoo/fs_exporter/pkg/logutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
)

type fsExporterOptions struct {
	maxRequests int64
	logConfig   *logutil.LogConfig
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

	cmds.Flags().Int64Var(&o.maxRequests, "max-requests", 64, "Source directory to read from")
	cmds.Flags().StringVar(&o.logConfig.LogLevel, "log-level", "info", "log level")
	cmds.Flags().StringSliceVar(&o.logConfig.LogOutputs, "log-outputs", []string{"stderr"}, "log outputs is a list of URLs or file paths to write logging output to.(default|stdout|stderr|file paths)")

	cobra.CheckErr(o.logConfig.Validate())

	cmds.AddCommand(versionCmd)

	return cmds
}

func (o *fsExporterOptions) Run() error {
	fmt.Println("Run")
	fmt.Printf("%#v\n", o)
	fmt.Printf("%#v\n", o.logConfig)
	return nil
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

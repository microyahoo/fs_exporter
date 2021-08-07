package cmd

import (
	"fmt"
	"os"

	"github.com/microyahoo/fs_exporter/pkg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
)

type fsExporterOptions struct {
	maxRequests int64
	logConfig   *pkg.LogConfig
}

func newFSExporterOptions() *fsExporterOptions {
	return &fsExporterOptions{
		maxRequests: 64,
		logConfig:   pkg.NewConfig(),
	}
}

// NewFSExporterCommand returns a fs exporter command, it is a root command.
func NewFSExporterCommand() *cobra.Command {
	o := newFSExporterOptions()
	// rootCmd represents the base command when called without any subcommands
	cmds := &cobra.Command{
		Use:   "fs-exporter",
		Short: "An exporter for file systems",
		// Long: `A longer description that spans multiple lines and likely contains
		// examples and usage of using your application. For example:

		// Cobra is a CLI library for Go that empowers applications.
		// This application is a tool to generate the needed files
		// to quickly create a Cobra application.`,

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

	// cmds.AddCommand(zfsCmd)
	// cmds.AddCommand(glusterfsCmd)
	cmds.AddCommand(versionCmd)

	// opts := newCommandOptions(pkg.NewConfig())

	return cmds
}

func (o *fsExporterOptions) Run() error {
	fmt.Println("Run")
	return nil
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.fs_exporter.yaml)")
	// rootCmd.PersistentFlags().StringP("author", "a", "YOUR NAME", "author name for copyright attribution")
	// rootCmd.PersistentFlags().StringVarP(&userLicense, "license", "l", "", "name of license for the project")
	// rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")
	// rootCmd.Flags().StringVarP(&source, "source", "s", "", "Source directory to read from")
	// rootCmd.MarkFlagRequired("source")
	// rootCmd.MarkPersistentFlagRequired("region")

	// viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	// viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
	// viper.SetDefault("author", "Liang Zheng<zhengliang0901@gmail.com>")
	// viper.SetDefault("license", "apache")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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

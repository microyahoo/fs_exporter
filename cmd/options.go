package cmd

import "github.com/microyahoo/fs_exporter/pkg"

type commandOptions struct {
	Debug    bool
	LogLevel string
	config   *pkg.Config
}

func newCommandOptions(config *pkg.Config) *commandOptions {
	return &commandOptions{
		// Debug:    false,
		// LogLevel: "debug",
		config: config,
	}
}

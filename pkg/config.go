package pkg

import (
	"fmt"
	"sync"

	"github.com/microyahoo/fs_exporter/pkg/logutil"
	"go.uber.org/zap"
)

const (
	// DefaultName = "fs-exporter"

	DefaultLogOutput = "default"
	StdErrLogOutput  = "stderr"
	StdOutLogOutput  = "stdout"
)

// LogConfig holds the arguments for configuring logger
type LogConfig struct {
	// Name string `json:"name"`

	// LogLevel configures log level. Only supports debug, info, warn, error, panic, or fatal. Default 'info'.
	LogLevel string `json:"log-level"`
	// LogOutputs is either:
	//  - "default" as os.Stderr,
	//  - "stderr" as os.Stderr,
	//  - "stdout" as os.Stdout,
	//  - file path to append server logs to.
	// It can be multiple when "Logger" is zap.
	LogOutputs []string `json:"log-outputs"`

	// ZapLoggerBuilder is used to build the zap logger.
	ZapLoggerBuilder func(*LogConfig) error

	// logger logs server-side operations. The default is nil,
	// and "setupLogging" must be called before starting server.
	// Do not set logger directly.
	loggerMu *sync.RWMutex
	logger   *zap.Logger

	// loggerConfig is server logger configuration for Raft logger.
	// Must be either: "loggerConfig != nil" or "loggerCore != nil && loggerWriteSyncer != nil".
	loggerConfig *zap.Config
	// loggerCore is "zapcore.Core" for raft logger.
	// Must be either: "loggerConfig != nil" or "loggerCore != nil && loggerWriteSyncer != nil".
	// loggerCore        zapcore.Core
	// loggerWriteSyncer zapcore.WriteSyncer
}

// NewLogConfig creates a new Config populated with default values.
func NewConfig() *LogConfig {
	cfg := &LogConfig{
		// Name: DefaultName,

		loggerMu:   new(sync.RWMutex),
		logger:     nil,
		LogOutputs: []string{DefaultLogOutput},
		LogLevel:   logutil.DefaultLogLevel,
	}
	return cfg
}

// Validate ensures that '*LogConfig' fields are properly configured.
func (cfg *LogConfig) Validate() error {
	if err := cfg.setupLogging(); err != nil {
		return err
	}
	return nil
}

// GetLogger returns the logger.
func (cfg LogConfig) GetLogger() *zap.Logger {
	cfg.loggerMu.RLock()
	l := cfg.logger
	cfg.loggerMu.RUnlock()
	return l
}

// setupLogging initializes logging.
// Must be called after flag parsing or finishing configuring LogConfig.
func (cfg *LogConfig) setupLogging() error {
	if len(cfg.LogOutputs) == 0 {
		cfg.LogOutputs = []string{DefaultLogOutput}
	}
	if len(cfg.LogOutputs) > 1 {
		for _, v := range cfg.LogOutputs {
			if v == DefaultLogOutput {
				return fmt.Errorf("multi logoutput for %q is not supported yet", DefaultLogOutput)
			}
		}
	}

	outputPaths, errOutputPaths := make([]string, 0), make([]string, 0)
	for _, v := range cfg.LogOutputs {
		switch v {
		case DefaultLogOutput:
			outputPaths = append(outputPaths, StdErrLogOutput)
			errOutputPaths = append(errOutputPaths, StdErrLogOutput)

		case StdErrLogOutput:
			outputPaths = append(outputPaths, StdErrLogOutput)
			errOutputPaths = append(errOutputPaths, StdErrLogOutput)

		case StdOutLogOutput:
			outputPaths = append(outputPaths, StdOutLogOutput)
			errOutputPaths = append(errOutputPaths, StdOutLogOutput)

		default:
			outputPaths = append(outputPaths, v)
			errOutputPaths = append(errOutputPaths, v)
		}
	}

	copied := logutil.DefaultZapLoggerConfig
	copied.OutputPaths = outputPaths
	copied.ErrorOutputPaths = errOutputPaths
	// copied = logutil.MergeOutputPaths(copied)
	copied.Level = zap.NewAtomicLevelAt(logutil.ConvertToZapLevel(cfg.LogLevel))
	if cfg.ZapLoggerBuilder == nil {
		cfg.ZapLoggerBuilder = func(c *LogConfig) error {
			var err error
			c.logger, err = copied.Build()
			if err != nil {
				return err
			}
			zap.ReplaceGlobals(c.logger)
			c.loggerMu.Lock()
			defer c.loggerMu.Unlock()
			c.loggerConfig = &copied
			// c.loggerCore = nil
			// c.loggerWriteSyncer = nil
			return nil
		}
	}
	err := cfg.ZapLoggerBuilder(cfg)
	if err != nil {
		return err
	}
	return nil
}

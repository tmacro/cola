package main

import (
	"encoding/json"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
	"github.com/tmacro/cola/pkg/config"
)

var CLI struct {
	LogLevel  string   `help:"Set the log level." enum:"trace,debug,info,warn,error" default:"debug"`
	LogFormat string   `enum:"json,text" default:"text" help:"Set the log format. (json, text)"`
	Config    []string `short:"c" help:"Path to the configuration file." default:"appliance.hcl" type:"path"`
	Base      []string `short:"b" help:"Use this config as a base to extend from." type:"path"`
	Output    string   `short:"o" help:"Output file."`
}

func main() {
	kong.Parse(&CLI,
		kong.Name("cola-ignition"),
		kong.Description("Build Ignition configs for a COLA appliance"),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	logger := createLogger(CLI.LogLevel, CLI.LogFormat)

	cfg, err := loadConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	logger.Trace().Interface("config", cfg).Msg("Configuration loaded")

	ignCfg, err := buildIgnitionConfig(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to build Ignition config")
	}

	logger.Trace().Interface("ignition", ignCfg).Msg("Ignition config built")

	ignJson, err := json.MarshalIndent(&ignCfg, "", "  ")
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to marshal Ignition config")
	}

	var output io.Writer
	if CLI.Output != "" {
		output, err = os.Create(CLI.Output)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to open output file")
		}
	} else {
		output = os.Stdout
	}

	_, err = output.Write(ignJson)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to write Ignition config")
	}
}

func createLogger(level, format string) zerolog.Logger {
	var lvl zerolog.Level
	switch level {
	case "trace":
		lvl = zerolog.TraceLevel
	case "debug":
		lvl = zerolog.DebugLevel
	case "info":
		lvl = zerolog.InfoLevel
	case "warn":
		lvl = zerolog.WarnLevel
	case "error":
		lvl = zerolog.ErrorLevel
	default:
		panic("invalid log level: " + level)
	}

	var writer io.Writer
	switch format {
	case "json":
		writer = os.Stdout
	case "text":
		writer = zerolog.ConsoleWriter{Out: os.Stdout}
	default:
		panic("invalid log format: " + format)
	}
	return zerolog.New(writer).Level(lvl).With().Timestamp().Logger()
}

func loadConfig() (*config.ApplianceConfig, error) {
	var baseCfg *config.ApplianceConfig
	var err error

	if len(CLI.Base) > 0 {
		baseCfg, err = config.ReadConfig(CLI.Base, false)
		if err != nil {
			return nil, err
		}
	}

	cfg, err := config.ReadConfig(CLI.Config, true)
	if err != nil {
		return nil, err
	}

	merged := config.MergeConfigs(baseCfg, cfg)

	err = config.ValidateConfig(merged)
	if err != nil {
		return nil, err
	}

	return merged, nil
}

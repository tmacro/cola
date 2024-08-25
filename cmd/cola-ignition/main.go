package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/rs/zerolog"
)

var CLI struct {
	LogLevel  string `help:"Set the log level." enum:"trace,debug,info,warn,error" default:"debug"`
	LogFormat string `enum:"json,text" default:"text" help:"Set the log format. (json, text)"`
	Config    string `short:"c" help:"Path to the configuration file." type:"existingfile" default:"appliance.hcl"`
	Base      string `short:"b" help:"Use this config as a base to extend from." type:"existingfile"`
	Output    string `short:"o" help:"Output file."`
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

	ctx := context.Background()

	ctx = logger.WithContext(ctx)

	configPath, err := filepath.Abs(CLI.Config)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to get absolute path to configuration file")
	}

	var config ApplianceConfig
	err = hclsimple.DecodeFile(configPath, nil, &config)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	if CLI.Base != "" {
		var baseCfg ApplianceConfig
		err := hclsimple.DecodeFile(CLI.Base, nil, &baseCfg)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to load base configuration")
		}
		logger.Info().Interface("base_config", baseCfg).Msg("Base configuration loaded")
		config = mergeConfigs(baseCfg, config)
	}

	logger.Info().Interface("config", config).Msg("Configuration loaded")
	ignCfg, err := buildIgnitionConfig(configPath, config)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to build Ignition config")
	}

	logger.Info().Interface("ignition", ignCfg).Msg("Ignition config built")
	cfg, err := json.MarshalIndent(&ignCfg, "", "  ")
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

	_, err = output.Write(cfg)
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

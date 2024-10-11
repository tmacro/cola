package main

import (
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
	"github.com/tmacro/cola/pkg/config"
	"github.com/tmacro/cola/pkg/ignition"
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

	cfg, err := config.ReadConfig(CLI.Config, CLI.Base, true)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to read configuration")
	}

	err = config.ValidateConfig(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to validate configuration")
	}

	logger.Trace().Interface("config", cfg).Msg("Configuration loaded")

	opts := []ignition.GeneratorOpt{}
	if CLI.BundledExtensions {
		opts = append(opts, ignition.WithBundledExtensions())
	}

	ignJson, err := ignition.Generate(cfg, opts...)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to generate Ignition config")
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

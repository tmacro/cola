package main

import (
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
	Image     string   `short:"i" help:"Path to the Flatcar Linux image." type:"existingpath"`
	Ignition  string   `short:"g" help:"Path to the Ignition config." type:"existingpath"`
	Output    string   `short:"o" help:"Output file."`
}

func main() {
	kong.Parse(&CLI,
		kong.Name("cola-bundle"),
		kong.Description("Bundle sysexts and an Ignition config with a Flatcar Linux image"),
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

	workdir, err := os.MkdirTemp("", "cola-bundle")
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create temporary directory")
	}

	// defer os.RemoveAll(workdir)

	err = bundle(cfg, workdir, logger)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to bundle")
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

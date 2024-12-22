package main

import (
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
)

var CLI struct {
	LogLevel  string      `help:"Set the log level." enum:"trace,debug,info,warn,error" default:"debug"`
	LogFormat string      `enum:"json,text" default:"text" help:"Set the log format. (json, text)"`
	Generate  GenerateCmd `help:"Generate an Ignition config." cmd:""`
	Bundle    BundleCmd   `help:"Bundle sysexts and an Ignition config with a Flatcar Linux image." cmd:""`
}

func createLogger(level, format string) *zerolog.Logger {
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
	logger := zerolog.New(writer).Level(lvl).With().Timestamp().Logger()
	return &logger
}

func main() {
	cmd := kong.Parse(&CLI,
		kong.Name("cola"),
		// kong.Description("Build Ignition configs for a COLA appliance"),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	logger := createLogger(CLI.LogLevel, CLI.LogFormat)

	err := cmd.Run(logger)
	cmd.FatalIfErrorf(err)
}

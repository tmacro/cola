package main

import (
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var CLI struct {
	LogLevel  string      `help:"Set the log level." enum:"trace,debug,info,warn,error" default:"info"`
	LogFormat string      `enum:"json,text" default:"text" help:"Set the log format. (json, text)"`
	Generate  GenerateCmd `help:"Generate an Ignition config." cmd:""`
	Bundle    BundleCmd   `help:"Bundle sysexts and an Ignition config with a Flatcar Linux image." cmd:""`
}

func main() {
	cmd := kong.Parse(&CLI,
		kong.Name("cola"),
		// kong.Description("Build Ignition configs for a COLA appliance"),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	var logLevel zerolog.Level
	switch CLI.LogLevel {
	case "trace":
		logLevel = zerolog.TraceLevel
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	default:
		panic("invalid log level: " + CLI.LogLevel)
	}

	var writer io.Writer
	switch CLI.LogFormat {
	case "json":
		writer = os.Stdout
	case "text":
		writer = zerolog.ConsoleWriter{Out: os.Stdout}
	default:
		panic("invalid log format: " + CLI.LogFormat)
	}

	log.Logger = zerolog.New(writer).Level(logLevel).With().Timestamp().Logger()

	err := cmd.Run()
	cmd.FatalIfErrorf(err)
}

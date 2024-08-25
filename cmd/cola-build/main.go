package main

import (
	"context"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
)

var CLI struct {
	LogLevel  string `help:"Set the log level." enum:"trace,debug,info,warn,error" default:"debug"`
	LogFormat string `enum:"json,text" default:"text" help:"Set the log format. (json, text)"`
	Name      string `short:"n" help:"Name of the extension." required:""`
	Version   string `short:"v" help:"Version of the extension." required:""`
	Arch      string `short:"a" help:"Architecture of the extension." enum:"amd64,arm64"`
	Source    string `help:"Path to the extension source." type:"existingdir" required:""`
	Output    string `short:"o" help:"Output path for the extension." type:"existingdir" default:"."`
}

func main() {
	kong.Parse(&CLI,
		kong.Name("cola-build"),
		kong.Description("Build COLA extensions"),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	logger := createLogger(CLI.LogLevel, CLI.LogFormat)

	ctx := context.Background()

	ctx = logger.WithContext(ctx)

	err := validateExtension(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("invalid extension")
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

func validateExtension(ctx context.Context) error {
	ok, err := hasFile(CLI.Source + "/extension.yaml")
	if err != nil {
		return err
	}

	if !ok {
		return os.ErrNotExist
	}

	return nil
}

func hasFile(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return !info.IsDir(), nil
}

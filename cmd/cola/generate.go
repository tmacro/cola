package main

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/tmacro/cola/pkg/config"
	"github.com/tmacro/cola/pkg/ignition"
)

type GenerateCmd struct {
	Config            []string `short:"c" help:"Path to the configuration file or directory." type:"path"`
	Output            string   `short:"o" help:"Output file."`
	BundledExtensions bool     `short:"b" help:"Assume extensions are will be bundled into the image."`
	ExtensionDir      string   `short:"e" help:"Directory containing sysexts." type:"existingdir" optional:""`
}

func (cmd *GenerateCmd) Run(logger *zerolog.Logger) error {
	if len(cmd.Config) == 0 {
		logger.Fatal().Msg("No configuration file specified")
	}

	cfg, err := config.ReadConfig(cmd.Config, []string{}, false)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to read configuration")
	}

	err = config.ValidateConfig(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to validate configuration")
	}

	logger.Trace().Interface("config", cfg).Msg("Configuration loaded")

	opts := []ignition.GeneratorOpt{}
	if cmd.BundledExtensions {
		opts = append(opts, ignition.WithBundledExtensions())
	}

	if cmd.ExtensionDir != "" {
		opts = append(opts, ignition.WithExtensionDir(cmd.ExtensionDir))
	}

	ignJson, err := ignition.Generate(cfg, opts...)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to generate Ignition config")
	}

	var output io.Writer
	if cmd.Output != "" {
		output, err = os.Create(cmd.Output)
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

	return nil
}

package main

import (
	"io"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/tmacro/cola/pkg/config"
	"github.com/tmacro/cola/pkg/ignition"
)

type GenerateCmd struct {
	Config            []string `short:"c" help:"Path to the configuration file or directory." type:"path"`
	VarFile           []string `short:"v" help:"Path to the files containing variable values." type:"path"`
	Output            string   `short:"o" help:"Output file."`
	BundledExtensions bool     `short:"b" help:"Assume extensions are will be bundled into the image."`
	ExtensionDir      string   `short:"e" help:"Directory containing sysexts." type:"existingdir" optional:""`
}

func (cmd *GenerateCmd) Run() error {
	if len(cmd.Config) == 0 {
		log.Fatal().Msg("No configuration file specified")
	}

	cfg, err := config.ReadConfig(cmd.Config, cmd.VarFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read configuration")
	}

	err = config.ValidateConfig(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to validate configuration")
	}

	log.Trace().Interface("config", cfg).Msg("Configuration loaded")

	opts := []ignition.GeneratorOpt{}
	if cmd.BundledExtensions {
		workdir, err := os.MkdirTemp("", "cola-generate")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create temporary directory")
		}

		defer os.RemoveAll(workdir)

		err = fetchExtensionTransferConfigs(workdir, cmd.ExtensionDir, cfg)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to fetch extension transfer configs")
		}

		opts = append(opts, ignition.WithBundledExtensions(), ignition.WithExtensionDir(workdir))
	}

	ignJson, err := ignition.Generate(cfg, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to generate Ignition config")
	}

	var output io.Writer
	if cmd.Output != "" {
		output, err = os.Create(cmd.Output)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to open output file")
		}
	} else {
		output = os.Stdout
	}

	_, err = output.Write(ignJson)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to write Ignition config")
	}

	return nil
}

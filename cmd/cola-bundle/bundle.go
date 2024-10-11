package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/tmacro/cola/pkg/config"
	"github.com/tmacro/cola/pkg/ignition"
)

func bundle(cfg *config.ApplianceConfig, workdir string, logger zerolog.Logger) error {
	err := fetchExtensions(cfg, workdir, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to fetch extensions")
	}

	imagePath := filepath.Join(workdir, "flatcar_production_image.bin")
	err = exec.Command("cp", CLI.Image, imagePath).Run()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to copy image")
	}

	cleanupMounts, err := mountImage(imagePath, filepath.Join(workdir, "mnt"))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to mount image")
	}

	err = installSysExts(cfg, workdir, logger)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to install extensions")
		cleanupMounts()
		return err
	}

	err = installIgnition(workdir, CLI.Ignition, logger)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to install Ignition config")
		cleanupMounts()
		return err
	}

	cleanupMounts()

	err = exec.Command("mv", imagePath, CLI.Output).Run()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to move image to output")
	}

	return nil
}

func installSysExts(cfg *config.ApplianceConfig, workdir string, logger zerolog.Logger) error {
	installPath := filepath.Join(workdir, "mnt", "flatcar-root", "opt", "extensions")
	symlinkPath := filepath.Join(workdir, "mnt", "flatcar-root", "etc", "extensions")
	for _, ext := range cfg.Extensions {
		logger.Info().Str("name", ext.Name).Str("version", ext.Version).Str("arch", ext.Arch).Msg("Installing extension")
		err := os.MkdirAll(filepath.Join(installPath, ext.Name), 0o755)
		if err != nil {
			return fmt.Errorf("Failed to create extension directory: %w", err)
		}

		err = os.MkdirAll(filepath.Join(symlinkPath), 0o755)
		if err != nil {
			return fmt.Errorf("Failed to create symlink directory: %w", err)
		}

		// Copy extension files to the install path
		extFilename := ignition.FormatExtensionName(ext.Name, ext.Version, ext.Arch) + ".raw"
		extPath := filepath.Join(workdir, extFilename)
		extInstallPath := filepath.Join(installPath, ext.Name, extFilename)

		err = exec.Command("cp", extPath, extInstallPath).Run()
		if err != nil {
			return fmt.Errorf("Failed to copy extension file: %w", err)
		}

		err = exec.Command("ln", "-s", filepath.Join("/opt/extensions", ext.Name, extFilename), filepath.Join(symlinkPath, ext.Name+".raw")).Run()
		if err != nil {
			return fmt.Errorf("Failed to create symlink: %w", err)
		}

		transferCfg := filepath.Join(workdir, ext.Name+".conf")
		transferCfgInstallPath := filepath.Join("/etc", "sysupdate."+ext.Name+".d")

		err = os.MkdirAll(transferCfgInstallPath, 0o755)
		if err != nil {
			return fmt.Errorf("Failed to create transfer config directory: %w", err)
		}

		err = exec.Command("cp", transferCfg, transferCfgInstallPath).Run()
		if err != nil {
			return fmt.Errorf("Failed to copy transfer config: %w", err)
		}
	}

	return nil
}

func installIgnition(workdir, ignitionPath string, logger zerolog.Logger) error {
	err := exec.Command("cp", ignitionPath, filepath.Join(workdir, "mnt", "flatcar-oem", "config.ign")).Run()
	if err != nil {
		return fmt.Errorf("Failed to copy Ignition config: %w", err)
	}

	return nil
}

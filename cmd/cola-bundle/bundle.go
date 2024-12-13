package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/tmacro/cola/internal/templates"
	"github.com/tmacro/cola/pkg/config"
	"github.com/tmacro/cola/pkg/ignition"
)

func bundle(cfg *config.ApplianceConfig, workdir string, logger *zerolog.Logger) error {
	imagePath := filepath.Join(workdir, "flatcar_production_image.bin")
	err := copyFile(CLI.Image, imagePath, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to copy image")
	}

	err = fetchExtensions(cfg, workdir, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to fetch extensions")
	}

	cleanupMounts, err := mountImage(imagePath, filepath.Join(workdir, "mnt"), logger)
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

	err = cleanupMounts()
	if err != nil {
		logger.Fatal().Err(err).Msg("Error cleaning up mounts")
	}

	err = exec.Command("mv", imagePath, CLI.Output).Run()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to move image to output")
	}

	logger.Info().Str("output", CLI.Output).Msg("Image bundled successfully")

	return nil
}

func installSysExts(cfg *config.ApplianceConfig, workdir string, logger *zerolog.Logger) error {
	installPath := filepath.Join(workdir, "mnt", "root", "opt", "extensions")
	symlinkPath := filepath.Join(workdir, "mnt", "root", "etc", "extensions")

	transferCfgs := make([]templates.Tmpfile, 0)

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
		transferCfgImagePath := filepath.Join("/usr/lib/cola/etc", "sysupdate."+ext.Name+".d")

		err = os.MkdirAll(transferCfgInstallPath, 0o755)
		if err != nil {
			return fmt.Errorf("Failed to create transfer config directory: %w", err)
		}

		err = os.MkdirAll(transferCfgImagePath, 0o755)
		if err != nil {
			return fmt.Errorf("Failed to create transfer config directory: %w", err)
		}

		err = exec.Command("cp", transferCfg, transferCfgImagePath).Run()
		if err != nil {
			return fmt.Errorf("Failed to copy transfer config: %w", err)
		}

		transferCfgs = append(transferCfgs,
			templates.Tmpfile{
				Mode:   "C",
				Target: transferCfgInstallPath,
				Source: transferCfgImagePath,
			},
		)
	}

	tmpfileCfg, err := templates.SystemdTmpfileConfig(transferCfgs...)
	if err != nil {
		return fmt.Errorf("Failed to generate systemd-tmpfiles config: %w", err)
	}

	err = os.MkdirAll("/usr/lib/tmpfiles.d/", 0o755)
	if err != nil {
		return fmt.Errorf("Failed to create tmpfiles.d directory: %w", err)
	}

	f, err := os.Create("/usr/lib/tmpfiles.d/cola-extensions.conf")
	if err != nil {
		return fmt.Errorf("Failed to copy tmpfiles config: %w", err)
	}

	defer f.Close()

	_, err = f.WriteString(tmpfileCfg)
	if err != nil {
		return fmt.Errorf("Failed to write tmpfiles config: %w", err)
	}

	return nil
}

func installIgnition(workdir, ignitionPath string, logger *zerolog.Logger) error {
	logger.Info().Str("path", ignitionPath).Msg("Installing Ignition config")
	err := exec.Command("cp", ignitionPath, filepath.Join(workdir, "mnt", "oem", "config.ign")).Run()
	if err != nil {
		return fmt.Errorf("Failed to copy Ignition config: %w", err)
	}

	return nil
}

func copyFile(src, dest string, logger *zerolog.Logger) error {
	fin, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("Failed to open source file: %w", err)
	}

	defer fin.Close()

	fout, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("Failed to create destination file: %w", err)
	}

	defer fout.Close()

	_, err = io.Copy(fout, fin)
	if err != nil {
		return fmt.Errorf("Failed to copy file: %w", err)
	}

	return nil
}

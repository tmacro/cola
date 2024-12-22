package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/moby/sys/mount"
	"github.com/rs/zerolog"
	"github.com/tmacro/cola/internal/templates"
	"github.com/tmacro/cola/pkg/config"
	"github.com/tmacro/cola/pkg/download"
	"github.com/tmacro/cola/pkg/ignition"
	"github.com/tmacro/cola/pkg/losetup"
)

type BundleCmd struct {
	Config            []string `short:"c" help:"Path to the configuration file or directory." type:"path"`
	Base              []string `help:"Use this config as a base to extend from." type:"path"`
	Image             string   `short:"f" help:"Path to the Flatcar Linux image." type:"existingpath" required:""`
	GenIgnition       bool     `short:"g" help:"Generate the Ignition config. (cannot be used with --ignition)"`
	Ignition          string   `short:"i" help:"Path to the Ignition config." type:"existingpath" optional:""`
	Output            string   `short:"o" help:"Output file."`
	BundledExtensions bool     `short:"b" help:"Bundle extensions into the image."`
	ExtensionDir      string   `short:"e" help:"Directory containing sysexts." type:"existingdir" optional:""`
}

func (cmd *BundleCmd) Run(logger *zerolog.Logger) error {
	cfg, err := config.ReadConfig(cmd.Config, []string{}, false)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to read configuration")
	}

	err = config.ValidateConfig(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to validate configuration")
	}

	workdir, err := os.MkdirTemp("", "cola-bundle")
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create temporary directory")
	}

	defer os.RemoveAll(workdir)

	imagePath := filepath.Join(workdir, "flatcar_production_image.bin")
	err = copyFile(cmd.Image, imagePath)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to copy image")
	}

	err = fetchExtensions(workdir, cmd.ExtensionDir, cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to fetch extensions")
	}

	cleanupMounts, err := mountImage(imagePath, filepath.Join(workdir, "mnt"), logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to mount image")
	}

	err = installSysExts(cfg, workdir, logger)
	if err != nil {
		logger.Error().Err(err).Msg("failed to install extensions")
		cleanupMounts()
		return err
	}

	err = installIgnition(workdir, cmd.Ignition, logger)
	if err != nil {
		logger.Error().Err(err).Msg("failed to install Ignition config")
		cleanupMounts()
		return err
	}

	err = cleanupMounts()
	if err != nil {
		logger.Fatal().Err(err).Msg("Error cleaning up mounts")
	}

	return nil
}

func copyFile(src, dest string) error {
	fin, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}

	defer fin.Close()

	fout, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}

	defer fout.Close()

	_, err = io.Copy(fout, fin)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func copyOrDownload(srcPath, srcURL, destPath string, logger *zerolog.Logger) error {
	if srcPath != "" {
		if fileExists(srcPath) {
			logger.Info().Str("source", srcPath).Str("destination", destPath).Msg("Using existing file")
			if err := copyFile(srcPath, destPath); err != nil {
				return fmt.Errorf("failed to copy file: %w", err)
			}
			return nil
		}
	}

	logger.Info().Str("url", srcURL).Str("destination", destPath).Msg("Downloading file")
	if err := download.Get(srcURL, destPath); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

func fetchExtensions(workdir, extdir string, cfg *config.ApplianceConfig, logger *zerolog.Logger) error {
	for _, ext := range cfg.Extensions {
		extFilename := ignition.FormatExtensionName(ext.Name, ext.Version, ext.Arch) + ".raw"
		extDestPath := filepath.Join(workdir, extFilename)
		transferCfgDestPath := filepath.Join(workdir, ext.Name+".conf")

		extPath := ""
		transferCfgPath := ""
		if extdir != "" {
			extPath = filepath.Join(extdir, extFilename)
			transferCfgPath = filepath.Join(extdir, ext.Name+".conf")
		}

		extUrl := ignition.FormatExtensionURL(ext.BakeryUrl, ext.Name, ext.Version, ext.Arch)
		if err := copyOrDownload(extPath, extUrl, extDestPath, logger); err != nil {
			return fmt.Errorf("failed to download extension: %w", err)
		}

		transferCfgUrl := ignition.FormatExtensionTransferConfigURL(ext.BakeryUrl, ext.Name)
		if err := copyOrDownload(transferCfgPath, transferCfgUrl, transferCfgDestPath, logger); err != nil {
			return fmt.Errorf("failed to download transfer config: %w", err)
		}
	}

	return nil
}

const (
	OEM_PARTITION  = "p6"
	ROOT_PARTITION = "p9"
)

func mountImage(image, mountpoint string, logger *zerolog.Logger) (func() error, error) {
	loopDev, err := losetup.SetupDevice(image)
	if err != nil {
		return nil, err
	}

	logger.Debug().Str("device", loopDev).Msg("Mounted image")

	oemMount := filepath.Join(mountpoint, "oem")
	rootMount := filepath.Join(mountpoint, "root")

	err = os.MkdirAll(oemMount, 0o755)
	if err != nil {
		losetup.DetachDevice(loopDev)
		return nil, fmt.Errorf("failed to create OEM mount directory: %w", err)
	}

	err = os.MkdirAll(rootMount, 0o755)
	if err != nil {
		losetup.DetachDevice(loopDev)
		return nil, fmt.Errorf("failed to create root mount directory: %w", err)
	}

	cleanupMounts := func() error {
		hasError := false

		logger.Debug().Str("mountpoint", oemMount).Msg("Unmounting OEM partition")
		err := mount.Unmount(oemMount)
		if err != nil {
			hasError = true
			logger.Error().Err(err).Msg("failed to unmount OEM partition")
		}

		logger.Debug().Str("mountpoint", rootMount).Msg("Unmounting root partition")
		err = mount.Unmount(rootMount)
		if err != nil {
			hasError = true
			logger.Error().Err(err).Msg("failed to unmount root partition")
		}

		err = losetup.DetachDevice(loopDev)
		if err != nil {
			hasError = true
			logger.Error().Err(err).Msg("failed to detach loop device")
		}

		if hasError {
			return fmt.Errorf("errors occurred during image unmount")
		}

		logger.Debug().Msg("Unmounted image")
		return nil
	}

	oemPart := loopDev + OEM_PARTITION
	rootPart := loopDev + ROOT_PARTITION

	logger.Debug().Str("partition", oemPart).Str("mountpoint", oemMount).Msg("Mounting OEM partition")

	err = exec.Command("mount", "-o", "loop", loopDev+OEM_PARTITION, oemMount).Run()
	if err != nil {
		cleanupMounts()
		return nil, fmt.Errorf("failed to mount OEM partition: %w", err)
	}

	logger.Debug().Str("partition", rootPart).Str("mountpoint", rootMount).Msg("Mounting root partition")

	err = exec.Command("mount", "-o", "loop", loopDev+ROOT_PARTITION, rootMount).Run()
	if err != nil {
		cleanupMounts()
		return nil, fmt.Errorf("failed to mount root partition: %w", err)
	}

	return cleanupMounts, nil
}

func installSysExts(cfg *config.ApplianceConfig, workdir string, logger *zerolog.Logger) error {
	installPath := filepath.Join(workdir, "mnt", "root", "opt", "extensions")
	symlinkPath := filepath.Join(workdir, "mnt", "root", "etc", "extensions")

	transferCfgs := make([]templates.Tmpfile, 0)

	for _, ext := range cfg.Extensions {
		logger.Info().Str("name", ext.Name).Str("version", ext.Version).Str("arch", ext.Arch).Msg("Installing extension")
		err := os.MkdirAll(filepath.Join(installPath, ext.Name), 0o755)
		if err != nil {
			return fmt.Errorf("failed to create extension directory: %w", err)
		}

		err = os.MkdirAll(filepath.Join(symlinkPath), 0o755)
		if err != nil {
			return fmt.Errorf("failed to create symlink directory: %w", err)
		}

		// Copy extension files to the install path
		extFilename := ignition.FormatExtensionName(ext.Name, ext.Version, ext.Arch) + ".raw"
		extPath := filepath.Join(workdir, extFilename)
		extInstallPath := filepath.Join(installPath, ext.Name, extFilename)

		err = exec.Command("cp", extPath, extInstallPath).Run()
		if err != nil {
			return fmt.Errorf("failed to copy extension file: %w", err)
		}

		err = exec.Command("ln", "-s", filepath.Join("/opt/extensions", ext.Name, extFilename), filepath.Join(symlinkPath, ext.Name+".raw")).Run()
		if err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}

		transferCfg := filepath.Join(workdir, ext.Name+".conf")
		transferCfgInstallPath := filepath.Join("/etc", "sysupdate."+ext.Name+".d")
		transferCfgImagePath := filepath.Join("/usr/lib/cola/etc", "sysupdate."+ext.Name+".d")

		err = os.MkdirAll(transferCfgInstallPath, 0o755)
		if err != nil {
			return fmt.Errorf("failed to create transfer config directory: %w", err)
		}

		err = os.MkdirAll(transferCfgImagePath, 0o755)
		if err != nil {
			return fmt.Errorf("failed to create transfer config directory: %w", err)
		}

		err = exec.Command("cp", transferCfg, transferCfgImagePath).Run()
		if err != nil {
			return fmt.Errorf("failed to copy transfer config: %w", err)
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
		return fmt.Errorf("failed to generate systemd-tmpfiles config: %w", err)
	}

	err = os.MkdirAll("/usr/lib/tmpfiles.d/", 0o755)
	if err != nil {
		return fmt.Errorf("failed to create tmpfiles.d directory: %w", err)
	}

	f, err := os.Create("/usr/lib/tmpfiles.d/cola-extensions.conf")
	if err != nil {
		return fmt.Errorf("failed to copy tmpfiles config: %w", err)
	}

	defer f.Close()

	_, err = f.WriteString(tmpfileCfg)
	if err != nil {
		return fmt.Errorf("failed to write tmpfiles config: %w", err)
	}

	return nil
}

func installIgnition(workdir, ignitionPath string, logger *zerolog.Logger) error {
	logger.Info().Str("path", ignitionPath).Msg("Installing Ignition config")
	err := exec.Command("cp", ignitionPath, filepath.Join(workdir, "mnt", "oem", "config.ign")).Run()
	if err != nil {
		return fmt.Errorf("failed to copy Ignition config: %w", err)
	}

	return nil
}

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/moby/sys/mount"
	"github.com/rs/zerolog"
	"github.com/tmacro/cola/pkg/losetup"
)

const OEM_PARTITION = "p6"
const ROOT_PARTITION = "p9"

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
			logger.Error().Err(err).Msg("Failed to unmount OEM partition")
		}

		logger.Debug().Str("mountpoint", rootMount).Msg("Unmounting root partition")
		err = mount.Unmount(rootMount)
		if err != nil {
			hasError = true
			logger.Error().Err(err).Msg("Failed to unmount root partition")
		}

		err = losetup.DetachDevice(loopDev)
		if err != nil {
			hasError = true
			logger.Error().Err(err).Msg("Failed to detach loop device")
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

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/moby/sys/mount"
	"github.com/tmacro/cola/pkg/losetup"
)

const OEM_PARTITION = "p6"
const ROOT_PARTITION = "p9"

func mountImage(image, mountpoint string) (func(), error) {
	loopDev, err := losetup.SetupDevice(image)
	if err != nil {
		return nil, err
	}

	oemMount := filepath.Join(mountpoint, "flatcar-oem")
	rootMount := filepath.Join(mountpoint, "flatcar-root")

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

	cleanupMounts := func() {
		mount.Unmount(oemMount)
		mount.Unmount(rootMount)
		losetup.DetachDevice(loopDev)
	}

	err = exec.Command("mount", "-o", "loop", loopDev+OEM_PARTITION, oemMount).Run()
	if err != nil {
		cleanupMounts()
		return nil, fmt.Errorf("failed to mount OEM partition: %w", err)
	}

	err = exec.Command("mount", "-o", "loop", loopDev+ROOT_PARTITION, rootMount).Run()
	if err != nil {
		cleanupMounts()
		return nil, fmt.Errorf("failed to mount root partition: %w", err)
	}

	return cleanupMounts, nil
}

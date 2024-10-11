package losetup

import (
	"fmt"
	"os/exec"
	"strings"
)

func SetupDevice(path string) (string, error) {
	loopdev, err := exec.Command("losetup", "--partscan", "--find", "--show", path).Output()
	if err != nil {
		return "", fmt.Errorf("failed to mount loop device: %w", err)
	}

	return strings.TrimSpace(string(loopdev)), nil
}

func DetachDevice(loopdev string) error {
	return exec.Command("losetup", "--detach", loopdev).Run()
}

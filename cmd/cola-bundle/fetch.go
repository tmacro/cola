package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/tmacro/cola/pkg/config"
	"github.com/tmacro/cola/pkg/ignition"
)

// downloadFile downloads a file from a given URL and saves it to the specified destination path.
// It returns an error if the download or file creation fails.
func downloadFile(url, destPath string, logger zerolog.Logger) error {
	// Create the destination directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create the destination file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	// Send GET request to the URL
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to send GET request: %w", err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the response body to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to destination file: %w", err)
	}

	logger.Info().Str("url", url).Str("destination", destPath).Msg("File downloaded successfully")
	return nil
}

func fetchExtensions(cfg *config.ApplianceConfig, workdir string, logger zerolog.Logger) error {
	for _, ext := range cfg.Extensions {
		extPath := filepath.Join(workdir, ignition.FormatExtensionName(ext.Name, ext.Version, ext.Arch)+".raw")
		url := ignition.FormatExtensionURL(ext.BakeryUrl, ext.Name, ext.Version, ext.Arch)
		logger.Info().Str("url", url).Str("destination", extPath).Msg("Downloading extension")
		if err := downloadFile(url, extPath, logger); err != nil {
			return fmt.Errorf("failed to download extension: %w", err)
		}

		transferCfgPath := filepath.Join(workdir, ext.Name+".conf")
		transferCfgURL := ignition.FormatExtensionTransferConfigURL(ext.BakeryUrl, ext.Name)
		logger.Info().Str("url", transferCfgURL).Str("destination", transferCfgPath).Msg("Downloading transfer config")
		if err := downloadFile(transferCfgURL, transferCfgPath, logger); err != nil {
			return fmt.Errorf("failed to download transfer config: %w", err)
		}
	}

	return nil
}

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
func downloadFile(url, destPath string, logger *zerolog.Logger) error {
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

func copyOrDownload(srcPath, srcURL, destPath string, logger *zerolog.Logger) error {
	if srcPath != "" {
		if _, err := os.Stat(srcPath); err == nil {
			logger.Info().Str("source", srcPath).Str("destination", destPath).Msg("Using existing file")
			if err := copyFile(srcPath, destPath, logger); err != nil {
				return fmt.Errorf("failed to copy file: %w", err)
			}
			return nil
		}
	}

	logger.Info().Str("url", srcURL).Str("destination", destPath).Msg("Downloading file")
	return downloadFile(srcURL, destPath, logger)
}

func fetchExtensions(cfg *config.ApplianceConfig, workdir string, logger *zerolog.Logger) error {
	for _, ext := range cfg.Extensions {
		extFilename := ignition.FormatExtensionName(ext.Name, ext.Version, ext.Arch) + ".raw"
		extDestPath := filepath.Join(workdir, extFilename)
		transferCfgDestPath := filepath.Join(workdir, ext.Name+".conf")

		extPath := ""
		transferCfgPath := ""
		if CLI.ExtensionDir != "" {
			extPath = filepath.Join(CLI.ExtensionDir, extFilename)
			transferCfgPath = filepath.Join(CLI.ExtensionDir, ext.Name+".conf")
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

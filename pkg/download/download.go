package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func Get(url, destPath string) error {
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

	return nil
}

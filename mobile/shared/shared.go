package shared

import (
	"fmt"
	"os"
	"path/filepath"
)

// GarbleBinDir This bin dir is used in both the go redirection binary as well as the main garble
// binary so that we don't have magic strings.
func GarbleBinDir() (dir string, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir. %w", err)
	}

	return filepath.Join(homeDir, ".garble/bin"), nil
}

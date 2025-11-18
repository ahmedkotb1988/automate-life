package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath expands ~ and $HOME in file paths
// Also handles paths within command strings
func ExpandPath(path string) string {
	if path == "" {
		return path
	}

	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return path
	}

	// If it's a command with arguments, expand paths in the arguments
	if strings.Contains(path, " ") {
		parts := strings.Fields(path)
		for i, part := range parts {
			parts[i] = expandSinglePath(part, homeDir)
		}
		return strings.Join(parts, " ")
	}

	return expandSinglePath(path, homeDir)
}

// expandSinglePath expands a single path string
func expandSinglePath(path string, homeDir string) string {
	if path == "" {
		return path
	}

	// Expand $HOME anywhere in the path
	if strings.Contains(path, "$HOME") {
		path = strings.ReplaceAll(path, "$HOME", homeDir)
	}

	// Expand ~ at the beginning
	if strings.HasPrefix(path, "~/") {
		path = filepath.Join(homeDir, path[2:])
	} else if path == "~" {
		path = homeDir
	}

	return path
}

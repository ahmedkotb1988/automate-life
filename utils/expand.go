package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandEnvVars expands all environment variables in a string
// Uses os.ExpandEnv() to automatically expand $VAR and ${VAR}
// Also handles tilde (~) expansion for home directory
func ExpandEnvVars(s string) string {
	if s == "" {
		return s
	}

	// Expand all environment variables using os.ExpandEnv
	// This handles: $HOME, $USER, $PATH, $GOPATH, or ANY variable
	s = os.ExpandEnv(s)

	// Handle tilde expansion (~ is not an env var)
	// Need to handle tildes anywhere in the string (e.g., in commands)
	if strings.Contains(s, "~") {
		s = expandAllTildes(s)
	}

	return s
}

// expandAllTildes expands all tildes in a string (handles commands with ~ in arguments)
func expandAllTildes(s string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.Getenv("HOME")
		if homeDir == "" {
			return s
		}
	}

	// If the whole string is just ~, return home
	if s == "~" {
		return homeDir
	}

	// If it starts with ~/
	if strings.HasPrefix(s, "~/") {
		s = filepath.Join(homeDir, s[2:])
	}

	// Handle ~ followed by space (e.g., in "command ~/path")
	// Replace " ~/" with " " + homeDir + "/"
	s = strings.ReplaceAll(s, " ~/", " "+homeDir+"/")

	return s
}

// expandTilde expands ~ to the user's home directory (legacy, kept for backwards compatibility)
func expandTilde(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.Getenv("HOME")
		if homeDir == "" {
			return path
		}
	}

	if path == "~" {
		return homeDir
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:])
	}

	return path
}

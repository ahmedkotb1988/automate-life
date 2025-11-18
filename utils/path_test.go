package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	// Save original HOME and restore after tests
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set a test HOME directory
	testHome := "/Users/testuser"
	os.Setenv("HOME", testHome)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Tilde with path",
			input:    "~/.ssh/id_rsa",
			expected: filepath.Join(testHome, ".ssh/id_rsa"),
		},
		{
			name:     "Just tilde",
			input:    "~",
			expected: testHome,
		},
		{
			name:     "$HOME with path",
			input:    "$HOME/documents/file.txt",
			expected: filepath.Join(testHome, "documents/file.txt"),
		},
		{
			name:     "$HOME in middle of path",
			input:    "/opt/$HOME/bin",
			expected: "/opt/" + testHome + "/bin",
		},
		{
			name:     "Multiple $HOME replacements",
			input:    "$HOME/src/$HOME/test",
			expected: testHome + "/src/" + testHome + "/test",
		},
		{
			name:     "Absolute path without expansion",
			input:    "/usr/local/bin",
			expected: "/usr/local/bin",
		},
		{
			name:     "Command with tilde path",
			input:    "python ~/scripts/test.py",
			expected: "python " + filepath.Join(testHome, "scripts/test.py"),
		},
		{
			name:     "Command with $HOME path",
			input:    "go build -o $HOME/bin/app",
			expected: "go build -o " + filepath.Join(testHome, "bin/app"),
		},
		{
			name:     "Command with multiple paths",
			input:    "cp ~/src/file.txt ~/dest/",
			expected: "cp " + filepath.Join(testHome, "src/file.txt") + " " + filepath.Join(testHome, "dest/"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandPath(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExpandPathNoHOME(t *testing.T) {
	// Save original HOME and restore after tests
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Unset HOME
	os.Unsetenv("HOME")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Tilde when HOME not set",
			input:    "~/.ssh/id_rsa",
			expected: "~/.ssh/id_rsa",
		},
		{
			name:     "$HOME when HOME not set",
			input:    "$HOME/documents",
			expected: "$HOME/documents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandPath(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExpandSinglePath(t *testing.T) {
	testHome := "/Users/testuser"

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Tilde with path",
			input:    "~/documents",
			expected: filepath.Join(testHome, "documents"),
		},
		{
			name:     "$HOME replacement",
			input:    "$HOME/bin",
			expected: filepath.Join(testHome, "bin"),
		},
		{
			name:     "No expansion needed",
			input:    "/usr/bin",
			expected: "/usr/bin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandSinglePath(tt.input, testHome)
			if result != tt.expected {
				t.Errorf("expandSinglePath(%q, %q) = %q, want %q", tt.input, testHome, result, tt.expected)
			}
		})
	}
}

package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandEnvVars(t *testing.T) {
	// Save and restore original environment
	originalHome := os.Getenv("HOME")
	originalUser := os.Getenv("USER")
	originalCustom := os.Getenv("CUSTOM_VAR")
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("USER", originalUser)
		if originalCustom != "" {
			os.Setenv("CUSTOM_VAR", originalCustom)
		} else {
			os.Unsetenv("CUSTOM_VAR")
		}
	}()

	// Set test environment variables
	testHome := "/Users/testuser"
	testUser := "testuser"
	os.Setenv("HOME", testHome)
	os.Setenv("USER", testUser)
	os.Setenv("CUSTOM_VAR", "/custom/path")

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
			name:     "No variables",
			input:    "/usr/local/bin",
			expected: "/usr/local/bin",
		},
		{
			name:     "Tilde only",
			input:    "~",
			expected: testHome,
		},
		{
			name:     "Tilde with path",
			input:    "~/.ssh/id_rsa",
			expected: filepath.Join(testHome, ".ssh/id_rsa"),
		},
		{
			name:     "$HOME variable",
			input:    "$HOME/documents",
			expected: filepath.Join(testHome, "documents"),
		},
		{
			name:     "${HOME} variable",
			input:    "${HOME}/documents",
			expected: filepath.Join(testHome, "documents"),
		},
		{
			name:     "$USER variable",
			input:    "/home/$USER/files",
			expected: "/home/testuser/files",
		},
		{
			name:     "Custom environment variable",
			input:    "$CUSTOM_VAR/data",
			expected: "/custom/path/data",
		},
		{
			name:     "Multiple variables",
			input:    "$HOME/users/$USER/data",
			expected: testHome + "/users/testuser/data",
		},
		{
			name:     "Variable at end",
			input:    "/path/to/$USER",
			expected: "/path/to/testuser",
		},
		{
			name:     "Tilde with $HOME",
			input:    "~/projects/$USER",
			expected: filepath.Join(testHome, "projects/testuser"),
		},
		{
			name:     "Undefined variable",
			input:    "$UNDEFINED_VAR/path",
			expected: "/path", // os.ExpandEnv replaces undefined vars with empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandEnvVars(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandEnvVars(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExpandEnvVarsNoHOME(t *testing.T) {
	// Save original HOME
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
			expected: "/documents", // os.ExpandEnv replaces with empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandEnvVars(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandEnvVars(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExpandTilde(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

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
			name:     "No tilde",
			input:    "/usr/bin",
			expected: "/usr/bin",
		},
		{
			name:     "Just tilde",
			input:    "~",
			expected: testHome,
		},
		{
			name:     "Tilde with slash",
			input:    "~/documents",
			expected: filepath.Join(testHome, "documents"),
		},
		{
			name:     "Tilde in middle (not expanded)",
			input:    "/home/~user",
			expected: "/home/~user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandTilde(tt.input)
			if result != tt.expected {
				t.Errorf("expandTilde(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExpandEnvVarsRealWorld(t *testing.T) {
	// Test real-world config scenarios
	originalHome := os.Getenv("HOME")
	originalGoPath := os.Getenv("GOPATH")
	defer func() {
		os.Setenv("HOME", originalHome)
		if originalGoPath != "" {
			os.Setenv("GOPATH", originalGoPath)
		} else {
			os.Unsetenv("GOPATH")
		}
	}()

	testHome := "/Users/testuser"
	os.Setenv("HOME", testHome)
	os.Setenv("GOPATH", "/Users/testuser/go")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "SSH key path",
			input:    "~/.ssh/id_rsa",
			expected: filepath.Join(testHome, ".ssh/id_rsa"),
		},
		{
			name:     "Output directory",
			input:    "$HOME/builds/output",
			expected: filepath.Join(testHome, "builds/output"),
		},
		{
			name:     "GOPATH bin",
			input:    "$GOPATH/bin",
			expected: "/Users/testuser/go/bin",
		},
		{
			name:     "Mixed variables",
			input:    "$GOPATH/src/$USER/project",
			expected: "/Users/testuser/go/src/" + os.Getenv("USER") + "/project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandEnvVars(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandEnvVars(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

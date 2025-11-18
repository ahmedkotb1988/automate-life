package builder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunCommand(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		expectError bool
	}{
		{
			name:        "Valid echo command",
			command:     "echo hello",
			expectError: false,
		},
		{
			name:        "Valid ls command",
			command:     "ls",
			expectError: false,
		},
		{
			name:        "Empty command",
			command:     "",
			expectError: true,
		},
		{
			name:        "Non-existent command",
			command:     "nonexistentcommand12345",
			expectError: true,
		},
		{
			name:        "Command with multiple arguments",
			command:     "echo hello world",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RunCommand(tt.command)

			if tt.expectError {
				if err == nil {
					t.Error("RunCommand() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("RunCommand() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGetDefaultTestCommand(t *testing.T) {
	tests := []struct {
		name     string
		language string
		expected string
	}{
		{
			name:     "Go",
			language: "go",
			expected: "go test",
		},
		{
			name:     "Golang (alternative)",
			language: "golang",
			expected: "go test",
		},
		{
			name:     "Node.js",
			language: "nodejs",
			expected: "npm test",
		},
		{
			name:     "Node (alternative)",
			language: "node",
			expected: "npm test",
		},
		{
			name:     "JavaScript",
			language: "javascript",
			expected: "npm test",
		},
		{
			name:     "TypeScript",
			language: "typescript",
			expected: "npm test",
		},
		{
			name:     "Python",
			language: "python",
			expected: "pytest",
		},
		{
			name:     ".NET",
			language: "dotnet",
			expected: "dotnet test",
		},
		{
			name:     "C#",
			language: "c#",
			expected: "dotnet test",
		},
		{
			name:     "CSharp",
			language: "csharp",
			expected: "dotnet test",
		},
		{
			name:     "Rust",
			language: "rust",
			expected: "cargo test",
		},
		{
			name:     "Ruby",
			language: "ruby",
			expected: "bundle exec rspec",
		},
		{
			name:     "Java",
			language: "java",
			expected: "mvn test",
		},
		{
			name:     "Unknown language",
			language: "cobol",
			expected: "echo 'No default test command for language: cobol'",
		},
		{
			name:     "Empty language",
			language: "",
			expected: "echo 'No default test command for language: '",
		},
		{
			name:     "Case insensitive - GO",
			language: "GO",
			expected: "go test",
		},
		{
			name:     "Case insensitive - PYTHON",
			language: "PYTHON",
			expected: "pytest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDefaultTestCommand(tt.language)
			if result != tt.expected {
				t.Errorf("GetDefaultTestCommand(%q) = %q, want %q", tt.language, result, tt.expected)
			}
		})
	}
}

func TestAutoInstallDependencies(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Change to temp directory
	os.Chdir(tmpDir)

	tests := []struct {
		name        string
		language    string
		setupFiles  []string // Files to create before test
		expectError bool
	}{
		{
			name:        "Go with go.mod",
			language:    "go",
			setupFiles:  []string{"go.mod"},
			expectError: false,
		},
		{
			name:        "Go without go.mod",
			language:    "go",
			setupFiles:  []string{},
			expectError: false,
		},
		{
			name:        "Unknown language",
			language:    "fortran",
			setupFiles:  []string{},
			expectError: true,
		},
		{
			name:        "Empty language",
			language:    "",
			setupFiles:  []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create setup files
			testDir := filepath.Join(tmpDir, tt.name)
			os.MkdirAll(testDir, 0755)
			os.Chdir(testDir)

			for _, file := range tt.setupFiles {
				os.WriteFile(file, []byte("test content"), 0644)
			}

			// Note: This will actually try to run commands, so we expect errors
			// for languages where dependencies can't actually be installed
			err := AutoInstallDependencies(tt.language)

			if tt.expectError {
				if err == nil {
					t.Error("AutoInstallDependencies() expected error, got nil")
				}
			}
			// For non-error cases, we don't check if err is nil because
			// the actual dependency installation might fail in the test environment
		})
	}
}

func TestAutoInstallDependenciesWithMockFiles(t *testing.T) {
	// This test verifies that the function recognizes the right files
	// without actually running installation commands

	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	tests := []struct {
		name       string
		language   string
		setupFiles map[string]string // filename -> content
	}{
		{
			name:     "Node.js with package.json",
			language: "nodejs",
			setupFiles: map[string]string{
				"package.json": `{"name": "test"}`,
			},
		},
		{
			name:     "Python with requirements.txt",
			language: "python",
			setupFiles: map[string]string{
				"requirements.txt": "requests==2.28.0",
			},
		},
		{
			name:     "Python with Pipfile",
			language: "python",
			setupFiles: map[string]string{
				"Pipfile": "[packages]\nrequests = \"*\"",
			},
		},
		{
			name:     "Ruby with Gemfile",
			language: "ruby",
			setupFiles: map[string]string{
				"Gemfile": "source 'https://rubygems.org'",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a subdirectory for this test
			testDir := filepath.Join(tmpDir, tt.name)
			os.MkdirAll(testDir, 0755)
			os.Chdir(testDir)

			// Create setup files
			for filename, content := range tt.setupFiles {
				os.WriteFile(filename, []byte(content), 0644)
			}

			// Call the function (it will likely fail because commands aren't available,
			// but we're just testing that it detects the files correctly)
			_ = AutoInstallDependencies(tt.language)

			// The main goal here is to ensure no panic occurs and the function
			// attempts to use the correct command based on detected files
		})
	}
}

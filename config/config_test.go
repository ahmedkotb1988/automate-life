package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigTemplate(t *testing.T) {
	template := DefaultConfigTemplate()

	if template == "" {
		t.Error("DefaultConfigTemplate() returned empty string")
	}

	// Check if template contains expected keys
	expectedKeys := []string{
		"project",
		"git",
		"build",
		"azure",
		"environment",
		"repo_url",
		"auth_type",
		"language",
	}

	for _, key := range expectedKeys {
		if !contains(template, key) {
			t.Errorf("DefaultConfigTemplate() missing expected key: %s", key)
		}
	}
}

func TestCreateConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_config.json")

	content := `{"test": "data"}`

	// Test successful creation
	err := Create(testFile, content)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Create() did not create the file")
	}

	// Verify content
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}
	if string(data) != content {
		t.Errorf("File content = %q, want %q", string(data), content)
	}

	// Test creating file that already exists
	err = Create(testFile, content)
	if err == nil {
		t.Error("Create() should fail when file already exists")
	}
	if err.Error() != "config file already exists" {
		t.Errorf("Create() error = %q, want %q", err.Error(), "config file already exists")
	}
}

func TestLoadConfig(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", "/Users/testuser")

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_config.json")

	// Create a valid config file
	validConfig := `{
  "project": {
    "name": "TestProject",
    "type": "backend",
    "description": "Test Description"
  },
  "git": {
    "provider": "github",
    "repo_url": "https://github.com/test/repo",
    "branch": "main",
    "auth_type": "token",
    "username": "",
    "password": "",
    "token": "test-token",
    "ssh_key_path": "~/.ssh/id_rsa"
  },
  "build": {
    "language": "go",
    "install_command": "",
    "build_command": "go build",
    "test_command": "go test",
    "output_dir": "$HOME/bin"
  },
  "azure": {
    "subscription_id": "",
    "resource_group": "",
    "app_name": "",
    "deployment_type": "webapp",
    "region": "eastus"
  },
  "environment": {
    "variables": {
      "ENV": "test",
      "LOG_DIR": "~/logs"
    }
  }
}`

	err := os.WriteFile(testFile, []byte(validConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loading config
	cfg, err := Load(testFile)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify loaded values
	if cfg.Project.Name != "TestProject" {
		t.Errorf("Project.Name = %q, want %q", cfg.Project.Name, "TestProject")
	}
	if cfg.Git.Provider != "github" {
		t.Errorf("Git.Provider = %q, want %q", cfg.Git.Provider, "github")
	}
	if cfg.Git.AuthType != "token" {
		t.Errorf("Git.AuthType = %q, want %q", cfg.Git.AuthType, "token")
	}
	if cfg.Build.Language != "go" {
		t.Errorf("Build.Language = %q, want %q", cfg.Build.Language, "go")
	}

	// Verify paths were expanded
	expectedSSHPath := filepath.Join("/Users/testuser", ".ssh/id_rsa")
	if cfg.Git.SSHKeyPath != expectedSSHPath {
		t.Errorf("Git.SSHKeyPath = %q, want %q (path should be expanded)", cfg.Git.SSHKeyPath, expectedSSHPath)
	}

	expectedOutputDir := filepath.Join("/Users/testuser", "bin")
	if cfg.Build.OutputDir != expectedOutputDir {
		t.Errorf("Build.OutputDir = %q, want %q (path should be expanded)", cfg.Build.OutputDir, expectedOutputDir)
	}

	// Test loading non-existent file
	_, err = Load(filepath.Join(tmpDir, "nonexistent.json"))
	if err == nil {
		t.Error("Load() should fail for non-existent file")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.json")

	invalidJSON := `{"project": invalid json}`
	err := os.WriteFile(testFile, []byte(invalidJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = Load(testFile)
	if err == nil {
		t.Error("Load() should fail for invalid JSON")
	}
}

func TestExpandPaths(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	testHome := "/Users/testuser"
	os.Setenv("HOME", testHome)

	cfg := &Config{
		Git: GitConfig{
			SSHKeyPath: "~/.ssh/id_rsa",
		},
		Build: BuildConfig{
			OutputDir:      "$HOME/bin",
			BuildCommand:   "go build -o ~/app",
			TestCommand:    "go test",
			InstallCommand: "go mod download",
		},
		Environment: EnvironmentConfig{
			Variables: map[string]string{
				"DATA_DIR": "~/data",
				"LOG_PATH": "$HOME/logs/app.log",
				"PLAIN":    "/var/log",
			},
		},
	}

	cfg.ExpandPaths()

	// Check Git paths
	expectedSSH := filepath.Join(testHome, ".ssh/id_rsa")
	if cfg.Git.SSHKeyPath != expectedSSH {
		t.Errorf("Git.SSHKeyPath = %q, want %q", cfg.Git.SSHKeyPath, expectedSSH)
	}

	// Check Build paths
	expectedOutputDir := filepath.Join(testHome, "bin")
	if cfg.Build.OutputDir != expectedOutputDir {
		t.Errorf("Build.OutputDir = %q, want %q", cfg.Build.OutputDir, expectedOutputDir)
	}

	expectedBuildCmd := "go build -o " + filepath.Join(testHome, "app")
	if cfg.Build.BuildCommand != expectedBuildCmd {
		t.Errorf("Build.BuildCommand = %q, want %q", cfg.Build.BuildCommand, expectedBuildCmd)
	}

	// Check Environment variables
	expectedDataDir := filepath.Join(testHome, "data")
	if cfg.Environment.Variables["DATA_DIR"] != expectedDataDir {
		t.Errorf("Environment.Variables[DATA_DIR] = %q, want %q", cfg.Environment.Variables["DATA_DIR"], expectedDataDir)
	}

	expectedLogPath := filepath.Join(testHome, "logs/app.log")
	if cfg.Environment.Variables["LOG_PATH"] != expectedLogPath {
		t.Errorf("Environment.Variables[LOG_PATH] = %q, want %q", cfg.Environment.Variables["LOG_PATH"], expectedLogPath)
	}

	// Plain path should remain unchanged
	if cfg.Environment.Variables["PLAIN"] != "/var/log" {
		t.Errorf("Environment.Variables[PLAIN] = %q, want %q", cfg.Environment.Variables["PLAIN"], "/var/log")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && s != substr && len(s) >= len(substr) && s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsHelper(s, substr)
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

package tests

import (
	"automateLife/config"
	"os"
	"path/filepath"
	"testing"
)

func TestValidate(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Create a temporary directory for SSH key tests
	tmpDir := t.TempDir()
	testSSHKey := filepath.Join(tmpDir, "test_key")
	os.WriteFile(testSSHKey, []byte("test key"), 0600)

	tests := []struct {
		name        string
		config      config.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid config with token auth",
			config: config.Config{
				Git: config.GitConfig{
					RepoUrl:  "https://github.com/test/repo",
					AuthType: "token",
					Token:    "test-token",
				},
				Project: config.ProjectConfig{
					Type: "backend",
				},
			},
			expectError: false,
		},
		{
			name: "Valid config with basic auth",
			config: config.Config{
				Git: config.GitConfig{
					RepoUrl:  "https://github.com/test/repo",
					AuthType: "basic",
					UserName: "user",
					Password: "pass",
				},
				Project: config.ProjectConfig{
					Type: "backend",
				},
			},
			expectError: false,
		},
		{
			name: "Valid config with SSH auth",
			config: config.Config{
				Git: config.GitConfig{
					RepoUrl:    "git@github.com:test/repo.git",
					AuthType:   "ssh",
					SSHKeyPath: testSSHKey,
				},
				Project: config.ProjectConfig{
					Type: "backend",
				},
			},
			expectError: false,
		},
		{
			name: "Missing repo URL",
			config: config.Config{
				Git: config.GitConfig{
					AuthType: "token",
					Token:    "test-token",
				},
				Project: config.ProjectConfig{
					Type: "backend",
				},
			},
			expectError: true,
			errorMsg:    "git.repo_url is required",
		},
		{
			name: "Missing project type",
			config: config.Config{
				Git: config.GitConfig{
					RepoUrl:  "https://github.com/test/repo",
					AuthType: "token",
					Token:    "test-token",
				},
				Project: config.ProjectConfig{},
			},
			expectError: true,
			errorMsg:    "project.type is required",
		},
		{
			name: "Token auth missing token",
			config: config.Config{
				Git: config.GitConfig{
					RepoUrl:  "https://github.com/test/repo",
					AuthType: "token",
				},
				Project: config.ProjectConfig{
					Type: "backend",
				},
			},
			expectError: true,
			errorMsg:    "git.token is required when auth_type is 'token'",
		},
		{
			name: "Basic auth missing username",
			config: config.Config{
				Git: config.GitConfig{
					RepoUrl:  "https://github.com/test/repo",
					AuthType: "basic",
					Password: "pass",
				},
				Project: config.ProjectConfig{
					Type: "backend",
				},
			},
			expectError: true,
			errorMsg:    "git.username and git.password are required when auth_type is 'basic'",
		},
		{
			name: "Basic auth missing password",
			config: config.Config{
				Git: config.GitConfig{
					RepoUrl:  "https://github.com/test/repo",
					AuthType: "basic",
					UserName: "user",
				},
				Project: config.ProjectConfig{
					Type: "backend",
				},
			},
			expectError: true,
			errorMsg:    "git.username and git.password are required when auth_type is 'basic'",
		},
		{
			name: "SSH auth missing key path",
			config: config.Config{
				Git: config.GitConfig{
					RepoUrl:  "git@github.com:test/repo.git",
					AuthType: "ssh",
				},
				Project: config.ProjectConfig{
					Type: "backend",
				},
			},
			expectError: true,
			errorMsg:    "git.ssh_key_path is required when auth_type is 'ssh'",
		},
		{
			name: "SSH auth with non-existent key",
			config: config.Config{
				Git: config.GitConfig{
					RepoUrl:    "git@github.com:test/repo.git",
					AuthType:   "ssh",
					SSHKeyPath: "/nonexistent/key",
				},
				Project: config.ProjectConfig{
					Type: "backend",
				},
			},
			expectError: true,
			errorMsg:    "SSH key not found at",
		},
		{
			name: "Invalid auth type",
			config: config.Config{
				Git: config.GitConfig{
					RepoUrl:  "https://github.com/test/repo",
					AuthType: "invalid",
				},
				Project: config.ProjectConfig{
					Type: "backend",
				},
			},
			expectError: true,
			errorMsg:    "git.auth_type must be 'token', 'basic', or 'ssh'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Validate() expected error containing %q, got nil", tt.errorMsg)
				} else if !containsSubstring(err.Error(), tt.errorMsg) {
					t.Errorf("Validate() error = %q, want error containing %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateSSHWithExpandedPath(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create SSH directory and key
	sshDir := filepath.Join(tmpDir, ".ssh")
	os.Mkdir(sshDir, 0700)
	sshKey := filepath.Join(sshDir, "id_rsa")
	os.WriteFile(sshKey, []byte("test key"), 0600)

	cfg := config.Config{
		Git: config.GitConfig{
			RepoUrl:    "git@github.com:test/repo.git",
			AuthType:   "ssh",
			SSHKeyPath: "~/.ssh/id_rsa", // Using tilde
		},
		Project: config.ProjectConfig{
			Type: "backend",
		},
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate() should succeed with expanded path, got error: %v", err)
	}
}

// Helper function
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

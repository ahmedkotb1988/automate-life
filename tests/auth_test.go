package tests

import (
	"automateLife/config"
	"automateLife/git"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildAuthURL(t *testing.T) {
	tests := []struct {
		name        string
		config      config.GitConfig
		expected    string
		expectError bool
		errorMsg    string
	}{
		{
			name: "Token auth with HTTPS",
			config: config.GitConfig{
				RepoUrl:  "https://github.com/user/repo.git",
				AuthType: "token",
				Token:    "ghp_test123",
			},
			expected:    "https://github.com/user/repo.git",
			expectError: false,
		},
		{
			name: "Token auth with HTTP",
			config: config.GitConfig{
				RepoUrl:  "http://github.com/user/repo.git",
				AuthType: "token",
				Token:    "test_token",
			},
			expected:    "http://github.com/user/repo.git",
			expectError: false,
		},
		{
			name: "Basic auth with HTTPS",
			config: config.GitConfig{
				RepoUrl:  "https://github.com/user/repo.git",
				AuthType: "basic",
				UserName: "username",
				Password: "password",
			},
			expected:    "https://github.com/user/repo.git",
			expectError: false,
		},
		{
			name: "Basic auth with HTTP",
			config: config.GitConfig{
				RepoUrl:  "http://gitlab.com/user/repo.git",
				AuthType: "basic",
				UserName: "user",
				Password: "pass123",
			},
			expected:    "http://gitlab.com/user/repo.git",
			expectError: false,
		},
		{
			name: "SSH auth returns original URL",
			config: config.GitConfig{
				RepoUrl:    "git@github.com:user/repo.git",
				AuthType:   "ssh",
				SSHKeyPath: "/path/to/key",
			},
			expected:    "git@github.com:user/repo.git",
			expectError: false,
		},
		{
			name: "Token auth missing token",
			config: config.GitConfig{
				RepoUrl:  "https://github.com/user/repo.git",
				AuthType: "token",
				Token:    "",
			},
			expected:    "https://github.com/user/repo.git",
			expectError: false,
		},
		{
			name: "Token auth with invalid URL",
			config: config.GitConfig{
				RepoUrl:  "git@github.com:user/repo.git",
				AuthType: "token",
				Token:    "test_token",
			},
			expectError: true,
			errorMsg:    "repo_url must start with http:// or https://",
		},
		{
			name: "Basic auth missing username",
			config: config.GitConfig{
				RepoUrl:  "https://github.com/user/repo.git",
				AuthType: "basic",
				UserName: "",
				Password: "password",
			},
			expected:    "https://github.com/user/repo.git",
			expectError: false,
		},
		{
			name: "Basic auth missing password",
			config: config.GitConfig{
				RepoUrl:  "https://github.com/user/repo.git",
				AuthType: "basic",
				UserName: "username",
				Password: "",
			},
			expectError: false,
			expected:    "https://github.com/user/repo.git",
		},
		{
			name: "Basic auth with invalid URL",
			config: config.GitConfig{
				RepoUrl:  "git@github.com:user/repo.git",
				AuthType: "basic",
				UserName: "user",
				Password: "pass",
			},
			expectError: true,
			errorMsg:    "repo_url must start with http:// or https://",
		},
		{
			name: "Unsupported auth type",
			config: config.GitConfig{
				RepoUrl:  "https://github.com/user/repo.git",
				AuthType: "oauth",
			},
			expectError: true,
			errorMsg:    "unsupported auth type",
		},
		{
			name: "Empty repo URL",
			config: config.GitConfig{
				RepoUrl:  "",
				AuthType: "token",
				Token:    "test",
			},
			expectError: true,
			errorMsg:    "repo_url is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := git.BuildAuthURL(&tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("git.BuildAuthURL() expected error containing %q, got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("git.BuildAuthURL() error = %q, want error containing %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("git.BuildAuthURL() unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("git.BuildAuthURL() = %q, want %q", result, tt.expected)
				}
			}
		})
	}
}

func TestSetupSSH(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create a test SSH key
	sshDir := filepath.Join(tmpDir, ".ssh")
	os.Mkdir(sshDir, 0700)
	sshKey := filepath.Join(sshDir, "id_rsa")
	os.WriteFile(sshKey, []byte("test key"), 0600)

	tests := []struct {
		name        string
		keyPath     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid absolute path",
			keyPath:     sshKey,
			expectError: false,
		},
		{
			name:        "Valid path with tilde",
			keyPath:     "~/.ssh/id_rsa",
			expectError: false,
		},
		{
			name:        "Empty key path",
			keyPath:     "",
			expectError: true,
			errorMsg:    "ssh_key_path must be provided",
		},
		{
			name:        "Non-existent key",
			keyPath:     "/nonexistent/key",
			expectError: true,
			errorMsg:    "SSH key not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := git.SetupSSH(tt.keyPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("git.SetupSSH() expected error containing %q, got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("git.SetupSSH() error = %q, want error containing %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("git.SetupSSH() unexpected error: %v", err)
				}

				// Verify GIT_SSH_COMMAND is set
				gitSSHCmd := os.Getenv("GIT_SSH_COMMAND")
				if gitSSHCmd == "" {
					t.Error("git.SetupSSH() did not set GIT_SSH_COMMAND")
				}
				if !strings.Contains(gitSSHCmd, "ssh -i") {
					t.Errorf("GIT_SSH_COMMAND = %q, should contain 'ssh -i'", gitSSHCmd)
				}
			}
		})
	}
}

func TestGetProjectDirName(t *testing.T) {
	tests := []struct {
		name     string
		repoUrl  string
		expected string
	}{
		{
			name:     "HTTPS URL with .git",
			repoUrl:  "https://github.com/user/myproject.git",
			expected: "myproject",
		},
		{
			name:     "HTTPS URL without .git",
			repoUrl:  "https://github.com/user/myproject",
			expected: "myproject",
		},
		{
			name:     "SSH URL with .git",
			repoUrl:  "git@github.com:user/awesome-repo.git",
			expected: "awesome-repo",
		},
		{
			name:     "SSH URL without .git",
			repoUrl:  "git@github.com:user/awesome-repo",
			expected: "awesome-repo",
		},
		{
			name:     "URL with nested path",
			repoUrl:  "https://gitlab.com/group/subgroup/project.git",
			expected: "project",
		},
		{
			name:     "Simple name",
			repoUrl:  "myrepo",
			expected: "myrepo",
		},
		{
			name:     "Empty URL",
			repoUrl:  "",
			expected: "",
		},
		{
			name:     "URL ending with slash",
			repoUrl:  "https://github.com/user/repo/",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := git.GetProjectDirName(tt.repoUrl)
			if result != tt.expected {
				t.Errorf("git.GetProjectDirName(%q) = %q, want %q", tt.repoUrl, result, tt.expected)
			}
		})
	}
}


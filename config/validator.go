package config

import (
	"fmt"
	"os"
)

func (c *Config) Validate() error {
	if c.Git.RepoUrl == "" {
		return fmt.Errorf("git.repo_url is required")
	}
	if c.Project.Type == "" {
		return fmt.Errorf("project.type is required")
	}

	switch c.Git.AuthType {
	case "token":
		if c.Git.Token == "" {
			return fmt.Errorf("git.token is required when auth_type is 'token'")
		}
	case "basic":
		if c.Git.UserName == "" || c.Git.Password == "" {
			return fmt.Errorf("git.username and git.password are required when auth_type is 'basic'")
		}
	case "ssh":
		if c.Git.SSHKeyPath == "" {
			return fmt.Errorf("git.ssh_key_path is required when auth_type is 'ssh'")
		}
		if _, err := os.Stat(c.Git.SSHKeyPath); os.IsNotExist(err) {
			return fmt.Errorf("SSH key not found at: %s", c.Git.SSHKeyPath)
		}
	default:
		return fmt.Errorf("git.auth_type must be 'token', 'basic', or 'ssh'")
	}

	return nil
}

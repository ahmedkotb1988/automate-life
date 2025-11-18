package git

import (
	"automateLife/config"
	"automateLife/utils"
	"fmt"
	"os"
	"strings"
)

func BuildAuthURL(config *config.GitConfig) (string, error) {
	switch config.AuthType {
	case "token":
		return buildTokenURL(config)
	case "basic":
		return buildBasicAuthURL(config)
	case "ssh":
		return config.RepoUrl, nil
	default:
		return "", fmt.Errorf("unsupported auth type: %s . Use 'token', 'basic' or 'ssh'", config.AuthType)
	}
}

func buildTokenURL(config *config.GitConfig) (string, error) {
	if config.Token == "" {
		return "", fmt.Errorf("token is required when auth_type is 'token'")
	}

	if len(config.RepoUrl) == 0 {
		return "", fmt.Errorf("repo_url is empty")
	}
	if strings.HasPrefix(config.RepoUrl, "http://") {
		return strings.Replace(config.RepoUrl, "http://", fmt.Sprintf("http://%s@", config.Token), 1), nil
	} else if strings.HasPrefix(config.RepoUrl, "https://") {
		return strings.Replace(config.RepoUrl, "https://", fmt.Sprintf("https://%s@", config.Token), 1), nil
	}

	return "", fmt.Errorf("repo_url must start with http:// or https://")
}

func buildBasicAuthURL(config *config.GitConfig) (string, error) {
	if config.UserName == "" || config.Password == "" {
		return "", fmt.Errorf("username and password must not be empty when auth_type is 'basic'")
	}

	if strings.HasPrefix(config.RepoUrl, "http://") {
		credentials := fmt.Sprintf("%s:%s", config.UserName, config.Password)
		return strings.Replace(config.RepoUrl, "http://", fmt.Sprintf("http://%s@", credentials), 1), nil
	} else if strings.HasPrefix(config.RepoUrl, "https://") {
		credentials := fmt.Sprintf("%s:%s", config.UserName, config.Password)
		return strings.Replace(config.RepoUrl, "https://", fmt.Sprintf("https://%s@", credentials), 1), nil
	}

	return "", fmt.Errorf("repo_url must start with http:// or https://")
}

func SetupSSH(keyPath string) error {
	if keyPath == "" {
		return fmt.Errorf("ssh_key_path must be provided when auth_type is 'ssh'")
	}

	// Expand environment variables and tilde in the key path
	expandedPath := utils.ExpandEnvVars(keyPath)

	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		return fmt.Errorf("SSH key not found at %s (expanded from: %s)", expandedPath, keyPath)
	}

	sshCommand := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no", expandedPath)
	os.Setenv("GIT_SSH_COMMAND", sshCommand)

	return nil
}

func GetProjectDirName(repoUrl string) string {

	repoUrl = strings.TrimSuffix(repoUrl, ".git")
	parts := strings.Split(repoUrl, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

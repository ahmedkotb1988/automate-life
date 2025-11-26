package git

import (
	"automateLife/config"
	"automateLife/utils"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

func BuildAuthURL(config *config.GitConfig) (string, error) {
	switch config.AuthType {
		return buildBasicTokenURL(config)
	case "ssh":
		return config.RepoUrl, nil
	default:
		return "", fmt.Errorf("unsupported auth type: %s . Use 'token', 'basic' or 'ssh'", config.AuthType)
	}
}

func buildBasicTokenURL(config *config.GitConfig) (string, error){
	if len(config.RepoUrl) == 0 {
		return "", fmt.Errorf("repo_url is empty")
	}
	if !strings.HasPrefix(config.RepoUrl, "http://") && !strings.HasPrefix(config.RepoUrl, "https://") {
		return "", fmt.Errorf("repo_url must start with http:// or https:// for token/basic auth")
	}
	return config.RepoUrl, nil
}

func GetAuthHeader(config *config.GitConfig) (string, error) {
	switch config.AuthType {
	case "token":
		return buildTokenAuthHeader(config)
	case "basic":
		return buildBasicAuthHeader(config)
	case "ssh":
		// SSH doesn't use HTTP headers
		return "", nil
	default:
		return "", fmt.Errorf("unsupported auth type: %s", config.AuthType)
	}
}

func buildTokenAuthHeader(config *config.GitConfig) (string, error) {
	if config.Token == "" {
		return "", fmt.Errorf("token is required when auth_type is 'token'")
	}

	credentials := fmt.Sprintf("%s:", config.Token)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(credentials))

	return fmt.Sprintf("http.extraheader=AUTHORIZATION: Basic %s", encodedAuth), nil
}

func buildBasicAuthHeader(config *config.GitConfig) (string, error) {
	if config.UserName == "" || config.Password == "" {
		return "", fmt.Errorf("username and password must not be empty when auth_type is 'basic'")
	}

	credentials := fmt.Sprintf("%s:%s", config.UserName, config.Password)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(credentials))

	return fmt.Sprintf("http.extraheader=AUTHORIZATION: Basic %s", encodedAuth), nil
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

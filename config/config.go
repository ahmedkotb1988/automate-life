package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const DefaultConfigFileName = "ConfigFile.json"

type Config struct {
	Git         GitConfig         `json:"git"`
	Project     ProjectConfig     `json:"project"`
	Build       BuildConfig       `json:"build"`
	Azure       AzureConfig       `json:"azure"`
	Environment EnvironmentConfig `json:"environment"`
}

type GitConfig struct {
	RepoUrl    string `json:"repo_url"`
	AuthType   string `json:"auth_type"`
	UserName   string `json:"username"`
	Password   string `json:"password"`
	Branch     string `json:"branch"`
	Token      string `json:"token"`
	SSHKeyPath string `json:"ssh_key_path"`
}

type ProjectConfig struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type BuildConfig struct {
	Language       string `json:"language"` //go, dotnet, python
	InstallCommand string `json:"install_command"`
	BuildCommand   string `json:"build_command"`
	TestCommand    string `json:"test_command"`
	OutputDir      string `json:"output_dir"`
}

type AzureConfig struct {
	SubscriptionID string `json:"subscription_id"`
	ResourceGroup  string `json:"resource_group"`
	AppName        string `json:"app_name"`
	DeploymentType string `json:"deployment_type"` // "webapp", "container", "function"
	Region         string `json:"region"`
}

type EnvironmentConfig struct {
	Variables map[string]string `json:"variables"`
}

func DefaultConfigTemplate() string {
	return `{
  "project": {
    "name": "",
    "type": "backend",
    "description": ""
  },
  "git": {
    "repo_url": "",
    "branch": "main",
    "auth_type": "token",
    "username": "",
    "password": "",
    "token": "",
    "ssh_key_path": ""
  },
  "build": {
    "language": "go",
    "install_command": "",
    "build_command": "",
    "test_command": "",
    "output_dir": "./bin"
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
      "ENV": "production"
    }
  }
}`
}

func Load(fileName string) (*Config, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("config file not found: %w", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return &config, nil
}

func Create(fileName string, content string) error {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("config file already exists")
		}
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

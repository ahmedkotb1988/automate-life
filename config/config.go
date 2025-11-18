package config

import (
	"automateLife/utils"
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
	Provider   string `json:"provider"`
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
    "provider": "github",
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

	// Expand all paths in the config
	config.ExpandPaths()

	return &config, nil
}

// ExpandPaths expands all environment variables in all string fields of the config
// Users can use $VAR, ${VAR}, or ~ in any field
func (c *Config) ExpandPaths() {
	// Expand Git fields
	c.Git.Provider = utils.ExpandEnvVars(c.Git.Provider)
	c.Git.RepoUrl = utils.ExpandEnvVars(c.Git.RepoUrl)
	c.Git.AuthType = utils.ExpandEnvVars(c.Git.AuthType)
	c.Git.UserName = utils.ExpandEnvVars(c.Git.UserName)
	c.Git.Password = utils.ExpandEnvVars(c.Git.Password)
	c.Git.Branch = utils.ExpandEnvVars(c.Git.Branch)
	c.Git.Token = utils.ExpandEnvVars(c.Git.Token)
	c.Git.SSHKeyPath = utils.ExpandEnvVars(c.Git.SSHKeyPath)

	// Expand Project fields
	c.Project.Name = utils.ExpandEnvVars(c.Project.Name)
	c.Project.Type = utils.ExpandEnvVars(c.Project.Type)
	c.Project.Description = utils.ExpandEnvVars(c.Project.Description)

	// Expand Build fields
	c.Build.Language = utils.ExpandEnvVars(c.Build.Language)
	c.Build.InstallCommand = utils.ExpandEnvVars(c.Build.InstallCommand)
	c.Build.BuildCommand = utils.ExpandEnvVars(c.Build.BuildCommand)
	c.Build.TestCommand = utils.ExpandEnvVars(c.Build.TestCommand)
	c.Build.OutputDir = utils.ExpandEnvVars(c.Build.OutputDir)

	// Expand Azure fields
	c.Azure.SubscriptionID = utils.ExpandEnvVars(c.Azure.SubscriptionID)
	c.Azure.ResourceGroup = utils.ExpandEnvVars(c.Azure.ResourceGroup)
	c.Azure.AppName = utils.ExpandEnvVars(c.Azure.AppName)
	c.Azure.DeploymentType = utils.ExpandEnvVars(c.Azure.DeploymentType)
	c.Azure.Region = utils.ExpandEnvVars(c.Azure.Region)

	// Expand environment variable values
	for key, value := range c.Environment.Variables {
		c.Environment.Variables[key] = utils.ExpandEnvVars(value)
	}
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

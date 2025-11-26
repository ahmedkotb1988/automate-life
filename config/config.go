package config

import (
	"automateLife/utils"
	"encoding/json"
	"fmt"
	"os"
)

const DefaultConfigFileName = "ConfigFile.json"

type Config struct {
	Git              GitConfig              `json:"git"`
	Project          ProjectConfig          `json:"project"`
	Build            BuildConfig            `json:"build"`
	IOS              IOSConfig              `json:"ios,omitempty"`
	AppStoreConnect  AppStoreConnectConfig  `json:"app_store_connect,omitempty"`
	Azure            AzureConfig            `json:"azure"`
	Environment      EnvironmentConfig      `json:"environment"`
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

type IOSConfig struct {
	WorkspacePath       string `json:"workspace_path,omitempty"`       // Path to .xcworkspace file (optional, auto-detected if empty)
	ProjectPath         string `json:"project_path,omitempty"`         // Path to .xcodeproj file (optional, auto-detected if empty)
	Scheme              string `json:"scheme"`                         // Xcode scheme name
	Configuration       string `json:"configuration"`                  // Debug or Release
	SDK                 string `json:"sdk"`                            // iphoneos, iphonesimulator
	ExportMethod        string `json:"export_method"`                  // app-store, ad-hoc, development, enterprise
	BundleID            string `json:"bundle_id,omitempty"`            // App bundle identifier
	TeamID              string `json:"team_id,omitempty"`              // Apple Developer Team ID
	ProvisioningProfile string `json:"provisioning_profile,omitempty"` // Provisioning profile name (manual signing)
	CodeSignIdentity    string `json:"code_sign_identity,omitempty"`   // Code signing identity (manual signing)
	AutomaticSigning    bool   `json:"automatic_signing"`              // Use automatic code signing
	ArchivePath         string `json:"archive_path,omitempty"`         // Where to save .xcarchive (optional)
	ExportPath          string `json:"export_path,omitempty"`          // Where to export IPA (optional, defaults to current dir)
	UploadToTestFlight  bool   `json:"upload_to_testflight"`           // Whether to prompt for TestFlight upload
}

type AppStoreConnectConfig struct {
	AppleID            string `json:"apple_id"`              // Apple ID email
	AppSpecificPassword string `json:"app_specific_password"` // App-specific password for Apple ID
	APIKeyID           string `json:"api_key_id"`            // App Store Connect API Key ID
	APIIssuerID        string `json:"api_issuer_id"`         // API Issuer ID
	APIKeyPath         string `json:"api_key_path"`          // Path to .p8 API key file
	TeamID             string `json:"team_id"`               // App Store Connect Team ID
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

	// Expand iOS fields
	c.IOS.WorkspacePath = utils.ExpandEnvVars(c.IOS.WorkspacePath)
	c.IOS.ProjectPath = utils.ExpandEnvVars(c.IOS.ProjectPath)
	c.IOS.Scheme = utils.ExpandEnvVars(c.IOS.Scheme)
	c.IOS.Configuration = utils.ExpandEnvVars(c.IOS.Configuration)
	c.IOS.SDK = utils.ExpandEnvVars(c.IOS.SDK)
	c.IOS.ExportMethod = utils.ExpandEnvVars(c.IOS.ExportMethod)
	c.IOS.BundleID = utils.ExpandEnvVars(c.IOS.BundleID)
	c.IOS.TeamID = utils.ExpandEnvVars(c.IOS.TeamID)
	c.IOS.ProvisioningProfile = utils.ExpandEnvVars(c.IOS.ProvisioningProfile)
	c.IOS.CodeSignIdentity = utils.ExpandEnvVars(c.IOS.CodeSignIdentity)
	c.IOS.ArchivePath = utils.ExpandEnvVars(c.IOS.ArchivePath)
	c.IOS.ExportPath = utils.ExpandEnvVars(c.IOS.ExportPath)

	// Expand App Store Connect fields
	c.AppStoreConnect.AppleID = utils.ExpandEnvVars(c.AppStoreConnect.AppleID)
	c.AppStoreConnect.AppSpecificPassword = utils.ExpandEnvVars(c.AppStoreConnect.AppSpecificPassword)
	c.AppStoreConnect.APIKeyID = utils.ExpandEnvVars(c.AppStoreConnect.APIKeyID)
	c.AppStoreConnect.APIIssuerID = utils.ExpandEnvVars(c.AppStoreConnect.APIIssuerID)
	c.AppStoreConnect.APIKeyPath = utils.ExpandEnvVars(c.AppStoreConnect.APIKeyPath)
	c.AppStoreConnect.TeamID = utils.ExpandEnvVars(c.AppStoreConnect.TeamID)

	// Expand environment variable values
	for key, value := range c.Environment.Variables {
		c.Environment.Variables[key] = utils.ExpandEnvVars(value)
	}
}

// IOSConfigTemplate returns a config template for iOS projects
func IOSConfigTemplate() string {
	return `{
  "project": {
    "name": "",
    "type": "mobile",
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
    "language": "swift",
    "install_command": "pod install",
    "build_command": "",
    "test_command": "",
    "output_dir": "./build"
  },
  "ios": {
    "workspace_path": "",
    "project_path": "",
    "scheme": "",
    "configuration": "Release",
    "sdk": "iphoneos",
    "export_method": "app-store",
    "bundle_id": "",
    "team_id": "",
    "automatic_signing": true,
    "provisioning_profile": "",
    "code_sign_identity": "iPhone Distribution",
    "archive_path": "",
    "export_path": "",
    "upload_to_testflight": false
  },
  "app_store_connect": {
    "apple_id": "",
    "app_specific_password": "",
    "api_key_id": "",
    "api_issuer_id": "",
    "api_key_path": "",
    "team_id": ""
  },
  "environment": {
    "variables": {
      "ENV": "production"
    }
  }
}`
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

package config

import (
	"automateLife/utils"
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
		// Expand path in case it wasn't expanded yet
		expandedPath := utils.ExpandEnvVars(c.Git.SSHKeyPath)
		if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
			return fmt.Errorf("SSH key not found at: %s (expanded from: %s)", expandedPath, c.Git.SSHKeyPath)
		}
	default:
		return fmt.Errorf("git.auth_type must be 'token', 'basic', or 'ssh'")
	}

	// Validate iOS-specific configuration if language is swift or objective-c
	if c.Build.Language == "swift" || c.Build.Language == "objective-c" || c.Build.Language == "objc" {
		if err := c.validateIOSConfig(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) validateIOSConfig() error {
	// Workspace and project paths are optional (will be auto-detected)
	// But if both are provided, that's fine too

	// Check that scheme is provided
	if c.IOS.Scheme == "" {
		return fmt.Errorf("ios.scheme is required for iOS projects")
	}

	// If manual signing is used, require provisioning profile and code sign identity
	if !c.IOS.AutomaticSigning {
		if c.IOS.ProvisioningProfile == "" {
			return fmt.Errorf("ios.provisioning_profile is required when automatic_signing is false")
		}
		if c.IOS.CodeSignIdentity == "" {
			return fmt.Errorf("ios.code_sign_identity is required when automatic_signing is false")
		}
	}

	// Validate configuration
	if c.IOS.Configuration != "" {
		if c.IOS.Configuration != "Debug" && c.IOS.Configuration != "Release" {
			return fmt.Errorf("ios.configuration must be 'Debug' or 'Release'")
		}
	}

	// Validate SDK
	if c.IOS.SDK != "" {
		validSDKs := []string{"iphoneos", "iphonesimulator", "macosx"}
		isValid := false
		for _, sdk := range validSDKs {
			if c.IOS.SDK == sdk {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("ios.sdk must be one of: iphoneos, iphonesimulator, macosx")
		}
	}

	// Validate export method
	if c.IOS.ExportMethod != "" {
		validMethods := []string{"app-store", "ad-hoc", "development", "enterprise"}
		isValid := false
		for _, method := range validMethods {
			if c.IOS.ExportMethod == method {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("ios.export_method must be one of: app-store, ad-hoc, development, enterprise")
		}
	}

	// If upload to TestFlight is enabled, validate App Store Connect credentials
	if c.IOS.UploadToTestFlight {
		if err := c.validateAppStoreConnect(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) validateAppStoreConnect() error {
	// Check for API Key authentication (preferred method)
	hasAPIKey := c.AppStoreConnect.APIKeyID != "" &&
		c.AppStoreConnect.APIIssuerID != "" &&
		c.AppStoreConnect.APIKeyPath != ""

	// Check for Apple ID authentication (legacy method)
	hasAppleID := c.AppStoreConnect.AppleID != "" &&
		c.AppStoreConnect.AppSpecificPassword != ""

	if !hasAPIKey && !hasAppleID {
		return fmt.Errorf("app_store_connect credentials required for TestFlight upload. " +
			"Provide either (api_key_id, api_issuer_id, api_key_path) or (apple_id, app_specific_password)")
	}

	// If using API Key, verify the file exists
	if hasAPIKey {
		expandedPath := utils.ExpandEnvVars(c.AppStoreConnect.APIKeyPath)
		if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
			return fmt.Errorf("App Store Connect API key file not found at: %s", expandedPath)
		}
	}

	return nil
}

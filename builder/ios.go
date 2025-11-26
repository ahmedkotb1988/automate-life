package builder

import (
	"automateLife/config"
	"automateLife/ui"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IOSBuilder handles iOS project building operations
type IOSBuilder struct {
	Config        *config.Config
	workspacePath string
	projectPath   string
}

// NewIOSBuilder creates a new iOS builder instance
func NewIOSBuilder(cfg *config.Config) *IOSBuilder {
	builder := &IOSBuilder{Config: cfg}
	builder.detectPaths()
	return builder
}

// detectPaths auto-detects workspace and project paths if not provided
func (b *IOSBuilder) detectPaths() {
	// Use provided paths if available
	b.workspacePath = b.Config.IOS.WorkspacePath
	b.projectPath = b.Config.IOS.ProjectPath

	// If neither is provided, search for them
	if b.workspacePath == "" && b.projectPath == "" {
		ui.Info("Auto-detecting Xcode workspace/project...")

		// First look for .xcworkspace (preferred)
		matches, _ := filepath.Glob("*.xcworkspace")
		if len(matches) > 0 {
			b.workspacePath = matches[0]
			ui.Success(fmt.Sprintf("Found workspace: %s", b.workspacePath))
			return
		}

		// Fall back to .xcodeproj
		matches, _ = filepath.Glob("*.xcodeproj")
		if len(matches) > 0 {
			b.projectPath = matches[0]
			ui.Success(fmt.Sprintf("Found project: %s", b.projectPath))
			return
		}

		ui.Warning("No workspace or project found in current directory")
	}
}

// Build builds the iOS project using xcodebuild
func (b *IOSBuilder) Build() error {
	ui.Info("Building iOS project...")

	args := []string{"build"}

	// Add workspace or project
	if b.workspacePath != "" {
		args = append(args, "-workspace", b.workspacePath)
	} else if b.projectPath != "" {
		args = append(args, "-project", b.projectPath)
	} else {
		return fmt.Errorf("no workspace or project file found")
	}

	// Add scheme
	args = append(args, "-scheme", b.Config.IOS.Scheme)

	// Add configuration
	if b.Config.IOS.Configuration != "" {
		args = append(args, "-configuration", b.Config.IOS.Configuration)
	}

	// Add SDK
	if b.Config.IOS.SDK != "" {
		args = append(args, "-sdk", b.Config.IOS.SDK)
	}

	// Add code signing settings
	args = append(args, b.getCodeSigningArgs()...)

	ui.Info(fmt.Sprintf("Executing: xcodebuild %s", strings.Join(args, " ")))

	cmd := exec.Command("xcodebuild", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	ui.Success("Build completed successfully")
	return nil
}

// getCodeSigningArgs returns code signing arguments based on automatic vs manual signing
func (b *IOSBuilder) getCodeSigningArgs() []string {
	var args []string

	if b.Config.IOS.AutomaticSigning {
		// Automatic code signing
		args = append(args, "CODE_SIGN_STYLE=Automatic")

		if b.Config.IOS.TeamID != "" {
			args = append(args, "DEVELOPMENT_TEAM="+b.Config.IOS.TeamID)
		}

		if b.Config.IOS.BundleID != "" {
			args = append(args, "PRODUCT_BUNDLE_IDENTIFIER="+b.Config.IOS.BundleID)
		}
	} else {
		// Manual code signing
		args = append(args, "CODE_SIGN_STYLE=Manual")

		if b.Config.IOS.CodeSignIdentity != "" {
			args = append(args, "CODE_SIGN_IDENTITY="+b.Config.IOS.CodeSignIdentity)
		}

		if b.Config.IOS.ProvisioningProfile != "" {
			args = append(args, "PROVISIONING_PROFILE_SPECIFIER="+b.Config.IOS.ProvisioningProfile)
		}

		if b.Config.IOS.TeamID != "" {
			args = append(args, "DEVELOPMENT_TEAM="+b.Config.IOS.TeamID)
		}
	}

	return args
}

// Test runs tests for the iOS project
func (b *IOSBuilder) Test() error {
	ui.Info("Running iOS tests...")

	args := []string{"test"}

	// Add workspace or project
	if b.workspacePath != "" {
		args = append(args, "-workspace", b.workspacePath)
	} else if b.projectPath != "" {
		args = append(args, "-project", b.projectPath)
	} else {
		return fmt.Errorf("no workspace or project file found")
	}

	// Add scheme
	args = append(args, "-scheme", b.Config.IOS.Scheme)

	// Use simulator for tests by default
	sdk := "iphonesimulator"
	if b.Config.IOS.SDK == "iphonesimulator" || b.Config.IOS.SDK == "" {
		args = append(args, "-sdk", sdk)
		args = append(args, "-destination", "platform=iOS Simulator,name=iPhone 15")
	}

	ui.Info(fmt.Sprintf("Executing: xcodebuild %s", strings.Join(args, " ")))

	cmd := exec.Command("xcodebuild", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	ui.Success("All tests passed")
	return nil
}

// Archive creates an archive of the iOS project
func (b *IOSBuilder) Archive() (string, error) {
	ui.Info("Creating iOS archive...")

	archivePath := b.Config.IOS.ArchivePath
	if archivePath == "" {
		// Create archive in a temp directory
		tmpDir := os.TempDir()
		archivePath = filepath.Join(tmpDir, "app.xcarchive")
	}

	// Create directory if it doesn't exist
	archiveDir := filepath.Dir(archivePath)
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create archive directory: %w", err)
	}

	args := []string{"archive"}

	// Add workspace or project
	if b.workspacePath != "" {
		args = append(args, "-workspace", b.workspacePath)
	} else if b.projectPath != "" {
		args = append(args, "-project", b.projectPath)
	} else {
		return "", fmt.Errorf("no workspace or project file found")
	}

	// Add scheme
	args = append(args, "-scheme", b.Config.IOS.Scheme)

	// Add configuration
	if b.Config.IOS.Configuration != "" {
		args = append(args, "-configuration", b.Config.IOS.Configuration)
	}

	// Add SDK
	sdk := b.Config.IOS.SDK
	if sdk == "" {
		sdk = "iphoneos"
	}
	args = append(args, "-sdk", sdk)

	// Add archive path
	args = append(args, "-archivePath", archivePath)

	// Add code signing settings
	args = append(args, b.getCodeSigningArgs()...)

	ui.Info(fmt.Sprintf("Executing: xcodebuild %s", strings.Join(args, " ")))

	cmd := exec.Command("xcodebuild", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("archive failed: %w", err)
	}

	ui.Success(fmt.Sprintf("Archive created at: %s", archivePath))
	return archivePath, nil
}

// ExportIPA exports the archive to an IPA file
func (b *IOSBuilder) ExportIPA(archivePath string) (string, error) {
	ui.Info("Exporting IPA...")

	exportPath := b.Config.IOS.ExportPath
	if exportPath == "" {
		// Default to current working directory
		cwd, err := os.Getwd()
		if err != nil {
			exportPath = "."
		} else {
			exportPath = cwd
		}
		ui.Info(fmt.Sprintf("Export path not specified, using: %s", exportPath))
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(exportPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create export directory: %w", err)
	}

	// Create export options plist
	exportOptionsPlist := filepath.Join(exportPath, "ExportOptions.plist")
	if err := b.createExportOptions(exportOptionsPlist); err != nil {
		return "", fmt.Errorf("failed to create export options: %w", err)
	}

	args := []string{
		"-exportArchive",
		"-archivePath", archivePath,
		"-exportPath", exportPath,
		"-exportOptionsPlist", exportOptionsPlist,
	}

	ui.Info(fmt.Sprintf("Executing: xcodebuild %s", strings.Join(args, " ")))

	cmd := exec.Command("xcodebuild", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("export failed: %w", err)
	}

	// Find the IPA file - try multiple possible names
	var ipaPath string

	// First try with project name
	if b.Config.Project.Name != "" {
		possiblePath := filepath.Join(exportPath, b.Config.Project.Name+".ipa")
		if _, err := os.Stat(possiblePath); err == nil {
			ipaPath = possiblePath
		}
	}

	// If not found, search for any .ipa file in the export directory
	if ipaPath == "" {
		files, err := os.ReadDir(exportPath)
		if err == nil {
			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".ipa") {
					ipaPath = filepath.Join(exportPath, file.Name())
					break
				}
			}
		}
	}

	if ipaPath == "" {
		return "", fmt.Errorf("could not find IPA file in export directory: %s", exportPath)
	}

	ui.Success(fmt.Sprintf("IPA exported to: %s", ipaPath))
	return ipaPath, nil
}

// createExportOptions creates the ExportOptions.plist file
func (b *IOSBuilder) createExportOptions(path string) error {
	exportMethod := b.Config.IOS.ExportMethod
	if exportMethod == "" {
		exportMethod = "app-store"
	}

	teamID := b.Config.IOS.TeamID
	if teamID == "" {
		teamID = b.Config.AppStoreConnect.TeamID
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>method</key>
    <string>%s</string>
    <key>teamID</key>
    <string>%s</string>
    <key>uploadBitcode</key>
    <false/>
    <key>uploadSymbols</key>
    <true/>
    <key>compileBitcode</key>
    <false/>
</dict>
</plist>`, exportMethod, teamID)

	return os.WriteFile(path, []byte(content), 0644)
}

// UploadToTestFlight uploads the IPA to TestFlight
func (b *IOSBuilder) UploadToTestFlight(ipaPath string) error {
	ui.Info("Uploading to TestFlight...")

	var cmd *exec.Cmd

	// Check which authentication method to use
	if b.Config.AppStoreConnect.APIKeyID != "" {
		// Use API Key (preferred method)
		ui.Info("Using App Store Connect API Key authentication")

		args := []string{
			"altool",
			"--upload-app",
			"--type", "ios",
			"--file", ipaPath,
			"--apiKey", b.Config.AppStoreConnect.APIKeyID,
			"--apiIssuer", b.Config.AppStoreConnect.APIIssuerID,
		}

		cmd = exec.Command("xcrun", args...)
	} else if b.Config.AppStoreConnect.AppleID != "" {
		// Use Apple ID (legacy method)
		ui.Info("Using Apple ID authentication")

		args := []string{
			"altool",
			"--upload-app",
			"--type", "ios",
			"--file", ipaPath,
			"--username", b.Config.AppStoreConnect.AppleID,
			"--password", b.Config.AppStoreConnect.AppSpecificPassword,
		}

		cmd = exec.Command("xcrun", args...)
	} else {
		return fmt.Errorf("no valid App Store Connect credentials provided")
	}

	ui.Info(fmt.Sprintf("Executing: xcrun altool --upload-app..."))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("TestFlight upload failed: %w", err)
	}

	ui.Success("Successfully uploaded to TestFlight!")
	return nil
}

// InstallDependencies installs iOS project dependencies
func (b *IOSBuilder) InstallDependencies() error {
	ui.Info("Installing iOS dependencies...")

	// Check for CocoaPods
	if _, err := os.Stat("Podfile"); err == nil {
		ui.Info("Found Podfile, running pod install...")
		cmd := exec.Command("pod", "install")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("pod install failed: %w", err)
		}
		ui.Success("CocoaPods dependencies installed")
		return nil
	}

	// Check for Swift Package Manager (SPM)
	if _, err := os.Stat("Package.swift"); err == nil {
		ui.Info("Found Package.swift, resolving Swift packages...")
		// SPM packages are resolved automatically by xcodebuild
		ui.Success("Swift Package Manager will resolve packages during build")
		return nil
	}

	// Check for Carthage
	if _, err := os.Stat("Cartfile"); err == nil {
		ui.Info("Found Cartfile, running carthage update...")
		cmd := exec.Command("carthage", "update", "--platform", "iOS")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("carthage update failed: %w", err)
		}
		ui.Success("Carthage dependencies installed")
		return nil
	}

	ui.Info("No dependency manager detected, skipping...")
	return nil
}

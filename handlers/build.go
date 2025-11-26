package handlers

import (
	"automateLife/builder"
	"automateLife/config"
	"automateLife/git"
	"automateLife/ui"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func HandleBuild(fileName string) {
	cfg, err := config.Load(fileName)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to load config: %v", err))
		return
	}

	if err := cfg.Validate(); err != nil {
		ui.Error(fmt.Sprintf("Configuration validation failed: %v", err))
		return
	}

	projectDir := git.GetProjectDirName(cfg.Git.RepoUrl)
	if projectDir == "" {
		ui.Error("Could not determine project directory name")
		return
	}

	originalDir, _ := os.Getwd()
	fullProjectPath := projectDir

	// If project directory is relative, make it absolute
	if !filepath.IsAbs(projectDir) {
		fullProjectPath = filepath.Join(originalDir, projectDir)
	}

	if _, err := os.Stat(fullProjectPath); os.IsNotExist(err) {
		ui.Error(fmt.Sprintf("Project directory '%s' not found. Run 'automateLife start' first.", fullProjectPath))
		return
	}

	fmt.Printf("%s%s=== Building %s ===%s\n\n", ui.Bold, ui.Blue, cfg.Project.Name, ui.Reset)
	ui.Info(fmt.Sprintf("Project directory: %s", fullProjectPath))

	if err := os.Chdir(fullProjectPath); err != nil {
		ui.Error(fmt.Sprintf("Could not change to project directory: %v", err))
		return
	}
	defer os.Chdir(originalDir)

	// Set environment variables
	for key, value := range cfg.Environment.Variables {
		os.Setenv(key, value)
	}

	// Check if this is an iOS project
	isIOS := cfg.Build.Language == "swift" || cfg.Build.Language == "objective-c" || cfg.Build.Language == "objc"

	if isIOS {
		handleIOSBuild(cfg)
	} else {
		handleStandardBuild(cfg)
	}
}

func handleIOSBuild(cfg *config.Config) {
	iosBuilder := builder.NewIOSBuilder(cfg)

	// Step 1: Install dependencies
	fmt.Printf("%sStep 1:%s Installing dependencies...\n", ui.Bold, ui.Reset)
	if err := iosBuilder.InstallDependencies(); err != nil {
		ui.Warning(fmt.Sprintf("Dependency installation warning: %v", err))
	} else {
		ui.Success("Dependencies installed successfully\n")
	}

	// Step 2: Build
	fmt.Printf("\n%sStep 2:%s Building iOS project...\n", ui.Bold, ui.Reset)
	if err := iosBuilder.Build(); err != nil {
		ui.Error(fmt.Sprintf("Build failed: %v", err))
		return
	}

	// Step 3: Archive
	fmt.Printf("\n%sStep 3:%s Creating archive...\n", ui.Bold, ui.Reset)
	archivePath, err := iosBuilder.Archive()
	if err != nil {
		ui.Error(fmt.Sprintf("Archive failed: %v", err))
		return
	}

	// Step 4: Export IPA
	fmt.Printf("\n%sStep 4:%s Exporting IPA...\n", ui.Bold, ui.Reset)
	ipaPath, err := iosBuilder.ExportIPA(archivePath)
	if err != nil {
		ui.Error(fmt.Sprintf("Export failed: %v", err))
		return
	}

	fmt.Printf("\n%s%s✓ Build completed successfully!%s\n", ui.Bold, ui.Green, ui.Reset)
	ui.Success(fmt.Sprintf("IPA location: %s", ipaPath))

	// Step 5: Prompt for TestFlight upload (if configured)
	if cfg.IOS.UploadToTestFlight {
		// Check if App Store Connect credentials are provided
		hasCredentials := (cfg.AppStoreConnect.APIKeyID != "" && cfg.AppStoreConnect.APIIssuerID != "") ||
			cfg.AppStoreConnect.AppleID != ""

		if hasCredentials {
			fmt.Print("\nDo you want to upload to TestFlight now? (y/n): ")
			var response string
			fmt.Scanln(&response)

			if strings.ToLower(strings.TrimSpace(response)) == "y" {
				fmt.Printf("\n%sStep 5:%s Uploading to TestFlight...\n", ui.Bold, ui.Reset)
				if err := iosBuilder.UploadToTestFlight(ipaPath); err != nil {
					ui.Error(fmt.Sprintf("TestFlight upload failed: %v", err))
					ui.Info("Build artifacts are still available, you can upload manually")
					return
				}
				fmt.Printf("\n%s%s✓ Successfully uploaded to TestFlight!%s\n", ui.Bold, ui.Green, ui.Reset)
			} else {
				ui.Info("Skipping TestFlight upload")
				fmt.Printf("\nYou can manually upload later using:\n")
				fmt.Printf("  xcrun altool --upload-app --type ios --file %s\n", ipaPath)
			}
		} else {
			ui.Warning("TestFlight upload is enabled but App Store Connect credentials are not configured")
			fmt.Printf("\nYou can manually upload using:\n")
			fmt.Printf("  xcrun altool --upload-app --type ios --file %s\n", ipaPath)
		}
	} else {
		fmt.Printf("\n%s%sNote:%s TestFlight upload prompting is disabled in config\n", ui.Bold, ui.Yellow, ui.Reset)
		ui.Info("To enable TestFlight upload prompt, set 'upload_to_testflight: true' in your config")
		fmt.Printf("\nYou can manually upload using:\n")
		fmt.Printf("  xcrun altool --upload-app --type ios --file %s\n", ipaPath)
	}
}

func handleStandardBuild(cfg *config.Config) {
	// Step 1: Install dependencies
	if cfg.Build.InstallCommand != "" {
		fmt.Printf("%sStep 1:%s Installing dependencies...\n", ui.Bold, ui.Reset)
		if err := builder.RunCommand(cfg.Build.InstallCommand); err != nil {
			ui.Error(fmt.Sprintf("Dependency installation failed: %v", err))
			return
		}
		ui.Success("Dependencies installed successfully\n")
	} else {
		fmt.Printf("%sStep 1:%s Detecting and installing dependencies...\n", ui.Bold, ui.Reset)
		if err := builder.AutoInstallDependencies(cfg.Build.Language); err != nil {
			ui.Warning(fmt.Sprintf("Could not auto-install dependencies: %v", err))
		} else {
			ui.Success("Dependencies installed successfully\n")
		}
	}

	// Step 2: Build
	fmt.Printf("\n%sStep 2:%s Building project...\n", ui.Bold, ui.Reset)
	buildCommand := cfg.Build.BuildCommand
	if buildCommand == "" {
		buildCommand = builder.GetDefaultBuildCommand(cfg.Build.Language)
		ui.Info(fmt.Sprintf("Using default build command for %s: %s", cfg.Build.Language, buildCommand))
	}

	ui.Info(fmt.Sprintf("Executing: %s", buildCommand))

	if err := builder.RunCommand(buildCommand); err != nil {
		fmt.Printf("\n%s%s✗ Build failed!%s\n", ui.Bold, ui.Red, ui.Reset)
		ui.Error(fmt.Sprintf("Error: %v", err))
		return
	}

	fmt.Printf("\n%s%s✓ Build completed successfully!%s\n", ui.Bold, ui.Green, ui.Reset)
}

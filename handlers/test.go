package handlers

import (
	"automateLife/builder"
	"automateLife/config"
	"automateLife/git"
	"automateLife/ui"
	"fmt"
	"os"
	"path/filepath"
)

func HandleTest(fileName string) {
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
		ui.Info(fmt.Sprintf("Current directory: %s", originalDir))
		ui.Info(fmt.Sprintf("Looking for: %s", fullProjectPath))
		return
	}

	fmt.Printf("%s%s=== Running Tests for %s ===%s\n\n", ui.Bold, ui.Blue, cfg.Project.Name, ui.Reset)
	ui.Info(fmt.Sprintf("Project directory: %s", fullProjectPath))

	if err := os.Chdir(fullProjectPath); err != nil {
		ui.Error(fmt.Sprintf("Could not change to project directory: %v", err))
		return
	}
	defer os.Chdir(originalDir)

	currentDir, _ := os.Getwd()
	ui.Info(fmt.Sprintf("Changed to: %s", currentDir))

	// Set environment variables
	for key, value := range cfg.Environment.Variables {
		os.Setenv(key, value)
	}

	// Check if this is an iOS project
	isIOS := cfg.Build.Language == "swift" || cfg.Build.Language == "objective-c" || cfg.Build.Language == "objc"

	if isIOS {
		// Handle iOS testing
		handleIOSTest(cfg, fullProjectPath)
		return
	}

	// Install dependencies
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

	// Step 2: Discover test files
	fmt.Printf("%sStep 2:%s Discovering test files...\n", ui.Bold, ui.Reset)
	testFiles, err := builder.DiscoverTests(currentDir)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to discover tests: %v", err))
		return
	}

	if len(testFiles) == 0 {
		ui.Warning("No test files found in the project")
		return
	}

	// Step 3: Create unified test suite
	fmt.Printf("\n%sStep 3:%s Creating unified test suite...\n", ui.Bold, ui.Reset)
	unifiedDir, err := builder.CreateUnifiedTestSuite(testFiles, currentDir)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to create unified test suite: %v", err))
		return
	}
	defer builder.CleanupUnifiedTestSuite(currentDir)

	// Step 4: Run tests
	fmt.Printf("\n%sStep 4:%s Running tests...\n", ui.Bold, ui.Reset)
	testCommand := cfg.Build.TestCommand
	if testCommand == "" {
		testCommand = builder.GetDefaultTestCommand(cfg.Build.Language)
		ui.Info(fmt.Sprintf("Using default test command for %s: %s", cfg.Build.Language, testCommand))
	}

	// Run tests from the unified directory
	if err := os.Chdir(unifiedDir); err != nil {
		ui.Error(fmt.Sprintf("Failed to change to unified test directory: %v", err))
		return
	}

	ui.Info(fmt.Sprintf("Executing: %s", testCommand))

	if err := builder.RunCommand(testCommand); err != nil {
		fmt.Printf("\n%s%s✗ Tests failed!%s\n", ui.Bold, ui.Red, ui.Reset)
		ui.Info(fmt.Sprintf("Error: %v", err))
		return
	}

	fmt.Printf("\n%s%s✓ All tests passed successfully!%s\n", ui.Bold, ui.Green, ui.Reset)
}

func handleIOSTest(cfg *config.Config, projectPath string) {
	fmt.Printf("%s%s=== Running iOS Tests for %s ===%s\n\n", ui.Bold, ui.Blue, cfg.Project.Name, ui.Reset)

	// Change to project directory
	if err := os.Chdir(projectPath); err != nil {
		ui.Error(fmt.Sprintf("Could not change to project directory: %v", err))
		return
	}

	iosBuilder := builder.NewIOSBuilder(cfg)

	// Step 1: Install dependencies
	fmt.Printf("%sStep 1:%s Installing dependencies...\n", ui.Bold, ui.Reset)
	if err := iosBuilder.InstallDependencies(); err != nil {
		ui.Warning(fmt.Sprintf("Dependency installation warning: %v", err))
	} else {
		ui.Success("Dependencies installed successfully\n")
	}

	// Step 2: Run tests
	fmt.Printf("\n%sStep 2:%s Running iOS tests...\n", ui.Bold, ui.Reset)
	if err := iosBuilder.Test(); err != nil {
		fmt.Printf("\n%s%s✗ Tests failed!%s\n", ui.Bold, ui.Red, ui.Reset)
		ui.Error(fmt.Sprintf("Error: %v", err))
		return
	}

	fmt.Printf("\n%s%s✓ All iOS tests passed successfully!%s\n", ui.Bold, ui.Green, ui.Reset)
}

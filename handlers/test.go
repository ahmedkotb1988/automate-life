package handlers

import (
	"automateLife/builder"
	"automateLife/config"
	"automateLife/git"
	"automateLife/ui"
	"fmt"
	"os"
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

	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		ui.Error(fmt.Sprintf("Project directory '%s' not found. Run 'automateLife start' first.", projectDir))
		return
	}

	fmt.Printf("%s%s=== Running Tests for %s ===%s\n\n", ui.Bold, ui.Blue, cfg.Project.Name, ui.Reset)

	originalDir, _ := os.Getwd()
	if err := os.Chdir(projectDir); err != nil {
		ui.Error(fmt.Sprintf("Could not change to project directory: %v", err))
		return
	}
	defer os.Chdir(originalDir)

	// Set environment variables
	for key, value := range cfg.Environment.Variables {
		os.Setenv(key, value)
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

	// Run tests
	fmt.Printf("%sStep 2:%s Running tests...\n", ui.Bold, ui.Reset)
	testCommand := cfg.Build.TestCommand
	if testCommand == "" {
		testCommand = builder.GetDefaultTestCommand(cfg.Build.Language)
		ui.Info(fmt.Sprintf("Using default test command for %s: %s", cfg.Build.Language, testCommand))
	}

	if err := builder.RunCommand(testCommand); err != nil {
		fmt.Printf("\n%s%s✗ Tests failed!%s\n", ui.Bold, ui.Red, ui.Reset)
		return
	}

	fmt.Printf("\n%s%s✓ All tests passed successfully!%s\n", ui.Bold, ui.Green, ui.Reset)
}

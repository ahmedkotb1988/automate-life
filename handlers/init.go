package handlers

import (
	"automateLife/config"
	"automateLife/ui"
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
)

func HandleInit(fileName string) {
	content := config.DefaultConfigTemplate()

	if err := config.Create(fileName, content); err != nil {
		if err.Error() == "config file already exists" {
			fmt.Println(fileName + " already exists in your current directory")
		} else {
			ui.Error(fmt.Sprintf("Failed to create %s: %v", fileName, err))
		}
		return
	}

	ui.Success(fileName + " created successfully")
	fmt.Println("Do you wish to populate the config file? y/n")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "n":
		fmt.Println("Population process aborted, please populate the config file then run 'automatelife start'")
		return
	case "y":
		fmt.Println("Populating .....")
		if err := populateConfigInteractively(fileName); err != nil {
			ui.Error(fmt.Sprintf("Failed to populate config: %v", err))
			return
		}
		ui.Success("Config file populated successfully!")

		// Ask if user wants to start immediately
		fmt.Print("\nDo you want to start cloning the repository now? y/n\n")
		startReader := bufio.NewReader(os.Stdin)
		startInput, _ := startReader.ReadString('\n')
		startInput = strings.TrimSpace(startInput)

		if startInput == "y" {
			fmt.Print("\nStarting repository clone...\n\n")
			HandleStart(fileName)
		} else {
			fmt.Println("You can run 'automateLife start' later to begin cloning the repository")
		}
	default:
		fmt.Println("Invalid choice " + input)
		return
	}
}

func populateConfigInteractively(fileName string) error {
	// Load the existing config
	cfg, err := config.Load(fileName)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// 1. Select Git Provider
	providerPrompt := promptui.Select{
		Label: "Select Git Provider",
		Items: []string{"github", "gitlab", "bitbucket", "azure-devops"},
	}
	_, provider, err := providerPrompt.Run()
	if err != nil {
		return fmt.Errorf("provider selection failed: %w", err)
	}
	cfg.Git.Provider = provider

	// 2. Select Authentication Type
	authPrompt := promptui.Select{
		Label: "Select Git Authentication Type",
		Items: []string{"token", "password", "ssh"},
	}
	_, authType, err := authPrompt.Run()
	if err != nil {
		return fmt.Errorf("authentication selection failed: %w", err)
	}
	cfg.Git.AuthType = authType

	// 3. Select Language
	langPrompt := promptui.Select{
		Label: "Select Project Language",
		Items: []string{"go", "dotnet", "python", "nodejs", "java"},
	}
	_, language, err := langPrompt.Run()
	if err != nil {
		return fmt.Errorf("language selection failed: %w", err)
	}
	cfg.Build.Language = language

	// 4. Select Project Type
	projectTypePrompt := promptui.Select{
		Label: "Select Project Type",
		Items: []string{"backend", "frontend", "fullstack", "cli", "library"},
	}
	_, projectType, err := projectTypePrompt.Run()
	if err != nil {
		return fmt.Errorf("project type selection failed: %w", err)
	}
	cfg.Project.Type = projectType

	// 5. Select Deployment Type (only if using Azure DevOps)
	if provider == "azure-devops" {
		deploymentPrompt := promptui.Select{
			Label: "Select Azure Deployment Type",
			Items: []string{"webapp", "container", "function"},
		}
		_, deploymentType, err := deploymentPrompt.Run()
		if err != nil {
			return fmt.Errorf("deployment type selection failed: %w", err)
		}
		cfg.Azure.DeploymentType = deploymentType
	}

	// Now collect crucial inputs based on selections
	fmt.Println("\nPlease provide the following information:")

	// Project Name
	projectNamePrompt := promptui.Prompt{
		Label: "Project Name",
	}
	projectName, err := projectNamePrompt.Run()
	if err != nil {
		return fmt.Errorf("project name input failed: %w", err)
	}
	cfg.Project.Name = strings.TrimSpace(projectName)

	// Project Description
	projectDescPrompt := promptui.Prompt{
		Label: "Project Description (optional)",
	}
	projectDesc, _ := projectDescPrompt.Run()
	cfg.Project.Description = strings.TrimSpace(projectDesc)

	// Git Repository URL
	repoUrlPrompt := promptui.Prompt{
		Label: "Git Repository URL",
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("repository URL is required")
			}
			return nil
		},
	}
	repoUrl, err := repoUrlPrompt.Run()
	if err != nil {
		return fmt.Errorf("repository URL input failed: %w", err)
	}
	cfg.Git.RepoUrl = strings.TrimSpace(repoUrl)

	// Git Branch
	branchPrompt := promptui.Prompt{
		Label:   "Git Branch",
		Default: "main",
	}
	branch, err := branchPrompt.Run()
	if err != nil {
		return fmt.Errorf("branch input failed: %w", err)
	}
	cfg.Git.Branch = strings.TrimSpace(branch)

	// Authentication-specific fields
	switch authType {
	case "token":
		tokenPrompt := promptui.Prompt{
			Label: "Git Token",
			Mask:  '*',
			Validate: func(input string) error {
				if strings.TrimSpace(input) == "" {
					return fmt.Errorf("token is required")
				}
				return nil
			},
		}
		token, err := tokenPrompt.Run()
		if err != nil {
			return fmt.Errorf("token input failed: %w", err)
		}
		cfg.Git.Token = strings.TrimSpace(token)
		cfg.Git.Password = ""
		cfg.Git.UserName = ""
		cfg.Git.SSHKeyPath = ""

	case "password":
		usernamePrompt := promptui.Prompt{
			Label: "Git Username",
			Validate: func(input string) error {
				if strings.TrimSpace(input) == "" {
					return fmt.Errorf("username is required")
				}
				return nil
			},
		}
		username, err := usernamePrompt.Run()
		if err != nil {
			return fmt.Errorf("username input failed: %w", err)
		}
		cfg.Git.UserName = strings.TrimSpace(username)

		passwordPrompt := promptui.Prompt{
			Label: "Git Password",
			Mask:  '*',
			Validate: func(input string) error {
				if strings.TrimSpace(input) == "" {
					return fmt.Errorf("password is required")
				}
				return nil
			},
		}
		password, err := passwordPrompt.Run()
		if err != nil {
			return fmt.Errorf("password input failed: %w", err)
		}
		cfg.Git.Password = strings.TrimSpace(password)
		cfg.Git.Token = ""
		cfg.Git.SSHKeyPath = ""

	case "ssh":
		sshKeyPrompt := promptui.Prompt{
			Label:   "SSH Key Path",
			Default: "~/.ssh/id_rsa",
			Validate: func(input string) error {
				if strings.TrimSpace(input) == "" {
					return fmt.Errorf("SSH key path is required")
				}
				return nil
			},
		}
		sshKeyPath, err := sshKeyPrompt.Run()
		if err != nil {
			return fmt.Errorf("SSH key path input failed: %w", err)
		}
		cfg.Git.SSHKeyPath = strings.TrimSpace(sshKeyPath)
		cfg.Git.Token = ""
		cfg.Git.Password = ""
		cfg.Git.UserName = ""
	}

	// Build commands (optional, can be auto-detected later)
	buildCmdPrompt := promptui.Prompt{
		Label: fmt.Sprintf("Build Command for %s (optional)", language),
	}
	buildCmd, _ := buildCmdPrompt.Run()
	cfg.Build.BuildCommand = strings.TrimSpace(buildCmd)

	testCmdPrompt := promptui.Prompt{
		Label: fmt.Sprintf("Test Command for %s (optional)", language),
	}
	testCmd, _ := testCmdPrompt.Run()
	cfg.Build.TestCommand = strings.TrimSpace(testCmd)

	// Azure Configuration (only if using Azure DevOps)
	if provider == "azure-devops" {
		fmt.Println("\nAzure Configuration:")

		azureAppNamePrompt := promptui.Prompt{
			Label: "Azure App Name",
		}
		azureAppName, err := azureAppNamePrompt.Run()
		if err != nil {
			return fmt.Errorf("Azure app name input failed: %w", err)
		}
		cfg.Azure.AppName = strings.TrimSpace(azureAppName)

		azureResourceGroupPrompt := promptui.Prompt{
			Label: "Azure Resource Group",
		}
		azureResourceGroup, err := azureResourceGroupPrompt.Run()
		if err != nil {
			return fmt.Errorf("Azure resource group input failed: %w", err)
		}
		cfg.Azure.ResourceGroup = strings.TrimSpace(azureResourceGroup)

		azureSubscriptionPrompt := promptui.Prompt{
			Label: "Azure Subscription ID",
		}
		azureSubscription, err := azureSubscriptionPrompt.Run()
		if err != nil {
			return fmt.Errorf("Azure subscription input failed: %w", err)
		}
		cfg.Azure.SubscriptionID = strings.TrimSpace(azureSubscription)

		azureRegionPrompt := promptui.Prompt{
			Label:   "Azure Region",
			Default: "eastus",
		}
		azureRegion, err := azureRegionPrompt.Run()
		if err != nil {
			return fmt.Errorf("Azure region input failed: %w", err)
		}
		cfg.Azure.Region = strings.TrimSpace(azureRegion)
	}

	// Save the updated config
	return saveConfig(fileName, cfg)
}

func saveConfig(fileName string, cfg *config.Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(fileName, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

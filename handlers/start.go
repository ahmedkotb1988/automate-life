package handlers

import (
	"automateLife/config"
	"automateLife/git"
	"automateLife/ui"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func HandleStart(fileName string) {
	cfg, err := config.Load(fileName)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to load config: %v", err))
		fmt.Println("Please run 'automateLife init' to create a config file")
		return
	}

	if cfg.Git.RepoUrl == "" {
		ui.Error("repo_url cannot be empty")
		return
	}

	// Handle SSH authentication
	if cfg.Git.AuthType == "ssh" {
		if err := git.SetupSSH(cfg.Git.SSHKeyPath); err != nil {
			ui.Error(err.Error())
			return
		}
		ui.Info(fmt.Sprintf("Using SSH authentication with key: %s", cfg.Git.SSHKeyPath))
	}

	repoUrl, err := git.BuildAuthURL(&cfg.Git)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to build repo URL: %v", err))
		return
	}

	authHeader, err := git.GetAuthHeader(&cfg.Git)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to build auth header: %v", err))
		return
	}

	// Get project directory name
	projectDir := git.GetProjectDirName(cfg.Git.RepoUrl)
	if projectDir == "" {
		ui.Error("Could not determine project directory name from repo URL")
		return
	}

	// Check if repository is already cloned
	if isRepoAlreadyCloned(projectDir) {
		ui.Info(fmt.Sprintf("Repository already cloned in directory: %s", projectDir))
		ui.Success("Skipping clone step")

		// Optionally update the repository
		fmt.Print("\nDo you want to pull latest changes? (y/n): ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "y" {
			if err := pullLatestChanges(projectDir, cfg.Git.Branch); err != nil {
				ui.Warning(fmt.Sprintf("Failed to pull latest changes: %v", err))
				ui.Info("Continuing with existing code...")
			} else {
				ui.Success("Repository updated successfully!")
			}
		}
	} else {
		// Disable Git credential helper
		os.Setenv("GIT_TERMINAL_PROMPT", "0")
		os.Setenv("GCM_INTERACTIVE", "never")

		// Prepare git clone command
		var cmd *exec.Cmd
		var args []string

		// Add auth header if present (for token/basic auth)
		if authHeader != "" {
			args = append(args, "-c", authHeader)
		}

		args = append(args, "clone")
		if cfg.Git.Branch != "" && cfg.Git.Branch != "main" {
			args = append(args, "-b", cfg.Git.Branch)
			fmt.Printf("Cloning repository (branch: %s%s%s%s) .....\n", ui.Bold, cfg.Git.Branch, ui.Reset, "")
		} else {
			fmt.Println("Cloning repository .....")
		}

		args = append(args, repoUrl)

		cmd = exec.Command("git", args...)

		cmd.Env = append(os.Environ(),
			"GIT_TERMINAL_PROMPT=0",
			"GCM_INTERACTIVE=never",
			"GIT_ASKPASS=echo",
		)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err = cmd.Run(); err != nil {
			ui.Error(fmt.Sprintf("Cloning repo failed: %v", err))
			fmt.Println("\nTroubleshooting tips:")
			fmt.Println("  1. Verify your PAT has the correct permissions (Code: Read)")
			fmt.Println("  2. Check if the PAT has expired")
			fmt.Println("  3. Ensure the repo_url is correct")
			return
		}

		ui.Success("Repo cloned successfully!")
	}

	// Ask what user wants to do next
	fmt.Println("\nWhat would you like to do next?")
	fmt.Println("  1. Run tests")
	fmt.Println("  2. Build project")
	fmt.Println("  3. Build and test")
	fmt.Println("  4. Skip (do nothing)")
	fmt.Print("\nEnter your choice (1-4): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "1":
		fmt.Print("\nStarting tests...\n\n")
		HandleTest(fileName)
	case "2":
		fmt.Print("\nStarting build...\n\n")
		HandleBuild(fileName)
	case "3":
		fmt.Print("\nRunning tests first...\n\n")
		HandleTest(fileName)
		fmt.Print("\nTests completed. Starting build...\n\n")
		HandleBuild(fileName)
	case "4":
		if cfg.Project.Name != "" {
			fmt.Println("\nNext steps:")
			fmt.Println("  cd into your project directory")
			fmt.Printf("  Run %s%sautomateLife test%s to run tests\n", ui.Bold, ui.Blue, ui.Reset)
			fmt.Printf("  Run %s%sautomateLife build%s to build the project\n", ui.Bold, ui.Blue, ui.Reset)
		}
	default:
		fmt.Println("Invalid choice, skipping...")
		if cfg.Project.Name != "" {
			fmt.Println("\nNext steps:")
			fmt.Println("  cd into your project directory")
			fmt.Printf("  Run %s%sautomateLife test%s to run tests\n", ui.Bold, ui.Blue, ui.Reset)
			fmt.Printf("  Run %s%sautomateLife build%s to build the project\n", ui.Bold, ui.Blue, ui.Reset)
		}
	}
}

// isRepoAlreadyCloned checks if a directory exists and contains a .git folder
func isRepoAlreadyCloned(projectDir string) bool {
	// Check if directory exists
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return false
	}

	// Check if it's a git repository
	gitDir := projectDir + "/.git"
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return false
	}

	return true
}

// pullLatestChanges updates the repository with the latest changes from remote
func pullLatestChanges(projectDir string, branch string) error {
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(projectDir); err != nil {
		return fmt.Errorf("failed to change to project directory: %w", err)
	}

	ui.Info("Pulling latest changes...")

	// Pull from the specified branch or current branch
	args := []string{"pull"}
	if branch != "" {
		args = append(args, "origin", branch)
	}

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}

	return nil
}

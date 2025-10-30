package handlers

import (
	"automateLife/config"
	"automateLife/git"
	"automateLife/ui"
	"fmt"
	"os"
	"os/exec"
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

	if cfg.Project.Name != "" {
		fmt.Println("\nNext steps:")
		fmt.Println("  cd into your project directory")
		fmt.Printf("  Run %s%sautomateLife test%s to run tests\n", ui.Bold, ui.Blue, ui.Reset)
	}
}

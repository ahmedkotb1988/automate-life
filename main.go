package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// ANSI escape codes
const bold string = "\033[1m"

const red string = "\033[31m"
const green string = "\033[32m"

// yellow := "\033[33m"
const blue string = "\033[34m"

// magenta := "\033[35m"
// cyan := "\033[36m"
const reset string = "\033[0m"

type ConfigJSON struct {
	Git         GitConfig         `json:"git"`
	Project     ProjectConfig     `json:"project"`
	Build       BuildConfig       `json:"build"`
	Azure       AzureConfig       `json:"azure"`
	Environment EnvironmentConfig `json:"environment"`
}

type GitConfig struct {
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

func main() {

	fileName := "ConfigFile.json"

	content := `{
  "project": {
    "name": "",
    "type": "backend",
    "description": ""
  },
  "git": {
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

	args := os.Args

	if len(args) != 1 && args[1] == "init" {
		handleInit(fileName, content)
	} else if args[1] == "start" {
		handleStart(fileName)
	} else if args[1] == "verify" {
		_, err := handleVerify(fileName)

		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("%s%sDirectory verified successfully and ready for automation. Run 'automateLife start' to automate!%s\n", bold, green, reset)
	} else if args[1] == "test" {
		handleTest(fileName)
	} else {
		fmt.Print(`
    _         _                        _         _     _  __      
   / \  _   _| |_ ___  _ __ ___   __ _| |_ ___  | |   (_)/ _| ___ 
  / _ \| | | | __/ _ \| '_ ' _ \ / _' | __/ _ \ | |   | | |_ / _ \
 / ___ \ |_| | || (_) | | | | | | (_| | ||  __/ | |___| |  _|  __/
/_/   \_\__,_|\__\___/|_| |_| |_|\__,_|\__\___| |_____|_|_|  \___|
                                                                    
`)

		fmt.Printf("Welcome to %s%sAutomate Life%s, your gate way to automation\n\n", bold, green, reset)

		fmt.Printf("Run %s%sautomate%s then one of the following commands to start:\n\n", bold, blue, reset)
		fmt.Println("init: creates a config file in your current directory")
		fmt.Println("start: starts the automation process using the created config file")
		fmt.Println("verify: verifies that the current directory has the necessary parameters for automation")
		fmt.Println("test: runs the tests deployed in your project")
	}

}

func handleInit(fileName string, content string) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)

	if err != nil {
		if os.IsExist(err) {
			fmt.Println(fileName + " Already Exists in your current directory")
			return
		} else {
			log.Fatalf("Failed to create "+fileName+" In your current directory, please make sure you have the right permissions: %v", err)
			return
		}
	} else {
		fmt.Println(fileName + " created successfully")
	}

	if file != nil {
		defer file.Close()
		_, err := file.WriteString(content)
		if err != nil {
			log.Fatalf("failed to write to file: %v", err.Error())
			return
		}
	}
}

func handleStart(fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(fileName + " doesn't exist in your current directory, please run 'automateLife init' to create it, or make sure you're in the right directory")
		return
	}
	defer file.Close()

	var config ConfigJSON

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&config)

	if err != nil {
		fmt.Println("Unable to decode " + fileName + " Please make sure the file is created and in proper format")
		return
	}

	if config.Git.RepoUrl == "" {
		fmt.Printf("%s%sError:%s repo_url cannot be empty\n", bold, red, reset)
		return
	}

	if config.Git.AuthType == "ssh" {
		if err := setupSSHKey(config.Git.SSHKeyPath); err != nil {
			fmt.Printf("%s%sError:%s %v\n", bold, red, reset, err)
			return
		}
		fmt.Printf("%sInfo:%s Using SSH authentication with key: %s\n", bold+blue, reset, config.Git.SSHKeyPath)
	}

	authUrl, err := buildAuthenticatedURL(&config.Git)

	if err != nil {
		fmt.Printf("%sError:%s Failed to build authenticated URL: %v\n", bold+red, reset, err)
		return
	}
	// authUrl := fmt.Sprintf("http://%s:@azure.adek.gov.ae/PortalCollection/Rayah%%20Enhanced/_git/Rayah%%20iOS", url.QueryEscape("503099"))
	// fmt.Printf("----------URL-------------: \n%s\n", authUrl)
	// cmd := exec.Command("git", "clone", authUrl)

	var cmd *exec.Cmd

	if config.Git.Branch != "" && config.Git.Branch != "main" {
		cmd = exec.Command("git", "clone", "-b", config.Git.Branch, authUrl)
		fmt.Printf("Cloning repository (branch:%s%s%s) ......\n", bold, config.Git.Branch, reset)
	} else {
		cmd = exec.Command("git", "clone", authUrl)
		fmt.Println("Cloning Repository .....")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err = cmd.Run(); err != nil {
		fmt.Printf("%sError:%s Cloning repo failed: %v\n ", bold+red, reset, err)
		return
	}

	fmt.Printf("\n%s%sRepo cloned successfully!%s\n", bold, green, reset)

	if config.Project.Name != "" {
		fmt.Printf("\nNext Steps:\n")
		fmt.Printf(" cd into your project directory\n")
		fmt.Printf(" Run %s%sautomateLife test%s to run tests\n", bold, blue, reset)
	}
}

func handleVerify(fileName string) (*ConfigJSON, error) {

	file, err := os.Open(fileName)

	if err != nil {
		fmt.Printf("%s was not found in the working directory. Please make sure to run 'automateLife init' to create a new config file, or make sure you are in the right directory\n", fileName)
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var config ConfigJSON
	err = decoder.Decode(&config)

	if err != nil {
		fmt.Printf("%s does not have the necessary configurations for the process to start, please make sure the file is properly populated\n", fileName)
		return nil, err
	}

	if config.Git.RepoUrl == "" {
		return nil, fmt.Errorf("git.repo_url Cannot be empty")
	}

	if config.Project.Type == "" {
		return nil, fmt.Errorf("project.type is required")
	}

	switch config.Git.AuthType {
	case "token":
		if config.Git.Token == "" {
			return nil, fmt.Errorf("git.token is required when auth_type is 'token'")
		}
	case "basic":
		if config.Git.UserName == "" || config.Git.Password == "" {
			return nil, fmt.Errorf("git.username and git.password are required when auth_type is 'basic'")
		}
	case "ssh":
		if config.Git.SSHKeyPath == "" {
			return nil, fmt.Errorf("git.ssh_key_path is required when auth_type is 'ssh'")
		}

		if _, err := os.Stat(config.Git.SSHKeyPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("SSH key not found at: %s", config.Git.SSHKeyPath)
		}
	default:
		return nil, fmt.Errorf("git.auth_type must be 'token', 'basic', or 'ssh'")
	}

	return &config, nil
}

func buildAuthenticatedURL(config *GitConfig) (string, error) {
	switch config.AuthType {
	case "token":
		return buildTokenURL(config)
	case "basic":
		return buildBasicAuthURL(config)
	case "ssh":
		return config.RepoUrl, nil
	default:
		return "", fmt.Errorf("unsupported auth type: %s . Use 'token', 'basic' or 'ssh'", config.AuthType)
	}
}

func buildTokenURL(config *GitConfig) (string, error) {
	if config.Token == "" {
		return "", fmt.Errorf("token is required when auth_type is 'token'")
	}

	if len(config.RepoUrl) == 0 {
		return "", fmt.Errorf("repo_url is empty")
	}
	if strings.HasPrefix(config.RepoUrl, "http://") {
		return strings.Replace(config.RepoUrl, "http://", fmt.Sprintf("http://%s@", config.Token), 1), nil
	} else if strings.HasPrefix(config.RepoUrl, "https://") {
		return strings.Replace(config.RepoUrl, "https://", fmt.Sprintf("https://%s@", config.Token), 1), nil
	}

	return "", fmt.Errorf("repo_url must start with http:// or https://")
}

func buildBasicAuthURL(config *GitConfig) (string, error) {
	if config.UserName == "" || config.Password == "" {
		return "", fmt.Errorf("username and password must not be empty when auth_type is 'basic'")
	}

	if strings.HasPrefix(config.RepoUrl, "http://") {
		credentials := fmt.Sprintf("%s:%s", config.UserName, config.Password)
		return strings.Replace(config.RepoUrl, "http://", fmt.Sprintf("http://%s@", credentials), 1), nil
	} else if strings.HasPrefix(config.RepoUrl, "https://") {
		credentials := fmt.Sprintf("%s:%s", config.UserName, config.Password)
		return strings.Replace(config.RepoUrl, "https://", fmt.Sprintf("https://%s@", credentials), 1), nil
	}

	return "", fmt.Errorf("repo_url must start with http:// or https://")
}

func setupSSHKey(keyPath string) error {
	if keyPath == "" {
		return fmt.Errorf("ssh_key_path must be provided when auth_type is 'ssh'")
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return fmt.Errorf("SSH key not found at %s", keyPath)
	}

	sshCommand := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no", keyPath)
	os.Setenv("GIT_SSH_COMMAND", sshCommand)

	return nil
}

func handleTest(fileName string) {
	config, err := handleVerify(fileName)
	if err != nil {
		fmt.Printf("%sError%s: Configuration validation failed: %v\n", bold+red, reset, err)
		return
	}

	projectDir := getProjectDirectoryName(config.Git.RepoUrl)
	if projectDir == "" {
		fmt.Printf("%sError:%s Could not determine project directory name\n", bold+red, reset)
		return
	}

	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		fmt.Printf("%sError:%s Project directory '%s' not found. Run 'automateLife start' first to clone the repository.\n", bold+red, reset, projectDir)
		return
	}

	fmt.Printf("%s%s=== Running Tests for %s ===%s\n\n", bold, blue, config.Project.Name, reset)

	originalDir, err := os.Getwd()

	if err != nil {
		fmt.Printf("%sError:%s Could not get current directory: %v\n", bold+red, reset, err)
		return
	}

	if err := os.Chdir(projectDir); err != nil {
		fmt.Printf("%sError:%s Could not change to project directory: %v\n", bold+red, reset, err)
		return
	}
	defer os.Chdir(originalDir)

	for key, value := range config.Environment.Variables {
		os.Setenv(key, value)
	}

	if config.Build.InstallCommand != "" {
		fmt.Printf("%sStep 1:%s Installing dependencies...\n", bold, reset)

		if err := runCommand(config.Build.InstallCommand); err != nil {
			fmt.Printf("%sError:%s Dependency installation failed: %v\n", bold+red, reset, err)
			return
		}
		fmt.Printf("%s%s✓ Dependencies installed successfully%s\n\n", bold, green, reset)
	} else {

		fmt.Printf("%sStep 1:%s Detecting and installing dependencies...\n", bold, reset)
		if err := autoInstallDependencies(config.Build.Language); err != nil {
			fmt.Printf("%sWarning:%s Could not auto-install dependencies: %v\n", bold+red, reset, err)
		} else {
			fmt.Printf("%s%s✓ Dependencies installed successfully%s\n\n", bold, green, reset)
		}
	}

	fmt.Printf("%sStep 2:%s Running tests...\n", bold, reset)
	testCommand := config.Build.TestCommand

	if testCommand == "" {
		testCommand = getDefaultTestCommand(config.Build.Language)
		fmt.Printf("%sInfo:%s Using default test command for %s: %s\n", bold+blue, reset, config.Build.Language, testCommand)
	}

	if err := runCommand(testCommand); err != nil {
		fmt.Printf("\n%s%s Tests failed!%s\n", bold, "\033[31m", reset)
		return
	}

	fmt.Printf("\n%s%s All tests passed successfully!%s\n", bold, green, reset)
}

func getProjectDirectoryName(repoUrl string) string {

	repoUrl = strings.TrimSuffix(repoUrl, ".git")
	parts := strings.Split(repoUrl, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func runCommand(command string) error {

	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func autoInstallDependencies(language string) error {
	switch strings.ToLower(language) {
	case "go", "golang":
		if _, err := os.Stat("go.mod"); err != nil {
			return runCommand("go mod download")
		}
	case "node", "nodejs", "javascript", "typescript":

		if _, err := os.Stat("package.json"); err != nil {
			if _, err := os.Stat("yarn.lock"); err != nil {
				return runCommand("yarn install")
			}
			return runCommand("npm install")
		}
	case "python":
		if _, err := os.Stat("requirements.txt"); err != nil {
			return runCommand("pip install -r requirements.txt")
		}
		if _, err := os.Stat("Pipfile"); err != nil {
			return runCommand("pipenv install")
		}
	case "dotnet", "c#", "csharp":
		return runCommand("dotnet restore")
	case "rust":
		return runCommand("cargo fetch")
	case "ruby":
		if _, err := os.Stat("Gemfile"); err != nil {
			return runCommand("bundle install")
		}
	}
	return fmt.Errorf("could not determine how to install dependencies for language: %s", language)
}

func getDefaultTestCommand(language string) string {
	switch strings.ToLower(language) {
	case "go", "golang":
		return "go test"
	case "node", "nodejs", "javascript", "typescript":
		return "npm test"
	case "python":
		return "pytest"
	case "dotnet", "c#", "csharp":
		return "dotnet test"
	case "rust":
		return "cargo test"
	case "ruby":
		return "bundle exec rspec"
	case "java":
		return "mvn test"
	default:
		return "echo 'No default test command for language: " + language + "'"
	}
}

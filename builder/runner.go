package builder

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func RunCommand(command string) error {

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

func AutoInstallDependencies(language string) error {
	switch strings.ToLower(language) {
	case "go", "golang":
		if _, err := os.Stat("go.mod"); err == nil {
			return RunCommand("go mod download")
		}
		return nil // No go.mod, skip dependency installation
	case "node", "nodejs", "javascript", "typescript":
		if _, err := os.Stat("package.json"); err == nil {
			if _, err := os.Stat("yarn.lock"); err == nil {
				return RunCommand("yarn install")
			}
			return RunCommand("npm install")
		}
		return nil // No package.json, skip dependency installation
	case "python":
		if _, err := os.Stat("requirements.txt"); err == nil {
			return RunCommand("pip install -r requirements.txt")
		}
		if _, err := os.Stat("Pipfile"); err == nil {
			return RunCommand("pipenv install")
		}
		return nil // No requirements file, skip dependency installation
	case "dotnet", "c#", "csharp":
		return RunCommand("dotnet restore")
	case "rust":
		return RunCommand("cargo fetch")
	case "ruby":
		if _, err := os.Stat("Gemfile"); err == nil {
			return RunCommand("bundle install")
		}
		return nil // No Gemfile, skip dependency installation
	}
	return fmt.Errorf("could not determine how to install dependencies for language: %s", language)
}

func GetDefaultTestCommand(language string) string {
	switch strings.ToLower(language) {
	case "go", "golang":
		return "go test ./..." // Test all packages recursively
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

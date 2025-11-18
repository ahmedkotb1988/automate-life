# AutomateLife

A powerful CLI tool for automating Git repository cloning, testing, CI/CD pipeline setup, and deployment to cloud platforms.

## Features

- 🔧 **Interactive Configuration**: User-friendly CLI prompts for project setup
- 🔐 **Multiple Auth Methods**: Support for Token, SSH, and Basic authentication
- 🌐 **Multi-Provider Support**: Works with GitHub, GitLab, Bitbucket, and Azure DevOps
- 🧪 **Automated Testing**: Auto-detection and execution of tests for multiple languages
- 🚀 **CI/CD Ready**: Automatic pipeline configuration
- ☁️ **Cloud Deployment**: Azure deployment support (more providers coming soon)
- 📦 **Dependency Management**: Automatic dependency detection and installation
- 🔄 **Path Expansion**: Smart handling of `~` and `$HOME` in paths

## Supported Languages

- Go
- Python
- Node.js / JavaScript / TypeScript
- .NET / C#
- Java
- Rust
- Ruby

## Installation

### Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd Automation

# Build the binary
go build -o automateLife

# Optional: Move to PATH
sudo mv automateLife /usr/local/bin/
```

### Requirements

- Go 1.25.3 or higher
- Git installed and configured

## Quick Start

### 1. Initialize Configuration

```bash
automateLife init
```

This will:
- Create a `ConfigFile.json` in your current directory
- Guide you through interactive prompts to configure:
  - Git provider (GitHub, GitLab, etc.)
  - Authentication method (Token, SSH, Password)
  - Project details (name, type, language)
  - Build and test commands
  - Azure deployment settings (if using Azure DevOps)

### 2. Start Cloning

```bash
automateLife start
```

This will:
- Clone your repository using the configured authentication
- Optionally run tests immediately after cloning

### 3. Run Tests

```bash
automateLife test
```

This will:
- Install dependencies (auto-detected or custom commands)
- Run tests using language-specific defaults or custom commands

## Configuration

### Configuration File Structure

```json
{
  "project": {
    "name": "MyProject",
    "type": "backend",
    "description": "My awesome project"
  },
  "git": {
    "provider": "github",
    "repo_url": "https://github.com/user/repo.git",
    "branch": "main",
    "auth_type": "token",
    "token": "ghp_yourtoken",
    "ssh_key_path": "~/.ssh/id_rsa"
  },
  "build": {
    "language": "go",
    "install_command": "go mod download",
    "build_command": "go build",
    "test_command": "go test ./...",
    "output_dir": "./bin"
  },
  "azure": {
    "subscription_id": "your-subscription-id",
    "resource_group": "your-resource-group",
    "app_name": "your-app-name",
    "deployment_type": "webapp",
    "region": "eastus"
  },
  "environment": {
    "variables": {
      "ENV": "production"
    }
  }
}
```

### Authentication Methods

#### Token Authentication
```json
{
  "auth_type": "token",
  "token": "your-personal-access-token"
}
```

#### SSH Authentication
```json
{
  "auth_type": "ssh",
  "ssh_key_path": "~/.ssh/id_rsa"
}
```

#### Basic Authentication
```json
{
  "auth_type": "password",
  "username": "your-username",
  "password": "your-password"
}
```

## Commands

| Command | Description |
|---------|-------------|
| `automateLife init` | Initialize configuration file |
| `automateLife start` | Clone repository and optionally run tests |
| `automateLife test` | Run tests on cloned repository |
| `automateLife verify` | Verify configuration is valid |

## Path Expansion

AutomateLife automatically expands `~` and `$HOME` in all path configurations:

```json
{
  "ssh_key_path": "~/.ssh/id_rsa",           // ✅ Expands to /Users/username/.ssh/id_rsa
  "output_dir": "$HOME/builds",              // ✅ Expands to /Users/username/builds
  "build_command": "go build -o ~/bin/app"  // ✅ Expands paths in commands too
}
```

## Development

### Project Structure

```
.
├── builder/         # Build and test command execution
├── config/          # Configuration management
├── git/            # Git authentication and operations
├── handlers/       # Command handlers (init, start, test)
├── ui/             # User interface utilities
├── utils/          # Utility functions (path expansion, etc.)
└── main.go         # Entry point
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run verbose
go test ./... -v

# Run specific package
go test ./utils -v
```

### Test Coverage

- **utils**: 100.0%
- **git**: 97.2%
- **config**: 95.3%
- **builder**: 81.8%

See [TEST_SUMMARY.md](TEST_SUMMARY.md) for detailed test documentation.

### Code Quality

```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run

# Build
go build -o automateLife
```

## Examples

### Example 1: Simple Go Project

```bash
# Initialize
automateLife init

# Follow prompts:
# - Provider: github
# - Auth: token
# - Language: go
# - Type: backend

# Start
automateLife start
# Select 'y' to run tests immediately
```

### Example 2: Python Project with SSH

```bash
automateLife init

# Follow prompts:
# - Provider: gitlab
# - Auth: ssh
# - SSH Key: ~/.ssh/id_rsa
# - Language: python
# - Type: backend

automateLife start
automateLife test
```

### Example 3: Azure DevOps Deployment

```bash
automateLife init

# Follow prompts:
# - Provider: azure-devops
# - Auth: token
# - Language: dotnet
# - Deployment Type: webapp
# - Configure Azure settings

automateLife start
```

## Environment Variables

AutomateLife automatically sets and uses:

- `HOME`: User home directory (auto-detected)
- `GIT_SSH_COMMAND`: SSH configuration for Git
- `GIT_TERMINAL_PROMPT`: Disabled for non-interactive auth
- Custom environment variables from config

## Troubleshooting

### SSH Authentication Issues

```bash
# Ensure SSH key exists
ls -la ~/.ssh/id_rsa

# Test SSH connection
ssh -T git@github.com

# Check SSH agent
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_rsa
```

### Token Authentication Issues

- Verify token has correct permissions (Code: Read/Write)
- Check token hasn't expired
- Ensure token is correctly copied (no extra spaces)

### Path Issues

- Use `~` or `$HOME` instead of hardcoded paths
- AutomateLife expands paths automatically
- Check paths with: `echo ~` and `echo $HOME`

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`go test ./...`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Roadmap

- [ ] Support for more cloud providers (AWS, GCP)
- [ ] Docker integration
- [ ] Kubernetes deployment
- [ ] Multi-repository support
- [ ] CI/CD template generation
- [ ] Webhook integration
- [ ] Slack/Discord notifications
- [ ] Performance metrics tracking

## Support

- 📧 Email: [Your Email]
- 🐛 Issues: [GitHub Issues](link-to-issues)
- 📖 Documentation: [Wiki](link-to-wiki)

## Acknowledgments

- Built with Go
- Uses [promptui](https://github.com/manifoldco/promptui) for interactive prompts
- Inspired by automation needs in DevOps workflows

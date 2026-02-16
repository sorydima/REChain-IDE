# Environment Setup Guide

This document describes the environment setup for the REChain Quantum-CrossAI IDE Engine project.

## Required Tools

### Core Dependencies

| Tool | Version | Purpose | Installation |
|------|---------|---------|--------------|
| Go | 1.21+ | Backend development | [Download](https://golang.org/dl/) |
| Node.js | 18.x | Frontend/Documentation | [Download](https://nodejs.org/) |
| Git | 2.35+ | Version control | [Download](https://git-scm.com/) |

### Optional Tools

| Tool | Purpose | Installation |
|------|---------|--------------|
| Docker | Containerization | [Docker Desktop](https://www.docker.com/products/docker-desktop) |
| Make | Build automation | Included with Git on Windows / `brew install make` on macOS |
| golangci-lint | Go linting | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` |
| govulncheck | Security scanning | `go install golang.org/x/vuln/cmd/govulncheck@latest` |

## IDE/Editor Setup

### Visual Studio Code

Recommended extensions:
- Go (golang.go)
- ES7+ React/Redux/React-Native snippets (dsznajder.es7-react-js-snippets)
- Prettier (esbenp.prettier-vscode)
- ESLint (dbaeumer.vscode-eslint)
- Markdown All in One (yzhang.markdown-all-in-one)

Settings (`.vscode/settings.json`):
```json
{
  "go.formatTool": "gofmt",
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.vulncheckPath": "govulncheck",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": "explicit"
  }
}
```

### GoLand / IntelliJ IDEA

1. Install the Go plugin
2. Configure Go SDK: File → Project Structure → SDKs
3. Enable `go fmt` on save: Settings → Tools → File Watchers

## Environment Variables

Create a `.env` file in the project root (do not commit this file):

```bash
# Development
DEBUG=true
LOG_LEVEL=debug

# API Keys (if needed)
# API_KEY=your_api_key_here

# Database (if applicable)
# DATABASE_URL=postgres://user:pass@localhost/rechain

# External services
# DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/...
```

## Initial Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/rechain-ai/rechain-ide.git
   cd rechain-ide
   ```

2. **Run the setup script:**
   ```bash
   ./scripts/setup-repo.sh
   ```

3. **Verify installation:**
   ```bash
   cd rechain-ide
   go version
   go build ./...
   ```

## Git Configuration

### Required Settings

```bash
# Set your identity
git config user.name "Your Name"
git config user.email "your.email@example.com"

# Set default branch name
git config init.defaultBranch main

# Configure line endings (Windows)
git config core.autocrlf true

# Configure line endings (macOS/Linux)
git config core.autocrlf input
```

### Install Git Hooks

```bash
./scripts/install-hooks.sh
```

## Workspace Structure

```
rechain-ide/
├── rechain-ide/          # Main Go module
│   ├── cmd/              # Application entry points
│   ├── internal/         # Internal packages
│   ├── pkg/              # Public packages
│   └── go.mod
├── docs/                 # Documentation
├── scripts/              # Utility scripts
├── .github/              # GitHub configuration
├── .gitlab/              # GitLab configuration
└── README.md
```

## Common Commands

### Development

```bash
# Build the project
make build

# Run tests
make test

# Run linting
make lint

# Format code
make fmt

# Clean build artifacts
make clean
```

### Git Workflow

```bash
# Create a new feature branch
./scripts/create-branch.sh feature my-feature

# Check if PR is ready
./scripts/check-pr.sh

# Bump version
./scripts/bump-version.sh patch
```

## Troubleshooting

### Go Module Issues

If you encounter module-related errors:

```bash
cd rechain-ide
go mod tidy
go mod download
```

### Git Hooks Not Running

Ensure hooks are executable:

```bash
chmod +x .git/hooks/*
```

On Windows, Git Bash should handle this automatically.

### Line Ending Issues

If you see `^M` characters or related errors:

```bash
# Re-normalize line endings
git add --renormalize .
git commit -m "Normalize line endings"
```

## Docker Setup (Optional)

```bash
# Build development image
docker build -f Dockerfile.dev -t rechain-ide:dev .

# Run development container
docker run -it -v $(pwd):/workspace rechain-ide:dev

# Using docker-compose
docker-compose up -d
```

## CI/CD Testing Locally

### GitHub Actions

Use [act](https://github.com/nektos/act) to test workflows locally:

```bash
# Install act
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Run workflows locally
act push
act pull_request
```

### GitLab CI

Use [GitLab Runner](https://docs.gitlab.com/runner/) locally:

```bash
# Register and run a local runner
gitlab-runner exec docker build
```

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Git Documentation](https://git-scm.com/doc)
- [Conventional Commits](https://www.conventionalcommits.org/)

## Support

If you encounter issues:
1. Check existing [issues](https://github.com/rechain-ai/rechain-ide/issues)
2. Join our [Discord](https://discord.gg/rechain)
3. Create a new issue with the "help wanted" label

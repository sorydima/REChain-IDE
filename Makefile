# Makefile for REChain IDE project

# Variables
PROJECT_NAME := rechain-ide
GO_MODULES := agents cli kernel orchestrator quantum rag shared vscode-extension web6-3d windsrif-api

# Default target
.PHONY: help
help: ## Display this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build all Go modules
	@echo "Building all Go modules..."
	@for module in $(GO_MODULES); do \
		echo "Building $$module..."; \
		cd $(PROJECT_NAME)/$$module && go build ./... && cd ../..; \
	done

.PHONY: test
test: ## Run tests for all Go modules
	@echo "Running tests for all Go modules..."
	@for module in $(GO_MODULES); do \
		echo "Testing $$module..."; \
		cd $(PROJECT_NAME)/$$module && go test ./... && cd ../..; \
	done

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@find . -name "*.exe" -delete
	@find . -name "*.out" -delete
	@find . -name "*.test" -delete
	@rm -rf dist/

.PHONY: docker-build
docker-build: ## Build Docker images
	@echo "Building Docker images..."
	docker-compose build

.PHONY: docker-up
docker-up: ## Start Docker containers
	@echo "Starting Docker containers..."
	docker-compose up -d

.PHONY: docker-down
docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	docker-compose down

.PHONY: docs-serve
docs-serve: ## Serve documentation locally
	@echo "Serving documentation..."
	docker-compose up -d docs

.PHONY: validate
validate: ## Run validation scripts
	@echo "Running validation scripts..."
	@for script in scripts/*.ps1; do \
		echo "Running $$script..."; \
		powershell -ExecutionPolicy Bypass -File "$$script"; \
	done

.PHONY: dev
dev: ## Start development environment
	@echo "Starting development environment..."
	./scripts/dev.ps1

.PHONY: install
install: ## Install dependencies
	@echo "Installing dependencies..."
	@for module in $(GO_MODULES); do \
		echo "Installing dependencies for $$module..."; \
		cd $(PROJECT_NAME)/$$module && go mod tidy && cd ../..; \
	done

# Git-related targets
.PHONY: git-setup
git-setup: ## Setup git hooks and initial configuration
	@echo "Setting up git hooks..."
	@./scripts/setup-repo.sh

.PHONY: git-hooks
git-hooks: ## Install git hooks
	@echo "Installing git hooks..."
	@./scripts/install-hooks.sh

.PHONY: check-pr
check-pr: ## Check if current branch is ready for PR
	@echo "Checking PR readiness..."
	@./scripts/check-pr.sh

.PHONY: create-branch
create-branch: ## Create a new feature branch (usage: make create-branch TYPE=feature NAME=my-feature)
	@if [ -z "$(TYPE)" ] || [ -z "$(NAME)" ]; then \
		echo "Usage: make create-branch TYPE=feature NAME=my-feature"; \
		echo "Types: feature, bugfix, hotfix, docs, refactor"; \
		exit 1; \
	fi
	@./scripts/create-branch.sh $(TYPE) $(NAME)

.PHONY: bump-version
bump-version: ## Bump version (usage: make bump-version LEVEL=patch)
	@if [ -z "$(LEVEL)" ]; then \
		echo "Usage: make bump-version LEVEL=patch"; \
		echo "Levels: major, minor, patch"; \
		exit 1; \
	fi
	@./scripts/bump-version.sh $(LEVEL)

.PHONY: clean-branches
clean-branches: ## Clean up merged branches
	@echo "Cleaning up merged branches..."
	@git branch --merged | grep -v "^\*" | grep -v "main\|develop" | xargs -n 1 git branch -d || true
	@echo "Fetching prune info..."
	@git fetch --prune

.PHONY: changelog
changelog: ## Generate changelog
	@echo "Generating changelog..."
	@git log --pretty=format:"- %s (%h)" --no-merges $$(git describe --tags --abbrev=0 2>/dev/null || echo HEAD~10)..HEAD > CHANGELOG.new.md
	@cat CHANGELOG.new.md
	@rm CHANGELOG.new.md
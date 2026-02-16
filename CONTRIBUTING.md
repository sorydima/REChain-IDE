# Contributing to REChain Quantum-CrossAI IDE Engine

Thank you for your interest in contributing to the REChain Quantum-CrossAI IDE Engine! We welcome contributions from the community and are excited to see what you'll build.

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md) to ensure a positive experience for all contributors.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/your-username/rechain-ide.git`
3. Create a new branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Commit your changes: `git commit -am 'Add some feature'`
6. Push to the branch: `git push origin feature/your-feature-name`
7. Create a new Pull Request

## Development Setup

See our [Development Setup Guide](docs/dev-setup.md) for detailed instructions on setting up your development environment.

## Project Structure

The project is organized into several modules within the `rechain-ide/` directory:

- `agents/` - AI agents for code assistance
- `cli/` - Command-line interface
- `kernel/` - Core IDE kernel
- `orchestrator/` - Cross-AI model orchestrator
- `quantum/` - Quantum computing integration
- `rag/` - Retrieval-Augmented Generation system
- `shared/` - Shared libraries
- `vscode-extension/` - VS Code extension
- `web6-3d/` - 3D visualization engine

## Coding Standards

### Go Code

- Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Run `go vet` to identify potential issues
- Write tests for your code
- Use meaningful variable and function names
- Add comments for exported functions and types

### TypeScript Code

- Follow the [TypeScript Coding Guidelines](https://github.com/Microsoft/TypeScript/wiki/Coding-guidelines)
- Use 2-space indentation
- Use meaningful variable and function names
- Add JSDoc comments for functions and classes

## Testing

- Write unit tests for all new functionality
- Ensure all tests pass before submitting a pull request
- Run the full test suite with `make test`

## Documentation

- Update documentation when making changes to functionality
- Add examples for new features
- Ensure all public APIs are documented

## Pull Request Process

1. Ensure any install or build dependencies are removed before the end of the layer when doing a build
2. Update the README.md with details of changes to the interface, this includes new environment variables, exposed ports, useful file locations and container parameters
3. Increase the version numbers in any examples files and the README.md to the new version that this Pull Request would represent
4. You may merge the Pull Request in once you have the sign-off of two other developers, or if you do not have permission to do that, you may request the second reviewer to merge it for you

## Reporting Issues

Before submitting an issue, please check the existing issues to avoid duplicates. When submitting a new issue, please include:

1. A clear and descriptive title
2. Steps to reproduce the issue
3. Expected behavior
4. Actual behavior
5. Environment information (OS, Go version, etc.)

## Security

If you discover a security vulnerability, please follow our [Security Policy](SECURITY.md) for responsible disclosure.

## Questions?

If you have any questions about contributing, feel free to ask in the discussions section or contact the maintainers.

# Contributing to goDev

Thank you for your interest in contributing to **goDev**! We welcome contributions from everyone and are grateful for any help you'd like to offer.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Coding Guidelines](#coding-guidelines)
- [Submitting a Pull Request](#submitting-a-pull-request)
- [Reporting Bugs](#reporting-bugs)
- [Requesting Features](#requesting-features)

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to [15264938+anburocky3@users.noreply.github.com](mailto:15264938+anburocky3@users.noreply.github.com).

## How Can I Contribute?

### Reporting Bugs

If you've found a bug, please create an issue on GitHub with the following details:

- A clear, descriptive title
- Steps to reproduce the issue
- Expected vs. actual behavior
- Your operating system (Windows/macOS/Linux)
- Service versions (Apache, Nginx, MySQL, PHP) if applicable

### Suggesting Features

We'd love to hear your feature ideas! Please create an issue with:

- A clear, descriptive title
- A detailed description of the proposed feature
- Any relevant examples or mockups
- Explanation of why this feature would be useful

### Writing Code

- Fork the repository and create your branch from `main`
- Implement your changes
- Test thoroughly on your platform
- Submit a pull request with a clear description

## Development Setup

### Prerequisites

- [Go](https://golang.org/dl/) 1.25.0 or later
- [Bun](https://bun.sh/) - Fast JavaScript runtime and package manager
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

### Local Development

```bash
# Clone the repository
git clone https://github.com/yourusername/goDev.git
cd goDev

# Install Go dependencies
go mod download

# Install frontend dependencies
cd frontend
bun install
cd ..

# Run in development mode (hot-reload enabled)
wails dev
```

### Building for Production

```bash
# Build for your current platform
wails build

# Build for a specific platform
wails build -platform windows/amd64
wails build -platform darwin/arm64
wails build -platform linux/amd64
```

## Coding Guidelines

### Go Backend

- Follow [Effective Go](https://golang.org/doc/effective_go) conventions
- Use `gofmt` or `goimports` for formatting
- Ensure all functions return errors — no silent failures
- Keep functions focused and modular
- Add comments for exported functions and types

### React Frontend

- Use TypeScript for all new components
- Follow existing component structure and patterns
- Use Tailwind CSS for styling — avoid inline styles
- Keep components small and reusable
- Use functional components with hooks

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>(<scope>): <description>

[optional body]
```

**Types:**

| Type       | Description                               |
| ---------- | ----------------------------------------- |
| `feat`     | A new feature                             |
| `fix`      | A bug fix                                 |
| `docs`     | Documentation changes                     |
| `style`    | Code style (formatting, semicolons, etc.) |
| `refactor` | Code refactoring                          |
| `test`     | Adding or updating tests                  |
| `chore`    | Maintenance tasks, dependencies           |

**Examples:**

```
feat(nginx): add graceful reload support
fix(mysql): resolve data directory permission issue on macOS
docs: update installation instructions
```

## Submitting a Pull Request

1. Fork the repository
2. Create a feature or fix branch (`feat/add-phpmyadmin` or `fix/nginx-reload`)
3. Make your changes with clear, focused commits
4. Ensure the application builds and runs without errors
5. Update documentation if needed
6. Push to your fork and submit the Pull Request

### PR Checklist

- [ ] Description clearly explains the changes
- [ ] Code follows the project's coding guidelines
- [ ] Changes have been tested on at least one platform
- [ ] Documentation has been updated if needed
- [ ] No unnecessary files are included in the commit

## Project Structure

```
goDev/
├── frontend/          # React + TypeScript frontend
│   ├── src/           # Source components
│   └── public/        # Static assets
├── generated/         # Generated config files (Apache, MySQL)
├── tools/             # Bundled service binaries
│   ├── apache/
│   ├── mysql/
│   ├── nginx/
│   └── php/
├── www/               # Default web root directory
├── wails.json         # Wails configuration
├── config.yaml        # Service configuration
├── services.go        # Service orchestration logic
└── tray_*.go          # System tray integration
```

## Getting Help

- Check existing [Issues](https://github.com/yourusername/goDev/issues) for known problems
- Read the [README](README.md) for setup and usage instructions
- Feel free to ask questions in GitHub Discussions or open a new issue

## License

By contributing to goDev, you agree that your contributions will be licensed under the [MIT License](LICENSE).

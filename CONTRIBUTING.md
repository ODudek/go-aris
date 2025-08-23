# Contributing to Aris

We love your input! We want to make contributing to Aris as easy and transparent as possible, whether it's:

- Reporting a bug
- Discussing the current state of the code
- Submitting a fix
- Proposing new features
- Becoming a maintainer

## Development Process

We use GitHub to host code, to track issues and feature requests, as well as accept pull requests.

## Pull Requests

Pull requests are the best way to propose changes to the codebase. We actively welcome your pull requests:

1. Fork the repo and create your branch from `main`.
2. If you've added code that should be tested, add tests.
3. If you've changed APIs, update the documentation.
4. Ensure the test suite passes.
5. Make sure your code lints.
6. Issue that pull request!

## Development Setup

### Prerequisites

- Go 1.20 or higher
- Git

### Setting up the development environment

```bash
# Clone your fork
git clone https://github.com/your-username/go-aris.git
cd go-aris

# Add upstream remote
git remote add upstream https://github.com/ODudek/go-aris.git

# Install dependencies (if any)
go mod tidy
```

### Running tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./resp -v
```

### Code Style

- Follow standard Go conventions and formatting
- Use `gofmt` to format your code
- Run `go vet` to check for common mistakes
- Write clear, descriptive commit messages

## Any contributions you make will be under the MIT Software License

In short, when you submit code changes, your submissions are understood to be under the same [MIT License](LICENSE) that covers the project.

## Report bugs using GitHub's [issue tracker](https://github.com/ODudek/go-aris/issues)

We use GitHub issues to track public bugs. Report a bug by [opening a new issue](https://github.com/ODudek/go-aris/issues/new); it's that easy!

**Great Bug Reports** tend to have:

- A quick summary and/or background
- Steps to reproduce
  - Be specific!
  - Give sample code if you can
- What you expected would happen
- What actually happens
- Notes (possibly including why you think this might be happening, or stuff you tried that didn't work)

## Feature Requests

We welcome feature requests! Please [open an issue](https://github.com/ODudek/go-aris/issues/new) to discuss new features before implementing them.

## Project Priorities

Currently, the main priorities are:

1. Implementing core Redis commands (SET, GET, PING, etc.)
2. Adding in-memory storage
3. Improving test coverage
4. Adding data persistence
5. Supporting concurrent connections

## License

By contributing, you agree that your contributions will be licensed under its MIT License.
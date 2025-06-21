# Contributing to Galick

Thank you for considering contributing to Galick! This document outlines the process for contributing to the project.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct. Please be respectful and considerate of others.

## How to Contribute

### Reporting Bugs

If you find a bug in the code, please create an issue with the following information:

- A clear, descriptive title
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Any relevant logs or screenshots
- Your environment (OS, Go version, Vegeta version, etc.)

### Suggesting Features

We welcome feature suggestions! Please create an issue with:

- A clear, descriptive title
- A detailed description of the feature
- Why you believe it would be valuable
- Any implementation ideas you have

### Pull Requests

1. Fork the repository
2. Create a new branch from `main`
3. Make your changes
4. Add or update tests as needed
5. Ensure all tests pass with `make test`
6. Make sure your code follows the project's style guidelines
7. Submit a pull request

## Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/kanywst/galick.git
   cd galick
   ```

2. Install dependencies:
   ```bash
   make setup-dev
   ```

3. Run tests:
   ```bash
   make test
   ```

## Coding Style

- Follow standard Go coding conventions
- Use meaningful variable and function names
- Write clear comments
- Include tests for new functionality

## Versioning

We use [Semantic Versioning](https://semver.org/). Please make sure your changes are compatible with this versioning system.

## License

By contributing to Galick, you agree that your contributions will be licensed under the project's MIT license.

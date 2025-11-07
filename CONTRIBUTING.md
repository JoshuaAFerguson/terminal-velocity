# Contributing to Terminal Velocity

Thank you for your interest in contributing to Terminal Velocity! This document provides guidelines and instructions for contributing.

## Getting Started

### Prerequisites

- Go 1.23 or higher
- PostgreSQL 14+ (for database features)
- Git
- A terminal emulator
- SSH client (for testing)

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/s0v3r1gn/terminal-velocity.git
   cd terminal-velocity
   ```

2. **Install dependencies**
   ```bash
   make install-deps
   # or
   go mod download
   ```

3. **Set up configuration**
   ```bash
   cp configs/config.example.yaml configs/config.yaml
   # Edit configs/config.yaml with your settings
   ```

4. **Set up database** (optional, for full features)
   ```bash
   make setup-db
   ```

5. **Run the server**
   ```bash
   make run
   # or
   go run cmd/server/main.go
   ```

6. **Connect via SSH**
   ```bash
   ssh -p 2222 username@localhost
   ```

## Development Workflow

### Branching Strategy

- `main` - Stable, production-ready code
- `develop` - Integration branch for features
- `feature/*` - New features
- `bugfix/*` - Bug fixes
- `refactor/*` - Code refactoring

### Making Changes

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write clean, idiomatic Go code
   - Follow the existing code style
   - Add tests for new functionality
   - Update documentation as needed

3. **Test your changes**
   ```bash
   make test
   make lint
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add new feature description"
   ```

   Use conventional commit messages:
   - `feat:` - New feature
   - `fix:` - Bug fix
   - `docs:` - Documentation changes
   - `refactor:` - Code refactoring
   - `test:` - Adding tests
   - `chore:` - Maintenance tasks

5. **Push and create PR**
   ```bash
   git push origin feature/your-feature-name
   ```
   Then create a Pull Request on GitHub.

## Code Guidelines

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Keep functions small and focused
- Write descriptive variable names

### Project Structure

```
terminal-velocity/
├── cmd/server/          # Main application entry
├── internal/            # Private application code
│   ├── server/         # SSH server & sessions
│   ├── game/           # Game engine modules
│   ├── ui/             # Terminal UI components
│   ├── database/       # Database layer
│   └── models/         # Data models
├── pkg/                # Public libraries
├── configs/            # Configuration files
└── scripts/            # Tools and migrations
```

### Testing

- Write unit tests for game logic
- Use table-driven tests where appropriate
- Mock external dependencies
- Aim for >70% coverage on critical paths

Example test:
```go
func TestPlayer_CanAfford(t *testing.T) {
    tests := []struct {
        name    string
        credits int64
        amount  int64
        want    bool
    }{
        {"sufficient credits", 1000, 500, true},
        {"insufficient credits", 100, 500, false},
        {"exact amount", 500, 500, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p := &Player{Credits: tt.credits}
            got := p.CanAfford(tt.amount)
            if got != tt.want {
                t.Errorf("CanAfford() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Documentation

- Add godoc comments for public functions
- Update README.md for user-facing changes
- Update ROADMAP.md when completing phases
- Document complex algorithms

Example documentation:
```go
// CalculateProfit returns the profit from a trade transaction.
// It takes the buy price, sell price, and quantity and returns
// the total profit (or loss if negative).
func CalculateProfit(buyPrice, sellPrice int64, quantity int) int64 {
    return (sellPrice - buyPrice) * int64(quantity)
}
```

## Areas for Contribution

### High Priority
- Universe generation algorithms
- Combat system implementation
- Mission generation
- UI/UX improvements
- Performance optimization

### Good First Issues
- Adding new ship types
- Adding new commodities
- Writing tests
- Documentation improvements
- Bug fixes

### Ideas Welcome
- New mission types
- Special events
- Random encounters
- Achievements
- Quality of life features

## Reporting Issues

### Bug Reports

Include:
- Terminal Velocity version
- Operating system
- Steps to reproduce
- Expected behavior
- Actual behavior
- Relevant logs

### Feature Requests

Include:
- Clear description of the feature
- Use case / motivation
- Proposed implementation (optional)
- Impact on gameplay

## Community Guidelines

- Be respectful and constructive
- Help others learn
- Focus on the code, not the person
- Assume good intentions
- Have fun building a great game!

## Questions?

- Open a GitHub issue
- Check existing documentation
- Review the ROADMAP.md

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Terminal Velocity! 🚀

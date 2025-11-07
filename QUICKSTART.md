# Terminal Velocity - Quick Start Guide

This guide will help you get Terminal Velocity running quickly.

## Prerequisites

### Required
- **Go 1.23+**: [Install Go](https://golang.org/doc/install)
- **Git**: For cloning the repository

### Optional (for full features)
- **PostgreSQL 14+**: For persistent multiplayer universe
- **Make**: For using the Makefile commands

## Quick Start (Development Mode)

### 1. Install Go

**On Ubuntu/Debian:**
```bash
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
```

**On macOS (with Homebrew):**
```bash
brew install go
```

**Verify installation:**
```bash
go version
```

### 2. Clone and Setup

```bash
# Clone repository
git clone https://github.com/s0v3r1gn/terminal-velocity.git
cd terminal-velocity

# Install dependencies
go mod download

# Create config file
cp configs/config.example.yaml configs/config.yaml
```

### 3. Run Server (Simple Mode - No Database)

For quick testing without PostgreSQL:

```bash
go run cmd/server/main.go --port 2222
```

The server will start on port 2222.

### 4. Connect via SSH

In another terminal:

```bash
ssh -p 2222 player@localhost
```

Enter any username/password (auth is currently placeholder).

## Full Setup (With Database)

### 1. Install PostgreSQL

**On Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
```

**On macOS:**
```bash
brew install postgresql@14
brew services start postgresql@14
```

### 2. Create Database

```bash
# Switch to postgres user
sudo -u postgres psql

# In PostgreSQL shell:
CREATE DATABASE terminal_velocity;
CREATE USER terminal_velocity WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE terminal_velocity TO terminal_velocity;
\q
```

### 3. Initialize Schema

```bash
psql -U postgres -d terminal_velocity -f scripts/schema.sql
```

### 4. Configure Database Connection

Edit `configs/config.yaml`:

```yaml
database:
  url: "postgres://terminal_velocity:your_secure_password@localhost:5432/terminal_velocity?sslmode=disable"
```

### 5. Run Server

```bash
make run
# or
go run cmd/server/main.go
```

## Using Makefile Commands

The project includes a Makefile for common tasks:

```bash
# Show all available commands
make help

# Install dependencies
make install-deps

# Build binary
make build

# Run server
make run

# Run tests
make test

# Format code
make fmt

# Clean build artifacts
make clean

# Complete dev setup
make dev-setup
```

## Project Status

**Current Phase**: Phases 0-7 Complete! ✅

Terminal Velocity is feature-complete for core gameplay with 29+ interconnected systems:

- ✅ Full multiplayer support (chat, factions, PvP, trading)
- ✅ Complete trading and economy system
- ✅ Turn-based combat with tactical AI
- ✅ Quest & storyline system with branching narratives
- ✅ Dynamic server events and competitions
- ✅ Interactive tutorial system
- ✅ Server administration tools
- ✅ Advanced ship customization
- ✅ Achievements, leaderboards, missions
- ✅ Session management with auto-save

**Next**: Integration testing, balance tuning, and community testing

See [ROADMAP.md](ROADMAP.md) for full development history.

## Development

### Running Tests

```bash
make test
```

### Code Quality

```bash
# Format code
make fmt

# Run linter (requires golangci-lint)
make lint

# Run go vet
make vet
```

### Building for Production

```bash
# Build single binary
make build

# Build for all platforms
make release
```

## Configuration

Key configuration options in `configs/config.yaml`:

```yaml
server:
  port: 2222              # SSH port
  max_players: 100        # Max concurrent players

game:
  starting_credits: 10000 # Starting money
  num_systems: 100        # Universe size
  enable_pvp: true        # Enable player combat

database:
  url: "postgres://..."   # Database connection
```

## Troubleshooting

### Go not found
Make sure Go is installed and in your PATH:
```bash
go version
```

### Port already in use
Change the port in config or command line:
```bash
go run cmd/server/main.go --port 2223
```

### Database connection failed
Check PostgreSQL is running:
```bash
sudo systemctl status postgresql
# or on macOS
brew services list
```

### SSH connection refused
Ensure server is running and firewall allows port 2222.

## Next Steps

1. **Read the full [README.md](README.md)** for detailed information
2. **Check [ROADMAP.md](ROADMAP.md)** to see development progress
3. **Join development** by reading [CONTRIBUTING.md](CONTRIBUTING.md)
4. **Report issues** on GitHub

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [PostgreSQL Docs](https://www.postgresql.org/docs/)
- [SSH Protocol](https://www.openssh.com/)

## Getting Help

- GitHub Issues: Report bugs or request features
- Documentation: Check the docs/ directory
- Code: Read the source code (it's documented!)

---

Welcome to Terminal Velocity! 🚀

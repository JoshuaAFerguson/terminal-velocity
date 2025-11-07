# Terminal Velocity

[![CI](https://github.com/JoshuaAFerguson/terminal-velocity/actions/workflows/ci.yml/badge.svg)](https://github.com/JoshuaAFerguson/terminal-velocity/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/JoshuaAFerguson/terminal-velocity)](https://goreportcard.com/report/github.com/JoshuaAFerguson/terminal-velocity)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/JoshuaAFerguson/terminal-velocity)](https://go.dev/)

A multiplayer space trading and combat game inspired by Escape Velocity, playable entirely through SSH.

## Overview

Terminal Velocity is a persistent multiplayer universe where players can:
- Trade commodities across star systems
- Engage in tactical turn-based combat
- Form player factions and control territory
- Complete missions and build reputation
- Progress from a humble shuttle to commanding capital ships

## Features

### Core Gameplay
- **Space Trading**: Buy low, sell high across a dynamic economy
- **Ship Progression**: Upgrade from shuttles to destroyers to capital ships
- **Turn-based Combat**: Tactical space battles with multiple weapon types
- **Mission System**: Delivery, combat, escort, and bounty missions

### Multiplayer Features
- **Player Factions**: Create guilds/corporations with shared goals
- **Territory Control**: Claim systems and earn passive income
- **Player Trading**: Direct player-to-player commerce
- **PvP Combat**: Consensual duels, faction wars, and piracy
- **Real-time Interactions**: See other players in your system
- **Communication**: Global, faction, and system chat channels

### Technical Features
- Playable via SSH from anywhere
- Persistent universe with real-time updates
- Beautiful terminal UI using Bubble Tea
- PostgreSQL backend for reliability

## Quick Start

### Docker (Recommended)

```bash
# Clone repository
git clone https://github.com/JoshuaAFerguson/terminal-velocity.git
cd terminal-velocity

# Configure environment
cp .env.example .env
# Edit .env and set DB_PASSWORD

# Start the stack
docker-compose up -d

# Connect to game
ssh -p 2222 username@localhost
```

See [Docker Guide](docs/DOCKER.md) for detailed instructions.

### Manual Setup

```bash
# Install dependencies
go mod download

# Set up database
psql -U postgres -f scripts/schema.sql

# Configure server
cp configs/config.example.yaml configs/config.yaml
# Edit configs/config.yaml with your settings

# Run server
go run cmd/server/main.go
```

### Client Connection

```bash
ssh -p 2222 player@your-server-address
```

## Development

### Project Structure

```
terminal-velocity/
├── cmd/
│   └── server/          # Server entry point
├── internal/
│   ├── server/          # SSH server and session management
│   ├── game/            # Game engine
│   │   ├── universe/    # Star systems, planets, navigation
│   │   ├── ship/        # Ship management and stats
│   │   ├── combat/      # Combat system
│   │   ├── trading/     # Economy and trading
│   │   ├── faction/     # Player factions
│   │   └── mission/     # Mission generation and tracking
│   ├── ui/              # Terminal UI components
│   ├── database/        # Database layer
│   └── models/          # Data models
├── pkg/
│   └── utils/           # Shared utilities
├── configs/             # Configuration files
└── scripts/             # Database migrations, tools
```

### Phase 1 Goals (Current)

- [x] Research and planning
- [x] Multiplayer feature design
- [ ] Basic SSH server with authentication
- [ ] Universe generation (systems, planets, routes)
- [ ] Navigation between systems
- [ ] Basic UI framework
- [ ] Database schema and persistence

## Technology Stack

- **Language**: Go 1.23+
- **UI Framework**: Bubble Tea + Lipgloss
- **Database**: PostgreSQL with pgx
- **SSH**: golang.org/x/crypto/ssh
- **Testing**: Go testing + testify

## Roadmap

See [ROADMAP.md](ROADMAP.md) for detailed development phases.

## License

MIT License - See LICENSE file for details

## Contributing

Contributions welcome! Please read CONTRIBUTING.md first.

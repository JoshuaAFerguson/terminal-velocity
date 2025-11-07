# Terminal Velocity

[![CI](https://github.com/JoshuaAFerguson/terminal-velocity/actions/workflows/ci.yml/badge.svg)](https://github.com/JoshuaAFerguson/terminal-velocity/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/JoshuaAFerguson/terminal-velocity)](https://goreportcard.com/report/github.com/JoshuaAFerguson/terminal-velocity)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/JoshuaAFerguson/terminal-velocity)](https://go.dev/)

A multiplayer space trading and combat game inspired by Escape Velocity, playable entirely through SSH.

## Overview

Terminal Velocity is a space trading and combat game inspired by Escape Velocity. Currently in active development with a fully playable single-player experience.

**What's Playable Now**:
- ✅ Dynamic trading economy with 15 commodities
- ✅ 11 ship types with full progression system
- ✅ Turn-based combat with tactical AI opponents
- ✅ Equipment customization (9 weapons, 15 outfits)
- ✅ Mission board with 4 mission types
- ✅ Reputation and bounty system with 6 NPC factions
- ✅ Loot and salvage from combat victories

**Coming Soon**:
- 🔄 Random encounters (pirates, traders, distress calls)
- 🔄 News and dynamic universe events
- 🔄 Multiplayer features (player factions, PvP, territory control)

## Features

### ✅ Implemented Features

#### Core Gameplay
- **Space Trading System**: Dynamic economy with 15 commodities
  - Supply and demand price fluctuations
  - Tech level modifiers affecting prices
  - 5+ profitable trade routes documented
  - Real-time market updates
  - Illegal goods tracking (contraband)

- **Ship Progression**: 11 ship types from Shuttle to Battleship
  - Complete ship statistics (hull, shields, speed, cargo, maneuverability)
  - Combat rating requirements for advanced ships
  - Trade-in system (70% value)
  - Side-by-side ship comparison tools
  - Performance ratings (combat, trading, speed)

- **Equipment & Outfitting**: Comprehensive customization system
  - 9 weapon types (laser, missile, plasma, railgun)
  - 15 outfit types (shields, hull, cargo, fuel, engines)
  - Tab-based installation interface
  - Equipment removal with 50% refund
  - Real-time stat calculations with bonuses

- **Turn-based Combat**: Tactical space battles
  - Full-screen tactical display with ASCII radar
  - 5 AI difficulty levels (Easy to Ace)
  - Weapon mechanics (range, accuracy, cooldown, ammo)
  - Shield penetration and critical hits
  - Target selection with threat assessment
  - Combat log with scrolling messages

- **Reputation & Bounty System**:
  - Faction reputation tracking (−100 to +100)
  - 8 combat event types affecting reputation
  - Bounty system with expiration
  - Legal status (clean, offender, wanted, fugitive)
  - Faction reinforcement mechanics

- **Loot & Salvage System**:
  - Dynamic loot generation from combat
  - Cargo recovery (30-60% survival rate)
  - Equipment salvaging (weapons, outfits)
  - Rare item drops (4 rarity tiers: legendary, epic, rare, uncommon)
  - 6 unique rare items worth 25K-250K credits

- **Mission System**: Mission board with multiple types
  - 4 mission types: Delivery, Combat, Bounty, Trading
  - Mission requirements (combat rating, reputation)
  - Deadline tracking with auto-expiration
  - Progress tracking for multi-part missions
  - Accept/decline mechanics with validation
  - 5 concurrent active mission limit

- **Fleet Management**:
  - Ship inventory with multiple owned ships
  - Active ship selection and switching
  - Ship renaming functionality
  - Configuration viewer (weapons, outfits, cargo)
  - Status displays (hull, shields, fuel)

#### Technical Features
- **SSH Server**: Multi-method authentication
  - Password authentication with bcrypt
  - SSH public key authentication
  - Multiple keys per account support
  - Account management CLI tool

- **Database Layer**: PostgreSQL with pgx/v5
  - Player, ship, system, and market repositories
  - Connection pooling
  - Transaction support with rollback
  - Migration system

- **Universe**: Procedurally generated galaxy
  - 100+ star systems
  - MST-based jump route connectivity
  - 6 NPC factions with territories and relationships
  - Tech level distribution (1-10)

- **UI Framework**: Beautiful terminal interface
  - BubbleTea + Lipgloss styling
  - Multiple game screens (trading, shipyard, outfitter, combat, missions)
  - Tab-based navigation
  - Visual status bars and indicators

### 🔄 In Development

- **Mission Integration**: Reward application and reputation effects
- **Random Encounters**: Pirates, traders, distress calls
- **News System**: Dynamic universe events

### 📋 Planned Features

#### Multiplayer Features (Phase 6)
- **Player Factions**: Create guilds/corporations with shared goals
- **Territory Control**: Claim systems and earn passive income
- **Player Trading**: Direct player-to-player commerce
- **PvP Combat**: Consensual duels, faction wars, and piracy
- **Real-time Interactions**: See other players in your system
- **Communication**: Global, faction, and system chat channels

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
│   ├── server/          # SSH game server
│   ├── accounts/        # Account management CLI
│   └── genmap/          # Universe generation tool
├── internal/
│   ├── server/          # SSH server and session management
│   ├── database/        # Database repositories (pgx)
│   │   ├── player.go    # Player CRUD operations
│   │   ├── system.go    # Universe persistence
│   │   ├── ship.go      # Ship management
│   │   └── market.go    # Market data
│   ├── models/          # Data models
│   │   ├── player.go    # Player, ship, cargo
│   │   ├── universe.go  # Systems, planets, routes
│   │   ├── trading.go   # Commodities, markets
│   │   ├── faction.go   # NPC factions
│   │   ├── ship.go      # Ship types, weapons, outfits
│   │   └── mission.go   # Mission structures
│   ├── combat/          # Combat system
│   │   ├── weapons.go   # Weapon mechanics
│   │   ├── ai.go        # Enemy AI
│   │   ├── reputation.go # Reputation & bounties
│   │   └── loot.go      # Loot generation
│   ├── missions/        # Mission system
│   │   └── manager.go   # Mission lifecycle
│   ├── tui/             # Terminal UI (BubbleTea)
│   │   ├── model.go     # Main TUI model
│   │   ├── trading.go   # Trading screen
│   │   ├── cargo.go     # Cargo management
│   │   ├── shipyard.go  # Ship purchasing
│   │   ├── outfitter.go # Equipment installation
│   │   ├── combat.go    # Combat interface
│   │   └── missions.go  # Mission board
│   └── universe/        # Universe generation
├── configs/             # Configuration files
├── docs/                # Documentation
└── scripts/             # Database migrations
```

### Development Status

**Current Phase**: Phase 5 - Missions & Progression (20% complete)

**Completed Phases**:
- ✅ **Phase 1**: Foundation & Navigation
  - SSH authentication (password + public key)
  - PostgreSQL database layer
  - Universe generation (100+ systems)
  - BubbleTea UI framework

- ✅ **Phase 2**: Core Economy
  - Trading system with 15 commodities
  - Dynamic market engine
  - Cargo management
  - Balanced economy

- ✅ **Phase 3**: Ship Progression
  - 11 ship types
  - Shipyard with comparison
  - Outfitter system
  - Fleet management

- ✅ **Phase 4**: Combat System
  - Turn-based combat
  - 9 weapon types
  - AI opponents (5 difficulty levels)
  - Reputation & bounty system
  - Loot & salvage

**In Progress**:
- 🔄 Mission system integration
- 🔄 Random encounter system
- 🔄 News/events system

**Milestones**:
- ✅ M1: Playable Prototype (trading works)
- ✅ M1.5: Single-Player Complete (combat + progression)
- 🎯 M2: Feature Complete (target: end of Phase 6)

## Technology Stack

- **Language**: Go 1.23+
- **UI Framework**: Bubble Tea + Lipgloss
- **Database**: PostgreSQL with pgx
- **SSH**: golang.org/x/crypto/ssh
- **Testing**: Go testing + testify

## Game Statistics

- **11** ship types (Shuttle, Fighter, Light Freighter, Heavy Freighter, Corvette, Frigate, Destroyer, Heavy Destroyer, Cruiser, Heavy Cruiser, Battleship)
- **15** commodities (Food, Water, Medicine, Luxury Goods, Electronics, Machinery, Weapons, Fuel, Ore, Rare Metals, Radioactives, Gems, Art, Narcotics, Slaves)
- **9** weapon types (Pulse Laser, Beam Laser, Heavy Laser, Light Missile, Heavy Missile, Plasma Cannon, Light Railgun, Heavy Railgun, Gatling Laser)
- **15** outfit types (Shield Boosters, Hull Reinforcement, Cargo Pods, Fuel Tanks, Engine Upgrades)
- **100+** star systems with procedural generation
- **6** NPC factions (UEF, Rigel Outer Marches, Free Traders Guild, Free Worlds Alliance, Crimson Syndicate, Auroran Empire)
- **4** mission types (Delivery, Combat, Bounty, Trading)
- **5** AI difficulty levels (Easy, Medium, Hard, Expert, Ace)

## Documentation

- [ROADMAP.md](ROADMAP.md) - Detailed development phases
- [CHANGELOG.md](CHANGELOG.md) - Complete change history
- [UNIVERSE_DESIGN.md](docs/UNIVERSE_DESIGN.md) - Galaxy structure and lore
- [FACTION_RELATIONS.md](docs/FACTION_RELATIONS.md) - NPC faction relationships
- [ECONOMY_BALANCE.md](docs/ECONOMY_BALANCE.md) - Trade route analysis
- [DOCKER.md](docs/DOCKER.md) - Docker deployment guide
- [IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md) - Feature tracking

## Roadmap

See [ROADMAP.md](ROADMAP.md) for detailed development phases and current status.

**Next Steps**:
1. Complete mission system integration
2. Implement random encounters
3. Add news/events system
4. Begin multiplayer features (Phase 6)

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

**Ways to Contribute**:
- Report bugs via GitHub Issues
- Suggest features or improvements
- Submit pull requests
- Write documentation
- Create content (missions, ships, commodities)
- Balance testing and feedback

## License

MIT License - See [LICENSE](LICENSE) file for details.

## Acknowledgments

Inspired by the classic Escape Velocity series by Ambrosia Software.

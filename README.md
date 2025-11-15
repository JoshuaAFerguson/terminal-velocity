# Terminal Velocity

[![CI](https://github.com/JoshuaAFerguson/terminal-velocity/actions/workflows/ci.yml/badge.svg)](https://github.com/JoshuaAFerguson/terminal-velocity/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/JoshuaAFerguson/terminal-velocity)](https://goreportcard.com/report/github.com/JoshuaAFerguson/terminal-velocity)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/JoshuaAFerguson/terminal-velocity)](https://go.dev/)
[![Sponsor](https://img.shields.io/badge/Sponsor-Terminal%20Velocity-ff69b4?logo=github-sponsors)](https://github.com/sponsors/JoshuaAFerguson)

A feature-rich multiplayer space trading and combat game inspired by Escape Velocity, playable entirely through SSH.

## Overview

Terminal Velocity is a **production-ready** comprehensive space trading and combat game with **full multiplayer support**. **All 20 development phases complete** with 245+ features, 41 interactive TUI screens, and 78,000+ lines of Go code!

**🎮 Production-Ready Features** (245+ total - see [FEATURES.md](FEATURES.md)):
- ✅ **Core Gameplay**: Trading (15 commodities), Combat (9 weapons, 5 AI levels), 11 ship types
- ✅ **Fleet Management**: Own 6 ships, escorts, formations, fleet combat
- ✅ **Advanced Systems**: Mining (12 resources), salvage, manufacturing, crafting, tech research
- ✅ **Player Stations**: Build & manage stations with production modules
- ✅ **Ship Capture**: Classic boarding mechanics with crew-based combat
- ✅ **Marketplace**: Auctions, contracts (4 types), bounties with time-based mechanics
- ✅ **Social**: Friends, mail system, notifications, player profiles
- ✅ **Multiplayer**: 4 chat channels, factions, alliances, diplomacy, territory control
- ✅ **PvP**: Arena system (5 types), tournaments, ranked/unranked modes, spectator
- ✅ **Content**: Quests (7 types), missions (4 types), dynamic events (10 types)
- ✅ **Progression**: Achievements, leaderboards (8 categories), reputation (6 NPC factions)
- ✅ **Universe**: Wormholes (4 types), nebulae, black holes, anomalies, 100+ systems
- ✅ **Infrastructure**: 2FA, password reset, metrics, backups, admin tools (RBAC)
- ✅ **Security**: 9.5/10 rating, rate limiting, auto-banning, audit logging

## 🚀 Recent Updates (2025-11-15)

**Major Feature Release** - New Roadmap Features + Production-Ready Infrastructure:

**🆕 New Gameplay Systems** (~7,200 lines of code):
- ✅ **Social Features** (Phase 9): Friends, mail, notifications, enhanced chat commands
- ✅ **Ship Capture** (Phase 12): Classic Escape Velocity boarding mechanics
- ✅ **Mining & Salvage** (Phase 15): 12 resource types, 3 target types, rarity system
- ✅ **Player Marketplace** (Phase 10): Auctions, contracts, bounties with time-based mechanics

**🔒 Production-Ready Infrastructure** - 61 critical bugs fixed + Enhanced Observability:
- ✅ **Security Fixes**: 6 money duplication exploits eliminated with atomic transactions
- ✅ **Concurrency Safety**: 15 race conditions fixed, all managers thread-safe
- ✅ **Input Validation**: 30+ fixes preventing memory exhaustion and injection attacks
- ✅ **Resource Management**: 3 goroutine leaks fixed, proper shutdown handling
- ✅ **Database Performance**: 17 strategic indexes added (10-100x improvement expected)
- ✅ **Enhanced Observability**: Latency histograms (p50/p95/p99), error categorization
- ✅ **Health Monitoring**: Comprehensive health checks with status indicators
- ✅ **Regression Tests**: 15+ tests ensuring bug fixes don't regress
- ✅ **TUI Integration**: All Phase 20+ screens fully integrated (Fleet, Friends, Marketplace, Notifications)
- ✅ **Build System**: Entire project compiles successfully with no errors or warnings

See [CHANGELOG.md](CHANGELOG.md) for complete details and [docs/SECURITY_AUDIT.md](docs/SECURITY_AUDIT.md) for security analysis.

## Features

### 🎯 Core Gameplay

#### Trading & Economy
- **Dynamic Market System**: 15 commodities with real-time price fluctuations
- **Supply & Demand**: Tech level modifiers, illegal goods tracking
- **Profitable Routes**: Multiple documented trade routes
- **Player Trading**: Direct player-to-player commerce with escrow

#### Ship Systems
- **11 Ship Types**: Shuttle → Battleship progression
- **Advanced Outfitting**: 6 equipment slot types, 16 unique items
- **Loadout System**: Save/load/clone ship configurations
- **Fleet Management**: Own multiple ships, switch between them
- **Performance Ratings**: Combat, trading, and speed metrics

#### Combat
- **Turn-Based Tactical**: Full-screen display with ASCII radar
- **9 Weapon Types**: Lasers, missiles, plasma, railguns
- **5 AI Difficulty Levels**: Easy to Ace with unique behaviors
- **PvP Combat**: Consensual duels, faction wars, piracy
- **Loot & Salvage**: 4 rarity tiers, rare item drops
- **Ship Capture (NEW)**: Classic Escape Velocity boarding mechanics
  - Two-stage process: board then capture enemy ships
  - Crew-based success calculations with casualties
  - Disable requirements: <25% hull, <10% shields
  - Cooldown system and thread-safe operation tracking

#### Resource Gathering (NEW)
- **Mining System**: Extract resources from asteroids
  - 12 resource types: ores, crystals, rare earth materials
  - 3 target types: asteroids, derelicts, debris fields
  - Rarity tiers: common, uncommon, rare, legendary
  - Equipment bonuses: mining lasers +25%/level, cargo scanner +15%
- **Salvage Operations**: Recover valuables from derelicts
  - Salvage weapons and outfits from rare ships
  - Scrap metal and components from all targets
  - Time-based cycle extraction (15s per cycle)
  - Scanner integration for revealing hidden resources

#### Reputation & Progression
- **Faction System**: 6 NPC factions with dynamic relationships
- **Reputation Tracking**: −100 to +100 per faction
- **Bounty System**: Legal status (clean → fugitive)
- **Achievements**: Track milestones and unlock rewards
- **Leaderboards**: Compete globally in multiple categories

### 📖 Content Systems

#### Quests & Storylines
- **7 Quest Types**: Main, Side, Faction, Daily, Chain, Hidden, Event
- **12 Objective Types**: Deliver, destroy, travel, collect, investigate, and more
- **Branching Narratives**: Player choices affect quest outcomes
- **"The Void Threat"**: Multi-quest main storyline
- **Comprehensive Rewards**: Credits, XP, items, reputation, unlocks

#### Missions
- **4 Mission Types**: Delivery, Combat, Bounty, Trading
- **Mission Board**: Browse and accept available missions
- **Progress Tracking**: Monitor objectives and deadlines
- **Reputation Requirements**: Unlock advanced missions

#### Dynamic Events
- **10 Event Types**: Trading competitions, combat tournaments, boss encounters, festivals
- **Community Goals**: Server-wide objectives with shared progress
- **Event Leaderboards**: Real-time rankings and rewards
- **Event Modifiers**: Temporary bonuses (2x credits, 1.5x XP, 2x drops)
- **5 Pre-defined Events**: Trade challenges, PvP tournaments, expeditions, boss fights

#### Random Encounters
- **Encounter System**: Pirates, traders, police, distress calls
- **Dynamic Spawns**: Based on system security and faction control
- **Loot Opportunities**: Combat rewards and salvage

#### News System
- **Dynamic News**: Universe events, player achievements, faction updates
- **10+ Event Types**: Combat victories, trade milestones, territorial changes
- **News Feed**: Stay informed about the galaxy

### 👥 Multiplayer Features

#### Communication
- **4 Chat Channels**: Global, System, Faction, Direct Messages
- **Enhanced Chat Commands**: `/whisper`, `/who`, `/roll`, `/me`, `/ignore` and more
- **Dice Rolling**: Full dice notation support (1d6, 2d10+5, etc.)
- **Player Presence**: See who's online and where
- **Announcements**: Server-wide notifications

#### Social Features (NEW)
- **Friends System**: Send/accept friend requests, manage friend list
- **Mail System**: Player-to-player messaging with credit/item attachments
- **Notifications**: 9 notification types with expiration tracking
- **Privacy Controls**: Block unwanted players, filter interactions
- **Online Status**: See when friends are online and where they are

#### Player Marketplace (NEW)
- **Auction House**: Time-based auctions (1h - 7 days)
  - Bid on ships, outfits, commodities, and special items
  - Instant buyout option with premium pricing
  - Bid history tracking and automatic expiry
  - Credit escrow and seller payouts
- **Contract System**: Player-posted missions
  - 4 contract types: courier, assassination, escort, bounty hunt
  - Claim and complete contracts for rewards
  - Failure penalties and expiry tracking
- **Bounty Board**: Post bounties on other players
  - Minimum 5000 credits, 10% posting fee
  - Automatic claim on target kill
  - Multiple bounties stack for big rewards

#### Factions & Territory
- **Player Factions**: Create guilds with shared treasury
- **Territory Control**: Claim systems, earn passive income
- **Faction Wars**: Coordinate attacks and defense
- **Member Management**: Ranks, permissions, invitations

#### Player Interaction
- **Player Visibility**: Real-time player locations
- **Direct Trading**: Exchange credits and items
- **PvP Combat**: Consensual and faction-based combat
- **Leaderboards**: Credits, combat rating, trade volume, exploration

### 🛠️ Infrastructure & Polish

#### Server Administration
- **4 Admin Roles**: Player, Moderator, Admin, SuperAdmin
- **20+ Permissions**: Granular access control (RBAC)
- **Moderation Tools**: Ban/mute with expiration tracking
- **Server Metrics**: Real-time performance monitoring
- **Audit Logging**: Complete action history (10,000 buffer)
- **Settings Management**: Configure economy, difficulty, rules

#### Session Management
- **Auto-Persistence**: Automatic saving every 30 seconds
- **Server-Authoritative**: No player-controlled saves
- **Session Tracking**: Monitor activity and connections
- **Graceful Disconnect**: Final save on exit

#### Player Experience
- **Interactive Tutorial**: 7 categories, 20+ steps with hints
- **Context-Aware Help**: Tutorials trigger based on actions
- **Settings System**: 6 categories, 5 color schemes including colorblind
- **Achievement Tracking**: Unlock milestones and badges
- **Notification System**: Event alerts, rewards, updates

### 🎨 Technical Features

- **SSH Server**: Multi-method authentication (password + public key)
- **PostgreSQL Database**: Full persistence with pgx connection pooling
- **BubbleTea UI**: Beautiful terminal interface with Lipgloss styling
- **Thread-Safe**: Concurrent operations with sync.RWMutex throughout
- **Background Workers**: Event scheduling, metrics collection, session cleanup
- **100+ Star Systems**: Procedurally generated with MST jump routes

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
docker compose up -d

# Connect to game
ssh -p 2222 username@localhost
```

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

### First-Time Players

When you connect, you'll be greeted by an interactive tutorial system that guides you through:
1. Basic navigation and UI
2. Trading fundamentals
3. Ship management
4. Combat basics
5. Mission system
6. Multiplayer features

## Game Statistics

### Codebase
- **78,002** lines of Go code
- **41** interactive TUI screens
- **48** internal packages
- **30+** database tables
- **14** database repositories
- **100+** tests passing
- **245+** documented features

### Content
- **11** ship types (Shuttle → Flagship)
- **15** commodities + **12** mineable resources
- **9** weapon types + **16** outfit items
- **100+** star systems with **4** wormhole types
- **6** NPC factions + player factions/alliances
- **7** quest types + **4** mission types
- **10** dynamic event types
- **5** PvP arena types
- **4** contract types + bounty system

### Infrastructure
- **20+** manager systems
- **4** admin roles with **20+** permissions
- **8** leaderboard categories
- **4** chat channels
- **9** notification types
- **Security Rating:** 9.5/10

## Development Status

**Current Status**: ✅ **All 20 Phases Complete - Production Ready!**

### Completed Development Phases

**Core Systems** (Phases 0-8):
- ✅ **Phase 0-1**: Foundation, Navigation, Universe Generation
- ✅ **Phase 2-3**: Trading Economy, Ship Progression
- ✅ **Phase 4-5**: Combat System, Missions, Quests, Events
- ✅ **Phase 6-7**: Multiplayer, Factions, Territory, Admin Tools
- ✅ **Phase 8**: Enhanced TUI Integration (56 tests passing)

**Advanced Features** (Phases 9-20):
- ✅ **Phase 9**: Social & Communication (friends, mail, notifications)
- ✅ **Phase 10**: Marketplace & Economy (auctions, contracts, bounties)
- ✅ **Phase 11**: Fleet Management (6 ships, escorts, formations)
- ✅ **Phase 12**: Ship Capture & Boarding (crew-based combat)
- ✅ **Phase 13**: Diplomacy & Alliances (inter-faction relations)
- ✅ **Phase 14**: Advanced Faction Systems (wars, conquest)
- ✅ **Phase 15**: Mining & Salvage (12 resources, derelicts)
- ✅ **Phase 16**: Advanced Systems (wormholes, cloaking, jump drives)
- ✅ **Phase 17**: Manufacturing & Crafting (stations, blueprints, tech tree)
- ✅ **Phase 18**: Competitive Systems (arenas, tournaments, rankings)
- ✅ **Phase 19**: Quality of Life (UI polish, automation, accessibility)
- ✅ **Phase 20**: Security & Infrastructure V2 (2FA, password reset, enhanced metrics)

See [ROADMAP.md](ROADMAP.md) for complete details on all phases.

### Project Milestones

- ✅ **M1**: Playable Prototype (Phases 0-4)
- ✅ **M1.5**: Single-Player Complete (Phases 5-7)
- ✅ **M2**: Multiplayer Functional (Phase 8)
- ✅ **M3**: Advanced Features (Phases 9-16)
- ✅ **M4**: Production Infrastructure (Phases 17-20)
- 🎯 **M5**: Public Launch (Beta testing → Production deployment)

## Technology Stack

- **Language**: Go 1.24+
- **UI Framework**: Bubble Tea + Lipgloss
- **Database**: PostgreSQL with pgx
- **SSH**: golang.org/x/crypto/ssh
- **Concurrency**: sync.RWMutex, context, goroutines
- **Testing**: Go testing + testify

## Documentation

### Essential Guides
- **[📖 Wiki](https://github.com/JoshuaAFerguson/terminal-velocity/wiki)** - Comprehensive player and developer guides
- **[FEATURES.md](FEATURES.md)** - Complete catalog of all 245+ features
- **[ROADMAP.md](ROADMAP.md)** - All 20 phases with complete implementation details
- [QUICKSTART.md](QUICKSTART.md) - Quick start guide
- [CHANGELOG.md](CHANGELOG.md) - Version history
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) - Community standards
- [SECURITY.md](SECURITY.md) - Security policy and vulnerability reporting

### Wiki Highlights
- [Getting Started](https://github.com/JoshuaAFerguson/terminal-velocity/wiki/Getting-Started) - Installation and first steps
- [Gameplay Guide](https://github.com/JoshuaAFerguson/terminal-velocity/wiki/Gameplay-Guide) - Core mechanics
- [Trading Guide](https://github.com/JoshuaAFerguson/terminal-velocity/wiki/Trading-Guide) - Economic strategies
- [FAQ](https://github.com/JoshuaAFerguson/terminal-velocity/wiki/FAQ) - Common questions
- [Architecture Overview](https://github.com/JoshuaAFerguson/terminal-velocity/wiki/Architecture-Overview) - Technical documentation

## Project Structure

```
terminal-velocity/
├── cmd/
│   ├── server/          # SSH game server
│   ├── accounts/        # Account management CLI
│   └── genmap/          # Universe generation tool
├── internal/
│   ├── server/          # SSH server & session management
│   ├── database/        # PostgreSQL repositories (pgx)
│   ├── models/          # Data models (player, universe, trading, etc.)
│   ├── combat/          # Combat system & AI
│   ├── missions/        # Mission lifecycle management
│   ├── quests/          # Quest & storyline system
│   ├── events/          # Dynamic events manager
│   ├── achievements/    # Achievement tracking
│   ├── news/            # News generation system
│   ├── leaderboards/    # Player rankings
│   ├── chat/            # Multiplayer chat
│   ├── factions/        # Player faction system
│   ├── territory/       # Territory control
│   ├── trade/           # Player-to-player trading
│   ├── pvp/             # PvP combat system
│   ├── presence/        # Player presence tracking
│   ├── encounters/      # Random encounter system
│   ├── outfitting/      # Equipment & loadouts
│   ├── settings/        # Player settings management
│   ├── tutorial/        # Tutorial & onboarding
│   ├── admin/           # Server administration
│   ├── session/         # Session & auto-save
│   ├── tui/             # Terminal UI (BubbleTea)
│   └── universe/        # Universe generation
├── configs/             # Configuration files
├── docs/                # Documentation
└── scripts/             # Database migrations
```

## Community

### Get Involved
- **[💬 Discussions](https://github.com/JoshuaAFerguson/terminal-velocity/discussions)** - Join the community!
  - [Announcements](https://github.com/JoshuaAFerguson/terminal-velocity/discussions/categories/announcements) - Official updates
  - [Q&A](https://github.com/JoshuaAFerguson/terminal-velocity/discussions/categories/q-a) - Ask questions
  - [Show and Tell](https://github.com/JoshuaAFerguson/terminal-velocity/discussions/categories/show-and-tell) - Share achievements
  - [Ideas](https://github.com/JoshuaAFerguson/terminal-velocity/discussions/categories/ideas) - Suggest features
- **[🐛 Issues](https://github.com/JoshuaAFerguson/terminal-velocity/issues)** - Report bugs and track progress

### Support Development
- **[💝 Sponsor](https://github.com/sponsors/JoshuaAFerguson)** - Support Terminal Velocity development
  - See [SPONSORS.md](SPONSORS.md) for our amazing sponsors
  - Tiers from $5/month to custom partnerships
  - Exclusive benefits including early access, in-game recognition, and direct input on features

## Production Infrastructure

Terminal Velocity includes production-ready monitoring, backup, and security features:

### Observability & Monitoring
- **Prometheus Metrics**: Full observability with `/metrics` endpoint on port 8080
- **Stats Dashboard**: Human-readable `/stats` page with real-time server statistics
- **Enhanced Metrics**: `/stats/enhanced` with latency percentiles and error tracking
- **Performance Profiling**: `/stats/performance` with color-coded health indicators
- **Health Checks**: `/health` endpoint with comprehensive status (healthy/degraded/unhealthy)
- **Metrics Tracked**:
  - Connections, players, game activity, economy, database performance
  - Operation latencies (p50/p95/p99 percentiles)
  - Error categorization and recent error history
  - Throughput metrics (trades/min, combat/min, queries/min)
  - Resource utilization and cache performance

```bash
curl http://localhost:8080/metrics            # Prometheus format
curl http://localhost:8080/stats              # HTML dashboard
curl http://localhost:8080/stats/enhanced     # Latency & error tracking
curl http://localhost:8080/stats/performance  # Performance profiling
curl http://localhost:8080/health             # Health status (JSON)
```

### Automated Backups
- **Automated Backups**: `scripts/backup.sh` with compression and retention policies
- **Easy Restore**: `scripts/restore.sh` with safety checks and verification
- **Cron Integration**: Example crontab for scheduled backups
- **Flexible Configuration**: Retention by days and count, custom backup locations

```bash
./scripts/backup.sh -d /var/backups -r 30 -c 50  # 30 days, keep 50
./scripts/restore.sh --list                       # List backups
```

### Rate Limiting & Security
- **Connection Limits**: 5 concurrent connections per IP, 20/minute rate limit
- **Auth Protection**: 5 failed attempts = 15 minute lockout
- **Auto-Banning**: 20 failed attempts = 24 hour automatic ban
- **Brute Force Protection**: Per-IP tracking with automatic cleanup

See [CLAUDE.md](CLAUDE.md) for detailed configuration options.

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

**Ways to Contribute**:
- Report bugs via [GitHub Issues](https://github.com/JoshuaAFerguson/terminal-velocity/issues)
- Suggest features in [Discussions](https://github.com/JoshuaAFerguson/terminal-velocity/discussions/categories/ideas)
- Submit pull requests
- Write documentation
- Create content (quests, events, ships)
- Balance testing and feedback
- Multiplayer testing

## Roadmap

✅ **All 20 development phases complete!** Terminal Velocity is production-ready with 245+ features implemented and tested.

### Next Steps (Launch Preparation)

1. **Beta Testing** (2-4 weeks)
   - Recruit 10-20 beta testers
   - Performance monitoring under real load
   - Bug fixes and polish

2. **Balance Tuning** (1-2 weeks)
   - Economy balance based on playtesting
   - Combat difficulty adjustments
   - Progression pacing refinement

3. **Performance Optimization** (1 week)
   - Load testing with 100+ concurrent players
   - Database query optimization
   - Caching strategy refinement

4. **Public Launch** (TBD)
   - Community announcement
   - 48-hour intensive monitoring
   - Rapid response to issues

### Documentation

- [ROADMAP.md](ROADMAP.md) - Complete phase-by-phase development history
- [FEATURES.md](FEATURES.md) - Comprehensive catalog of all 245+ features
- [CHANGELOG.md](CHANGELOG.md) - Detailed version history

## License

MIT License - See [LICENSE](LICENSE) file for details.

## Acknowledgments

Inspired by the classic Escape Velocity series by Ambrosia Software. Built with love for terminal-based gaming and the SSH community.

---

**Ready to play?** `ssh -p 2222 username@your-server-address`

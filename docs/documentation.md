---
layout: page
title: Documentation
permalink: /documentation/
description: Technical documentation, architecture, API references, and system design
---

# Technical Documentation

Comprehensive technical documentation for Terminal Velocity, covering architecture, game systems, APIs, and operations.

---

## 📖 Essential Documents

### Project Overview

**[README.md](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/README.md)**
Complete project overview with quick start, features, and development status.

**[FEATURES.md](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/FEATURES.md)**
Comprehensive catalog of all 245+ features organized by category.

**[ROADMAP.md](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/ROADMAP.md)**
Complete development history across all 20 phases with detailed feature breakdowns.

**[CHANGELOG.md](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/CHANGELOG.md)**
Detailed version history with all changes, fixes, and enhancements.

**[CLAUDE.md](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/CLAUDE.md)**
Complete developer reference guide with architecture, patterns, and development workflows.

---

## 🏗️ Architecture Documentation

### System Architecture

**[Architecture Refactoring]({{ '/ARCHITECTURE_REFACTORING' | relative_url }})**
Detailed design for future client-server architecture split:
- gRPC-based communication
- Horizontal scalability
- Multi-client support (SSH, web, native)
- State synchronization strategies
- Migration roadmap

**[API Integration Plan]({{ '/API_INTEGRATION_PLAN' | relative_url }})**
Plan for API-based architecture:
- API surface design
- Authentication flow
- State management
- Protocol specifications

**[API Migration Guide]({{ '/API_MIGRATION_GUIDE' | relative_url }})**
Step-by-step guide for migrating to API-based architecture:
- Phase-by-phase approach
- Backward compatibility
- Testing strategies
- Rollback procedures

### Current Architecture

**Monolithic Design** (Current):
```
SSH Server (Port 2222)
  ├─ Authentication & Session Management
  ├─ TUI Layer (BubbleTea)
  │   └─ Direct access to managers
  ├─ Game Logic Layer (Managers)
  │   └─ Direct database access
  └─ Database Layer (Repositories)
      └─ PostgreSQL connection pool
```

**Key Components**:
- **SSH Server**: `internal/server/` - Authentication, session management
- **TUI Framework**: `internal/tui/` - 41 BubbleTea screens
- **Game Managers**: 30+ manager packages for game systems
- **Database Layer**: 14 repositories with pgx connection pooling

See [CLAUDE.md](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/CLAUDE.md) for complete architecture details.

---

## 🎮 Game Systems Documentation

### Core Systems

**[Economy & Balance]({{ '/ECONOMY_BALANCE' | relative_url }})**
Trading economy design and balancing:
- Commodity pricing algorithms
- Supply/demand mechanics
- Tech level modifiers
- Profitable trade routes
- Market simulation

**[NPC Factions]({{ '/NPC_FACTIONS_SUMMARY' | relative_url }})**
Design of the 6 NPC factions:
- Government types
- Faction relationships
- Territory distribution
- Reputation mechanics

**[Faction Relations]({{ '/FACTION_RELATIONS' | relative_url }})**
Diplomatic relationships between factions:
- War/peace/neutral/allied states
- Relationship matrix
- Dynamic faction events

**[Universe Design]({{ '/UNIVERSE_DESIGN' | relative_url }})**
Procedural universe generation:
- Galaxy distribution algorithm
- MST-based jump routes
- Tech level radial distribution
- Planet generation

### Multiplayer Systems

**[Chat System]({{ '/CHAT_SYSTEM' | relative_url }})**
Real-time chat with 4 channels:
- Global, system, faction, DM channels
- Enhanced chat commands
- Dice rolling support
- Message broadcasting

**[Player Factions]({{ '/PLAYER_FACTIONS' | relative_url }})**
Player-created organizations:
- Creation and management
- Treasury system
- Ranks and permissions
- Member management

**[Territory Control]({{ '/TERRITORY_CONTROL' | relative_url }})**
System ownership mechanics:
- Claiming systems
- Passive income generation
- Territory conflicts
- Control timers

**[Player Trading]({{ '/PLAYER_TRADING' | relative_url }})**
P2P trading system:
- Trade initiation
- Escrow mechanics
- Item/credit transfers
- Trade history

**[PvP Combat]({{ '/PVP_COMBAT' | relative_url }})**
Player vs player combat:
- Consensual duels
- Faction warfare
- Combat balance
- Rewards/penalties

**[Leaderboards]({{ '/LEADERBOARDS' | relative_url }})**
Competitive rankings:
- 8 categories
- Real-time updates
- Seasonal boards
- Weekly/monthly/all-time

**[Player Presence]({{ '/PLAYER_PRESENCE' | relative_url }})**
Real-time player tracking:
- Online status
- Location updates
- Timeout handling

### Advanced Systems

**[Outfitter System]({{ '/OUTFITTER_SYSTEM' | relative_url }})**
Equipment and loadouts:
- 16+ equipment items
- 6 slot types
- Loadout management
- Stats recalculation

**[Settings System]({{ '/SETTINGS_SYSTEM' | relative_url }})**
Player preferences:
- 6 setting categories
- 5 color schemes
- JSON persistence
- Default resets

**[Tutorial System]({{ '/TUTORIAL_SYSTEM' | relative_url }})**
Onboarding and help:
- 7 tutorial categories
- 20+ steps
- Context-sensitive help
- Progress tracking

**[Admin System]({{ '/ADMIN_SYSTEM' | relative_url }})**
Server administration:
- RBAC (4 roles, 20+ permissions)
- Moderation tools
- Audit logging
- Server settings

**[Inventory System]({{ '/INVENTORY_SYSTEM_SPEC' | relative_url }})**
UUID-based item tracking:
- Hybrid commodity/unique item system
- Location tracking
- Transfer audit
- Batch operations

**[Inventory System Complete]({{ '/INVENTORY_SYSTEM_COMPLETE' | relative_url }})**
Implementation summary and verification.

**[Inventory Verification]({{ '/INVENTORY_SYSTEM_VERIFICATION' | relative_url }})**
Testing and validation of inventory system.

**[Marketplace Forms]({{ '/MARKETPLACE_FORMS_COMPLETE' | relative_url }})**
Player marketplace UI:
- Auction creation
- Contract posting
- Bounty submission
- Form validation

---

## 🛠️ Operations & Infrastructure

### Monitoring & Metrics

**[Metrics & Monitoring]({{ '/METRICS_MONITORING' | relative_url }})**
Production observability:
- Prometheus metrics endpoint
- HTML stats dashboard
- Enhanced metrics (p50/p95/p99)
- Performance profiling
- Health checks

**Metrics Endpoints**:
- `/metrics` - Prometheus-compatible format
- `/stats` - Human-readable HTML dashboard
- `/stats/enhanced` - Latency & error tracking
- `/stats/performance` - Performance profiling
- `/health` - Health status (JSON)

**Tracked Metrics**:
- Connection metrics (total, active, failed, duration)
- Player metrics (active, logins, registrations, peak)
- Game activity (trades, combat, missions, quests)
- Economy (credits, market volume, trade volume)
- Database (queries, errors, connection pool)
- System (uptime, goroutines, memory)

### Security & Protection

**[Rate Limiting]({{ '/RATE_LIMITING' | relative_url }})**
Connection and authentication protection:
- Connection limits (5 concurrent, 20/min per IP)
- Auth rate limiting (5 attempts = 15min lockout)
- Auto-ban system (20 failures = 24h ban)
- Per-IP tracking
- Brute force protection

**[Security Audit]({{ '/SECURITY_AUDIT' | relative_url }})**
Comprehensive security analysis:
- Security rating: 9.5/10
- Vulnerability assessment
- 61 critical bugs fixed
- Security best practices
- Remediation tracking

**Security Features**:
- Persistent SSH host keys (MITM prevention)
- Password complexity requirements
- Username validation with regex
- Two-factor authentication (TOTP)
- Password reset system
- Input sanitization
- Audit logging

### Backup & Recovery

**[Backup & Restore]({{ '/BACKUP_RESTORE' | relative_url }})**
Automated backup system:
- Compression with gzip
- Retention policies (days and count)
- Automatic cleanup
- Safe restore with prompts
- Progress tracking
- Cron integration

**Backup Features**:
- Manual backups: `./scripts/backup.sh`
- Automated via cron
- Configurable retention
- List available backups
- Point-in-time recovery

### Database

**Schema**: `scripts/schema.sql`
Complete database schema with 30+ tables.

**Migrations**: `scripts/migrations/`
Database migration system:
- Version tracking
- Up/down migrations
- Migration status
- Rollback support

**Migration Commands**:
```bash
./scripts/migrate.sh status   # Check status
./scripts/migrate.sh up       # Apply pending
./scripts/migrate.sh down     # Rollback
./scripts/migrate.sh reset    # Reset all (DANGEROUS)
```

---

## 🧪 Testing & Quality

### Test Coverage

**[Load Testing Report]({{ '/LOAD_TESTING_REPORT' | relative_url }})**
Performance testing results:
- Concurrent user testing
- Database performance
- Memory profiling
- Bottleneck identification

**Test Suite**:
- **56 TUI Tests** (17 integration, 39 unit) - All passing
- **15 Regression Tests** - Bug fix verification
- **12 Marketplace Form Tests** - Form validation
- **14 Inventory Component Tests** - ItemPicker, ItemList
- **100+ Total Tests** - ~70% coverage

**Testing Tools**:
- Go testing framework
- Race detector (`go test -race`)
- Coverage reporting (`make coverage`)
- Load testing tool for 1000+ items

### Code Quality

**Linting**: `.golangci.yml`
Configuration for golangci-lint with:
- Enabled linters: errcheck, gosimple, govet, ineffassign, staticcheck, gofmt, goimports, misspell, dupl
- Disabled linters with reasoning
- TUI file exemptions
- Test file relaxed requirements

**Commands**:
```bash
make lint      # Run linters
make fmt       # Format code
make vet       # Run go vet
make test      # Run tests with race detector
make coverage  # View coverage in browser
```

---

## 📦 Deployment

### Docker

**[Docker Documentation]({{ '/DOCKER' | relative_url }})**
Containerized deployment:
- Docker Compose setup
- PostgreSQL configuration
- Environment variables
- Volume management
- Networking

**Commands**:
```bash
docker compose up -d        # Start stack
docker compose down         # Stop stack
docker compose logs -f      # View logs
docker compose restart      # Restart services
make docker-clean           # Remove all artifacts
```

### Production Deployment

**Requirements**:
- Go 1.24+
- PostgreSQL 12+
- Linux/macOS server
- Minimum 2GB RAM
- 10GB disk space

**Setup Process**:
1. Clone repository
2. Install dependencies
3. Configure database
4. Generate universe
5. Create admin account
6. Configure server settings
7. Start server
8. Set up monitoring
9. Configure backups

See [Server Setup Guide]({{ '/guides/server-setup' | relative_url }}) for complete instructions.

---

## 🔧 Development

### Getting Started

**[CLAUDE.md](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/CLAUDE.md)**
Complete developer reference covering:
- Project overview
- Building and running
- Architecture patterns
- Common development tasks
- Code conventions
- Testing strategies
- Troubleshooting

**Development Commands**:
```bash
make build          # Build server binary
make run            # Run in development mode
make test           # Run tests
make lint           # Run linters
make dev-setup      # Complete dev environment setup
make watch          # Auto-rebuild on changes
```

### Contributing

**[CONTRIBUTING.md](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/CONTRIBUTING.md)**
Contribution guidelines:
- Code style
- Commit messages
- Pull request process
- Testing requirements
- Documentation standards

**[CODE_OF_CONDUCT.md](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/CODE_OF_CONDUCT.md)**
Community standards and expectations.

### API Documentation

**[API.md]({{ '/API' | relative_url }})**
Internal API documentation:
- Repository interfaces
- Manager APIs
- Service contracts
- Data models

---

## 📚 Reference Materials

### Implementation Status

**[Implementation Status]({{ '/IMPLEMENTATION_STATUS' | relative_url }})**
Historical tracking of feature implementation across phases.

**[Phase 1 Status]({{ '/PHASE_1_STATUS' | relative_url }})**
Detailed Phase 1 implementation breakdown.

**[Phase 1B Example]({{ '/PHASE_1B_EXAMPLE' | relative_url }})**
Example implementation from Phase 1B.

### UI Documentation

**[UI Prototypes]({{ '/UI_PROTOTYPES' | relative_url }})**
TUI screen designs and prototypes:
- Screen layouts
- Navigation flows
- Component designs
- Styling guidelines

**[Screen Navigation Test]({{ '/SCREEN_NAVIGATION_TEST' | relative_url }})**
Testing screen navigation and transitions.

### GitHub & CI/CD

**[GitHub Setup]({{ '/GITHUB_SETUP' | relative_url }})**
Repository setup and CI/CD:
- GitHub Actions workflows
- Branch protection
- Release process
- Issue templates

---

## 🗂️ Documentation by Category

### Architecture
- [Architecture Refactoring]({{ '/ARCHITECTURE_REFACTORING' | relative_url }})
- [API Integration Plan]({{ '/API_INTEGRATION_PLAN' | relative_url }})
- [API Migration Guide]({{ '/API_MIGRATION_GUIDE' | relative_url }})
- [API Documentation]({{ '/API' | relative_url }})

### Game Design
- [Economy & Balance]({{ '/ECONOMY_BALANCE' | relative_url }})
- [Universe Design]({{ '/UNIVERSE_DESIGN' | relative_url }})
- [NPC Factions]({{ '/NPC_FACTIONS_SUMMARY' | relative_url }})
- [Faction Relations]({{ '/FACTION_RELATIONS' | relative_url }})

### Game Systems (14 docs)
- [Chat System]({{ '/CHAT_SYSTEM' | relative_url }})
- [Player Factions]({{ '/PLAYER_FACTIONS' | relative_url }})
- [Territory Control]({{ '/TERRITORY_CONTROL' | relative_url }})
- [Player Trading]({{ '/PLAYER_TRADING' | relative_url }})
- [PvP Combat]({{ '/PVP_COMBAT' | relative_url }})
- [Leaderboards]({{ '/LEADERBOARDS' | relative_url }})
- [Player Presence]({{ '/PLAYER_PRESENCE' | relative_url }})
- [Outfitter System]({{ '/OUTFITTER_SYSTEM' | relative_url }})
- [Settings System]({{ '/SETTINGS_SYSTEM' | relative_url }})
- [Admin System]({{ '/ADMIN_SYSTEM' | relative_url }})
- [Tutorial System]({{ '/TUTORIAL_SYSTEM' | relative_url }})
- [Inventory System Spec]({{ '/INVENTORY_SYSTEM_SPEC' | relative_url }})
- [Inventory Complete]({{ '/INVENTORY_SYSTEM_COMPLETE' | relative_url }})
- [Marketplace Forms]({{ '/MARKETPLACE_FORMS_COMPLETE' | relative_url }})

### Operations
- [Metrics & Monitoring]({{ '/METRICS_MONITORING' | relative_url }})
- [Rate Limiting]({{ '/RATE_LIMITING' | relative_url }})
- [Backup & Restore]({{ '/BACKUP_RESTORE' | relative_url }})
- [Security Audit]({{ '/SECURITY_AUDIT' | relative_url }})
- [Load Testing Report]({{ '/LOAD_TESTING_REPORT' | relative_url }})

### Development
- [CLAUDE.md](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/CLAUDE.md)
- [Docker Documentation]({{ '/DOCKER' | relative_url }})
- [GitHub Setup]({{ '/GITHUB_SETUP' | relative_url }})
- [UI Prototypes]({{ '/UI_PROTOTYPES' | relative_url }})

---

## 🔗 External Resources

- **[GitHub Repository](https://github.com/JoshuaAFerguson/terminal-velocity)** - Source code
- **[GitHub Wiki](https://github.com/JoshuaAFerguson/terminal-velocity/wiki)** - Community wiki
- **[Discussions](https://github.com/JoshuaAFerguson/terminal-velocity/discussions)** - Q&A and community
- **[Issues](https://github.com/JoshuaAFerguson/terminal-velocity/issues)** - Bug reports and features

---

<div class="footer-cta">
  <h2>Need Help?</h2>
  <p>Can't find what you're looking for?</p>
  <a href="https://github.com/JoshuaAFerguson/terminal-velocity/discussions" class="cta-button">Ask the Community</a>
  <a href="https://github.com/JoshuaAFerguson/terminal-velocity/issues" class="cta-button">Report an Issue</a>
</div>

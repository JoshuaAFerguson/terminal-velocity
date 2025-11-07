# Changelog

All notable changes to Terminal Velocity will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Enhanced Authentication System (Phase 1, Issue #2 partial)**:
  - SSH public key authentication support
  - Multi-method authentication (password + SSH keys simultaneously)
  - Email field for player accounts
  - SSH key repository with fingerprint tracking
  - Support for multiple SSH keys per account
  - SSH key management (add, remove, deactivate, track last used)
  - Password-optional accounts (SSH-key-only authentication)
  - Account management CLI tool (`accounts`)
  - Server configuration for auth methods and registration
  - Placeholder for interactive registration (coming soon)
- **Database Layer (Phase 1, Issue #1)**:
  - PostgreSQL connection management with pgx driver
  - Connection pooling with configurable limits
  - Transaction support with automatic rollback
  - Player repository (registration, authentication, CRUD, credits, reputation)
  - System repository (systems, planets, jump routes, bulk operations)
  - Migration runner for schema initialization
  - Integration tests for player repository
  - Server integration with database authentication
- Initial project structure and repository setup
- Comprehensive GitHub templates:
  - Bug report template with game state tracking
  - Feature request template
  - Enhancement template for roadmap items
  - Documentation template
  - Deployment/configuration template
  - Question/help template
  - Enhanced pull request template
  - Template configuration with helpful links
- Universe generation system:
  - Procedural system and planet generation
  - 6 NPC factions with territory distribution
  - MST-based jump route connectivity
  - Name generation system (Greek+Constellation, real stars, catalog numbers)
  - Comprehensive test coverage (100% connectivity verification)
- CLI tools:
  - `genmap` - Universe preview and statistics tool
- Data models:
  - Player (progression, credits, reputation)
  - Ship (types, cargo, weapons, outfits)
  - Universe (systems, planets, jump routes)
  - Factions (player and NPC factions)
  - Trading (commodities, market data)
  - Missions (framework for mission system)
  - NPC Factions (6 standard factions: UEF, ROM, FTG, FWA, Crimson, Auroran)
- Database schema:
  - PostgreSQL schema with 20+ tables
  - Player, ship, system, planet, faction tables
  - Trading and mission tables
  - Indexes and constraints
- SSH server:
  - Basic SSH server implementation
  - Database-backed authentication with bcrypt
  - Session handling with player tracking
  - Host key generation
  - Graceful shutdown support
  - Online status tracking
- Docker support:
  - Multi-stage Dockerfile
  - docker-compose.yml with PostgreSQL and PgAdmin
  - .dockerignore optimization
  - .env.example template
  - Health checks and auto-initialization
- Documentation:
  - README.md with quickstart and features
  - ROADMAP.md with 8 development phases
  - UNIVERSE_DESIGN.md (galaxy structure and factions)
  - FACTION_RELATIONS.md (politics and conflicts)
  - DOCKER.md (comprehensive deployment guide)
  - IMPLEMENTATION_STATUS.md (development tracking)
  - CODE_OF_CONDUCT.md (Contributor Covenant v2.1)
  - SECURITY.md (vulnerability reporting)
  - CONTRIBUTING.md (contribution guidelines)
- GitHub project management:
  - 8 GitHub projects (one per development phase)
  - 41 issues organized by phase
  - 25+ labels (phase labels, category labels)
  - Issue and PR templates
- CI/CD:
  - GitHub Actions workflow for testing and building
  - Docker workflow for multi-platform builds (amd64, arm64)
  - Automated publishing to GitHub Container Registry
  - Branch protection rules
- Makefile with common development tasks
- Go module setup with dependencies
- **Trading System (Phase 2, Issue #7)**:
  - Interactive trading UI with market view
  - Buy/sell transactions with database persistence
  - Real-time market price updates based on supply/demand
  - Commodity filtering by tech level
  - Stock validation and credit checks
  - Transaction rollback on errors
- **Cargo Management (Phase 2, Issue #8)**:
  - Cargo hold visualization
  - Jettison functionality with quantity control
  - Real-time space calculations
  - Ship type integration
  - Sorted cargo display
- **Economy Balance (Phase 2, Issue #9)**:
  - Comprehensive ECONOMY_BALANCE.md documentation
  - 5 profitable trade routes with ROI analysis
  - Contraband price increases (20-50%)
  - Improved tech level modifiers (0.05 → 0.07)
  - Ship progression economics documented
- **Ship Types (Phase 3, Issue #10)**:
  - 11 standard ship types (Shuttle to Battleship)
  - Complete stat definitions (combat, cargo, speed, etc.)
  - Combat rating requirements
  - Class-based progression system
- **Shipyard System (Phase 3, Issue #11)**:
  - Ship browsing with affordability checking
  - Purchase functionality with validation
  - Trade-in system (70% value)
  - Ship details viewer
  - Combat rating enforcement
  - Database persistence
- **Outfitter System (Phase 3, Issue #12)**:
  - 9 weapon types (laser, missile, plasma, railgun)
  - 15 outfit types (shields, hull, cargo, fuel, engines)
  - Tab-based navigation interface
  - Equipment installation with space validation
  - Equipment removal with 50% refund
  - Real-time ship stats with bonuses
  - Transaction safety with rollback
- **Ship Comparison Tools (Phase 3, Issue #13)**:
  - Side-by-side ship comparison view
  - Visual statistics bars for all stats
  - Cost-benefit analysis with trade-in calculations
  - Performance ratings (combat, trading, speed, overall)
  - Star-based rating display (0-10 stars)
  - Upgrade path recommendations
  - Value-per-credit analysis
  - Specific recommendations (combat vs trading focus)
- **Ship Management Screens (Phase 3, Issue #14)**:
  - Ship inventory view with all owned ships
  - Active ship indicator and selection
  - Ship switching functionality
  - Ship renaming with input validation
  - Configuration viewer (weapons, outfits, cargo)
  - Hull/shield/fuel status display
  - Maintenance status indicators
  - Comprehensive ship details view
- **Weapon Systems (Phase 4, Issue #15)**:
  - Extended weapon model with combat mechanics
  - Weapon type differentiation (laser, missile, plasma, railgun)
  - Complete weapon stat definitions (damage, cooldown, energy, ammo, etc.)
  - Damage calculation system with shield/hull damage split
  - Range and accuracy systems with distance penalties
  - Ammo tracking for missile weapons (capacity and consumption)
  - Energy cost tracking for energy weapons
  - Weapon cooldown system (time between shots)
  - Shield penetration mechanics (railguns bypass shields)
  - Critical hit system (10% chance for 1.5x damage)
  - Hit chance calculation (accuracy, distance, evasion)
  - Projectile speed differentiation
  - DPS calculation utilities
  - Weapon state tracking (ammo, cooldowns, last fired turn)
  - Ammo reload functionality
- **Combat AI (Phase 4, Issue #16)**:
  - 5 AI difficulty levels (Easy, Medium, Hard, Expert, Ace)
  - Level-based attributes (aggression, accuracy, reaction time)
  - Intelligent target selection with threat assessment
  - Multi-factor targeting (hull/shield damage, threat level, distance)
  - Weapon usage strategies (range optimization, ammo conservation)
  - Shield penetration awareness in weapon selection
  - Strategic missile usage (save for weakened targets)
  - Evasion patterns based on ship condition
  - Retreat conditions with morale system
  - Level-dependent retreat thresholds
  - Formation flying framework
  - Position maintenance in formations
  - AI action priority system
  - Morale system (affected by hull damage)
  - Reaction time delays (0.25s-2.0s based on level)
  - Accuracy modifiers per difficulty level
- **Combat UI (Phase 4, Issue #17)**:
  - Full-screen tactical display with turn-based combat
  - Real-time ship status displays (hull/shields with percentage bars)
  - Target selection interface with hull/shield status
  - Weapon selection and control panel
  - Weapon status display (damage, range, cooldown, ammo)
  - Shield and armor indicators with visual bars
  - ASCII tactical radar (20x20 grid)
  - Radar zoom levels (1x-5x)
  - Enemy position tracking on radar (P=Player, E=Enemy, T=Target)
  - Combat log with scrolling messages (last 10 messages)
  - Turn number tracking
  - Player/enemy turn system
  - Fire weapon command (F key)
  - End turn command (E key)
  - Shield regeneration per turn
  - Weapon cooldown visualization
  - Ammo tracking display for missiles
  - Victory detection (all enemies destroyed)
  - Weapon firing with full damage integration
  - Target destruction and removal
  - Three viewing modes (tactical, target_select, weapons)
- **Reputation System (Phase 4, Issue #18)**:
  - ReputationChange tracking with faction cascade effects
  - 8 reputation event types (kill hostile/neutral/ally/civilian, defend, piracy, bounty)
  - CalculateCombatReputation() with multi-faction relationship system
  - Bounty system with expiration tracking
  - BountyInfo structure (amount, reason, expiration)
  - Legal status tracking (clean, offender, wanted, fugitive)
  - LegalStatus with crime count and severity tracking
  - Bounty amount calculation based on ship value and crime severity
  - Pay-off system (1.5x bounty cost for bribes)
  - Faction reinforcement logic based on reputation and territory
  - WillFactionsReinforce() with turn-based arrival system
  - Reinforcement strength calculation (1-5 ships based on patrol strength)
  - Reinforcement delay based on faction capabilities (2-5 turns)
  - Hostility level calculation (allied/friendly/neutral/unfriendly/hostile/at_war)
  - ApplyReputationChanges() with -100 to +100 clamping
  - Active bounty tracking across multiple factions
  - GetReputationChangeMessage() for player feedback
- **Loot and Salvage System (Phase 4, Issue #19)**:
  - LootDrop structure for combat rewards
  - GenerateLoot() with dynamic reward calculation
  - Ship destruction drop system (10-20% of ship value)
  - Cargo recovery mechanics (30-60% survival rate)
  - Outfit salvaging (40% chance per outfit)
  - Weapon salvaging (30-45% chance, higher for hostile ships)
  - Credit rewards including bounty payouts
  - Rare item drop system with 4 rarity tiers:
    - Legendary (5% chance): Ancient artifacts worth 250K
    - Epic (15% chance): Prototype components, neural processors (75-100K)
    - Rare (30% chance): Military plans, jump drive data (50-60K)
    - Uncommon (50% chance): Fusion cores (25K)
  - 6 standard rare items (artifacts, components, data, contraband)
  - Rarity-weighted drop system based on ship class and hostility
  - Cargo space calculation for loot
  - CanCarryLoot() validation before salvage
  - ApplyLoot() with automatic cargo integration
  - Salvage time calculation (2-5 turns based on loot complexity)
  - Total loot value tracking
  - Formatted loot summaries with credit shorthand (K/M)
  - SalvageSpecificItem() for targeted salvage attempts

### Changed
- Module path corrected from github.com/s0v3r1gn to github.com/JoshuaAFerguson
- Trading tech level modifier improved (0.05 → 0.07) for better balance
- Contraband prices increased for better risk/reward balance

### Deprecated
- N/A (initial release)

### Removed
- N/A (initial release)

### Fixed
- N/A (initial release)

### Security
- Non-root Docker container user
- Multi-stage Docker builds for minimal attack surface
- No hardcoded secrets in configuration
- Bcrypt password hashing support in models

## Release History

### [0.1.0] - Unreleased
- Initial development phase
- Phase 1: Foundation & Navigation (in progress)

---

## Version Numbering

Terminal Velocity follows semantic versioning:
- **Major version** (X.0.0): Breaking changes, major new features
- **Minor version** (0.X.0): New features, backwards compatible
- **Patch version** (0.0.X): Bug fixes, minor improvements

## Development Phases

- **Phase 1**: Foundation & Navigation (in progress)
- **Phase 2**: Core Economy
- **Phase 3**: Ship Progression
- **Phase 4**: Combat System
- **Phase 5**: Missions & Progression
- **Phase 6**: Multiplayer Features
- **Phase 7**: Polish & Content
- **Phase 8**: Advanced Features

See [ROADMAP.md](ROADMAP.md) for detailed phase breakdowns.

## Links

- [GitHub Repository](https://github.com/JoshuaAFerguson/terminal-velocity)
- [Issue Tracker](https://github.com/JoshuaAFerguson/terminal-velocity/issues)
- [Project Boards](https://github.com/JoshuaAFerguson/terminal-velocity/projects)
- [Discussions](https://github.com/JoshuaAFerguson/terminal-velocity/discussions)

[Unreleased]: https://github.com/JoshuaAFerguson/terminal-velocity/compare/HEAD
[0.1.0]: https://github.com/JoshuaAFerguson/terminal-velocity/releases/tag/v0.1.0

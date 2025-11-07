# Changelog

All notable changes to Terminal Velocity will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
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

### Changed
- N/A (initial release)

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

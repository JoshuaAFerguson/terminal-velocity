---
layout: page
title: Guides
permalink: /guides/
description: Step-by-step tutorials for players, server administrators, and developers
---

# Guides & Tutorials

Comprehensive guides to help you get started with Terminal Velocity, whether you're a player, server administrator, or developer.

---

## 🎮 Player Guides

### [Getting Started]({{ '/guides/getting-started' | relative_url }})
**New to Terminal Velocity?** Start here!

Learn how to:
- Connect via SSH
- Create your account
- Navigate the game interface
- Understand basic controls
- Complete your first trade
- Get your first mission

**Difficulty**: Beginner
**Time**: 15 minutes
**Prerequisites**: SSH client installed

---

### Gameplay Guides

#### Trading Guide
Learn the art of interstellar commerce:
- Understanding commodity markets
- Finding profitable trade routes
- Tech level effects on pricing
- Supply and demand mechanics
- Advanced trading strategies

*Coming soon - See [ECONOMY_BALANCE.md]({{ '/ECONOMY_BALANCE' | relative_url }}) for detailed economics*

#### Combat Guide
Master tactical combat:
- Weapon types and effectiveness
- AI difficulty levels
- Combat tactics and strategies
- Ship outfitting for combat
- PvP combat mechanics

*Coming soon - See combat documentation in codebase*

#### Ship Progression Guide
Upgrade your way to the stars:
- Ship types and roles
- Cost-effective upgrade paths
- Equipment recommendations
- Loadout strategies
- Fleet management basics

*Coming soon*

---

## 🖥️ Server Administrator Guides

### [Server Setup]({{ '/guides/server-setup' | relative_url }})
**Want to run your own server?** Complete setup guide!

Learn how to:
- Install prerequisites (Go, PostgreSQL)
- Configure the database
- Generate the universe
- Set up user accounts
- Configure server settings
- Run and monitor the server

**Difficulty**: Intermediate
**Time**: 30-60 minutes
**Prerequisites**: Linux/macOS system, root/sudo access

---

### Administration Guides

#### Admin Tools Guide
Master server administration:
- Using the admin panel
- Role-based access control (RBAC)
- Player moderation (ban/mute)
- Viewing audit logs
- Server configuration

See: [Admin System Documentation]({{ '/ADMIN_SYSTEM' | relative_url }})

#### Monitoring & Metrics Guide
Keep your server healthy:
- Prometheus metrics endpoint
- Stats dashboard overview
- Performance profiling
- Health checks
- Alerting setup

See: [Metrics & Monitoring Documentation]({{ '/METRICS_MONITORING' | relative_url }})

#### Backup & Recovery Guide
Protect your data:
- Automated backup setup
- Retention policies
- Restoring from backup
- Disaster recovery
- Cron job configuration

See: [Backup & Restore Documentation]({{ '/BACKUP_RESTORE' | relative_url }})

#### Security Guide
Secure your server:
- Rate limiting configuration
- Authentication best practices
- Auto-ban system
- 2FA setup
- Security auditing

See: [Rate Limiting Documentation]({{ '/RATE_LIMITING' | relative_url }})

---

## 👨‍💻 Developer Guides

### Development Setup
Set up your development environment:
- Clone the repository
- Install dependencies
- Database setup for development
- Running tests
- Code formatting and linting

See: [CLAUDE.md](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/CLAUDE.md) - Complete development guide

---

### Contributing Guides

#### Contributing Guide
Join the development effort:
- Code style guidelines
- Commit message format
- Pull request process
- Testing requirements
- Documentation standards

See: [CONTRIBUTING.md](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/CONTRIBUTING.md)

#### Adding Features
Extend Terminal Velocity:
- Adding a new TUI screen
- Creating a new game system
- Database migrations
- Writing tests
- Integrating with existing systems

See: **Common Development Tasks** in [CLAUDE.md](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/CLAUDE.md)

#### Testing Guide
Ensure quality:
- Running the test suite
- Writing TUI tests
- Integration testing
- Race condition testing
- Coverage reporting

See: **Testing** section in [CLAUDE.md](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/CLAUDE.md)

---

## 📚 Reference Documentation

### Quick References

- **[Features Catalog]({{ '/features' | relative_url }})** - All 245+ features
- **[Technical Documentation]({{ '/documentation' | relative_url }})** - Architecture, API, systems
- **[Roadmap](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/ROADMAP.md)** - Development history
- **[Changelog](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/CHANGELOG.md)** - Version history

### System Documentation

- [Chat System]({{ '/CHAT_SYSTEM' | relative_url }})
- [Player Factions]({{ '/PLAYER_FACTIONS' | relative_url }})
- [Territory Control]({{ '/TERRITORY_CONTROL' | relative_url }})
- [Player Trading]({{ '/PLAYER_TRADING' | relative_url }})
- [PvP Combat]({{ '/PVP_COMBAT' | relative_url }})
- [Leaderboards]({{ '/LEADERBOARDS' | relative_url }})
- [Player Presence]({{ '/PLAYER_PRESENCE' | relative_url }})
- [Outfitter System]({{ '/OUTFITTER_SYSTEM' | relative_url }})
- [Settings System]({{ '/SETTINGS_SYSTEM' | relative_url }})
- [Tutorial System]({{ '/TUTORIAL_SYSTEM' | relative_url }})

---

## 🎯 Learning Paths

### New Player Path
1. [Getting Started]({{ '/guides/getting-started' | relative_url }}) - Connect and create account
2. In-game Tutorial - Learn the basics (7 categories)
3. Trading Guide - Make your first fortune
4. Ship Progression Guide - Upgrade your ship
5. Combat Guide - Defend yourself
6. Multiplayer Features - Join the community

### Server Admin Path
1. [Server Setup]({{ '/guides/server-setup' | relative_url }}) - Get running
2. Admin Tools Guide - Learn moderation
3. Monitoring Guide - Track performance
4. Backup Guide - Protect data
5. Security Guide - Harden your server

### Developer Path
1. Development Setup - Clone and build
2. [CLAUDE.md](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/CLAUDE.md) - Complete dev guide
3. Architecture Documentation - Understand the design
4. Contributing Guide - Submit your first PR
5. Adding Features - Build something new

---

## 🆘 Troubleshooting

### Common Issues

**Connection Problems**:
- Verify SSH port (default 2222)
- Check firewall settings
- Ensure server is running
- Test with `ssh -vvv` for debug output

**Database Issues**:
- Verify PostgreSQL is running
- Check connection credentials
- Run schema migrations
- Review database logs

**Performance Issues**:
- Check metrics dashboard
- Review database indexes
- Monitor connection pool
- Analyze slow queries

**Build Errors**:
- Ensure Go 1.24+ is installed
- Run `go mod download`
- Check for import errors
- Verify all dependencies

See [CLAUDE.md - Troubleshooting](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/CLAUDE.md#troubleshooting) for detailed solutions.

---

## 📞 Getting Help

**Community Support**:
- [GitHub Discussions](https://github.com/JoshuaAFerguson/terminal-velocity/discussions) - Q&A, ideas, show & tell
- [GitHub Issues](https://github.com/JoshuaAFerguson/terminal-velocity/issues) - Bug reports, feature requests

**Documentation**:
- [Technical Documentation]({{ '/documentation' | relative_url }})
- [Features Catalog]({{ '/features' | relative_url }})
- [Project README](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/README.md)

**Development**:
- [CLAUDE.md](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/CLAUDE.md) - Complete developer reference
- [Contributing Guide](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/CONTRIBUTING.md)

---

## 📝 Guide Status

| Guide | Status | Last Updated |
|-------|--------|--------------|
| Getting Started | ✅ Available | 2025-11-15 |
| Server Setup | ✅ Available | 2025-11-15 |
| Trading Guide | 📝 Planned | - |
| Combat Guide | 📝 Planned | - |
| Ship Progression | 📝 Planned | - |
| Admin Tools | 📖 See Docs | 2025-11-15 |
| Monitoring | 📖 See Docs | 2025-11-15 |
| Backup & Recovery | 📖 See Docs | 2025-11-15 |
| Security | 📖 See Docs | 2025-11-15 |
| Development Setup | 📖 See CLAUDE.md | 2025-11-15 |
| Contributing | 📖 See CONTRIBUTING.md | 2025-11-15 |
| Adding Features | 📖 See CLAUDE.md | 2025-11-15 |

---

<div class="footer-cta">
  <h2>Ready to Get Started?</h2>
  <p>Choose your path and dive into Terminal Velocity!</p>
  <a href="{{ '/guides/getting-started' | relative_url }}" class="cta-button">Start Playing</a>
  <a href="{{ '/guides/server-setup' | relative_url }}" class="cta-button">Run a Server</a>
</div>

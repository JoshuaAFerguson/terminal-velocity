---
layout: home
title: Terminal Velocity
---

# Welcome to Terminal Velocity

A multiplayer space trading and combat game inspired by **Escape Velocity**, playable entirely through **SSH**. Navigate a persistent universe, trade commodities, upgrade ships, engage in combat, and form factions—all within a terminal UI built with BubbleTea.

<div class="hero-section">
  <div class="hero-content">
    <h2>🚀 Play via SSH</h2>
    <pre><code>ssh -p 2222 username@terminalvelocity.game</code></pre>
  </div>
</div>

## ✨ Key Features

<div class="feature-grid">
  <div class="feature-card">
    <h3>🌌 Persistent Universe</h3>
    <p>Explore 100+ star systems with procedurally generated trade routes, planets, and economies.</p>
  </div>

  <div class="feature-card">
    <h3>💰 Dynamic Economy</h3>
    <p>Trade 15 commodities across systems with supply/demand pricing. Build your fortune!</p>
  </div>

  <div class="feature-card">
    <h3>⚔️ Combat System</h3>
    <p>Turn-based tactical combat with AI opponents. Upgrade your ship and dominate the battlefield.</p>
  </div>

  <div class="feature-card">
    <h3>🏛️ Player Factions</h3>
    <p>Form organizations with shared treasuries, territory control, and faction warfare.</p>
  </div>

  <div class="feature-card">
    <h3>🗨️ Multiplayer Chat</h3>
    <p>5 chat channels including global, system-local, faction, direct messages, and trade.</p>
  </div>

  <div class="feature-card">
    <h3>🎯 Missions & Quests</h3>
    <p>Dynamic storylines, branching narratives, and procedurally generated missions.</p>
  </div>
</div>

## 📊 Project Status

**Current Version**: 0.20.0
**Status**: ✅ **All 20 Phases Complete** - Production Ready
**Lines of Code**: 78,002
**UI Screens**: 41
**Features**: 245+

### Recent Achievements

- ✅ **Phase 20**: Production infrastructure (metrics, rate limiting, backups)
- ✅ **Phase 19**: Tutorial system with context-sensitive help
- ✅ **Phase 18**: Server administration with RBAC and audit logging
- ✅ **Phase 17**: Settings system with 5 color schemes
- ✅ **Phase 16**: Enhanced outfitter with loadout management

[View Complete Roadmap →](/roadmap)

## 🎮 Quick Start

### Playing the Game

1. **Install SSH Client** (most systems have it built-in)
2. **Connect to Server**:
   ```bash
   ssh -p 2222 username@terminalvelocity.game
   ```
3. **Create Account** (if registration is enabled) or contact admin
4. **Start Playing!** Navigate with arrow keys, select with Enter

[Full Getting Started Guide →](/guides/getting-started)

### Running Your Own Server

```bash
# Clone repository
git clone https://github.com/JoshuaAFerguson/terminal-velocity.git
cd terminal-velocity

# Quick setup (uses initialization script)
./scripts/init-server.sh

# Or manual setup
make dev-setup
make build
make run
```

[Server Setup Guide →](/guides/server-setup)

## 📚 Documentation

<div class="doc-links">
  <a href="/features" class="doc-link">
    <strong>Features</strong>
    <p>Explore all 245+ features across 20 development phases</p>
  </a>

  <a href="/guides" class="doc-link">
    <strong>Guides</strong>
    <p>Step-by-step tutorials for players and developers</p>
  </a>

  <a href="/documentation" class="doc-link">
    <strong>Technical Docs</strong>
    <p>API references, architecture, and system design</p>
  </a>

  <a href="https://github.com/JoshuaAFerguson/terminal-velocity" class="doc-link">
    <strong>GitHub Repository</strong>
    <p>Source code, issues, and contribution guidelines</p>
  </a>
</div>

## 🛠️ Tech Stack

- **Language**: Go 1.24+
- **Database**: PostgreSQL (pgx/v5)
- **TUI Framework**: BubbleTea + Lipgloss
- **SSH Server**: golang.org/x/crypto/ssh
- **Testing**: Go testing + testify
- **CI/CD**: GitHub Actions

## 🌟 Highlights

### 245+ Features Across 20 Phases

- **Core Gameplay** (Phases 0-8): Trading, combat, ships, reputation, missions
- **Multiplayer** (Phases 9-15): Chat, factions, territory, P2P trading, PvP, leaderboards
- **Polish** (Phases 16-19): Outfitting, settings, admin, tutorials
- **Production** (Phase 20): Metrics, rate limiting, backups

### Production-Ready Infrastructure

- **Metrics**: Prometheus-compatible endpoint + HTML stats dashboard
- **Security**: Rate limiting, auto-ban system, RBAC with 20+ permissions
- **Operations**: Automated backups with retention policies
- **Monitoring**: Health checks, performance tracking, audit logging

## 📈 Statistics

<div class="stats-grid">
  <div class="stat">
    <div class="stat-number">78,002</div>
    <div class="stat-label">Lines of Code</div>
  </div>

  <div class="stat">
    <div class="stat-number">245+</div>
    <div class="stat-label">Features</div>
  </div>

  <div class="stat">
    <div class="stat-number">41</div>
    <div class="stat-label">UI Screens</div>
  </div>

  <div class="stat">
    <div class="stat-number">20</div>
    <div class="stat-label">Phases Complete</div>
  </div>
</div>

## 🤝 Contributing

Terminal Velocity is an open-source project. Contributions are welcome!

- **Report Bugs**: [GitHub Issues](https://github.com/JoshuaAFerguson/terminal-velocity/issues)
- **Suggest Features**: [Discussions](https://github.com/JoshuaAFerguson/terminal-velocity/discussions)
- **Submit PRs**: See [Contributing Guide](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/CONTRIBUTING.md)

## 📜 License

Terminal Velocity is licensed under the MIT License. See [LICENSE](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/LICENSE) for details.

## 📞 Contact

- **GitHub**: [@JoshuaAFerguson](https://github.com/JoshuaAFerguson)
- **Repository**: [terminal-velocity](https://github.com/JoshuaAFerguson/terminal-velocity)
- **Issues**: [Report a bug or request a feature](https://github.com/JoshuaAFerguson/terminal-velocity/issues)

---

<div class="footer-cta">
  <h2>Ready to Launch?</h2>
  <p>Join the universe and start your adventure today!</p>
  <pre><code>ssh -p 2222 username@terminalvelocity.game</code></pre>
</div>

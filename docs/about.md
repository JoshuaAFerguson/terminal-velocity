---
layout: page
title: About
permalink: /about/
description: Project history, motivation, technology stack, and development journey
---

# About Terminal Velocity

Terminal Velocity is a feature-rich multiplayer space trading and combat game inspired by the classic **Escape Velocity** series, designed to be played entirely through **SSH** with a beautiful terminal-based interface.

---

## 🚀 Project Vision

**Mission**: Create a comprehensive, production-ready multiplayer space game that captures the spirit of classic space trading games while leveraging modern technology and the unique charm of terminal interfaces.

**Philosophy**:
- **Accessibility**: Play from anywhere with just an SSH client
- **Depth**: Rich, interconnected game systems
- **Community**: Multiplayer-first design with social features
- **Quality**: Production-ready infrastructure and security
- **Open Source**: Transparent development, community-driven

---

## 📖 Project History

### The Beginning (2024)

Terminal Velocity began as an exploration of what's possible with SSH-based gaming. Inspired by the classic **Escape Velocity** series by Ambrosia Software and the resurgence of terminal-based applications, the project aimed to prove that rich, engaging multiplayer experiences could exist entirely in the terminal.

### Development Journey (20 Phases)

**Phase 0-1** (Foundation): SSH server, database integration, universe generation
**Phase 2-3** (Core Gameplay): Trading economy, ship progression
**Phase 4-5** (Content): Combat system, missions, quests, events
**Phase 6-8** (Multiplayer): Chat, factions, territory, PvP, leaderboards
**Phase 9-15** (Advanced Features): Social systems, marketplace, fleet management, mining
**Phase 16-18** (Deep Systems): Manufacturing, crafting, stations, competitive arenas
**Phase 19-20** (Production Ready): QoL improvements, security hardening, infrastructure

### Milestones

- **M1**: Playable Prototype (Phases 0-4) ✅
- **M1.5**: Single-Player Complete (Phases 5-7) ✅
- **M2**: Multiplayer Functional (Phase 8) ✅
- **M3**: Advanced Features (Phases 9-16) ✅
- **M4**: Production Infrastructure (Phases 17-20) ✅
- **M5**: Public Launch (Beta testing → Production) 🎯

---

## 💡 Motivation

### Why Terminal-Based?

**Simplicity**: No graphics engine, no complex rendering. Just connect via SSH and play.

**Accessibility**: Works on any device with an SSH client - from Raspberry Pi to high-end workstation, from Linux to Windows to macOS.

**Nostalgia**: Captures the spirit of early text-based games while incorporating modern game design.

**Technical Challenge**: Demonstrates that compelling multiplayer experiences can be built with minimal client requirements.

**Developer Friendly**: Terminal UI makes it easy to iterate, debug, and extend. No asset pipelines, no shader compilation, just code.

### Why Escape Velocity?

The Escape Velocity series (1996-2002) represented the pinnacle of single-player space trading games:
- Rich universe with hundreds of systems
- Deep ship progression and customization
- Engaging storylines with branching narratives
- Modding community that kept games alive for decades

Terminal Velocity aims to capture that magic while adding:
- **Multiplayer**: Play with friends, form factions, compete on leaderboards
- **Modern Infrastructure**: Cloud-ready, scalable, production-grade
- **Active Development**: Regular updates, community involvement, transparent roadmap

---

## 🛠️ Technology Stack

### Core Technologies

**Language**: [Go 1.24+](https://go.dev/)
Chosen for:
- Excellent concurrency support (goroutines, channels)
- Strong type system with interfaces
- Fast compilation and excellent tooling
- Great standard library for SSH, networking, and HTTP
- Simple deployment (single binary)

**Database**: [PostgreSQL](https://www.postgresql.org/) with [pgx/v5](https://github.com/jackc/pgx)
Chosen for:
- Rock-solid reliability and ACID compliance
- Excellent performance for game data
- Rich feature set (JSONB, arrays, full-text search)
- Best-in-class connection pooling with pgx
- Mature ecosystem and tooling

**TUI Framework**: [BubbleTea](https://github.com/charmbracelet/bubbletea) + [Lipgloss](https://github.com/charmbracelet/lipgloss)
Chosen for:
- Elm-inspired architecture (Model-View-Update)
- Clean separation of concerns
- Excellent async handling via commands
- Beautiful styling with Lipgloss
- Active development and community

**SSH Server**: [golang.org/x/crypto/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh)
Chosen for:
- Official Go implementation
- Multi-method authentication (password + public key)
- Channel-based I/O perfect for BubbleTea
- Solid security and compliance

### Infrastructure

**Containerization**: Docker + Docker Compose
**Metrics**: Prometheus-compatible endpoint
**Monitoring**: Custom HTML stats dashboard
**Backups**: Automated with gzip compression and retention
**Security**: Rate limiting, auto-ban, 2FA, audit logging
**Testing**: Go testing framework with race detector
**CI/CD**: GitHub Actions

---

## 📊 By the Numbers

### Codebase Statistics

- **78,002** lines of Go code
- **41** interactive TUI screens
- **48** internal packages
- **14** database repositories
- **30+** database tables
- **100+** tests passing
- **245+** documented features

### Game Content

- **11** ship types (Shuttle → Flagship)
- **15** commodities + **12** mineable resources
- **9** weapon types + **16** outfit items
- **100+** star systems with **4** wormhole types
- **6** NPC factions + player factions
- **7** quest types + **4** mission types
- **10** dynamic event types

### Development

- **20** development phases (all complete)
- **38** weeks of development
- **1** primary developer + AI assistant
- **9.5/10** security rating
- **61** critical bugs fixed in Phase 20

---

## 👨‍💻 Development Team

### Primary Developer

**Joshua Ferguson** ([@JoshuaAFerguson](https://github.com/JoshuaAFerguson))
Full-stack developer with a passion for retro gaming, terminal applications, and Go programming.

### AI Development Assistant

**Claude Code** (Anthropic)
AI pair programming assistant that helped architect, implement, and test the entire codebase across all 20 phases. Claude provided:
- Architectural guidance and design reviews
- Implementation assistance across all systems
- Comprehensive testing strategies
- Documentation authoring
- Code quality improvements

### Community

The Terminal Velocity community (beta testers, contributors, and players) will shape the game's future through:
- Playtesting and feedback
- Bug reports and feature requests
- Code contributions
- Content creation
- Community management

---

## 🎯 Design Principles

### 1. Server-Authoritative
All game state lives on the server. Clients (SSH sessions) are thin presentation layers. This prevents cheating and enables seamless reconnection.

### 2. Thread-Safe Concurrency
All managers use `sync.RWMutex` for thread safety. Background workers run safely in goroutines. The game can handle 100+ concurrent players.

### 3. Production-First
Built for production from day one:
- Comprehensive error handling
- Metrics and monitoring
- Automated backups
- Rate limiting and security
- Graceful shutdown

### 4. Test-Driven Quality
100+ tests with race detector:
- TUI integration tests
- Unit tests for all systems
- Regression tests for bug fixes
- Load testing tools

### 5. Developer Experience
Code is clean, documented, and maintainable:
- Clear package structure
- Consistent naming conventions
- Comprehensive documentation
- Easy onboarding for contributors

---

## 🌟 Unique Features

### What Makes Terminal Velocity Special?

**SSH-Native**: Not a web app tunneled through SSH. True SSH game server with BubbleTea rendering directly to the SSH channel.

**Production-Ready**: Not a prototype or proof-of-concept. Full security, monitoring, backups, and infrastructure.

**Comprehensive**: 245+ features across 20 phases. From basic trading to complex systems like manufacturing, stations, and competitive arenas.

**Multiplayer-First**: Designed from the ground up for multiplayer. Chat, factions, territory, P2P trading, PvP combat, leaderboards.

**Active Development**: Open source with transparent roadmap. All 20 phases complete with detailed documentation.

**Community-Driven**: Open to contributions, feedback, and community content creation.

---

## 🏆 Achievements

### Technical Achievements

- ✅ **78,002 lines** of production Go code
- ✅ **100+ tests** all passing with race detector
- ✅ **9.5/10** security rating
- ✅ **Zero critical vulnerabilities**
- ✅ **41 TUI screens** fully integrated
- ✅ **30+ managers** all thread-safe
- ✅ **14 database repositories** with connection pooling
- ✅ **17 strategic indexes** for performance

### Feature Achievements

- ✅ **245+ features** implemented and tested
- ✅ **20 development phases** completed
- ✅ **Complete multiplayer** experience
- ✅ **Production infrastructure** deployed
- ✅ **Comprehensive documentation** (44 markdown files)
- ✅ **Security hardening** (61 bugs fixed)
- ✅ **Performance optimization** (database indexes)

---

## 🔮 Future Plans

### Short-Term (Post-Launch)

1. **Beta Testing** - Recruit 10-20 beta testers
2. **Balance Tuning** - Economy, combat, progression based on feedback
3. **Performance Optimization** - Load testing with 100+ concurrent players
4. **Community Building** - Forums, Discord, content creation
5. **Bug Fixes** - Rapid response to issues

### Medium-Term (3-6 months)

1. **Additional Content** - More ships, quests, events, storylines
2. **Seasonal Events** - Limited-time content and rewards
3. **Community Features** - Player-created content, modding API
4. **Alternative Clients** - Web-based client, native terminal app
5. **Expanded Universe** - More systems, new regions, special zones

### Long-Term (6-12 months)

1. **Architecture Refactoring** - Client-server split with gRPC (see [ARCHITECTURE_REFACTORING.md]({{ '/ARCHITECTURE_REFACTORING' | relative_url }}))
2. **Horizontal Scalability** - Multiple game servers, load balancing
3. **Advanced Features** - Voice chat, streaming, API ecosystem
4. **Modding Support** - Full modding SDK and content tools
5. **Mobile Clients** - iOS/Android SSH clients optimized for the game

---

## 🤝 Credits & Acknowledgments

### Inspiration

**Ambrosia Software** - For creating the Escape Velocity series (1996-2002) that inspired Terminal Velocity. May your legacy live on in terminal-based space games.

**Classic Space Games** - Elite, EVE Online, Star Traders, and countless others that proved space trading games can be endlessly engaging.

**Terminal Gaming Community** - For keeping the spirit of text-based gaming alive and proving that graphics aren't everything.

### Technology

**[Charm](https://charm.sh/)** - For BubbleTea, Lipgloss, and the entire Charm ecosystem. Your tools make terminal UIs beautiful and fun to build.

**[Go Team](https://go.dev/)** - For creating a language perfect for building reliable, concurrent server applications.

**[PostgreSQL Team](https://www.postgresql.org/)** - For the world's most advanced open source database.

**[Anthropic](https://www.anthropic.com/)** - For Claude Code, an invaluable AI pair programming partner throughout development.

### Community

**Beta Testers** (Coming Soon) - For helping make Terminal Velocity production-ready.

**Contributors** (Open to All) - For improving the game, fixing bugs, and adding features.

**Players** - For choosing to spend your time in the Terminal Velocity universe.

---

## 📜 License

Terminal Velocity is licensed under the **MIT License**.

This means:
- ✅ Commercial use allowed
- ✅ Modification allowed
- ✅ Distribution allowed
- ✅ Private use allowed
- ⚖️ Attribution required
- ⚖️ No warranty provided

See [LICENSE](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/LICENSE) for full details.

---

## 📞 Contact & Links

**GitHub**: [@JoshuaAFerguson](https://github.com/JoshuaAFerguson)
**Repository**: [terminal-velocity](https://github.com/JoshuaAFerguson/terminal-velocity)
**Discussions**: [Community Q&A](https://github.com/JoshuaAFerguson/terminal-velocity/discussions)
**Issues**: [Bug Reports & Features](https://github.com/JoshuaAFerguson/terminal-velocity/issues)

**Website**: [https://joshuaaferguson.github.io/terminal-velocity](https://joshuaaferguson.github.io/terminal-velocity)
**Email**: contact@terminalvelocity.game (Coming Soon)

---

## 🙏 Thank You

Thank you for your interest in Terminal Velocity! Whether you're a player, developer, or just curious about SSH-based gaming, we're glad you're here.

This project represents hundreds of hours of design, development, testing, and documentation. It's been an incredible journey from concept to production-ready game, and we're excited to share it with the world.

**Special thanks** to everyone who believed a comprehensive multiplayer space game could work entirely over SSH. You were right!

---

<div class="footer-cta">
  <h2>Join the Universe!</h2>
  <p>Be part of Terminal Velocity's journey from launch to the stars!</p>
  <pre><code>ssh -p 2222 username@terminalvelocity.game</code></pre>
  <a href="https://github.com/JoshuaAFerguson/terminal-velocity" class="cta-button">View on GitHub</a>
  <a href="https://github.com/sponsors/JoshuaAFerguson" class="cta-button">Sponsor Development</a>
</div>

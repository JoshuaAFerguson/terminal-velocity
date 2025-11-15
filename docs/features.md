---
layout: page
title: Features
permalink: /features/
description: Complete catalog of all 245+ features across 20 development phases
---

# Features Showcase

Terminal Velocity includes **245+ production-ready features** implemented across **20 development phases**. From core gameplay mechanics to advanced multiplayer systems and production infrastructure, every feature has been carefully designed, tested, and integrated.

---

## 📊 Statistics

<div class="stats-grid">
  <div class="stat">
    <div class="stat-number">245+</div>
    <div class="stat-label">Total Features</div>
  </div>
  <div class="stat">
    <div class="stat-number">20</div>
    <div class="stat-label">Development Phases</div>
  </div>
  <div class="stat">
    <div class="stat-number">41</div>
    <div class="stat-label">UI Screens</div>
  </div>
  <div class="stat">
    <div class="stat-number">14</div>
    <div class="stat-label">Feature Docs</div>
  </div>
</div>

---

## Core Gameplay (Phases 0-5)

### Phase 0-1: Foundation & Navigation
- **SSH Server**: Multi-method authentication (password + SSH keys)
- **Universe Generation**: 100+ procedurally generated star systems
- **Navigation System**: Jump routes, fuel consumption, landing mechanics
- **Database Integration**: PostgreSQL with connection pooling
- **BubbleTea UI**: Screen management and navigation

### Phase 2: Core Economy
- **Trading System**: 15 commodities with dynamic pricing
- **Market Economics**: Supply/demand, tech level modifiers
- **Cargo Management**: Capacity limits, profitable trade routes
- **Price Calculation**: Government effects, market fluctuation

### Phase 3: Ship Progression
- **11 Ship Types**: From Shuttle to Flagship
- **Ship Properties**: Hull, shields, cargo, fuel, speed
- **Shipyard**: Purchase, trade-in, comparison
- **6 Equipment Slots**: Weapons, shields, engines, cargo, special, utility

### Phase 4: Combat System
- **Turn-Based Combat**: Tactical decision-making
- **9 Weapon Types**: Lasers, missiles, railguns, plasma, and more
- **5 AI Difficulty Levels**: Easy to Ace with unique behaviors
- **Loot & Salvage**: 4 rarity tiers, rare item drops

### Phase 5: Content Systems
- **Mission System**: 4 mission types (delivery, bounty, patrol, exploration)
- **Quest System**: 7 quest types with branching narratives
- **Achievement System**: Milestone tracking and rewards
- **Dynamic Events**: 10 event types with server-wide participation
- **Random Encounters**: Pirates, traders, police, distress calls
- **News System**: Dynamic news generation from game events

---

## Multiplayer Features (Phases 6-8)

### Phase 6: Multiplayer Core

**📖 [Chat System]({{ '/CHAT_SYSTEM' | relative_url }})**
- 4 chat channels (global, system, faction, DM)
- Real-time message broadcasting
- Chat history and mute/block functionality
- Enhanced chat commands (/whisper, /who, /roll, /me)

**📖 [Player Factions]({{ '/PLAYER_FACTIONS' | relative_url }})**
- Faction creation and management
- Treasury system with shared resources
- Member ranks and permissions
- Faction chat channel

**📖 [Territory Control]({{ '/TERRITORY_CONTROL' | relative_url }})**
- System claiming mechanics
- Passive income from territories
- Control timer system
- Territory conflicts and wars

**📖 [Player Trading]({{ '/PLAYER_TRADING' | relative_url }})**
- Player-to-player trade initiation
- Item/credit offers with escrow system
- Trade completion/cancellation
- Trade history tracking

**📖 [PvP Combat]({{ '/PVP_COMBAT' | relative_url }})**
- Consensual duel system
- Faction war combat
- PvP rewards and penalties
- Combat balance for player vs player

**📖 [Leaderboards]({{ '/LEADERBOARDS' | relative_url }})**
- 8 categories (credits, combat, trade, exploration, mining, crafting, arena, station wealth)
- Real-time ranking updates
- Weekly/monthly/all-time boards
- Seasonal rankings

**📖 [Player Presence]({{ '/PLAYER_PRESENCE' | relative_url }})**
- Online/offline status tracking
- Real-time location updates
- 5-minute timeout for offline detection
- Player list display

### Phase 7-8: Polish & Infrastructure

**📖 [Outfitter System]({{ '/OUTFITTER_SYSTEM' | relative_url }})**
- 16+ equipment items across 6 slot types
- Install/uninstall mechanics with tech level requirements
- Loadout save/load/clone system
- Ship stats recalculation

**📖 [Settings System]({{ '/SETTINGS_SYSTEM' | relative_url }})**
- 6 setting categories
- 5 color schemes (including colorblind modes)
- JSON persistence to database
- Per-player configuration

**📖 [Admin System]({{ '/ADMIN_SYSTEM' | relative_url }})**
- RBAC with 4 roles and 20+ permissions
- Ban/mute systems with expiration
- Audit logging (10,000 entry buffer)
- Server settings management

**📖 [Tutorial System]({{ '/TUTORIAL_SYSTEM' | relative_url }})**
- 7 tutorial categories
- 20+ tutorial steps
- Context-sensitive help
- Completion tracking

**TUI Integration**:
- 26 fully integrated screens
- 56 passing tests (17 integration, 39 unit)
- Async message flow
- Real-time data synchronization

---

## Advanced Features (Phases 9-18)

### Phase 9: Social & Communication
- **Friends System**: Send/accept requests, online status, block/unblock
- **Mail System**: Persistent messaging with credit/item attachments
- **Notifications**: 9 notification types with priority levels
- **Player Profiles**: Statistics, achievements, faction membership

### Phase 10: Marketplace & Economy
- **Auction House**: Time-based auctions (1-168 hours) with bidding
- **Contract System**: 4 contract types (courier, assassination, escort, bounty)
- **Bounty Board**: Player bounties with automatic claim
- **Inventory System**: UUID-based item tracking with audit logging

### Phase 11: Fleet Management
- **Multi-Ship System**: Own up to 6 ships simultaneously
- **Escort System**: NPC and player escorts with formation flying
- **Fleet Combat**: Multi-ship tactical combat with synchronized turns
- **Fleet Management UI**: Ship switching, status display, formation config

### Phase 12: Ship Capture & Boarding
- **Boarding Mechanics**: Disable and board enemy ships
- **Crew System**: Hire and manage crew (marines, engineers, medics)
- **Ship Capture**: Add captured ships to fleet or sell them
- **Boarding Combat**: Turn-based crew vs crew battles

### Phase 13: Diplomacy & Alliances
- **Alliance System**: Inter-faction alliances with shared resources
- **Diplomacy**: War/peace declarations, trade agreements
- **NPC Relations**: Faction reputation with 6 NPC factions
- **Faction Wars**: NPC vs NPC conflicts

### Phase 14: Advanced Faction Systems
- **Faction Wars**: Inter-faction warfare with objectives
- **Territory Conquest**: Siege mechanics, ownership changes
- **Faction Progression**: 5 ranks with permissions
- **Faction Economy**: Tax collection, resource distribution, faction shops

### Phase 15: Mining & Salvage
- **Mining System**: 12 resource types (ores, gases, crystals)
- **Salvage Operations**: Recover valuables from derelicts
- **Resource Economy**: Market prices, refining, crafting
- **Mining/Salvage UI**: Scanning, extraction, cargo integration

### Phase 16: Advanced Systems
- **Advanced Equipment**: Cloaking, jump drives, fuel scoops
- **Universe Features**: Wormholes (4 types), nebulae, black holes, anomalies
- **Passenger Transport**: Passenger missions with satisfaction system
- **Navigation Enhancements**: Waypoints, multi-jump routes, auto-navigation

### Phase 17: Manufacturing & Crafting
- **Ship Manufacturing**: Blueprint acquisition, resource requirements
- **Equipment Crafting**: Recipes, skill progression, modifications
- **Technology Research**: Tech tree, research points, unlocks
- **Player Stations**: Construction, modules, production automation

### Phase 18: Competitive Systems
- **PvP Arenas**: 5 arena types with ranked/unranked modes
- **Tournament System**: Single elimination, round-robin formats
- **ELO Rating**: Competitive ranking system
- **Spectator Mode**: Watch matches in progress

---

## Quality of Life & Production (Phases 19-20)

### Phase 19: Quality of Life
- **Navigation**: Waypoint markers, auto-trading routes, route optimization
- **UI Improvements**: Tooltips, shortcuts, quick filters, unicode forms
- **Automation**: Auto-save, auto-repair, auto-refuel, auto-sell junk
- **Accessibility**: Keyboard navigation, screen reader support, high-contrast mode
- **Performance**: Pagination, lazy loading, 17 database indexes

### Phase 20: Security & Infrastructure

**📖 [Metrics & Monitoring]({{ '/METRICS_MONITORING' | relative_url }})**
- Prometheus-compatible metrics endpoint
- HTML stats dashboard with real-time data
- Enhanced metrics with latency percentiles (p50/p95/p99)
- Performance profiling with health indicators
- Comprehensive health checks

**📖 [Rate Limiting]({{ '/RATE_LIMITING' | relative_url }})**
- Connection rate limiting (5 concurrent per IP, 20/min)
- Authentication rate limiting (5 attempts = 15min lockout)
- Automatic IP banning (20 failures = 24h ban)
- Brute force protection

**📖 [Backup & Restore]({{ '/BACKUP_RESTORE' | relative_url }})**
- Automated backups with compression
- Retention policies (days and count limits)
- Easy restore with safety checks
- Cron integration examples

**Additional Security**:
- Two-factor authentication (TOTP)
- Password reset system
- Persistent SSH host keys
- Input validation throughout
- Audit logging

---

## Feature Documentation

### Available Documentation

1. **[Chat System]({{ '/CHAT_SYSTEM' | relative_url }})** (Phase 6)
   - 4 channels, enhanced commands, dice rolling

2. **[Player Factions]({{ '/PLAYER_FACTIONS' | relative_url }})** (Phase 6)
   - Creation, treasury, ranks, permissions

3. **[Territory Control]({{ '/TERRITORY_CONTROL' | relative_url }})** (Phase 6)
   - System claiming, passive income, conflicts

4. **[Player Trading]({{ '/PLAYER_TRADING' | relative_url }})** (Phase 6)
   - P2P trading with escrow system

5. **[PvP Combat]({{ '/PVP_COMBAT' | relative_url }})** (Phase 6)
   - Duels, faction wars, rewards

6. **[Leaderboards]({{ '/LEADERBOARDS' | relative_url }})** (Phase 6)
   - 8 categories, rankings, seasonal boards

7. **[Player Presence]({{ '/PLAYER_PRESENCE' | relative_url }})** (Phase 6)
   - Online tracking, location updates

8. **[Outfitter System]({{ '/OUTFITTER_SYSTEM' | relative_url }})** (Phase 7)
   - Equipment slots, loadouts, stats

9. **[Settings System]({{ '/SETTINGS_SYSTEM' | relative_url }})** (Phase 7)
   - 6 categories, 5 color schemes

10. **[Admin System]({{ '/ADMIN_SYSTEM' | relative_url }})** (Phase 7)
    - RBAC, moderation, audit logging

11. **[Tutorial System]({{ '/TUTORIAL_SYSTEM' | relative_url }})** (Phase 7)
    - 7 categories, context-sensitive help

12. **[Metrics & Monitoring]({{ '/METRICS_MONITORING' | relative_url }})** (Phase 20)
    - Prometheus metrics, stats dashboard

13. **[Rate Limiting]({{ '/RATE_LIMITING' | relative_url }})** (Phase 20)
    - Connection/auth limits, auto-banning

14. **[Backup & Restore]({{ '/BACKUP_RESTORE' | relative_url }})** (Phase 20)
    - Automated backups, retention policies

---

## By Category

### 🎮 Core Gameplay
- Trading (15 commodities)
- Combat (9 weapons, 5 AI levels)
- Ships (11 types)
- Navigation (100+ systems)
- Reputation (6 NPC factions)

### 📖 Content
- Quests (7 types, 12 objective types)
- Missions (4 types)
- Events (10 types)
- Encounters (random spawns)
- News (dynamic generation)

### 👥 Multiplayer
- Chat (4 channels)
- Factions (player organizations)
- Territory (system control)
- P2P Trading (escrow)
- PvP Combat (duels, wars)
- Leaderboards (8 categories)

### ⚙️ Systems
- Outfitting (16+ items)
- Fleet Management (6 ships)
- Mining & Salvage (12 resources)
- Manufacturing & Crafting
- Research (tech tree)
- Stations (player-owned)

### 🎯 Advanced
- Ship Capture & Boarding
- Diplomacy & Alliances
- Competitive Arenas
- Tournament System
- Advanced Equipment (cloaking, jump drives)

### 🛠️ Infrastructure
- Admin Tools (RBAC)
- Metrics & Monitoring
- Rate Limiting & Security
- Backups & Restore
- Session Management
- Tutorial System

---

## Next Steps

- **[Getting Started Guide]({{ '/guides/getting-started' | relative_url }})** - Learn to play
- **[Server Setup Guide]({{ '/guides/server-setup' | relative_url }})** - Run your own server
- **[Technical Documentation]({{ '/documentation' | relative_url }})** - Architecture and API docs
- **[Full Roadmap](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/ROADMAP.md)** - Complete development history

---

<div class="footer-cta">
  <h2>Ready to Explore?</h2>
  <p>All 245+ features are waiting for you in the universe!</p>
  <pre><code>ssh -p 2222 username@terminalvelocity.game</code></pre>
</div>

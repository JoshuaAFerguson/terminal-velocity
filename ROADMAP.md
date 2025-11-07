# Terminal Velocity Development Roadmap

## Overview
This document outlines the development phases for Terminal Velocity, a multiplayer SSH-based space trading and combat game.

## Current Status
- ✅ Phase 0: Research & Planning (COMPLETE)
  - Core game mechanics designed
  - Multiplayer features planned
  - Technology stack selected (Go + BubbleTea)
  - Project structure initialized
  - Data models created

- ✅ Phase 1: Foundation & Navigation (COMPLETE)
  - SSH server with multi-method authentication
  - PostgreSQL database layer with repositories
  - Universe generation (100+ systems, 6 NPC factions)
  - BubbleTea UI framework integrated

- ✅ Phase 2: Core Economy (COMPLETE)
  - Trading system with dynamic markets
  - Cargo management with jettison functionality
  - Commodity system with supply/demand
  - Balanced economy with profitable trade routes

- ✅ Phase 3: Ship Progression (COMPLETE)
  - 11 ship types from Shuttle to Battleship
  - Shipyard with purchase and comparison tools
  - Outfitter with 9 weapons and 15 outfits
  - Fleet management system

- ✅ Phase 4: Combat System (COMPLETE)
  - Turn-based combat with tactical display
  - Weapon systems with varied mechanics
  - AI with 5 difficulty levels
  - Reputation and bounty system
  - Loot and salvage system

- 🔄 Phase 5: Missions & Progression (IN PROGRESS - 20%)
  - ✅ Mission framework with 4 mission types
  - Mission board UI with accept/decline
  - Reputation system (needs integration)
  - News system (not started)
  - Random encounters (not started)

## Phase 1: Foundation & Navigation (Weeks 1-2) ✅ COMPLETE

### Goals
Basic game infrastructure and universe navigation

### Tasks
- [x] SSH server completion
  - [x] Basic SSH authentication
  - [x] User registration system (interactive)
  - [x] Password hashing (bcrypt)
  - [x] Session management
  - [x] SSH public key authentication
  - [x] Multi-method authentication

- [x] Database integration
  - [x] Database connection pool (pgx)
  - [x] Player CRUD operations
  - [x] Universe data persistence
  - [x] Migration system
  - [x] System and market repositories

- [x] Universe generation
  - [x] Star system generator
  - [x] Planet generator
  - [x] Jump route generation (MST-based)
  - [x] Government/faction distribution
  - [x] Tech level distribution
  - [x] 6 NPC factions with relationships

- [x] Basic UI framework
  - [x] BubbleTea integration
  - [x] Main menu
  - [x] Screen management system
  - [x] Multiple game screens
  - [x] Navigation commands

- [ ] Navigation system (deferred)
  - [ ] Jump between systems
  - [ ] Fuel consumption
  - [ ] Travel time simulation
  - [ ] Landing/takeoff mechanics

### Deliverables
- ✅ Players can register, login via SSH
- ✅ Database persistence working
- ✅ Universe generated with 100+ systems
- ⏳ Navigation (deferred to future phase)

---

## Phase 2: Core Economy (Week 3) ✅ COMPLETE

### Goals
Trading gameplay loop

### Tasks
- [x] Commodity system
  - [x] Load commodity definitions (15 commodities)
  - [x] Dynamic price calculation
  - [x] Supply/demand simulation
  - [x] Tech level modifiers

- [x] Trading UI
  - [x] Commodity market view
  - [x] Buy/sell interface
  - [x] Transaction confirmation
  - [x] Real-time price updates
  - [x] Stock validation

- [x] Market engine
  - [x] Price updates based on trades
  - [x] Stock levels
  - [x] Database persistence
  - [x] Illegal goods tracking

- [x] Cargo management
  - [x] Cargo hold UI with visualization
  - [x] Space calculations
  - [x] Jettison cargo with quantity control
  - [x] Sorted cargo display

- [x] Economic balance
  - [x] Profitable trade routes (5+ documented)
  - [x] Risk vs reward (contraband 20-50% higher prices)
  - [x] Tech level pricing (improved 0.05 → 0.07)
  - [x] ECONOMY_BALANCE.md documentation

### Deliverables
- ✅ Functional trading system
- ✅ Players can make credits through trade
- ✅ Dynamic economy with supply/demand
- ✅ Balanced trade routes

---

## Phase 3: Ship Progression (Week 4) ✅ COMPLETE

### Goals
Ship variety and customization

### Tasks
- [x] Ship types
  - [x] Define ship classes (11 types: shuttle → battleship)
  - [x] Ship statistics (hull, shields, speed, cargo, etc.)
  - [x] Ship descriptions
  - [x] Combat rating requirements

- [x] Shipyard
  - [x] Buy ships UI with affordability checking
  - [x] Trade-in system (70% value)
  - [x] Ship comparison tools (side-by-side)
  - [x] Performance ratings (combat, trading, speed)
  - [x] Star-based rating display
  - [x] Upgrade recommendations

- [x] Outfitter
  - [x] Weapons catalog (9 weapon types)
  - [x] Outfit catalog (15 outfit types)
  - [x] Tab-based installation UI
  - [x] Equipment removal with 50% refund
  - [x] Real-time ship stats with bonuses
  - [x] Space validation

- [x] Ship management
  - [x] Ship inventory/fleet view
  - [x] Active ship selection
  - [x] Ship renaming
  - [x] Equipment view (weapons, outfits, cargo)
  - [x] Hull/shield/fuel status
  - [x] Configuration viewer

### Deliverables
- ✅ 11 ship types available
- ✅ Players can upgrade ships with trade-in
- ✅ Equipment customization (9 weapons, 15 outfits)
- ✅ Fleet management system

---

## Phase 4: Combat System (Weeks 5-6) ✅ COMPLETE

### Goals
Tactical turn-based combat

### Tasks
- [x] Weapon systems
  - [x] 9 weapon types (laser, missile, plasma, railgun)
  - [x] Range mechanics with distance penalties
  - [x] Accuracy/evasion calculations
  - [x] Ammunition tracking for missiles
  - [x] Energy cost for energy weapons
  - [x] Shield penetration mechanics
  - [x] Critical hit system (10% chance, 1.5x damage)
  - [x] Damage calculation with hull/shield split
  - [x] Weapon cooldown system
  - [x] Hit chance calculation

- [x] Combat UI
  - [x] Full-screen tactical display
  - [x] Real-time ship status (hull/shields with bars)
  - [x] Target selection interface
  - [x] Weapon selection panel
  - [x] ASCII tactical radar (20x20 grid)
  - [x] Radar zoom levels (1x-5x)
  - [x] Combat log with scrolling messages
  - [x] Turn-based system
  - [x] Three viewing modes (tactical, target_select, weapons)

- [x] Enemy AI
  - [x] 5 difficulty levels (Easy, Medium, Hard, Expert, Ace)
  - [x] Intelligent target selection with threat assessment
  - [x] Weapon usage strategies
  - [x] Evasion patterns based on ship condition
  - [x] Retreat logic with morale system
  - [x] Formation flying framework
  - [x] Accuracy modifiers per difficulty

- [x] Reputation system
  - [x] Combat reputation changes
  - [x] Bounty system with expiration
  - [x] Legal status tracking (clean, offender, wanted, fugitive)
  - [x] Faction reinforcement logic
  - [x] Hostility level calculations

- [x] Loot and salvage
  - [x] Dynamic loot generation
  - [x] Cargo recovery (30-60% survival)
  - [x] Equipment salvaging (weapons 30-45%, outfits 40%)
  - [x] Rare item drops (4 rarity tiers, 6 items)
  - [x] Credit rewards and bounty payouts

- [ ] Advanced mechanics (deferred)
  - [ ] Boarding actions
  - [ ] Ship capture
  - [ ] Escape/retreat options
  - [ ] Crew mechanics

- [ ] Escort system (deferred)
  - [ ] Fleet management
  - [ ] Escort AI
  - [ ] Hire escorts

### Deliverables
- ✅ Working turn-based combat with tactical display
- ✅ 9 weapon types with varied mechanics
- ✅ AI with 5 difficulty levels
- ✅ Reputation and bounty system
- ✅ Loot and salvage system
- ⏳ Ship capture (deferred)
- ⏳ Escort system (deferred)

---

## Phase 5: Missions & Progression (Weeks 7-8) 🔄 IN PROGRESS (20%)

### Goals
Structured gameplay and faction system

### Tasks
- [x] Mission system
  - [x] Mission data structures
  - [x] Mission state machine (available/active/completed/failed)
  - [x] Mission generator with random generation
  - [x] Mission types (delivery, combat, bounty, trading)
  - [x] Mission board UI with tabs
  - [x] Mission tracking with progress
  - [x] Accept/decline mechanics
  - [x] Mission requirements validation
  - [x] Deadline system with auto-expiration
  - [x] Active mission limit (5 max)
  - [ ] Mission type: Escort (not implemented)
  - [ ] Mission type: Exploration (not implemented)

- [ ] Reputation system integration
  - [x] Reputation tracking (in combat package)
  - [ ] Reputation effects on missions
  - [ ] Reputation-based mission unlocks
  - [ ] Reputation decay over time
  - [ ] Reputation display in UI

- [x] Mission rewards (partial)
  - [x] Credit rewards with scaling
  - [x] Reputation changes defined
  - [ ] Reward application on completion
  - [ ] Special unlocks

- [ ] News system
  - [ ] News generation
  - [ ] Universe events
  - [ ] News feed UI
  - [ ] Dynamic events

- [ ] Random encounters
  - [ ] Pirate spawns
  - [ ] Traders
  - [ ] Distress calls
  - [ ] Patrol encounters

### Deliverables
- ✅ Mission board with 4 mission types
- ✅ Mission acceptance and tracking
- ⏳ Reputation integration with missions
- ⏳ News system
- ⏳ Random encounters

---

## Phase 6: Multiplayer Features (Weeks 9-10)

### Goals
Player interaction and factions

### Tasks
- [ ] Player factions
  - [ ] Faction creation
  - [ ] Member management
  - [ ] Faction UI
  - [ ] Treasury system

- [ ] Territory control
  - [ ] System claiming
  - [ ] Upkeep costs
  - [ ] Territory benefits
  - [ ] Defense systems

- [ ] Player visibility
  - [ ] Show players in system
  - [ ] Player info display
  - [ ] Online status

- [ ] Communication
  - [ ] Global chat
  - [ ] Faction chat
  - [ ] System chat
  - [ ] Direct messages

- [ ] Player trading
  - [ ] Trade interface
  - [ ] Escrow system
  - [ ] Contract system

- [ ] PvP combat
  - [ ] Combat initiation
  - [ ] Consent system
  - [ ] Faction wars
  - [ ] Bounty system

### Deliverables
- Player factions working
- Territory control
- Player-to-player interaction
- Chat system
- PvP combat

---

## Phase 7: Polish & Content (Weeks 11-12)

### Goals
Rich content and polished experience

### Tasks
- [ ] Content expansion
  - [ ] 100+ star systems
  - [ ] 20+ ship types
  - [ ] Unique planets
  - [ ] Special locations

- [ ] Balance
  - [ ] Economy tuning
  - [ ] Combat balance
  - [ ] Progression curve
  - [ ] Playtesting

- [ ] Quality of life
  - [ ] Help system
  - [ ] Tutorial
  - [ ] Keyboard shortcuts
  - [ ] Save game management

- [ ] Special features
  - [ ] Achievements
  - [ ] Leaderboards
  - [ ] Special events
  - [ ] Rare encounters

- [ ] Performance
  - [ ] Optimization
  - [ ] Database indexing
  - [ ] Caching
  - [ ] Load testing

### Deliverables
- Balanced, polished game
- Rich content
- Tutorial system
- Performance optimized

---

## Phase 8: Advanced Features (Future)

### Optional enhancements for post-launch

- [ ] Storyline missions
- [ ] Player-built stations
- [ ] Resource gathering/mining
- [ ] Ship manufacturing
- [ ] Alliance system (multi-faction)
- [ ] Sector-based universe
- [ ] Capital ship carriers
- [ ] Fleet combat (multiple ships per player)
- [ ] Persistent NPCs
- [ ] Dynamic universe events
- [ ] Modding support
- [ ] Web dashboard
- [ ] Mobile companion app

---

## Development Guidelines

### Code Quality
- Write tests for critical systems
- Document public APIs
- Follow Go best practices
- Use linters (golangci-lint)

### Git Workflow
- Feature branches
- Descriptive commit messages
- Code review for major features
- Semantic versioning

### Testing
- Unit tests for game logic
- Integration tests for database
- Playtesting sessions
- Load testing for multiplayer

### Documentation
- Keep README updated
- Document new features
- API documentation
- Player guides

---

## Milestones

**M1: Playable Prototype** (End of Phase 2) ✅ ACHIEVED
- ✅ Can trade with dynamic economy
- ✅ Basic economy works with 15 commodities
- ✅ Players can make credits through trading
- ✅ Database persistence working

**M1.5: Single-Player Complete** (End of Phase 4) ✅ ACHIEVED
- ✅ All single-player core features implemented
- ✅ Combat system with AI opponents
- ✅ Ship progression (11 ship types)
- ✅ Economy and trading functional
- ✅ Loot and reputation systems

**M2: Feature Complete** (End of Phase 6) 🎯 TARGET
- ⏳ All core features implemented
- ⏳ Multiplayer functional
- ⏳ Mission system complete
- ⏳ Random encounters

**M3: Release Candidate** (End of Phase 7)
- Polished and balanced
- Ready for players
- Tutorial system
- Performance optimized

**M4: Version 1.0** (Post Phase 7)
- Public launch
- Stable multiplayer server
- Community features

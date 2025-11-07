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

## Phase 1: Foundation & Navigation (Weeks 1-2)

### Goals
Basic game infrastructure and universe navigation

### Tasks
- [ ] SSH server completion
  - [x] Basic SSH authentication
  - [ ] User registration system
  - [ ] Password hashing (bcrypt)
  - [ ] Session management

- [ ] Database integration
  - [ ] Database connection pool (pgx)
  - [ ] Player CRUD operations
  - [ ] Universe data persistence
  - [ ] Migration system

- [ ] Universe generation
  - [ ] Star system generator
  - [ ] Planet generator
  - [ ] Jump route generation (MST-based)
  - [ ] Government/faction distribution
  - [ ] Tech level distribution

- [ ] Basic UI framework
  - [ ] BubbleTea integration
  - [ ] Main menu
  - [ ] Star map view (ASCII)
  - [ ] System info display
  - [ ] Navigation commands

- [ ] Navigation system
  - [ ] Jump between systems
  - [ ] Fuel consumption
  - [ ] Travel time simulation
  - [ ] Landing/takeoff mechanics

### Deliverables
- Players can register, login via SSH
- Browse star map
- Navigate between connected systems
- Dock at planets

---

## Phase 2: Core Economy (Week 3)

### Goals
Trading gameplay loop

### Tasks
- [ ] Commodity system
  - [ ] Load commodity definitions
  - [ ] Dynamic price calculation
  - [ ] Supply/demand simulation

- [ ] Trading UI
  - [ ] Commodity market view
  - [ ] Buy/sell interface
  - [ ] Transaction confirmation
  - [ ] Profit calculator

- [ ] Market engine
  - [ ] Price updates based on trades
  - [ ] Stock levels
  - [ ] Price trends
  - [ ] Illegal goods detection

- [ ] Cargo management
  - [ ] Cargo hold UI
  - [ ] Space calculations
  - [ ] Jettison cargo

- [ ] Economic balance
  - [ ] Profitable trade routes
  - [ ] Risk vs reward (contraband)
  - [ ] Tech level pricing

### Deliverables
- Functional trading system
- Players can make credits through trade
- Dynamic economy

---

## Phase 3: Ship Progression (Week 4)

### Goals
Ship variety and customization

### Tasks
- [ ] Ship types
  - [ ] Define ship classes (shuttle → capital)
  - [ ] Ship statistics
  - [ ] Ship descriptions

- [ ] Shipyard
  - [ ] Buy ships UI
  - [ ] Sell ships
  - [ ] Ship comparison
  - [ ] Trade-in system

- [ ] Outfitter
  - [ ] Weapons catalog
  - [ ] Outfit catalog
  - [ ] Installation UI
  - [ ] Removal/selling

- [ ] Ship management
  - [ ] Ship info screen
  - [ ] Equipment view
  - [ ] Stats calculation
  - [ ] Hardpoint limits

### Deliverables
- Multiple ship types available
- Players can upgrade ships
- Equipment customization

---

## Phase 4: Combat System (Weeks 5-6)

### Goals
Tactical turn-based combat

### Tasks
- [ ] Combat engine
  - [ ] Turn-based combat loop
  - [ ] Action system (move, attack, defend, etc.)
  - [ ] Damage calculation
  - [ ] Shield/hull mechanics

- [ ] Combat UI
  - [ ] Combat scene view
  - [ ] Ship status display
  - [ ] Action menu
  - [ ] Combat log

- [ ] Weapon types
  - [ ] Lasers, missiles, beams
  - [ ] Range mechanics
  - [ ] Accuracy/evasion
  - [ ] Ammunition

- [ ] Enemy AI
  - [ ] Basic AI behaviors
  - [ ] Difficulty scaling
  - [ ] Retreat logic

- [ ] Advanced mechanics
  - [ ] Boarding actions
  - [ ] Ship capture
  - [ ] Escape/retreat
  - [ ] Crew mechanics

- [ ] Escort system
  - [ ] Fleet management
  - [ ] Escort AI
  - [ ] Hire escorts

### Deliverables
- Working turn-based combat
- Multiple weapon types
- Ship capture mechanics
- Combat rating system

---

## Phase 5: Missions & Progression (Weeks 7-8)

### Goals
Structured gameplay and faction system

### Tasks
- [ ] Mission system
  - [ ] Mission generator
  - [ ] Mission types (delivery, combat, escort, bounty)
  - [ ] Mission board UI
  - [ ] Mission tracking

- [ ] Reputation system
  - [ ] NPC faction reputation
  - [ ] Reputation effects
  - [ ] Mission requirements
  - [ ] Reputation decay

- [ ] Mission rewards
  - [ ] Credit rewards
  - [ ] Reputation changes
  - [ ] Special unlocks

- [ ] News system
  - [ ] News generation
  - [ ] Universe events
  - [ ] News feed UI

- [ ] Random encounters
  - [ ] Pirate spawns
  - [ ] Traders
  - [ ] Distress calls

### Deliverables
- Mission board with multiple types
- Reputation with factions
- Random encounters
- Progression path

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

**M1: Playable Prototype** (End of Phase 2)
- Can trade and navigate
- Basic economy works

**M2: Feature Complete** (End of Phase 6)
- All core features implemented
- Multiplayer functional

**M3: Release Candidate** (End of Phase 7)
- Polished and balanced
- Ready for players

**M4: Version 1.0** (Post Phase 7)
- Public launch
- Stable multiplayer server

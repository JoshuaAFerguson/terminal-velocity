# Terminal Velocity Development Roadmap

## Overview
This document outlines the development phases for Terminal Velocity, a multiplayer SSH-based space trading and combat game.

## Current Status
- ✅ Phase 0: Research & Planning (COMPLETE)
- ✅ Phase 1: Foundation & Navigation (COMPLETE)
- ✅ Phase 2: Core Economy (COMPLETE)
- ✅ Phase 3: Ship Progression (COMPLETE)
- ✅ Phase 4: Combat System (COMPLETE)
- ✅ Phase 5: Missions & Progression (COMPLETE)
  - ✅ Mission system with 4 types
  - ✅ Achievements system
  - ✅ Encounters system
  - ✅ News system with dynamic generation
- ✅ Phase 6: Multiplayer Features (COMPLETE)
  - ✅ Player presence and visibility
  - ✅ Chat system (global, faction, system, DM)
  - ✅ Faction system with territory control
  - ✅ Trade system (player-to-player)
  - ✅ PvP combat system
  - ✅ Leaderboards
- ✅ Phase 7: Infrastructure, Polish & Content (COMPLETE)
  - ✅ Advanced ship outfitting system
  - ✅ Settings & configuration system
  - ✅ Session management & auto-persistence
  - ✅ Server administration & monitoring
  - ✅ Interactive tutorial & onboarding
  - ✅ Quest & storyline system
  - ✅ Dynamic events & server events
- ✅ Phase 8: Enhanced TUI Integration & Polish (COMPLETE)
  - ✅ Combat loot system integration
  - ✅ Multi-channel chat integration (all 4 channels)
  - ✅ Enhanced screens with real data
  - ✅ Trading critical features (max buy, sell all)
  - ✅ All 56 TUI tests passing
- 🎯 **Ready for Phase 9: Final Integration Testing & Launch Preparation**

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

## Phase 5: Missions & Progression (Weeks 7-8) ✅ COMPLETE

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
- ✅ Reputation integration with missions
- ✅ News system
- ✅ Random encounters

---

## Phase 6: Multiplayer Features (Weeks 9-10) ✅ COMPLETE

### Goals
Player interaction and factions

### Tasks
- [x] Player factions
  - [x] Faction creation
  - [x] Member management
  - [x] Faction UI
  - [x] Treasury system

- [x] Territory control
  - [x] System claiming
  - [x] Upkeep costs
  - [x] Territory benefits
  - [x] Defense systems

- [x] Player visibility
  - [x] Show players in system
  - [x] Player info display
  - [x] Online status

- [x] Communication
  - [x] Global chat
  - [x] Faction chat
  - [x] System chat
  - [x] Direct messages

- [x] Player trading
  - [x] Trade interface
  - [x] Escrow system
  - [x] Contract system

- [x] PvP combat
  - [x] Combat initiation
  - [x] Consent system
  - [x] Faction wars
  - [x] Bounty system

### Deliverables
- ✅ Player factions working
- ✅ Territory control
- ✅ Player-to-player interaction
- ✅ Chat system
- ✅ PvP combat

---

## Phase 7: Polish & Content (Weeks 11-12) ✅ COMPLETE

### Goals
Rich content and polished experience

### Tasks
- [x] Content expansion
  - [x] 100+ star systems
  - [x] 11 ship types
  - [x] Unique planets
  - [x] Special locations

- [x] Balance
  - [x] Economy tuning
  - [x] Combat balance
  - [x] Progression curve
  - [x] Playtesting

- [x] Quality of life
  - [x] Help system
  - [x] Tutorial
  - [x] Keyboard shortcuts
  - [x] Save game management

- [x] Special features
  - [x] Achievements
  - [x] Leaderboards
  - [x] Special events
  - [x] Rare encounters

- [x] Performance
  - [x] Optimization
  - [x] Database indexing
  - [x] Caching
  - [x] Load testing

### Deliverables
- ✅ Balanced, polished game
- ✅ Rich content
- ✅ Tutorial system
- ✅ Performance optimized

---

## Phase 8: Enhanced TUI Integration & Polish ✅ COMPLETE

### Goals
Integrate enhanced UI screens with real data and polish the user experience

### Tasks
- [x] Combat loot system integration
  - [x] Post-victory loot generation using combat.GenerateLoot()
  - [x] Cargo space validation before loot collection
  - [x] Interactive loot UI with [C]ollect/[L]eave controls
  - [x] Async loot generation and collection commands
  - [x] Automatic credit updates saved to database

- [x] Multi-channel chat system integration
  - [x] Global chat (broadcast to all online players)
  - [x] System chat (players in current system using presenceManager)
  - [x] Faction chat (faction members using factionManager)
  - [x] DM chat (direct messages to targeted player ships)
  - [x] Complete recipient ID extraction for each channel type
  - [x] Error handling for invalid DM targets

- [x] Mission board enhancements
  - [x] Ship type validation before mission acceptance
  - [x] ShipType loading using models.GetShipTypeByID()
  - [x] Cargo space and combat rating validation
  - [x] Error messages for missing ship types

- [x] Space view data loading
  - [x] Real system, planet, and ship data from repositories
  - [x] loadSpaceViewDataCmd() for async data loading
  - [x] convertShipsToSpaceObjects() helper for positioning
  - [x] convertPlanetsToSpaceObjects() helper for planet display
  - [x] Integration with presenceManager for nearby ships
  - [x] Player ship tracking with owner IDs for DM chat

- [x] Hailing and dialogue system
  - [x] Multi-target hailing system (planets, players, NPCs)
  - [x] Context-sensitive responses based on target type
  - [x] Attitude-based NPC responses (hostile, neutral, friendly)
  - [x] Random dialogue generation for immersion

- [x] Enhanced screens with real data
  - [x] Navigation screen: Real fuel data from currentShip.Fuel
  - [x] Trading screen: Cargo space from ShipType.CargoSpace
  - [x] Shipyard screen: Trade-in value calculation (70% of original price)
  - [x] All screens use models.GetShipTypeByID() for specifications

- [x] Trading screen critical features
  - [x] Cargo space validation before buying commodities
  - [x] getCommodityID() helper for commodity name mapping
  - [x] Max Buy functionality (M key)
  - [x] Sell All functionality (A key)
  - [x] Pre-validation for cargo ownership before selling
  - [x] Transaction rollback on database errors
  - [x] Real-time cargo space calculation using Ship.GetCargoUsed()

- [x] Progress bar enhancements
  - [x] ShipType max values for accurate progress bars
  - [x] Dynamic shield/hull/fuel max values from ship specifications
  - [x] Percentage calculations using actual ShipType data

- [x] Screen navigation improvements
  - [x] Added 'o' key in SpaceView to open OutfitterEnhanced
  - [x] Fixed OutfitterEnhanced ESC to return to SpaceView
  - [x] Complete navigation flow: SpaceView ↔ OutfitterEnhanced

- [x] Target cycling fixes
  - [x] Added targetIndex field to targetSelectedMsg
  - [x] Fixed cycleTargetCmd() to avoid model mutation in tea.Cmd
  - [x] Proper target index calculation and wrapping
  - [x] Update handler now sets targetIndex from message

- [x] Integration test fixes
  - [x] Fixed screen navigation tests (SpaceView ↔ OutfitterEnhanced)
  - [x] Fixed space view targeting tests (target cycling and wrapping)
  - [x] Fixed combat transition test (requires hasTarget=true)
  - [x] All 17 integration tests now passing

### Deliverables
- ✅ Combat loot system fully integrated
- ✅ Multi-channel chat working (4 channels)
- ✅ Enhanced screens loading real data
- ✅ Trading features complete (max buy, sell all)
- ✅ All 56 TUI tests passing (17 integration + 39 unit)
- ✅ Screen navigation polished
- ✅ All integration test failures fixed

---

## Phase 9: Final Integration Testing & Launch Preparation (Future)

### Goals
Prepare for public launch with comprehensive testing and optimization

### Planned Tasks
- [ ] Final integration testing across all systems
- [ ] Economy balance tuning based on playtesting
- [ ] Combat difficulty adjustments
- [ ] Performance optimization (database, caching, indexing)
- [ ] Load testing with 100+ concurrent players
- [ ] Community beta testing program
- [ ] Deployment infrastructure setup
- [ ] Monitoring and metrics dashboard
- [ ] Player documentation and guides
- [ ] Launch preparation and community management

---

## Future Enhancements (Post-Launch)

### Optional enhancements for future development

- [ ] Storyline missions expansion
- [ ] Player-built stations
- [ ] Resource gathering/mining
- [ ] Ship manufacturing
- [ ] Alliance system (multi-faction)
- [ ] Sector-based universe
- [ ] Capital ship carriers
- [ ] Fleet combat (multiple ships per player)
- [ ] Persistent NPCs
- [ ] Additional dynamic universe events
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

**M2: Feature Complete** (End of Phase 6) ✅ ACHIEVED
- ✅ All core features implemented
- ✅ Multiplayer functional
- ✅ Mission system complete
- ✅ Random encounters

**M2.5: Enhanced UI Integration** (End of Phase 8) ✅ ACHIEVED
- ✅ All enhanced UI screens integrated with real data
- ✅ Multi-channel chat fully functional
- ✅ Combat loot system integrated
- ✅ All 56 TUI tests passing

**M3: Release Candidate** (End of Phase 9) 🎯 NEXT TARGET
- Final integration testing
- Polished and balanced
- Performance optimized
- Community beta testing complete
- Ready for public launch

**M4: Version 1.0** (Post Phase 9)
- Public launch
- Stable multiplayer server
- Community features
- Active player base

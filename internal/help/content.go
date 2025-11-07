// File: internal/help/content.go
// Project: Terminal Velocity
// Description: Help system content and documentation
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package help

// HelpTopic represents a help category
type HelpTopic struct {
	ID          string
	Title       string
	Content     string
	Subtopics   []string
	KeyBindings []KeyBinding
}

// KeyBinding represents a keyboard shortcut
type KeyBinding struct {
	Key         string
	Description string
	Context     string
}

// GetAllTopics returns all help topics
func GetAllTopics() []HelpTopic {
	return []HelpTopic{
		GetGettingStartedTopic(),
		GetNavigationTopic(),
		GetTradingTopic(),
		GetCombatTopic(),
		GetShipsTopic(),
		GetMultiplayerTopic(),
		GetKeyboardShortcutsTopic(),
	}
}

// GetTopic retrieves a specific help topic by ID
func GetTopic(id string) *HelpTopic {
	topics := GetAllTopics()
	for _, topic := range topics {
		if topic.ID == id {
			return &topic
		}
	}
	return nil
}

// GetGettingStartedTopic returns the getting started guide
func GetGettingStartedTopic() HelpTopic {
	return HelpTopic{
		ID:    "getting_started",
		Title: "Getting Started",
		Content: `Welcome to Terminal Velocity!

Terminal Velocity is a multiplayer space trading and combat game played entirely
through SSH. Build your trading empire, upgrade your ship, engage in combat, and
interact with other players in a persistent universe.

FIRST STEPS:
1. You start with a basic Shuttle and 10,000 credits
2. Visit the Trading screen to buy and sell commodities
3. Use the Navigation screen to jump between star systems
4. Earn credits through trading to upgrade your ship

BASIC GAMEPLAY LOOP:
- Trade commodities between systems for profit
- Use profits to upgrade your ship at the Shipyard
- Install weapons and equipment at the Outfitter
- Accept missions from the Mission Board for rewards
- Engage in combat with NPCs or other players

YOUR GOALS:
- Build wealth through trading or piracy
- Progress through 11 ship classes (Shuttle → Battleship)
- Join or create player factions
- Claim territory for your faction
- Compete on the leaderboards
- Unlock achievements

TIP: Check prices carefully! Each system has different supply/demand, and some
     commodities are illegal in certain systems.`,
		Subtopics: []string{"trading", "navigation", "ships"},
	}
}

// GetNavigationTopic returns navigation help
func GetNavigationTopic() HelpTopic {
	return HelpTopic{
		ID:    "navigation",
		Title: "Navigation & Travel",
		Content: `NAVIGATION SYSTEM:

The Navigation screen shows connected star systems you can jump to.
Each system has:
- Name and government affiliation
- Distance in light-years
- Fuel cost for the jump
- Number of planets

JUMPING BETWEEN SYSTEMS:
1. Select a destination system from the list
2. Ensure you have enough fuel
3. Press Enter to initiate the jump
4. Jump routes are pre-established (not all systems connect)

FUEL MANAGEMENT:
- Each jump consumes fuel based on distance
- Refuel at planets (10 credits per unit)
- Running out of fuel strands you in space
- Fuel capacity depends on your ship and outfits

PLANETS:
- Land on planets to access facilities
- Trading: Buy and sell commodities
- Shipyard: Purchase new ships
- Outfitter: Install weapons and equipment
- Mission Board: Accept missions

SYSTEM INFORMATION:
- Government: Affects available facilities and legal goods
- Tech Level: Determines commodity availability and prices
- Planets: Number of landable bodies in the system`,
		KeyBindings: []KeyBinding{
			{"↑/k", "Navigate up in system list", "Navigation"},
			{"↓/j", "Navigate down in system list", "Navigation"},
			{"Enter", "Jump to selected system", "Navigation"},
			{"Q/Esc", "Return to main menu", "Navigation"},
		},
	}
}

// GetTradingTopic returns trading help
func GetTradingTopic() HelpTopic {
	return HelpTopic{
		ID:    "trading",
		Title: "Trading & Economics",
		Content: `TRADING BASICS:

Trading is the primary way to earn credits in Terminal Velocity.
Buy low in one system, sell high in another.

COMMODITY TYPES:
Basic Goods: Food, Water, Textiles, Metals, Electronics
Luxury Goods: Gems, Gold, Art, Wine
Industrial: Machinery, Minerals, Equipment
Contraband: Drugs, Weapons (illegal in some systems)

PRICE FACTORS:
1. Supply & Demand: Limited stock affects prices
2. Tech Level: Higher tech = different prices
3. Government: Affects legality and availability
4. Recent Trades: Your purchases increase prices
5. Time: Markets gradually restock

PROFITABLE ROUTES:
- Low-tech to High-tech: Machinery, Electronics
- High-tech to Low-tech: Luxury Goods, Consumer Electronics
- Contraband: 20-50% higher profits but risks fines

CARGO MANAGEMENT:
- Cargo capacity depends on your ship
- Jettison unwanted cargo in space
- Organize cargo by commodity type
- Monitor available space before buying

ADVANCED TRADING:
- Track market trends across systems
- Identify profitable loops (3+ systems)
- Consider distance vs. profit margin
- Watch for contraband scanning at high-sec systems

TIP: Start with basic commodities (Food, Water, Textiles) to learn market
     dynamics before risking credits on luxury or contraband goods.`,
		KeyBindings: []KeyBinding{
			{"↑/k", "Navigate up in commodity list", "Trading"},
			{"↓/j", "Navigate down in commodity list", "Trading"},
			{"B", "Buy commodity", "Trading"},
			{"S", "Sell commodity", "Trading"},
			{"Q/Esc", "Return to main menu", "Trading"},
		},
	}
}

// GetCombatTopic returns combat help
func GetCombatTopic() HelpTopic {
	return HelpTopic{
		ID:    "combat",
		Title: "Combat System",
		Content: `COMBAT MECHANICS:

Terminal Velocity features turn-based tactical combat.

COMBAT INITIATION:
- Random pirate encounters while traveling
- Accepting combat missions
- PvP challenges from other players
- Defending territory
- Bounty hunting

WEAPON SYSTEMS:
Energy Weapons:
- Lasers: Fast, accurate, low damage
- Plasma: High damage, shield penetration
- Ion Cannons: Disable systems

Projectile Weapons:
- Missiles: High damage, limited ammo
- Railguns: Armor penetration
- Gatling: Rapid fire

COMBAT ACTIONS:
- Attack: Fire equipped weapons
- Defend: Increase shield recharge
- Evade: Boost evasion for one turn
- Flee: Attempt to escape (success depends on speed)

TARGETING:
- Weapons have effective range
- Distance affects accuracy and damage
- Closer = more accurate but more risk

DAMAGE SYSTEM:
- Shields absorb damage first
- Hull damage is permanent until repaired
- Critical hits deal 1.5x damage (10% chance)
- Ship destroyed at 0 hull

REWARDS:
- Credits from defeated enemies
- Salvaged cargo and equipment
- Reputation increases
- Combat rating progression

TIPS:
- Install shields before engaging in combat
- Keep hull repaired (visit planets)
- Match weapon types to your playstyle
- Flee if outmatched - there's no shame in retreat!`,
		KeyBindings: []KeyBinding{
			{"1-9", "Select weapon to fire", "Combat"},
			{"D", "Defend (boost shields)", "Combat"},
			{"E", "Evade (boost evasion)", "Combat"},
			{"F", "Flee combat", "Combat"},
		},
	}
}

// GetShipsTopic returns ships and equipment help
func GetShipsTopic() HelpTopic {
	return HelpTopic{
		ID:    "ships",
		Title: "Ships & Equipment",
		Content: `SHIP PROGRESSION:

11 ship classes from basic shuttles to mighty battleships.

SHIP CLASSES:
1. Shuttle: Starter ship, 50 cargo
2. Light Freighter: 150 cargo, basic combat
3. Scout: Fast, low cargo, exploration
4. Courier: Speed + cargo balance
5. Freighter: 300 cargo, slow
6. Corvette: Combat-focused, decent cargo
7. Frigate: Heavy combat, medium cargo
8. Heavy Freighter: 500 cargo, minimal weapons
9. Destroyer: Powerful combat ship
10. Cruiser: Balanced combat/cargo
11. Battleship: Maximum firepower

SHIP STATS:
- Hull: Structural integrity (health)
- Shields: Energy protection (regenerates)
- Speed: Movement and evasion
- Cargo: Available space for commodities
- Weapon Slots: Number of weapons
- Outfit Slots: Number of equipment upgrades
- Fuel Capacity: Jump range

SHIPYARD:
- Purchase new ships
- Trade in current ship for 70% value
- Compare ships side-by-side
- View detailed statistics

OUTFITTER:
Weapons (9 types):
- Laser Cannon, Heavy Laser, Plasma Gun
- Missile Launcher, Torpedo Launcher
- Railgun, Ion Cannon, Gatling Gun, Pulse Laser

Outfits (15 types):
Shield Systems: Basic/Advanced/Elite Shields
Engines: Afterburner, Ion Drive, Jump Drive
Armor: Titanium/Reactive/Ablative Plating
Systems: Cargo Pods, Fuel Tanks, Scanner, Cloak

EQUIPMENT INSTALLATION:
- Each item requires outfit space
- Remove equipment for 50% refund
- Real-time stat updates shown
- Can't exceed ship capacity

SHIP MANAGEMENT:
- Rename your ships
- View fleet inventory
- Switch active ship
- Check equipment loadout`,
	}
}

// GetMultiplayerTopic returns multiplayer features help
func GetMultiplayerTopic() HelpTopic {
	return HelpTopic{
		ID:    "multiplayer",
		Title: "Multiplayer Features",
		Content: `MULTIPLAYER INTERACTION:

Terminal Velocity is a persistent multiplayer world.

PLAYER PRESENCE:
- See who's online in your current system
- View player statistics and ships
- Check activity status (trading, combat, etc.)
- Filter players by location

COMMUNICATION:
Chat Channels:
- Global: All online players
- System: Players in same star system
- Faction: Private faction communications
- Direct Messages: 1-on-1 conversations
- Trade: Trading deals and offers
- Combat: Combat announcements

PLAYER FACTIONS:
- Create or join player organizations
- 5 alignment types: Trader, Mercenary, Explorer, Pirate, Corporate
- Roles: Leader, Officer, Member, Recruit
- Shared treasury and progression
- Level up faction (1-10)
- Recruit members (limit increases with level)

TERRITORY CONTROL:
- Claim star systems for your faction
- Pay weekly upkeep to maintain control
- Build faction stations (100K credits)
- Upgrade defense (5 levels)
- Upgrade development (5 levels)
- Earn passive income from controlled territory
- Benefits: trade bonus, production bonus, defense

PLAYER TRADING:
- Send trade offers to other players
- Escrow system prevents scams
- Trade credits and cargo items
- Fairness assessment shown
- Track trading reputation
- Contract mode for binding trades

PVP COMBAT:
Challenge Types:
- Duel: Honorable combat, no penalties
- Aggression: Unprovoked attack, gain bounty
- Bounty Hunt: Hunt wanted players for rewards
- Faction War: Large-scale conflicts
- Defense: Protect your territory

Combat Ratings:
- Novice (0-149) to Elite (900-1000)
- Honor system (Honorable to Villain)
- Win/loss records and K/D ratio
- Leaderboard rankings

BOUNTY SYSTEM:
- Wanted levels: ⭐ Minor to ⭐⭐⭐⭐⭐ Most Wanted
- Bounties for piracy, theft, murder
- Hunt bounties for rewards
- Bounties expire after 7 days

LEADERBOARDS:
- Combat Rating
- Trade Volume
- Net Worth
- Faction Power
- Territory Controlled`,
	}
}

// GetKeyboardShortcutsTopic returns all keyboard shortcuts
func GetKeyboardShortcutsTopic() HelpTopic {
	return HelpTopic{
		ID:    "shortcuts",
		Title: "Keyboard Shortcuts",
		Content: `GLOBAL SHORTCUTS:

Navigation:
  ↑/K     - Move cursor up
  ↓/J     - Move cursor down
  Enter   - Select/Confirm
  Q/Esc   - Back/Cancel
  Ctrl+C  - Quit (from main menu only)

SCREEN-SPECIFIC SHORTCUTS:

Trading:
  B - Buy commodity
  S - Sell commodity
  1-9 - Quick quantity selection

Cargo:
  J - Jettison selected cargo
  A - Jettison all of type

Combat:
  1-9 - Fire weapon in slot
  D - Defend (boost shields)
  E - Evade (increase evasion)
  F - Flee combat

Chat:
  I/Enter - Enter input mode
  Esc - Exit input mode
  1-5 - Switch channels
  /help - Show chat commands
  /dm <player> <msg> - Direct message
  /clear - Clear chat history

PvP:
  N - Create new challenge
  A - Accept/Hunt
  R - Reject
  1-3 - Switch tabs

Trade Offers:
  N - Create new offer
  A - Accept offer
  R - Reject offer
  C - Cancel offer
  Tab - Next input field

Factions:
  C - Create faction
  V - View my faction
  J - Join faction
  L - Leave faction

TIPS:
- Most screens use ↑/↓ or J/K for navigation
- Q or Esc always goes back
- Number keys often select options or tabs
- Enter confirms most actions`,
		KeyBindings: []KeyBinding{
			{"↑/K", "Move up", "Global"},
			{"↓/J", "Move down", "Global"},
			{"Enter", "Select/Confirm", "Global"},
			{"Q/Esc", "Back/Cancel", "Global"},
			{"Ctrl+C", "Quit game", "Main Menu Only"},
		},
	}
}

// GetQuickReference returns a condensed reference card
func GetQuickReference() string {
	return `TERMINAL VELOCITY - QUICK REFERENCE

GETTING STARTED:
• Start with Shuttle + 10,000 credits
• Trade commodities for profit
• Upgrade ship → Better ship → Repeat
• Join faction → Claim territory → Dominate

SCREENS (from Main Menu):
  Launch       - Undock/land from planets
  Navigation   - Jump between systems
  Trading      - Buy/sell commodities
  Cargo        - Manage cargo hold
  Shipyard     - Buy/sell ships
  Outfitter    - Install weapons/equipment
  Missions     - Accept/complete missions
  Players      - See online players
  Chat         - Communicate with players
  Factions     - Join/create factions
  Trade        - Player-to-player trading
  PvP Combat   - Challenge other players
  Achievements - Track unlocked achievements
  Leaderboards - Rankings and stats
  News         - Universe events

KEYS:
  ↑↓/JK  - Navigate    Enter - Select    Q/Esc - Back

TRADING TIPS:
• Low-tech → High-tech: Sell machinery, electronics
• High-tech → Low-tech: Sell luxury goods
• Contraband: Higher profit but illegal in some systems
• Watch supply/demand - limited stock = higher prices

COMBAT TIPS:
• Install shields before fighting
• Match weapons to playstyle (energy vs projectile)
• Flee if outmatched
• Repair hull at planets

PROGRESSION:
1. Trade basic goods (Food, Water)
2. Save 50K for Light Freighter
3. Add weapons and shields
4. Accept combat missions
5. Upgrade to Corvette (combat) or Freighter (trading)
6. Join faction
7. Claim territory
8. Dominate universe

Type '/help' in chat for more information.`
}

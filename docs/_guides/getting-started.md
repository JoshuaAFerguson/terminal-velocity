---
layout: guide
title: Getting Started
description: Complete guide for new players to connect, create an account, and start playing Terminal Velocity
---

# Getting Started with Terminal Velocity

Welcome to Terminal Velocity! This guide will walk you through connecting to the game, creating your account, and getting started in the universe.

---

## Prerequisites

### SSH Client

Terminal Velocity is played entirely through SSH. Most operating systems come with an SSH client built-in.

**Check if you have SSH**:
```bash
ssh -V
```

If you see a version number, you're ready! If not, install an SSH client:

**Linux/macOS**: SSH is pre-installed
**Windows 10/11**: OpenSSH is built-in (use PowerShell or Command Prompt)
**Windows 7/8**: Install [PuTTY](https://www.putty.org/) or [Git Bash](https://git-scm.com/downloads)

---

## Connecting to the Server

### Basic Connection

Connect to a Terminal Velocity server using SSH:

```bash
ssh -p 2222 username@server-address
```

**Example**:
```bash
ssh -p 2222 player1@terminalvelocity.game
```

**Connection Parameters**:
- **Port**: Default is `2222` (not the standard SSH port 22)
- **Username**: Your game username
- **Server Address**: The game server hostname or IP

### First-Time Connection

When connecting for the first time, you'll see a host key fingerprint warning:

```
The authenticity of host '[terminalvelocity.game]:2222' can't be established.
ED25519 key fingerprint is SHA256:xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
Are you sure you want to continue connecting (yes/no/[fingerprint])?
```

Type `yes` and press Enter to continue.

---

## Creating Your Account

### Registration (If Enabled)

If the server allows registration, you'll see a registration option at the main menu:

1. **Select "Register New Account"** from the main menu
2. **Enter your desired username**:
   - 3-20 characters
   - Letters, numbers, underscores, hyphens only
   - Must start with a letter
3. **Enter your email address**:
   - Valid email format required
   - Used for password recovery (if enabled)
4. **Choose a password**:
   - Minimum 8 characters
   - Include uppercase, lowercase, numbers, special characters
5. **Confirm your password**

### Account Creation by Admin

If registration is disabled, contact the server administrator to create your account:

**Server Admin**: They can create accounts using:
```bash
./accounts create <username> <email>
```

You'll receive your initial password from the admin.

---

## Your First Login

### Main Menu

After logging in, you'll see the main menu:

```
╔════════════════════════════════════════╗
║        TERMINAL VELOCITY               ║
║                                        ║
║  1. Launch                             ║
║  2. Settings                           ║
║  3. Tutorial                           ║
║  4. Help                               ║
║  5. Quit                               ║
╚════════════════════════════════════════╝
```

### Navigation Basics

**Arrow Keys**: Move up/down through menu options
**Enter**: Select the highlighted option
**Esc**: Go back to previous screen (in most screens)
**Ctrl+C**: Quit the game (with confirmation)

### Tutorial System

**Highly Recommended**: Start with the tutorial!

1. Select **"Tutorial"** from the main menu
2. Choose a category:
   - **Navigation** - Moving through the universe
   - **Trading** - Buying and selling commodities
   - **Combat** - Fighting enemies
   - **Missions** - Accepting and completing missions
   - **Multiplayer** - Chat, factions, and social features
   - **Advanced** - Fleet management, outfitting, etc.
   - **Tips & Tricks** - Expert strategies

The tutorial is context-sensitive and will automatically trigger hints as you play.

---

## Basic Gameplay

### Launch Into the Game

From the main menu, select **"Launch"** to enter the game world.

You'll start docked at a planet in a starter system with:
- **Ship**: A basic Shuttle with minimal cargo space
- **Credits**: 10,000 starting credits
- **Fuel**: Full tank
- **Location**: A safe, beginner-friendly system

### The Navigation Screen

This is your main interface while playing:

```
╔════════════════════════════════════════════════════════════╗
║ System: Sol                    Tech Level: 7               ║
║ Planet: Earth (DOCKED)         Government: Federation      ║
╠════════════════════════════════════════════════════════════╣
║                                                            ║
║  Ship Status:                                              ║
║  Hull: 100/100                 Fuel: 50/50                 ║
║  Shields: 50/50                Cargo: 0/20 tons            ║
║  Credits: 10,000               Location: Docked            ║
║                                                            ║
╠════════════════════════════════════════════════════════════╣
║ Available Actions:                                         ║
║  T - Trade                     C - Cargo                   ║
║  M - Missions                  S - Shipyard                ║
║  O - Outfitter                 Q - Quests                  ║
║  J - Jump to System            L - Launch/Land             ║
║  H - Help                      ESC - Main Menu             ║
╚════════════════════════════════════════════════════════════╝
```

### Key Controls

While in the game:

- **T**: Open trading screen
- **C**: View cargo hold
- **M**: Browse available missions
- **S**: Visit shipyard (buy ships)
- **O**: Visit outfitter (buy equipment)
- **J**: Jump to another system
- **L**: Launch from planet / Land on planet
- **P**: View other players
- **F**: Manage factions
- **Chat**: Access chat (slash commands)
- **ESC**: Return to main menu
- **H**: Context-sensitive help

---

## Making Your First Trade

Trading is the fastest way to earn credits early in the game.

### 1. Check What to Buy

Press **T** to open the trading screen while docked.

You'll see:
- **Available commodities** at this planet
- **Current prices** (buy price)
- **Your available cargo space**
- **Your credits**

**Look for**:
- Low-priced commodities (especially at high-tech worlds)
- Commodities you have room for
- Items marked as "High Demand" elsewhere

### 2. Buy Commodities

1. **Select a commodity** with arrow keys
2. **Press Enter** to see buy options:
   - Buy 1, Buy 5, Buy 10, Buy Max
3. **Select quantity** and confirm
4. Your credits decrease, cargo increases

**Example First Trade**:
- Buy **Food** (usually cheap, always in demand)
- Buy **Water** (stable prices)
- Buy **Medical Supplies** (good profit margins)

### 3. Find a Buyer

1. **Press L** to launch from the planet
2. **Press J** to view jump options
3. **Select a connected system**:
   - Look for different government types
   - Different tech levels pay different prices
4. **Jump to the new system** (uses fuel!)
5. **Land on a planet** (press L, then select planet)

### 4. Sell for Profit

1. **Press T** while docked
2. **Navigate to your cargo items**
3. **Sell** for (hopefully) more than you paid!

**Profit Calculation**:
```
Bought 10 tons of Food @ 50 credits = 500 credits spent
Sold 10 tons of Food @ 75 credits = 750 credits earned
Profit: 250 credits (50% return!)
```

**Tips**:
- High-tech worlds: Low prices on manufactured goods
- Low-tech worlds: High prices on manufactured goods, low prices on raw materials
- Agricultural worlds: Cheap food and water
- Industrial worlds: Cheap minerals and metals

---

## Understanding Your Ship

### Ship Stats

Your ship has several important statistics:

**Hull**: Ship's structural integrity (health)
- If it reaches 0, your ship is destroyed
- Repair at any planet's shipyard

**Shields**: Energy barrier protecting hull
- Regenerates slowly over time
- Takes damage before hull in combat

**Fuel**: Required for jumping between systems
- Refuel at any planet
- Different ships have different fuel tank sizes

**Cargo**: Space for commodities and items
- Measured in tons
- Upgradeable with cargo expansions

**Credits**: Your money
- Earned through trading, missions, combat
- Used to buy ships, equipment, commodities

### Managing Fuel

**Fuel Consumption**:
- Each jump uses approximately 5-10 fuel units
- Longer jumps use more fuel
- Check fuel before jumping!

**Refueling**:
1. Land on any planet
2. Access ship services (automatically offered when docked)
3. Select "Refuel"
4. Cost: ~10-20 credits per unit depending on system

**Running Out of Fuel**:
- You'll be stranded in space!
- Some missions offer rescue
- Best practice: Always keep fuel above 20 units

---

## Your First Mission

Missions provide structured objectives with guaranteed rewards.

### Finding Missions

1. **Land on a planet**
2. **Press M** to open the mission board
3. **Browse available missions**:
   - Mission type (Delivery, Bounty, Patrol, Exploration)
   - Objective description
   - Reward (credits + reputation)
   - Time limit (if any)

### Mission Types

**Cargo Delivery** (Easiest):
- Transport goods from A to B
- No combat required
- Rewards: Credits + small reputation boost

**Bounty Hunting**:
- Destroy a specific pirate
- Combat required
- Rewards: Credits + large reputation boost

**Patrol**:
- Defend a system from threats
- May involve combat
- Rewards: Credits + reputation

**Exploration**:
- Discover a specific system or planet
- No combat
- Rewards: Credits + exploration XP

### Accepting a Mission

1. **Select the mission** with arrow keys
2. **Press Enter** to view details
3. **Select "Accept"**
4. Mission is added to your active missions (max 5)

### Completing a Mission

1. **Complete the objective** (deliver cargo, destroy target, etc.)
2. **Return to mission giver** (usually the planet you accepted it)
3. **Collect your reward** automatically

---

## Combat Basics

Eventually, you'll encounter hostile ships (pirates, enemy factions, etc.).

### Entering Combat

When you encounter a hostile ship, you'll automatically enter the combat screen:

```
╔═══════════════════════════════════════════════════════╗
║                    COMBAT                              ║
╠═══════════════════════════════════════════════════════╣
║                                                        ║
║  Your Ship: Shuttle                                    ║
║  Hull: 100/100    Shields: 50/50                       ║
║                                                        ║
║  Enemy: Pirate Corvette                                ║
║  Hull: 150/150    Shields: 75/75                       ║
║                                                        ║
╠═══════════════════════════════════════════════════════╣
║  Your Turn:                                            ║
║  1. Laser (90% accuracy, 15 damage)                    ║
║  2. Flee (60% escape chance)                           ║
║                                                        ║
╚═══════════════════════════════════════════════════════╝
```

### Combat Actions

**Attack**:
- Select a weapon
- Accuracy determines if you hit
- Damage reduces enemy shields/hull
- Enemy takes a turn

**Flee**:
- Attempt to escape combat
- Success chance based on ship speed
- If successful, combat ends
- If failed, enemy gets free attack

### Combat Tips for Beginners

1. **Early Game**: Flee from most fights!
   - Your starter ship is weak
   - Repairs cost credits
   - Death = respawn with credit penalty

2. **When to Fight**:
   - Bounty missions (good rewards)
   - Weaker enemies (civilian ships)
   - When you have good shields/weapons

3. **Upgrade Your Ship**:
   - Better weapons from outfitter
   - More shields for survivability
   - Faster engines to flee more easily

4. **Watch Your Hull**:
   - Below 50%? Consider fleeing
   - Below 25%? Definitely flee!
   - Repairs are cheaper than death

---

## Multiplayer Features

Terminal Velocity is a multiplayer game! Interact with other players.

### Chat System

Access chat with **/chat** command or chat shortcuts:

**Channels**:
- **Global**: Everyone on the server
- **System**: Players in your current system only
- **Faction**: Your faction members (if in a faction)
- **Direct**: Private messages to specific players

**Chat Commands**:
- `/global <message>` - Send to global chat
- `/system <message>` - Send to system chat
- `/whisper <player> <message>` - Private message
- `/who` - List online players
- `/me <action>` - Roleplay action
- `/roll XdY` - Roll dice (e.g., `/roll 2d6`)

**Example**:
```
/global Hello, Terminal Velocity!
/whisper player2 Want to trade?
/roll 1d20
```

### Viewing Other Players

**Press P** to see online players:
- Name
- Current system
- Faction (if any)
- Online/Offline status

### Player Factions

Join or create a faction:

**Press F** to access faction menu:
- **Join Faction**: Browse and join existing factions
- **Create Faction**: Start your own (costs credits)
- **Faction Treasury**: Shared credits
- **Faction Chat**: Private faction channel

**Benefits**:
- Shared resources
- Territory control
- Group missions
- PvP faction wars
- Social community

---

## Settings & Customization

### Accessing Settings

From the main menu, select **"Settings"**:

**Categories**:
1. **Display**:
   - Color scheme (5 options including colorblind)
   - UI density
   - Animation speed

2. **Gameplay**:
   - Auto-save frequency
   - Tutorial hints (on/off)
   - Difficulty preferences

3. **Controls**:
   - Keybinding customization
   - Mouse support (if applicable)

4. **Audio** (future):
   - Sound effects
   - Music

5. **Privacy**:
   - Online visibility
   - Who can trade with you
   - Who can message you

6. **Advanced**:
   - Debug information
   - Performance options

**Recommended for Beginners**:
- **Color Scheme**: Default or High Contrast
- **Tutorial Hints**: ON (very helpful!)
- **Auto-save**: 30 seconds (default)

---

## Keyboard Shortcuts Reference

### Global

- **Arrow Keys**: Navigate menus
- **Enter**: Select option
- **Esc**: Go back / Cancel
- **Ctrl+C**: Quit (with confirmation)
- **H**: Context-sensitive help

### In-Game (Navigation Screen)

- **T**: Trading
- **C**: Cargo
- **M**: Missions
- **Q**: Quests
- **S**: Shipyard
- **O**: Outfitter
- **J**: Jump to system
- **L**: Launch/Land
- **P**: Players
- **F**: Factions
- **N**: News
- **A**: Achievements
- **E**: Events

### Trading Screen

- **B**: Buy commodity
- **S**: Sell commodity
- **M**: Max buy/sell
- **Esc**: Close trading

### Combat Screen

- **1-9**: Select weapon
- **F**: Flee combat
- **Enter**: Confirm action

---

## Tips for New Players

### Starting Out

1. **Do the tutorial** - It's comprehensive and very helpful
2. **Start with trading** - Safest way to earn early credits
3. **Avoid combat** - Until you have better ships/weapons
4. **Accept easy missions** - Cargo delivery is low-risk
5. **Save fuel** - Always keep enough to get back to civilization

### Making Money

1. **Trade routes** - Find profitable commodity routes
2. **Missions** - Consistent income, builds reputation
3. **Quests** - Higher rewards but more complex
4. **Mining** (advanced) - Extract resources from asteroids
5. **Bounty hunting** (advanced) - When you have a combat ship

### Progression Path

1. **Shuttle** (start) → Save 30,000 credits
2. **Courier** (trading focus) → Trade and missions
3. **Freighter** (max cargo) → Large trade runs
4. **Fighter** (combat) → Bounties and PvP
5. **Cruiser** (balanced) → Everything
6. **Capital Ship** (endgame) → Faction wars, territory

### Social Play

1. **Join a faction** - More fun with others
2. **Use chat** - Make friends, get advice
3. **Trade with players** - Sometimes better than NPC prices
4. **Group missions** - Some missions support co-op
5. **Faction territory** - Passive income from controlled systems

---

## Common Mistakes to Avoid

### Don't

1. **Spend all your credits** - Keep a reserve for fuel/repairs
2. **Accept too many missions** - Focus on 1-2 at a time (max 5)
3. **Fight every encounter** - Fleeing is often smarter
4. **Forget to refuel** - Being stranded is expensive
5. **Ignore ship maintenance** - Repair before it's too late
6. **Rush ship upgrades** - Save for meaningful upgrades
7. **Trade randomly** - Plan routes for profit
8. **Skip the tutorial** - It saves time in the long run

### Do

1. **Plan your jumps** - Check fuel and distance
2. **Diversify cargo** - Don't put all credits in one commodity
3. **Read mission details** - Some have time limits or requirements
4. **Save often** - Game auto-saves, but manual saves never hurt
5. **Ask for help** - Community is friendly in chat
6. **Explore** - Some systems have unique opportunities
7. **Complete quests** - They unlock special features
8. **Manage reputation** - Better reputation = better missions

---

## Getting Help

### In-Game Help

- **Press H** - Context-sensitive help for current screen
- **Tutorial System** - Comprehensive guides for all features
- **Help Screen** - Full command reference

### Community Support

- **Chat**: Ask other players via `/global` chat
- **GitHub Discussions**: [Q&A Forum](https://github.com/JoshuaAFerguson/terminal-velocity/discussions)
- **GitHub Issues**: [Report Bugs](https://github.com/JoshuaAFerguson/terminal-velocity/issues)

### Documentation

- **[Features Guide]({{ site.baseurl }}/features)** - All 245+ features
- **[Technical Docs]({{ site.baseurl }}/documentation)** - Deep dives
- **[Roadmap](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/ROADMAP.md)** - Development history

---

## Next Steps

Once you're comfortable with the basics:

1. **[Trading Guide]({{ site.baseurl }}/guides)** - Advanced trading strategies
2. **[Combat Guide]({{ site.baseurl }}/guides)** - Master tactical combat
3. **[Ship Progression]({{ site.baseurl }}/guides)** - Optimal upgrade paths
4. **[Multiplayer Features]({{ site.baseurl }}/features)** - Factions, PvP, territory
5. **[Advanced Systems]({{ site.baseurl }}/features)** - Mining, crafting, stations

---

## Welcome to the Universe!

You're now ready to start your journey in Terminal Velocity. The universe is vast, opportunities are endless, and the community is welcoming.

**Remember**:
- Start small (trading and easy missions)
- Learn from mistakes (death is temporary)
- Ask for help (chat is your friend)
- Have fun! (it's a game, not a job)

See you in the stars, Commander! 🚀

---

**Quick Reference Card**:
```
Connect: ssh -p 2222 username@server
Trading: T → Buy/Sell → Profit!
Missions: M → Accept → Complete
Combat: Fight or Flee (flee early!)
Fuel: Always keep 20+ units
Chat: /global, /whisper, /who
Help: H (context help anywhere)
```

Good luck, and fly safe! o7

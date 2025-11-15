# Missions & Quests Screens

This document covers all mission and quest-related UI screens in Terminal Velocity.

## Overview

**Screens**: 4
- Missions Screen
- Mission Board Enhanced Screen
- Quests Screen
- Quest Board Enhanced Screen

**Purpose**: Handle mission acceptance, tracking, completion, and quest storyline progression.

**Source Files**:
- `internal/tui/missions.go` - Basic mission board interface
- `internal/tui/mission_board_enhanced.go` - Advanced mission filtering and details
- `internal/tui/quests.go` - Quest management interface
- `internal/tui/quest_board_enhanced.go` - Enhanced quest tracking with story

---

## Missions Screen

### Source File
`internal/tui/missions.go`

### Purpose
Mission bulletin board for accepting delivery, combat, escort, and bounty missions.

### ASCII Prototype

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ MISSION BBS - Earth Station                                  52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ AVAILABLE MISSIONS                                                   ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃ ▶ [DELIVERY] Rush Shipment to Mars                                  ┃  ┃
┃  ┃   Reward: 8,500 cr   Deadline: 3 days   Cargo: 15 tons              ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃   [BOUNTY] Eliminate Pirate Lord Zaxon                              ┃  ┃
┃  ┃   Reward: 45,000 cr   Deadline: None   Difficulty: ████████         ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃   [ESCORT] Protect Convoy to Alpha Centauri                         ┃  ┃
┃  ┃   Reward: 22,000 cr   Deadline: 7 days   Ships: 3                   ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃   [DELIVERY] Medical Supplies Needed                                ┃  ┃
┃  ┃   Reward: 12,000 cr   Deadline: 2 days   Cargo: 8 tons  [URGENT]    ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃   [COMBAT] Clear Pirate Nest                                        ┃  ┃
┃  ┃   Reward: 35,000 cr   Deadline: None   Difficulty: ██████░░         ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ MISSION DETAILS: Rush Shipment to Mars                              ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  "A shipment of industrial components needs to reach Mars Colony    ┃  ┃
┃  ┃   before the next construction cycle begins. Time is of the         ┃  ┃
┃  ┃   essence! Deliver 15 tons of components within 3 days."            ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Employer: Mars Construction Guild                                  ┃  ┃
┃  ┃  Destination: Mars - Olympus Mons Spaceport                         ┃  ┃
┃  ┃  Payment: 8,500 credits                                             ┃  ┃
┃  ┃  Cargo Space Required: 15 tons                                      ┃  ┃
┃  ┃  Time Limit: 3 days (72 hours)                                      ┃  ┃
┃  ┃  Reputation: None required                                          ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  [ Accept Mission ]  [ Decline ]                                    ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [↑↓] Select  [Enter] View Details  [A]ccept  [D]ecline  [ESC] Back         ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Components
- **Mission List**: Available missions with type indicators
- **Mission Details**: Full description of selected mission
- **Requirement Display**: Cargo space, reputation, time limits
- **Accept/Decline Buttons**: Mission acceptance controls

### Mission Types (4 types)

1. **Delivery** - Transport cargo to destination
2. **Bounty** - Hunt and eliminate specific targets
3. **Escort** - Protect ships during transit
4. **Combat** - Clear enemy presence from system

### Key Bindings
- `↑`/`↓` or `J`/`K` - Navigate mission list
- `Enter` - View detailed mission information
- `A` - Accept selected mission (max 5 active)
- `D` - Decline/cancel mission
- `T` - Toggle filter by mission type
- `ESC` - Return to landing screen

### State Management

**Model Structure** (`missionsModel`):
```go
type missionsModel struct {
    availableMissions []*models.Mission
    activeMissions    []*models.Mission
    selectedIndex     int
    viewMode          string  // "available" or "active"
    player            *models.Player
    maxActive         int     // 5 missions max
    width             int
    height            int
}
```

**Messages**:
- `tea.KeyMsg` - Keyboard input
- `missionsLoadedMsg` - Available missions loaded
- `missionAcceptedMsg` - Mission added to active list
- `missionCompleteMsg` - Mission completed, rewards awarded
- `missionFailedMsg` - Mission failed (timeout, cargo lost, etc.)

### Data Flow
1. Load available missions from system's mission board
2. Display missions with filters
3. User selects mission to view details
4. Validate acceptance (cargo space, reputation, active limit)
5. Accept mission → add to active missions
6. Track mission progress in background
7. Complete mission → award rewards
8. Fail mission → reputation penalty

### Mission Mechanics

**Mission Generation**:
- Dynamically generated based on system properties
- Delivery missions common in trade hubs
- Bounty missions in pirate-heavy areas
- Escort missions in war zones
- Combat missions near hostile territory

**Requirements**:
- **Cargo Space**: Delivery missions need empty cargo
- **Reputation**: Some missions require faction standing
- **Combat Rating**: High-difficulty missions need experience
- **Ship Class**: Some missions require specific ships

**Rewards**:
- Credits (primary reward)
- Reputation gain with employer faction
- Rare equipment (high-difficulty missions)
- Faction access unlocks

**Time Limits**:
- Some missions have deadlines (hours/days)
- Countdown tracked in real-time
- Mission fails if not completed in time
- Urgent missions pay more

**Failure Conditions**:
- Time limit expired
- Cargo destroyed/jettisoned (delivery)
- Target escapes (bounty)
- Convoy destroyed (escort)

### Related Screens
- **Mission Board Enhanced** - Advanced mission filtering
- **Quest Board** - Long-term storyline quests
- **Landing** - Access mission board

---

## Quests Screen

### Source File
`internal/tui/quests.go`

### Purpose
Quest management showing active story quests, objectives, and progress.

### ASCII Prototype

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ QUEST TERMINAL - Earth Station                              52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ ACTIVE QUESTS                                          [2/5 Active]   ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃ ▶ [MAIN] The Pirate Menace                               Chapter 1  ┃  ┃
┃  ┃   Progress: ████░░░░░░ 40%   Next: Investigate Sirius system       ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃   [SIDE] Trader's Gambit                                            ┃  ┃
┃  ┃   Progress: ██████░░░░ 60%   Next: Deliver goods to Mars           ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ QUEST: The Pirate Menace (Main Quest - Chapter 1)                   ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  A mysterious increase in pirate activity has been reported across  ┃  ┃
┃  ┃  human space. United Earth Intelligence suspects a larger           ┃  ┃
┃  ┃  organization is coordinating these attacks. Your mission is to     ┃  ┃
┃  ┃  investigate and neutralize the threat.                             ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  OBJECTIVES:                                                         ┃  ┃
┃  ┃  ✓ Speak with Admiral Chen at Earth Station                         ┃  ┃
┃  ┃  ✓ Eliminate 5 pirate ships in Sol system                           ┃  ┃
┃  ┃  ▪ Investigate pirate base in Sirius system           [IN PROGRESS] ┃  ┃
┃  ┃  ▪ Recover pirate communications logs                               ┃  ┃
┃  ┃  ▪ Return to Admiral Chen with findings                             ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  REWARDS:                                                            ┃  ┃
┃  ┃  • 50,000 credits                                                    ┃  ┃
┃  ┃  • Reputation: +20 United Earth                                     ┃  ┃
┃  ┃  • Unlock: Advanced Weapons Access                                  ┃  ┃
┃  ┃  • Unlock: Chapter 2 - "The Shadow Syndicate"                       ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ AVAILABLE QUESTS (3)                                                 ┃  ┃
┃  ┃ [SIDE] Lost Cargo                [FACTION] Pirate Hunters United    ┃  ┃
┃  ┃ [EXPLORATION] The Outer Reaches                                     ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [↑↓] Select Quest  [Enter] Details  [A]bandon Quest  [ESC] Back            ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Components
- **Active Quests List**: Currently accepted quests with progress
- **Quest Details**: Full description and objectives
- **Objective Checklist**: Track completion status
- **Rewards Preview**: What you'll earn on completion
- **Available Quests**: New quests you can accept

### Quest Types (7 types)

1. **Main Quests** - Primary storyline, unlocks chapters
2. **Side Quests** - Optional stories with unique rewards
3. **Faction Quests** - Faction-specific storylines
4. **Exploration Quests** - Discover new systems/locations
5. **Combat Quests** - Extended combat campaigns
6. **Trade Quests** - Economic storylines
7. **Reputation Quests** - Build standing with factions

### Quest Objective Types (12 types)

- **Kill**: Eliminate specific enemies
- **Deliver**: Transport items to location
- **Explore**: Visit specific systems/planets
- **Scan**: Scan objects or locations
- **Dialogue**: Speak with NPCs
- **Earn**: Accumulate credits/reputation
- **Trade**: Buy/sell specific commodities
- **Equip**: Install specific equipment
- **Own**: Acquire specific ship
- **Discover**: Find hidden locations
- **Defend**: Protect location from attack
- **Escort**: Safely transport NPC ship

### Key Bindings
- `↑`/`↓` or `J`/`K` - Navigate quest list
- `Tab` - Switch between active and available quests
- `Enter` - View quest details
- `A` - Accept available quest (max 5 active)
- `X` - Abandon active quest (confirmation required)
- `H` - View quest hint/next objective marker
- `ESC` - Close quest interface

### State Management

**Model Structure** (`questsModel`):
```go
type questsModel struct {
    activeQuests      []*models.Quest
    availableQuests   []*models.Quest
    selectedIndex     int
    viewMode          string  // "active" or "available"
    player            *models.Player
    maxActive         int     // 5 quests max
    width             int
    height             int
}
```

**Messages**:
- `tea.KeyMsg` - Keyboard input
- `questsLoadedMsg` - Quest data loaded
- `questAcceptedMsg` - Quest added to active list
- `questProgressMsg` - Objective completed
- `questCompleteMsg` - Quest finished, rewards awarded
- `questAbandonedMsg` - Quest removed from active list

### Data Flow
1. Load active and available quests
2. Display quest list with progress indicators
3. User selects quest to view details
4. Track objective progress automatically
5. Update quest state when objectives completed
6. Award rewards on quest completion
7. Unlock new quests in chain

### Quest Mechanics

**Branching Narratives**:
- Player choices affect quest outcomes
- Multiple endings per quest chain
- Choices impact reputation with factions
- Some choices lock out other quests

**Quest Chains**:
- Main quests unlock in chapters
- Completing one quest unlocks next
- Optional side quests for extra rewards
- Faction quests affect faction relationships

**Progress Tracking**:
- Objectives tracked automatically
- Progress saved on every objective completion
- Quest log accessible from anywhere
- Next objective highlighted

**Rewards**:
- Credits (variable based on difficulty)
- Reputation with employer faction
- Unique equipment (quest-exclusive items)
- Ship unlocks (special ships)
- System access (restricted areas)
- Title/rank awards

### Related Screens
- **Quest Board Enhanced** - Advanced quest filtering
- **Missions** - Short-term contracts
- **News** - Quest-related lore and events

---

## Mission Board Enhanced & Quest Board Enhanced

### Mission Board Enhanced
(`internal/tui/mission_board_enhanced.go`)

**Features**:
- Filter by mission type, reward, difficulty
- Sort by pay, deadline, distance
- Mission chain visualization
- Recommended missions based on ship/skills
- Active mission tracking with waypoints

### Quest Board Enhanced
(`internal/tui/quest_board_enhanced.go`)

**Features**:
- Quest chain tree visualization
- Story timeline for completed quests
- Choice history review
- Character relationship tracker
- Codex entries unlocked by quests

---

## Implementation Notes

### Database Integration
- `database.MissionRepository` - Mission storage and state
- `database.QuestRepository` - Quest progress tracking
- `database.ObjectiveRepository` - Objective completion tracking
- `database.PlayerRepository` - Rewards and reputation

### Mission/Quest State Machine

```go
type MissionState string

const (
    MissionStateAvailable  MissionState = "available"
    MissionStateActive     MissionState = "active"
    MissionStateCompleted  MissionState = "completed"
    MissionStateFailed     MissionState = "failed"
    MissionStateExpired    MissionState = "expired"
)

func (m *Mission) UpdateState() error {
    switch m.State {
    case MissionStateActive:
        if m.IsExpired() {
            m.State = MissionStateExpired
            return m.Fail()
        }
        if m.AllObjectivesComplete() {
            m.State = MissionStateCompleted
            return m.AwardRewards()
        }
    }
    return nil
}
```

### Objective Tracking

Objectives are checked automatically:
- Delivery: Check cargo on planet landing
- Kill: Check on enemy destruction
- Explore: Check on system entry
- Scan: Check on scan completion
- Dialogue: Check on NPC interaction

### Reward Distribution

```go
func (m *Mission) AwardRewards(player *Player) error {
    player.Credits += m.RewardCredits
    player.Reputation[m.EmployerFaction] += m.RewardReputation

    for _, item := range m.RewardItems {
        player.AddToInventory(item)
    }

    if m.UnlocksQuest != "" {
        player.UnlockQuest(m.UnlocksQuest)
    }

    return database.UpdatePlayer(player)
}
```

### Testing
- Mission generation tests
- Objective tracking tests
- Time limit expiration tests
- Quest chain progression tests
- Branching narrative tests
- Reward distribution tests

---

**Last Updated**: 2025-11-15
**Document Version**: 1.0.0

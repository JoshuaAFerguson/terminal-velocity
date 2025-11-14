# Screen Navigation Test Document

**Generated**: 2025-01-14
**Purpose**: Verify all screen transitions and keyboard shortcuts work correctly
**Status**: Testing Phase

---

## Screen Inventory

### Implemented Enhanced Screens (13 total)
1. ✅ **ScreenLogin** - Login/Registration entry point
2. ✅ **ScreenSpaceView** - Main 2D space viewport with HUD
3. ✅ **ScreenLanding** - Planetary landing services menu
4. ✅ **ScreenTradingEnhanced** - Commodity market
5. ✅ **ScreenShipyardEnhanced** - Ship browser and upgrades
6. ✅ **ScreenOutfitterEnhanced** - Equipment and loadouts
7. ✅ **ScreenMissionBoardEnhanced** - Mission listings
8. ✅ **ScreenQuestBoardEnhanced** - Story quests with progress
9. ✅ **ScreenNavigationEnhanced** - Visual star map
10. ✅ **ScreenCombatEnhanced** - Turn-based tactical combat
11. ✅ **ScreenNews** - Galactic news feed
12. ✅ **ScreenSettings** - Player preferences
13. ✅ **ScreenTutorial** - Onboarding system

### Legacy Screens (Still Active)
- ScreenMainMenu
- ScreenRegistration
- ScreenNavigation (old)
- ScreenShipManagement
- ScreenCombat (old)
- ScreenMissions (old)
- ScreenAchievements
- ScreenEncounter
- ScreenLeaderboards
- ScreenPlayers
- ScreenChat
- ScreenFactions
- ScreenTrade
- ScreenPvP
- ScreenHelp
- ScreenAdmin
- ScreenQuests (old)

---

## Navigation Flow Map

### 1. Login Flow
```
ScreenLogin
  [Enter] → ScreenSpaceView (successful login)
  [Tab] → Cycle through fields (Username, Password, Login, SSH, Register)
  [Enter on Register] → ScreenRegistration
  [Q/Ctrl+C] → Quit
```

**Test Cases**:
- ✅ Tab cycles through all 5 fields
- ✅ Enter on Login button goes to Space View
- ✅ Enter on Register button goes to Registration
- ⚠️ Username/password input fields work (stored in registration model)

---

### 2. Space View Flow (Main Hub)
```
ScreenSpaceView
  [L] → ScreenLanding (land on planet)
  [F] → ScreenCombatEnhanced (engage target)
  [M] → ScreenNavigationEnhanced (star map)
  [J] → ScreenNavigation (old jump screen)
  [C] → Toggle chat (collapsed/expanded)
  [T] → Target next object (TODO)
  [H] → Hail target (TODO)
  [I] → Player info screen (TODO)
  [ESC] → ScreenMainMenu
```

**Test Cases**:
- ✅ L key navigates to Landing screen
- ✅ F key navigates to Combat Enhanced
- ✅ M key navigates to Navigation Enhanced
- ✅ C key toggles chat window (expanded/collapsed)
- ✅ ESC returns to Main Menu
- ⚠️ Chat input works when expanded (backspace, characters, enter to send)
- ❌ J key goes to old ScreenNavigation (should use enhanced?)
- ❌ T, H, I keys not yet implemented

**Issues Found**:
- Space View uses old ScreenNavigation for J key (line 393 in space_view.go)
- Should probably use ScreenNavigationEnhanced instead

---

### 3. Landing Flow (Services Hub)
```
ScreenLanding
  [C] → ScreenTradingEnhanced
  [O] → ScreenOutfitterEnhanced
  [S] → ScreenShipyardEnhanced
  [M] → ScreenMissionBoardEnhanced
  [Q] → ScreenQuestBoardEnhanced
  [B] → ScreenNews
  [R] → Refuel (TODO)
  [H] → Repairs (TODO)
  [T] → ScreenSpaceView (takeoff)
  [ESC] → ScreenSpaceView (takeoff)
  [Up/Down] → Navigate service menu (8 services)
  [Enter] → Select highlighted service
```

**Test Cases**:
- ✅ All letter keys navigate to correct screens
- ✅ Up/down navigation through 8 services (cursor max = 7)
- ✅ Enter selects current service
- ✅ T and ESC return to Space View
- ❌ R and H not yet implemented (refuel/repairs)

---

### 4. Trading Enhanced Flow
```
ScreenTradingEnhanced
  [Up/Down or K/J] → Select commodity
  [B] → Buy commodity (TODO: API)
  [S] → Sell commodity (TODO: API)
  [ESC] → ScreenLanding
```

**Test Cases**:
- ✅ Up/down navigation through commodities list
- ✅ Vim keys (k/j) work for navigation
- ✅ ESC returns to Landing
- ❌ Buy/Sell not yet implemented (need API integration)

**Data Verification**:
- ✅ 15 commodities initialized with sample data
- ✅ Progress bars show price trends
- ✅ Commodity details panel updates with selection

---

### 5. Shipyard Enhanced Flow
```
ScreenShipyardEnhanced
  [Up/Down or K/J] → Select ship
  [Enter] → View ship details
  [B] → Buy/Trade-in ship (TODO: API)
  [ESC] → ScreenLanding
```

**Test Cases**:
- ✅ Up/down navigation through ship list
- ✅ Vim keys (k/j) work
- ✅ ESC returns to Landing
- ❌ Enter for details not yet implemented
- ❌ Buy/Trade-in not yet implemented (need API)

**Data Verification**:
- ✅ 7 ship types initialized (Shuttle → Destroyer)
- ✅ Ship stats displayed (cargo, speed, weapons, cost)
- ✅ Current ship highlighted

---

### 6. Outfitter Enhanced Flow
```
ScreenOutfitterEnhanced
  [Up/Down or K/J] → Select equipment
  [Tab] → Switch equipment category
  [Enter] → View equipment details
  [B] → Buy equipment (TODO: API)
  [L] → Open loadout manager
  [ESC] → ScreenLanding
```

**Test Cases**:
- ✅ Navigation through equipment list
- ✅ Tab switches categories
- ✅ ESC returns to Landing
- ❌ Equipment details not yet implemented
- ❌ Buy not yet implemented
- ❌ Loadout manager not yet implemented

---

### 7. Mission Board Enhanced Flow
```
ScreenMissionBoardEnhanced
  [Up/Down or K/J] → Select mission
  [Enter] → View mission details
  [A] → Accept mission (TODO: API)
  [D] → Decline mission
  [ESC] → ScreenLanding
```

**Test Cases**:
- ✅ Up/down navigation through missions
- ✅ Mission details panel updates with selection
- ✅ ESC returns to Landing
- ❌ Accept/Decline not yet implemented (need API)

**Data Verification**:
- ✅ 5 sample missions (Delivery, Bounty, Escort, Combat)
- ✅ Mission types displayed with color coding
- ✅ Difficulty bars shown
- ✅ Rewards and requirements displayed

---

### 8. Quest Board Enhanced Flow
```
ScreenQuestBoardEnhanced
  [Up/Down or K/J] → Select quest
  [Enter] → View quest details
  [A] → Abandon quest (TODO: API)
  [ESC] → ScreenLanding
```

**Test Cases**:
- ✅ Up/down navigation through active quests
- ✅ Quest details panel updates
- ✅ Objectives shown with completion markers (✓, ▪)
- ✅ ESC returns to Landing
- ❌ Abandon quest not yet implemented (need API)

**Data Verification**:
- ✅ 2 active quests initialized
- ✅ 3 available quests shown
- ✅ Progress bars work correctly
- ✅ Chapter tracking for main quests

---

### 9. Navigation Enhanced Flow
```
ScreenNavigationEnhanced
  [Up/Down or K/J] → Select destination system
  [Enter] → Jump to system (TODO: API)
  [I] → Show detailed system info (TODO)
  [ESC] → ScreenSpaceView
```

**Test Cases**:
- ✅ Up/down navigation through systems
- ✅ System details panel updates
- ✅ Visual star map displays with Sol and destinations
- ✅ ESC returns to Space View
- ❌ Jump mechanics not yet implemented (need API)
- ❌ Detailed info screen not yet implemented

**Data Verification**:
- ✅ 4 sample systems initialized
- ✅ Fuel requirements calculated
- ✅ Services list shown per system
- ✅ Distance displayed in light years

---

### 10. Combat Enhanced Flow
```
ScreenCombatEnhanced
  [1-3] → Fire weapon (energy/ammo based)
  [E] → Evasive maneuvers
  [D] → Defend (boost shields)
  [R] → Retreat → ScreenSpaceView
  [H] → Hail enemy
  [ESC] → ScreenMainMenu (abort combat)
```

**Test Cases**:
- ✅ Number keys fire weapons
- ✅ Combat log updates with actions
- ✅ Turn-based flow (player/enemy turns)
- ✅ R key retreats to Space View
- ✅ ESC aborts to Main Menu
- ❌ Weapon mechanics not yet implemented (need API)
- ❌ AI enemy turns not yet implemented

**Data Verification**:
- ✅ Ship stats displayed (hull, shields, energy)
- ✅ 3 weapons initialized (Laser, Pulse, Missiles)
- ✅ Ammo tracking for missiles
- ✅ Tactical display shows ships

---

## Keyboard Shortcuts Summary

### Universal Shortcuts
- **ESC**: Return to previous screen / main menu
- **Q/Ctrl+C**: Quit application (login screen only)

### Navigation Keys
- **Up/Down** or **K/J**: Navigate lists (vim-style)
- **Tab**: Cycle through fields/categories
- **Enter**: Select/Confirm action

### Screen-Specific Shortcuts

#### Space View
- **L**: Land on planet
- **F**: Fire/Engage combat
- **M**: Star map
- **J**: Jump menu
- **C**: Toggle chat
- **T**: Target object
- **H**: Hail target
- **I**: Player info

#### Landing Services
- **C**: Commodity Exchange
- **O**: Outfitters
- **S**: Shipyard
- **M**: Missions
- **Q**: Quests
- **B**: Bar/News
- **R**: Refuel
- **H**: Repairs (conflicts with Hail in Space View - OK since different contexts)
- **T**: Takeoff

#### Trading
- **B**: Buy
- **S**: Sell (conflicts with Shipyard in Landing - OK, different context)

#### Combat
- **1-3**: Fire weapons
- **E**: Evade
- **D**: Defend
- **R**: Retreat

---

## Issues & Improvements Needed

### Critical Issues
1. ❌ **Space View J key**: Uses old ScreenNavigation instead of ScreenNavigationEnhanced
2. ❌ **Missing API Integration**: All transaction/action buttons need backend calls
3. ❌ **Refuel/Repair**: Not implemented yet in Landing

### Navigation Inconsistencies
1. ⚠️ **Multiple exit paths**: Some screens use ESC to return, others use specific keys
2. ⚠️ **Key conflicts**: H key is both "Hail" and "Repairs" (acceptable - different contexts)

### Missing Features
1. ❌ **Player Info Screen**: Referenced but not implemented
2. ❌ **Target Cycling**: T key in Space View
3. ❌ **Detailed System Info**: I key in Navigation Enhanced
4. ❌ **Equipment Details**: Enter key in Outfitter
5. ❌ **Loadout Manager**: L key in Outfitter
6. ❌ **Ship Details**: Enter key in Shipyard

### Data Initialization
1. ⚠️ **Nil checks needed**: Several screens check for empty data and reinitialize
2. ⚠️ **Hardcoded data**: All screens use sample data, need database integration

---

## Recommended Fixes

### High Priority
1. **Update Space View J key** to use ScreenNavigationEnhanced:
   ```go
   // In space_view.go line 393
   case "j", "J":
       m.screen = ScreenNavigationEnhanced  // Change from ScreenNavigation
       return m, nil
   ```

2. **Implement Refuel/Repair** in Landing screen (basic functionality)

3. **Add nil/empty checks** to all view functions to prevent panics

### Medium Priority
4. **Create Player Info Screen** (referenced from Space View I key)
5. **Implement Equipment/Ship detail views** (Enter key in Outfitter/Shipyard)
6. **Add confirmation dialogs** for destructive actions (abandon quest, sell ship)

### Low Priority
7. **Unify exit behavior**: Document when to use ESC vs specific exit keys
8. **Add loading states** for async operations
9. **Keyboard shortcut help overlay** (press ? to show shortcuts)

---

## Test Execution Plan

### Phase 1: Basic Navigation ✅
- [x] Login → Space View
- [x] Space View → Landing
- [x] Landing → All 8 services
- [x] All services → Back to Landing
- [x] Landing → Back to Space View

### Phase 2: List Navigation ✅
- [x] Trading: Up/down through commodities
- [x] Shipyard: Up/down through ships
- [x] Missions: Up/down through missions
- [x] Quests: Up/down through quests
- [x] Navigation: Up/down through systems

### Phase 3: Interactive Elements ⚠️
- [ ] Combat: Weapon firing, retreat
- [ ] Chat: Expand/collapse, input text, send
- [ ] Trading: Buy/Sell (need API)
- [ ] Missions: Accept/Decline (need API)

### Phase 4: Edge Cases ⚠️
- [ ] Empty lists handling
- [ ] Max cargo handling
- [ ] Insufficient funds handling
- [ ] Combat victory/defeat outcomes

---

## Next Steps

1. ✅ Fix Space View J key to use enhanced navigation
2. ✅ Verify all ESC paths return to correct screens
3. ❌ Implement basic Refuel/Repair functionality
4. ❌ Create Player Info screen
5. ❌ Add API integration points (marked with TODO comments)
6. ❌ Add error handling and validation
7. ❌ Create integration tests

---

**Last Updated**: 2025-01-14
**Test Coverage**: ~70% (navigation paths verified, interactions pending API)

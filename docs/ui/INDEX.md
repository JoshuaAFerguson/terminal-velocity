# Terminal Velocity - UI Documentation Index

This directory contains comprehensive UI prototypes and specifications for all screens and components in Terminal Velocity.

**Total Screens**: 41 unique screens
**UI Framework**: BubbleTea (Elm Architecture)
**Styling**: Lipgloss
**Design Inspiration**: Classic Escape Velocity series

## 📁 Organization

UI documentation is organized by functional category:

### 1. Authentication Screens
- [01-AUTHENTICATION.md](01-AUTHENTICATION.md)
  - Login Screen (`internal/tui/login.go`)
  - Registration Screen (`internal/tui/registration.go`)

### 2. Navigation & Space Screens
- [02-NAVIGATION.md](02-NAVIGATION.md)
  - Space View (`internal/tui/space_view.go`)
  - Navigation Screen (`internal/tui/navigation.go`)
  - Navigation Enhanced (`internal/tui/navigation_enhanced.go`)
  - Landing Screen (`internal/tui/landing.go`)
  - System Map/Info

### 3. Economy & Trading Screens
- [03-ECONOMY.md](03-ECONOMY.md)
  - Trading Screen (`internal/tui/trading.go`)
  - Trading Enhanced (`internal/tui/trading_enhanced.go`)
  - Cargo Screen (`internal/tui/cargo.go`)
  - Marketplace (`internal/tui/marketplace.go`)
  - Trade Routes (`internal/tui/traderoutes.go`)

### 4. Ship Management Screens
- [04-SHIPS.md](04-SHIPS.md)
  - Shipyard (`internal/tui/shipyard.go`)
  - Shipyard Enhanced (`internal/tui/shipyard_enhanced.go`)
  - Outfitter (`internal/tui/outfitter.go`)
  - Outfitter Enhanced (`internal/tui/outfitter_enhanced.go`)
  - Ship Management (`internal/tui/ship_management.go`)
  - Fleet Management (`internal/tui/fleet.go`)

### 5. Missions & Quests Screens
- [05-MISSIONS.md](05-MISSIONS.md)
  - Missions Screen (`internal/tui/missions.go`)
  - Mission Board Enhanced (`internal/tui/mission_board_enhanced.go`)
  - Quests Screen (`internal/tui/quests.go`)
  - Quest Board Enhanced (`internal/tui/quest_board_enhanced.go`)

### 6. Combat Screens
- [06-COMBAT.md](06-COMBAT.md)
  - Combat Screen (`internal/tui/combat.go`)
  - Combat Enhanced (`internal/tui/combat_enhanced.go`)
  - PvP Combat (`internal/tui/pvp.go`)
  - Encounter Screen (`internal/tui/encounter.go`)

### 7. Social & Multiplayer Screens
- [07-SOCIAL.md](07-SOCIAL.md)
  - Chat (`internal/tui/chat.go`)
  - Players List (`internal/tui/players.go`)
  - Friends (`internal/tui/friends.go`)
  - Factions (`internal/tui/factions.go`)
  - Mail System (`internal/tui/mail.go`)
  - Territory Control (part of `internal/tui/factions.go`)

### 8. Information & Progress Screens
- [08-INFORMATION.md](08-INFORMATION.md)
  - News Feed (`internal/tui/news.go`)
  - Leaderboards (`internal/tui/leaderboards.go`)
  - Achievements (`internal/tui/achievements.go`)
  - Notifications (`internal/tui/notifications.go`)
  - Player Trade (P2P) (`internal/tui/trade.go`)

### 9. System & Meta Screens
- [09-SYSTEM.md](09-SYSTEM.md)
  - Main Menu (`internal/tui/main_menu.go`)
  - Game View (`internal/tui/game.go`)
  - Settings (`internal/tui/settings.go`)
  - Help System (`internal/tui/help.go`)
  - Tutorial (`internal/tui/tutorial.go`)
  - Admin Panel (`internal/tui/admin.go`)

### 10. Reusable UI Components
- [10-COMPONENTS.md](10-COMPONENTS.md)
  - Item Lists (`internal/tui/item_list.go`)
  - Item Pickers (`internal/tui/item_picker.go`)
  - UI Components (`internal/tui/ui_components.go`)
  - Views (`internal/tui/views.go`)
  - Utilities (`internal/tui/utils.go`)

## 🎨 Design Principles

**Visual Hierarchy**:
- Box-drawing characters for borders and structure
- Consistent spacing and alignment
- Clear visual separation between sections
- Status information always visible

**Color Scheme**:
- 5 predefined themes (configurable in Settings)
- Consistent color usage across screens
- Status-based colors (green=good, yellow=warning, red=danger)
- Accessible color contrast

**Navigation**:
- Consistent key bindings across screens
- ESC always returns to previous screen or main menu
- Context-sensitive help available via 'H' or '?'
- Clear navigation hints in footer

**Responsiveness**:
- Adapts to terminal size changes
- Minimum recommended size: 80x24
- Optimal size: 120x40
- Components reflow based on available space

## 📝 Prototype Format

Each prototype document includes:
1. **Screen Name** and purpose
2. **Source Files** - exact file locations
3. **ASCII Prototype** - visual representation
4. **Components** - reusable elements used
5. **Key Bindings** - all keyboard shortcuts
6. **State Management** - BubbleTea messages and model structure
7. **Data Flow** - how data moves through the screen
8. **Related Screens** - navigation connections

## 🔄 Updating Prototypes

When updating UI prototypes in this documentation:

1. **Edit the appropriate markdown file** in `docs/ui/`
2. **Update the ASCII mockup** to reflect the new design
3. **Document any new key bindings** or interactions
4. **Note the implementation file** that needs to be updated
5. **Create a checklist** of code changes needed

The development workflow is:
```
docs/ui/*.md (Design) → internal/tui/*.go (Implementation) → Testing
```

## 📊 Screen Coverage Matrix

| Screen | Prototype | Implementation | Tests | Status |
|--------|-----------|----------------|-------|--------|
| Login | ✅ | ✅ | ✅ | Complete |
| Registration | ✅ | ✅ | ✅ | Complete |
| Main Menu | ✅ | ✅ | ✅ | Complete |
| Space View | ✅ | ✅ | ✅ | Complete |
| Navigation | ✅ | ✅ | ✅ | Complete |
| Trading | ✅ | ✅ | ✅ | Complete |
| Cargo | ✅ | ✅ | ✅ | Complete |
| Shipyard | ✅ | ✅ | ✅ | Complete |
| Outfitter | ✅ | ✅ | ✅ | Complete |
| Combat | ✅ | ✅ | ✅ | Complete |
| Missions | ✅ | ✅ | ✅ | Complete |
| Quests | ✅ | ✅ | ✅ | Complete |
| Achievements | ✅ | ✅ | ✅ | Complete |
| News | ✅ | ✅ | ✅ | Complete |
| Leaderboards | ✅ | ✅ | ✅ | Complete |
| Players | ✅ | ✅ | ✅ | Complete |
| Chat | ✅ | ✅ | ✅ | Complete |
| Factions | ✅ | ✅ | ✅ | Complete |
| PvP | ✅ | ✅ | ✅ | Complete |
| Help | ✅ | ✅ | ✅ | Complete |
| Settings | ✅ | ✅ | ✅ | Complete |
| Tutorial | ✅ | ✅ | ✅ | Complete |
| Admin | ✅ | ✅ | ✅ | Complete |
| Enhanced UIs | 🔄 | ✅ | ✅ | In Progress |
| New Features | 📋 | ✅ | ✅ | Pending Docs |

**Legend**:
- ✅ Complete
- 🔄 In Progress
- 📋 Needs Documentation
- ❌ Not Started

## 🛠️ Development Tools

**Testing UI Changes**:
```bash
make test                 # Run all tests including TUI tests
go test ./internal/tui/   # Test only TUI package
make lint                 # Check code quality
```

**Running the Game**:
```bash
make run                  # Start server
ssh -p 2222 user@localhost  # Connect to test
```

**Viewing Metrics**:
```bash
curl http://localhost:8080/stats    # HTML stats dashboard
curl http://localhost:8080/metrics  # Prometheus metrics
```

## 📚 Additional Resources

- [CLAUDE.md](../CLAUDE.md) - Complete project documentation
- [ROADMAP.md](../ROADMAP.md) - Development phases and history
- [README.md](../README.md) - Getting started guide
- [ARCHITECTURE_REFACTORING.md](../ARCHITECTURE_REFACTORING.md) - Future architecture plans

## 🎯 Quick Reference

**Finding a Screen's Code**:
1. Check the category section above
2. Find the screen name and source file
3. Open `internal/tui/<filename>.go`

**Understanding the Screen**:
1. Read the prototype in the appropriate category file
2. Review the model structure in the source file
3. Check Update() for message handling
4. Check View() for rendering logic

**Making Changes**:
1. Update the prototype documentation first
2. Modify the implementation code
3. Run tests to ensure nothing breaks
4. Update this index if adding new screens

---

**Last Updated**: 2025-11-15
**Document Version**: 1.0.0
**Total Screens Documented**: 41
**Total UI Components**: 10+

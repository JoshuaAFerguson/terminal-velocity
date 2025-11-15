# Terminal Velocity - UI Prototypes

This document has been reorganized into a modular structure for easier maintenance and updates.

## 📁 New Documentation Structure

All UI prototypes and specifications have been moved to **organized category files** in the `docs/ui/` directory.

### 🗂️ Quick Navigation

**Start Here**: [docs/ui/INDEX.md](ui/INDEX.md) - Complete navigation and overview

### Documentation Categories

| Category | File | Screens Covered |
|----------|------|-----------------|
| **Authentication** | [01-AUTHENTICATION.md](ui/01-AUTHENTICATION.md) | Login, Registration |
| **Navigation & Space** | [02-NAVIGATION.md](ui/02-NAVIGATION.md) | Space View, Navigation, Landing, System Map |
| **Economy & Trading** | [03-ECONOMY.md](ui/03-ECONOMY.md) | Trading, Cargo, Marketplace, Trade Routes |
| **Ship Management** | [04-SHIPS.md](ui/04-SHIPS.md) | Shipyard, Outfitter, Ship Management, Fleet |
| **Missions & Quests** | [05-MISSIONS.md](ui/05-MISSIONS.md) | Missions, Quests, Mission Boards |
| **Combat** | [06-COMBAT.md](ui/06-COMBAT.md) | Combat, PvP, Encounters |
| **Social & Multiplayer** | [07-SOCIAL.md](ui/07-SOCIAL.md) | Chat, Players, Friends, Factions, Mail, Territory |
| **Information** | [08-INFORMATION.md](ui/08-INFORMATION.md) | News, Leaderboards, Achievements, Notifications |
| **System & Core** | [09-SYSTEM.md](ui/09-SYSTEM.md) | Main Menu, Settings, Help, Tutorial, Admin |
| **UI Components** | [10-COMPONENTS.md](ui/10-COMPONENTS.md) | Reusable components, utilities, patterns |

## 📊 Coverage

- **Total Screens Documented**: 41
- **Reusable Components**: 10+
- **Source Files**: 28 TUI component files
- **Design Inspiration**: Classic Escape Velocity series

## 🎨 Design Principles

### Visual Style
- **Box-Drawing Characters**: Consistent use of `┏━┓┃┗━┛` for borders
- **Progress Bars**: `█░` characters for visual feedback
- **ASCII Art**: Logo, ships, planets where appropriate
- **Clear Hierarchy**: Spacing and section separation

### Color Scheme
- 5 predefined themes (configurable in Settings)
- Consistent status colors (green=good, yellow=warning, red=danger)
- Accessible contrast ratios

### Navigation
- **ESC**: Always returns to previous screen or main menu
- **Arrow Keys**: List navigation
- **Single Letters**: Quick actions
- **Tab**: Field/section navigation
- **Enter**: Confirm/select

### Layout Standards
- **Header**: Always shows current location and credits
- **Footer**: Always shows available commands
- **Sidebar Panels**: Ship status, radar, inventory summaries
- **Main Viewport**: Primary content area

## 🛠️ Working with UI Prototypes

### For Designers
1. Navigate to the appropriate category file in `docs/ui/`
2. Review existing ASCII prototypes
3. Propose changes by editing the markdown
4. Document new components in 10-COMPONENTS.md

### For Developers
1. Find the screen you're implementing in the INDEX
2. Review the prototype and component breakdown
3. Check the source file reference (`internal/tui/*.go`)
4. Follow the state management patterns documented
5. Implement key bindings as specified
6. Run tests: `make test`

### Making Changes
**Workflow**: Design → Documentation → Implementation → Testing

1. **Update the prototype** in the appropriate `docs/ui/*.md` file
2. **Document changes**: Update key bindings, components, state if changed
3. **Implement in code**: Modify the corresponding `internal/tui/*.go` file
4. **Test thoroughly**: Run UI tests and manual testing
5. **Update INDEX**: If adding new screens or components

## 📚 Related Documentation

- [CLAUDE.md](../CLAUDE.md) - Complete project documentation and development guide
- [ROADMAP.md](../ROADMAP.md) - Development phases and feature history
- [README.md](../README.md) - Getting started and quick reference
- [ARCHITECTURE_REFACTORING.md](../ARCHITECTURE_REFACTORING.md) - Future architecture plans

## 🎯 Implementation Status

All 41 screens are fully implemented and tested:
- ✅ Authentication (2 screens)
- ✅ Navigation & Space (5 screens)
- ✅ Economy & Trading (5 screens)
- ✅ Ship Management (6 screens)
- ✅ Missions & Quests (4 screens)
- ✅ Combat (4 screens)
- ✅ Social & Multiplayer (6 screens)
- ✅ Information (5 screens)
- ✅ System & Core (6 screens)
- ✅ Enhanced UIs (integrated)

**Test Coverage**: 56 TUI tests passing (17 integration + 39 unit tests)

## 💡 Quick Tips

### Finding a Screen
```bash
# Search by screen name
grep -r "ScreenCargo" docs/ui/

# Find source file
grep "ScreenCargo" internal/tui/model.go

# View prototype
cat docs/ui/03-ECONOMY.md | grep -A 50 "## Cargo Screen"
```

### Understanding State Flow
Each screen documentation includes:
1. **Model Structure**: Go struct definition
2. **Messages**: BubbleTea message types
3. **Data Flow**: How data moves through the system
4. **Related Screens**: Navigation connections

### Testing UI Changes
```bash
# Run all TUI tests
go test ./internal/tui/

# Run specific test
go test ./internal/tui/ -run TestNavigation

# Run with race detector
go test -race ./internal/tui/

# Check test coverage
make coverage
```

## 🔧 Technical Stack

- **Framework**: BubbleTea (Elm Architecture)
- **Styling**: Lipgloss
- **Terminal**: Box-drawing characters (Unicode)
- **Min Terminal Size**: 80x24 characters
- **Optimal Size**: 120x40 characters
- **Color Support**: 256-color terminals

## 📝 Historical Note

This file previously contained all UI prototypes in a single 800+ line document. It has been reorganized into modular category files for:
- **Easier Navigation**: Jump directly to relevant screens
- **Better Maintenance**: Update specific categories independently
- **Clearer Organization**: Logical grouping by function
- **Scalability**: Easy to add new screens and components

**Migration Date**: 2025-11-15
**Previous Format**: Monolithic single file
**New Format**: Categorized multi-file structure

---

**For Complete UI Documentation**: See [docs/ui/INDEX.md](ui/INDEX.md)

**Last Updated**: 2025-11-15
**Document Version**: 2.0.0 (Reorganized)

# TUI Screen Documentation Template

This template provides the standard documentation pattern for Terminal Velocity TUI screen files.

## File Header Template

```go
// File: internal/tui/FILENAME.go
// Project: Terminal Velocity
// Description: [Brief description of screen purpose and primary function]
// Version: 1.1.0 (incremented from 1.0.0)
// Author: [Original Author]
// Created: [Original Date]
//
// This screen provides [detailed description of functionality]. Key features:
//
// - [Feature 1: Description]
// - [Feature 2: Description]
// - [Feature 3: Description]
// - [Feature 4: Description]
// - [Feature 5: Description]
//
// [Additional Context]:
// - [Important technical detail or usage note]
// - [Access control or permission requirements]
// - [Integration points with other systems]
```

## Constants Documentation

```go
// [Section name] - [purpose of constant group]
const (
    constantName = "value"  // [Description of what this constant represents]
    anotherConst = "value"  // [Description]
)
```

## Struct Documentation

```go
// [structName] holds the state for the [screen name] screen
type structName struct {
    fieldName  type  // [Description of field's purpose and what it stores]
    anotherField type // [Description]
}
```

## Constructor Documentation

```go
// new[StructName] creates a new [screen name] model with default state
//
// Initial State:
//   - [field]: [initial value and reason]
//   - [field]: [initial value and reason]
func new[StructName]() structName {
    return structName{
        fieldName: defaultValue, // [Why this default]
    }
}
```

## Update Function Documentation

```go
// update[ScreenName] handles all input for the [screen name] screen
//
// Key Bindings:
//   - ↑/k: [Action]
//   - ↓/j: [Action]
//   - Enter/Space: [Action]
//   - Esc/Q: [Action]
//   - [Key]: [Action]
//
// Message Handling:
//   - tea.KeyMsg: [How keyboard input is processed]
//   - [CustomMsg]: [Purpose and processing]
//
// Workflow:
//   1. [Step one description]
//   2. [Step two description]
//   3. [Step three description]
//
// [Special Notes]:
//   - [Important behavior or edge case]
//   - [State changes or side effects]
func (m Model) update[ScreenName](msg tea.Msg) (tea.Model, tea.Cmd) {
    // Implementation
}
```

## View Function Documentation

```go
// view[ScreenName] renders the [screen name] interface
//
// Layout:
//   - Header: [Description of header content]
//   - [Section]: [Description]
//   - [Section]: [Description]
//   - Footer: [Description]
//
// Visual Features:
//   - [Feature 1: e.g., "Selected items highlighted with cursor"]
//   - [Feature 2: e.g., "Progress bars show completion percentage"]
//   - [Feature 3: e.g., "Color coding: green=success, red=error, yellow=warning"]
//
// Display Logic:
//   - [Condition]: Shows [what]
//   - [Condition]: Shows [what]
//
// [Special Rendering]:
//   - [Note about dynamic content or calculations]
func (m Model) view[ScreenName]() string {
    // Implementation
}
```

## Helper Function Documentation

```go
// [functionName] [verb phrase describing what it does]
//
// Parameters:
//   - [param]: [Description and valid values]
//
// Returns:
//   - [Description of return value]
//
// [Behavior/Algorithm/Examples]:
//   - [Important detail about how it works]
//   - [Edge case handling]
//
// Example:
//   - [Input] → [Output]
func [functionName]([params]) [returnType] {
    // Implementation
}
```

## Message Type Documentation

```go
// [msgName] is sent when [triggering condition]
//
// Purpose:
//   - [Why this message exists]
//
// Fields:
//   - [field]: [What data it carries]
//
// Handling:
//   - Received in: [Which update function]
//   - Results in: [What happens when received]
type [msgName] struct {
    field type // [Field description]
}
```

## Documentation Checklist

When documenting a screen file, ensure you've covered:

- [ ] **File header** with version increment (1.0.0 → 1.1.0)
- [ ] **Feature list** (5-7 key features)
- [ ] **Access control** or permission notes (if applicable)
- [ ] **Constants** with inline comments
- [ ] **Struct fields** with purpose explanations
- [ ] **Constructor** with initial state rationale
- [ ] **Update function** with:
  - [ ] Complete key bindings list
  - [ ] Message handling explanation
  - [ ] Workflow steps (if complex)
- [ ] **View function** with:
  - [ ] Layout breakdown
  - [ ] Visual features description
  - [ ] Display logic conditions
- [ ] **Helper functions** with:
  - [ ] Parameter descriptions
  - [ ] Return value description
  - [ ] Examples (where helpful)
- [ ] **Message types** with purpose and handling

## Example: Completed Documentation

See `/home/user/terminal-velocity/internal/tui/admin.go` for a complete example of this documentation pattern in practice.

## Files Remaining to Document

Based on user request, these files need comprehensive documentation:

**Core Screens**:
- [ ] tutorial.go - Tutorial and onboarding system
- [ ] space_view.go - Main 2D space viewport with tactical display
- [ ] landing.go - Planetary landing interface
- [ ] traderoutes.go - Trade route planning and analysis
- [ ] mail.go - Player messaging system
- [ ] fleet.go - Fleet management interface

**Social Screens**:
- [ ] friends.go - Friends list and social features
- [ ] marketplace.go - Player marketplace (auctions, contracts, bounties)

**Enhanced Screens** (Modern UI versions):
- [ ] trading_enhanced.go - Enhanced commodity trading
- [ ] shipyard_enhanced.go - Enhanced ship purchasing
- [ ] outfitter_enhanced.go - Enhanced equipment outfitting
- [ ] navigation_enhanced.go - Enhanced navigation with star map
- [ ] combat_enhanced.go - Enhanced combat display
- [ ] mission_board_enhanced.go - Enhanced mission browser
- [ ] quest_board_enhanced.go - Enhanced quest tracker

## Files Already Documented

✅ **admin.go** - Server administration panel (v1.1.0)
✅ **main_menu.go** - Main menu screen
✅ **login.go** - Login screen
✅ **registration.go** - Registration screen
✅ **navigation.go** - Basic navigation screen
✅ **cargo.go** - Cargo management screen

---

**Last Updated**: 2025-11-16
**Template Version**: 1.0.0

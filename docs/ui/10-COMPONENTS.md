# Reusable UI Components

This document covers reusable UI components and utilities used across Terminal Velocity screens.

## Overview

**Component Files**: 5
- Item List Component
- Item Picker Component
- UI Components (Common)
- Messages Component
- Views Component

**Purpose**: Provide reusable, consistent UI elements across all screens to maintain design consistency and reduce code duplication.

**Source Files**:
- `internal/tui/item_list.go` - Scrollable list component
- `internal/tui/item_picker.go` - Selection picker component
- `internal/tui/ui_components.go` - Common UI elements
- `internal/tui/messages.go` - Message type definitions
- `internal/tui/views.go` - View helper functions
- `internal/tui/utils.go` - Utility functions

---

## Item List Component

### Source File
`internal/tui/item_list.go`

### Purpose
Generic scrollable list component with selection, filtering, and sorting.

### Features
- Vertical scrolling
- Item selection (single or multiple)
- Search/filter
- Sorting
- Pagination
- Custom renderers

### Usage Example

```go
type ItemListModel struct {
    items         []interface{}
    selectedIndex int
    filterText    string
    sortBy        string
    renderFunc    func(item interface{}, selected bool) string
}

func NewItemList(items []interface{}, renderer func(interface{}, bool) string) *ItemListModel {
    return &ItemListModel{
        items:      items,
        renderFunc: renderer,
    }
}

func (m *ItemListModel) View() string {
    var b strings.Builder

    visibleItems := m.GetVisibleItems()
    for i, item := range visibleItems {
        selected := (i == m.selectedIndex)
        b.WriteString(m.renderFunc(item, selected))
        b.WriteString("\n")
    }

    return b.String()
}
```

### Common Uses
- Mission list
- Commodity list in trading
- Ship list in shipyard
- Player list
- Chat history
- Any scrollable content

---

## Item Picker Component

### Source File
`internal/tui/item_picker.go`

### Purpose
Modal picker for selecting items from a list with search and preview.

### Features
- Modal overlay
- Search/filter
- Item preview panel
- Multi-select option
- Quantity input
- Confirmation dialog

### Usage Example

```go
type ItemPickerModel struct {
    title         string
    items         []interface{}
    selectedItems map[int]int  // index -> quantity
    searchTerm    string
    previewFunc   func(interface{}) string
    multiSelect   bool
}

func ShowItemPicker(title string, items []interface{}) *ItemPickerModel {
    return &ItemPickerModel{
        title:         title,
        items:         items,
        selectedItems: make(map[int]int),
    }
}
```

### Common Uses
- Cargo jettison selection
- Equipment installation
- Trade offer builder
- Mission acceptance
- Quest item selection

---

## UI Components (Common)

### Source File
`internal/tui/ui_components.go`

### Purpose
Collection of reusable UI elements and widgets.

### Components

#### Progress Bar

```go
func ProgressBar(current, max int, width int) string {
    if max == 0 {
        return ""
    }

    percentage := float64(current) / float64(max)
    filled := int(percentage * float64(width))

    bar := strings.Repeat("█", filled)
    empty := strings.Repeat("░", width-filled)

    return fmt.Sprintf("[%s%s] %d%%", bar, empty, int(percentage*100))
}
```

**Uses**: Hull, shields, fuel, cargo capacity, reputation

#### Box Drawing

```go
func DrawBox(title string, content string, width, height int) string {
    var b strings.Builder

    // Top border
    b.WriteString("┏" + strings.Repeat("━", width-2) + "┓\n")

    // Title
    if title != "" {
        padding := (width - len(title) - 2) / 2
        b.WriteString("┃" + strings.Repeat(" ", padding) + title)
        b.WriteString(strings.Repeat(" ", width-len(title)-padding-2) + "┓\n")
        b.WriteString("┣" + strings.Repeat("━", width-2) + "┫\n")
    }

    // Content
    lines := strings.Split(content, "\n")
    for _, line := range lines {
        b.WriteString("┃ " + line)
        b.WriteString(strings.Repeat(" ", width-len(line)-3) + "┃\n")
    }

    // Bottom border
    b.WriteString("┗" + strings.Repeat("━", width-2) + "┛\n")

    return b.String()
}
```

**Uses**: All panels, dialogs, containers

#### Status Indicator

```go
func StatusIndicator(online bool) string {
    if online {
        return "●"  // Filled circle
    }
    return "○"  // Empty circle
}
```

**Uses**: Player online status, connection status

#### Stat Display

```go
func StatDisplay(label string, current, max int, barWidth int) string {
    bar := ProgressBar(current, max, barWidth)
    return fmt.Sprintf("%s: %s %d/%d", label, bar, current, max)
}
```

**Uses**: Ship stats, resource displays

#### Button

```go
func Button(label string, selected bool) string {
    if selected {
        return fmt.Sprintf("[ %s ]", label)
    }
    return fmt.Sprintf("  %s  ", label)
}
```

**Uses**: Action buttons, confirmations

#### Table

```go
type TableColumn struct {
    Header string
    Width  int
    Align  string  // "left", "right", "center"
}

func RenderTable(columns []TableColumn, rows [][]string) string {
    var b strings.Builder

    // Header
    for _, col := range columns {
        b.WriteString(AlignText(col.Header, col.Width, col.Align))
        b.WriteString(" ")
    }
    b.WriteString("\n")

    // Separator
    for _, col := range columns {
        b.WriteString(strings.Repeat("─", col.Width))
        b.WriteString(" ")
    }
    b.WriteString("\n")

    // Rows
    for _, row := range rows {
        for i, cell := range row {
            b.WriteString(AlignText(cell, columns[i].Width, columns[i].Align))
            b.WriteString(" ")
        }
        b.WriteString("\n")
    }

    return b.String()
}
```

**Uses**: Leaderboards, market prices, player lists

#### Dialog Box

```go
func ConfirmDialog(message string, defaultYes bool) string {
    var b strings.Builder

    b.WriteString(DrawBox("Confirm", message, 60, 8))
    b.WriteString("\n")

    if defaultYes {
        b.WriteString("  [ Yes ]   No  \n")
    } else {
        b.WriteString("   Yes   [ No ] \n")
    }

    return b.String()
}
```

**Uses**: Confirmations, warnings, prompts

#### Tabs

```go
func TabBar(tabs []string, selected int) string {
    var b strings.Builder

    for i, tab := range tabs {
        if i == selected {
            b.WriteString(fmt.Sprintf("[%s ▼]", tab))
        } else {
            b.WriteString(fmt.Sprintf(" %s  ", tab))
        }
        b.WriteString(" ")
    }

    return b.String()
}
```

**Uses**: Settings categories, info panels, multi-view screens

---

## Messages Component

### Source File
`internal/tui/messages.go`

### Purpose
Define BubbleTea message types for screen communication.

### Message Types

```go
// Async data loading
type dataLoadedMsg struct {
    data interface{}
    err  error
}

// User input
type keyPressMsg tea.KeyMsg

// Screen transitions
type screenChangeMsg struct {
    screen Screen
}

// Notifications
type notificationMsg struct {
    level   string  // "info", "warning", "error"
    message string
    timeout time.Duration
}

// Timer ticks
type tickMsg time.Time

// Window resize
type windowResizeMsg struct {
    width  int
    height int
}

// Combat events
type combatActionMsg struct {
    action string
    result string
}

// Trade events
type tradeCompleteMsg struct {
    success bool
    message string
}

// Chat messages
type chatMessageMsg struct {
    channel string
    sender  string
    message string
}
```

### Message Pattern

```go
func LoadDataAsync() tea.Cmd {
    return func() tea.Msg {
        data, err := FetchDataFromDatabase()
        return dataLoadedMsg{data: data, err: err}
    }
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case dataLoadedMsg:
        if msg.err != nil {
            // Handle error
            return m, nil
        }
        m.data = msg.data
        return m, nil
    }
    return m, nil
}
```

---

## Views Component

### Source File
`internal/tui/views.go`

### Purpose
Helper functions for common view patterns.

### View Helpers

```go
// Center text horizontally
func CenterText(text string, width int) string {
    if len(text) >= width {
        return text
    }
    padding := (width - len(text)) / 2
    return strings.Repeat(" ", padding) + text
}

// Truncate text with ellipsis
func TruncateText(text string, maxLen int) string {
    if len(text) <= maxLen {
        return text
    }
    return text[:maxLen-3] + "..."
}

// Word wrap text
func WrapText(text string, width int) []string {
    words := strings.Fields(text)
    var lines []string
    var currentLine string

    for _, word := range words {
        if len(currentLine)+len(word)+1 > width {
            lines = append(lines, currentLine)
            currentLine = word
        } else {
            if currentLine != "" {
                currentLine += " "
            }
            currentLine += word
        }
    }

    if currentLine != "" {
        lines = append(lines, currentLine)
    }

    return lines
}

// Pad text to width
func PadText(text string, width int, align string) string {
    if len(text) >= width {
        return text[:width]
    }

    padding := width - len(text)
    switch align {
    case "left":
        return text + strings.Repeat(" ", padding)
    case "right":
        return strings.Repeat(" ", padding) + text
    case "center":
        leftPad := padding / 2
        rightPad := padding - leftPad
        return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
    default:
        return text
    }
}

// Format currency
func FormatCredits(amount int) string {
    return fmt.Sprintf("%s cr", formatNumber(amount))
}

// Format large numbers with commas
func formatNumber(n int) string {
    s := fmt.Sprintf("%d", n)
    var result string

    for i, c := range s {
        if i > 0 && (len(s)-i)%3 == 0 {
            result += ","
        }
        result += string(c)
    }

    return result
}

// Format time duration
func FormatDuration(d time.Duration) string {
    if d < time.Minute {
        return fmt.Sprintf("%ds", int(d.Seconds()))
    }
    if d < time.Hour {
        return fmt.Sprintf("%dm", int(d.Minutes()))
    }
    return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}
```

---

## Utils Component

### Source File
`internal/tui/utils.go`

### Purpose
General utility functions for TUI operations.

### Utilities

```go
// Clamp value between min and max
func Clamp(value, min, max int) int {
    if value < min {
        return min
    }
    if value > max {
        return max
    }
    return value
}

// Calculate percentage
func Percentage(current, total int) int {
    if total == 0 {
        return 0
    }
    return (current * 100) / total
}

// Interpolate between two values
func Lerp(start, end, t float64) float64 {
    return start + t*(end-start)
}

// Color text (terminal colors)
func ColorText(text string, color string) string {
    colors := map[string]string{
        "red":    "\033[31m",
        "green":  "\033[32m",
        "yellow": "\033[33m",
        "blue":   "\033[34m",
        "reset":  "\033[0m",
    }

    if c, ok := colors[color]; ok {
        return c + text + colors["reset"]
    }
    return text
}

// Get terminal size
func GetTerminalSize() (width, height int, err error) {
    width, height, err = term.GetSize(int(os.Stdout.Fd()))
    return
}
```

---

## Component Design Patterns

### Composition Over Inheritance

Components are designed to be composed:

```go
type Screen struct {
    header    *HeaderComponent
    content   *ContentComponent
    footer    *FooterComponent
    sidebar   *SidebarComponent
}

func (s *Screen) View() string {
    var b strings.Builder

    b.WriteString(s.header.View())
    b.WriteString(s.content.View())
    b.WriteString(s.footer.View())

    return b.String()
}
```

### State Management

Components maintain minimal state:

```go
type Component struct {
    // State
    selected int
    items    []string

    // Styling
    width  int
    height int
}

func (c *Component) Update(msg tea.Msg) tea.Cmd {
    // Handle updates
    return nil
}

func (c *Component) View() string {
    // Render component
    return ""
}
```

### Theming

Components support theming through lipgloss:

```go
import "github.com/charmbracelet/lipgloss"

var (
    titleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#00FF00"))

    selectedStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("#444444")).
        Foreground(lipgloss.Color("#FFFFFF"))

    errorStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FF0000"))
)
```

---

## Best Practices

### Component Development

1. **Single Responsibility**: Each component does one thing well
2. **Reusability**: Design for multiple contexts
3. **Configurability**: Accept options for customization
4. **Consistency**: Follow established patterns
5. **Documentation**: Document props and usage

### Performance

1. **Lazy Rendering**: Only render visible content
2. **Memoization**: Cache expensive calculations
3. **Efficient Updates**: Minimize re-renders
4. **String Building**: Use `strings.Builder` for concatenation

### Testing

```go
func TestProgressBar(t *testing.T) {
    bar := ProgressBar(50, 100, 10)
    expected := "[█████░░░░░] 50%"
    if bar != expected {
        t.Errorf("Expected %s, got %s", expected, bar)
    }
}

func TestCenterText(t *testing.T) {
    centered := CenterText("Hello", 11)
    expected := "   Hello   "
    if centered != expected {
        t.Errorf("Expected '%s', got '%s'", expected, centered)
    }
}
```

---

## Implementation Notes

### BubbleTea Integration

Components follow BubbleTea's Model-View-Update pattern:

```go
type Component interface {
    Update(msg tea.Msg) (Component, tea.Cmd)
    View() string
}
```

### Lipgloss Styling

Use Lipgloss for advanced styling:

```go
import "github.com/charmbracelet/lipgloss"

box := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("#00FF00")).
    Padding(1, 2).
    Width(50).
    Render("Content")
```

### Responsive Design

Components adapt to terminal size:

```go
func (c *Component) Resize(width, height int) {
    c.width = width
    c.height = height
    c.recalculateLayout()
}
```

---

**Last Updated**: 2025-11-15
**Document Version**: 1.0.0

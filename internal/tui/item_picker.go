// File: internal/tui/item_picker.go
// Project: Terminal Velocity
// Description: Reusable item picker component for selecting items from inventory
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-11-15

package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

// ItemPickerMode determines how the picker behaves
type ItemPickerMode int

const (
	ItemPickerModeSingle ItemPickerMode = iota // Select single item
	ItemPickerModeMulti                        // Select multiple items (checkbox)
)

// ItemPickerFilter determines which items to show
type ItemPickerFilter int

const (
	FilterAll ItemPickerFilter = iota
	FilterWeapons
	FilterOutfits
	FilterSpecial
	FilterQuest
	FilterAvailable // Only ship/station (excludes mail/escrow/auction)
	FilterShip      // Only items on ship
	FilterStation   // Only items in station storage
)

// ItemPickerModel is a reusable component for picking items from inventory
type ItemPickerModel struct {
	// Configuration
	mode         ItemPickerMode
	filter       ItemPickerFilter
	title        string
	maxSelection int // 0 = unlimited (for multi mode)

	// Data
	itemRepo *database.ItemRepository
	playerID uuid.UUID
	items    []*models.PlayerItem
	allItems []*models.PlayerItem // Unfiltered items for search

	// State
	cursor       int
	selected     map[uuid.UUID]bool // For multi-select mode
	searchQuery  string
	searchMode   bool
	error        string
	loading      bool

	// Scroll support
	scrollOffset int
	viewportHeight int // Number of items visible at once
}

// NewItemPicker creates a new item picker component
func NewItemPicker(repo *database.ItemRepository, playerID uuid.UUID) *ItemPickerModel {
	return &ItemPickerModel{
		mode:           ItemPickerModeMulti,
		filter:         FilterAvailable,
		title:          "Select Items",
		maxSelection:   0,
		itemRepo:       repo,
		playerID:       playerID,
		cursor:         0,
		selected:       make(map[uuid.UUID]bool),
		searchQuery:    "",
		searchMode:     false,
		loading:        false,
		scrollOffset:   0,
		viewportHeight: 10, // Show 10 items at a time
	}
}

// SetMode sets the picker mode (single or multi select)
func (p *ItemPickerModel) SetMode(mode ItemPickerMode) {
	p.mode = mode
	if mode == ItemPickerModeSingle {
		p.selected = make(map[uuid.UUID]bool)
	}
}

// SetFilter sets the item filter
func (p *ItemPickerModel) SetFilter(filter ItemPickerFilter) {
	p.filter = filter
}

// SetTitle sets the picker title
func (p *ItemPickerModel) SetTitle(title string) {
	p.title = title
}

// SetMaxSelection sets max items selectable (0 = unlimited)
func (p *ItemPickerModel) SetMaxSelection(max int) {
	p.maxSelection = max
}

// LoadItems loads items from the repository
func (p *ItemPickerModel) LoadItems() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		var items []*models.PlayerItem
		var err error

		switch p.filter {
		case FilterAll:
			items, err = p.itemRepo.GetPlayerItems(ctx, p.playerID)
		case FilterAvailable:
			items, err = p.itemRepo.GetAvailableItems(ctx, p.playerID)
		case FilterWeapons:
			items, err = p.itemRepo.GetItemsByType(ctx, p.playerID, models.ItemTypeWeapon)
		case FilterOutfits:
			items, err = p.itemRepo.GetItemsByType(ctx, p.playerID, models.ItemTypeOutfit)
		case FilterSpecial:
			items, err = p.itemRepo.GetItemsByType(ctx, p.playerID, models.ItemTypeSpecial)
		case FilterQuest:
			items, err = p.itemRepo.GetItemsByType(ctx, p.playerID, models.ItemTypeQuest)
		case FilterShip:
			// Filter available items to only ship location
			allItems, err := p.itemRepo.GetAvailableItems(ctx, p.playerID)
			if err == nil {
				for _, item := range allItems {
					if item.Location == models.LocationShip {
						items = append(items, item)
					}
				}
			}
		case FilterStation:
			// Filter available items to only station location
			allItems, err := p.itemRepo.GetAvailableItems(ctx, p.playerID)
			if err == nil {
				for _, item := range allItems {
					if item.Location == models.LocationStationStorage {
						items = append(items, item)
					}
				}
			}
		default:
			items, err = p.itemRepo.GetAvailableItems(ctx, p.playerID)
		}

		return itemsLoadedMsg{items: items, err: err}
	}
}

type itemsLoadedMsg struct {
	items []*models.PlayerItem
	err   error
}

// Update handles input for the item picker
func (p *ItemPickerModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case itemsLoadedMsg:
		if msg.err != nil {
			p.error = msg.err.Error()
			p.loading = false
			return nil
		}
		p.items = msg.items
		p.allItems = msg.items
		p.loading = false
		p.applySearch() // Apply any existing search
		return nil

	case tea.KeyMsg:
		if p.searchMode {
			return p.handleSearchInput(msg)
		}
		return p.handleNormalInput(msg)
	}

	return nil
}

// handleNormalInput handles keyboard input in normal mode
func (p *ItemPickerModel) handleNormalInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "up", "k":
		if p.cursor > 0 {
			p.cursor--
			p.adjustScroll()
		}

	case "down", "j":
		if p.cursor < len(p.items)-1 {
			p.cursor++
			p.adjustScroll()
		}

	case "g": // Go to top
		p.cursor = 0
		p.scrollOffset = 0

	case "G": // Go to bottom
		if len(p.items) > 0 {
			p.cursor = len(p.items) - 1
			p.adjustScroll()
		}

	case "pgup":
		p.cursor -= p.viewportHeight
		if p.cursor < 0 {
			p.cursor = 0
		}
		p.adjustScroll()

	case "pgdown":
		p.cursor += p.viewportHeight
		if p.cursor >= len(p.items) {
			p.cursor = len(p.items) - 1
		}
		p.adjustScroll()

	case " ", "enter":
		if len(p.items) == 0 {
			return nil
		}

		currentItem := p.items[p.cursor]

		if p.mode == ItemPickerModeSingle {
			// Single select: toggle current item only
			p.selected = make(map[uuid.UUID]bool)
			p.selected[currentItem.ID] = true
		} else {
			// Multi select: toggle current item
			if p.selected[currentItem.ID] {
				delete(p.selected, currentItem.ID)
			} else {
				// Check max selection limit
				if p.maxSelection > 0 && len(p.selected) >= p.maxSelection {
					p.error = fmt.Sprintf("Maximum %d items allowed", p.maxSelection)
					return nil
				}
				p.selected[currentItem.ID] = true
				p.error = "" // Clear error
			}
		}

	case "/":
		// Enter search mode
		p.searchMode = true
		p.searchQuery = ""
		p.error = ""

	case "ctrl+a":
		// Select all (multi-select only)
		if p.mode == ItemPickerModeMulti {
			if p.maxSelection == 0 || len(p.items) <= p.maxSelection {
				for _, item := range p.items {
					p.selected[item.ID] = true
				}
			} else {
				p.error = fmt.Sprintf("Cannot select all: max %d items allowed", p.maxSelection)
			}
		}

	case "ctrl+d":
		// Deselect all
		p.selected = make(map[uuid.UUID]bool)
		p.error = ""
	}

	return nil
}

// handleSearchInput handles keyboard input in search mode
func (p *ItemPickerModel) handleSearchInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		// Exit search mode
		p.searchMode = false
		p.searchQuery = ""
		p.items = p.allItems
		p.cursor = 0
		p.scrollOffset = 0

	case "enter":
		// Apply search and exit search mode
		p.searchMode = false
		p.applySearch()
		p.cursor = 0
		p.scrollOffset = 0

	case "backspace":
		if len(p.searchQuery) > 0 {
			p.searchQuery = p.searchQuery[:len(p.searchQuery)-1]
		}

	case "ctrl+u":
		// Clear search query
		p.searchQuery = ""

	default:
		// Add character to search query
		if len(msg.String()) == 1 {
			p.searchQuery += msg.String()
		}
	}

	return nil
}

// applySearch filters items based on search query
func (p *ItemPickerModel) applySearch() {
	if p.searchQuery == "" {
		p.items = p.allItems
		return
	}

	query := strings.ToLower(p.searchQuery)
	var filtered []*models.PlayerItem

	for _, item := range p.allItems {
		itemName := strings.ToLower(item.GetEquipmentName())
		equipID := strings.ToLower(item.EquipmentID)

		if strings.Contains(itemName, query) || strings.Contains(equipID, query) {
			filtered = append(filtered, item)
		}
	}

	p.items = filtered
}

// adjustScroll adjusts scroll offset to keep cursor in view
func (p *ItemPickerModel) adjustScroll() {
	// Scroll down if cursor is below visible area
	if p.cursor >= p.scrollOffset+p.viewportHeight {
		p.scrollOffset = p.cursor - p.viewportHeight + 1
	}

	// Scroll up if cursor is above visible area
	if p.cursor < p.scrollOffset {
		p.scrollOffset = p.cursor
	}
}

// View renders the item picker
func (p *ItemPickerModel) View() string {
	var s strings.Builder

	// Title
	s.WriteString(subtitleStyle.Render(fmt.Sprintf("=== %s ===", p.title)) + "\n\n")

	// Error display
	if p.error != "" {
		s.WriteString(errorStyle.Render(p.error) + "\n\n")
	}

	// Loading state
	if p.loading {
		s.WriteString(helpStyle.Render("Loading items...") + "\n")
		return s.String()
	}

	// Filter info
	filterName := p.getFilterName()
	modeName := "Single"
	if p.mode == ItemPickerModeMulti {
		modeName = "Multi"
	}
	s.WriteString(fmt.Sprintf("Filter: %s  |  Mode: %s  |  Items: %d",
		highlightStyle.Render(filterName),
		highlightStyle.Render(modeName),
		len(p.items)) + "\n")

	// Selection count
	if len(p.selected) > 0 {
		s.WriteString(fmt.Sprintf("Selected: %s",
			statsStyle.Render(fmt.Sprintf("%d", len(p.selected)))) + "\n")
	}

	// Search bar
	if p.searchMode {
		s.WriteString(fmt.Sprintf("\nSearch: %s_", highlightStyle.Render(p.searchQuery)) + "\n")
	} else if p.searchQuery != "" {
		s.WriteString(fmt.Sprintf("\nSearch: %s (press / to edit)", helpStyle.Render(p.searchQuery)) + "\n")
	}

	s.WriteString("\n")

	// Items list
	if len(p.items) == 0 {
		s.WriteString(helpStyle.Render("No items found") + "\n\n")
	} else {
		s.WriteString(p.renderItemsList())
	}

	// Help footer
	if p.searchMode {
		s.WriteString(renderFooter("Type to search  •  Enter: Apply  •  ESC: Cancel"))
	} else {
		if p.mode == ItemPickerModeMulti {
			s.WriteString(renderFooter("↑/↓: Navigate  •  Space: Toggle  •  /: Search  •  Ctrl+A: All  •  Ctrl+D: None"))
		} else {
			s.WriteString(renderFooter("↑/↓: Navigate  •  Enter: Select  •  /: Search"))
		}
	}

	return s.String()
}

// renderItemsList renders the scrollable list of items
func (p *ItemPickerModel) renderItemsList() string {
	var s strings.Builder

	// Calculate visible range
	start := p.scrollOffset
	end := p.scrollOffset + p.viewportHeight
	if end > len(p.items) {
		end = len(p.items)
	}

	// Render visible items
	for i := start; i < end; i++ {
		item := p.items[i]
		s.WriteString(p.renderItem(i, item))
	}

	// Scroll indicators
	if p.scrollOffset > 0 {
		s.WriteString(helpStyle.Render("  ↑ More items above\n"))
	}
	if end < len(p.items) {
		s.WriteString(helpStyle.Render("  ↓ More items below\n"))
	}

	return s.String()
}

// renderItem renders a single item in the list
func (p *ItemPickerModel) renderItem(index int, item *models.PlayerItem) string {
	cursor := "  "
	if index == p.cursor {
		cursor = "> "
	}

	checkbox := "[ ]"
	if p.selected[item.ID] {
		checkbox = "[✓]"
	}

	itemName := item.GetDisplayName()
	location := item.GetLocationName()

	// Apply styling based on cursor position
	if index == p.cursor {
		itemName = selectedMenuItemStyle.Render(itemName)
		checkbox = selectedMenuItemStyle.Render(checkbox)
	} else if p.selected[item.ID] {
		itemName = highlightStyle.Render(itemName)
		checkbox = highlightStyle.Render(checkbox)
	}

	if p.mode == ItemPickerModeMulti {
		return fmt.Sprintf("%s%s %s  (%s)\n", cursor, checkbox, itemName, location)
	}

	// Single select mode - no checkbox
	return fmt.Sprintf("%s%s  (%s)\n", cursor, itemName, location)
}

// getFilterName returns a human-readable filter name
func (p *ItemPickerModel) getFilterName() string {
	switch p.filter {
	case FilterAll:
		return "All Items"
	case FilterWeapons:
		return "Weapons"
	case FilterOutfits:
		return "Outfits"
	case FilterSpecial:
		return "Special"
	case FilterQuest:
		return "Quest Items"
	case FilterAvailable:
		return "Available"
	case FilterShip:
		return "Ship"
	case FilterStation:
		return "Station Storage"
	default:
		return "Unknown"
	}
}

// GetSelectedItems returns the currently selected item IDs
func (p *ItemPickerModel) GetSelectedItems() []uuid.UUID {
	var selected []uuid.UUID
	for id := range p.selected {
		selected = append(selected, id)
	}
	return selected
}

// GetSelectedCount returns the number of selected items
func (p *ItemPickerModel) GetSelectedCount() int {
	return len(p.selected)
}

// ClearSelection clears all selected items
func (p *ItemPickerModel) ClearSelection() {
	p.selected = make(map[uuid.UUID]bool)
	p.error = ""
}

// Reset resets the picker to initial state
func (p *ItemPickerModel) Reset() {
	p.cursor = 0
	p.selected = make(map[uuid.UUID]bool)
	p.searchQuery = ""
	p.searchMode = false
	p.error = ""
	p.scrollOffset = 0
	p.items = nil
	p.allItems = nil
}

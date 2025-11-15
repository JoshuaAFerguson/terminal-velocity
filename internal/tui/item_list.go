// File: internal/tui/item_list.go
// Project: Terminal Velocity
// Description: Read-only item list component for displaying inventory
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-11-15

package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

// ItemListGrouping determines how items are grouped
type ItemListGrouping int

const (
	GroupByNone ItemListGrouping = iota
	GroupByType                  // Group by weapon/outfit/special/quest
	GroupByLocation              // Group by ship/station/mail/etc
)

// ItemListSorting determines how items are sorted
type ItemListSorting int

const (
	SortByName ItemListSorting = iota
	SortByType
	SortByAcquiredDate
)

// ItemListModel is a read-only component for displaying items
type ItemListModel struct {
	// Configuration
	title     string
	grouping  ItemListGrouping
	sorting   ItemListSorting
	showStats bool // Show item stats/properties inline

	// Data
	itemRepo *database.ItemRepository
	playerID uuid.UUID
	items    []*models.PlayerItem

	// State
	cursor       int
	error        string
	loading      bool
	scrollOffset int
	viewportHeight int

	// Filtering (read-only but can filter what's shown)
	filter ItemPickerFilter
}

// NewItemList creates a new item list component
func NewItemList(repo *database.ItemRepository, playerID uuid.UUID) *ItemListModel {
	return &ItemListModel{
		title:          "Your Items",
		grouping:       GroupByType,
		sorting:        SortByName,
		showStats:      false,
		itemRepo:       repo,
		playerID:       playerID,
		cursor:         0,
		loading:        false,
		scrollOffset:   0,
		viewportHeight: 15, // Show 15 items at a time
		filter:         FilterAll,
	}
}

// SetTitle sets the list title
func (l *ItemListModel) SetTitle(title string) {
	l.title = title
}

// SetGrouping sets how items should be grouped
func (l *ItemListModel) SetGrouping(grouping ItemListGrouping) {
	l.grouping = grouping
}

// SetSorting sets how items should be sorted
func (l *ItemListModel) SetSorting(sorting ItemListSorting) {
	l.sorting = sorting
}

// SetShowStats enables/disables inline stats display
func (l *ItemListModel) SetShowStats(show bool) {
	l.showStats = show
}

// SetFilter sets the item filter
func (l *ItemListModel) SetFilter(filter ItemPickerFilter) {
	l.filter = filter
}

// LoadItems loads items from the repository
func (l *ItemListModel) LoadItems() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		var items []*models.PlayerItem
		var err error

		switch l.filter {
		case FilterAll:
			items, err = l.itemRepo.GetPlayerItems(ctx, l.playerID)
		case FilterAvailable:
			items, err = l.itemRepo.GetAvailableItems(ctx, l.playerID)
		case FilterWeapons:
			items, err = l.itemRepo.GetItemsByType(ctx, l.playerID, models.ItemTypeWeapon)
		case FilterOutfits:
			items, err = l.itemRepo.GetItemsByType(ctx, l.playerID, models.ItemTypeOutfit)
		case FilterSpecial:
			items, err = l.itemRepo.GetItemsByType(ctx, l.playerID, models.ItemTypeSpecial)
		case FilterQuest:
			items, err = l.itemRepo.GetItemsByType(ctx, l.playerID, models.ItemTypeQuest)
		case FilterShip:
			allItems, err := l.itemRepo.GetAvailableItems(ctx, l.playerID)
			if err == nil {
				for _, item := range allItems {
					if item.Location == models.LocationShip {
						items = append(items, item)
					}
				}
			}
		case FilterStation:
			allItems, err := l.itemRepo.GetAvailableItems(ctx, l.playerID)
			if err == nil {
				for _, item := range allItems {
					if item.Location == models.LocationStationStorage {
						items = append(items, item)
					}
				}
			}
		default:
			items, err = l.itemRepo.GetPlayerItems(ctx, l.playerID)
		}

		return itemsLoadedMsg{items: items, err: err}
	}
}

// Update handles input for the item list
func (l *ItemListModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case itemsLoadedMsg:
		if msg.err != nil {
			l.error = msg.err.Error()
			l.loading = false
			return nil
		}
		l.items = msg.items
		l.sortItems()
		l.loading = false
		return nil

	case tea.KeyMsg:
		return l.handleInput(msg)
	}

	return nil
}

// handleInput handles keyboard input
func (l *ItemListModel) handleInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "up", "k":
		if l.cursor > 0 {
			l.cursor--
			l.adjustScroll()
		}

	case "down", "j":
		if l.cursor < len(l.items)-1 {
			l.cursor++
			l.adjustScroll()
		}

	case "g": // Go to top
		l.cursor = 0
		l.scrollOffset = 0

	case "G": // Go to bottom
		if len(l.items) > 0 {
			l.cursor = len(l.items) - 1
			l.adjustScroll()
		}

	case "pgup":
		l.cursor -= l.viewportHeight
		if l.cursor < 0 {
			l.cursor = 0
		}
		l.adjustScroll()

	case "pgdown":
		l.cursor += l.viewportHeight
		if l.cursor >= len(l.items) {
			l.cursor = len(l.items) - 1
		}
		l.adjustScroll()

	case "s":
		// Cycle through sorting options
		switch l.sorting {
		case SortByName:
			l.sorting = SortByType
		case SortByType:
			l.sorting = SortByAcquiredDate
		case SortByAcquiredDate:
			l.sorting = SortByName
		}
		l.sortItems()
		l.cursor = 0
		l.scrollOffset = 0

	case "t":
		// Toggle grouping
		switch l.grouping {
		case GroupByNone:
			l.grouping = GroupByType
		case GroupByType:
			l.grouping = GroupByLocation
		case GroupByLocation:
			l.grouping = GroupByNone
		}
	}

	return nil
}

// sortItems sorts the items based on current sorting setting
func (l *ItemListModel) sortItems() {
	if l.items == nil {
		return
	}

	switch l.sorting {
	case SortByName:
		sort.Slice(l.items, func(i, j int) bool {
			return l.items[i].GetEquipmentName() < l.items[j].GetEquipmentName()
		})
	case SortByType:
		sort.Slice(l.items, func(i, j int) bool {
			if l.items[i].ItemType == l.items[j].ItemType {
				return l.items[i].GetEquipmentName() < l.items[j].GetEquipmentName()
			}
			return l.items[i].ItemType < l.items[j].ItemType
		})
	case SortByAcquiredDate:
		sort.Slice(l.items, func(i, j int) bool {
			return l.items[i].AcquiredAt.After(l.items[j].AcquiredAt)
		})
	}
}

// adjustScroll adjusts scroll offset to keep cursor in view
func (l *ItemListModel) adjustScroll() {
	if l.cursor >= l.scrollOffset+l.viewportHeight {
		l.scrollOffset = l.cursor - l.viewportHeight + 1
	}

	if l.cursor < l.scrollOffset {
		l.scrollOffset = l.cursor
	}
}

// View renders the item list
func (l *ItemListModel) View() string {
	var s strings.Builder

	// Title
	s.WriteString(subtitleStyle.Render(fmt.Sprintf("=== %s ===", l.title)) + "\n\n")

	// Error display
	if l.error != "" {
		s.WriteString(errorStyle.Render(l.error) + "\n\n")
	}

	// Loading state
	if l.loading {
		s.WriteString(helpStyle.Render("Loading items...") + "\n")
		return s.String()
	}

	// Summary
	totalItems := len(l.items)
	s.WriteString(fmt.Sprintf("Total Items: %s  |  Sort: %s  |  Group: %s\n\n",
		statsStyle.Render(fmt.Sprintf("%d", totalItems)),
		highlightStyle.Render(l.getSortingName()),
		highlightStyle.Render(l.getGroupingName())) + "\n")

	// Items list
	if len(l.items) == 0 {
		s.WriteString(helpStyle.Render("No items in inventory") + "\n\n")
	} else {
		if l.grouping == GroupByNone {
			s.WriteString(l.renderFlatList())
		} else {
			s.WriteString(l.renderGroupedList())
		}
	}

	// Help footer
	s.WriteString(renderFooter("↑/↓: Navigate  •  s: Sort  •  t: Toggle Group"))

	return s.String()
}

// renderFlatList renders a simple flat list
func (l *ItemListModel) renderFlatList() string {
	var s strings.Builder

	start := l.scrollOffset
	end := l.scrollOffset + l.viewportHeight
	if end > len(l.items) {
		end = len(l.items)
	}

	for i := start; i < end; i++ {
		item := l.items[i]
		s.WriteString(l.renderItem(i, item))
	}

	// Scroll indicators
	if l.scrollOffset > 0 {
		s.WriteString(helpStyle.Render("  ↑ More items above\n"))
	}
	if end < len(l.items) {
		s.WriteString(helpStyle.Render("  ↓ More items below\n"))
	}

	return s.String()
}

// renderGroupedList renders items grouped by type or location
func (l *ItemListModel) renderGroupedList() string {
	var s strings.Builder

	groups := l.groupItems()

	for _, groupName := range l.getSortedGroupNames(groups) {
		items := groups[groupName]
		if len(items) == 0 {
			continue
		}

		// Group header
		s.WriteString(highlightStyle.Render(fmt.Sprintf("\n%s (%d)", groupName, len(items))) + "\n")

		// Group items
		for _, item := range items {
			s.WriteString(l.renderItemSimple(item))
		}
	}

	return s.String()
}

// groupItems groups items based on current grouping setting
func (l *ItemListModel) groupItems() map[string][]*models.PlayerItem {
	groups := make(map[string][]*models.PlayerItem)

	for _, item := range l.items {
		var groupName string

		switch l.grouping {
		case GroupByType:
			switch item.ItemType {
			case models.ItemTypeWeapon:
				groupName = "Weapons"
			case models.ItemTypeOutfit:
				groupName = "Outfits"
			case models.ItemTypeSpecial:
				groupName = "Special Items"
			case models.ItemTypeQuest:
				groupName = "Quest Items"
			default:
				groupName = "Other"
			}
		case GroupByLocation:
			groupName = item.GetLocationName()
		default:
			groupName = "All"
		}

		groups[groupName] = append(groups[groupName], item)
	}

	return groups
}

// getSortedGroupNames returns group names in sorted order
func (l *ItemListModel) getSortedGroupNames(groups map[string][]*models.PlayerItem) []string {
	names := make([]string, 0, len(groups))
	for name := range groups {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// renderItem renders a single item with cursor
func (l *ItemListModel) renderItem(index int, item *models.PlayerItem) string {
	cursor := "  "
	if index == l.cursor {
		cursor = "> "
	}

	itemName := item.GetDisplayName()
	location := item.GetLocationName()

	if index == l.cursor {
		itemName = selectedMenuItemStyle.Render(itemName)
	}

	line := fmt.Sprintf("%s%s  (%s)", cursor, itemName, location)

	if l.showStats {
		props, err := item.GetProperties()
		if err == nil && (len(props.Mods) > 0 || len(props.Upgrades) > 0) {
			line += "  " + helpStyle.Render(l.formatProperties(props))
		}
	}

	return line + "\n"
}

// renderItemSimple renders an item without cursor (for grouped view)
func (l *ItemListModel) renderItemSimple(item *models.PlayerItem) string {
	itemName := item.GetEquipmentName()
	location := item.GetLocationName()

	line := fmt.Sprintf("  • %s  (%s)", itemName, location)

	if l.showStats {
		props, err := item.GetProperties()
		if err == nil && (len(props.Mods) > 0 || len(props.Upgrades) > 0) {
			line += "  " + helpStyle.Render(l.formatProperties(props))
		}
	}

	return line + "\n"
}

// formatProperties formats item properties for display
func (l *ItemListModel) formatProperties(props *models.ItemProperties) string {
	var parts []string

	if len(props.Mods) > 0 {
		parts = append(parts, fmt.Sprintf("Mods: %s", strings.Join(props.Mods, ", ")))
	}

	if len(props.Upgrades) > 0 {
		var upgrades []string
		for name, level := range props.Upgrades {
			upgrades = append(upgrades, fmt.Sprintf("%s+%d", name, level))
		}
		parts = append(parts, strings.Join(upgrades, ", "))
	}

	return strings.Join(parts, " | ")
}

// getSortingName returns a human-readable sorting name
func (l *ItemListModel) getSortingName() string {
	switch l.sorting {
	case SortByName:
		return "Name"
	case SortByType:
		return "Type"
	case SortByAcquiredDate:
		return "Acquired"
	default:
		return "Unknown"
	}
}

// getGroupingName returns a human-readable grouping name
func (l *ItemListModel) getGroupingName() string {
	switch l.grouping {
	case GroupByNone:
		return "None"
	case GroupByType:
		return "Type"
	case GroupByLocation:
		return "Location"
	default:
		return "Unknown"
	}
}

// GetCurrentItem returns the currently selected item (under cursor)
func (l *ItemListModel) GetCurrentItem() *models.PlayerItem {
	if l.cursor < 0 || l.cursor >= len(l.items) {
		return nil
	}
	return l.items[l.cursor]
}

// GetItemCount returns the total number of items
func (l *ItemListModel) GetItemCount() int {
	return len(l.items)
}

// Reset resets the list to initial state
func (l *ItemListModel) Reset() {
	l.cursor = 0
	l.error = ""
	l.scrollOffset = 0
	l.items = nil
}

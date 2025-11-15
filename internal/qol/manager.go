// File: internal/qol/manager.go
// Project: Terminal Velocity
// Description: Quality of life features including waypoints, auto-trading, and polish
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package qol

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/google/uuid"
)

var log = logger.WithComponent("QoL")

// Manager handles quality of life features
type Manager struct {
	mu sync.RWMutex

	// QoL data
	waypoints      map[uuid.UUID]*WaypointSet     // player_id -> waypoints
	autoTraders    map[uuid.UUID]*AutoTrader      // player_id -> auto-trader
	tradeRoutes    map[uuid.UUID]*TradeRoute      // route_id -> trade route
	quickCommands  map[uuid.UUID]map[string]string // player_id -> command shortcuts
	notifications  map[uuid.UUID]*NotificationPrefs // player_id -> preferences

	// Configuration
	config QoLConfig

	// Repositories
	playerRepo *database.PlayerRepository
	systemRepo *database.SystemRepository

	// Callbacks
	onWaypointReached func(playerID uuid.UUID, waypoint *Waypoint)
	onAutoTradeComplete func(playerID uuid.UUID, profit int64)
	onNotification func(playerID uuid.UUID, notification *Notification)

	// Background workers
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// QoLConfig defines QoL system parameters
type QoLConfig struct {
	// Waypoint settings
	MaxWaypoints         int           // Max waypoints per player
	WaypointAutoNav      bool          // Enable auto-navigation
	WaypointNotifyRange  float64       // Distance to notify approaching waypoint

	// Auto-trading settings
	AutoTradeEnabled     bool          // Enable auto-trading
	AutoTradeInterval    time.Duration // How often to check trade routes
	AutoTradeFuelReserve float64       // Minimum fuel to keep
	AutoTradeMaxJumps    int           // Max jumps for route

	// Notification settings
	EnableNotifications  bool          // Global notification toggle
	NotificationSound    bool          // Sound notifications
	NotificationHistory  int           // How many to keep

	// Quick command settings
	MaxQuickCommands     int           // Max shortcuts per player
	CommandAliasLimit    int           // Max characters in alias

	// Auto-save settings
	AutoSaveInterval     time.Duration // How often to auto-save
	AutoSaveOnJump       bool          // Save before each jump
	AutoSaveOnTrade      bool          // Save after trades
}

// DefaultQoLConfig returns sensible defaults
func DefaultQoLConfig() QoLConfig {
	return QoLConfig{
		MaxWaypoints:         20,
		WaypointAutoNav:      true,
		WaypointNotifyRange:  5.0,
		AutoTradeEnabled:     true,
		AutoTradeInterval:    5 * time.Minute,
		AutoTradeFuelReserve: 50.0,
		AutoTradeMaxJumps:    5,
		EnableNotifications:  true,
		NotificationSound:    true,
		NotificationHistory:  50,
		MaxQuickCommands:     10,
		CommandAliasLimit:    20,
		AutoSaveInterval:     5 * time.Minute,
		AutoSaveOnJump:       true,
		AutoSaveOnTrade:      true,
	}
}

// NewManager creates a new QoL manager
func NewManager(playerRepo *database.PlayerRepository, systemRepo *database.SystemRepository) *Manager {
	return &Manager{
		waypoints:     make(map[uuid.UUID]*WaypointSet),
		autoTraders:   make(map[uuid.UUID]*AutoTrader),
		tradeRoutes:   make(map[uuid.UUID]*TradeRoute),
		quickCommands: make(map[uuid.UUID]map[string]string),
		notifications: make(map[uuid.UUID]*NotificationPrefs),
		config:        DefaultQoLConfig(),
		playerRepo:    playerRepo,
		systemRepo:    systemRepo,
		stopChan:      make(chan struct{}),
	}
}

// Start begins background workers
func (m *Manager) Start() {
	m.wg.Add(1)
	go m.autoTradeWorker()
	log.Info("QoL manager started")
}

// Stop gracefully shuts down the manager
func (m *Manager) Stop() {
	close(m.stopChan)
	m.wg.Wait()
	log.Info("QoL manager stopped")
}

// SetCallbacks sets all QoL callbacks
func (m *Manager) SetCallbacks(
	onWaypointReached func(playerID uuid.UUID, waypoint *Waypoint),
	onAutoTradeComplete func(playerID uuid.UUID, profit int64),
	onNotification func(playerID uuid.UUID, notification *Notification),
) {
	m.onWaypointReached = onWaypointReached
	m.onAutoTradeComplete = onAutoTradeComplete
	m.onNotification = onNotification
}

// ============================================================================
// DATA STRUCTURES
// ============================================================================

// WaypointSet represents a player's waypoints
type WaypointSet struct {
	PlayerID  uuid.UUID
	Waypoints []*Waypoint
	Active    *uuid.UUID // Currently active waypoint
}

// Waypoint represents a navigation marker
type Waypoint struct {
	ID          uuid.UUID
	Name        string
	Description string
	SystemID    uuid.UUID
	SystemName  string
	PlanetID    *uuid.UUID // Optional: specific planet
	PlanetName  string
	Color       string     // For display
	Icon        string     // Icon character
	CreatedAt   time.Time
	Visited     bool
}

// AutoTrader represents an automated trading configuration
type AutoTrader struct {
	PlayerID     uuid.UUID
	Enabled      bool
	Routes       []uuid.UUID // Trade route IDs
	MinProfit    int64       // Minimum profit to execute
	LastRun      time.Time
	TotalProfit  int64
	TradesExecuted int
}

// TradeRoute represents a profitable trade route
type TradeRoute struct {
	ID            uuid.UUID
	PlayerID      uuid.UUID
	Name          string
	FromSystemID  uuid.UUID
	ToSystemID    uuid.UUID
	FromSystem    string
	ToSystem      string
	Commodity     string
	BuyPrice      int64
	SellPrice     int64
	Profit        int64
	Distance      float64
	LastUpdated   time.Time
	Active        bool
}

// NotificationPrefs represents player notification preferences
type NotificationPrefs struct {
	PlayerID       uuid.UUID
	EnabledTypes   map[string]bool // notification_type -> enabled
	History        []*Notification
	MaxHistory     int
}

// Notification represents a player notification
type Notification struct {
	ID        uuid.UUID
	Type      string    // "waypoint", "trade", "combat", "system"
	Title     string
	Message   string
	Priority  string    // "low", "normal", "high", "critical"
	Timestamp time.Time
	Read      bool
}

// QuickCommand represents a command shortcut
type QuickCommand struct {
	Alias   string
	Command string
	Args    []string
}

// ============================================================================
// WAYPOINT SYSTEM
// ============================================================================

// AddWaypoint adds a new waypoint for a player
func (m *Manager) AddWaypoint(ctx context.Context, playerID uuid.UUID, name, description string, systemID uuid.UUID, systemName string, planetID *uuid.UUID, planetName string) (*Waypoint, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get or create waypoint set
	waypointSet := m.waypoints[playerID]
	if waypointSet == nil {
		waypointSet = &WaypointSet{
			PlayerID:  playerID,
			Waypoints: []*Waypoint{},
		}
		m.waypoints[playerID] = waypointSet
	}

	// Check limit
	if len(waypointSet.Waypoints) >= m.config.MaxWaypoints {
		return nil, fmt.Errorf("maximum waypoints reached (%d)", m.config.MaxWaypoints)
	}

	// Create waypoint
	waypoint := &Waypoint{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		SystemID:    systemID,
		SystemName:  systemName,
		PlanetID:    planetID,
		PlanetName:  planetName,
		Color:       "cyan",
		Icon:        "⭐",
		CreatedAt:   time.Now(),
		Visited:     false,
	}

	waypointSet.Waypoints = append(waypointSet.Waypoints, waypoint)

	log.Info("Waypoint added: player=%s, name=%s, system=%s", playerID, name, systemName)
	return waypoint, nil
}

// RemoveWaypoint removes a waypoint
func (m *Manager) RemoveWaypoint(ctx context.Context, playerID, waypointID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	waypointSet := m.waypoints[playerID]
	if waypointSet == nil {
		return fmt.Errorf("no waypoints found")
	}

	// Find and remove waypoint
	for i, waypoint := range waypointSet.Waypoints {
		if waypoint.ID == waypointID {
			waypointSet.Waypoints = append(waypointSet.Waypoints[:i], waypointSet.Waypoints[i+1:]...)
			log.Info("Waypoint removed: player=%s, waypoint=%s", playerID, waypoint.Name)
			return nil
		}
	}

	return fmt.Errorf("waypoint not found")
}

// SetActiveWaypoint sets a waypoint as the navigation target
func (m *Manager) SetActiveWaypoint(ctx context.Context, playerID, waypointID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	waypointSet := m.waypoints[playerID]
	if waypointSet == nil {
		return fmt.Errorf("no waypoints found")
	}

	// Verify waypoint exists
	found := false
	for _, waypoint := range waypointSet.Waypoints {
		if waypoint.ID == waypointID {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("waypoint not found")
	}

	waypointSet.Active = &waypointID
	log.Info("Active waypoint set: player=%s, waypoint=%s", playerID, waypointID)
	return nil
}

// GetWaypoints retrieves all waypoints for a player
func (m *Manager) GetWaypoints(playerID uuid.UUID) []*Waypoint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	waypointSet := m.waypoints[playerID]
	if waypointSet == nil {
		return []*Waypoint{}
	}

	// Return copy
	waypoints := make([]*Waypoint, len(waypointSet.Waypoints))
	copy(waypoints, waypointSet.Waypoints)
	return waypoints
}

// CheckWaypointProximity checks if player is near an active waypoint
func (m *Manager) CheckWaypointProximity(playerID uuid.UUID, currentSystemID uuid.UUID) {
	m.mu.RLock()
	waypointSet := m.waypoints[playerID]
	m.mu.RUnlock()

	if waypointSet == nil || waypointSet.Active == nil {
		return
	}

	// Find active waypoint
	for _, waypoint := range waypointSet.Waypoints {
		if waypoint.ID == *waypointSet.Active {
			if waypoint.SystemID == currentSystemID && !waypoint.Visited {
				waypoint.Visited = true
				log.Info("Waypoint reached: player=%s, waypoint=%s", playerID, waypoint.Name)

				if m.onWaypointReached != nil {
					go m.onWaypointReached(playerID, waypoint)
				}
			}
			break
		}
	}
}

// ============================================================================
// AUTO-TRADING SYSTEM
// ============================================================================

// EnableAutoTrading enables automatic trading for a player
func (m *Manager) EnableAutoTrading(ctx context.Context, playerID uuid.UUID, minProfit int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.config.AutoTradeEnabled {
		return fmt.Errorf("auto-trading is disabled")
	}

	autoTrader := m.autoTraders[playerID]
	if autoTrader == nil {
		autoTrader = &AutoTrader{
			PlayerID: playerID,
			Routes:   []uuid.UUID{},
		}
		m.autoTraders[playerID] = autoTrader
	}

	autoTrader.Enabled = true
	autoTrader.MinProfit = minProfit

	log.Info("Auto-trading enabled: player=%s, min_profit=%d", playerID, minProfit)
	return nil
}

// DisableAutoTrading disables automatic trading
func (m *Manager) DisableAutoTrading(ctx context.Context, playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	autoTrader := m.autoTraders[playerID]
	if autoTrader == nil {
		return fmt.Errorf("auto-trading not configured")
	}

	autoTrader.Enabled = false
	log.Info("Auto-trading disabled: player=%s", playerID)
	return nil
}

// AddTradeRoute adds a profitable trade route
func (m *Manager) AddTradeRoute(ctx context.Context, playerID uuid.UUID, name string, fromSystemID, toSystemID uuid.UUID, fromSystem, toSystem, commodity string, buyPrice, sellPrice int64, distance float64) (*TradeRoute, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	route := &TradeRoute{
		ID:           uuid.New(),
		PlayerID:     playerID,
		Name:         name,
		FromSystemID: fromSystemID,
		ToSystemID:   toSystemID,
		FromSystem:   fromSystem,
		ToSystem:     toSystem,
		Commodity:    commodity,
		BuyPrice:     buyPrice,
		SellPrice:    sellPrice,
		Profit:       sellPrice - buyPrice,
		Distance:     distance,
		LastUpdated:  time.Now(),
		Active:       true,
	}

	m.tradeRoutes[route.ID] = route

	// Add to auto-trader if exists
	if autoTrader := m.autoTraders[playerID]; autoTrader != nil {
		autoTrader.Routes = append(autoTrader.Routes, route.ID)
	}

	log.Info("Trade route added: player=%s, route=%s, profit=%d", playerID, name, route.Profit)
	return route, nil
}

// GetTradeRoutes retrieves all active trade routes for a player
func (m *Manager) GetTradeRoutes(playerID uuid.UUID) []*TradeRoute {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var routes []*TradeRoute
	for _, route := range m.tradeRoutes {
		if route.PlayerID == playerID && route.Active {
			routes = append(routes, route)
		}
	}
	return routes
}

// autoTradeWorker runs automatic trading
func (m *Manager) autoTradeWorker() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.AutoTradeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.processAutoTrades()
		case <-m.stopChan:
			return
		}
	}
}

// processAutoTrades executes profitable trades automatically
func (m *Manager) processAutoTrades() {
	m.mu.RLock()
	traders := make([]*AutoTrader, 0, len(m.autoTraders))
	for _, trader := range m.autoTraders {
		if trader.Enabled {
			traders = append(traders, trader)
		}
	}
	m.mu.RUnlock()

	for _, trader := range traders {
		// TODO: Implement actual trade execution logic
		// For now, just track that it ran
		trader.LastRun = time.Now()
	}
}

// ============================================================================
// QUICK COMMANDS
// ============================================================================

// AddQuickCommand adds a command shortcut
func (m *Manager) AddQuickCommand(ctx context.Context, playerID uuid.UUID, alias, command string, args []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(alias) > m.config.CommandAliasLimit {
		return fmt.Errorf("alias too long (max %d characters)", m.config.CommandAliasLimit)
	}

	commands := m.quickCommands[playerID]
	if commands == nil {
		commands = make(map[string]string)
		m.quickCommands[playerID] = commands
	}

	if len(commands) >= m.config.MaxQuickCommands {
		return fmt.Errorf("maximum quick commands reached (%d)", m.config.MaxQuickCommands)
	}

	// Store full command with args
	fullCommand := command
	for _, arg := range args {
		fullCommand += " " + arg
	}

	commands[alias] = fullCommand
	log.Info("Quick command added: player=%s, alias=%s, command=%s", playerID, alias, command)
	return nil
}

// GetQuickCommand retrieves a quick command by alias
func (m *Manager) GetQuickCommand(playerID uuid.UUID, alias string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	commands := m.quickCommands[playerID]
	if commands == nil {
		return "", false
	}

	command, exists := commands[alias]
	return command, exists
}

// RemoveQuickCommand removes a quick command
func (m *Manager) RemoveQuickCommand(ctx context.Context, playerID uuid.UUID, alias string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	commands := m.quickCommands[playerID]
	if commands == nil {
		return fmt.Errorf("no quick commands found")
	}

	if _, exists := commands[alias]; !exists {
		return fmt.Errorf("quick command not found")
	}

	delete(commands, alias)
	log.Info("Quick command removed: player=%s, alias=%s", playerID, alias)
	return nil
}

// ============================================================================
// NOTIFICATION SYSTEM
// ============================================================================

// SendNotification sends a notification to a player
func (m *Manager) SendNotification(playerID uuid.UUID, notifType, title, message, priority string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	prefs := m.notifications[playerID]
	if prefs == nil {
		prefs = &NotificationPrefs{
			PlayerID:     playerID,
			EnabledTypes: make(map[string]bool),
			History:      []*Notification{},
			MaxHistory:   m.config.NotificationHistory,
		}
		m.notifications[playerID] = prefs
	}

	// Check if notification type is enabled
	if enabled, exists := prefs.EnabledTypes[notifType]; exists && !enabled {
		return
	}

	notification := &Notification{
		ID:        uuid.New(),
		Type:      notifType,
		Title:     title,
		Message:   message,
		Priority:  priority,
		Timestamp: time.Now(),
		Read:      false,
	}

	// Add to history
	prefs.History = append(prefs.History, notification)

	// Trim history if needed
	if len(prefs.History) > prefs.MaxHistory {
		prefs.History = prefs.History[len(prefs.History)-prefs.MaxHistory:]
	}

	log.Debug("Notification sent: player=%s, type=%s, title=%s", playerID, notifType, title)

	if m.onNotification != nil {
		go m.onNotification(playerID, notification)
	}
}

// GetNotifications retrieves recent notifications for a player
func (m *Manager) GetNotifications(playerID uuid.UUID, limit int) []*Notification {
	m.mu.RLock()
	defer m.mu.RUnlock()

	prefs := m.notifications[playerID]
	if prefs == nil {
		return []*Notification{}
	}

	// Return most recent notifications
	start := 0
	if len(prefs.History) > limit {
		start = len(prefs.History) - limit
	}

	notifications := make([]*Notification, len(prefs.History)-start)
	copy(notifications, prefs.History[start:])
	return notifications
}

// MarkNotificationRead marks a notification as read
func (m *Manager) MarkNotificationRead(playerID, notificationID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	prefs := m.notifications[playerID]
	if prefs == nil {
		return fmt.Errorf("no notifications found")
	}

	for _, notification := range prefs.History {
		if notification.ID == notificationID {
			notification.Read = true
			return nil
		}
	}

	return fmt.Errorf("notification not found")
}

// SetNotificationPreference sets a notification type preference
func (m *Manager) SetNotificationPreference(playerID uuid.UUID, notifType string, enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	prefs := m.notifications[playerID]
	if prefs == nil {
		prefs = &NotificationPrefs{
			PlayerID:     playerID,
			EnabledTypes: make(map[string]bool),
			History:      []*Notification{},
			MaxHistory:   m.config.NotificationHistory,
		}
		m.notifications[playerID] = prefs
	}

	prefs.EnabledTypes[notifType] = enabled
	log.Info("Notification preference set: player=%s, type=%s, enabled=%v", playerID, notifType, enabled)
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// GetStats returns QoL statistics
func (m *Manager) GetStats() QoLStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := QoLStats{
		ActiveAutoTraders: 0,
		TotalWaypoints:    0,
		TotalTradeRoutes:  len(m.tradeRoutes),
	}

	for _, trader := range m.autoTraders {
		if trader.Enabled {
			stats.ActiveAutoTraders++
		}
	}

	for _, waypointSet := range m.waypoints {
		stats.TotalWaypoints += len(waypointSet.Waypoints)
	}

	return stats
}

// QoLStats contains QoL statistics
type QoLStats struct {
	ActiveAutoTraders int `json:"active_auto_traders"`
	TotalWaypoints    int `json:"total_waypoints"`
	TotalTradeRoutes  int `json:"total_trade_routes"`
}

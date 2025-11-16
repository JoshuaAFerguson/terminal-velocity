// File: internal/traderoutes/calculator.go
// Project: Terminal Velocity
// Description: Trade route calculator and optimization
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-14

package traderoutes

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

var log = logger.WithComponent("TradeRoutes")

// Calculator provides trade route optimization
type Calculator struct {
	systemRepo *database.SystemRepository
	marketRepo *database.MarketRepository
}

// NewCalculator creates a new trade route calculator
func NewCalculator(systemRepo *database.SystemRepository, marketRepo *database.MarketRepository) *Calculator {
	return &Calculator{
		systemRepo: systemRepo,
		marketRepo: marketRepo,
	}
}

// TradeRoute represents a profitable trade route
type TradeRoute struct {
	FromSystem   *models.StarSystem
	ToSystem     *models.StarSystem
	Commodity    string
	BuyPrice     float64
	SellPrice    float64
	ProfitPerUnit float64
	Distance     int
	JumpPath     []uuid.UUID
	ProfitPerJump float64
	TotalProfit  int64 // For max cargo
	ROI          float64 // Return on investment percentage
}

// RouteOptions configures route finding
type RouteOptions struct {
	MaxJumps      int     // Maximum jumps to consider (0 = unlimited)
	MinProfit     float64 // Minimum profit per unit
	CargoCapacity int     // Ship's cargo capacity
	CurrentSystem uuid.UUID
	IncludeIllegal bool   // Include illegal goods
	MaxDistance   int     // Maximum total distance
}

// DefaultRouteOptions returns sensible defaults
func DefaultRouteOptions() *RouteOptions {
	return &RouteOptions{
		MaxJumps:       5,
		MinProfit:      10.0,
		CargoCapacity:  10,
		IncludeIllegal: false,
		MaxDistance:    10,
	}
}

// FindBestRoutes finds the most profitable trade routes
func (c *Calculator) FindBestRoutes(ctx context.Context, opts *RouteOptions) ([]*TradeRoute, error) {
	if opts == nil {
		opts = DefaultRouteOptions()
	}

	log.Debug("Finding trade routes: maxJumps=%d, minProfit=%.2f, cargo=%d",
		opts.MaxJumps, opts.MinProfit, opts.CargoCapacity)

	// Get all systems
	systems, err := c.systemRepo.ListSystems(ctx)
	if err != nil {
		log.Error("Failed to list systems: %v", err)
		return nil, err
	}

	// Build routes
	routes := make([]*TradeRoute, 0)

	// For each system pair, check all commodities
	for _, fromSystem := range systems {
		for _, toSystem := range systems {
			if fromSystem.ID == toSystem.ID {
				continue
			}

			// Calculate distance
			distance := c.calculateDistance(fromSystem, toSystem)
			if opts.MaxDistance > 0 && distance > opts.MaxDistance {
				continue
			}

			// Find jump path
			jumpPath, jumps := c.findShortestPath(systems, fromSystem.ID, toSystem.ID, opts.MaxJumps)
			if jumpPath == nil || (opts.MaxJumps > 0 && jumps > opts.MaxJumps) {
				continue
			}

			// Check all commodities
			for _, commodity := range models.StandardCommodities {
				// Skip illegal goods if not included
				if !opts.IncludeIllegal && commodity.IsIllegal(fromSystem.GovernmentID) {
					continue
				}

				// Calculate estimated prices (without market data)
				// Buy at fromSystem, sell at toSystem
				buyModifier := models.GetPriceModifier(commodity.TechLevel, fromSystem.TechLevel, false)
				sellModifier := models.GetPriceModifier(commodity.TechLevel, toSystem.TechLevel, false)

				buyPrice := float64(commodity.BasePrice) * buyModifier
				sellPrice := float64(commodity.BasePrice) * sellModifier

				profit := sellPrice - buyPrice

				// Skip unprofitable routes
				if profit < opts.MinProfit {
					continue
				}

				// Calculate total profit for full cargo
				totalProfit := int64(profit * float64(opts.CargoCapacity))
				roi := (profit / buyPrice) * 100

				route := &TradeRoute{
					FromSystem:    fromSystem,
					ToSystem:      toSystem,
					Commodity:     commodity.Name,
					BuyPrice:      buyPrice,
					SellPrice:     sellPrice,
					ProfitPerUnit: profit,
					Distance:      distance,
					JumpPath:      jumpPath,
					ProfitPerJump: profit / float64(jumps),
					TotalProfit:   totalProfit,
					ROI:           roi,
				}

				routes = append(routes, route)
			}
		}
	}

	// Sort by total profit (descending)
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].TotalProfit > routes[j].TotalProfit
	})

	// Return top 50 routes
	if len(routes) > 50 {
		routes = routes[:50]
	}

	log.Debug("Found %d profitable trade routes", len(routes))
	return routes, nil
}

// FindRoutesFromSystem finds best routes starting from a specific system
func (c *Calculator) FindRoutesFromSystem(ctx context.Context, systemID uuid.UUID, opts *RouteOptions) ([]*TradeRoute, error) {
	if opts == nil {
		opts = DefaultRouteOptions()
	}
	opts.CurrentSystem = systemID

	log.Debug("Finding routes from system %s", systemID)

	allRoutes, err := c.FindBestRoutes(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Filter to routes starting from this system
	filtered := make([]*TradeRoute, 0)
	for _, route := range allRoutes {
		if route.FromSystem.ID == systemID {
			filtered = append(filtered, route)
		}
	}

	return filtered, nil
}

// FindRoutesBetween finds profitable routes between two specific systems
func (c *Calculator) FindRoutesBetween(ctx context.Context, fromID, toID uuid.UUID, cargoCapacity int) ([]*TradeRoute, error) {
	log.Debug("Finding routes between %s and %s", fromID, toID)

	// Get systems
	fromSystem, err := c.systemRepo.GetSystemByID(ctx, fromID)
	if err != nil {
		return nil, err
	}

	toSystem, err := c.systemRepo.GetSystemByID(ctx, toID)
	if err != nil {
		return nil, err
	}

	// Get all systems for pathfinding
	systems, err := c.systemRepo.ListSystems(ctx)
	if err != nil {
		return nil, err
	}

	// Find jump path
	jumpPath, _ := c.findShortestPath(systems, fromID, toID, 0)
	if jumpPath == nil {
		log.Warn("No path found between systems")
		return nil, nil
	}

	distance := c.calculateDistance(fromSystem, toSystem)

	routes := make([]*TradeRoute, 0)

	// Check all commodities
	for _, commodity := range models.StandardCommodities {
		// Calculate estimated prices (without market data)
		buyModifier := models.GetPriceModifier(commodity.TechLevel, fromSystem.TechLevel, false)
		sellModifier := models.GetPriceModifier(commodity.TechLevel, toSystem.TechLevel, false)

		buyPrice := float64(commodity.BasePrice) * buyModifier
		sellPrice := float64(commodity.BasePrice) * sellModifier

		profit := sellPrice - buyPrice

		if profit > 0 {
			totalProfit := int64(profit * float64(cargoCapacity))
			roi := (profit / buyPrice) * 100

			route := &TradeRoute{
				FromSystem:    fromSystem,
				ToSystem:      toSystem,
				Commodity:     commodity.Name,
				BuyPrice:      buyPrice,
				SellPrice:     sellPrice,
				ProfitPerUnit: profit,
				Distance:      distance,
				JumpPath:      jumpPath,
				ProfitPerJump: profit / float64(len(jumpPath)-1),
				TotalProfit:   totalProfit,
				ROI:           roi,
			}

			routes = append(routes, route)
		}
	}

	// Sort by profit per unit
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].ProfitPerUnit > routes[j].ProfitPerUnit
	})

	return routes, nil
}

// calculateDistance calculates Euclidean distance between systems
func (c *Calculator) calculateDistance(from, to *models.StarSystem) int {
	dx := float64(to.Position.X - from.Position.X)
	dy := float64(to.Position.Y - from.Position.Y)
	return int(math.Sqrt(dx*dx + dy*dy))
}

// findShortestPath finds the shortest jump route between two systems using Dijkstra's
// shortest path algorithm.
//
// Algorithm: Dijkstra's Shortest Path (without priority queue)
//  1. Initialize:
//     - Set distance[start] = 0, all others = ∞
//     - Mark all nodes unvisited
//     - Build adjacency list from ConnectedSystems
//  2. Main loop:
//     - Find unvisited node with minimum distance (linear search in O(V))
//     - If found destination or no reachable nodes, stop
//     - Mark current node as visited
//     - For each neighbor of current node:
//       * Calculate tentative distance = dist[current] + 1 (all edges weight 1)
//       * If tentative < known distance, update dist[neighbor] and prev[neighbor]
//  3. Reconstruct path:
//     - Backtrack from destination using prev[] pointers
//     - Build path by prepending each system until reaching start
//
// Time Complexity: O(V²) where V is number of systems
//   - O(V) iterations of main loop
//   - O(V) to find minimum distance node each iteration
//   - O(E) total edge relaxations across all iterations
//   - Dominated by O(V²) minimum-finding
//
// Note: This could be optimized to O((V + E) log V) using a priority queue/min-heap,
// but for typical universe sizes (<1000 systems), the simpler implementation is adequate
// and easier to understand.
//
// Pathfinding Properties:
//   - All jump routes have equal cost (1 jump per edge)
//   - This effectively makes it BFS, but Dijkstra generalizes better
//   - Returns shortest path in terms of jump count, not distance
//   - Path is guaranteed to be optimal (shortest possible)
//
// Parameters:
//   - systems: All star systems in the universe
//   - fromID: Starting system UUID
//   - toID: Destination system UUID
//   - maxJumps: Maximum jumps allowed (0 = unlimited)
//
// Returns:
//   - path: Ordered list of system UUIDs from start to destination (nil if no path)
//   - jumps: Number of jumps in path (len(path) - 1)
//
// Thread Safety: Safe for concurrent calls (read-only operations on systems).
func (c *Calculator) findShortestPath(systems []*models.StarSystem, fromID, toID uuid.UUID, maxJumps int) ([]uuid.UUID, int) {
	// Build adjacency map
	adjacency := make(map[uuid.UUID][]uuid.UUID)
	systemMap := make(map[uuid.UUID]*models.StarSystem)

	for _, system := range systems {
		systemMap[system.ID] = system
		adjacency[system.ID] = system.ConnectedSystems
	}

	// Dijkstra's algorithm
	dist := make(map[uuid.UUID]int)
	prev := make(map[uuid.UUID]uuid.UUID)
	visited := make(map[uuid.UUID]bool)

	// Initialize distances
	for _, system := range systems {
		dist[system.ID] = math.MaxInt32
	}
	dist[fromID] = 0

	for {
		// Find unvisited node with minimum distance
		var current uuid.UUID
		minDist := math.MaxInt32
		found := false

		for id := range dist {
			if !visited[id] && dist[id] < minDist {
				current = id
				minDist = dist[id]
				found = true
			}
		}

		if !found || current == toID {
			break
		}

		visited[current] = true

		// Check max jumps constraint
		if maxJumps > 0 && dist[current] >= maxJumps {
			continue
		}

		// Update neighbors
		for _, neighborID := range adjacency[current] {
			newDist := dist[current] + 1

			if newDist < dist[neighborID] {
				dist[neighborID] = newDist
				prev[neighborID] = current
			}
		}
	}

	// Reconstruct path
	if dist[toID] == math.MaxInt32 {
		return nil, 0 // No path found
	}

	path := make([]uuid.UUID, 0)
	current := toID

	for current != fromID {
		path = append([]uuid.UUID{current}, path...)
		var ok bool
		current, ok = prev[current]
		if !ok {
			return nil, 0
		}
	}

	path = append([]uuid.UUID{fromID}, path...)
	return path, len(path) - 1
}

// NavigationPath represents a planned route through space
type NavigationPath struct {
	Systems      []*models.StarSystem
	TotalJumps   int
	TotalDistance int
	FuelRequired int
	Waypoints    []string // System names
}

// PlanRoute creates a navigation plan between two systems
func (c *Calculator) PlanRoute(ctx context.Context, fromID, toID uuid.UUID) (*NavigationPath, error) {
	log.Debug("Planning route from %s to %s", fromID, toID)

	// Get all systems
	systems, err := c.systemRepo.ListSystems(ctx)
	if err != nil {
		return nil, err
	}

	// Find shortest path
	jumpPath, jumps := c.findShortestPath(systems, fromID, toID, 0)
	if jumpPath == nil {
		log.Warn("No path found between systems")
		return nil, nil
	}

	// Build navigation path
	systemMap := make(map[uuid.UUID]*models.StarSystem)
	for _, system := range systems {
		systemMap[system.ID] = system
	}

	pathSystems := make([]*models.StarSystem, 0, len(jumpPath))
	waypoints := make([]string, 0, len(jumpPath))
	totalDistance := 0

	for i, sysID := range jumpPath {
		system, exists := systemMap[sysID]
		if !exists {
			return nil, fmt.Errorf("system not found in map: %s", sysID)
		}
		pathSystems = append(pathSystems, system)
		waypoints = append(waypoints, system.Name)

		if i > 0 {
			prevSystem, exists := systemMap[jumpPath[i-1]]
			if !exists {
				return nil, fmt.Errorf("previous system not found in map: %s", jumpPath[i-1])
			}
			totalDistance += c.calculateDistance(prevSystem, system)
		}
	}

	// Estimate fuel (1 fuel per jump)
	fuelRequired := jumps

	nav := &NavigationPath{
		Systems:       pathSystems,
		TotalJumps:    jumps,
		TotalDistance: totalDistance,
		FuelRequired:  fuelRequired,
		Waypoints:     waypoints,
	}

	log.Debug("Route planned: %d jumps, %d distance, %d fuel", jumps, totalDistance, fuelRequired)
	return nav, nil
}

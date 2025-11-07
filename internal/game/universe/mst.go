package universe

import (
	"sort"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// Edge represents a connection between two systems with a distance
type Edge struct {
	From     uuid.UUID
	To       uuid.UUID
	Distance float64
}

// UnionFind data structure for Kruskal's algorithm
type UnionFind struct {
	parent map[uuid.UUID]uuid.UUID
	rank   map[uuid.UUID]int
}

// NewUnionFind creates a new union-find structure
func NewUnionFind(systems []models.StarSystem) *UnionFind {
	uf := &UnionFind{
		parent: make(map[uuid.UUID]uuid.UUID),
		rank:   make(map[uuid.UUID]int),
	}

	for i := range systems {
		uf.parent[systems[i].ID] = systems[i].ID
		uf.rank[systems[i].ID] = 0
	}

	return uf
}

// Find finds the root of a set
func (uf *UnionFind) Find(id uuid.UUID) uuid.UUID {
	if uf.parent[id] != id {
		uf.parent[id] = uf.Find(uf.parent[id]) // Path compression
	}
	return uf.parent[id]
}

// Union merges two sets
func (uf *UnionFind) Union(a, b uuid.UUID) bool {
	rootA := uf.Find(a)
	rootB := uf.Find(b)

	if rootA == rootB {
		return false // Already in same set
	}

	// Union by rank
	if uf.rank[rootA] < uf.rank[rootB] {
		uf.parent[rootA] = rootB
	} else if uf.rank[rootA] > uf.rank[rootB] {
		uf.parent[rootB] = rootA
	} else {
		uf.parent[rootB] = rootA
		uf.rank[rootA]++
	}

	return true
}

// generateMinimumSpanningTree creates an MST using Kruskal's algorithm
func generateMinimumSpanningTree(systems []models.StarSystem) []Edge {
	edges := make([]Edge, 0)

	// Create all possible edges
	for i := range systems {
		for j := i + 1; j < len(systems); j++ {
			distance := systems[i].Position.DistanceTo(systems[j].Position)
			edges = append(edges, Edge{
				From:     systems[i].ID,
				To:       systems[j].ID,
				Distance: distance,
			})
		}
	}

	// Sort edges by distance
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].Distance < edges[j].Distance
	})

	// Kruskal's algorithm
	uf := NewUnionFind(systems)
	mst := make([]Edge, 0)

	for _, edge := range edges {
		if uf.Union(edge.From, edge.To) {
			mst = append(mst, edge)

			// MST is complete when we have n-1 edges
			if len(mst) == len(systems)-1 {
				break
			}
		}
	}

	return mst
}

// addExtraConnections adds additional routes beyond MST for interesting topology
func (g *Generator) addExtraConnections(systems []models.StarSystem, mstEdges []Edge) {
	// Create a map for quick lookup
	systemMap := make(map[uuid.UUID]*models.StarSystem)
	for i := range systems {
		systemMap[systems[i].ID] = &systems[i]
	}

	// Add MST edges
	for _, edge := range mstEdges {
		if system, ok := systemMap[edge.From]; ok {
			system.AddConnection(edge.To)
		}
		if system, ok := systemMap[edge.To]; ok {
			system.AddConnection(edge.From)
		}
	}

	// Add extra connections to create interesting shortcuts and alternate routes
	for i := range systems {
		currentConnections := len(systems[i].ConnectedSystems)

		// If system has too few connections, add more
		if currentConnections < g.config.MinConnections {
			needed := g.config.MinConnections - currentConnections
			g.addNearestConnections(&systems[i], systems, needed)
		}

		// Randomly add extra connections (30% chance per system)
		if g.rand.Float64() < 0.3 && currentConnections < g.config.MaxConnections {
			extraConnections := 1 + g.rand.Intn(2) // 1-2 extra
			g.addNearestConnections(&systems[i], systems, extraConnections)
		}
	}
}

// addNearestConnections adds n nearest unconnected systems
func (g *Generator) addNearestConnections(system *models.StarSystem, allSystems []models.StarSystem, n int) {
	type distPair struct {
		id       uuid.UUID
		distance float64
	}

	// Find unconnected systems
	unconnected := make([]distPair, 0)

	for i := range allSystems {
		if allSystems[i].ID == system.ID {
			continue
		}

		// Skip if already connected
		if system.IsConnectedTo(allSystems[i].ID) {
			continue
		}

		dist := system.Position.DistanceTo(allSystems[i].Position)
		unconnected = append(unconnected, distPair{allSystems[i].ID, dist})
	}

	// Sort by distance
	sort.Slice(unconnected, func(i, j int) bool {
		return unconnected[i].distance < unconnected[j].distance
	})

	// Add n nearest
	added := 0
	for i := 0; i < len(unconnected) && added < n; i++ {
		if len(system.ConnectedSystems) < g.config.MaxConnections {
			system.AddConnection(unconnected[i].id)
			added++
		}
	}
}

// UpdatedGenerateJumpRoutes creates jump routes using MST + extra connections
func (g *Generator) UpdatedGenerateJumpRoutes(systems []models.StarSystem) {
	// Phase 1: Create minimum spanning tree (ensures all connected)
	mstEdges := generateMinimumSpanningTree(systems)

	// Phase 2: Add MST edges and extra connections
	g.addExtraConnections(systems, mstEdges)

	// Phase 3: Make connections bidirectional
	g.makeBidirectional(systems)
}

// makeBidirectional ensures all connections are bidirectional
func (g *Generator) makeBidirectional(systems []models.StarSystem) {
	systemMap := make(map[uuid.UUID]*models.StarSystem)
	for i := range systems {
		systemMap[systems[i].ID] = &systems[i]
	}

	for i := range systems {
		for _, connectedID := range systems[i].ConnectedSystems {
			if connectedSystem, ok := systemMap[connectedID]; ok {
				// Ensure reverse connection exists
				if !connectedSystem.IsConnectedTo(systems[i].ID) {
					connectedSystem.AddConnection(systems[i].ID)
				}
			}
		}
	}
}

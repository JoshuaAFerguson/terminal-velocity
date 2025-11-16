// File: internal/game/universe/mst.go
// Project: Terminal Velocity
// Description: Procedural universe generation: mst
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

// Package universe implements Kruskal's Minimum Spanning Tree (MST) algorithm for
// generating jump route networks between star systems.
//
// The MST algorithm ensures:
//   - All systems are reachable (graph is connected)
//   - Total jump route distance is minimized
//   - No cycles exist in the base network
//
// Algorithm: Kruskal's MST with Union-Find
//  1. Create all possible edges between systems (O(n²))
//  2. Sort edges by distance (O(n² log n))
//  3. Greedily add shortest edges that don't create cycles (O(n² α(n)))
//     where α(n) is the inverse Ackermann function (effectively constant)
//  4. Stop when we have n-1 edges (tree complete)
//
// After MST generation, additional edges are added to create shortcuts and alternate
// routes for more interesting gameplay topology.
//
// Time Complexity: O(n² log n) where n is the number of systems
// Space Complexity: O(n²) for edge storage
package universe

import (
	"sort"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

// Edge represents a potential hyperspace jump route between two star systems.
//
// During MST generation, all possible edges are created and sorted by distance.
// The shortest edges that don't create cycles are selected to form the base
// jump route network.
type Edge struct {
	From     uuid.UUID  // Source system ID
	To       uuid.UUID  // Destination system ID
	Distance float64    // Euclidean distance in light-years
}

// UnionFind implements the disjoint-set data structure for Kruskal's algorithm.
//
// This structure efficiently tracks which systems are connected and prevents
// cycle formation when adding edges to the MST. It uses two optimizations:
//
//  1. Path Compression: During Find(), flatten tree structure to speed up future queries
//  2. Union by Rank: During Union(), attach smaller tree under larger tree's root
//
// Time Complexity:
//   - Find(): O(α(n)) amortized where α is inverse Ackermann (effectively O(1))
//   - Union(): O(α(n)) amortized
//
// These optimizations make UnionFind nearly O(1) for practical purposes, allowing
// Kruskal's algorithm to run efficiently even on large graphs.
//
// Thread Safety: NOT thread-safe. Should only be used within a single goroutine.
type UnionFind struct {
	parent map[uuid.UUID]uuid.UUID  // parent[x] = parent of node x in the forest
	rank   map[uuid.UUID]int        // rank[x] = approximate depth of tree rooted at x
}

// NewUnionFind creates and initializes a union-find structure for the given systems.
//
// Initially, each system is in its own set (i.e., each system is its own root with rank 0).
// As edges are added during MST construction, sets are merged until all systems belong
// to a single connected component.
//
// Parameters:
//   - systems: Slice of all star systems to track
//
// Returns:
//
//	Initialized UnionFind structure ready for use in Kruskal's algorithm
func NewUnionFind(systems []models.StarSystem) *UnionFind {
	uf := &UnionFind{
		parent: make(map[uuid.UUID]uuid.UUID),
		rank:   make(map[uuid.UUID]int),
	}

	// Initialize each system as its own parent (separate set)
	for i := range systems {
		uf.parent[systems[i].ID] = systems[i].ID
		uf.rank[systems[i].ID] = 0
	}

	return uf
}

// Find finds the root representative of the set containing the given system.
//
// This implements path compression optimization: as we traverse up the tree to find
// the root, we update all intermediate nodes to point directly to the root. This
// flattens the tree structure and makes future Find() operations faster.
//
// Algorithm (with path compression):
//  1. If node is its own parent, it's the root - return it
//  2. Otherwise, recursively find the root
//  3. Update node's parent to point directly to root (compression)
//  4. Return the root
//
// Time Complexity: O(α(n)) amortized where α is the inverse Ackermann function
//
// Parameters:
//   - id: System UUID to find the set representative for
//
// Returns:
//
//	UUID of the root system representing this set
func (uf *UnionFind) Find(id uuid.UUID) uuid.UUID {
	if uf.parent[id] != id {
		uf.parent[id] = uf.Find(uf.parent[id]) // Path compression
	}
	return uf.parent[id]
}

// Union merges the sets containing systems a and b.
//
// This implements union by rank optimization: always attach the shorter tree under
// the root of the taller tree. This keeps trees balanced and prevents degeneration
// into linked lists.
//
// Algorithm (union by rank):
//  1. Find roots of both sets
//  2. If roots are the same, sets already merged - return false
//  3. Compare ranks:
//     - Attach smaller rank tree under larger rank tree
//     - If ranks equal, pick one as root and increment its rank
//  4. Return true to indicate successful merge
//
// Time Complexity: O(α(n)) amortized
//
// Parameters:
//   - a: First system UUID
//   - b: Second system UUID
//
// Returns:
//
//	true if sets were merged (were previously separate)
//	false if sets were already connected (would create cycle)
//
// This return value is critical for Kruskal's algorithm to avoid adding edges
// that would create cycles in the MST.
func (uf *UnionFind) Union(a, b uuid.UUID) bool {
	rootA := uf.Find(a)
	rootB := uf.Find(b)

	if rootA == rootB {
		return false // Already in same set - adding edge would create cycle
	}

	// Union by rank: attach smaller tree under larger tree
	if uf.rank[rootA] < uf.rank[rootB] {
		uf.parent[rootA] = rootB
	} else if uf.rank[rootA] > uf.rank[rootB] {
		uf.parent[rootB] = rootA
	} else {
		// Ranks equal - arbitrarily choose rootA as new root
		uf.parent[rootB] = rootA
		uf.rank[rootA]++  // Increment rank since tree got deeper
	}

	return true
}

// generateMinimumSpanningTree creates a Minimum Spanning Tree of jump routes using
// Kruskal's algorithm.
//
// The MST ensures all star systems are reachable via jump routes while minimizing the
// total distance traveled. This creates a connected graph with exactly n-1 edges for
// n systems (the definition of a tree).
//
// Algorithm (Kruskal's MST):
//  1. Generate all possible edges (complete graph)
//     - For n systems, this creates n(n-1)/2 edges
//     - Each edge stores the Euclidean distance between systems
//  2. Sort all edges by distance (shortest first)
//  3. Greedily select edges:
//     - Take shortest edge that doesn't create a cycle
//     - Use Union-Find to detect cycles in O(α(n)) time
//     - Add edge to MST and continue
//  4. Stop when MST has n-1 edges (tree is complete)
//
// Time Complexity: O(n² log n)
//   - O(n²) to generate all edges
//   - O(n² log n) to sort edges
//   - O(n² α(n)) for union-find operations ≈ O(n²)
//   - Total: dominated by sorting
//
// Space Complexity: O(n²) to store all candidate edges
//
// Why Kruskal's Algorithm?
//   - Produces globally optimal solution (minimal total distance)
//   - Simple to implement with Union-Find
//   - Well-suited for dense graphs (many possible routes)
//   - Deterministic results for reproducible universe generation
//
// Parameters:
//   - systems: All star systems in the universe
//
// Returns:
//
//	Slice of edges forming the MST (exactly len(systems)-1 edges)
//
// The returned MST provides the base jump route network. Additional edges are added
// later to create shortcuts and gameplay variety.
func generateMinimumSpanningTree(systems []models.StarSystem) []Edge {
	edges := make([]Edge, 0)

	// Phase 1: Create all possible edges (complete graph)
	// For n systems, this generates n(n-1)/2 edges
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

	// Phase 2: Sort edges by distance (shortest first)
	// This is the core of Kruskal's greedy approach
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].Distance < edges[j].Distance
	})

	// Phase 3: Build MST using Union-Find
	uf := NewUnionFind(systems)
	mst := make([]Edge, 0)

	// Greedily add shortest edges that don't create cycles
	for _, edge := range edges {
		// Union returns true if sets were merged (no cycle)
		if uf.Union(edge.From, edge.To) {
			mst = append(mst, edge)

			// MST is complete when we have exactly n-1 edges
			// Early termination optimization
			if len(mst) == len(systems)-1 {
				break
			}
		}
	}

	return mst
}

// addExtraConnections adds additional jump routes beyond the MST to create shortcuts
// and alternate paths.
//
// While the MST ensures connectivity with minimal total distance, it can create long
// detours between distant systems. Extra connections provide:
//   - Shortcuts that reduce travel time
//   - Alternate routes for strategic gameplay
//   - More interesting navigation choices
//   - Better connectivity for hub systems
//
// Algorithm:
//  1. Apply MST edges to system connectivity
//  2. For each system, ensure it meets MinConnections requirement
//  3. Add random extra connections (30% chance) up to MaxConnections
//  4. Make all connections bidirectional
//
// The extra connections are chosen by finding the nearest unconnected systems,
// which creates shortcuts to nearby clusters while maintaining reasonable distances.
//
// Parameters:
//   - systems: All star systems (modified in-place)
//   - mstEdges: Edges from the MST to use as base connectivity
//
// Thread Safety: NOT thread-safe. Modifies systems slice in-place.
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

// UpdatedGenerateJumpRoutes creates the complete jump route network using a three-phase
// approach: MST generation, extra connection addition, and bidirectional enforcement.
//
// This is the main entry point for jump route generation and creates a connected graph
// with interesting topology for gameplay.
//
// Three-Phase Algorithm:
//
// Phase 1 - MST Generation:
//   - Use Kruskal's algorithm to create minimum spanning tree
//   - Ensures all systems are reachable
//   - Minimizes total route distance
//   - Creates exactly n-1 edges for n systems
//
// Phase 2 - Extra Connections:
//   - Add edges beyond MST for shortcuts and alternate routes
//   - Ensure MinConnections requirement (default: 2)
//   - Randomly add extra edges up to MaxConnections (default: 5)
//   - 30% chance per system to get 1-2 extra connections
//
// Phase 3 - Bidirectional Enforcement:
//   - Make all jump routes bidirectional
//   - If system A can jump to B, then B can jump to A
//   - This matches player expectations and simplifies pathfinding
//
// Resulting Network Properties:
//   - Connected: All systems reachable from any starting point
//   - Sparse: Average connectivity 2-5 jumps per system
//   - Varied: Mix of hub systems (many connections) and remote systems (few connections)
//   - Realistic: Shorter jumps are more common (MST minimizes distance)
//
// Parameters:
//   - systems: All star systems to connect (modified in-place)
//
// Time Complexity: O(n² log n) dominated by MST generation
//
// Thread Safety: NOT thread-safe. Modifies systems slice in-place.
func (g *Generator) UpdatedGenerateJumpRoutes(systems []models.StarSystem) {
	// Phase 1: Create minimum spanning tree (ensures all connected)
	mstEdges := generateMinimumSpanningTree(systems)

	// Phase 2: Add MST edges and extra connections
	g.addExtraConnections(systems, mstEdges)

	// Phase 3: Make connections bidirectional
	g.makeBidirectional(systems)
}

// makeBidirectional ensures all jump routes work in both directions.
//
// For each system A with a connection to system B, this adds a reciprocal connection
// from B to A if it doesn't already exist. This creates symmetric navigation where
// jumpable routes work in both directions.
//
// Algorithm:
//  1. Build system lookup map for O(1) access
//  2. For each system A:
//     a. For each connection A → B:
//        b. Check if B → A exists
//        c. If not, add B → A
//
// Time Complexity: O(n * m) where n is systems and m is average connections per system
//
// Parameters:
//   - systems: All star systems (modified in-place to add reverse connections)
//
// Thread Safety: NOT thread-safe. Modifies systems slice in-place.
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

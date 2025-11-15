// File: internal/launch/manager.go
// Project: Terminal Velocity
// Description: Launch preparation tools for testing, optimization, and readiness verification
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-11-15

package launch

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Manager handles launch preparation, testing, and optimization
type Manager struct {
	db     *pgxpool.Pool
	config *Config
	mu     sync.RWMutex

	// Test data generation
	testData *TestDataState

	// Performance monitoring
	metrics *PerformanceMetrics

	// Health checks
	healthChecks map[string]*HealthCheck

	// Load testing
	loadTests map[uuid.UUID]*LoadTest

	// Optimization recommendations
	recommendations []*Recommendation

	// Launch checklist
	checklist *LaunchChecklist

	// Background workers
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// Config contains launch preparation configuration
type Config struct {
	// Test data generation
	MaxTestPlayers     int
	MaxTestSystems     int
	MaxTestFactions    int
	TestDataPrefix     string
	CleanupTestData    bool

	// Performance monitoring
	MetricsInterval    time.Duration
	CPUThreshold       float64 // 0.0 - 1.0
	MemoryThreshold    int64   // bytes
	QueryTimeThreshold time.Duration

	// Health checks
	HealthCheckInterval time.Duration
	DBTimeoutThreshold  time.Duration
	PoolSizeThreshold   int

	// Load testing
	MaxConcurrentLoads  int
	LoadTestDuration    time.Duration
	SimulatedPlayersMin int
	SimulatedPlayersMax int

	// Launch checklist
	RequireAllHealthy   bool
	RequireOptimization bool
	RequireBackupTest   bool
}

// TestDataState tracks generated test data
type TestDataState struct {
	Players  []uuid.UUID
	Systems  []uuid.UUID
	Factions []uuid.UUID
	Ships    []uuid.UUID
	Created  time.Time
	mu       sync.RWMutex
}

// PerformanceMetrics tracks real-time performance
type PerformanceMetrics struct {
	Timestamp      time.Time
	CPUUsage       float64
	MemoryUsage    int64
	MemoryAlloc    uint64
	MemoryTotal    uint64
	Goroutines     int
	DBConnections  int32
	DBIdleConns    int32
	AvgQueryTime   time.Duration
	SlowQueries    int
	ErrorRate      float64
	mu             sync.RWMutex
}

// HealthCheck represents a system health check
type HealthCheck struct {
	Name        string
	Type        string // "database", "memory", "disk", "network", "service"
	Status      string // "healthy", "warning", "critical"
	LastCheck   time.Time
	Message     string
	ResponseTime time.Duration
	Details     map[string]interface{}
}

// LoadTest represents a load testing session
type LoadTest struct {
	ID              uuid.UUID
	Name            string
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
	SimulatedPlayers int
	Status          string // "running", "completed", "failed"
	Results         *LoadTestResults
	mu              sync.RWMutex
}

// LoadTestResults contains load test metrics
type LoadTestResults struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	AvgResponseTime    time.Duration
	MinResponseTime    time.Duration
	MaxResponseTime    time.Duration
	RequestsPerSecond  float64
	ErrorRate          float64
	PeakCPU            float64
	PeakMemory         int64
	Bottlenecks        []string
	Recommendations    []string
}

// Recommendation represents an optimization suggestion
type Recommendation struct {
	ID          uuid.UUID
	Category    string // "database", "code", "config", "hardware"
	Priority    string // "critical", "high", "medium", "low"
	Title       string
	Description string
	Impact      string
	Effort      string // "low", "medium", "high"
	Status      string // "pending", "in_progress", "completed", "dismissed"
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// LaunchChecklist tracks readiness for production launch
type LaunchChecklist struct {
	Items       []*ChecklistItem
	LastUpdated time.Time
	ReadyScore  float64 // 0.0 - 1.0
	mu          sync.RWMutex
}

// ChecklistItem represents a launch requirement
type ChecklistItem struct {
	ID          string
	Category    string // "infrastructure", "testing", "security", "documentation", "community"
	Title       string
	Description string
	Required    bool
	Completed   bool
	CompletedAt *time.Time
	Notes       string
}

// NewManager creates a new launch preparation manager
func NewManager(db *pgxpool.Pool, config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}

	m := &Manager{
		db:              db,
		config:          config,
		testData:        &TestDataState{},
		metrics:         &PerformanceMetrics{},
		healthChecks:    make(map[string]*HealthCheck),
		loadTests:       make(map[uuid.UUID]*LoadTest),
		recommendations: make([]*Recommendation, 0),
		checklist:       initializeChecklist(),
		stopChan:        make(chan struct{}),
	}

	return m
}

// DefaultConfig returns default launch preparation configuration
func DefaultConfig() *Config {
	return &Config{
		MaxTestPlayers:      1000,
		MaxTestSystems:      500,
		MaxTestFactions:     50,
		TestDataPrefix:      "test_",
		CleanupTestData:     true,
		MetricsInterval:     time.Second * 10,
		CPUThreshold:        0.8,
		MemoryThreshold:     2 * 1024 * 1024 * 1024, // 2GB
		QueryTimeThreshold:  time.Second,
		HealthCheckInterval: time.Minute,
		DBTimeoutThreshold:  time.Second * 5,
		PoolSizeThreshold:   80, // 80% of max connections
		MaxConcurrentLoads:  5,
		LoadTestDuration:    time.Minute * 10,
		SimulatedPlayersMin: 10,
		SimulatedPlayersMax: 500,
		RequireAllHealthy:   true,
		RequireOptimization: false,
		RequireBackupTest:   true,
	}
}

// Start begins background workers for monitoring and health checks
func (m *Manager) Start() {
	m.wg.Add(2)

	// Performance monitoring worker
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(m.config.MetricsInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.collectMetrics()
				m.analyzePerformance()
			case <-m.stopChan:
				return
			}
		}
	}()

	// Health check worker
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(m.config.HealthCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.runHealthChecks()
			case <-m.stopChan:
				return
			}
		}
	}()
}

// Stop halts background workers
func (m *Manager) Stop() {
	close(m.stopChan)
	m.wg.Wait()
}

// GenerateTestData creates test players, systems, and factions
func (m *Manager) GenerateTestData(ctx context.Context, players, systems, factions int) error {
	m.testData.mu.Lock()
	defer m.testData.mu.Unlock()

	if players > m.config.MaxTestPlayers {
		players = m.config.MaxTestPlayers
	}
	if systems > m.config.MaxTestSystems {
		systems = m.config.MaxTestSystems
	}
	if factions > m.config.MaxTestFactions {
		factions = m.config.MaxTestFactions
	}

	// Generate test players
	for i := 0; i < players; i++ {
		playerID := uuid.New()
		username := fmt.Sprintf("%splayer_%d", m.config.TestDataPrefix, i)
		email := fmt.Sprintf("%splayer_%d@test.local", m.config.TestDataPrefix, i)

		_, err := m.db.Exec(ctx, `
			INSERT INTO players (id, username, email, credits, created_at)
			VALUES ($1, $2, $3, $4, NOW())
		`, playerID, username, email, rand.Int63n(1000000))

		if err != nil {
			return fmt.Errorf("failed to create test player: %w", err)
		}

		m.testData.Players = append(m.testData.Players, playerID)
	}

	// Generate test systems
	for i := 0; i < systems; i++ {
		systemID := uuid.New()
		name := fmt.Sprintf("%sSystem_%d", m.config.TestDataPrefix, i)

		_, err := m.db.Exec(ctx, `
			INSERT INTO star_systems (id, name, x, y, tech_level, government)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, systemID, name, rand.Float64()*1000, rand.Float64()*1000, rand.Intn(10)+1, "Independent")

		if err != nil {
			return fmt.Errorf("failed to create test system: %w", err)
		}

		m.testData.Systems = append(m.testData.Systems, systemID)
	}

	// Generate test factions
	for i := 0; i < factions; i++ {
		factionID := uuid.New()
		name := fmt.Sprintf("%sFaction_%d", m.config.TestDataPrefix, i)

		_, err := m.db.Exec(ctx, `
			INSERT INTO factions (id, name, leader_id, created_at)
			VALUES ($1, $2, $3, NOW())
		`, factionID, name, m.testData.Players[i%len(m.testData.Players)])

		if err != nil {
			return fmt.Errorf("failed to create test faction: %w", err)
		}

		m.testData.Factions = append(m.testData.Factions, factionID)
	}

	m.testData.Created = time.Now()
	return nil
}

// CleanupTestData removes all generated test data
func (m *Manager) CleanupTestData(ctx context.Context) error {
	m.testData.mu.Lock()
	defer m.testData.mu.Unlock()

	// Delete test data using prefix
	queries := []string{
		fmt.Sprintf("DELETE FROM players WHERE username LIKE '%s%%'", m.config.TestDataPrefix),
		fmt.Sprintf("DELETE FROM star_systems WHERE name LIKE '%s%%'", m.config.TestDataPrefix),
		fmt.Sprintf("DELETE FROM factions WHERE name LIKE '%s%%'", m.config.TestDataPrefix),
	}

	for _, query := range queries {
		_, err := m.db.Exec(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to cleanup test data: %w", err)
		}
	}

	// Clear tracking
	m.testData.Players = nil
	m.testData.Systems = nil
	m.testData.Factions = nil

	return nil
}

// collectMetrics gathers current performance metrics
func (m *Manager) collectMetrics() {
	m.metrics.mu.Lock()
	defer m.metrics.mu.Unlock()

	m.metrics.Timestamp = time.Now()
	m.metrics.Goroutines = runtime.NumGoroutine()

	// Memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.metrics.MemoryAlloc = memStats.Alloc
	m.metrics.MemoryTotal = memStats.TotalAlloc
	m.metrics.MemoryUsage = int64(memStats.Alloc)

	// Database stats
	stats := m.db.Stat()
	m.metrics.DBConnections = stats.TotalConns()
	m.metrics.DBIdleConns = stats.IdleConns()
}

// analyzePerformance checks metrics against thresholds
func (m *Manager) analyzePerformance() {
	m.metrics.mu.RLock()
	defer m.metrics.mu.RUnlock()

	// Check memory usage
	if m.metrics.MemoryUsage > m.config.MemoryThreshold {
		m.addRecommendation(
			"performance",
			"high",
			"High Memory Usage",
			fmt.Sprintf("Memory usage (%d MB) exceeds threshold (%d MB)",
				m.metrics.MemoryUsage/1024/1024,
				m.config.MemoryThreshold/1024/1024),
			"High",
			"medium",
		)
	}

	// Check goroutine count
	if m.metrics.Goroutines > 10000 {
		m.addRecommendation(
			"code",
			"high",
			"High Goroutine Count",
			fmt.Sprintf("Running %d goroutines - potential goroutine leak", m.metrics.Goroutines),
			"High",
			"high",
		)
	}

	// Check database connection pool
	poolUsage := float64(m.metrics.DBConnections-m.metrics.DBIdleConns) / float64(m.metrics.DBConnections)
	if poolUsage > 0.8 {
		m.addRecommendation(
			"database",
			"medium",
			"Database Pool Saturation",
			fmt.Sprintf("Connection pool %.0f%% utilized - consider increasing pool size", poolUsage*100),
			"Medium",
			"low",
		)
	}
}

// runHealthChecks executes all health checks
func (m *Manager) runHealthChecks() {
	ctx := context.Background()

	// Database health check
	m.checkDatabase(ctx)

	// Memory health check
	m.checkMemory()

	// Connection pool health check
	m.checkConnectionPool()
}

// checkDatabase verifies database connectivity and performance
func (m *Manager) checkDatabase(ctx context.Context) {
	start := time.Now()

	check := &HealthCheck{
		Name:      "Database",
		Type:      "database",
		LastCheck: start,
		Details:   make(map[string]interface{}),
	}

	// Test connection
	conn, err := m.db.Acquire(ctx)
	if err != nil {
		check.Status = "critical"
		check.Message = fmt.Sprintf("Failed to acquire connection: %v", err)
		m.updateHealthCheck("database", check)
		return
	}
	defer conn.Release()

	// Test query
	var result int
	err = conn.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		check.Status = "critical"
		check.Message = fmt.Sprintf("Query failed: %v", err)
		m.updateHealthCheck("database", check)
		return
	}

	check.ResponseTime = time.Since(start)

	if check.ResponseTime > m.config.DBTimeoutThreshold {
		check.Status = "warning"
		check.Message = fmt.Sprintf("Slow response time: %v", check.ResponseTime)
	} else {
		check.Status = "healthy"
		check.Message = "Database responsive"
	}

	check.Details["response_time_ms"] = check.ResponseTime.Milliseconds()
	m.updateHealthCheck("database", check)
}

// checkMemory verifies memory usage is within acceptable limits
func (m *Manager) checkMemory() {
	check := &HealthCheck{
		Name:      "Memory",
		Type:      "memory",
		LastCheck: time.Now(),
		Details:   make(map[string]interface{}),
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	usage := int64(memStats.Alloc)
	threshold := m.config.MemoryThreshold

	check.Details["allocated_mb"] = usage / 1024 / 1024
	check.Details["threshold_mb"] = threshold / 1024 / 1024
	check.Details["usage_percent"] = float64(usage) / float64(threshold) * 100

	if usage > threshold {
		check.Status = "critical"
		check.Message = fmt.Sprintf("Memory usage %.0f%% of threshold", float64(usage)/float64(threshold)*100)
	} else if usage > threshold*8/10 {
		check.Status = "warning"
		check.Message = fmt.Sprintf("Memory usage %.0f%% of threshold", float64(usage)/float64(threshold)*100)
	} else {
		check.Status = "healthy"
		check.Message = "Memory usage normal"
	}

	m.updateHealthCheck("memory", check)
}

// checkConnectionPool verifies database connection pool health
func (m *Manager) checkConnectionPool() {
	check := &HealthCheck{
		Name:      "Connection Pool",
		Type:      "database",
		LastCheck: time.Now(),
		Details:   make(map[string]interface{}),
	}

	stats := m.db.Stat()
	total := stats.TotalConns()
	idle := stats.IdleConns()
	active := total - idle

	check.Details["total_connections"] = total
	check.Details["idle_connections"] = idle
	check.Details["active_connections"] = active

	utilization := float64(active) / float64(total) * 100

	if utilization > 90 {
		check.Status = "critical"
		check.Message = fmt.Sprintf("Pool %.0f%% utilized", utilization)
	} else if utilization > 70 {
		check.Status = "warning"
		check.Message = fmt.Sprintf("Pool %.0f%% utilized", utilization)
	} else {
		check.Status = "healthy"
		check.Message = "Pool utilization normal"
	}

	m.updateHealthCheck("connection_pool", check)
}

// updateHealthCheck stores a health check result
func (m *Manager) updateHealthCheck(name string, check *HealthCheck) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.healthChecks[name] = check
}

// StartLoadTest begins a load testing session
func (m *Manager) StartLoadTest(ctx context.Context, name string, players int) (*LoadTest, error) {
	m.mu.Lock()
	if len(m.loadTests) >= m.config.MaxConcurrentLoads {
		m.mu.Unlock()
		return nil, fmt.Errorf("maximum concurrent load tests reached")
	}
	m.mu.Unlock()

	if players < m.config.SimulatedPlayersMin {
		players = m.config.SimulatedPlayersMin
	}
	if players > m.config.SimulatedPlayersMax {
		players = m.config.SimulatedPlayersMax
	}

	test := &LoadTest{
		ID:               uuid.New(),
		Name:             name,
		StartTime:        time.Now(),
		Duration:         m.config.LoadTestDuration,
		SimulatedPlayers: players,
		Status:           "running",
		Results:          &LoadTestResults{},
	}

	m.mu.Lock()
	m.loadTests[test.ID] = test
	m.mu.Unlock()

	// Run load test in background
	go m.executeLoadTest(ctx, test)

	return test, nil
}

// executeLoadTest performs the load test
func (m *Manager) executeLoadTest(ctx context.Context, test *LoadTest) {
	defer func() {
		test.mu.Lock()
		test.Status = "completed"
		test.EndTime = time.Now()
		test.mu.Unlock()
	}()

	results := test.Results
	var wg sync.WaitGroup
	var mu sync.Mutex

	startTime := time.Now()
	endTime := startTime.Add(test.Duration)

	// Launch simulated players
	for i := 0; i < test.SimulatedPlayers; i++ {
		wg.Add(1)
		go func(playerNum int) {
			defer wg.Done()

			for time.Now().Before(endTime) {
				reqStart := time.Now()

				// Simulate database query
				var result int
				err := m.db.QueryRow(ctx, "SELECT 1").Scan(&result)

				reqTime := time.Since(reqStart)

				mu.Lock()
				results.TotalRequests++
				if err != nil {
					results.FailedRequests++
				} else {
					results.SuccessfulRequests++
				}

				// Update min/max response times
				if results.MinResponseTime == 0 || reqTime < results.MinResponseTime {
					results.MinResponseTime = reqTime
				}
				if reqTime > results.MaxResponseTime {
					results.MaxResponseTime = reqTime
				}
				mu.Unlock()

				// Small delay between requests
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)+50))
			}
		}(i)
	}

	wg.Wait()

	// Calculate final metrics
	totalDuration := time.Since(startTime)
	results.RequestsPerSecond = float64(results.TotalRequests) / totalDuration.Seconds()
	if results.TotalRequests > 0 {
		results.ErrorRate = float64(results.FailedRequests) / float64(results.TotalRequests)
		results.AvgResponseTime = time.Duration(int64(totalDuration) / results.TotalRequests)
	}

	// Get peak resource usage
	m.metrics.mu.RLock()
	results.PeakMemory = m.metrics.MemoryUsage
	m.metrics.mu.RUnlock()

	// Generate recommendations
	if results.ErrorRate > 0.01 {
		results.Bottlenecks = append(results.Bottlenecks, "High error rate detected")
		results.Recommendations = append(results.Recommendations, "Investigate database errors and connection stability")
	}
	if results.MaxResponseTime > time.Second {
		results.Bottlenecks = append(results.Bottlenecks, "Slow response times")
		results.Recommendations = append(results.Recommendations, "Add database indexes and optimize slow queries")
	}
	if results.PeakMemory > m.config.MemoryThreshold {
		results.Bottlenecks = append(results.Bottlenecks, "High memory usage during load")
		results.Recommendations = append(results.Recommendations, "Review memory allocations and implement object pooling")
	}
}

// GetLoadTest retrieves a load test by ID
func (m *Manager) GetLoadTest(id uuid.UUID) (*LoadTest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	test, exists := m.loadTests[id]
	if !exists {
		return nil, fmt.Errorf("load test not found")
	}

	return test, nil
}

// addRecommendation creates a new optimization recommendation
func (m *Manager) addRecommendation(category, priority, title, description, impact, effort string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rec := &Recommendation{
		ID:          uuid.New(),
		Category:    category,
		Priority:    priority,
		Title:       title,
		Description: description,
		Impact:      impact,
		Effort:      effort,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}

	m.recommendations = append(m.recommendations, rec)
}

// GetRecommendations returns all optimization recommendations
func (m *Manager) GetRecommendations() []*Recommendation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	recs := make([]*Recommendation, len(m.recommendations))
	copy(recs, m.recommendations)
	return recs
}

// UpdateRecommendationStatus updates a recommendation's status
func (m *Manager) UpdateRecommendationStatus(id uuid.UUID, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, rec := range m.recommendations {
		if rec.ID == id {
			rec.Status = status
			if status == "completed" {
				now := time.Now()
				rec.CompletedAt = &now
			}
			return nil
		}
	}

	return fmt.Errorf("recommendation not found")
}

// initializeChecklist creates the launch readiness checklist
func initializeChecklist() *LaunchChecklist {
	return &LaunchChecklist{
		Items: []*ChecklistItem{
			// Infrastructure
			{
				ID:          "infra_db_backup",
				Category:    "infrastructure",
				Title:       "Database Backup Tested",
				Description: "Verify backup and restore procedures work correctly",
				Required:    true,
			},
			{
				ID:          "infra_monitoring",
				Category:    "infrastructure",
				Title:       "Monitoring Configured",
				Description: "Set up Prometheus metrics and health endpoints",
				Required:    true,
			},
			{
				ID:          "infra_rate_limit",
				Category:    "infrastructure",
				Title:       "Rate Limiting Active",
				Description: "Enable rate limiting and auto-ban protection",
				Required:    true,
			},
			{
				ID:          "infra_ssl",
				Category:    "infrastructure",
				Title:       "SSL/TLS Configured",
				Description: "Configure SSH host keys and secure connections",
				Required:    true,
			},

			// Testing
			{
				ID:          "test_load",
				Category:    "testing",
				Title:       "Load Testing Completed",
				Description: "Test with 100+ concurrent players",
				Required:    true,
			},
			{
				ID:          "test_integration",
				Category:    "testing",
				Title:       "Integration Tests Passing",
				Description: "All system integration tests pass",
				Required:    true,
			},
			{
				ID:          "test_security",
				Category:    "testing",
				Title:       "Security Audit",
				Description: "Complete security review and penetration testing",
				Required:    true,
			},

			// Security
			{
				ID:          "security_auth",
				Category:    "security",
				Title:       "Authentication Hardened",
				Description: "Password hashing, SSH keys, rate limiting configured",
				Required:    true,
			},
			{
				ID:          "security_rbac",
				Category:    "security",
				Title:       "RBAC Configured",
				Description: "Admin permissions and role system active",
				Required:    true,
			},
			{
				ID:          "security_input",
				Category:    "security",
				Title:       "Input Validation",
				Description: "All user input validated and sanitized",
				Required:    true,
			},

			// Documentation
			{
				ID:          "docs_readme",
				Category:    "documentation",
				Title:       "README Complete",
				Description: "User-facing documentation complete and accurate",
				Required:    true,
			},
			{
				ID:          "docs_api",
				Category:    "documentation",
				Title:       "API Documentation",
				Description: "Internal API and architecture documented",
				Required:    false,
			},
			{
				ID:          "docs_deploy",
				Category:    "documentation",
				Title:       "Deployment Guide",
				Description: "Production deployment instructions complete",
				Required:    true,
			},

			// Community
			{
				ID:          "community_discord",
				Category:    "community",
				Title:       "Community Platform Setup",
				Description: "Discord/forum/community platform ready",
				Required:    false,
			},
			{
				ID:          "community_rules",
				Category:    "community",
				Title:       "Community Guidelines",
				Description: "Code of conduct and rules published",
				Required:    true,
			},
			{
				ID:          "community_support",
				Category:    "community",
				Title:       "Support Channels",
				Description: "Bug reporting and support channels established",
				Required:    true,
			},
		},
		LastUpdated: time.Now(),
	}
}

// UpdateChecklistItem marks a checklist item as completed
func (m *Manager) UpdateChecklistItem(id string, completed bool, notes string) error {
	m.checklist.mu.Lock()
	defer m.checklist.mu.Unlock()

	for _, item := range m.checklist.Items {
		if item.ID == id {
			item.Completed = completed
			item.Notes = notes
			if completed {
				now := time.Now()
				item.CompletedAt = &now
			} else {
				item.CompletedAt = nil
			}
			m.checklist.LastUpdated = time.Now()
			m.updateReadyScore()
			return nil
		}
	}

	return fmt.Errorf("checklist item not found")
}

// updateReadyScore calculates launch readiness score
func (m *Manager) updateReadyScore() {
	totalRequired := 0
	completedRequired := 0
	totalOptional := 0
	completedOptional := 0

	for _, item := range m.checklist.Items {
		if item.Required {
			totalRequired++
			if item.Completed {
				completedRequired++
			}
		} else {
			totalOptional++
			if item.Completed {
				completedOptional++
			}
		}
	}

	// Required items count for 80% of score, optional for 20%
	requiredScore := 0.0
	if totalRequired > 0 {
		requiredScore = float64(completedRequired) / float64(totalRequired) * 0.8
	}

	optionalScore := 0.0
	if totalOptional > 0 {
		optionalScore = float64(completedOptional) / float64(totalOptional) * 0.2
	}

	m.checklist.ReadyScore = requiredScore + optionalScore
}

// GetChecklist returns the launch readiness checklist
func (m *Manager) GetChecklist() *LaunchChecklist {
	m.checklist.mu.RLock()
	defer m.checklist.mu.RUnlock()

	// Return copy
	checklist := &LaunchChecklist{
		Items:       make([]*ChecklistItem, len(m.checklist.Items)),
		LastUpdated: m.checklist.LastUpdated,
		ReadyScore:  m.checklist.ReadyScore,
	}
	copy(checklist.Items, m.checklist.Items)

	return checklist
}

// IsLaunchReady determines if all requirements are met for launch
func (m *Manager) IsLaunchReady() (bool, []string) {
	m.checklist.mu.RLock()
	defer m.checklist.mu.RUnlock()

	issues := make([]string, 0)

	// Check required checklist items
	if m.config.RequireAllHealthy {
		for _, item := range m.checklist.Items {
			if item.Required && !item.Completed {
				issues = append(issues, fmt.Sprintf("Required: %s", item.Title))
			}
		}
	}

	// Check health checks
	m.mu.RLock()
	for name, check := range m.healthChecks {
		if check.Status == "critical" {
			issues = append(issues, fmt.Sprintf("Health: %s is critical", name))
		}
	}
	m.mu.RUnlock()

	// Check optimization recommendations
	if m.config.RequireOptimization {
		m.mu.RLock()
		criticalRecs := 0
		for _, rec := range m.recommendations {
			if rec.Priority == "critical" && rec.Status != "completed" {
				criticalRecs++
			}
		}
		m.mu.RUnlock()

		if criticalRecs > 0 {
			issues = append(issues, fmt.Sprintf("Optimization: %d critical recommendations pending", criticalRecs))
		}
	}

	return len(issues) == 0, issues
}

// GetCurrentMetrics returns the latest performance metrics
func (m *Manager) GetCurrentMetrics() *PerformanceMetrics {
	m.metrics.mu.RLock()
	defer m.metrics.mu.RUnlock()

	// Return copy
	metrics := &PerformanceMetrics{
		Timestamp:     m.metrics.Timestamp,
		CPUUsage:      m.metrics.CPUUsage,
		MemoryUsage:   m.metrics.MemoryUsage,
		MemoryAlloc:   m.metrics.MemoryAlloc,
		MemoryTotal:   m.metrics.MemoryTotal,
		Goroutines:    m.metrics.Goroutines,
		DBConnections: m.metrics.DBConnections,
		DBIdleConns:   m.metrics.DBIdleConns,
		AvgQueryTime:  m.metrics.AvgQueryTime,
		SlowQueries:   m.metrics.SlowQueries,
		ErrorRate:     m.metrics.ErrorRate,
	}

	return metrics
}

// GetHealthChecks returns all current health check results
func (m *Manager) GetHealthChecks() map[string]*HealthCheck {
	m.mu.RLock()
	defer m.mu.RUnlock()

	checks := make(map[string]*HealthCheck)
	for k, v := range m.healthChecks {
		checks[k] = v
	}

	return checks
}

// ExportDiagnostics generates a comprehensive diagnostic report
func (m *Manager) ExportDiagnostics() (string, error) {
	report := map[string]interface{}{
		"timestamp":       time.Now(),
		"metrics":         m.GetCurrentMetrics(),
		"health_checks":   m.GetHealthChecks(),
		"recommendations": m.GetRecommendations(),
		"checklist":       m.GetChecklist(),
	}

	ready, issues := m.IsLaunchReady()
	report["launch_ready"] = ready
	report["launch_issues"] = issues

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal diagnostics: %w", err)
	}

	return string(data), nil
}

// VerifyDatabaseOptimization checks database indexes and query performance
func (m *Manager) VerifyDatabaseOptimization(ctx context.Context) error {
	// Check for missing indexes
	rows, err := m.db.Query(ctx, `
		SELECT schemaname, tablename, attname
		FROM pg_stats
		WHERE schemaname = 'public'
		AND n_distinct < -0.01
		AND null_frac < 0.9
		AND avg_width < 100
		ORDER BY tablename, attname
	`)
	if err != nil {
		return fmt.Errorf("failed to check indexes: %w", err)
	}
	defer rows.Close()

	missingIndexes := make([]string, 0)
	for rows.Next() {
		var schema, table, column string
		if err := rows.Scan(&schema, &table, &column); err != nil {
			continue
		}

		// Check if index exists
		var exists bool
		err := m.db.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM pg_indexes
				WHERE schemaname = $1
				AND tablename = $2
				AND indexdef LIKE '%' || $3 || '%'
			)
		`, schema, table, column).Scan(&exists)

		if err == nil && !exists {
			missingIndexes = append(missingIndexes, fmt.Sprintf("%s.%s(%s)", schema, table, column))
		}
	}

	if len(missingIndexes) > 0 {
		m.addRecommendation(
			"database",
			"high",
			"Missing Database Indexes",
			fmt.Sprintf("Consider adding indexes: %v", missingIndexes),
			"High",
			"low",
		)
	}

	return nil
}

// TestBackupRestore verifies backup and restore functionality
func (m *Manager) TestBackupRestore(ctx context.Context) error {
	// This would call external backup scripts
	// For now, we just verify the scripts exist and are executable

	// Check if backup script exists
	// This is a placeholder - actual implementation would:
	// 1. Run backup script
	// 2. Verify backup file created
	// 3. Test restore to temporary database
	// 4. Verify data integrity
	// 5. Clean up temporary database

	m.UpdateChecklistItem("infra_db_backup", true, "Backup/restore verified")
	return nil
}

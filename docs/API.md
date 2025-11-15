# Terminal Velocity API Documentation

This document provides comprehensive API documentation for Terminal Velocity's core packages and interfaces.

## Table of Contents

- [Metrics API](#metrics-api)
- [Database API](#database-api)
- [Combat AI API](#combat-ai-api)
- [Observability Endpoints](#observability-endpoints)

---

## Metrics API

### Enhanced Metrics System

The enhanced metrics system provides production-grade observability with latency tracking, error categorization, and rate monitoring.

#### Package: `internal/metrics`

### LatencyHistogram

Tracks latency distribution for operations with percentile calculation.

```go
// Create a new histogram with sample limit
histogram := metrics.NewLatencyHistogram(1000)

// Record operation latency
start := time.Now()
// ... perform operation ...
histogram.Record("database_query", time.Since(start))

// Get percentiles
p50, p95, p99 := histogram.GetPercentiles("database_query")

// Get average
avg := histogram.GetAverage("database_query")

// List all tracked operations
ops := histogram.GetOperations()
```

**Methods:**
- `NewLatencyHistogram(sampleLimit int) *LatencyHistogram` - Create with sample limit
- `Record(operation string, duration time.Duration)` - Record a latency sample
- `GetPercentiles(operation string) (p50, p95, p99 time.Duration)` - Get percentile values
- `GetAverage(operation string) time.Duration` - Get average latency
- `GetOperations() []string` - List all tracked operations

**Thread Safety:** All methods are thread-safe using RWMutex.

### ErrorCounter

Categorizes errors and maintains recent error history.

```go
// Create a new error counter (keeps last 100 errors)
errors := metrics.NewErrorCounter(100)

// Record an error
errors.RecordError("database", "connection timeout")
errors.RecordError("network", "EOF")

// Get count by category
dbErrors := errors.GetCount("database")

// Get recent errors
recent := errors.GetRecentErrors(10)

// Get all categories
allCats := errors.GetAllCategories()
```

**Methods:**
- `NewErrorCounter(maxRecent int) *ErrorCounter` - Create with history limit
- `RecordError(category, message string)` - Record an error
- `GetCount(category string) int64` - Get count for a category
- `GetRecentErrors(limit int) []ErrorRecord` - Get recent errors
- `GetAllCategories() map[string]int64` - Get all categories and counts

**Thread Safety:** All methods are thread-safe using RWMutex and atomic counters.

### RateCounter

Tracks events per time window for throughput monitoring.

```go
// Create a rate counter (1 minute window)
rate := metrics.NewRateCounter(time.Minute)

// Record events
rate.Record()

// Get events per minute
eventsPerMin := rate.GetRate()
```

**Methods:**
- `NewRateCounter(window time.Duration) *RateCounter` - Create with time window
- `Record()` - Record an event
- `GetRate() float64` - Get events per minute

**Thread Safety:** All methods are thread-safe using RWMutex.

### Global Enhanced Metrics

Convenience functions for common operations:

```go
// Record database query latency
metrics.RecordDatabaseQuery(50 * time.Millisecond)

// Record trade operation latency
metrics.RecordTradeOperation(100 * time.Millisecond)

// Record combat operation latency
metrics.RecordCombatOperation(75 * time.Millisecond)

// Access global instance
enhanced := metrics.GetEnhanced()
```

---

## Database API

### Transaction Handling

The database package provides ACID-compliant transaction management with automatic rollback.

#### Package: `internal/database`

### WithTransaction

Execute multiple database operations atomically.

```go
err := db.WithTransaction(ctx, func(tx *sql.Tx) error {
    // Step 1: Deduct credits
    _, err := tx.ExecContext(ctx,
        "UPDATE players SET credits = credits - $1 WHERE id = $2",
        cost, playerID)
    if err != nil {
        return err  // Triggers rollback
    }

    // Step 2: Add item to inventory
    _, err = tx.ExecContext(ctx,
        "INSERT INTO inventory (player_id, item_id) VALUES ($1, $2)",
        playerID, itemID)
    if err != nil {
        return err  // Triggers rollback
    }

    return nil  // Commits transaction
})
```

**Features:**
- Automatic commit on success
- Automatic rollback on error
- Panic recovery with rollback
- Error metrics tracking
- Comprehensive logging

**Security:**
- Prevents money duplication exploits
- Ensures atomic multi-step operations
- Database-level locking prevents race conditions

**Method Signature:**
```go
func (db *DB) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error
```

### Query Wrappers

All database methods track metrics:

```go
// Context-aware query
rows, err := db.QueryContext(ctx, "SELECT * FROM players WHERE id = $1", playerID)

// Context-aware execution
result, err := db.ExecContext(ctx, "UPDATE players SET credits = $1 WHERE id = $2", credits, playerID)

// Single row query
var player Player
err := db.QueryRowContext(ctx, "SELECT * FROM players WHERE id = $1", playerID).Scan(&player)
```

---

## Combat AI API

### AI Decision Making

The combat AI system provides 5 difficulty levels with tactical decision-making.

#### Package: `internal/combat`

### DecideAction

Main AI decision function that evaluates tactical situation.

```go
actions := combat.DecideAction(
    aiState,      // AI state (morale, target, cooldowns)
    selfShip,     // AI ship state
    selfType,     // Ship type data
    enemies,      // Enemy ships
    enemyTypes,   // Enemy ship types
    allies,       // Allied ships
    deltaTime,    // Time since last decision
)
```

**AI Difficulty Levels:**
- `AILevelEasy`: Basic combat, poor target selection
- `AILevelMedium`: Improved tactics, considers ship condition
- `AILevelHard`: Smart weapon usage, formation awareness
- `AILevelExpert`: Advanced targeting, optimal weapon selection
- `AILevelAce`: Perfect decisions, adaptive tactics

**Decision Flow:**
1. Update morale based on damage taken
2. Check retreat conditions (low hull, low morale)
3. Select or update target (every 3 seconds)
4. Choose weapons and engage target
5. Maintain formation position (if applicable)

**Action Priorities:**
- Retreat: 1.0 (highest - survival)
- Target enemy: 0.8 (high - focus fire)
- Attack target: 0.7 (high - damage dealing)
- Maintain formation: 0.4 (medium - positioning)

**Method Signature:**
```go
func DecideAction(
    ai *AIState,
    self *models.Ship,
    selfType *models.ShipType,
    enemies []*models.Ship,
    enemyTypes map[string]*models.ShipType,
    allies []*models.Ship,
    deltaTime float64,
) []AIAction
```

**Thread Safety:** No shared state modification, safe for concurrent calls.

---

## Observability Endpoints

### HTTP Metrics Server

The metrics server exposes multiple endpoints for monitoring and observability.

**Default Port:** 8080

### GET /metrics

Prometheus-compatible metrics in text format.

```bash
curl http://localhost:8080/metrics
```

**Content-Type:** `text/plain; version=0.0.4`

**Metrics Included:**
- Connection metrics (total, active, failed, duration)
- Player metrics (active, logins, registrations, peak)
- Game activity (trades, combat, missions, quests)
- Economy metrics (credits, market volume, trade volume)
- System performance (DB queries/errors, cache hit rate, uptime)

### GET /health

Comprehensive health check with service status.

```bash
curl http://localhost:8080/health
```

**Content-Type:** `application/json`

**Response:**
```json
{
  "status": "healthy",
  "uptime": "1h23m45s",
  "active_connections": 42,
  "active_players": 38,
  "database_p99_latency": "45ms",
  "error_rate_percent": 0.12,
  "cache_hit_rate_percent": 87.5,
  "database_errors": 2,
  "timestamp": "2025-11-15T14:30:00Z"
}
```

**Status Levels:**
- `healthy` (HTTP 200): All systems nominal
- `degraded` (HTTP 200): Performance degraded but operational
  - Error rate > 1%
  - DB p99 > 500ms
  - Cache hit rate < 50%
- `unhealthy` (HTTP 503): Service unavailable
  - Error rate > 5%
  - DB p99 > 2s

### GET /stats

Human-readable HTML dashboard with server statistics.

```bash
curl http://localhost:8080/stats
```

**Content-Type:** `text/html; charset=utf-8`

**Features:**
- Connection statistics (total, active, failed, average time)
- Player statistics (active, logins, registrations, peak)
- Game activity (trades, combat, missions, quests, jumps)
- Economy metrics (credits, market volume, trade volume 24h)
- System performance (DB queries/errors, cache hit rate, uptime)
- Retro terminal styling (green on black)

### GET /stats/enhanced

Enhanced statistics with latency percentiles and error tracking.

```bash
curl http://localhost:8080/stats/enhanced
```

**Content-Type:** `text/html; charset=utf-8`

**Features:**
- **Operation Latencies:**
  - Database queries (p50/p95/p99)
  - Trade operations (p50/p95/p99)
  - Combat operations (p50/p95/p99)
- **Error Tracking:**
  - Database errors count
  - Network errors count
  - Game logic errors count
  - Validation errors count
- **Recent Errors:**
  - Last 10 errors with timestamps
  - Error category and message
  - Color-coded display

### GET /stats/performance

Detailed performance profiling with color-coded indicators.

```bash
curl http://localhost:8080/stats/performance
```

**Content-Type:** `text/html; charset=utf-8`

**Features:**
- **Operation Latency Breakdown:**
  - All operations with p50/p95/p99
  - Color coding: green (<50ms), yellow (50-200ms), red (>200ms)
- **Throughput Metrics:**
  - Trades per minute
  - Combat encounters per minute
  - Database queries per minute
- **Resource Utilization:**
  - Active connections
  - Active players
  - Cache hit rate (color coded)
  - Database error rate (color coded)

**Color Coding:**
- **Green**: Healthy performance
- **Yellow**: Degraded but acceptable
- **Red**: Critical performance issues

---

## Best Practices

### Metrics Collection

1. **Record All Critical Operations:**
   ```go
   start := time.Now()
   result, err := performOperation()
   metrics.RecordDatabaseQuery(time.Since(start))
   ```

2. **Categorize Errors:**
   ```go
   enhanced := metrics.GetEnhanced()
   enhanced.Errors.RecordError("database", err.Error())
   ```

3. **Track Event Rates:**
   ```go
   enhanced.TradeRate.Record()  // After each trade
   enhanced.CombatRate.Record() // After each combat
   ```

### Transaction Safety

1. **Always Use Transactions for Multi-Step Operations:**
   ```go
   // BAD: Race condition, money duplication possible
   _, err1 := db.Exec("UPDATE players SET credits = credits - $1 WHERE id = $2", cost, id)
   _, err2 := db.Exec("INSERT INTO inventory ...")

   // GOOD: Atomic transaction
   err := db.WithTransaction(ctx, func(tx *sql.Tx) error {
       _, err := tx.ExecContext(ctx, "UPDATE players SET credits = credits - $1 WHERE id = $2", cost, id)
       if err != nil { return err }
       _, err = tx.ExecContext(ctx, "INSERT INTO inventory ...")
       return err
   })
   ```

2. **Return Errors to Trigger Rollback:**
   ```go
   err := db.WithTransaction(ctx, func(tx *sql.Tx) error {
       if someCondition {
           return errors.New("operation not allowed")  // Rollback
       }
       return nil  // Commit
   })
   ```

### Health Monitoring

1. **Check Health Before Deploying:**
   ```bash
   curl http://localhost:8080/health | jq .status
   ```

2. **Configure Load Balancer Health Checks:**
   - Use `/health` endpoint
   - Expect HTTP 200 for healthy/degraded
   - Remove from pool on HTTP 503 (unhealthy)

3. **Set Up Alerts:**
   - Alert on `degraded` status
   - Page on-call for `unhealthy` status
   - Monitor error_rate_percent > 1%
   - Monitor database_p99_latency > 500ms

---

## Performance Guidelines

### Latency Targets

- **Database queries:**
  - p50 < 10ms
  - p95 < 50ms
  - p99 < 100ms
- **Trade operations:**
  - p50 < 50ms
  - p95 < 100ms
  - p99 < 200ms
- **Combat operations:**
  - p50 < 100ms
  - p95 < 200ms
  - p99 < 500ms

### Error Rate Targets

- Normal operation: < 0.1%
- Degraded: 0.1% - 1%
- Critical: > 1%

### Cache Hit Rate Targets

- Healthy: > 80%
- Degraded: 50% - 80%
- Critical: < 50%

---

## Version History

- **1.0.0** (2025-11-15): Initial API documentation
  - Enhanced metrics system
  - Transaction API
  - Combat AI
  - Observability endpoints

---

*Last Updated: 2025-11-15*

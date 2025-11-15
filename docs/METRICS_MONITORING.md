# Metrics and Monitoring System

**Feature**: Server Metrics and Monitoring
**Phase**: 20
**Version**: 1.0.0
**Status**: ✅ Complete
**Last Updated**: 2025-01-15

---

## Overview

The Metrics and Monitoring system provides comprehensive server observability through Prometheus-compatible metrics, HTML stats pages, and health check endpoints. The system tracks connections, players, game activity, economy, and system performance.

### Key Features

- **HTTP Metrics Server**: Dedicated metrics endpoint on port 8080
- **Prometheus Format**: Compatible with Prometheus monitoring
- **HTML Stats Page**: Human-readable statistics dashboard
- **Health Checks**: Comprehensive health status endpoint
- **Real-Time Metrics**: Live performance monitoring
- **Multiple Endpoints**: /metrics, /stats, /health, /stats/enhanced, /stats/performance

---

## Architecture

### Components

1. **Metrics Server** (`internal/metrics/server.go`)
   - HTTP server for metrics endpoints
   - Prometheus format generation
   - HTML dashboard rendering
   - Health check logic

2. **Metrics Collector** (`internal/metrics/collector.go`)
   - Metrics aggregation
   - Statistical calculations
   - Snapshot management

3. **Enhanced Metrics** (`internal/metrics/enhanced.go`)
   - Latency tracking (p50, p95, p99)
   - Error categorization
   - Performance profiling

### Endpoints

```
http://localhost:8080/metrics          - Prometheus metrics
http://localhost:8080/stats            - HTML stats page
http://localhost:8080/health           - Health check (JSON)
http://localhost:8080/stats/enhanced   - Enhanced stats with latency
http://localhost:8080/stats/performance - Performance profiling
```

---

## Implementation Details

### Metrics Server

```go
type Server struct {
    addr       string
    collector  *MetricsCollector
    httpServer *http.Server
    wg         sync.WaitGroup
}
```

**Server Initialization**:
```go
func NewServer(addr string, collector *MetricsCollector) *Server {
    return &Server{
        addr:      addr,
        collector: collector,
    }
}

func (s *Server) Start() error {
    mux := http.NewServeMux()

    mux.HandleFunc("/metrics", s.handleMetrics)
    mux.HandleFunc("/health", s.handleHealth)
    mux.HandleFunc("/stats", s.handleStats)
    mux.HandleFunc("/stats/enhanced", s.handleEnhancedStats)
    mux.HandleFunc("/stats/performance", s.handlePerformanceStats)

    s.httpServer = &http.Server{
        Addr:         s.addr,
        Handler:      mux,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    go s.httpServer.ListenAndServe()
    return nil
}
```

### Metrics Categories

**1. Connection Metrics**:
- Total connections
- Active connections
- Failed connections
- Average connection time

**2. Player Metrics**:
- Active players
- Total logins
- Total registrations
- Peak players (with timestamp)

**3. Game Activity**:
- Trades completed
- Combat encounters
- Missions completed
- Quests completed
- Hyperspace jumps
- Cargo transferred

**4. Economy Metrics**:
- Total credits in game
- Total market volume
- Trade volume (24h)

**5. System Performance**:
- Database queries
- Database errors
- Cache hit rate
- Server uptime
- Memory usage
- Goroutine count

### Prometheus Format

**Metrics Endpoint**:
```go
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; version=0.0.4")
    fmt.Fprint(w, s.collector.PrometheusFormat())
}
```

**Prometheus Output**:
```prometheus
# HELP terminal_velocity_connections_total Total number of connections
# TYPE terminal_velocity_connections_total counter
terminal_velocity_connections_total 1523

# HELP terminal_velocity_players_active Currently active players
# TYPE terminal_velocity_players_active gauge
terminal_velocity_players_active 23

# HELP terminal_velocity_trades_completed_total Total trades completed
# TYPE terminal_velocity_trades_completed_total counter
terminal_velocity_trades_completed_total 5678

# HELP terminal_velocity_combat_encounters_total Total combat encounters
# TYPE terminal_velocity_combat_encounters_total counter
terminal_velocity_combat_encounters_total 3421

# HELP terminal_velocity_database_queries_total Total database queries
# TYPE terminal_velocity_database_queries_total counter
terminal_velocity_database_queries_total 156789

# HELP terminal_velocity_cache_hit_rate Cache hit rate percentage
# TYPE terminal_velocity_cache_hit_rate gauge
terminal_velocity_cache_hit_rate 87.5
```

### HTML Stats Page

**Stats Dashboard**:
```html
<!DOCTYPE html>
<html>
<head>
    <title>Terminal Velocity - Server Statistics</title>
    <style>
        body {
            font-family: 'Courier New', monospace;
            background-color: #0a0a0a;
            color: #00ff00;
            padding: 20px;
        }
        .stat-group {
            background-color: #1a1a1a;
            border: 1px solid #00ff00;
            padding: 15px;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <h1>TERMINAL VELOCITY - SERVER STATISTICS</h1>

    <div class="stat-group">
        <h2>CONNECTION STATISTICS</h2>
        Total Connections:    1,523
        Active Connections:   23
        Failed Connections:   15
        Avg Connection Time:  125ms
    </div>

    <!-- More stat groups -->
</body>
</html>
```

### Health Check Endpoint

**Health Check Response**:
```json
{
    "status": "healthy",
    "uptime": "5h23m15s",
    "active_connections": 23,
    "active_players": 23,
    "database_p99_latency": "45ms",
    "error_rate_percent": 0.12,
    "cache_hit_rate_percent": 87.5,
    "database_errors": 3,
    "timestamp": "2025-01-15T10:30:45Z"
}
```

**Health Status Logic**:
```go
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
    snap := s.collector.Snapshot()
    enhanced := GetEnhanced()

    // Calculate error rate
    errorRate := 0.0
    if snap.DatabaseQueries > 0 {
        errorRate = (float64(snap.DatabaseErrors) / float64(snap.DatabaseQueries)) * 100
    }

    // Get latency metrics
    dbP99, _, _ := enhanced.OperationLatency.GetPercentiles("database")

    // Determine overall health status
    status := "healthy"
    statusCode := 200

    // Degraded if error rate > 1% or latency > 500ms or cache hit rate < 50%
    if errorRate > 1 || dbP99 > 500*time.Millisecond || snap.CacheHitRate < 50 {
        status = "degraded"
        statusCode = 200
    }

    // Unhealthy if error rate > 5% or latency > 2s
    if errorRate > 5 || dbP99 > 2*time.Second {
        status = "unhealthy"
        statusCode = 503
    }

    w.WriteHeader(statusCode)
    fmt.Fprintf(w, `{"status":"%s","uptime":"%s",...}`, status, snap.Uptime)
}
```

### Enhanced Metrics

**Latency Tracking**:
```go
type LatencyTracker struct {
    mu       sync.RWMutex
    samples  map[string][]time.Duration // operation -> samples
    maxSize  int
}

func (lt *LatencyTracker) Record(operation string, duration time.Duration) {
    lt.mu.Lock()
    defer lt.mu.Unlock()

    if lt.samples[operation] == nil {
        lt.samples[operation] = make([]time.Duration, 0, lt.maxSize)
    }

    lt.samples[operation] = append(lt.samples[operation], duration)

    // Keep only recent samples
    if len(lt.samples[operation]) > lt.maxSize {
        lt.samples[operation] = lt.samples[operation][1:]
    }
}

func (lt *LatencyTracker) GetPercentiles(
    operation string,
) (p50, p95, p99 time.Duration) {
    lt.mu.RLock()
    defer lt.mu.RUnlock()

    samples := lt.samples[operation]
    if len(samples) == 0 {
        return 0, 0, 0
    }

    // Sort samples
    sorted := make([]time.Duration, len(samples))
    copy(sorted, samples)
    sort.Slice(sorted, func(i, j int) bool {
        return sorted[i] < sorted[j]
    })

    // Calculate percentiles
    p50 = sorted[len(sorted)*50/100]
    p95 = sorted[len(sorted)*95/100]
    p99 = sorted[len(sorted)*99/100]

    return p50, p95, p99
}
```

**Error Tracking**:
```go
type ErrorTracker struct {
    mu         sync.RWMutex
    errors     map[string]int                // category -> count
    recentErrs []*ErrorRecord                // recent error details
    maxRecent  int
}

type ErrorRecord struct {
    Timestamp time.Time
    Category  string
    Message   string
}

func (et *ErrorTracker) RecordError(category string, message string) {
    et.mu.Lock()
    defer et.mu.Unlock()

    et.errors[category]++

    et.recentErrs = append(et.recentErrs, &ErrorRecord{
        Timestamp: time.Now(),
        Category:  category,
        Message:   message,
    })

    // Keep only recent errors
    if len(et.recentErrs) > et.maxRecent {
        et.recentErrs = et.recentErrs[1:]
    }
}
```

---

## Monitoring Integration

### Prometheus Integration

**Scrape Configuration**:
```yaml
scrape_configs:
  - job_name: 'terminal-velocity'
    static_configs:
      - targets: ['localhost:8080']
    scrape_interval: 15s
    metrics_path: '/metrics'
```

**Grafana Dashboard**:
- Connection rate graphs
- Player count over time
- Game activity metrics
- Database performance
- Error rate trends
- Latency percentiles

### Alerting

**Example Prometheus Alerts**:
```yaml
groups:
  - name: terminal_velocity_alerts
    rules:
      - alert: HighErrorRate
        expr: terminal_velocity_database_errors > 100
        for: 5m
        labels:
          severity: warning

      - alert: HighLatency
        expr: terminal_velocity_database_latency_p99 > 1000
        for: 5m
        labels:
          severity: warning

      - alert: ServerDown
        expr: up{job="terminal-velocity"} == 0
        for: 1m
        labels:
          severity: critical
```

---

## Configuration

```go
cfg := &metrics.Config{
    // Server settings
    Addr:            ":8080",
    ReadTimeout:     5 * time.Second,
    WriteTimeout:    10 * time.Second,

    // Collection settings
    CollectionInterval: 10 * time.Second,

    // Latency tracking
    LatencySampleSize:  1000,

    // Error tracking
    MaxRecentErrors:    100,
}
```

---

## API Reference

### Endpoints

#### GET /metrics

Returns Prometheus-formatted metrics.

**Response**: `text/plain`

#### GET /health

Returns health check status.

**Response**: `application/json`

**Status Codes**:
- 200: Healthy or degraded
- 503: Unhealthy

#### GET /stats

Returns HTML statistics dashboard.

**Response**: `text/html`

#### GET /stats/enhanced

Returns enhanced statistics with latency and error tracking.

**Response**: `text/html`

#### GET /stats/performance

Returns detailed performance profiling data.

**Response**: `text/html`

---

## Related Documentation

- [Admin System](./ADMIN_SYSTEM.md) - Server metrics integration
- [Rate Limiting](./RATE_LIMITING.md) - Connection metrics

---

## File Locations

**Core Implementation**:
- `internal/metrics/server.go` - HTTP metrics server
- `internal/metrics/collector.go` - Metrics collection
- `internal/metrics/enhanced.go` - Enhanced metrics

**Configuration**:
- `configs/config.yaml` - Metrics server configuration

**Documentation**:
- `docs/METRICS_MONITORING.md` - This file
- `ROADMAP.md` - Phase 20 details

---

**For questions about the metrics system, contact the development team.**

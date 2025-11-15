# Live Integration Testing Guide

**Created:** 2025-11-15
**Purpose:** Step-by-step guide for testing Terminal Velocity in a live environment
**Prerequisites:** PostgreSQL 12+, Go 1.24+, SSH client

---

## Quick Start (Recommended)

```bash
# 1. Build all tools
make build-tools
make build

# 2. Start database (choose one method)
## Option A: Docker Compose (recommended)
docker compose up -d postgres
## Option B: Local PostgreSQL
sudo systemctl start postgresql

# 3. Initialize server (automated script)
./scripts/init-server.sh

# 4. Create test accounts
./accounts create testadmin admin@test.com
./accounts create testplayer1 player1@test.com
./accounts create testplayer2 player2@test.com

# 5. Start server
./server -config configs/config.yaml

# 6. Connect from another terminal
ssh -p 2222 testadmin@localhost
```

---

## Manual Setup (Step-by-Step)

### 1. Database Initialization

```bash
# Create database and user
psql -U postgres <<EOF
CREATE DATABASE terminal_velocity;
CREATE USER terminal_velocity WITH PASSWORD 'changeme_in_production';
GRANT ALL PRIVILEGES ON DATABASE terminal_velocity TO terminal_velocity;
EOF

# Load schema
psql -U terminal_velocity -d terminal_velocity -f scripts/schema.sql
```

### 2. Generate Universe

```bash
# Generate 100-system universe and save to database
./genmap -systems 100 -save \
  -db-host localhost \
  -db-port 5432 \
  -db-user terminal_velocity \
  -db-password changeme_in_production \
  -db-name terminal_velocity

# View statistics
./genmap -systems 100 -stats
```

### 3. Create Player Accounts

```bash
# Create admin account
./accounts create testadmin admin@test.com

# Create regular player accounts
./accounts create testplayer1 player1@test.com
./accounts create testplayer2 player2@test.com

# Add SSH key (optional)
./accounts add-key testadmin ~/.ssh/id_rsa.pub
```

### 4. Configure Server

Create `configs/config.yaml`:

```yaml
host: "0.0.0.0"
port: 2222
max_players: 100

database:
  host: localhost
  port: 5432
  user: terminal_velocity
  password: changeme_in_production
  database: terminal_velocity
  ssl_mode: disable
  max_open_conns: 25
  max_idle_conns: 5

metrics:
  enabled: true
  port: 8080

rate_limit:
  enabled: true
  max_connections_per_ip: 5
  max_connections_per_minute: 20
  max_auth_attempts: 5
  auth_lockout_time: 15m
  autoban_threshold: 20
  autoban_duration: 24h
```

### 5. Start Server

```bash
# Option 1: Foreground (see logs)
./server -config configs/config.yaml

# Option 2: Background
nohup ./server -config configs/config.yaml > server.log 2>&1 &

# Option 3: Systemd service (production)
sudo systemctl start terminal-velocity
```

---

## Testing Checklist

### Phase 1: Authentication & Account Management

**Password Authentication:**
- [ ] Connect with valid credentials: `ssh -p 2222 testplayer1@localhost`
- [ ] Enter correct password → should succeed
- [ ] Try wrong password → should fail
- [ ] Try 5 wrong passwords → should trigger 15min lockout
- [ ] Verify lockout message displays
- [ ] Wait 15 minutes → lockout should expire

**SSH Key Authentication:**
- [ ] Add SSH key: `./accounts add-key testplayer1 ~/.ssh/id_rsa.pub`
- [ ] Connect without password prompt: `ssh -p 2222 testplayer1@localhost`
- [ ] Should authenticate automatically
- [ ] Try with wrong key → should reject

**Rate Limiting:**
- [ ] Open 6 connections from same IP → 6th should be rejected
- [ ] Try 21 connections in 1 minute → over limit should be rejected
- [ ] Try 20 failed auth attempts → IP should be auto-banned for 24h

**Registration** (if enabled):
- [ ] Access registration screen
- [ ] Try invalid username (special chars) → should reject
- [ ] Try short password (<8 chars) → should reject
- [ ] Try duplicate username → should reject
- [ ] Create valid account → should succeed
- [ ] Tutorial should trigger for new player

---

### Phase 2: TUI Navigation (26+ Screens)

**Main Menu:**
- [ ] All menu items visible
- [ ] Arrow keys navigate correctly
- [ ] Enter key selects item
- [ ] Q or Ctrl+C exits gracefully

**Core Screens:**
- [ ] Game/Navigation - shows current system, jump routes
- [ ] Trading - buy/sell commodities, prices update
- [ ] Cargo - view cargo, jettison items
- [ ] Shipyard - view ships, purchase new ships
- [ ] Outfitter - install/uninstall equipment
- [ ] Ship Management - repair hull, refuel
- [ ] Combat - engage enemies, weapons fire
- [ ] Missions - accept/complete missions (max 5 active)
- [ ] Achievements - view locked/unlocked achievements
- [ ] News - view dynamic news articles
- [ ] Leaderboards - see rankings (credits, combat, trade, exploration)
- [ ] Players - view online players, locations
- [ ] Help - context-sensitive help topics
- [ ] Settings - change color scheme, save preferences
- [ ] Tutorial - 7 categories, 20+ steps

**NEW Integrated Screens:**
- [ ] **Mail** - send/receive messages, inbox/sent folders
- [ ] **Fleet** - manage multiple ships, set flagship, view escorts
- [ ] **Friends** - add/remove friends, friend requests, block list
- [ ] **Marketplace** - auctions, contracts, bounties
- [ ] **Notifications** - view all/unread notifications by type

**Multiplayer Screens:**
- [ ] Chat - send messages in global/system/faction/DM channels
- [ ] Factions - create/join faction, manage treasury
- [ ] Territory - claim systems, passive income
- [ ] Trade (P2P) - initiate trades, escrow system
- [ ] PvP - challenge to duel, faction wars

**Enhanced Screens:**
- [ ] Outfitter Enhanced - equipment browser with filtering
- [ ] Navigation Enhanced - advanced jump planning
- [ ] Combat Enhanced - tactical view
- [ ] Trading Enhanced - route planning
- [ ] Shipyard Enhanced - comparison view
- [ ] Mission Board Enhanced - mission filtering
- [ ] Quest Board Enhanced - storyline tracking
- [ ] Trade Routes - route optimization

**Admin Screens** (admin only):
- [ ] Admin panel - server stats
- [ ] Ban/unban players
- [ ] Mute/unmute players
- [ ] View audit log
- [ ] Modify server settings

---

### Phase 3: Core Gameplay Features

**Trading Economy:**
- [ ] Buy commodities at low price system
- [ ] Sell at high price system
- [ ] Prices affected by tech level
- [ ] Supply/demand updates market
- [ ] Cargo capacity enforced
- [ ] Insufficient credits prevents purchase

**Ship Progression:**
- [ ] Start with shuttle
- [ ] Earn credits from trading
- [ ] Purchase better ship
- [ ] Cargo transfers correctly
- [ ] Ship stats update

**Combat System:**
- [ ] Encounter pirates/enemies
- [ ] Weapon selection works
- [ ] Fire weapons (damage calculated)
- [ ] Enemy AI responds (5 difficulty levels)
- [ ] Victory awards loot (4 rarity tiers)
- [ ] Defeat respawns player
- [ ] Flee option works

**Outfitting:**
- [ ] Browse equipment (6 slot types, 16 items)
- [ ] Install equipment (slots enforced)
- [ ] Uninstall equipment (refund)
- [ ] Save loadout configuration
- [ ] Load saved loadout
- [ ] Clone loadout

**Missions & Quests:**
- [ ] Accept mission (4 types: cargo, combat, explore, bounty)
- [ ] Track progress
- [ ] Complete mission (rewards)
- [ ] Abandon mission
- [ ] Max 5 active missions enforced
- [ ] Quest chains progress
- [ ] Branching narratives work

**Reputation:**
- [ ] Actions affect reputation (6 NPC factions)
- [ ] Range: -100 to +100
- [ ] Bounties tracked
- [ ] Reputation affects encounters

---

### Phase 4: Multiplayer Features

**Chat System:**
- [ ] Send message in global channel (all players see)
- [ ] Send in system channel (players in same system)
- [ ] Send in faction channel (faction members only)
- [ ] Send direct message to player
- [ ] Receive messages real-time
- [ ] Chat history persists
- [ ] Muted players cannot send

**Player Presence:**
- [ ] See online players
- [ ] See player locations (system/planet)
- [ ] Presence updates real-time
- [ ] Offline after 5min timeout

**Factions:**
- [ ] Create faction (costs credits)
- [ ] Join faction (invitation)
- [ ] Leave faction
- [ ] Deposit to treasury
- [ ] Withdraw from treasury (if authorized)
- [ ] Promote/demote members (if leader)
- [ ] Kick members (if leader)
- [ ] 4 ranks: member, officer, deputy, leader

**Territory:**
- [ ] Claim unclaimed system
- [ ] Cannot claim owned system
- [ ] Passive income generated
- [ ] Territory visualization

**P2P Trading:**
- [ ] Initiate trade with online player
- [ ] Offer items and credits
- [ ] Accept/reject trade
- [ ] Escrow prevents scamming
- [ ] Trade completion atomic
- [ ] Cannot trade with offline players

**PvP Combat:**
- [ ] Challenge player to duel
- [ ] Accept/decline challenge
- [ ] Consensual combat only
- [ ] Faction war combat
- [ ] Rewards distributed
- [ ] Losses penalized

**Leaderboards:**
- [ ] Credits leaderboard updates
- [ ] Combat leaderboard updates
- [ ] Trade volume leaderboard updates
- [ ] Exploration leaderboard updates

---

### Phase 5: Dynamic Systems

**Events:**
- [ ] Events trigger (10 types)
- [ ] Event progress tracked
- [ ] Event leaderboards update
- [ ] Event rewards distributed
- [ ] Events end on schedule

**Encounters:**
- [ ] Random encounters trigger
- [ ] Pirates attack
- [ ] Traders offer goods
- [ ] Police scan
- [ ] Distress calls
- [ ] Choices affect outcome

**News:**
- [ ] News generated from events (10+ types)
- [ ] Articles chronological
- [ ] News updates dynamically

**Achievements:**
- [ ] Achievements unlock
- [ ] Progress tracked
- [ ] Notifications display
- [ ] Incremental achievements work

---

### Phase 6: Server Features

**Metrics Endpoint:**
```bash
# Check metrics
curl http://localhost:8080/metrics | grep terminal_velocity

# View stats page
curl http://localhost:8080/stats

# Health check
curl http://localhost:8080/health
```

**Verify Metrics:**
- [ ] Connection metrics (total, active, failed)
- [ ] Player metrics (active, logins, peak)
- [ ] Game activity (trades, combat, missions)
- [ ] Economy (total credits, market volume)
- [ ] System performance (DB queries, cache hit rate)

**Session Management:**
- [ ] Auto-save every 30 seconds
- [ ] Disconnect preserves state
- [ ] Reconnect restores session
- [ ] Graceful shutdown saves all

**Admin Tools:**
- [ ] Ban player (with reason and expiration)
- [ ] Unban player
- [ ] Mute player (with expiration)
- [ ] Unmute player
- [ ] View audit log (actions logged)
- [ ] Audit log buffer (10,000 entries)
- [ ] RBAC enforced (4 roles, 20+ permissions)

---

### Phase 7: Security Testing

**Input Validation:**
- [ ] Try injecting SQL in username → should sanitize
- [ ] Try control characters in chat → should filter
- [ ] Try ANSI escape codes → should filter
- [ ] Try buffer overflow (very long input) → should limit
- [ ] Try null bytes → should filter
- [ ] Registration email max 254 chars
- [ ] Password max 128 chars
- [ ] Chat messages max 200 chars

**SQL Injection Prevention:**
```bash
# All these should fail safely:
testuser'; DROP TABLE players; --
testuser" OR "1"="1
testuser\x00admin
```

**Command Injection Prevention:**
- [ ] Try shell metacharacters in all inputs
- [ ] Verify no system commands executed

**Rate Limiting:**
- [ ] Connection rate limit works
- [ ] Auth rate limit works
- [ ] Auto-ban works
- [ ] Lockout timers accurate

**Session Security:**
- [ ] Cannot hijack other sessions
- [ ] Session tokens secure
- [ ] Horizontal privilege escalation prevented
- [ ] Vertical privilege escalation prevented

---

### Phase 8: Performance Testing

**Load Testing:**
```bash
# Simulate 10 concurrent connections
for i in {1..10}; do
  ssh -p 2222 testplayer$i@localhost &
done

# Monitor server
curl http://localhost:8080/stats
ps aux | grep server
```

**Performance Metrics:**
- [ ] 10 concurrent players - smooth
- [ ] 50 concurrent players - acceptable
- [ ] 100 concurrent players - target load
- [ ] Response times <100ms for most operations
- [ ] Memory usage stable
- [ ] CPU usage acceptable
- [ ] No memory leaks (run for hours)
- [ ] No goroutine leaks

**Race Conditions:**
```bash
# Build with race detector
go build -race -o server cmd/server/main.go

# Run and test concurrency
./server -config configs/config.yaml

# Should report no race conditions
```

**Database Performance:**
- [ ] Connection pool doesn't exhaust
- [ ] Queries fast (<50ms average)
- [ ] No deadlocks
- [ ] Indexes used correctly

---

## Troubleshooting

### Server Won't Start

**Check configuration:**
```bash
./server -config configs/config.yaml
# Look for error messages
```

**Common issues:**
- Database not running
- Wrong credentials in config
- Port 2222 already in use
- Schema not loaded

**Fix:**
```bash
# Check database
psql -U terminal_velocity -d terminal_velocity -c "SELECT version();"

# Check port
netstat -tlnp | grep 2222

# Reload schema
psql -U terminal_velocity -d terminal_velocity -f scripts/schema.sql
```

### Cannot Connect via SSH

**Test SSH server:**
```bash
ssh -v -p 2222 testplayer1@localhost
# Look for detailed error
```

**Common issues:**
- Server not running
- Wrong port
- Firewall blocking
- No accounts created

**Fix:**
```bash
# Check server is running
ps aux | grep server

# Check accounts exist
psql -U terminal_velocity -d terminal_velocity -c "SELECT username FROM players;"

# Create account
./accounts create testplayer1 test@example.com
```

### Database Connection Issues

**Check database is accessible:**
```bash
psql -U terminal_velocity -d terminal_velocity
```

**If connection fails:**
```bash
# Start PostgreSQL
sudo systemctl start postgresql

# Or with Docker
docker compose up -d postgres

# Check logs
journalctl -u postgresql -f
```

### Performance Issues

**Monitor server:**
```bash
# View metrics
curl http://localhost:8080/stats

# Check processes
top -p $(pgrep -f 'server')

# Check database
psql -U terminal_velocity -d terminal_velocity -c "
SELECT pid, usename, application_name, state, query
FROM pg_stat_activity
WHERE datname = 'terminal_velocity';
"
```

**Optimize:**
- Increase connection pool size
- Add database indexes
- Enable caching
- Profile with pprof

---

## Test Reports

### Bug Report Template

```markdown
**Title:** Brief description

**Severity:** Critical / High / Medium / Low

**Category:** Authentication / UI / Database / Gameplay / Multiplayer / etc.

**Steps to Reproduce:**
1. Step one
2. Step two
3. Step three

**Expected Behavior:**
What should happen

**Actual Behavior:**
What actually happens

**Environment:**
- Server version/commit
- Go version
- PostgreSQL version
- OS
- Number of concurrent players

**Logs/Screenshots:**
Relevant error messages

**Possible Fix:** (optional)
Ideas for fixing
```

### Performance Metrics Template

```markdown
**Test Date:** YYYY-MM-DD
**Duration:** X hours
**Concurrent Users:** X

**Metrics:**
- Average login time: X ms
- Average market operation: X ms
- Average combat turn: X ms
- Average database query: X ms
- Memory usage (idle): X MB
- Memory usage (peak): X MB
- CPU usage (average): X%
- CPU usage (peak): X%

**Issues Found:**
- List any performance issues

**Recommendations:**
- List optimization suggestions
```

---

## Next Steps After Testing

1. **Document all bugs found** using bug report template
2. **Record performance metrics**
3. **Create GitHub issues** for confirmed bugs
4. **Prioritize fixes** (Critical → High → Medium → Low)
5. **Re-test after fixes**
6. **Gather user feedback**
7. **Plan beta testing** with real users

---

**End of Live Testing Guide**

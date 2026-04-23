# Comprehensive Testing Checklist

This checklist covers integration testing across all Terminal Velocity systems. Use it when validating releases, running live-environment regression passes, or onboarding new testers. It is a human-driven checklist; automated tests live alongside source as `*_test.go` and run via `make test`.

## Pre-Testing Setup

**Environment Preparation**:
- [ ] Fresh database with `make setup-db`
- [ ] Universe generated with `./genmap -systems 100 -save`
- [ ] At least 3 test accounts created with different roles (admin, regular, new player)
- [ ] SSH keys added for public key authentication testing
- [ ] Server running with metrics enabled (`make run`)
- [ ] Metrics endpoint accessible (`curl http://localhost:8080/health`)
- [ ] Log files being written to configured location
- [ ] Database backup tested (`./scripts/backup.sh`)

**Test Accounts to Create**:
- [ ] `testadmin` - Admin user with full permissions
- [ ] `testplayer1` - Regular player for gameplay testing
- [ ] `testplayer2` - Second player for multiplayer features
- [ ] `testnewbie` - Fresh account for tutorial/onboarding testing

## 1. Authentication & Account Management

**Password Authentication**:
- [ ] Connect via SSH with valid password
- [ ] Reject invalid password
- [ ] Test rate limiting (5 failed attempts triggers 15min lockout)
- [ ] Test auto-ban (20 failures = 24h ban)
- [ ] Verify lockout message displays correctly
- [ ] Test successful login after lockout expires

**SSH Key Authentication**:
- [ ] Add SSH public key via `./accounts add-key`
- [ ] Connect using SSH key (no password prompt)
- [ ] Test with invalid key (should reject)
- [ ] Test with multiple keys for same account
- [ ] Verify fingerprint matching works correctly

**Account Registration** (if enabled):
- [ ] Access registration screen from main menu
- [ ] Create new account with valid username/email/password
- [ ] Reject invalid usernames (special chars, too short/long)
- [ ] Reject weak passwords
- [ ] Reject duplicate usernames
- [ ] Reject duplicate emails
- [ ] Verify new player starts at default location with starter ship
- [ ] Check tutorial triggers for new players

**Connection Rate Limiting**:
- [ ] Test 5 concurrent connections per IP limit
- [ ] Test 20 connections per minute per IP limit
- [ ] Verify error messages for rate limit violations
- [ ] Check metrics track connection attempts correctly

## 2. User Interface Testing

**Main Menu Screen**:
- [ ] All menu options visible and correctly labeled
- [ ] Navigation with up/down arrow keys works
- [ ] Selection with Enter key works
- [ ] Quit option (Ctrl+C or Q) exits gracefully
- [ ] Help text displays correctly

**Game/Navigation Screen**:
- [ ] Current system displays with correct info
- [ ] Current planet displays (if docked)
- [ ] Ship status shows (fuel, hull, shields, cargo)
- [ ] Jump to connected systems works
- [ ] Land on planets works
- [ ] Takeoff from planets works
- [ ] Fuel consumption calculated correctly
- [ ] Cannot jump without sufficient fuel
- [ ] Jump route visualization works
- [ ] System info panel updates correctly

**Trading Screen**:
- [ ] Market prices display for all commodities (15 items)
- [ ] Buy commodities (deducts credits, adds cargo)
- [ ] Sell commodities (adds credits, removes cargo)
- [ ] Cargo capacity limits enforced
- [ ] Insufficient credits prevents purchase
- [ ] Prices update based on supply/demand
- [ ] Tech level affects available commodities
- [ ] Transaction history updates

**Cargo Screen**:
- [ ] All cargo items listed with quantities
- [ ] Total cargo weight displayed
- [ ] Cargo capacity shown (current/max)
- [ ] Jettison cargo works
- [ ] Empty cargo bay displays message
- [ ] Cargo sorting works (if implemented)

**Shipyard Screen**:
- [ ] Available ships listed with stats (11 ship types)
- [ ] Ship prices displayed
- [ ] Cannot buy ship without sufficient credits
- [ ] Ship purchase transfers cargo correctly
- [ ] Ship purchase preserves credits correctly
- [ ] Current ship highlighted/indicated
- [ ] Ship stats comparison works
- [ ] Trade-in value calculated correctly

**Outfitter Screen**:
- [ ] All equipment categories visible (6 slot types)
- [ ] Available items listed (16 equipment items)
- [ ] Equipment prices displayed
- [ ] Install equipment (deducts credits, adds to ship)
- [ ] Uninstall equipment (refund credits, removes from ship)
- [ ] Slot limits enforced (e.g., max weapons)
- [ ] Equipment requirements checked (tech level, ship size)
- [ ] Ship stats update when equipment changes
- [ ] Cannot install without sufficient credits

**OutfitterEnhanced Screen**:
- [ ] Equipment browser with filtering works
- [ ] Equipment details panel shows stats
- [ ] Install/uninstall from enhanced UI
- [ ] Visual indicators for installed equipment
- [ ] Slot capacity visualization works
- [ ] Equipment comparison works

**Loadout System**:
- [ ] Save current loadout with custom name
- [ ] Load saved loadout (equipment changes to match)
- [ ] Clone loadout to new name
- [ ] Delete saved loadout
- [ ] List all saved loadouts
- [ ] Loadout validation (checks slot limits, requirements)

**Ship Management Screen**:
- [ ] Ship status overview displays
- [ ] Repair hull option works (deducts credits)
- [ ] Recharge shields option works (if applicable)
- [ ] Refuel option works (deducts credits)
- [ ] Cannot repair/refuel without credits
- [ ] Service costs calculated correctly
- [ ] Ship stats update after services

**Combat Screen**:
- [ ] Combat UI initializes when encountering enemy
- [ ] Player and enemy stats displayed
- [ ] Weapon selection works
- [ ] Fire weapon (damage calculation, accuracy)
- [ ] Enemy AI takes turns (5 difficulty levels)
- [ ] Combat log shows actions
- [ ] Combat ends on victory (loot awarded)
- [ ] Combat ends on defeat (respawn logic)
- [ ] Flee option works (escape chance)
- [ ] Shield damage vs hull damage works correctly
- [ ] Different weapon types work (energy, projectile, missile)

**Missions Screen**:
- [ ] Available missions listed (4 mission types)
- [ ] Accept mission (max 5 active)
- [ ] Cannot accept more than 5 missions
- [ ] Mission details displayed
- [ ] Mission progress tracking works
- [ ] Complete mission (rewards awarded)
- [ ] Abandon mission works
- [ ] Mission objectives update correctly
- [ ] Different mission types work (cargo, combat, explore, bounty)

**Quests Screen**:
- [ ] Available quests listed (7 quest types)
- [ ] Quest details and objectives displayed (12 objective types)
- [ ] Accept quest
- [ ] Quest progress tracking works
- [ ] Branching narrative choices work
- [ ] Complete quest (rewards, story progression)
- [ ] Quest chain progression works
- [ ] Quest markers/hints display correctly

**Achievements Screen**:
- [ ] All achievements listed
- [ ] Locked vs unlocked indicated
- [ ] Progress bars for incremental achievements
- [ ] Achievement descriptions shown
- [ ] Recent achievements highlighted
- [ ] Achievement categories organized
- [ ] Unlock notification triggers correctly

**Events Screen**:
- [ ] Active events listed (10 event types)
- [ ] Event details and objectives shown
- [ ] Event progress tracking works
- [ ] Event leaderboards displayed
- [ ] Participate in event
- [ ] Event rewards distributed correctly
- [ ] Event timer/countdown works
- [ ] Event completion notification

**Encounter Screen**:
- [ ] Random encounters trigger (pirates, traders, police, distress)
- [ ] Encounter dialog displays
- [ ] Encounter choices presented
- [ ] Choice selection works
- [ ] Encounter outcomes apply correctly
- [ ] Loot from encounters awarded
- [ ] Reputation changes from encounters

**News Screen**:
- [ ] News articles listed (10+ event types)
- [ ] News sorted by date (newest first)
- [ ] News details readable
- [ ] News pagination works (if many articles)
- [ ] News updates dynamically from game events
- [ ] Different news types display correctly

**Leaderboards Screen**:
- [ ] All categories accessible (4: credits, combat, trade, exploration)
- [ ] Top players listed with rankings
- [ ] Player's own rank displayed
- [ ] Stats shown for each category
- [ ] Leaderboard updates correctly
- [ ] Ties handled appropriately

**Players Screen**:
- [ ] Online players listed
- [ ] Player locations shown
- [ ] Player status (docked/in space)
- [ ] Player details viewable
- [ ] View player profile works
- [ ] Presence updates in real-time (5min timeout)

**Chat Screen**:
- [ ] All channels accessible (global, system, faction, DM)
- [ ] Send message to channel
- [ ] Receive messages in real-time
- [ ] Channel switching works
- [ ] Direct message to specific player
- [ ] Chat history scrolls correctly
- [ ] Muted players cannot send messages
- [ ] Chat formatting works

**Factions Screen**:
- [ ] Create new faction (deducts creation cost)
- [ ] Join existing faction
- [ ] Leave faction
- [ ] View faction members
- [ ] View faction treasury
- [ ] Deposit to faction treasury (if leader/officer)
- [ ] Withdraw from treasury (if authorized)
- [ ] Promote/demote members (if leader)
- [ ] Kick members (if leader)
- [ ] Faction ranks display correctly

**Territory Screen**:
- [ ] View controlled systems
- [ ] Claim system (if faction has resources)
- [ ] Cannot claim already-claimed system
- [ ] Passive income from territories
- [ ] Territory control timer works
- [ ] Territory visualization works

**Trade Screen** (Player-to-Player):
- [ ] Initiate trade with online player
- [ ] Offer items and credits
- [ ] Accept/reject trade offer
- [ ] Escrow system prevents cheating
- [ ] Trade completion transfers items correctly
- [ ] Trade cancellation returns items
- [ ] Cannot trade with offline players

**PvP Screen**:
- [ ] Challenge player to duel
- [ ] Accept/decline duel challenge
- [ ] Consensual duels work correctly
- [ ] Faction war combat works
- [ ] PvP combat follows same rules as PvE
- [ ] PvP rewards awarded correctly
- [ ] PvP losses penalized appropriately

**Help Screen**:
- [ ] Help topics listed (context-sensitive)
- [ ] Help content displays correctly
- [ ] Help navigation works
- [ ] Search help (if implemented)
- [ ] Help content accurate and helpful

**Settings Screen**:
- [ ] All setting categories accessible (6 categories)
- [ ] Color scheme selection (5 schemes)
- [ ] Color preview works
- [ ] Save settings persists to database
- [ ] Load settings on login
- [ ] Default settings reset works
- [ ] Settings validation works

**Admin Screen**:
- [ ] Only accessible to admin users
- [ ] View server stats
- [ ] Ban player (with expiration)
- [ ] Unban player
- [ ] Mute player (with expiration)
- [ ] Unmute player
- [ ] View audit log (10,000 entries)
- [ ] Server settings modification
- [ ] Permission checks enforce RBAC (4 roles, 20+ permissions)

**Tutorial Screen**:
- [ ] Tutorial triggers for new players
- [ ] All categories accessible (7 categories)
- [ ] Tutorial steps display correctly (20+ steps)
- [ ] Step progression works
- [ ] Skip tutorial option works
- [ ] Tutorial completion tracked
- [ ] Tutorial help context-sensitive

## 3. Core Gameplay Systems

**Ship Systems**:
- [ ] Hull damage tracked correctly
- [ ] Shield recharge works
- [ ] Fuel consumption accurate
- [ ] Cargo capacity enforced
- [ ] Ship upgrades apply correctly
- [ ] Ship destruction/respawn works

**Economy & Trading**:
- [ ] Supply/demand affects prices
- [ ] Tech level affects availability
- [ ] Trade profit calculations correct
- [ ] Market refreshes periodically
- [ ] Illegal commodities tracked (if implemented)
- [ ] Trade volume affects economy

**Reputation System**:
- [ ] Reputation with 6 NPC factions tracked
- [ ] Reputation range: -100 to +100
- [ ] Reputation affects prices (if implemented)
- [ ] Reputation affects encounters
- [ ] Bounty system works
- [ ] Reputation changes from actions

**Loot & Salvage**:
- [ ] Loot drops from combat (4 rarity tiers)
- [ ] Rare items drop correctly
- [ ] Loot awarded to cargo
- [ ] Salvage mechanics work
- [ ] Loot rarity affects value

**Universe & Navigation**:
- [ ] 100+ systems generated correctly
- [ ] Jump routes MST-based
- [ ] Tech levels distributed radially
- [ ] Government assignments work
- [ ] System information accurate
- [ ] Jump visualization works

## 4. Content Systems

**Quest System**:
- [ ] All 7 quest types available
- [ ] All 12 objective types work
- [ ] Branching narratives function
- [ ] Quest rewards awarded
- [ ] Quest completion tracked
- [ ] Quest chains progress correctly

**Mission System**:
- [ ] All 4 mission types available (cargo, combat, explore, bounty)
- [ ] Mission generation works
- [ ] Mission progress tracked
- [ ] Mission completion rewards
- [ ] Mission time limits enforced
- [ ] Mission failures handled

**Dynamic Events**:
- [ ] All 10 event types trigger
- [ ] Event scheduling works
- [ ] Event participation works
- [ ] Event leaderboards update
- [ ] Event rewards distributed
- [ ] Events end correctly

**Random Encounters**:
- [ ] Pirates encounter works
- [ ] Traders encounter works
- [ ] Police encounter works
- [ ] Distress call works
- [ ] Encounter frequency appropriate
- [ ] Encounter outcomes balanced

**Achievements**:
- [ ] Achievement tracking works
- [ ] Achievement unlock triggers
- [ ] Achievement notifications display
- [ ] Achievement progress saved
- [ ] All achievement types work

**News System**:
- [ ] News generated from events (10+ types)
- [ ] News articles created dynamically
- [ ] News displayed chronologically
- [ ] News updates regularly
- [ ] News content accurate

## 5. Multiplayer Features

**Player Presence**:
- [ ] Online players tracked
- [ ] Player locations updated
- [ ] Offline timeout works (5 minutes)
- [ ] Presence broadcasts to others
- [ ] Presence updates efficient

**Chat System**:
- [ ] Global chat broadcasts to all
- [ ] System chat limited to current system
- [ ] Faction chat limited to faction members
- [ ] Direct messages work
- [ ] Chat history persists
- [ ] Muted players blocked

**Faction System**:
- [ ] Faction creation works
- [ ] Faction joining works
- [ ] Faction treasury works
- [ ] Member management works
- [ ] Rank system works
- [ ] Faction permissions enforced

**Territory Control**:
- [ ] System claiming works
- [ ] Passive income generated
- [ ] Territory conflicts work
- [ ] Territory visualization accurate
- [ ] Territory control timer works

**Player Trading**:
- [ ] Trade initiation works
- [ ] Trade offers display correctly
- [ ] Escrow system prevents exploits
- [ ] Trade completion atomic
- [ ] Trade cancellation safe

**PvP Combat**:
- [ ] Duel challenges work
- [ ] Consensual combat only
- [ ] Faction wars work
- [ ] PvP rewards distributed
- [ ] PvP balance fair

**Leaderboards**:
- [ ] Credits leaderboard accurate
- [ ] Combat leaderboard accurate
- [ ] Trade leaderboard accurate
- [ ] Exploration leaderboard accurate
- [ ] Rankings update correctly

## 6. Infrastructure & Administration

**Server Administration**:
- [ ] RBAC enforced (4 roles: owner, admin, moderator, helper)
- [ ] All 20+ permissions checked
- [ ] Ban system works (with expiration)
- [ ] Mute system works (with expiration)
- [ ] Audit log records actions (10,000 buffer)
- [ ] Admin commands work
- [ ] Permission violations blocked

**Session Management**:
- [ ] Auto-save works (30 second interval)
- [ ] Session persistence across reconnects
- [ ] Graceful disconnection handling
- [ ] Session cleanup on logout
- [ ] Concurrent session handling

**Metrics & Monitoring**:
- [ ] Prometheus metrics endpoint works (`/metrics`)
- [ ] HTML stats page works (`/stats`)
- [ ] Health check endpoint works (`/health`)
- [ ] Connection metrics accurate
- [ ] Player metrics accurate
- [ ] Game activity metrics accurate
- [ ] Economy metrics accurate
- [ ] System performance metrics accurate

**Error Handling**:
- [ ] Database errors handled gracefully
- [ ] Network errors handled gracefully
- [ ] Invalid input rejected safely
- [ ] Error messages user-friendly
- [ ] Errors logged appropriately
- [ ] Retry logic works (exponential backoff)

**Logging**:
- [ ] Log levels work (debug, info, warn, error)
- [ ] Log files created correctly
- [ ] Log rotation works (if implemented)
- [ ] Sensitive data not logged (passwords)
- [ ] Logs parseable and useful

## 7. Database Operations

**Connection Pooling**:
- [ ] Connection pool initializes correctly
- [ ] Pool size configurable
- [ ] Connections reused efficiently
- [ ] Pool cleanup on shutdown
- [ ] Connection errors handled

**Repositories**:
- [ ] PlayerRepository CRUD works
- [ ] SystemRepository CRUD works
- [ ] SSHKeyRepository CRUD works
- [ ] ShipRepository CRUD works
- [ ] MarketRepository CRUD works
- [ ] All other repositories work
- [ ] Transaction handling works

**Migrations**:
- [ ] Migration tracking works (`schema_migrations` table)
- [ ] Migration up/down works
- [ ] Migration status accurate
- [ ] Migration failures rollback
- [ ] Migration scripts correct

**Backups**:
- [ ] Manual backup works (`./scripts/backup.sh`)
- [ ] Backup compression works (gzip)
- [ ] Backup retention enforced
- [ ] Old backup cleanup works
- [ ] Restore works (`./scripts/restore.sh`)
- [ ] Restore prompts for confirmation
- [ ] Large database backups work

## 8. Performance Testing

**Load Testing**:
- [ ] Test with 10 concurrent players
- [ ] Test with 50 concurrent players
- [ ] Test with 100+ concurrent players
- [ ] Response times acceptable under load
- [ ] No race conditions detected (`go test -race`)
- [ ] Memory usage stable over time
- [ ] CPU usage acceptable under load

**Database Performance**:
- [ ] Query performance acceptable
- [ ] Indexes used correctly
- [ ] Connection pool performs well
- [ ] No connection pool exhaustion
- [ ] Database locks minimal
- [ ] Transaction deadlocks handled

**Background Workers**:
- [ ] Event scheduler performs efficiently
- [ ] Metrics collection efficient
- [ ] Cleanup tasks run on schedule
- [ ] Auto-save doesn't block gameplay
- [ ] Worker goroutines don't leak

## 9. Security Testing

**Input Validation**:
- [ ] SQL injection prevented (parameterized queries)
- [ ] Command injection prevented
- [ ] XSS not applicable (terminal UI)
- [ ] Buffer overflow prevented
- [ ] Path traversal prevented
- [ ] Invalid UUIDs rejected

**Authentication Security**:
- [ ] Passwords hashed correctly (bcrypt/scrypt)
- [ ] SSH keys validated correctly
- [ ] Session tokens secure
- [ ] Rate limiting effective
- [ ] Auto-ban prevents brute force
- [ ] Timing attacks mitigated

**Authorization**:
- [ ] RBAC enforced everywhere
- [ ] Permission checks cannot be bypassed
- [ ] Horizontal privilege escalation prevented
- [ ] Vertical privilege escalation prevented
- [ ] Resource access controlled

**Data Protection**:
- [ ] Passwords never logged
- [ ] Sensitive data encrypted at rest (if applicable)
- [ ] Sensitive data encrypted in transit (SSH)
- [ ] Database credentials secured
- [ ] API tokens secured (if applicable)

## 10. Edge Cases & Error Conditions

**Boundary Testing**:
- [ ] Cargo at max capacity
- [ ] Credits at 0
- [ ] Fuel at 0 (cannot jump)
- [ ] Hull at 0 (ship destroyed)
- [ ] Reputation at -100 and +100
- [ ] Maximum missions (5) active
- [ ] Empty markets
- [ ] Single-player universe

**Concurrent Operations**:
- [ ] Multiple players trading same item
- [ ] Simultaneous faction treasury access
- [ ] Concurrent territory claims
- [ ] Race conditions in PvP
- [ ] Concurrent database writes

**Network Conditions**:
- [ ] Handle disconnections gracefully
- [ ] Reconnect preserves state
- [ ] Timeout handling works
- [ ] Partial message handling
- [ ] SSH connection drops handled

**Database Failures**:
- [ ] Connection loss handled
- [ ] Query timeout handled
- [ ] Transaction failures rollback
- [ ] Database unavailable handled
- [ ] Retry logic works

**Resource Exhaustion**:
- [ ] Memory limits respected
- [ ] File descriptor limits respected
- [ ] Database connection limits enforced
- [ ] Goroutine limits (no leaks)
- [ ] Disk space handling

## Testing Workflow

**Daily Testing Routine**:
1. Start fresh server instance
2. Test 3-5 major features from checklist
3. Document any bugs found
4. Test bug fixes immediately
5. Run automated tests: `make test`
6. Check metrics dashboard
7. Review logs for errors

**Weekly Testing Routine**:
1. Full database backup/restore test
2. Load testing with multiple clients
3. Review all open bugs
4. Test recent code changes thoroughly
5. Cross-feature integration tests
6. Performance profiling
7. Security audit of new features

**Pre-Release Testing**:
1. Complete entire checklist
2. Multi-hour stress test
3. Security penetration testing
4. Fresh player experience (tutorial → endgame)
5. All multiplayer features with real players
6. Backup/restore full cycle
7. Migration testing (fresh install vs upgrade)
8. Documentation accuracy review

## Bug Reporting Template

When you find a bug, document it with:

```
**Bug Title**: Brief description

**Severity**: Critical / High / Medium / Low

**Category**: Authentication / UI / Database / Gameplay / etc.

**Steps to Reproduce**:
1. Step one
2. Step two
3. Step three

**Expected Behavior**:
What should happen

**Actual Behavior**:
What actually happens

**Error Messages**:
Any error messages or logs

**Environment**:
- Go version
- PostgreSQL version
- OS
- Server commit hash

**Possible Fix** (optional):
Ideas for fixing the issue
```

## Performance Benchmarking

**Baseline Metrics to Record**:
- Average player login time: _______
- Average market price calculation: _______
- Average combat turn processing: _______
- Average database query time: _______
- Memory usage (idle): _______
- Memory usage (10 players): _______
- Memory usage (50 players): _______
- CPU usage (idle): _______
- CPU usage (10 players): _______
- CPU usage (50 players): _______

**Performance Goals**:
- Player login: < 500ms
- Market operations: < 100ms
- Combat turns: < 200ms
- Database queries: < 50ms
- UI responsiveness: < 16ms (60 FPS feel)
- Memory per player: < 10MB
- Support 100+ concurrent players

## Final Integration Testing Report

After completing the checklist, create a report:

**Features Tested**: ___ / Total
**Bugs Found**: ___
**Critical Bugs**: ___
**High Priority Bugs**: ___
**Medium Priority Bugs**: ___
**Low Priority Bugs**: ___
**Performance Issues**: ___
**Security Issues**: ___

**Overall Assessment**: Ready for Beta / Needs Work / Not Ready

**Notes**:
- Major issues found
- Features needing refinement
- Performance bottlenecks identified
- Security concerns
- Recommendations for beta testing

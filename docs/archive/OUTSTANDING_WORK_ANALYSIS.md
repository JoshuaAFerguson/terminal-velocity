# Outstanding Work Analysis
**Date:** 2025-11-15
**Session:** complete-remaining-todos

## Executive Summary

Comprehensive codebase scan completed. All simple/straightforward TODOs from previous sessions have been completed. Remaining items require significant architectural work and should be planned as separate major features.

## Completed in Recent Commits

✅ Mining resource tracking (migration 016)
✅ Crafting skill system (migration 017)
✅ Research points system (migration 018)
✅ Manufacturing resource management
✅ Arena matchmaking queue with ELO
✅ Tournament bracket generation
✅ Mail backend item attachment support
✅ Trade screen functionality
✅ Auto-trader framework
✅ PvP combat initiation
✅ Fleet UI improvements

## Outstanding Items Requiring Major Work

### 1. Session Manager Integration
**Location:** `internal/session/manager.go`
**Status:** Implemented but not integrated
**Impact:** Low (current event-driven persistence works well)
**Effort:** Medium-High (server lifecycle integration)

**Current State:**
- Session manager package exists with autosave worker
- Not used anywhere in codebase
- Server uses simple map for active sessions

**Required Work:**
1. Integrate into `internal/server/server.go`
2. Hook into SSH session lifecycle
3. Modify TUI to report state changes to session manager
4. Test autosave functionality
5. Add configuration options

**Recommendation:** Defer to post-launch. Current architecture is working.

### 2. Inventory/Item Selection System
**Blocks:** Mail attachments UI, Marketplace forms (auctions, contracts, bounties)
**Impact:** High (affects 4 features)
**Effort:** High (requires data model refactoring)

**Current Issue:**
- Ship cargo uses `[]CargoItem` with string commodity IDs
- Mail/marketplace expect `[]uuid.UUID` for item references
- No UUID-based inventory system exists

**Affected Features:**
1. Mail item attachments (`internal/tui/mail.go:676`)
2. Auction creation form (`internal/tui/marketplace.go:408`)
3. Contract creation form (`internal/tui/marketplace.go:443`)
4. Bounty posting form (`internal/tui/marketplace.go:477`)

**Solutions:**

**Option A:** Refactor cargo to use UUIDs (breaking change)
- Add UUID field to CargoItem
- Migrate existing cargo data
- Update all cargo operations

**Option B:** Hybrid system (RECOMMENDED)
- Keep commodity cargo as-is
- Create separate "items" table for weapons/outfits/special items
- Only allow attaching actual items (not bulk commodities)

**Option C:** Commodity-specific attachment system
- Create commodity attachment type (ID + quantity)
- Extend mail/marketplace to handle both types
- More complex but preserves current data model

**Recommendation:** Option B (hybrid). Most realistic for item trading.

### 3. Marketplace Form UIs
**Location:** `internal/tui/marketplace.go:871-904`
**Status:** Placeholder views only
**Blocked By:** Item selection system (#2)

**Current State:**
- Forms show "Press Enter to create with defaults"
- Hardcoded placeholder values used
- No user input fields

**Required Work (after inventory system):**
1. Multi-field forms with tab navigation
2. Item picker (from player inventory)
3. Numeric input for prices/quantities
4. Duration picker for auctions/contracts
5. Validation and error handling

### 4. Enhanced Integration Tests
**Location:** `internal/tui/integration_test.go`
**Status:** Placeholders with log statements
**Impact:** Low (basic tests passing)
**Effort:** Medium

**Missing Tests:**
- Trading screen with manager integration
- Shipyard screen with manager integration
- Mission board with manager integration
- Quest board with manager integration

**Required Work:**
1. Create manager mocks
2. Test manager integration points
3. Verify data flow
4. Test error handling

### 5. Combat AI Enhancement
**Location:** `internal/tui/combat_enhanced.go:686`
**Status:** Simple random AI
**Impact:** Low (functional but basic)
**Effort:** Low-Medium

**Current:** Random action selection (33% each: fire/evade/defend)
**Enhanced:** Use tactical AI from `internal/combat/ai.go`

**Benefits:**
- More challenging combat
- Difficulty-based AI
- Target prioritization
- Tactical decision making

## Scan Methodology

**Patterns Searched:**
- TODO, FIXME, XXX, HACK, PLACEHOLDER, NOTE
- `return nil, nil` (placeholder returns)
- Empty error handling
- "would come from form" comments
- Stub/mock implementations

**Files Analyzed:** 120+ Go files across:
- internal/tui/* (28 files)
- internal/*/* (managers, models, database)
- cmd/* (entry points)

**Result:** No simple standalone TODOs remaining

## Recommendations

### Short Term (Next Sprint)
1. **Document inventory system requirements** - Create detailed spec for Option B
2. **Prototype item picker UI** - Test feasibility with mock data
3. **Plan data migration** - If going with Option A/B

### Medium Term (Post-Launch)
1. **Implement inventory system** - Foundation for multiple features
2. **Build marketplace forms** - Once inventory picker ready
3. **Add integration tests** - Improve test coverage
4. **Enhanced combat AI** - Polish gameplay experience

### Long Term (Future Versions)
1. **Session manager integration** - If scaling to multiple servers
2. **Player housing/storage** - Extended inventory system
3. **Item crafting** - Uses same item picker UI

## Conclusion

The codebase is in excellent shape with all quick-win TODOs completed. Remaining work items are interconnected features requiring coordinated implementation. Primary blocker is the inventory/item selection system, which should be architected carefully as it affects multiple game systems.

**Next Steps:**
1. Prioritize inventory system design
2. Create technical specification
3. Plan phased implementation
4. Consider beta testing for data migration

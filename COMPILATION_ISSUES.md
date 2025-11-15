# Compilation Issues - Remaining Work

**Date:** 2025-11-15
**Status:** Partial fixes applied, additional work needed

---

## ✅ FIXED (Committed)

### 1. Mail Repository Field Name Mismatches
**File:** `internal/database/mail_repository.go`
**Status:** ✅ FIXED

**Changes:**
- Updated all queries to use correct Mail struct fields from `social.go`:
  - `from_player` → `sender_id`
  - `to_player` → `receiver_id`
  - `read` → `is_read`
  - `deleted_by` (array) → `is_deleted` (boolean)
- Updated all Scan operations to match new structure
- Fixed nullable SenderID handling with sql.NullString

### 2. Chat Commands UUID Type Mismatches
**File:** `internal/chat/commands.go`
**Status:** ✅ FIXED

**Changes:**
- Fixed UUID comparisons (was comparing to empty string "")
- Changed to compare against `uuid.Nil`
- Fixed field name: `player.CurrentShip` → `player.ShipName`/`player.ShipType`
- Updated to handle nullable CurrentPlanet pointer

### 3. Unused Import Cleanup
**Files:** `internal/database/social_repository.go`
**Status:** ✅ FIXED

**Changes:**
- Removed unused `"time"` import

---

## ⚠️ REMAINING ISSUES (Need Fixes)

### 1. Unused Imports
**Low Priority** - Quick fixes

```
internal/marketplace/manager.go:18 - unused "models" import
internal/arena/manager.go:13 - unused "math/rand" import
internal/mining/manager.go:19 - unused "models" import
```

**Fix:** Remove or comment out unused imports.

---

### 2. Ship Model Field Name Issues
**Medium Priority** - Structural inconsistency

**File:** `internal/capture/manager.go`
**Problem:** Code using old field names that don't exist in Ship model

**Errors:**
```
target.CurrentShields → should be: target.Shields
target.MaxShields → needs ShipType lookup: GetShipTypeByID(ship.TypeID).MaxShields
target.CurrentHull → should be: target.Hull
target.MaxHull → needs ShipType lookup: GetShipTypeByID(ship.TypeID).MaxHull
attackerShip.PlayerID → should be: attackerShip.OwnerID
defenderShip.PlayerID → should be: defenderShip.OwnerID
attempt.DefenderShip.CargoCapacity → needs ShipType lookup
attempt.AttackerShip.CargoCapacity → needs ShipType lookup
```

**Ship Model (from internal/models/ship.go):**
```go
type Ship struct {
    ID      uuid.UUID
    OwnerID uuid.UUID  // NOT PlayerID
    TypeID  string
    Hull    int        // NOT CurrentHull
    Shields int        // NOT CurrentShields
    Fuel    int
    Cargo   []CargoItem
    // ...
}

// Max values come from ShipType:
shipType := models.GetShipTypeByID(ship.TypeID)
maxHull := shipType.MaxHull
maxShields := shipType.MaxShields
cargoSpace := shipType.CargoSpace
```

**Recommended Fix:**
1. Replace `CurrentHull` with `Hull`
2. Replace `CurrentShields` with `Shields`
3. Replace `PlayerID` with `OwnerID`
4. For max values, load ShipType and access from there

---

### 3. API Converter Type Mismatches
**Medium Priority** - API layer issues

**File:** `internal/api/server/converters.go`
**Problem:** Struct field names/types don't match API definitions

**Errors:**
```
Line 98:  speed (int32) used as float64
Line 118: unknown field "Name" in api.Weapon
Line 120: weapon.Range (string) used as int32
Line 121: unknown field "Type" in api.Weapon
Line 122: weapon.Accuracy (int32) used as float64
Line 139: unknown field "Type" in api.Outfit
Line 140: unknown field "Effects" in api.Outfit
```

**Issue:** The `api.Weapon` and `api.Outfit` struct definitions in `internal/api/types.go` don't match what the converter is trying to set.

**Recommended Fix:**
1. Check `internal/api/types.go` for actual field names in `api.Weapon` and `api.Outfit`
2. Update `convertWeaponsToAPI()` and `convertOutfitsToAPI()` to match
3. May need to use different field names or add type conversions

---

### 4. Type Conversion Issues (int vs float64)
**Low Priority** - Type safety

**File:** `internal/shipsystems/manager.go`
**Problem:** Comparing/assigning int to float64

**Errors:**
```
Line 229: ship.Shields (int) < m.config.CloakActivationCost (float64)
Line 234: ship.Shields (int) -= m.config.CloakActivationCost (float64)
Line 425: ship.Fuel (int) < fuelCost (float64)
Line 430: ship.Fuel (int) -= fuelCost (float64)
```

**Recommended Fix:**
```go
// Option 1: Convert ship values to float64 for comparison
if float64(ship.Shields) < m.config.CloakActivationCost {
    ship.Shields -= int(m.config.CloakActivationCost)
}

// Option 2: Change config to use int instead of float64
// (Better - simpler and no precision loss)
```

---

### 5. Unused Variable
**Low Priority** - Code cleanup

**File:** `internal/manufacturing/manager.go:394`
**Error:** `researchCost` declared and not used

**Fix:** Either use the variable or remove the declaration.

---

## 📋 PRIORITY FIX ORDER

1. **HIGH:** Ship model field names in `capture/manager.go` (prevents compilation)
2. **HIGH:** API converter field mismatches (prevents compilation)
3. **MEDIUM:** Type conversion issues in `shipsystems/manager.go`
4. **LOW:** Unused imports (arena, marketplace, mining)
5. **LOW:** Unused variable in manufacturing

---

## 🎯 QUICK WIN SCRIPT

To fix low-priority issues quickly:

```bash
# Remove unused imports
sed -i '/\t"math\/rand"/d' internal/arena/manager.go
sed -i '/\t"github.com\/JoshuaAFerguson\/terminal-velocity\/internal\/models"/d' internal/marketplace/manager.go
sed -i '/\t"github.com\/JoshuaAFerguson\/terminal-velocity\/internal\/models"/d' internal/mining/manager.go

# Comment out unused variable
sed -i 's/researchCost :=/\/\/ researchCost :=/g' internal/manufacturing/manager.go
```

---

## 📝 NOTES

These issues existed in the codebase before our cleanup work. The main fixes we completed (mail repository and chat commands) were blocking compilation of database and chat packages.

The remaining issues are in Phase 20+ features (arena, marketplace, capture, manufacturing, mining) that were added but not fully integrated with the updated Ship model structure.

**Recommendation:** Create separate issues/PRs for each remaining category to systematically clean up the codebase.

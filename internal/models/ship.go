// File: internal/models/ship.go
// Project: Terminal Velocity
// Description: Data models for ship
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package models

import "github.com/google/uuid"

// Ship represents a spacecraft owned by a player or NPC.
//
// Ships are the primary vehicles in Terminal Velocity, used for trading, combat,
// and exploration across star systems. Each ship has a base type (ShipType) that
// defines its fundamental capabilities, which can be enhanced through equipment
// and outfitting.
//
// The ship's state includes:
//   - Current damage levels (hull, shields)
//   - Fuel reserves for hyperspace jumps
//   - Cargo hold contents
//   - Installed weapons and equipment
//   - Crew complement
//
// Ships can be upgraded, repaired, and customized at planets with appropriate
// services (shipyard, outfitter).
type Ship struct {
	// ID is the unique identifier for this ship instance
	ID uuid.UUID `json:"id"`

	// OwnerID is the UUID of the player who owns this ship
	OwnerID uuid.UUID `json:"owner_id"`

	// TypeID references the ShipType.ID that defines this ship's base characteristics
	// (e.g., "shuttle", "destroyer", "battleship")
	TypeID string `json:"type_id"`

	// Name is the player-assigned name for this ship (e.g., "The Millennium Falcon")
	Name string `json:"name"`

	// Hull is the current hull integrity (hit points)
	// Valid range: 0 to ShipType.MaxHull
	// At 0, the ship is destroyed
	Hull int `json:"hull"`

	// Shields is the current shield strength
	// Valid range: 0 to (ShipType.MaxShields + outfit bonuses)
	// Shields regenerate each combat turn based on ShipType.ShieldRegen
	Shields int `json:"shields"`

	// Fuel is the current fuel units available for hyperspace jumps
	// Valid range: 0 to (ShipType.MaxFuel + outfit bonuses)
	// Each jump consumes fuel based on jump distance
	Fuel int `json:"fuel"`

	// Cargo is the list of commodities currently in the cargo hold
	// Total cargo weight cannot exceed (ShipType.CargoSpace + outfit bonuses)
	Cargo []CargoItem `json:"cargo"`

	// Crew is the current crew count aboard the ship
	// Valid range: 0 to ShipType.MaxCrew
	// Some missions or ship functions may require minimum crew levels
	Crew int `json:"crew"`

	// Weapons is the list of installed weapon IDs (references Weapon.ID)
	// Array length cannot exceed ShipType.WeaponSlots
	// Each weapon occupies one weapon slot and consumes outfit space
	Weapons []string `json:"weapons"`

	// WeaponAmmo tracks current ammunition for each weapon slot
	// Map key is the weapon slot index (0-based), value is current ammo count
	// Only relevant for weapons with AmmoCapacity > 0 (missiles, torpedoes)
	// Omitted from JSON if empty
	WeaponAmmo map[int]int `json:"weapon_ammo,omitempty"`

	// Outfits is the list of installed outfit IDs (references Outfit.ID)
	// These provide bonuses to ship capabilities (shields, hull, cargo, fuel, speed)
	// Total outfit space used cannot exceed ShipType.OutfitSpace
	Outfits []string `json:"outfits"`
}

// ShipType defines a class of ship with its base characteristics.
//
// ShipTypes are templates that define the fundamental capabilities of ships
// in the game. There are 11 standard ship types ranging from basic Shuttles
// to powerful Battleships, organized into classes:
//   - Shuttles: Light transport vessels (Shuttle, Courier)
//   - Fighters: Combat-focused ships (Interceptor, Viper)
//   - Freighters: Heavy cargo vessels (Hauler, Bulk Freighter)
//   - Corvettes: Armed patrol vessels (Gunship, Frigate)
//   - Destroyers: Heavy warships
//   - Cruisers: Capital warships (Cruiser, Battleship)
//
// See ship_types.go for the StandardShipTypes array containing all definitions.
type ShipType struct {
	// ID is the unique identifier for this ship type (e.g., "shuttle", "battleship")
	ID string `json:"id"`

	// Name is the display name (e.g., "Shuttle", "Battleship")
	Name string `json:"name"`

	// Description provides flavor text about the ship type
	Description string `json:"description"`

	// Price is the base purchase cost in credits
	// Range: 25,000 (Shuttle) to 3,000,000 (Battleship)
	Price int64 `json:"price"`

	// MaxHull is the maximum hull integrity (hit points)
	// Range: 100 (Shuttle) to 1,500 (Battleship)
	// Can be increased with hull plating outfits
	MaxHull int `json:"max_hull"`

	// MaxShields is the maximum shield strength
	// Range: 50 (Shuttle) to 1,200 (Battleship)
	// Can be increased with shield booster outfits
	MaxShields int `json:"max_shields"`

	// ShieldRegen is the shield regeneration per combat turn
	// Range: 5 (Shuttle) to 75 (Battleship)
	// Higher values allow faster shield recovery in extended combat
	ShieldRegen int `json:"shield_regen"`

	// MaxFuel is the maximum fuel capacity in units
	// Range: 100 (Shuttle) to 350 (Battleship)
	// Can be increased with fuel tank outfits
	// Each hyperspace jump consumes fuel based on distance
	MaxFuel int `json:"max_fuel"`

	// CargoSpace is the cargo capacity in tons
	// Range: 10 (fighters) to 200 (Bulk Freighter)
	// Can be increased with cargo pod outfits
	// Each commodity unit weighs 1 ton
	CargoSpace int `json:"cargo_space"`

	// MaxCrew is the maximum crew capacity
	// Range: 1 (Interceptor) to 100 (Battleship)
	// Some ship functions may require minimum crew levels
	MaxCrew int `json:"max_crew"`

	// Speed determines initiative order in combat
	// Range: 2 (slow capital ships) to 10 (fast fighters)
	// Higher speed acts first in combat turns
	Speed int `json:"speed"`

	// Maneuverability provides evasion bonus in combat
	// Range: 2 (capital ships) to 12 (Interceptor)
	// Higher values make the ship harder to hit
	Maneuverability int `json:"maneuver"`

	// WeaponSlots is the maximum number of weapons that can be installed
	// Range: 1 (Shuttle) to 12 (Battleship)
	// Each weapon occupies one slot and consumes outfit space
	WeaponSlots int `json:"weapon_slots"`

	// OutfitSpace is the total space available for weapons and outfits
	// Range: 10 (Shuttle) to 100 (Battleship)
	// Weapons and outfits each have an OutfitSpace cost
	OutfitSpace int `json:"outfit_space"`

	// MinCombatRating is the minimum combat rating required to purchase
	// Range: 0 (starter ships) to 20 (Battleship)
	// Prevents new players from buying ships they can't handle
	MinCombatRating int `json:"min_combat_rating"`

	// Class categorizes the ship type for filtering and display
	// Valid values: shuttle, fighter, freighter, corvette, destroyer, cruiser, capital
	Class string `json:"class"`
}

// CargoItem represents a commodity stored in a ship's cargo hold.
//
// Cargo items are stacked by commodity type, with each item tracking
// the commodity ID and quantity. Each unit weighs 1 ton against the
// ship's cargo capacity limit.
type CargoItem struct {
	// CommodityID references the Commodity.ID being carried (e.g., "food", "weapons")
	CommodityID string `json:"commodity_id"`

	// Quantity is the number of units of this commodity
	// Must be > 0; empty cargo items should be removed from the cargo array
	Quantity int `json:"quantity"`
}

// Weapon represents an installable weapon system for ships.
//
// Weapons are the primary means of dealing damage in combat. There are 4 weapon
// types with different characteristics:
//   - Laser: Fast-firing, energy-based, no ammo (Pulse Laser, Beam Laser, Heavy Laser)
//   - Missile: High damage, ammo-based, slower (Missile Launcher, Torpedo Launcher)
//   - Plasma: Balanced, moderate energy use (Plasma Cannon, Plasma Turret)
//   - Railgun: Very high damage, kinetic, bypasses shields (Railgun, Heavy Railgun)
//
// See equipment.go for the StandardWeapons array containing all weapon definitions.
type Weapon struct {
	// ID is the unique identifier for this weapon type (e.g., "pulse_laser")
	ID string `json:"id"`

	// Name is the display name (e.g., "Pulse Laser")
	Name string `json:"name"`

	// Damage is the base damage dealt per hit
	// Range: 15 (Pulse Laser) to 100 (Heavy Railgun)
	Damage int `json:"damage"`

	// Range is the descriptive range category
	// Valid values: "short", "medium", "long"
	Range string `json:"range"`

	// RangeValue is the actual range in units for hit chance calculations
	// Range: 350 (short) to 1200 (long)
	// Targets beyond this range cannot be hit
	RangeValue int `json:"range_value"`

	// Type categorizes the weapon's damage mechanism
	// Valid values: "laser", "missile", "plasma", "railgun"
	Type string `json:"type"`

	// Accuracy is the base hit chance percentage
	// Range: 65-90
	// Actual hit chance modified by target's maneuverability and range
	Accuracy int `json:"accuracy"`

	// OutfitSpace is the amount of outfit space this weapon consumes
	// Range: 5-20
	// Larger, more powerful weapons consume more space
	OutfitSpace int `json:"outfit_space"`

	// Price is the purchase cost in credits
	// Range: 5,000 to 75,000
	Price int64 `json:"price"`

	// Cooldown is the time in seconds between shots
	// Range: 0.5 (fast) to 4.0 (very slow)
	// Lower values allow more frequent attacks
	Cooldown float64 `json:"cooldown"`

	// EnergyCost is the energy consumed per shot (for energy weapons)
	// Range: 10-60
	// Only applies to laser, plasma, and railgun weapons
	// Missile weapons use ammo instead
	EnergyCost int `json:"energy_cost"`

	// AmmoCapacity is the maximum ammunition count (for missile weapons)
	// Range: 10-20 for missiles, 0 for energy weapons
	// Ammo must be reloaded when depleted
	AmmoCapacity int `json:"ammo_capacity"`

	// AmmoConsumption is the ammo used per shot (for missile weapons)
	// Typically 1 for all missile weapons
	AmmoConsumption int `json:"ammo_consumption"`

	// ProjectileSpeed is the speed of fired projectiles in units per second
	// Range: 300 (torpedoes) to 1800 (heavy railgun)
	// Faster projectiles are harder to evade
	ProjectileSpeed int `json:"projectile_speed"`

	// ShieldPenetration is the percentage of damage that bypasses shields
	// Range: 0.0 (no penetration) to 0.5 (50% penetration)
	// Kinetic weapons (railguns) have higher penetration than energy weapons
	ShieldPenetration float64 `json:"shield_penetration"`
}

// Outfit represents ship equipment that enhances capabilities.
//
// Outfits provide passive bonuses to ship characteristics and don't occupy
// weapon slots, but they do consume outfit space. There are 16 standard outfits
// organized into types:
//   - Shield Boosters: Increase max shields (Mk1: +50, Mk2: +100, Mk3: +200)
//   - Hull Plating: Increase max hull (Mk1: +50, Mk2: +100, Mk3: +200)
//   - Cargo Pods: Increase cargo space (Small: +10, Medium: +20, Large: +40)
//   - Fuel Tanks: Increase fuel capacity (Small: +50, Medium: +100, Large: +200)
//   - Engine Upgrades: Increase speed (Mk1: +1, Mk2: +2, Mk3: +3)
//
// See equipment.go for the StandardOutfits array containing all definitions.
type Outfit struct {
	// ID is the unique identifier for this outfit (e.g., "shield_booster_mk1")
	ID string `json:"id"`

	// Name is the display name (e.g., "Shield Booster Mk1")
	Name string `json:"name"`

	// Description provides flavor text about the outfit
	Description string `json:"description"`

	// Type categorizes the outfit for filtering and display
	// Valid values: shield_booster, hull_reinforcement, cargo_pod, fuel_tank, engine
	Type string `json:"type"`

	// ShieldBonus is the increase to maximum shields
	// Range: 0 (not a shield outfit) to 200 (Mk3)
	// Omitted from JSON if 0
	ShieldBonus int `json:"shield_bonus,omitempty"`

	// HullBonus is the increase to maximum hull
	// Range: 0 (not a hull outfit) to 200 (Mk3)
	// Omitted from JSON if 0
	HullBonus int `json:"hull_bonus,omitempty"`

	// CargoBonus is the increase to cargo capacity in tons
	// Range: 0 (not a cargo outfit) to 40 (Large)
	// Omitted from JSON if 0
	CargoBonus int `json:"cargo_bonus,omitempty"`

	// FuelBonus is the increase to fuel capacity
	// Range: 0 (not a fuel outfit) to 200 (Large)
	// Omitted from JSON if 0
	FuelBonus int `json:"fuel_bonus,omitempty"`

	// SpeedBonus is the increase to ship speed
	// Range: 0 (not an engine outfit) to 3 (Mk3)
	// Affects both combat initiative and travel speed
	// Omitted from JSON if 0
	SpeedBonus int `json:"speed_bonus,omitempty"`

	// OutfitSpace is the amount of outfit space this outfit consumes
	// Range: 5-25
	// Must be available in ship's OutfitSpace to install
	OutfitSpace int `json:"outfit_space"`

	// Price is the purchase cost in credits
	// Range: 4,000 to 50,000
	Price int64 `json:"price"`
}

// GetCargoUsed returns total cargo space used
func (s *Ship) GetCargoUsed() int {
	total := 0
	for _, item := range s.Cargo {
		total += item.Quantity
	}
	return total
}

// GetCargoSpace returns available cargo space (requires ship type)
func (s *Ship) GetCargoSpace(shipType *ShipType) int {
	return shipType.CargoSpace - s.GetCargoUsed()
}

// CanAddCargo checks if ship has space for cargo
func (s *Ship) CanAddCargo(quantity int, shipType *ShipType) bool {
	return s.GetCargoSpace(shipType) >= quantity
}

// AddCargo adds cargo to ship
func (s *Ship) AddCargo(commodityID string, quantity int) {
	// Check if we already have this commodity
	for i := range s.Cargo {
		if s.Cargo[i].CommodityID == commodityID {
			s.Cargo[i].Quantity += quantity
			return
		}
	}
	// New commodity
	s.Cargo = append(s.Cargo, CargoItem{
		CommodityID: commodityID,
		Quantity:    quantity,
	})
}

// RemoveCargo removes cargo from ship (returns false if not enough)
func (s *Ship) RemoveCargo(commodityID string, quantity int) bool {
	for i := range s.Cargo {
		if s.Cargo[i].CommodityID == commodityID {
			if s.Cargo[i].Quantity < quantity {
				return false
			}
			s.Cargo[i].Quantity -= quantity
			if s.Cargo[i].Quantity == 0 {
				// Remove empty cargo entry
				s.Cargo = append(s.Cargo[:i], s.Cargo[i+1:]...)
			}
			return true
		}
	}
	return false
}

// GetCommodityQuantity returns quantity of a commodity in cargo
func (s *Ship) GetCommodityQuantity(commodityID string) int {
	for _, item := range s.Cargo {
		if item.CommodityID == commodityID {
			return item.Quantity
		}
	}
	return 0
}

// GetOutfitSpaceUsed returns total outfit space used by installed outfits
func (s *Ship) GetOutfitSpaceUsed() int {
	total := 0
	for _, outfitID := range s.Outfits {
		outfit := GetOutfitByID(outfitID)
		if outfit != nil {
			total += outfit.OutfitSpace
		}
	}
	// Also count weapons
	for _, weaponID := range s.Weapons {
		weapon := GetWeaponByID(weaponID)
		if weapon != nil {
			total += weapon.OutfitSpace
		}
	}
	return total
}

// GetOutfitSpaceAvailable returns available outfit space
func (s *Ship) GetOutfitSpaceAvailable(shipType *ShipType) int {
	used := s.GetOutfitSpaceUsed()
	return shipType.OutfitSpace - used
}

// CanAddOutfit checks if ship has space for an outfit
func (s *Ship) CanAddOutfit(outfit *Outfit, shipType *ShipType) bool {
	return s.GetOutfitSpaceAvailable(shipType) >= outfit.OutfitSpace
}

// CanAddWeapon checks if ship has space for a weapon and weapon slot
func (s *Ship) CanAddWeapon(weapon *Weapon, shipType *ShipType) bool {
	// Check weapon slots
	if len(s.Weapons) >= shipType.WeaponSlots {
		return false
	}
	// Check outfit space
	return s.GetOutfitSpaceAvailable(shipType) >= weapon.OutfitSpace
}

// GetWeaponAmmo returns the current ammo for a weapon slot
func (s *Ship) GetWeaponAmmo(slotIndex int) int {
	if s.WeaponAmmo == nil {
		return 0
	}
	return s.WeaponAmmo[slotIndex]
}

// SetWeaponAmmo sets the current ammo for a weapon slot
func (s *Ship) SetWeaponAmmo(slotIndex, ammo int) {
	if s.WeaponAmmo == nil {
		s.WeaponAmmo = make(map[int]int)
	}
	s.WeaponAmmo[slotIndex] = ammo
}

// ConsumeAmmo consumes ammo from a weapon slot (returns false if insufficient)
func (s *Ship) ConsumeAmmo(slotIndex, amount int) bool {
	currentAmmo := s.GetWeaponAmmo(slotIndex)
	if currentAmmo < amount {
		return false
	}
	s.SetWeaponAmmo(slotIndex, currentAmmo-amount)
	return true
}

// ReloadWeapon reloads a weapon to full capacity
func (s *Ship) ReloadWeapon(slotIndex int) {
	if slotIndex >= 0 && slotIndex < len(s.Weapons) {
		weapon := GetWeaponByID(s.Weapons[slotIndex])
		if weapon != nil {
			s.SetWeaponAmmo(slotIndex, weapon.AmmoCapacity)
		}
	}
}

// ReloadAllWeapons reloads all weapons to full capacity
func (s *Ship) ReloadAllWeapons() {
	for i := range s.Weapons {
		s.ReloadWeapon(i)
	}
}

package models

import "github.com/google/uuid"

// Ship represents a spacecraft owned by a player or NPC
type Ship struct {
	ID      uuid.UUID `json:"id"`
	OwnerID uuid.UUID `json:"owner_id"`
	TypeID  string    `json:"type_id"` // References ShipType
	Name    string    `json:"name"`

	// Status
	Hull    int `json:"hull"`
	Shields int `json:"shields"`
	Fuel    int `json:"fuel"`

	// Cargo
	Cargo []CargoItem `json:"cargo"`

	// Crew
	Crew int `json:"crew"`

	// Installed equipment (IDs of outfits)
	Weapons []string `json:"weapons"`
	Outfits []string `json:"outfits"`
}

// ShipType defines a class of ship
type ShipType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	// Cost
	Price int64 `json:"price"`

	// Combat stats
	MaxHull     int `json:"max_hull"`
	MaxShields  int `json:"max_shields"`
	ShieldRegen int `json:"shield_regen"` // per turn

	// Capacity
	MaxFuel    int `json:"max_fuel"`
	CargoSpace int `json:"cargo_space"`
	MaxCrew    int `json:"max_crew"`

	// Performance
	Speed           int `json:"speed"`    // Initiative in combat
	Maneuverability int `json:"maneuver"` // Evasion bonus

	// Hardpoints
	WeaponSlots int `json:"weapon_slots"`
	OutfitSpace int `json:"outfit_space"`

	// Requirements
	MinCombatRating int `json:"min_combat_rating"`

	// Classification
	Class string `json:"class"` // shuttle, fighter, freighter, corvette, destroyer, cruiser, capital
}

// CargoItem represents cargo in a ship's hold
type CargoItem struct {
	CommodityID string `json:"commodity_id"`
	Quantity    int    `json:"quantity"`
}

// Weapon represents an installed weapon
type Weapon struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Damage            int     `json:"damage"`
	Range             string  `json:"range"`       // short, medium, long
	RangeValue        int     `json:"range_value"` // actual range in units
	Type              string  `json:"type"`        // laser, missile, plasma, railgun
	Accuracy          int     `json:"accuracy"`
	OutfitSpace       int     `json:"outfit_space"`
	Price             int64   `json:"price"`
	Cooldown          float64 `json:"cooldown"`           // seconds between shots
	EnergyCost        int     `json:"energy_cost"`        // energy per shot (for energy weapons)
	AmmoCapacity      int     `json:"ammo_capacity"`      // max ammo (for missile weapons)
	AmmoConsumption   int     `json:"ammo_consumption"`   // ammo used per shot
	ProjectileSpeed   int     `json:"projectile_speed"`   // units per second
	ShieldPenetration float64 `json:"shield_penetration"` // 0.0-1.0, percentage that bypasses shields
}

// Outfit represents ship equipment/upgrades
type Outfit struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"` // shield_booster, cargo_pod, fuel_tank, etc.

	// Effects
	ShieldBonus int `json:"shield_bonus,omitempty"`
	HullBonus   int `json:"hull_bonus,omitempty"`
	CargoBonus  int `json:"cargo_bonus,omitempty"`
	FuelBonus   int `json:"fuel_bonus,omitempty"`
	SpeedBonus  int `json:"speed_bonus,omitempty"`

	OutfitSpace int   `json:"outfit_space"`
	Price       int64 `json:"price"`
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

package models

// Standard weapons available in the game
var StandardWeapons = []Weapon{
	// Laser Weapons (fast firing, energy-based, no ammo)
	{
		ID:                "pulse_laser",
		Name:              "Pulse Laser",
		Damage:            15,
		Range:             "medium",
		RangeValue:        500,
		Type:              "laser",
		Accuracy:          85,
		OutfitSpace:       5,
		Price:             5000,
		Cooldown:          0.5,  // 2 shots per second
		EnergyCost:        10,   // moderate energy consumption
		ProjectileSpeed:   1000, // very fast
		ShieldPenetration: 0.0,  // no shield penetration
	},
	{
		ID:                "beam_laser",
		Name:              "Beam Laser",
		Damage:            25,
		Range:             "medium",
		RangeValue:        600,
		Type:              "laser",
		Accuracy:          80,
		OutfitSpace:       8,
		Price:             12000,
		Cooldown:          1.0,  // 1 shot per second
		EnergyCost:        20,   // high energy consumption
		ProjectileSpeed:   1200, // very fast
		ShieldPenetration: 0.1,  // slight shield penetration
	},
	{
		ID:                "heavy_laser",
		Name:              "Heavy Laser",
		Damage:            40,
		Range:             "long",
		RangeValue:        800,
		Type:              "laser",
		Accuracy:          75,
		OutfitSpace:       12,
		Price:             25000,
		Cooldown:          1.5, // slower firing
		EnergyCost:        35,  // very high energy consumption
		ProjectileSpeed:   1000,
		ShieldPenetration: 0.15, // modest shield penetration
	},

	// Missile Weapons (high damage, ammo-based, slower)
	{
		ID:                "missile_launcher",
		Name:              "Missile Launcher",
		Damage:            50,
		Range:             "long",
		RangeValue:        1000,
		Type:              "missile",
		Accuracy:          70,
		OutfitSpace:       10,
		Price:             15000,
		Cooldown:          2.0, // slow reload
		AmmoCapacity:      20,  // 20 missiles
		AmmoConsumption:   1,   // 1 missile per shot
		ProjectileSpeed:   400, // slower projectile
		ShieldPenetration: 0.2, // good shield penetration
	},
	{
		ID:                "torpedo_launcher",
		Name:              "Torpedo Launcher",
		Damage:            80,
		Range:             "long",
		RangeValue:        1200,
		Type:              "missile",
		Accuracy:          65,
		OutfitSpace:       15,
		Price:             35000,
		Cooldown:          3.0, // very slow reload
		AmmoCapacity:      10,  // 10 torpedoes
		AmmoConsumption:   1,   // 1 torpedo per shot
		ProjectileSpeed:   300, // slow projectile
		ShieldPenetration: 0.4, // excellent shield penetration
	},

	// Plasma Weapons (balanced, moderate energy use)
	{
		ID:                "plasma_cannon",
		Name:              "Plasma Cannon",
		Damage:            35,
		Range:             "medium",
		RangeValue:        550,
		Type:              "plasma",
		Accuracy:          75,
		OutfitSpace:       10,
		Price:             20000,
		Cooldown:          1.2,  // moderate firing rate
		EnergyCost:        25,   // moderate energy consumption
		ProjectileSpeed:   600,  // moderate speed
		ShieldPenetration: 0.25, // good shield penetration
	},
	{
		ID:                "plasma_turret",
		Name:              "Plasma Turret",
		Damage:            30,
		Range:             "short",
		RangeValue:        350,
		Type:              "plasma",
		Accuracy:          90,
		OutfitSpace:       12,
		Price:             18000,
		Cooldown:          0.8, // faster firing
		EnergyCost:        18,  // lower energy consumption
		ProjectileSpeed:   700,
		ShieldPenetration: 0.2, // good shield penetration
	},

	// Railgun Weapons (very high damage, kinetic, bypasses some shields)
	{
		ID:                "railgun",
		Name:              "Railgun",
		Damage:            60,
		Range:             "long",
		RangeValue:        900,
		Type:              "railgun",
		Accuracy:          70,
		OutfitSpace:       14,
		Price:             40000,
		Cooldown:          2.5,  // slow firing
		EnergyCost:        40,   // high energy consumption
		ProjectileSpeed:   1500, // extremely fast
		ShieldPenetration: 0.35, // excellent shield penetration
	},
	{
		ID:                "heavy_railgun",
		Name:              "Heavy Railgun",
		Damage:            100,
		Range:             "long",
		RangeValue:        1000,
		Type:              "railgun",
		Accuracy:          65,
		OutfitSpace:       20,
		Price:             75000,
		Cooldown:          4.0,  // very slow firing
		EnergyCost:        60,   // very high energy consumption
		ProjectileSpeed:   1800, // extremely fast
		ShieldPenetration: 0.5,  // massive shield penetration
	},
}

// Standard outfits available in the game
var StandardOutfits = []Outfit{
	// Shield Boosters
	{
		ID:          "shield_booster_mk1",
		Name:        "Shield Booster Mk1",
		Description: "Increases maximum shield capacity",
		Type:        "shield_booster",
		ShieldBonus: 50,
		OutfitSpace: 5,
		Price:       8000,
	},
	{
		ID:          "shield_booster_mk2",
		Name:        "Shield Booster Mk2",
		Description: "Advanced shield enhancement system",
		Type:        "shield_booster",
		ShieldBonus: 100,
		OutfitSpace: 8,
		Price:       18000,
	},
	{
		ID:          "shield_booster_mk3",
		Name:        "Shield Booster Mk3",
		Description: "Military-grade shield amplifier",
		Type:        "shield_booster",
		ShieldBonus: 200,
		OutfitSpace: 12,
		Price:       40000,
	},

	// Hull Reinforcement
	{
		ID:          "hull_plating_mk1",
		Name:        "Hull Plating Mk1",
		Description: "Additional armor plating",
		Type:        "hull_reinforcement",
		HullBonus:   50,
		OutfitSpace: 5,
		Price:       6000,
	},
	{
		ID:          "hull_plating_mk2",
		Name:        "Hull Plating Mk2",
		Description: "Composite armor enhancement",
		Type:        "hull_reinforcement",
		HullBonus:   100,
		OutfitSpace: 8,
		Price:       15000,
	},
	{
		ID:          "hull_plating_mk3",
		Name:        "Hull Plating Mk3",
		Description: "Military-grade armor system",
		Type:        "hull_reinforcement",
		HullBonus:   200,
		OutfitSpace: 12,
		Price:       35000,
	},

	// Cargo Expansions
	{
		ID:          "cargo_pod_small",
		Name:        "Small Cargo Pod",
		Description: "Adds 10 tons of cargo space",
		Type:        "cargo_pod",
		CargoBonus:  10,
		OutfitSpace: 8,
		Price:       5000,
	},
	{
		ID:          "cargo_pod_medium",
		Name:        "Medium Cargo Pod",
		Description: "Adds 20 tons of cargo space",
		Type:        "cargo_pod",
		CargoBonus:  20,
		OutfitSpace: 15,
		Price:       12000,
	},
	{
		ID:          "cargo_pod_large",
		Name:        "Large Cargo Pod",
		Description: "Adds 40 tons of cargo space",
		Type:        "cargo_pod",
		CargoBonus:  40,
		OutfitSpace: 25,
		Price:       25000,
	},

	// Fuel Tanks
	{
		ID:          "fuel_tank_small",
		Name:        "Small Fuel Tank",
		Description: "Adds 50 units of fuel capacity",
		Type:        "fuel_tank",
		FuelBonus:   50,
		OutfitSpace: 5,
		Price:       4000,
	},
	{
		ID:          "fuel_tank_medium",
		Name:        "Medium Fuel Tank",
		Description: "Adds 100 units of fuel capacity",
		Type:        "fuel_tank",
		FuelBonus:   100,
		OutfitSpace: 10,
		Price:       9000,
	},
	{
		ID:          "fuel_tank_large",
		Name:        "Large Fuel Tank",
		Description: "Adds 200 units of fuel capacity",
		Type:        "fuel_tank",
		FuelBonus:   200,
		OutfitSpace: 18,
		Price:       20000,
	},

	// Engine Upgrades
	{
		ID:          "engine_upgrade_mk1",
		Name:        "Engine Upgrade Mk1",
		Description: "Increases ship speed",
		Type:        "engine",
		SpeedBonus:  1,
		OutfitSpace: 10,
		Price:       10000,
	},
	{
		ID:          "engine_upgrade_mk2",
		Name:        "Engine Upgrade Mk2",
		Description: "Advanced thruster system",
		Type:        "engine",
		SpeedBonus:  2,
		OutfitSpace: 15,
		Price:       25000,
	},
	{
		ID:          "engine_upgrade_mk3",
		Name:        "Engine Upgrade Mk3",
		Description: "Military-grade propulsion",
		Type:        "engine",
		SpeedBonus:  3,
		OutfitSpace: 20,
		Price:       50000,
	},
}

// GetWeaponByID finds a weapon by its ID
func GetWeaponByID(id string) *Weapon {
	for i := range StandardWeapons {
		if StandardWeapons[i].ID == id {
			return &StandardWeapons[i]
		}
	}
	return nil
}

// GetOutfitByID finds an outfit by its ID
func GetOutfitByID(id string) *Outfit {
	for i := range StandardOutfits {
		if StandardOutfits[i].ID == id {
			return &StandardOutfits[i]
		}
	}
	return nil
}

// GetWeaponsByType returns all weapons of a given type
func GetWeaponsByType(weaponType string) []Weapon {
	var result []Weapon
	for _, weapon := range StandardWeapons {
		if weapon.Type == weaponType {
			result = append(result, weapon)
		}
	}
	return result
}

// GetOutfitsByType returns all outfits of a given type
func GetOutfitsByType(outfitType string) []Outfit {
	var result []Outfit
	for _, outfit := range StandardOutfits {
		if outfit.Type == outfitType {
			result = append(result, outfit)
		}
	}
	return result
}

// CalculateShipBonuses calculates total bonuses from installed outfits
func CalculateShipBonuses(outfitIDs []string) (shields, hull, cargo, fuel, speed int) {
	for _, id := range outfitIDs {
		outfit := GetOutfitByID(id)
		if outfit != nil {
			shields += outfit.ShieldBonus
			hull += outfit.HullBonus
			cargo += outfit.CargoBonus
			fuel += outfit.FuelBonus
			speed += outfit.SpeedBonus
		}
	}
	return
}

// File: internal/outfitting/equipment.go
// Project: Terminal Velocity
// Description: Pre-defined equipment catalog
// Version: 1.0.0
// Author: Terminal Velocity Development Team
// Created: 2025-01-07

package outfitting

import "github.com/JoshuaAFerguson/terminal-velocity/internal/models"

// Weapon equipment

func createLaserCannon() *models.Equipment {
	return &models.Equipment{
		ID:          "laser_cannon_mk1",
		Name:        "Laser Cannon Mk1",
		Description: "Standard energy weapon with moderate damage and high accuracy",
		Category:    models.CategoryWeapon,
		SlotType:    models.SlotWeapon,
		SlotSize:    1,
		MinTechLevel: 2,
		Price:       15000,
		OutfitSpace: 10,
		Rarity:      "common",
		Stats: models.EquipmentStats{
			Damage:            25,
			Range:             150,
			Accuracy:          85,
			Cooldown:          1.5,
			EnergyCost:        15,
			ShieldPenetration: 0.3,
		},
	}
}

func createPlasmaTurret() *models.Equipment {
	return &models.Equipment{
		ID:          "plasma_turret_mk2",
		Name:        "Plasma Turret Mk2",
		Description: "Heavy plasma weapon with high damage but slower rate of fire",
		Category:    models.CategoryWeapon,
		SlotType:    models.SlotWeapon,
		SlotSize:    2,
		MinTechLevel: 4,
		Price:       45000,
		OutfitSpace: 25,
		Rarity:      "uncommon",
		Stats: models.EquipmentStats{
			Damage:            60,
			Range:             120,
			Accuracy:          70,
			Cooldown:          3.0,
			EnergyCost:        40,
			ShieldPenetration: 0.5,
		},
	}
}

func createRailgun() *models.Equipment {
	return &models.Equipment{
		ID:          "railgun_heavy",
		Name:        "Heavy Railgun",
		Description: "Long-range kinetic weapon with extreme shield penetration",
		Category:    models.CategoryWeapon,
		SlotType:    models.SlotWeapon,
		SlotSize:    2,
		MinTechLevel: 5,
		Price:       80000,
		OutfitSpace: 30,
		Rarity:      "rare",
		Stats: models.EquipmentStats{
			Damage:            75,
			Range:             250,
			Accuracy:          90,
			Cooldown:          4.0,
			EnergyCost:        50,
			ShieldPenetration: 0.8,
		},
	}
}

func createMissileLauncher() *models.Equipment {
	return &models.Equipment{
		ID:          "missile_launcher",
		Name:        "Missile Launcher",
		Description: "Guided missile system with high damage potential",
		Category:    models.CategoryWeapon,
		SlotType:    models.SlotWeapon,
		SlotSize:    2,
		MinTechLevel: 3,
		Price:       35000,
		OutfitSpace: 20,
		Rarity:      "uncommon",
		Stats: models.EquipmentStats{
			Damage:      100,
			Range:       200,
			Accuracy:    95,
			Cooldown:    5.0,
			AmmoCapacity: 20,
		},
	}
}

// Shield equipment

func createBasicShield() *models.Equipment {
	return &models.Equipment{
		ID:          "shield_basic",
		Name:        "Basic Shield Generator",
		Description: "Entry-level shield system for small ships",
		Category:    models.CategoryDefense,
		SlotType:    models.SlotShield,
		SlotSize:    1,
		MinTechLevel: 1,
		Price:       10000,
		OutfitSpace: 15,
		Rarity:      "common",
		Stats: models.EquipmentStats{
			ShieldHP:    100,
			ShieldRegen: 5,
		},
	}
}

func createAdvancedShield() *models.Equipment {
	return &models.Equipment{
		ID:          "shield_advanced",
		Name:        "Advanced Shield Array",
		Description: "High-capacity shield system with fast regeneration",
		Category:    models.CategoryDefense,
		SlotType:    models.SlotShield,
		SlotSize:    2,
		MinTechLevel: 4,
		Price:       50000,
		OutfitSpace: 35,
		Rarity:      "uncommon",
		Stats: models.EquipmentStats{
			ShieldHP:    250,
			ShieldRegen: 15,
		},
	}
}

func createMilitaryShield() *models.Equipment {
	return &models.Equipment{
		ID:          "shield_military",
		Name:        "Military-Grade Shield",
		Description: "Top-tier shield system with extreme durability",
		Category:    models.CategoryDefense,
		SlotType:    models.SlotShield,
		SlotSize:    3,
		MinTechLevel: 6,
		RequiredLicense: "Military License",
		Price:       150000,
		OutfitSpace: 50,
		Rarity:      "military",
		Stats: models.EquipmentStats{
			ShieldHP:    500,
			ShieldRegen: 25,
			ArmorRating: 20,
		},
	}
}

// Engine equipment

func createBasicEngine() *models.Equipment {
	return &models.Equipment{
		ID:          "engine_basic",
		Name:        "Standard Ion Drive",
		Description: "Reliable ion propulsion system",
		Category:    models.CategoryPropulsion,
		SlotType:    models.SlotEngine,
		SlotSize:    1,
		MinTechLevel: 1,
		Price:       12000,
		OutfitSpace: 20,
		Rarity:      "common",
		Stats: models.EquipmentStats{
			SpeedBonus: 10,
			TurnRate:   5,
		},
	}
}

func createAfterburnerEngine() *models.Equipment {
	return &models.Equipment{
		ID:          "engine_afterburner",
		Name:        "Afterburner Drive",
		Description: "High-speed engine with afterburner capability",
		Category:    models.CategoryPropulsion,
		SlotType:    models.SlotEngine,
		SlotSize:    2,
		MinTechLevel: 3,
		Price:       40000,
		OutfitSpace: 30,
		Rarity:      "uncommon",
		Stats: models.EquipmentStats{
			SpeedBonus:       20,
			TurnRate:         10,
			AfterburnerBoost: 50,
		},
	}
}

func createMilitaryEngine() *models.Equipment {
	return &models.Equipment{
		ID:          "engine_military",
		Name:        "Military Fusion Drive",
		Description: "Cutting-edge propulsion with unmatched performance",
		Category:    models.CategoryPropulsion,
		SlotType:    models.SlotEngine,
		SlotSize:    3,
		MinTechLevel: 6,
		RequiredLicense: "Military License",
		Price:       120000,
		OutfitSpace: 45,
		Rarity:      "military",
		Stats: models.EquipmentStats{
			SpeedBonus:       35,
			TurnRate:         20,
			AfterburnerBoost: 80,
		},
	}
}

// Reactor equipment

func createBasicReactor() *models.Equipment {
	return &models.Equipment{
		ID:          "reactor_basic",
		Name:        "Fission Reactor",
		Description: "Basic power generation system",
		Category:    models.CategoryPower,
		SlotType:    models.SlotReactor,
		SlotSize:    1,
		MinTechLevel: 1,
		Price:       15000,
		OutfitSpace: 25,
		Rarity:      "common",
		Stats: models.EquipmentStats{
			EnergyOutput:  100,
			EnergyStorage: 500,
		},
	}
}

func createFusionReactor() *models.Equipment {
	return &models.Equipment{
		ID:          "reactor_fusion",
		Name:        "Fusion Reactor",
		Description: "Advanced fusion-based power system",
		Category:    models.CategoryPower,
		SlotType:    models.SlotReactor,
		SlotSize:    2,
		MinTechLevel: 4,
		Price:       60000,
		OutfitSpace: 40,
		Rarity:      "uncommon",
		Stats: models.EquipmentStats{
			EnergyOutput:  250,
			EnergyStorage: 1200,
		},
	}
}

func createAntimatterReactor() *models.Equipment {
	return &models.Equipment{
		ID:          "reactor_antimatter",
		Name:        "Antimatter Reactor",
		Description: "Experimental antimatter power core with massive output",
		Category:    models.CategoryPower,
		SlotType:    models.SlotReactor,
		SlotSize:    3,
		MinTechLevel: 7,
		Price:       200000,
		OutfitSpace: 60,
		Rarity:      "experimental",
		Stats: models.EquipmentStats{
			EnergyOutput:  500,
			EnergyStorage: 2500,
		},
	}
}

// Utility equipment

func createCargoPod() *models.Equipment {
	return &models.Equipment{
		ID:          "cargo_pod",
		Name:        "Cargo Pod",
		Description: "External cargo storage module",
		Category:    models.CategoryUtility,
		SlotType:    models.SlotUtility,
		SlotSize:    1,
		MinTechLevel: 1,
		Price:       5000,
		OutfitSpace: 10,
		Rarity:      "common",
		Stats: models.EquipmentStats{
			CargoBonus: 50,
		},
	}
}

func createFuelTank() *models.Equipment {
	return &models.Equipment{
		ID:          "fuel_tank",
		Name:        "Extended Fuel Tank",
		Description: "Additional fuel storage for long journeys",
		Category:    models.CategoryUtility,
		SlotType:    models.SlotUtility,
		SlotSize:    1,
		MinTechLevel: 1,
		Price:       8000,
		OutfitSpace: 15,
		Rarity:      "common",
		Stats: models.EquipmentStats{
			FuelBonus: 100,
		},
	}
}

func createScanner() *models.Equipment {
	return &models.Equipment{
		ID:          "scanner_advanced",
		Name:        "Advanced Scanner Array",
		Description: "Long-range scanning and detection system",
		Category:    models.CategoryUtility,
		SlotType:    models.SlotUtility,
		SlotSize:    1,
		MinTechLevel: 3,
		Price:       25000,
		OutfitSpace: 12,
		Rarity:      "uncommon",
		Stats: models.EquipmentStats{
			ScannerRange: 500,
			JumpRange:    2,
		},
	}
}

func createRepairDrone() *models.Equipment {
	return &models.Equipment{
		ID:          "repair_drone",
		Name:        "Automated Repair Drone",
		Description: "Self-repairing hull maintenance system",
		Category:    models.CategoryUtility,
		SlotType:    models.SlotUtility,
		SlotSize:    2,
		MinTechLevel: 4,
		Price:       35000,
		OutfitSpace: 20,
		Rarity:      "rare",
		Stats: models.EquipmentStats{
			RepairRate: 10,
			HullBonus:  50,
		},
	}
}

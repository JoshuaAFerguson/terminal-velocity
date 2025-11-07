package models

// Standard ship types available in the game
var StandardShipTypes = []ShipType{
	// Starter Ships
	{
		ID:              "shuttle",
		Name:            "Shuttle",
		Description:     "A basic short-range transport vessel. Slow and lightly armed, but affordable.",
		Price:           25000,
		MaxHull:         100,
		MaxShields:      50,
		ShieldRegen:     5,
		MaxFuel:         100,
		CargoSpace:      20,
		MaxCrew:         2,
		Speed:           3,
		Maneuverability: 5,
		WeaponSlots:     1,
		OutfitSpace:     10,
		MinCombatRating: 0,
		Class:           "shuttle",
	},
	{
		ID:              "courier",
		Name:            "Courier",
		Description:     "Fast light transport with decent cargo capacity. Popular with traders.",
		Price:           50000,
		MaxHull:         150,
		MaxShields:      75,
		ShieldRegen:     7,
		MaxFuel:         150,
		CargoSpace:      40,
		MaxCrew:         3,
		Speed:           7,
		Maneuverability: 8,
		WeaponSlots:     2,
		OutfitSpace:     15,
		MinCombatRating: 0,
		Class:           "shuttle",
	},

	// Fighters
	{
		ID:              "interceptor",
		Name:            "Interceptor",
		Description:     "Light fighter with exceptional speed and maneuverability.",
		Price:           75000,
		MaxHull:         120,
		MaxShields:      100,
		ShieldRegen:     10,
		MaxFuel:         120,
		CargoSpace:      10,
		MaxCrew:         1,
		Speed:           10,
		Maneuverability: 12,
		WeaponSlots:     3,
		OutfitSpace:     20,
		MinCombatRating: 2,
		Class:           "fighter",
	},
	{
		ID:              "viper",
		Name:            "Viper",
		Description:     "Balanced fighter with good firepower and durability.",
		Price:           120000,
		MaxHull:         200,
		MaxShields:      150,
		ShieldRegen:     12,
		MaxFuel:         140,
		CargoSpace:      15,
		MaxCrew:         2,
		Speed:           8,
		Maneuverability: 10,
		WeaponSlots:     4,
		OutfitSpace:     25,
		MinCombatRating: 3,
		Class:           "fighter",
	},

	// Freighters
	{
		ID:              "hauler",
		Name:            "Hauler",
		Description:     "Large cargo vessel with massive storage capacity. Slow but profitable.",
		Price:           150000,
		MaxHull:         300,
		MaxShields:      100,
		ShieldRegen:     8,
		MaxFuel:         200,
		CargoSpace:      100,
		MaxCrew:         5,
		Speed:           3,
		Maneuverability: 3,
		WeaponSlots:     2,
		OutfitSpace:     20,
		MinCombatRating: 0,
		Class:           "freighter",
	},
	{
		ID:              "bulk_freighter",
		Name:            "Bulk Freighter",
		Description:     "Massive commercial vessel. Maximum cargo capacity.",
		Price:           300000,
		MaxHull:         400,
		MaxShields:      150,
		ShieldRegen:     10,
		MaxFuel:         250,
		CargoSpace:      200,
		MaxCrew:         10,
		Speed:           2,
		Maneuverability: 2,
		WeaponSlots:     3,
		OutfitSpace:     30,
		MinCombatRating: 0,
		Class:           "freighter",
	},

	// Corvettes
	{
		ID:              "gunship",
		Name:            "Gunship",
		Description:     "Armed patrol vessel. Good balance of firepower and cargo space.",
		Price:           250000,
		MaxHull:         350,
		MaxShields:      200,
		ShieldRegen:     15,
		MaxFuel:         180,
		CargoSpace:      50,
		MaxCrew:         8,
		Speed:           6,
		Maneuverability: 7,
		WeaponSlots:     5,
		OutfitSpace:     35,
		MinCombatRating: 5,
		Class:           "corvette",
	},
	{
		ID:              "frigate",
		Name:            "Frigate",
		Description:     "Military escort vessel with heavy armament.",
		Price:           400000,
		MaxHull:         500,
		MaxShields:      300,
		ShieldRegen:     20,
		MaxFuel:         200,
		CargoSpace:      60,
		MaxCrew:         15,
		Speed:           5,
		Maneuverability: 6,
		WeaponSlots:     6,
		OutfitSpace:     45,
		MinCombatRating: 7,
		Class:           "corvette",
	},

	// Destroyers
	{
		ID:              "destroyer",
		Name:            "Destroyer",
		Description:     "Heavy warship with formidable firepower.",
		Price:           750000,
		MaxHull:         700,
		MaxShields:      500,
		ShieldRegen:     30,
		MaxFuel:         250,
		CargoSpace:      80,
		MaxCrew:         25,
		Speed:           4,
		Maneuverability: 4,
		WeaponSlots:     8,
		OutfitSpace:     60,
		MinCombatRating: 10,
		Class:           "destroyer",
	},

	// Cruisers
	{
		ID:              "cruiser",
		Name:            "Cruiser",
		Description:     "Capital-class warship. Massive firepower and durability.",
		Price:           1500000,
		MaxHull:         1000,
		MaxShields:      800,
		ShieldRegen:     50,
		MaxFuel:         300,
		CargoSpace:      100,
		MaxCrew:         50,
		Speed:           3,
		Maneuverability: 3,
		WeaponSlots:     10,
		OutfitSpace:     80,
		MinCombatRating: 15,
		Class:           "cruiser",
	},
	{
		ID:              "battleship",
		Name:            "Battleship",
		Description:     "Ultimate warship. Devastating in combat.",
		Price:           3000000,
		MaxHull:         1500,
		MaxShields:      1200,
		ShieldRegen:     75,
		MaxFuel:         350,
		CargoSpace:      120,
		MaxCrew:         100,
		Speed:           2,
		Maneuverability: 2,
		WeaponSlots:     12,
		OutfitSpace:     100,
		MinCombatRating: 20,
		Class:           "capital",
	},
}

// GetShipTypeByID returns a ship type by its ID
func GetShipTypeByID(id string) *ShipType {
	for i := range StandardShipTypes {
		if StandardShipTypes[i].ID == id {
			return &StandardShipTypes[i]
		}
	}
	return nil
}

// GetShipTypesByClass returns all ship types in a class
func GetShipTypesByClass(class string) []ShipType {
	var result []ShipType
	for _, shipType := range StandardShipTypes {
		if shipType.Class == class {
			result = append(result, shipType)
		}
	}
	return result
}

// GetAffordableShipTypes returns ship types within a price range
func GetAffordableShipTypes(maxPrice int64) []ShipType {
	var result []ShipType
	for _, shipType := range StandardShipTypes {
		if shipType.Price <= maxPrice {
			result = append(result, shipType)
		}
	}
	return result
}

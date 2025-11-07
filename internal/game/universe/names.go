// File: internal/game/universe/names.go
// Project: Terminal Velocity
// Description: Procedural universe generation: names
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package universe

import (
	"fmt"
	"math/rand"
)

// NameGenerator generates star system names
type NameGenerator struct {
	rand      *rand.Rand
	usedNames map[string]bool
}

// NewNameGenerator creates a new name generator
func NewNameGenerator(r *rand.Rand) *NameGenerator {
	return &NameGenerator{
		rand:      r,
		usedNames: make(map[string]bool),
	}
}

// Greek letter prefixes for star names
var greekLetters = []string{
	"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta", "Eta", "Theta",
	"Iota", "Kappa", "Lambda", "Mu", "Nu", "Xi", "Omicron", "Pi",
	"Rho", "Sigma", "Tau", "Upsilon", "Phi", "Chi", "Psi", "Omega",
}

// Constellation names for star suffixes
var constellations = []string{
	"Centauri", "Eridani", "Ceti", "Draconis", "Leonis", "Aquarii", "Orionis",
	"Scorpii", "Cassiopeiae", "Andromedae", "Lyrae", "Cygni", "Aquilae",
	"Ursae", "Bootis", "Virginis", "Geminorum", "Tauri", "Sagittarii",
	"Capricorni", "Piscium", "Arietis", "Cancri", "Librae", "Persei",
	"Herculis", "Ophiuchi", "Serpentis", "Coronae", "Hydrae",
}

// Real star names (for variety)
var realStars = []string{
	"Sirius", "Canopus", "Arcturus", "Vega", "Capella", "Rigel", "Procyon",
	"Betelgeuse", "Achernar", "Altair", "Aldebaran", "Antares", "Spica",
	"Pollux", "Fomalhaut", "Deneb", "Regulus", "Adhara", "Castor", "Bellatrix",
	"Elnath", "Miaplacidus", "Alnilam", "Alnitak", "Alnair", "Alioth",
	"Dubhe", "Mirfak", "Wezen", "Sargas", "Kaus Australis", "Avior",
	"Alkaid", "Menkalinan", "Atria", "Alhena", "Peacock", "Alsephina",
	"Mirzam", "Alphard", "Hamal", "Polaris", "Alderamin", "Denebola",
}

// Procedural name components
var namePrefix = []string{
	"New", "Neo", "Nova", "Omega", "Proxima", "Ultima", "Prima", "Kepler",
	"Ross", "Gliese", "Wolf", "Lacaille", "Luyten", "Barnard", "Kruger",
	"Groombridge", "Lalande", "Struve", "Innes", "van", "Stein",
}

var nameSuffix = []string{
	"Prime", "Secundus", "Tertius", "Major", "Minor", "Station", "Outpost",
	"Haven", "Refuge", "Bastion", "Forge", "Reach", "Crossing", "Gate",
	"Nexus", "Hub", "Point", "Junction", "Terminal", "Threshold",
}

// GenerateSystemName generates a unique star system name
func (ng *NameGenerator) GenerateSystemName() string {
	maxAttempts := 100

	for i := 0; i < maxAttempts; i++ {
		var name string

		// Different naming strategies
		switch ng.rand.Intn(4) {
		case 0:
			// Greek + Constellation (e.g., "Alpha Centauri")
			name = ng.generateGreekConstellation()
		case 1:
			// Real star name
			name = realStars[ng.rand.Intn(len(realStars))]
		case 2:
			// Prefix + number (e.g., "Kepler-452")
			name = ng.generateCatalogName()
		case 3:
			// Prefix + Suffix (e.g., "New Haven")
			name = ng.generateCompoundName()
		}

		// Check if unique
		if !ng.usedNames[name] {
			ng.usedNames[name] = true
			return name
		}
	}

	// Fallback: generate guaranteed unique name
	return ng.generateFallbackName()
}

// generateGreekConstellation generates Greek letter + constellation name
func (ng *NameGenerator) generateGreekConstellation() string {
	greek := greekLetters[ng.rand.Intn(len(greekLetters))]
	constellation := constellations[ng.rand.Intn(len(constellations))]
	return fmt.Sprintf("%s %s", greek, constellation)
}

// generateCatalogName generates catalog-style name (e.g., "Kepler-442")
func (ng *NameGenerator) generateCatalogName() string {
	prefix := namePrefix[ng.rand.Intn(len(namePrefix))]
	number := ng.rand.Intn(9999) + 1
	return fmt.Sprintf("%s-%d", prefix, number)
}

// generateCompoundName generates compound name (e.g., "New Horizon")
func (ng *NameGenerator) generateCompoundName() string {
	prefix := namePrefix[ng.rand.Intn(len(namePrefix))]
	suffix := nameSuffix[ng.rand.Intn(len(nameSuffix))]
	return fmt.Sprintf("%s %s", prefix, suffix)
}

// generateFallbackName generates guaranteed unique name
func (ng *NameGenerator) generateFallbackName() string {
	counter := len(ng.usedNames)
	name := fmt.Sprintf("System-%d", counter)
	ng.usedNames[name] = true
	return name
}

// System descriptions based on faction and type
var coreDescriptions = []string{
	"A highly developed core world with massive orbital installations and billions of inhabitants.",
	"Capital of a sector, this system hosts impressive military and civilian infrastructure.",
	"A wealthy industrial hub with state-of-the-art shipyards and manufacturing facilities.",
	"Home to one of humanity's most prestigious universities and research centers.",
	"A major trade nexus where goods from across the galaxy change hands.",
}

var midDescriptions = []string{
	"A prosperous trade station serves as the heart of this busy system.",
	"Mining operations and refineries dot the asteroid belts of this resource-rich system.",
	"A growing colonial world striving to match the prosperity of the core systems.",
	"Agricultural domes and hydroponics stations feed millions across nearby systems.",
	"This system's strategic location makes it a valuable waypoint for traders.",
}

var outerDescriptions = []string{
	"A rugged frontier settlement where hardy colonists eke out a living.",
	"Distant from central authority, this system is a haven for independent traders and prospectors.",
	"Lawlessness and opportunity go hand in hand in this remote outpost.",
	"This barely-settled system sees more pirates than law enforcement patrols.",
	"A lonely outpost at the edge of civilized space, where self-reliance is everything.",
}

var edgeDescriptions = []string{
	"An alien world of incomprehensible architecture and technology.",
	"Mysterious signals emanate from the installations orbiting these strange planets.",
	"Few humans have visited this system and returned to tell the tale.",
	"The border of known space, where humanity meets the unknown.",
	"Advanced technology beyond human understanding is evident throughout this system.",
}

var independentDescriptions = []string{
	"An independent system that jealously guards its autonomy.",
	"Free from major faction control, this system charts its own course.",
	"A neutral ground where ships of all allegiances meet for trade.",
	"This system's fierce independence has kept major powers at bay.",
	"A hodgepodge of different cultures and peoples call this diverse system home.",
}

// GenerateDescription generates a system description based on government and distance
func GenerateDescription(governmentID string, distanceFromSol float64) string {
	descriptions := independentDescriptions

	switch governmentID {
	case "united_earth_federation", "republic_of_mars":
		if distanceFromSol < 30 {
			descriptions = coreDescriptions
		} else {
			descriptions = midDescriptions
		}
	case "free_traders_guild":
		descriptions = midDescriptions
	case "frontier_worlds":
		descriptions = outerDescriptions
	case "auroran_empire":
		descriptions = edgeDescriptions
	case "independent":
		descriptions = independentDescriptions
	}

	// Return random description from appropriate set
	return descriptions[rand.Intn(len(descriptions))]
}

// Planet description templates
var planetPrefixes = []string{
	"A rocky", "A barren", "A lush", "An icy", "A volcanic", "A desert",
	"A temperate", "A toxic", "A radiation-scarred", "A terraformed",
	"An oceanic", "A jungle-covered", "A mountainous", "A gaseous",
}

var planetMidparts = []string{
	"world", "planet", "moon", "dwarf planet", "terrestrial body",
}

var planetSuffixes = []string{
	"with a thin atmosphere.",
	"rich in mineral resources.",
	"hosting a thriving colony.",
	"barely suitable for habitation.",
	"under active terraforming.",
	"with ancient ruins dotting its surface.",
	"covered in sprawling cities.",
	"home to unique flora and fauna.",
	"with valuable ore deposits.",
	"serving as a military outpost.",
	"functioning as a research station.",
	"operating as a commercial hub.",
	"known for its agricultural output.",
	"famous for its shipyards.",
	"hosting a major spaceport.",
}

// GeneratePlanetDescription generates a planet/station description
func GeneratePlanetDescription(r *rand.Rand, isStation bool) string {
	if isStation {
		stationTypes := []string{
			"A massive orbital station serving as the system's commercial hub.",
			"A military starbase bristling with weapons and defenses.",
			"A research station dedicated to advanced scientific studies.",
			"A mining platform processing ore from nearby asteroids.",
			"A shipyard where vessels are constructed and repaired.",
			"A trading post where merchants from across the galaxy meet.",
			"A refueling depot crucial for long-range voyages.",
		}
		return stationTypes[r.Intn(len(stationTypes))]
	}

	prefix := planetPrefixes[r.Intn(len(planetPrefixes))]
	midpart := planetMidparts[r.Intn(len(planetMidparts))]
	suffix := planetSuffixes[r.Intn(len(planetSuffixes))]

	return fmt.Sprintf("%s %s %s", prefix, midpart, suffix)
}

// GeneratePlanetName generates a planet name (system name + letter)
func GeneratePlanetName(systemName string, index int, r *rand.Rand) string {
	// 30% chance to be a station instead of planet
	if r.Float64() < 0.3 {
		stationNames := []string{
			"Station", "Outpost", "Hub", "Terminal", "Bastion",
			"Citadel", "Haven", "Nexus", "Gateway", "Port",
		}
		name := stationNames[r.Intn(len(stationNames))]
		return fmt.Sprintf("%s %s", systemName, name)
	}

	// Standard planet naming: System Name + Roman numeral or letter
	if r.Float64() < 0.5 {
		// Roman numerals (I, II, III, IV)
		romanNumerals := []string{"I", "II", "III", "IV", "V", "VI"}
		if index < len(romanNumerals) {
			return fmt.Sprintf("%s %s", systemName, romanNumerals[index])
		}
	}

	// Letters (A, B, C, D)
	return fmt.Sprintf("%s %c", systemName, rune('A'+index))
}

// File: internal/game/universe/names.go
// Project: Terminal Velocity
// Description: Procedural universe generation: names
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

// Package universe provides procedural universe generation including star system names,
// planet names, and descriptive text generation.
//
// The name generation system uses multiple strategies to create diverse and realistic-sounding
// names for celestial bodies:
//   - Greek letter + Constellation combinations (e.g., "Alpha Centauri")
//   - Real star names from Earth's night sky (e.g., "Sirius", "Vega")
//   - Catalog-style designations (e.g., "Kepler-452", "Ross-128")
//   - Compound descriptive names (e.g., "New Haven", "Proxima Station")
//
// All generated names are guaranteed to be unique within a universe through collision detection
// and fallback mechanisms.
package universe

import (
	"fmt"
	"math/rand"
)

// NameGenerator generates unique star system and planet names using procedural algorithms.
//
// The generator maintains a registry of used names to ensure uniqueness and employs
// multiple naming strategies to create diverse nomenclature:
//
// Naming Strategies:
//   1. Greek + Constellation (25% chance) - Scientific naming convention used for many real stars
//   2. Real Star Names (25% chance) - Names from Earth's visible stars
//   3. Catalog Designation (25% chance) - Survey-style numbering (Kepler-N, Ross-N, etc.)
//   4. Compound Names (25% chance) - Descriptive combinations (New Haven, Omega Terminal)
//
// Thread Safety:
//
//	NameGenerator is NOT thread-safe. Each generator should be used by a single goroutine
//	or protected with external synchronization if shared across goroutines.
type NameGenerator struct {
	rand      *rand.Rand       // Seeded random number generator for reproducible names
	usedNames map[string]bool  // Registry of already-generated names to prevent duplicates
}

// NewNameGenerator creates a new name generator with the given random source.
//
// Parameters:
//   - r: Seeded random number generator for reproducible name generation across server restarts
//
// Returns:
//
//	A new NameGenerator ready to produce unique star system and planet names
//
// The generator starts with an empty name registry and will track all generated names
// to ensure uniqueness.
func NewNameGenerator(r *rand.Rand) *NameGenerator {
	return &NameGenerator{
		rand:      r,
		usedNames: make(map[string]bool),
	}
}

// greekLetters contains the 24 letters of the Greek alphabet used in the Bayer designation
// system for naming stars. In astronomy, the brightest star in a constellation is typically
// designated Alpha, the second brightest Beta, and so on.
//
// This list is used to generate scientifically-plausible star names when combined with
// constellation names (e.g., "Alpha Centauri", "Beta Orionis").
var greekLetters = []string{
	"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta", "Eta", "Theta",
	"Iota", "Kappa", "Lambda", "Mu", "Nu", "Xi", "Omicron", "Pi",
	"Rho", "Sigma", "Tau", "Upsilon", "Phi", "Chi", "Psi", "Omega",
}

// constellations contains genitive (possessive) forms of constellation names used in
// the Bayer designation system. For example, "Centauri" is the genitive of "Centaurus",
// so "Alpha Centauri" means "Alpha of Centaurus".
//
// These are combined with Greek letters to create realistic-sounding star system names
// following astronomical naming conventions.
var constellations = []string{
	"Centauri", "Eridani", "Ceti", "Draconis", "Leonis", "Aquarii", "Orionis",
	"Scorpii", "Cassiopeiae", "Andromedae", "Lyrae", "Cygni", "Aquilae",
	"Ursae", "Bootis", "Virginis", "Geminorum", "Tauri", "Sagittarii",
	"Capricorni", "Piscium", "Arietis", "Cancri", "Librae", "Persei",
	"Herculis", "Ophiuchi", "Serpentis", "Coronae", "Hydrae",
}

// realStars contains names of the brightest stars visible from Earth, including stars
// from various cultures and historical astronomical catalogs. These names add variety
// and familiarity to the generated universe.
//
// Many of these names have Arabic or Latin origins and are the traditional names
// used in modern astronomy (e.g., Sirius is the brightest star in Earth's night sky,
// Betelgeuse is the red supergiant in Orion).
var realStars = []string{
	"Sirius", "Canopus", "Arcturus", "Vega", "Capella", "Rigel", "Procyon",
	"Betelgeuse", "Achernar", "Altair", "Aldebaran", "Antares", "Spica",
	"Pollux", "Fomalhaut", "Deneb", "Regulus", "Adhara", "Castor", "Bellatrix",
	"Elnath", "Miaplacidus", "Alnilam", "Alnitak", "Alnair", "Alioth",
	"Dubhe", "Mirfak", "Wezen", "Sargas", "Kaus Australis", "Avior",
	"Alkaid", "Menkalinan", "Atria", "Alhena", "Peacock", "Alsephina",
	"Mirzam", "Alphard", "Hamal", "Polaris", "Alderamin", "Denebola",
}

// namePrefix contains prefixes for catalog-style and compound names. These include:
//   - Directional/Temporal: New, Neo, Nova, Prima, Proxima, Ultima
//   - Astronomer surnames: Kepler, Ross, Gliese, Wolf, Lacaille, Luyten, Barnard, etc.
//
// Astronomer names are used in real star catalogs (e.g., Barnard's Star, Ross 128).
var namePrefix = []string{
	"New", "Neo", "Nova", "Omega", "Proxima", "Ultima", "Prima", "Kepler",
	"Ross", "Gliese", "Wolf", "Lacaille", "Luyten", "Barnard", "Kruger",
	"Groombridge", "Lalande", "Struve", "Innes", "van", "Stein",
}

// nameSuffix contains suffixes for compound names, including:
//   - Latin ordinals: Prime, Secundus, Tertius
//   - Relative size: Major, Minor
//   - Settlement types: Station, Outpost, Haven, Refuge, Bastion
//   - Infrastructure: Gate, Nexus, Hub, Terminal, Crossing
//
// These create evocative names suggesting human colonization and space infrastructure.
var nameSuffix = []string{
	"Prime", "Secundus", "Tertius", "Major", "Minor", "Station", "Outpost",
	"Haven", "Refuge", "Bastion", "Forge", "Reach", "Crossing", "Gate",
	"Nexus", "Hub", "Point", "Junction", "Terminal", "Threshold",
}

// GenerateSystemName generates a unique star system name using one of four strategies.
//
// Algorithm:
//  1. Select naming strategy (25% chance each):
//     - Greek + Constellation: "Alpha Centauri"
//     - Real star name: "Sirius"
//     - Catalog designation: "Kepler-442"
//     - Compound name: "New Haven"
//  2. Check if generated name is unique
//  3. If collision detected, retry up to 100 times
//  4. If all retries fail, fall back to guaranteed unique "System-N" format
//
// Returns:
//
//	A unique system name that hasn't been used before in this generator
//
// The 100-attempt limit prevents infinite loops in edge cases where the name space
// is nearly exhausted, though this is unlikely in practice with thousands of possible
// combinations.
//
// Thread Safety: NOT thread-safe. Callers must serialize access if used concurrently.
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

// generateGreekConstellation generates a Bayer designation-style name by combining
// a Greek letter with a constellation genitive form.
//
// Examples: "Alpha Centauri", "Beta Orionis", "Gamma Draconis"
//
// This follows the real astronomical naming convention established by Johann Bayer
// in 1603, where stars are designated by Greek letters within their constellations.
//
// Returns:
//
//	A Greek letter + constellation name combination (e.g., "Alpha Centauri")
func (ng *NameGenerator) generateGreekConstellation() string {
	greek := greekLetters[ng.rand.Intn(len(greekLetters))]
	constellation := constellations[ng.rand.Intn(len(constellations))]
	return fmt.Sprintf("%s %s", greek, constellation)
}

// generateCatalogName generates a star catalog-style designation with an astronomer's
// name followed by a catalog number.
//
// Examples: "Kepler-442", "Ross-128", "Gliese-581"
//
// This mimics real astronomical catalogs like the Kepler catalog (exoplanet discoveries),
// Ross catalog (nearby stars), and Gliese catalog (nearby stars < 25 parsecs).
//
// Returns:
//
//	A catalog-style name with format "<Astronomer>-<Number>" where number is 1-9999
func (ng *NameGenerator) generateCatalogName() string {
	prefix := namePrefix[ng.rand.Intn(len(namePrefix))]
	number := ng.rand.Intn(9999) + 1
	return fmt.Sprintf("%s-%d", prefix, number)
}

// generateCompoundName generates a two-word compound name suggesting human settlement
// or infrastructure.
//
// Examples: "New Haven", "Proxima Station", "Nova Terminal"
//
// These names evoke colonization efforts and space infrastructure, giving systems
// a sense of human presence and purpose.
//
// Returns:
//
//	A compound name with format "<Prefix> <Suffix>"
func (ng *NameGenerator) generateCompoundName() string {
	prefix := namePrefix[ng.rand.Intn(len(namePrefix))]
	suffix := nameSuffix[ng.rand.Intn(len(nameSuffix))]
	return fmt.Sprintf("%s %s", prefix, suffix)
}

// generateFallbackName generates a guaranteed unique name using sequential numbering.
//
// This is used as a last resort when all other naming strategies have failed to produce
// a unique name after 100 attempts. The counter-based approach ensures uniqueness by
// using the size of the usedNames map as an incrementing ID.
//
// Format: "System-<N>" where N is the count of previously generated names
//
// Returns:
//
//	A sequential system name that is guaranteed to be unique
//
// The generated name is immediately added to the usedNames registry to prevent
// future collisions.
func (ng *NameGenerator) generateFallbackName() string {
	counter := len(ng.usedNames)
	name := fmt.Sprintf("System-%d", counter)
	ng.usedNames[name] = true
	return name
}

// coreDescriptions contains flavor text for systems in the galactic core (within 30 LY of Sol).
// These systems are highly developed with advanced technology, large populations, and
// significant infrastructure. They represent the heart of human civilization in space.
var coreDescriptions = []string{
	"A highly developed core world with massive orbital installations and billions of inhabitants.",
	"Capital of a sector, this system hosts impressive military and civilian infrastructure.",
	"A wealthy industrial hub with state-of-the-art shipyards and manufacturing facilities.",
	"Home to one of humanity's most prestigious universities and research centers.",
	"A major trade nexus where goods from across the galaxy change hands.",
}

// midDescriptions contains flavor text for mid-range systems (30-60 LY from Sol).
// These are colonized but less developed than core worlds, focusing on trade,
// mining, and agriculture to support the growing human presence in space.
var midDescriptions = []string{
	"A prosperous trade station serves as the heart of this busy system.",
	"Mining operations and refineries dot the asteroid belts of this resource-rich system.",
	"A growing colonial world striving to match the prosperity of the core systems.",
	"Agricultural domes and hydroponics stations feed millions across nearby systems.",
	"This system's strategic location makes it a valuable waypoint for traders.",
}

// outerDescriptions contains flavor text for frontier systems (60-100 LY from Sol).
// These remote systems are barely settled, dangerous, and far from central authority.
// They represent the edge of civilized space where law enforcement is scarce.
var outerDescriptions = []string{
	"A rugged frontier settlement where hardy colonists eke out a living.",
	"Distant from central authority, this system is a haven for independent traders and prospectors.",
	"Lawlessness and opportunity go hand in hand in this remote outpost.",
	"This barely-settled system sees more pirates than law enforcement patrols.",
	"A lonely outpost at the edge of civilized space, where self-reliance is everything.",
}

// edgeDescriptions contains flavor text for edge systems (> 100 LY from Sol).
// These mysterious systems are controlled by the Auroran Empire and feature
// incomprehensible alien technology beyond human understanding.
var edgeDescriptions = []string{
	"An alien world of incomprehensible architecture and technology.",
	"Mysterious signals emanate from the installations orbiting these strange planets.",
	"Few humans have visited this system and returned to tell the tale.",
	"The border of known space, where humanity meets the unknown.",
	"Advanced technology beyond human understanding is evident throughout this system.",
}

// independentDescriptions contains flavor text for independent systems not controlled
// by major factions. These systems maintain neutrality and autonomy, serving as
// neutral meeting grounds for all factions.
var independentDescriptions = []string{
	"An independent system that jealously guards its autonomy.",
	"Free from major faction control, this system charts its own course.",
	"A neutral ground where ships of all allegiances meet for trade.",
	"This system's fierce independence has kept major powers at bay.",
	"A hodgepodge of different cultures and peoples call this diverse system home.",
}

// GenerateDescription generates a procedural description for a star system based on
// its governing faction and distance from Sol.
//
// Algorithm:
//  1. Determine description pool based on government type:
//     - UEF/ROM: Use distance-based descriptions (core/mid)
//     - Free Traders Guild: Use midDescriptions
//     - Frontier Worlds: Use outerDescriptions
//     - Auroran Empire: Use edgeDescriptions
//     - Independent: Use independentDescriptions
//  2. Randomly select one description from the appropriate pool
//
// Parameters:
//   - governmentID: The ID of the faction controlling this system
//   - distanceFromSol: Distance in light-years from Sol (Earth's system)
//
// Returns:
//
//	A procedurally-selected description string appropriate for the system's
//	location and political affiliation
//
// The distance thresholds (30 LY, 60 LY, 100 LY) correspond to the core, mid, outer,
// and edge radius configuration values used during universe generation.
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

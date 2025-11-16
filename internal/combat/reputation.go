// File: internal/combat/reputation.go
// Project: Terminal Velocity
// Description: Combat system: reputation - Faction reputation and bounty system
// Version: 1.1.0
// Author: Joshua Ferguson
// Created: 2025-01-07

// Package combat provides the reputation and legal system for faction relationships.
//
// This file implements reputation changes from combat actions and bounty tracking:
//   - Reputation changes from combat events (-100 to +100 scale)
//   - Legal status tracking (clean, offender, wanted, fugitive)
//   - Bounty system with expiration
//   - Faction reinforcement mechanics
//   - Cascading reputation effects (allies/enemies)
//
// Reputation Scale:
//   - +75 to +100: Allied (maximum cooperation)
//   - +25 to +74: Friendly (favorable treatment)
//   - -24 to +24: Neutral (standard interactions)
//   - -49 to -25: Unfriendly (hostility, higher prices)
//   - -74 to -50: Hostile (attacked on sight)
//   - -100 to -75: At War (kill on sight, reinforcements)
//
// Reputation Events and Changes:
//   - Kill Hostile: +5 rep with faction's enemies
//   - Kill Ally: -25 rep with faction (-35 if player was friendly)
//   - Kill Neutral: -15 rep with faction
//   - Kill Civilian: -30 rep with faction, -10 with all lawful factions
//   - Defend Ally: +10 rep with faction
//   - Bounty Collected: +8 rep with issuing faction
//
// Thread-safety: Functions are stateless and safe for concurrent calls.
package combat

import (
	"fmt"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
)

// ReputationChange represents a single reputation modification with a faction.
//
// Reputation changes are generated from combat events and applied to the player's
// standing with various factions. Multiple changes can result from a single action
// due to cascading effects through faction alliances and rivalries.
//
// Fields:
//   - FactionID: Faction whose reputation is changing
//   - Amount: Reputation delta (negative = loss, positive = gain)
//   - Reason: Human-readable explanation for the change
type ReputationChange struct {
	FactionID string
	Amount    int
	Reason    string
}

// BountyInfo tracks an active bounty placed on a player by a faction.
//
// Bounties are placed when players commit crimes against a faction. They can be
// collected by other players or NPCs, or paid off at a station for 1.5x the amount.
//
// Fields:
//   - FactionID: Faction that issued the bounty
//   - Amount: Bounty value in credits
//   - Reason: Description of the crime
//   - Expires: Unix timestamp when bounty expires (0 = never)
type BountyInfo struct {
	FactionID string
	Amount    int64
	Reason    string
	Expires   int64 // Unix timestamp
}

// LegalStatus represents a player's legal standing with a specific faction.
//
// Each faction tracks the player's criminal record independently. Status
// degrades with crimes and can improve over time or by paying off bounties.
//
// Status Levels:
//   - "clean": No criminal record
//   - "offender": Minor crimes committed
//   - "wanted": Significant criminal activity
//   - "fugitive": Severe crimes, hunted actively
//
// Fields:
//   - FactionID: Faction tracking this legal status
//   - Status: Current legal standing
//   - CrimesCount: Number of crimes committed
//   - LastOffense: Timestamp of most recent crime
//   - ActiveBounty: Current bounty if any
type LegalStatus struct {
	FactionID    string
	Status       string // "clean", "offender", "wanted", "fugitive"
	CrimesCount  int    // Number of crimes committed
	LastOffense  int64  // Unix timestamp
	ActiveBounty *BountyInfo
}

// ReputationEvent represents the type of combat event that affects reputation.
//
// These constants define all combat actions that trigger reputation changes.
// Each event type has specific reputation effects on the victim's faction,
// their allies, and their enemies.
type ReputationEvent string

const (
	EventKillHostile   ReputationEvent = "kill_hostile"   // Killed enemy of faction
	EventKillNeutral   ReputationEvent = "kill_neutral"   // Killed neutral ship
	EventKillAlly      ReputationEvent = "kill_ally"      // Killed faction ship or ally
	EventKillCivilian  ReputationEvent = "kill_civilian"  // Killed civilian ship
	EventDefendAlly    ReputationEvent = "defend_ally"    // Defended faction ship
	EventPirateAction  ReputationEvent = "pirate_action"  // Attacked lawful ship
	EventBountyPaid    ReputationEvent = "bounty_paid"    // Collected bounty
	EventBountyCleared ReputationEvent = "bounty_cleared" // Cleared bounty
)

// CalculateCombatReputation calculates all reputation changes from a combat event.
//
// This function generates reputation changes for the player based on their actions
// in combat. A single combat event can affect reputation with multiple factions due
// to alliance and rivalry relationships.
//
// Cascading Effects:
//   - Attacking a faction also affects their allies (penalty) and enemies (bonus)
//   - Helping a faction also benefits their allies
//   - Civilian attacks anger all lawful factions
//
// Special Modifiers:
//   - Low reputation provides bonus gains when helping (+2 bonus if negative rep)
//   - High reputation increases penalties for betrayal (-10 extra if >+25 rep)
//
// Parameters:
//   - event: Type of combat event (kill hostile, kill ally, etc.)
//   - victimFactionID: Faction of the ship involved in the event
//   - attackerReputation: Player's current reputation with victim faction
//
// Returns:
//   - []ReputationChange: All reputation changes to apply (can be multiple factions)
//
// Thread-safe: No shared state modification, safe for concurrent calls.
func CalculateCombatReputation(
	event ReputationEvent,
	victimFactionID string,
	attackerReputation int,
) []ReputationChange {

	changes := []ReputationChange{}

	victimFaction := models.GetFactionByID(victimFactionID)
	if victimFaction == nil {
		return changes
	}

	switch event {
	case EventKillHostile:
		// Killing an enemy of the faction
		// Gain reputation with all factions hostile to victim
		for _, factionID := range victimFaction.Enemies {
			amount := 5 // Base gain
			// Bonus for low reputation (help build trust)
			if attackerReputation < 0 {
				amount += 2
			}
			changes = append(changes, ReputationChange{
				FactionID: factionID,
				Amount:    amount,
				Reason:    fmt.Sprintf("Destroyed %s vessel", victimFaction.ShortName),
			})
		}

	case EventKillAlly:
		// Killing a faction ship or their ally
		// Major reputation loss with victim faction
		amount := -25
		// Larger penalty for positive reputation (betrayal)
		if attackerReputation > 25 {
			amount -= 10
		}
		changes = append(changes, ReputationChange{
			FactionID: victimFactionID,
			Amount:    amount,
			Reason:    fmt.Sprintf("Destroyed %s vessel (hostile act)", victimFaction.ShortName),
		})

		// Loss with all allies of victim
		for _, allyID := range victimFaction.Allies {
			changes = append(changes, ReputationChange{
				FactionID: allyID,
				Amount:    -15,
				Reason:    fmt.Sprintf("Attacked %s ally", victimFaction.ShortName),
			})
		}

		// Small gain with enemies
		for _, enemyID := range victimFaction.Enemies {
			changes = append(changes, ReputationChange{
				FactionID: enemyID,
				Amount:    2,
				Reason:    fmt.Sprintf("Attacked %s", victimFaction.ShortName),
			})
		}

	case EventKillNeutral:
		// Killing neutral ship
		// Moderate reputation loss with victim faction
		changes = append(changes, ReputationChange{
			FactionID: victimFactionID,
			Amount:    -15,
			Reason:    fmt.Sprintf("Destroyed %s vessel (unprovoked)", victimFaction.ShortName),
		})

		// Small loss with allies
		for _, allyID := range victimFaction.Allies {
			changes = append(changes, ReputationChange{
				FactionID: allyID,
				Amount:    -5,
				Reason:    fmt.Sprintf("Attacked %s", victimFaction.ShortName),
			})
		}

	case EventKillCivilian:
		// Killing civilian/merchant ship
		// Severe reputation loss
		changes = append(changes, ReputationChange{
			FactionID: victimFactionID,
			Amount:    -30,
			Reason:    fmt.Sprintf("Destroyed %s civilian vessel (piracy)", victimFaction.ShortName),
		})

		// Loss with all non-hostile factions
		for _, faction := range models.StandardNPCFactions {
			if faction.ID != victimFactionID && !faction.IsHostileTo(victimFactionID) {
				changes = append(changes, ReputationChange{
					FactionID: faction.ID,
					Amount:    -10,
					Reason:    "Piracy against civilians",
				})
			}
		}

	case EventDefendAlly:
		// Defended faction ship from attack
		changes = append(changes, ReputationChange{
			FactionID: victimFactionID,
			Amount:    10,
			Reason:    fmt.Sprintf("Defended %s vessel", victimFaction.ShortName),
		})

	case EventPirateAction:
		// Attacked lawful ship without provocation
		changes = append(changes, ReputationChange{
			FactionID: victimFactionID,
			Amount:    -20,
			Reason:    "Piracy",
		})

	case EventBountyPaid:
		// Collected bounty
		changes = append(changes, ReputationChange{
			FactionID: victimFactionID,
			Amount:    8,
			Reason:    "Bounty collected",
		})
	}

	return changes
}

// GetLegalStatusName returns the human-readable legal status
func GetLegalStatusName(status string) string {
	switch status {
	case "clean":
		return "Clean Record"
	case "offender":
		return "Offender"
	case "wanted":
		return "Wanted"
	case "fugitive":
		return "Fugitive"
	default:
		return "Unknown"
	}
}

// UpdateLegalStatus updates legal status based on crime severity
func UpdateLegalStatus(current *LegalStatus, crimeSeverity int) {
	current.CrimesCount++

	// Determine new status based on crimes and severity
	totalSeverity := current.CrimesCount * crimeSeverity

	switch {
	case totalSeverity >= 100:
		current.Status = "fugitive"
	case totalSeverity >= 50:
		current.Status = "wanted"
	case totalSeverity >= 10:
		current.Status = "offender"
	default:
		current.Status = "clean"
	}
}

// CalculateBountyAmount calculates the bounty placed on player for a crime.
//
// Bounty amounts vary based on crime severity and victim ship value:
//   - Base bounty: 10,000 credits minimum
//   - Ship value bonus: 20% of ship value
//   - Multipliers by crime type:
//     * Kill Civilian: 3.0x (severe)
//     * Kill Ally: 2.0x (major)
//     * Kill Neutral: 1.5x (moderate)
//     * Piracy: 1.0x (standard)
//
// Parameters:
//   - event: Type of crime committed
//   - shipValue: Value of victim ship (affects bounty size)
//
// Returns:
//   - int64: Bounty amount in credits (0 if event doesn't warrant bounty)
//
// Thread-safe: No shared state, safe for concurrent calls.
func CalculateBountyAmount(event ReputationEvent, shipValue int64) int64 {
	var multiplier float64

	switch event {
	case EventKillCivilian:
		multiplier = 3.0 // Severe bounty
	case EventKillAlly:
		multiplier = 2.0 // Major bounty
	case EventKillNeutral:
		multiplier = 1.5 // Moderate bounty
	case EventPirateAction:
		multiplier = 1.0 // Standard bounty
	default:
		return 0 // No bounty
	}

	baseBounty := int64(10000) // Minimum bounty
	shipBounty := int64(float64(shipValue) * 0.2 * multiplier)

	return baseBounty + shipBounty
}

// GetHostilityLevel converts reputation score to hostility status.
//
// Reputation Ranges:
//   - 75-100: "allied"
//   - 25-74: "friendly"
//   - -24-24: "neutral"
//   - -49--25: "unfriendly"
//   - -74--50: "hostile"
//   - -100--75: "at_war"
//
// Parameters:
//   - reputation: Reputation score (-100 to +100)
//
// Returns:
//   - string: Hostility level identifier
//
// Thread-safe: No shared state, safe for concurrent calls.
func GetHostilityLevel(reputation int) string {
	switch {
	case reputation >= 75:
		return "allied"
	case reputation >= 25:
		return "friendly"
	case reputation > -25:
		return "neutral"
	case reputation > -50:
		return "unfriendly"
	case reputation > -75:
		return "hostile"
	default:
		return "at_war"
	}
}

// WillFactionsReinforce determines if a faction will send reinforcements to combat.
//
// Reinforcements are sent when:
//   1. Combat is in faction's territory or allied territory
//   2. Player has negative reputation with faction
//   3. Sufficient combat turns have elapsed (delay depends on reputation)
//
// Reinforcement Delays:
//   - Reputation <-50: 2 turns (high priority)
//   - Reputation -50 to -25: 4 turns (medium priority)
//   - Reputation -25 to 0: 6 turns (low priority)
//   - Reputation >0: No reinforcements (friendly)
//
// Parameters:
//   - factionID: Faction to check for reinforcements
//   - reputation: Player's reputation with faction
//   - systemControllingFaction: Faction that controls current system
//   - combatTurns: Number of turns combat has lasted
//
// Returns:
//   - bool: true if reinforcements will arrive this turn
//
// Thread-safe: No shared state, safe for concurrent calls.
func WillFactionsReinforce(
	factionID string,
	reputation int,
	systemControllingFaction string,
	combatTurns int,
) bool {

	// Must be in their territory or allied territory
	faction := models.GetFactionByID(factionID)
	if faction == nil {
		return false
	}

	inTerritory := factionID == systemControllingFaction
	inAlliedTerritory := faction.IsAlliedWith(systemControllingFaction)

	if !inTerritory && !inAlliedTerritory {
		return false
	}

	// More likely if player has bad reputation
	if reputation < -50 {
		// High priority reinforcement
		return combatTurns >= 2
	} else if reputation < -25 {
		// Medium priority
		return combatTurns >= 4
	} else if reputation < 0 {
		// Low priority
		return combatTurns >= 6
	}

	// Won't reinforce against friendly player
	return false
}

// CalculateReinforcementStrength determines how many ships to send
func CalculateReinforcementStrength(
	factionID string,
	reputation int,
	playerShipValue int64,
) int {

	faction := models.GetFactionByID(factionID)
	if faction == nil {
		return 0
	}

	// Base on faction patrol strength
	baseStrength := faction.PatrolStrength

	// Adjust for reputation (worse rep = more ships)
	repMultiplier := 1.0
	if reputation < -75 {
		repMultiplier = 2.0
	} else if reputation < -50 {
		repMultiplier = 1.5
	} else if reputation < -25 {
		repMultiplier = 1.2
	}

	// Calculate number of ships (1-5 based on strength)
	ships := int(float64(baseStrength) * repMultiplier / 3.0)
	if ships < 1 {
		ships = 1
	}
	if ships > 5 {
		ships = 5
	}

	return ships
}

// GetReinforcementDelay returns turns before reinforcements arrive
func GetReinforcementDelay(factionPatrolStrength int) int {
	// Stronger factions respond faster
	switch {
	case factionPatrolStrength >= 8:
		return 2 // 2 turns
	case factionPatrolStrength >= 6:
		return 3 // 3 turns
	case factionPatrolStrength >= 4:
		return 4 // 4 turns
	default:
		return 5 // 5 turns
	}
}

// ApplyReputationChanges applies multiple reputation changes to player's standings.
//
// This function updates the player's reputation map with all changes from a combat
// event. Reputation values are clamped between -100 and +100.
//
// Parameters:
//   - playerReputation: Current reputation map (faction ID -> reputation score)
//   - changes: List of reputation changes to apply
//
// Returns:
//   - map[string]int: Updated reputation map with all changes applied
//
// Side Effects:
//   - Modifies playerReputation map in place
//   - Also returns the modified map for convenience
//
// Thread-safe: Modifies only passed parameter, safe for single player context.
func ApplyReputationChanges(
	playerReputation map[string]int,
	changes []ReputationChange,
) map[string]int {

	if playerReputation == nil {
		playerReputation = make(map[string]int)
	}

	for _, change := range changes {
		current := playerReputation[change.FactionID]
		newRep := current + change.Amount

		// Clamp reputation between -100 and 100
		if newRep > 100 {
			newRep = 100
		}
		if newRep < -100 {
			newRep = -100
		}

		playerReputation[change.FactionID] = newRep
	}

	return playerReputation
}

// GetReputationChangeMessage formats a reputation change message
func GetReputationChangeMessage(change ReputationChange) string {
	faction := models.GetFactionByID(change.FactionID)
	factionName := change.FactionID
	if faction != nil {
		factionName = faction.ShortName
	}

	direction := "increased"
	if change.Amount < 0 {
		direction = "decreased"
	}

	return fmt.Sprintf("%s reputation %s by %d: %s",
		factionName, direction, abs(change.Amount), change.Reason)
}

// abs returns absolute value of an integer
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// IsBountyActive checks if a bounty is still active
func IsBountyActive(bounty *BountyInfo, currentTime int64) bool {
	if bounty == nil {
		return false
	}
	return bounty.Expires > currentTime
}

// GetActiveBounties returns all active bounties for a player
func GetActiveBounties(legalStatuses []*LegalStatus, currentTime int64) []*BountyInfo {
	var active []*BountyInfo

	for _, status := range legalStatuses {
		if status.ActiveBounty != nil && IsBountyActive(status.ActiveBounty, currentTime) {
			active = append(active, status.ActiveBounty)
		}
	}

	return active
}

// GetTotalBountyValue calculates total bounty value across all factions
func GetTotalBountyValue(bounties []*BountyInfo) int64 {
	var total int64
	for _, bounty := range bounties {
		total += bounty.Amount
	}
	return total
}

// CanPayOffBounty checks if player can afford to pay off a bounty
func CanPayOffBounty(playerCredits int64, bountyAmount int64) bool {
	// Paying off costs 1.5x the bounty (bribe)
	cost := int64(float64(bountyAmount) * 1.5)
	return playerCredits >= cost
}

// PayOffBountyCost calculates the cost to pay off a bounty
func PayOffBountyCost(bountyAmount int64) int64 {
	return int64(float64(bountyAmount) * 1.5)
}

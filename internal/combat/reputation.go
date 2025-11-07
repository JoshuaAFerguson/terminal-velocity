// File: internal/combat/reputation.go
// Project: Terminal Velocity
// Description: Combat system: reputation
// Version: 1.0.0
// Author: Joshua Ferguson
// Created: 2025-01-07

package combat

import (
	"fmt"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
)

// ReputationChange represents a change in faction reputation
type ReputationChange struct {
	FactionID string
	Amount    int
	Reason    string
}

// BountyInfo tracks bounties on a player
type BountyInfo struct {
	FactionID string
	Amount    int64
	Reason    string
	Expires   int64 // Unix timestamp
}

// LegalStatus represents a player's legal standing with a faction
type LegalStatus struct {
	FactionID    string
	Status       string // "clean", "offender", "wanted", "fugitive"
	CrimesCount  int    // Number of crimes committed
	LastOffense  int64  // Unix timestamp
	ActiveBounty *BountyInfo
}

// ReputationEvent represents the type of combat event that affects reputation
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

// CalculateCombatReputation calculates reputation changes from combat
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

// CalculateBountyAmount calculates bounty amount for a crime
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

// GetHostilityLevel returns hostility level based on reputation
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

// WillFactionsReinforce checks if a faction will send reinforcements
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

// ApplyReputationChanges applies a list of reputation changes to a player
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

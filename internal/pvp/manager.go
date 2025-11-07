// File: internal/pvp/manager.go
// Project: Terminal Velocity
// Version: 1.0.0

package pvp

import (
	"errors"
	"sync"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/models"
	"github.com/google/uuid"
)

var (
	ErrChallengeNotFound    = errors.New("challenge not found")
	ErrNotAuthorized        = errors.New("not authorized for this action")
	ErrChallengeExpired     = errors.New("challenge has expired")
	ErrInvalidStatus        = errors.New("invalid challenge status for this operation")
	ErrPlayerNotFound       = errors.New("player not found")
	ErrBountyNotFound       = errors.New("bounty not found")
	ErrInsufficientFunds    = errors.New("insufficient funds for wager")
	ErrNotInSameSystem      = errors.New("players must be in same system")
	ErrCannotAttackSelf     = errors.New("cannot attack yourself")
)

// Manager handles all PvP combat operations
type Manager struct {
	mu         sync.RWMutex
	challenges map[uuid.UUID]*models.PvPChallenge       // Challenge ID -> Challenge
	byPlayer   map[uuid.UUID][]*models.PvPChallenge     // Player ID -> Challenges
	bounties   map[uuid.UUID]*models.Bounty             // Target ID -> Bounty
	stats      map[uuid.UUID]*models.PvPStats           // Player ID -> Stats
	results    []*models.PvPCombatResult                // Combat history
}

// NewManager creates a new PvP manager
func NewManager() *Manager {
	return &Manager{
		challenges: make(map[uuid.UUID]*models.PvPChallenge),
		byPlayer:   make(map[uuid.UUID][]*models.PvPChallenge),
		bounties:   make(map[uuid.UUID]*models.Bounty),
		stats:      make(map[uuid.UUID]*models.PvPStats),
		results:    []*models.PvPCombatResult{},
	}
}

// CreateChallenge creates a new PvP challenge
func (m *Manager) CreateChallenge(
	challengerID uuid.UUID,
	challengerName string,
	defenderID uuid.UUID,
	defenderName string,
	challengeType models.PvPChallengeType,
	systemID uuid.UUID,
	wager int64,
	message string,
) (*models.PvPChallenge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Cannot challenge yourself
	if challengerID == defenderID {
		return nil, ErrCannotAttackSelf
	}

	challenge := models.NewPvPChallenge(
		challengerID,
		challengerName,
		defenderID,
		defenderName,
		challengeType,
		systemID,
	)

	challenge.Wager = wager
	challenge.Message = message

	m.challenges[challenge.ID] = challenge
	m.byPlayer[challengerID] = append(m.byPlayer[challengerID], challenge)
	m.byPlayer[defenderID] = append(m.byPlayer[defenderID], challenge)

	// Ensure both players have stats
	m.ensureStats(challengerID)
	m.ensureStats(defenderID)

	return challenge, nil
}

// GetChallenge retrieves a challenge by ID
func (m *Manager) GetChallenge(challengeID uuid.UUID) (*models.PvPChallenge, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	challenge, exists := m.challenges[challengeID]
	if !exists {
		return nil, ErrChallengeNotFound
	}

	// Auto-expire if needed
	if challenge.IsExpired() {
		challenge.Status = models.ChallengeExpired
	}

	return challenge, nil
}

// GetPlayerChallenges returns all challenges for a player
func (m *Manager) GetPlayerChallenges(playerID uuid.UUID) []*models.PvPChallenge {
	m.mu.RLock()
	defer m.mu.RUnlock()

	challenges := m.byPlayer[playerID]

	// Auto-expire stale challenges
	for _, challenge := range challenges {
		if challenge.IsExpired() {
			challenge.Status = models.ChallengeExpired
		}
	}

	return challenges
}

// GetPendingChallenges returns pending challenges awaiting player response
func (m *Manager) GetPendingChallenges(playerID uuid.UUID) []*models.PvPChallenge {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pending := []*models.PvPChallenge{}
	for _, challenge := range m.byPlayer[playerID] {
		if challenge.IsExpired() {
			challenge.Status = models.ChallengeExpired
		}
		if challenge.Status == models.ChallengePending && challenge.DefenderID == playerID {
			pending = append(pending, challenge)
		}
	}

	return pending
}

// AcceptChallenge accepts a combat challenge
func (m *Manager) AcceptChallenge(challengeID uuid.UUID, playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	challenge, exists := m.challenges[challengeID]
	if !exists {
		return ErrChallengeNotFound
	}

	// Only defender can accept
	if challenge.DefenderID != playerID {
		return ErrNotAuthorized
	}

	// Check status
	if challenge.Status != models.ChallengePending {
		return ErrInvalidStatus
	}

	// Check expiry
	if challenge.IsExpired() {
		challenge.Status = models.ChallengeExpired
		return ErrChallengeExpired
	}

	challenge.Accept()
	challenge.Start() // Auto-start after acceptance

	return nil
}

// DeclineChallenge declines a combat challenge
func (m *Manager) DeclineChallenge(challengeID uuid.UUID, playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	challenge, exists := m.challenges[challengeID]
	if !exists {
		return ErrChallengeNotFound
	}

	// Only defender can decline
	if challenge.DefenderID != playerID {
		return ErrNotAuthorized
	}

	// Check status
	if challenge.Status != models.ChallengePending {
		return ErrInvalidStatus
	}

	challenge.Decline()

	return nil
}

// CompleteCombat completes a combat and records the result
func (m *Manager) CompleteCombat(
	challengeID uuid.UUID,
	winnerID uuid.UUID,
	creditsTransfer int64,
	winnerDamage int,
	loserDamage int,
) (*models.PvPCombatResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	challenge, exists := m.challenges[challengeID]
	if !exists {
		return nil, ErrChallengeNotFound
	}

	// Must be active
	if challenge.Status != models.ChallengeActive {
		return nil, ErrInvalidStatus
	}

	// Complete the challenge
	challenge.Complete(winnerID)

	// Determine loser
	loserID := challenge.DefenderID
	if winnerID == challenge.DefenderID {
		loserID = challenge.ChallengerID
	}

	// Create result
	result := &models.PvPCombatResult{
		ChallengeID:     challengeID,
		WinnerID:        winnerID,
		LoserID:         loserID,
		WinnerDamage:    winnerDamage,
		LoserDamage:     loserDamage,
		CreditsWon:      creditsTransfer + challenge.Wager,
		CreditsLost:     creditsTransfer + challenge.Wager,
		CargoLooted:     make(map[string]int),
		WinnerRepChange: 10,
		LoserRepChange:  -5,
		BountyAdded:     0,
		BountyClaimed:   0,
		Duration:        challenge.EndedAt.Sub(*challenge.StartedAt),
		Timestamp:       *challenge.EndedAt,
	}

	// Handle bounty effects
	if challenge.Type == models.ChallengeAggression {
		// Aggressor gets bounty if they win
		if winnerID == challenge.ChallengerID {
			bounty, exists := m.bounties[winnerID]
			if !exists {
				bounty = models.NewBounty(winnerID, challenge.ChallengerName, 10000, "Piracy", "System")
				m.bounties[winnerID] = bounty
			} else {
				bounty.AddCrime("kill", creditsTransfer)
			}
			result.BountyAdded = 10000
		}
	} else if challenge.Type == models.ChallengeBountyHunt {
		// Bounty hunter claims bounty if they win
		if bounty, exists := m.bounties[loserID]; exists && bounty.Active {
			bounty.Claim(winnerID)
			result.BountyClaimed = bounty.Amount
			result.CreditsWon += bounty.Amount
		}
	}

	// Update stats
	m.ensureStats(winnerID)
	m.ensureStats(loserID)

	m.stats[winnerID].RecordWin(challenge.Type, result.CreditsWon, int64(winnerDamage))
	m.stats[loserID].RecordLoss(result.CreditsLost, int64(loserDamage))

	// Store result
	m.results = append(m.results, result)

	return result, nil
}

// IssueBounty issues a bounty on a player
func (m *Manager) IssueBounty(targetID uuid.UUID, targetName string, amount int64, reason string, issuedBy string) *models.Bounty {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if bounty already exists
	if bounty, exists := m.bounties[targetID]; exists && bounty.Active {
		// Add to existing bounty
		bounty.Amount += amount
		bounty.CrimeValue += amount
		return bounty
	}

	// Create new bounty
	bounty := models.NewBounty(targetID, targetName, amount, reason, issuedBy)
	m.bounties[targetID] = bounty

	// Update stats
	m.ensureStats(targetID)
	m.stats[targetID].NotorietyLevel++

	return bounty
}

// GetBounty retrieves a bounty for a player
func (m *Manager) GetBounty(targetID uuid.UUID) (*models.Bounty, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	bounty, exists := m.bounties[targetID]
	if !exists || !bounty.Active {
		return nil, false
	}

	// Check expiry
	if bounty.IsExpired() {
		bounty.Active = false
		return nil, false
	}

	return bounty, true
}

// GetAllActiveBounties returns all active bounties
func (m *Manager) GetAllActiveBounties() []*models.Bounty {
	m.mu.RLock()
	defer m.mu.RUnlock()

	active := []*models.Bounty{}
	for _, bounty := range m.bounties {
		if bounty.Active && !bounty.IsExpired() {
			active = append(active, bounty)
		} else if bounty.IsExpired() {
			bounty.Active = false
		}
	}

	return active
}

// GetStats retrieves PvP stats for a player
func (m *Manager) GetStats(playerID uuid.UUID) *models.PvPStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if stats, exists := m.stats[playerID]; exists {
		return stats
	}

	return models.NewPvPStats(playerID)
}

// GetRecentResults returns recent combat results
func (m *Manager) GetRecentResults(limit int) []*models.PvPCombatResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.results) <= limit {
		return m.results
	}

	return m.results[len(m.results)-limit:]
}

// GetPlayerResults returns combat results for a specific player
func (m *Manager) GetPlayerResults(playerID uuid.UUID, limit int) []*models.PvPCombatResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := []*models.PvPCombatResult{}
	for i := len(m.results) - 1; i >= 0 && len(results) < limit; i-- {
		result := m.results[i]
		if result.WinnerID == playerID || result.LoserID == playerID {
			results = append(results, result)
		}
	}

	return results
}

// CleanupExpiredChallenges removes expired challenges
func (m *Manager) CleanupExpiredChallenges() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	cleaned := 0
	for id, challenge := range m.challenges {
		if challenge.IsExpired() {
			challenge.Status = models.ChallengeExpired
			cleaned++
		}

		// Remove very old completed challenges (older than 24 hours)
		if challenge.Status == models.ChallengeComplete &&
			challenge.EndedAt != nil &&
			challenge.EndedAt.Add(24*60*60*1000000000).Before(*challenge.EndedAt) {
			delete(m.challenges, id)

			// Remove from player lists
			m.removeFromPlayerList(challenge.ChallengerID, id)
			m.removeFromPlayerList(challenge.DefenderID, id)
		}
	}

	return cleaned
}

// CleanupExpiredBounties removes expired bounties
func (m *Manager) CleanupExpiredBounties() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	cleaned := 0
	for _, bounty := range m.bounties {
		if bounty.IsExpired() && bounty.Active {
			bounty.Active = false
			cleaned++
		}
	}

	return cleaned
}

// ensureStats creates stats if they don't exist
func (m *Manager) ensureStats(playerID uuid.UUID) {
	if _, exists := m.stats[playerID]; !exists {
		m.stats[playerID] = models.NewPvPStats(playerID)
	}
}

// removeFromPlayerList removes a challenge from a player's list
func (m *Manager) removeFromPlayerList(playerID uuid.UUID, challengeID uuid.UUID) {
	challenges := m.byPlayer[playerID]
	for i, challenge := range challenges {
		if challenge.ID == challengeID {
			m.byPlayer[playerID] = append(challenges[:i], challenges[i+1:]...)
			break
		}
	}
}

// GetLeaderboard returns top PvP players by combat rating
func (m *Manager) GetLeaderboard(limit int) []*models.PvPStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect all stats
	allStats := make([]*models.PvPStats, 0, len(m.stats))
	for _, stats := range m.stats {
		allStats = append(allStats, stats)
	}

	// Simple bubble sort by combat rating (good enough for small lists)
	for i := 0; i < len(allStats)-1; i++ {
		for j := 0; j < len(allStats)-i-1; j++ {
			if allStats[j].CombatRating < allStats[j+1].CombatRating {
				allStats[j], allStats[j+1] = allStats[j+1], allStats[j]
			}
		}
	}

	// Return top N
	if len(allStats) > limit {
		return allStats[:limit]
	}

	return allStats
}

// GetStats returns overall PvP statistics
func (m *Manager) GetSystemStats() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]int{
		"total_challenges":  len(m.challenges),
		"active_challenges": 0,
		"pending_challenges": 0,
		"active_bounties":   0,
		"total_combats":     len(m.results),
	}

	for _, challenge := range m.challenges {
		switch challenge.Status {
		case models.ChallengeActive:
			stats["active_challenges"]++
		case models.ChallengePending:
			stats["pending_challenges"]++
		}
	}

	for _, bounty := range m.bounties {
		if bounty.Active {
			stats["active_bounties"]++
		}
	}

	return stats
}

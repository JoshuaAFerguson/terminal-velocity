// File: internal/arena/manager.go
// Project: Terminal Velocity
// Description: Enhanced PvP system with arenas, tournaments, and spectator mode
// Version: 1.0.0
// Author: Claude Code
// Created: 2025-11-15

package arena

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JoshuaAFerguson/terminal-velocity/internal/database"
	"github.com/JoshuaAFerguson/terminal-velocity/internal/logger"
	"github.com/google/uuid"
)

var log = logger.WithComponent("Arena")

// Manager handles arena battles, tournaments, and spectating
type Manager struct {
	mu sync.RWMutex

	// Arena data
	arenas      map[uuid.UUID]*Arena
	matches     map[uuid.UUID]*Match
	tournaments map[uuid.UUID]*Tournament
	spectators  map[uuid.UUID][]uuid.UUID // match_id -> spectator player IDs
	rankings    map[string]*PlayerRanking  // player_id -> ranking
	matchQueue  map[MatchType][]*QueueEntry // matchmaking queue by match type

	// Configuration
	config ArenaConfig

	// Repositories
	playerRepo *database.PlayerRepository

	// Callbacks
	onMatchStart    func(match *Match)
	onMatchEnd      func(match *Match)
	onTournamentStart func(tournament *Tournament)
	onTournamentEnd   func(tournament *Tournament)
	onSpectatorJoin   func(matchID, spectatorID uuid.UUID)

	// Background workers
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// ArenaConfig defines arena system parameters
type ArenaConfig struct {
	// Arena settings
	ArenaCount              int           // Number of available arenas
	MatchQueueTimeout       time.Duration // Max time in matchmaking queue
	RankingDecayRate        float64       // Daily ranking decay

	// Match settings
	MatchDuration           time.Duration // Maximum match duration
	RoundDuration           time.Duration // Duration per round
	RespawnTime             time.Duration // Time before respawn in match
	SpectatorDelay          time.Duration // Spectator view delay
	MaxSpectatorsPerMatch   int           // Max spectators per match

	// Tournament settings
	MinTournamentPlayers    int           // Minimum players to start
	MaxTournamentPlayers    int           // Maximum tournament size
	TournamentEntryFee      int64         // Entry fee in credits
	TournamentPrizePool     float64       // % of entry fees as prizes
	TournamentBracketType   string        // "single_elimination", "double_elimination"

	// Ranking settings
	StartingELO             int           // Starting ELO rating
	ELOKFactor              int           // ELO calculation K-factor
	WinStreakBonus          int           // Bonus per win streak
	RankTiers               []string      // Rank tier names
}

// DefaultArenaConfig returns sensible defaults
func DefaultArenaConfig() ArenaConfig {
	return ArenaConfig{
		ArenaCount:            10,
		MatchQueueTimeout:     5 * time.Minute,
		RankingDecayRate:      0.02, // 2% per day
		MatchDuration:         15 * time.Minute,
		RoundDuration:         5 * time.Minute,
		RespawnTime:           10 * time.Second,
		SpectatorDelay:        5 * time.Second,
		MaxSpectatorsPerMatch: 50,
		MinTournamentPlayers:  4,
		MaxTournamentPlayers:  32,
		TournamentEntryFee:    10000,
		TournamentPrizePool:   0.90, // 90% of fees
		TournamentBracketType: "single_elimination",
		StartingELO:           1000,
		ELOKFactor:            32,
		WinStreakBonus:        50,
		RankTiers: []string{
			"Bronze", "Silver", "Gold", "Platinum", "Diamond", "Master", "Grandmaster",
		},
	}
}

// NewManager creates a new arena manager
func NewManager(playerRepo *database.PlayerRepository) *Manager {
	m := &Manager{
		arenas:      make(map[uuid.UUID]*Arena),
		matches:     make(map[uuid.UUID]*Match),
		tournaments: make(map[uuid.UUID]*Tournament),
		spectators:  make(map[uuid.UUID][]uuid.UUID),
		rankings:    make(map[string]*PlayerRanking),
		matchQueue:  make(map[MatchType][]*QueueEntry),
		config:      DefaultArenaConfig(),
		playerRepo:  playerRepo,
		stopChan:    make(chan struct{}),
	}

	// Initialize arenas
	m.initializeArenas()

	return m
}

// Start begins background workers
func (m *Manager) Start() {
	m.wg.Add(2)
	go m.maintenanceWorker()
	go m.matchmakingWorker()
	log.Info("Arena manager started")
}

// Stop gracefully shuts down the manager
func (m *Manager) Stop() {
	close(m.stopChan)
	m.wg.Wait()
	log.Info("Arena manager stopped")
}

// SetCallbacks sets all arena callbacks
func (m *Manager) SetCallbacks(
	onMatchStart func(match *Match),
	onMatchEnd func(match *Match),
	onTournamentStart func(tournament *Tournament),
	onTournamentEnd func(tournament *Tournament),
	onSpectatorJoin func(matchID, spectatorID uuid.UUID),
) {
	m.onMatchStart = onMatchStart
	m.onMatchEnd = onMatchEnd
	m.onTournamentStart = onTournamentStart
	m.onTournamentEnd = onTournamentEnd
	m.onSpectatorJoin = onSpectatorJoin
}

// ============================================================================
// DATA STRUCTURES
// ============================================================================

// Arena represents a PvP battleground
type Arena struct {
	ID          uuid.UUID
	Name        string
	Description string
	MapType     string // "asteroid_field", "nebula", "space_station", "debris_field"
	Size        string // "small", "medium", "large"
	Features    []string // "cover", "hazards", "power_ups", "objectives"
	Capacity    int    // Max players in match
	Status      string // "available", "occupied", "maintenance"
}

// Match represents an active PvP match
type Match struct {
	ID            uuid.UUID
	ArenaID       uuid.UUID
	Type          MatchType
	Players       []uuid.UUID
	Teams         map[string][]uuid.UUID // "red" -> player IDs, "blue" -> player IDs
	Scores        map[uuid.UUID]int      // player_id -> score
	StartTime     time.Time
	EndTime       time.Time
	Status        string // "waiting", "in_progress", "completed", "cancelled"
	Winner        uuid.UUID
	MatchData     *MatchData
	TournamentID  *uuid.UUID // If part of tournament
}

// MatchType defines the type of PvP match
type MatchType string

const (
	MatchTypeDuel          MatchType = "duel"           // 1v1
	MatchTypeTeamDeathmatch MatchType = "team_deathmatch" // Team vs Team
	MatchTypeFreeForAll    MatchType = "free_for_all"   // Everyone vs Everyone
	MatchTypeCaptureFlag   MatchType = "capture_flag"   // CTF mode
	MatchTypeKingOfHill    MatchType = "king_of_hill"   // Control point
	MatchTypeElimination   MatchType = "elimination"    // Last standing wins
)

// MatchData contains detailed match information
type MatchData struct {
	Kills         map[uuid.UUID]int
	Deaths        map[uuid.UUID]int
	Assists       map[uuid.UUID]int
	DamageDealt   map[uuid.UUID]float64
	DamageTaken   map[uuid.UUID]float64
	Objectives    map[uuid.UUID]int // Flags captured, points controlled, etc.
}

// Tournament represents a competitive tournament
type Tournament struct {
	ID            uuid.UUID
	Name          string
	Type          TournamentType
	EntryFee      int64
	PrizePool     int64
	MaxPlayers    int
	Participants  []uuid.UUID
	Bracket       *TournamentBracket
	CurrentRound  int
	StartTime     time.Time
	EndTime       time.Time
	Status        string // "registration", "in_progress", "completed"
	Winners       []uuid.UUID // 1st, 2nd, 3rd place
}

// TournamentType defines tournament format
type TournamentType string

const (
	TournamentSingleElimination TournamentType = "single_elimination"
	TournamentDoubleElimination TournamentType = "double_elimination"
	TournamentRoundRobin        TournamentType = "round_robin"
	TournamentSwiss             TournamentType = "swiss"
)

// TournamentBracket represents the tournament structure
type TournamentBracket struct {
	Rounds   []*TournamentRound
	Matches  map[uuid.UUID]*Match
}

// TournamentRound represents a round in the bracket
type TournamentRound struct {
	RoundNumber int
	Matches     []uuid.UUID // Match IDs
	Status      string      // "pending", "in_progress", "completed"
}

// PlayerRanking tracks a player's competitive ranking
type PlayerRanking struct {
	PlayerID      uuid.UUID
	ELO           int
	Tier          string
	Division      int       // 1-5 within tier
	Wins          int
	Losses        int
	WinStreak     int
	HighestELO    int
	TournamentsWon int
	LastMatchTime time.Time
}

// QueueEntry represents a player in the matchmaking queue
type QueueEntry struct {
	PlayerID  uuid.UUID
	MatchType MatchType
	ELO       int
	QueueTime time.Time
}

// ============================================================================
// MATCHMAKING
// ============================================================================

// QueueForMatch adds a player to matchmaking queue
func (m *Manager) QueueForMatch(ctx context.Context, playerID uuid.UUID, matchType MatchType) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if player already in a match
	for _, match := range m.matches {
		if match.Status == "in_progress" || match.Status == "waiting" {
			for _, pid := range match.Players {
				if pid == playerID {
					return fmt.Errorf("already in a match")
				}
			}
		}
	}

	// Check if player already in queue
	if queue, exists := m.matchQueue[matchType]; exists {
		for _, entry := range queue {
			if entry.PlayerID == playerID {
				return fmt.Errorf("already in queue for %s", matchType)
			}
		}
	}

	// Get player's ranking (or create default)
	ranking := m.getOrCreateRanking(playerID)

	// Add to queue
	entry := &QueueEntry{
		PlayerID:  playerID,
		MatchType: matchType,
		ELO:       ranking.ELO,
		QueueTime: time.Now(),
	}

	m.matchQueue[matchType] = append(m.matchQueue[matchType], entry)

	log.Info("Player queued for match: player=%s, type=%s, ELO=%d, queue_size=%d",
		playerID, matchType, ranking.ELO, len(m.matchQueue[matchType]))

	return nil
}

// CreateMatch creates a new PvP match
func (m *Manager) CreateMatch(ctx context.Context, matchType MatchType, players []uuid.UUID) (*Match, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find available arena
	var arena *Arena
	for _, a := range m.arenas {
		if a.Status == "available" {
			arena = a
			break
		}
	}

	if arena == nil {
		return nil, fmt.Errorf("no available arenas")
	}

	// Create match
	match := &Match{
		ID:        uuid.New(),
		ArenaID:   arena.ID,
		Type:      matchType,
		Players:   players,
		Teams:     make(map[string][]uuid.UUID),
		Scores:    make(map[uuid.UUID]int),
		StartTime: time.Now(),
		Status:    "waiting",
		MatchData: &MatchData{
			Kills:       make(map[uuid.UUID]int),
			Deaths:      make(map[uuid.UUID]int),
			Assists:     make(map[uuid.UUID]int),
			DamageDealt: make(map[uuid.UUID]float64),
			DamageTaken: make(map[uuid.UUID]float64),
			Objectives:  make(map[uuid.UUID]int),
		},
	}

	// Assign teams for team-based modes
	if matchType == MatchTypeTeamDeathmatch || matchType == MatchTypeCaptureFlag {
		mid := len(players) / 2
		match.Teams["red"] = players[:mid]
		match.Teams["blue"] = players[mid:]
	}

	m.matches[match.ID] = match
	arena.Status = "occupied"

	log.Info("Match created: type=%s, players=%d, arena=%s", matchType, len(players), arena.Name)

	if m.onMatchStart != nil {
		go m.onMatchStart(match)
	}

	return match, nil
}

// StartMatch begins a match
func (m *Manager) StartMatch(ctx context.Context, matchID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	match, exists := m.matches[matchID]
	if !exists {
		return fmt.Errorf("match not found")
	}

	if match.Status != "waiting" {
		return fmt.Errorf("match already started or completed")
	}

	match.Status = "in_progress"
	match.StartTime = time.Now()

	log.Info("Match started: match=%s, type=%s", matchID, match.Type)

	// Start match timer
	go m.matchTimerWorker(matchID)

	return nil
}

// EndMatch ends a match and calculates results
func (m *Manager) EndMatch(ctx context.Context, matchID uuid.UUID, winnerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	match, exists := m.matches[matchID]
	if !exists {
		return fmt.Errorf("match not found")
	}

	if match.Status != "in_progress" {
		return fmt.Errorf("match not in progress")
	}

	match.Status = "completed"
	match.EndTime = time.Now()
	match.Winner = winnerID

	// Update rankings
	m.updateRankingsAfterMatch(match)

	// Free arena
	if arena, exists := m.arenas[match.ArenaID]; exists {
		arena.Status = "available"
	}

	log.Info("Match ended: match=%s, winner=%s, duration=%v",
		matchID, winnerID, match.EndTime.Sub(match.StartTime))

	if m.onMatchEnd != nil {
		go m.onMatchEnd(match)
	}

	return nil
}

// ============================================================================
// SPECTATOR SYSTEM
// ============================================================================

// JoinAsSpectator adds a player as spectator to a match
func (m *Manager) JoinAsSpectator(ctx context.Context, matchID, spectatorID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check match exists and is in progress
	match, exists := m.matches[matchID]
	if !exists {
		return fmt.Errorf("match not found")
	}

	if match.Status != "in_progress" {
		return fmt.Errorf("match not in progress")
	}

	// Check spectator limit
	spectators := m.spectators[matchID]
	if len(spectators) >= m.config.MaxSpectatorsPerMatch {
		return fmt.Errorf("spectator limit reached")
	}

	// Check if already spectating
	for _, sid := range spectators {
		if sid == spectatorID {
			return fmt.Errorf("already spectating this match")
		}
	}

	// Add spectator
	m.spectators[matchID] = append(spectators, spectatorID)

	log.Info("Spectator joined: match=%s, spectator=%s", matchID, spectatorID)

	if m.onSpectatorJoin != nil {
		go m.onSpectatorJoin(matchID, spectatorID)
	}

	return nil
}

// LeaveSpectator removes a spectator from a match
func (m *Manager) LeaveSpectator(ctx context.Context, matchID, spectatorID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	spectators := m.spectators[matchID]
	for i, sid := range spectators {
		if sid == spectatorID {
			m.spectators[matchID] = append(spectators[:i], spectators[i+1:]...)
			log.Info("Spectator left: match=%s, spectator=%s", matchID, spectatorID)
			return nil
		}
	}

	return fmt.Errorf("not spectating this match")
}

// GetSpectators retrieves all spectators for a match
func (m *Manager) GetSpectators(matchID uuid.UUID) []uuid.UUID {
	m.mu.RLock()
	defer m.mu.RUnlock()

	spectators := m.spectators[matchID]
	// Return copy
	result := make([]uuid.UUID, len(spectators))
	copy(result, spectators)
	return result
}

// GetActiveMatches retrieves all active matches for spectating
func (m *Manager) GetActiveMatches() []*Match {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matches []*Match
	for _, match := range m.matches {
		if match.Status == "in_progress" {
			matches = append(matches, match)
		}
	}
	return matches
}

// ============================================================================
// TOURNAMENT SYSTEM
// ============================================================================

// CreateTournament creates a new tournament
func (m *Manager) CreateTournament(ctx context.Context, name string, tournamentType TournamentType, maxPlayers int) (*Tournament, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	tournament := &Tournament{
		ID:           uuid.New(),
		Name:         name,
		Type:         tournamentType,
		EntryFee:     m.config.TournamentEntryFee,
		PrizePool:    0,
		MaxPlayers:   maxPlayers,
		Participants: []uuid.UUID{},
		Status:       "registration",
	}

	m.tournaments[tournament.ID] = tournament

	log.Info("Tournament created: name=%s, type=%s, max_players=%d", name, tournamentType, maxPlayers)
	return tournament, nil
}

// RegisterForTournament adds a player to a tournament
func (m *Manager) RegisterForTournament(ctx context.Context, tournamentID, playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tournament, exists := m.tournaments[tournamentID]
	if !exists {
		return fmt.Errorf("tournament not found")
	}

	if tournament.Status != "registration" {
		return fmt.Errorf("registration closed")
	}

	if len(tournament.Participants) >= tournament.MaxPlayers {
		return fmt.Errorf("tournament full")
	}

	// Check if already registered
	for _, pid := range tournament.Participants {
		if pid == playerID {
			return fmt.Errorf("already registered")
		}
	}

	// Check entry fee
	player, err := m.playerRepo.GetByID(ctx, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %v", err)
	}
	if player.Credits < tournament.EntryFee {
		return fmt.Errorf("insufficient credits (need %d)", tournament.EntryFee)
	}

	// Deduct entry fee
	player.Credits -= tournament.EntryFee
	if err := m.playerRepo.Update(ctx, player); err != nil {
		return fmt.Errorf("failed to deduct credits: %v", err)
	}

	// Add to tournament
	tournament.Participants = append(tournament.Participants, playerID)
	tournament.PrizePool += int64(float64(tournament.EntryFee) * m.config.TournamentPrizePool)

	log.Info("Player registered for tournament: tournament=%s, player=%s", tournament.Name, playerID)

	// Auto-start if minimum reached
	if len(tournament.Participants) >= m.config.MinTournamentPlayers {
		// Could auto-start or wait for more players
	}

	return nil
}

// StartTournament begins a tournament
func (m *Manager) StartTournament(ctx context.Context, tournamentID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tournament, exists := m.tournaments[tournamentID]
	if !exists {
		return fmt.Errorf("tournament not found")
	}

	if tournament.Status != "registration" {
		return fmt.Errorf("tournament already started or completed")
	}

	if len(tournament.Participants) < m.config.MinTournamentPlayers {
		return fmt.Errorf("not enough participants (need %d)", m.config.MinTournamentPlayers)
	}

	// Generate bracket
	tournament.Bracket = m.generateBracket(tournament)
	tournament.Status = "in_progress"
	tournament.StartTime = time.Now()
	tournament.CurrentRound = 1

	log.Info("Tournament started: name=%s, participants=%d", tournament.Name, len(tournament.Participants))

	if m.onTournamentStart != nil {
		go m.onTournamentStart(tournament)
	}

	return nil
}

// generateBracket creates a tournament bracket
func (m *Manager) generateBracket(tournament *Tournament) *TournamentBracket {
	bracket := &TournamentBracket{
		Rounds:  []*TournamentRound{},
		Matches: make(map[uuid.UUID]*Match),
	}

	// For single elimination, calculate number of rounds
	numRounds := 0
	players := len(tournament.Participants)
	for players > 1 {
		players /= 2
		numRounds++
	}

	// Initialize rounds
	for i := 0; i < numRounds; i++ {
		bracket.Rounds = append(bracket.Rounds, &TournamentRound{
			RoundNumber: i + 1,
			Matches:     []uuid.UUID{},
			Status:      "pending",
		})
	}

	// Create first round matches
	// TODO: Implement bracket generation logic
	// For now, just create placeholder structure

	return bracket
}

// ============================================================================
// RANKING SYSTEM
// ============================================================================

// updateRankingsAfterMatch updates player rankings based on match results
func (m *Manager) updateRankingsAfterMatch(match *Match) {
	// Simple ELO update for now
	for _, playerID := range match.Players {
		ranking := m.getOrCreateRanking(playerID)

		if playerID == match.Winner {
			ranking.Wins++
			ranking.WinStreak++
			ranking.ELO += m.config.ELOKFactor + (ranking.WinStreak * m.config.WinStreakBonus / 10)
		} else {
			ranking.Losses++
			ranking.WinStreak = 0
			ranking.ELO -= m.config.ELOKFactor
		}

		// Update highest ELO
		if ranking.ELO > ranking.HighestELO {
			ranking.HighestELO = ranking.ELO
		}

		// Update tier
		ranking.Tier = m.calculateTier(ranking.ELO)
		ranking.LastMatchTime = time.Now()
	}
}

// getOrCreateRanking gets or creates a player ranking
func (m *Manager) getOrCreateRanking(playerID uuid.UUID) *PlayerRanking {
	key := playerID.String()
	if ranking, exists := m.rankings[key]; exists {
		return ranking
	}

	ranking := &PlayerRanking{
		PlayerID:   playerID,
		ELO:        m.config.StartingELO,
		Tier:       m.config.RankTiers[0],
		Division:   5,
		HighestELO: m.config.StartingELO,
	}
	m.rankings[key] = ranking
	return ranking
}

// calculateTier determines rank tier based on ELO
func (m *Manager) calculateTier(elo int) string {
	// Simple tier calculation
	if elo < 800 {
		return m.config.RankTiers[0] // Bronze
	} else if elo < 1100 {
		return m.config.RankTiers[1] // Silver
	} else if elo < 1400 {
		return m.config.RankTiers[2] // Gold
	} else if elo < 1700 {
		return m.config.RankTiers[3] // Platinum
	} else if elo < 2000 {
		return m.config.RankTiers[4] // Diamond
	} else if elo < 2300 {
		return m.config.RankTiers[5] // Master
	}
	return m.config.RankTiers[6] // Grandmaster
}

// GetPlayerRanking retrieves a player's ranking
func (m *Manager) GetPlayerRanking(playerID uuid.UUID) *PlayerRanking {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.getOrCreateRanking(playerID)
}

// GetLeaderboard retrieves top ranked players
func (m *Manager) GetLeaderboard(limit int) []*PlayerRanking {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect all rankings
	rankings := make([]*PlayerRanking, 0, len(m.rankings))
	for _, ranking := range m.rankings {
		rankings = append(rankings, ranking)
	}

	// Sort by ELO (simple bubble sort for now)
	for i := 0; i < len(rankings); i++ {
		for j := i + 1; j < len(rankings); j++ {
			if rankings[j].ELO > rankings[i].ELO {
				rankings[i], rankings[j] = rankings[j], rankings[i]
			}
		}
	}

	// Return top N
	if limit > 0 && limit < len(rankings) {
		return rankings[:limit]
	}
	return rankings
}

// ============================================================================
// BACKGROUND WORKERS
// ============================================================================

// matchmakingWorker processes the matchmaking queue
func (m *Manager) matchmakingWorker() {
	defer m.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.processMatchmakingQueue()
		case <-m.stopChan:
			return
		}
	}
}

// processMatchmakingQueue attempts to create matches from queued players
func (m *Manager) processMatchmakingQueue() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Process each match type queue
	for matchType, queue := range m.matchQueue {
		if len(queue) == 0 {
			continue
		}

		// Determine required player count
		requiredPlayers := m.getRequiredPlayers(matchType)

		// Remove timed out entries
		now := time.Now()
		validQueue := make([]*QueueEntry, 0, len(queue))
		for _, entry := range queue {
			if now.Sub(entry.QueueTime) < m.config.MatchQueueTimeout {
				validQueue = append(validQueue, entry)
			} else {
				log.Info("Player queue timeout: player=%s, type=%s", entry.PlayerID, matchType)
			}
		}
		queue = validQueue

		// Try to create matches
		for len(queue) >= requiredPlayers {
			// Find best match based on ELO (within ±200 ELO range)
			matched := m.findBestMatch(queue, requiredPlayers)
			if matched == nil {
				break
			}

			// Remove matched players from queue
			remainingQueue := make([]*QueueEntry, 0, len(queue))
			matchedIDs := make(map[uuid.UUID]bool)
			for _, entry := range matched {
				matchedIDs[entry.PlayerID] = true
			}
			for _, entry := range queue {
				if !matchedIDs[entry.PlayerID] {
					remainingQueue = append(remainingQueue, entry)
				}
			}
			queue = remainingQueue

			// Create match
			playerIDs := make([]uuid.UUID, len(matched))
			for i, entry := range matched {
				playerIDs[i] = entry.PlayerID
			}

			match, err := m.createMatchInternal(matchType, playerIDs)
			if err != nil {
				log.Error("Failed to create match from queue: %v", err)
				// Re-add players to queue
				queue = append(queue, matched...)
			} else {
				log.Info("Match created from queue: match=%s, type=%s, players=%d",
					match.ID, matchType, len(playerIDs))
			}
		}

		// Update queue
		m.matchQueue[matchType] = queue
	}
}

// getRequiredPlayers returns the number of players needed for a match type
func (m *Manager) getRequiredPlayers(matchType MatchType) int {
	switch matchType {
	case MatchTypeDuel:
		return 2
	case MatchTypeTeamDeathmatch:
		return 4 // 2v2
	case MatchTypeFreeForAll:
		return 4
	case MatchTypeCaptureFlag:
		return 6 // 3v3
	case MatchTypeKingOfHill:
		return 4
	case MatchTypeElimination:
		return 4
	default:
		return 2
	}
}

// findBestMatch finds the best group of players for a match based on ELO
func (m *Manager) findBestMatch(queue []*QueueEntry, count int) []*QueueEntry {
	if len(queue) < count {
		return nil
	}

	// Start with the player who's been waiting longest
	matched := []*QueueEntry{queue[0]}
	baseELO := queue[0].ELO

	// Find players within ±200 ELO
	for _, entry := range queue[1:] {
		if len(matched) >= count {
			break
		}

		eloDiff := entry.ELO - baseELO
		if eloDiff < 0 {
			eloDiff = -eloDiff
		}

		// Accept if within ELO range (±200)
		if eloDiff <= 200 {
			matched = append(matched, entry)
		}
	}

	// Only return if we have enough players
	if len(matched) < count {
		return nil
	}

	return matched[:count]
}

// createMatchInternal creates a match without locking (caller must hold lock)
func (m *Manager) createMatchInternal(matchType MatchType, players []uuid.UUID) (*Match, error) {
	// Find available arena
	var arena *Arena
	for _, a := range m.arenas {
		if a.Status == "available" {
			arena = a
			break
		}
	}

	if arena == nil {
		return nil, fmt.Errorf("no available arenas")
	}

	// Create match
	match := &Match{
		ID:        uuid.New(),
		ArenaID:   arena.ID,
		Type:      matchType,
		Players:   players,
		Teams:     make(map[string][]uuid.UUID),
		Scores:    make(map[uuid.UUID]int),
		StartTime: time.Now(),
		Status:    "waiting",
		MatchData: &MatchData{
			Kills:   make(map[uuid.UUID]int),
			Deaths:  make(map[uuid.UUID]int),
			Assists: make(map[uuid.UUID]int),
			Damage:  make(map[uuid.UUID]int),
		},
	}

	// Assign teams for team-based modes
	if matchType == MatchTypeTeamDeathmatch || matchType == MatchTypeCaptureFlag {
		half := len(players) / 2
		match.Teams["red"] = players[:half]
		match.Teams["blue"] = players[half:]
	}

	// Mark arena as occupied
	arena.Status = "occupied"

	// Store match
	m.matches[match.ID] = match

	// Trigger callback
	if m.onMatchStart != nil {
		go m.onMatchStart(match)
	}

	return match, nil
}

// LeaveQueue removes a player from the matchmaking queue
func (m *Manager) LeaveQueue(playerID uuid.UUID, matchType MatchType) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	queue, exists := m.matchQueue[matchType]
	if !exists {
		return fmt.Errorf("not in queue")
	}

	// Find and remove player
	newQueue := make([]*QueueEntry, 0, len(queue))
	found := false
	for _, entry := range queue {
		if entry.PlayerID != playerID {
			newQueue = append(newQueue, entry)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("not in queue")
	}

	m.matchQueue[matchType] = newQueue
	log.Info("Player left queue: player=%s, type=%s", playerID, matchType)

	return nil
}

// maintenanceWorker handles daily maintenance
func (m *Manager) maintenanceWorker() {
	defer m.wg.Done()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.processMaintenance()
		case <-m.stopChan:
			return
		}
	}
}

// processMaintenance processes daily arena maintenance
func (m *Manager) processMaintenance() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ranking decay for inactive players
	for _, ranking := range m.rankings {
		if time.Since(ranking.LastMatchTime) > 7*24*time.Hour {
			decay := int(float64(ranking.ELO) * m.config.RankingDecayRate)
			ranking.ELO -= decay
			if ranking.ELO < m.config.StartingELO/2 {
				ranking.ELO = m.config.StartingELO / 2
			}
		}
	}
}

// matchTimerWorker handles match timeout
func (m *Manager) matchTimerWorker(matchID uuid.UUID) {
	time.Sleep(m.config.MatchDuration)

	m.mu.Lock()
	match, exists := m.matches[matchID]
	if !exists || match.Status != "in_progress" {
		m.mu.Unlock()
		return
	}

	// Determine winner by score
	var winnerID uuid.UUID
	maxScore := -1
	for playerID, score := range match.Scores {
		if score > maxScore {
			maxScore = score
			winnerID = playerID
		}
	}

	match.Status = "completed"
	match.EndTime = time.Now()
	match.Winner = winnerID

	// Free arena
	if arena, exists := m.arenas[match.ArenaID]; exists {
		arena.Status = "available"
	}

	m.mu.Unlock()

	log.Info("Match timed out: match=%s, winner=%s", matchID, winnerID)
}

// ============================================================================
// INITIALIZATION
// ============================================================================

// initializeArenas creates default arenas
func (m *Manager) initializeArenas() {
	arenaTemplates := []struct {
		name        string
		description string
		mapType     string
		size        string
		features    []string
	}{
		{"Nebula Nexus", "Dense nebula with limited visibility", "nebula", "medium", []string{"cover", "hazards"}},
		{"Asteroid Belt Alpha", "Navigate through dense asteroids", "asteroid_field", "large", []string{"cover", "hazards"}},
		{"Derelict Station", "Abandoned space station ruins", "space_station", "medium", []string{"cover", "objectives"}},
		{"Debris Field Delta", "Battle wreckage from ancient war", "debris_field", "small", []string{"hazards", "power_ups"}},
		{"Crimson Arena", "Open space combat arena", "open_space", "medium", []string{"power_ups"}},
	}

	for _, template := range arenaTemplates {
		arena := &Arena{
			ID:          uuid.New(),
			Name:        template.name,
			Description: template.description,
			MapType:     template.mapType,
			Size:        template.size,
			Features:    template.features,
			Capacity:    8,
			Status:      "available",
		}
		m.arenas[arena.ID] = arena
	}

	log.Info("Initialized %d arenas", len(m.arenas))
}

// GetStats returns arena statistics
func (m *Manager) GetStats() ArenaStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := ArenaStats{
		AvailableArenas: 0,
		ActiveMatches:   0,
		TotalSpectators: 0,
		ActiveTournaments: 0,
	}

	for _, arena := range m.arenas {
		if arena.Status == "available" {
			stats.AvailableArenas++
		}
	}

	for _, match := range m.matches {
		if match.Status == "in_progress" {
			stats.ActiveMatches++
		}
	}

	for _, spectators := range m.spectators {
		stats.TotalSpectators += len(spectators)
	}

	for _, tournament := range m.tournaments {
		if tournament.Status == "in_progress" || tournament.Status == "registration" {
			stats.ActiveTournaments++
		}
	}

	return stats
}

// ArenaStats contains arena statistics
type ArenaStats struct {
	AvailableArenas   int `json:"available_arenas"`
	ActiveMatches     int `json:"active_matches"`
	TotalSpectators   int `json:"total_spectators"`
	ActiveTournaments int `json:"active_tournaments"`
}

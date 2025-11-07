package models

import (
	"time"

	"github.com/google/uuid"
)

// PlayerFaction represents a player-created organization
type PlayerFaction struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Tag       string    `json:"tag"` // 3-4 character identifier
	FounderID uuid.UUID `json:"founder_id"`
	CreatedAt time.Time `json:"created_at"`

	// Leadership
	LeaderID uuid.UUID   `json:"leader_id"`
	Officers []uuid.UUID `json:"officers"`

	// Members
	Members     []uuid.UUID `json:"members"`
	MemberLimit int         `json:"member_limit"`

	// Resources
	Treasury int64 `json:"treasury"`

	// Territory
	HomeSystem        *uuid.UUID  `json:"home_system,omitempty"`
	ControlledSystems []uuid.UUID `json:"controlled_systems"`

	// Progression
	Level      int   `json:"level"`
	Experience int64 `json:"experience"`

	// Properties
	Alignment    string `json:"alignment"` // trader, mercenary, explorer, pirate, corporate
	IsRecruiting bool   `json:"is_recruiting"`

	// Reputation with NPC governments
	Reputation map[string]int `json:"reputation"`

	// Settings
	TaxRate  float64         `json:"tax_rate"` // % of member income to treasury
	Settings FactionSettings `json:"settings"`
}

// FactionSettings contains faction configuration
type FactionSettings struct {
	PublicProfile     bool   `json:"public_profile"`
	AllowApplications bool   `json:"allow_applications"`
	RequireApproval   bool   `json:"require_approval"`
	MinCombatRating   int    `json:"min_combat_rating"`
	MOTD              string `json:"motd"` // Message of the day
}

// FactionMember represents a member's association with a faction
type FactionMember struct {
	FactionID    uuid.UUID `json:"faction_id"`
	PlayerID     uuid.UUID `json:"player_id"`
	Rank         string    `json:"rank"` // recruit, member, officer, leader
	JoinedAt     time.Time `json:"joined_at"`
	Contribution int64     `json:"contribution"` // Total credits contributed
}

// FactionRank enum
const (
	RankRecruit = "recruit"
	RankMember  = "member"
	RankOfficer = "officer"
	RankLeader  = "leader"
)

// FactionAlignment enum
const (
	AlignmentTrader    = "trader"
	AlignmentMercenary = "mercenary"
	AlignmentExplorer  = "explorer"
	AlignmentPirate    = "pirate"
	AlignmentCorporate = "corporate"
)

// FactionPerk represents unlockable faction benefits
type FactionPerk struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	RequiredLevel int    `json:"required_level"`
	Cost          int64  `json:"cost"`
	Active        bool   `json:"active"`
}

// NewPlayerFaction creates a new faction
func NewPlayerFaction(name, tag string, founderID uuid.UUID, alignment string) *PlayerFaction {
	now := time.Now()
	id := uuid.New()

	return &PlayerFaction{
		ID:                id,
		Name:              name,
		Tag:               tag,
		FounderID:         founderID,
		LeaderID:          founderID,
		CreatedAt:         now,
		Officers:          []uuid.UUID{},
		Members:           []uuid.UUID{founderID},
		MemberLimit:       10, // Starting limit
		Treasury:          0,
		ControlledSystems: []uuid.UUID{},
		Level:             1,
		Experience:        0,
		Alignment:         alignment,
		IsRecruiting:      false,
		Reputation:        make(map[string]int),
		TaxRate:           0.05, // 5% default
		Settings: FactionSettings{
			PublicProfile:     true,
			AllowApplications: true,
			RequireApproval:   true,
			MinCombatRating:   0,
			MOTD:              "Welcome to " + name + "!",
		},
	}
}

// IsMember checks if a player is a member
func (f *PlayerFaction) IsMember(playerID uuid.UUID) bool {
	for _, id := range f.Members {
		if id == playerID {
			return true
		}
	}
	return false
}

// IsOfficer checks if a player is an officer
func (f *PlayerFaction) IsOfficer(playerID uuid.UUID) bool {
	if f.LeaderID == playerID {
		return true
	}
	for _, id := range f.Officers {
		if id == playerID {
			return true
		}
	}
	return false
}

// IsLeader checks if a player is the leader
func (f *PlayerFaction) IsLeader(playerID uuid.UUID) bool {
	return f.LeaderID == playerID
}

// CanRecruit checks if faction can accept new members
func (f *PlayerFaction) CanRecruit() bool {
	return len(f.Members) < f.MemberLimit
}

// AddMember adds a player to the faction
func (f *PlayerFaction) AddMember(playerID uuid.UUID) bool {
	if !f.CanRecruit() || f.IsMember(playerID) {
		return false
	}
	f.Members = append(f.Members, playerID)
	return true
}

// RemoveMember removes a player from the faction
func (f *PlayerFaction) RemoveMember(playerID uuid.UUID) bool {
	// Can't remove leader
	if f.IsLeader(playerID) {
		return false
	}

	// Remove from officers if applicable
	for i, id := range f.Officers {
		if id == playerID {
			f.Officers = append(f.Officers[:i], f.Officers[i+1:]...)
			break
		}
	}

	// Remove from members
	for i, id := range f.Members {
		if id == playerID {
			f.Members = append(f.Members[:i], f.Members[i+1:]...)
			return true
		}
	}

	return false
}

// PromoteToOfficer promotes a member to officer
func (f *PlayerFaction) PromoteToOfficer(playerID uuid.UUID) bool {
	if !f.IsMember(playerID) || f.IsOfficer(playerID) {
		return false
	}
	f.Officers = append(f.Officers, playerID)
	return true
}

// DemoteFromOfficer demotes an officer to member
func (f *PlayerFaction) DemoteFromOfficer(playerID uuid.UUID) bool {
	if f.IsLeader(playerID) {
		return false // Can't demote leader
	}

	for i, id := range f.Officers {
		if id == playerID {
			f.Officers = append(f.Officers[:i], f.Officers[i+1:]...)
			return true
		}
	}
	return false
}

// Deposit adds credits to faction treasury
func (f *PlayerFaction) Deposit(amount int64) {
	if amount > 0 {
		f.Treasury += amount
	}
}

// Withdraw removes credits from treasury (returns false if insufficient)
func (f *PlayerFaction) Withdraw(amount int64) bool {
	if amount <= 0 || f.Treasury < amount {
		return false
	}
	f.Treasury -= amount
	return true
}

// AddExperience adds experience and handles leveling
func (f *PlayerFaction) AddExperience(xp int64) {
	f.Experience += xp

	// Level up thresholds (exponential)
	requiredXP := int64(f.Level * f.Level * 1000)
	if f.Experience >= requiredXP && f.Level < 10 {
		f.Level++
		f.MemberLimit += 5 // Increase member capacity on level up
	}
}

// GetFullName returns the faction name with tag
func (f *PlayerFaction) GetFullName() string {
	return "[" + f.Tag + "] " + f.Name
}

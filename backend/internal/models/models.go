package models

import (
	"time"
)

// Email// Pool represents a football pool
type Pool struct {
	ID             int       `json:"id" db:"pool_id"`
	Name           string    `json:"name" db:"pool_name"`
	Description    string    `json:"description" db:"description"`
	MaxPlayers     int       `json:"max_players" db:"max_members"`
	Season         int       `json:"season" db:"season_year"`
	PoolType       string    `json:"pool_type" db:"pool_type"`
	EntryFee       float64   `json:"entry_fee" db:"entry_fee"`
	IsActive       string    `json:"is_active" db:"status"`
	CreatedBy      int       `json:"created_by" db:"commissioner_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	
	// Additional fields for API responses
	CreatorName    string       `json:"creator_name,omitempty"`
	CurrentMembers int          `json:"current_members,omitempty"`
	UserRole       string       `json:"user_role,omitempty"`
	Members        []PoolMember `json:"members,omitempty"`
}

// EmailAccount represents an email-based account
type EmailAccount struct {
	EmailID      int       `json:"email_id" db:"email_id"`
	EmailAddress string    `json:"email_address" db:"email_address"`
	PasswordHash string    `json:"-" db:"password_hash"` // Never include in JSON responses
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// UserProfile represents a user profile associated with an email account
type UserProfile struct {
	UserID      int       `json:"user_id" db:"user_id"`
	EmailID     int       `json:"email_id" db:"email_id"`
	Username    string    `json:"username" db:"username"`
	DisplayName string    `json:"display_name" db:"display_name"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Role represents user roles within pools
type Role struct {
	RoleID      int    `json:"role_id" db:"role_id"`
	RoleName    string `json:"role_name" db:"role_name"`
	Description string `json:"description" db:"description"`
}

// NFLTeam represents an NFL team
type NFLTeam struct {
	TeamID           int        `json:"team_id" db:"team_id"`
	TeamName         string     `json:"team_name" db:"team_name"`
	TeamAbbreviation string     `json:"team_abbreviation" db:"team_abbreviation"`
	City             string     `json:"city" db:"city"`
	Conference       string     `json:"conference" db:"conference"`
	Division         string     `json:"division" db:"division"`
	LogoURL          *string    `json:"logo_url" db:"logo_url"`
	PrimaryColor     *string    `json:"primary_color" db:"primary_color"`
	SecondaryColor   *string    `json:"secondary_color" db:"secondary_color"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

// PoolMembership represents a user's membership in a pool
type PoolMembership struct {
	MembershipID int       `json:"membership_id" db:"membership_id"`
	PoolID       int       `json:"pool_id" db:"pool_id"`
	UserID       int       `json:"user_id" db:"user_id"`
	RoleID       int       `json:"role_id" db:"role_id"`
	JoinedAt     time.Time `json:"joined_at" db:"joined_at"`
}

// NFLGame represents an NFL game
type NFLGame struct {
	GameID        int       `json:"game_id" db:"game_id"`
	ExternalID    string    `json:"external_id" db:"external_id"`
	SeasonYear    int       `json:"season_year" db:"season_year"`
	Week          int       `json:"week" db:"week"`
	GameType      string    `json:"game_type" db:"game_type"`
	HomeTeamID    int       `json:"home_team_id" db:"home_team_id"`
	AwayTeamID    int       `json:"away_team_id" db:"away_team_id"`
	GameDate      time.Time `json:"game_date" db:"game_date"`
	HomeScore     int       `json:"home_score" db:"home_score"`
	AwayScore     int       `json:"away_score" db:"away_score"`
	Status        string    `json:"status" db:"status"`
	Quarter       int       `json:"quarter" db:"quarter"`
	TimeRemaining string    `json:"time_remaining" db:"time_remaining"`
	LastUpdated   time.Time `json:"last_updated" db:"last_updated"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// SeasonPick represents a user's team selection for the season
type SeasonPick struct {
	PickID          int       `json:"pick_id" db:"pick_id"`
	PoolID          int       `json:"pool_id" db:"pool_id"`
	UserID          int       `json:"user_id" db:"user_id"`
	TeamID          int       `json:"team_id" db:"team_id"`
	PickOrder       int       `json:"pick_order" db:"pick_order"`
	PointsScored    int       `json:"points_scored" db:"points_scored"`
	IsEliminated    bool      `json:"is_eliminated" db:"is_eliminated"`
	EliminationWeek int       `json:"elimination_week" db:"elimination_week"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// ChatMessage represents a chat message in a pool
type ChatMessage struct {
	ID          string    `json:"id,omitempty" db:"id"`
	PoolID      string    `json:"pool_id" db:"pool_id"`
	UserID      string    `json:"user_id" db:"user_id"`
	DisplayName string    `json:"display_name" db:"display_name"`
	Message     string    `json:"message" db:"message"`
	MessageType string    `json:"message_type" db:"message_type"`
	Timestamp   time.Time `json:"timestamp" db:"created_at"`
}

// API Request/Response Models

// LoginRequest represents login request data
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterRequest represents registration request data
type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	Username    string `json:"username" binding:"required,min=3,max=50"`
	DisplayName string `json:"display_name" binding:"required,min=1,max=100"`
}

// LoginResponse represents login response data
type LoginResponse struct {
	Token    string        `json:"token"`
	User     UserProfile   `json:"user"`
	Profiles []UserProfile `json:"profiles"`
}

// CreatePoolRequest represents pool creation request data
type CreatePoolRequest struct {
	PoolName       string                 `json:"pool_name" binding:"required,min=1,max=100"`
	SeasonYear     int                    `json:"season_year" binding:"required,min=2020,max=2030"`
	MaxMembers     int                    `json:"max_members" binding:"min=2,max=100"`
	EntryFee       float64               `json:"entry_fee" binding:"min=0"`
	PrizeStructure map[string]interface{} `json:"prize_structure"`
	Settings       map[string]interface{} `json:"settings"`
}

// CreatePickRequest represents pick creation request data
type CreatePickRequest struct {
	PoolID    int `json:"pool_id" binding:"required"`
	TeamID    int `json:"team_id" binding:"required"`
	PickOrder int `json:"pick_order" binding:"required,min=1,max=4"`
}

// ErrorResponse represents API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// SuccessResponse represents API success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// StandingsEntry represents a single entry in pool standings
type StandingsEntry struct {
	UserID      int    `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	TotalPoints int    `json:"total_points"`
	Rank        int    `json:"rank"`
	TeamsPicked int    `json:"teams_picked"`
	Eliminated  bool   `json:"eliminated"`
}

// ChatMessageWithUser represents a chat message with user information
type ChatMessageWithUser struct {
	ChatMessage
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

// GameWithTeams represents a game with team information
type GameWithTeams struct {
	NFLGame
	HomeTeam NFLTeam `json:"home_team"`
	AwayTeam NFLTeam `json:"away_team"`
}

// PickWithTeam represents a pick with team information
type PickWithTeam struct {
	SeasonPick
	Team NFLTeam `json:"team"`
}

// PoolWithDetails represents a pool with additional details
type PoolWithDetails struct {
	Pool
	CreatorName   string `json:"creator_name"`
	MemberCount   int    `json:"member_count"`
	UserRole      string `json:"user_role,omitempty"`
	UserMemberID  int    `json:"user_member_id,omitempty"`
}

// PoolMember represents a member of a pool
type PoolMember struct {
	UserID      int       `json:"user_id"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
}

// Standing represents a user's standing in a pool
type Standing struct {
	Position        int      `json:"position"`
	UserID          string   `json:"user_id"`
	DisplayName     string   `json:"display_name"`
	TotalPicks      int      `json:"total_picks"`
	CorrectPicks    int      `json:"correct_picks"`
	IncorrectPicks  int      `json:"incorrect_picks"`
	WinPercentage   float64  `json:"win_percentage"`
	Week            *int     `json:"week,omitempty"`
}

// UserStats represents detailed statistics for a user
type UserStats struct {
	UserID         string         `json:"user_id"`
	DisplayName    string         `json:"display_name"`
	PoolID         string         `json:"pool_id"`
	Season         int            `json:"season"`
	PoolType       string         `json:"pool_type"`
	TotalPicks     int            `json:"total_picks"`
	CorrectPicks   int            `json:"correct_picks"`
	IncorrectPicks int            `json:"incorrect_picks"`
	WinPercentage  float64        `json:"win_percentage"`
	WeeklyStats    []WeeklyStats  `json:"weekly_stats,omitempty"`
	RecentPicks    []RecentPick   `json:"recent_picks,omitempty"`
}

// WeeklyStats represents statistics for a specific week
type WeeklyStats struct {
	Week           int     `json:"week"`
	TotalPicks     int     `json:"total_picks"`
	CorrectPicks   int     `json:"correct_picks"`
	IncorrectPicks int     `json:"incorrect_picks"`
	WinPercentage  float64 `json:"win_percentage"`
}

// RecentPick represents a recent pick with game details
type RecentPick struct {
	GameID         string     `json:"game_id"`
	Week           int        `json:"week"`
	GameDate       time.Time  `json:"game_date"`
	Status         string     `json:"status"`
	HomeTeam       string     `json:"home_team"`
	HomeTeamName   string     `json:"home_team_name"`
	AwayTeam       string     `json:"away_team"`
	AwayTeamName   string     `json:"away_team_name"`
	HomeScore      *int       `json:"home_score"`
	AwayScore      *int       `json:"away_score"`
	PickedTeam     string     `json:"picked_team"`
	PickedTeamName string     `json:"picked_team_name"`
	Confidence     *int       `json:"confidence"`
	IsCorrect      *bool      `json:"is_correct"`
}

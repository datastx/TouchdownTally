package handlers

import (
	"database/sql"
	"strconv"

	"touchdown-tally/internal/config"
	"touchdown-tally/internal/models"
	"touchdown-tally/pkg/logger"
	"touchdown-tally/pkg/response"

	"github.com/gin-gonic/gin"
)

// GameHandler handles game-related requests
type GameHandler struct {
	db     *sql.DB
	config *config.Config
	logger *logger.Logger
}

// NewGameHandler creates a new GameHandler
func NewGameHandler(db *sql.DB, cfg *config.Config, logger *logger.Logger) *GameHandler {
	return &GameHandler{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

// List returns games with optional filters
func (h *GameHandler) List(c *gin.Context) {
	seasonYear := c.Query("season_year")
	week := c.Query("week")
	status := c.Query("status")

	query := `
		SELECT g.game_id, g.external_id, g.season_year, g.week, g.game_type,
		       g.home_team_id, g.away_team_id, g.game_date, g.home_score, g.away_score,
		       g.status, g.quarter, g.time_remaining, g.last_updated, g.created_at,
		       ht.team_name as home_team_name, ht.team_abbreviation as home_team_abbr,
		       ht.city as home_team_city, ht.conference as home_team_conference,
		       ht.division as home_team_division, ht.logo_url as home_team_logo,
		       ht.primary_color as home_team_primary, ht.secondary_color as home_team_secondary,
		       at.team_name as away_team_name, at.team_abbreviation as away_team_abbr,
		       at.city as away_team_city, at.conference as away_team_conference,
		       at.division as away_team_division, at.logo_url as away_team_logo,
		       at.primary_color as away_team_primary, at.secondary_color as away_team_secondary
		FROM nfl_games g
		JOIN nfl_teams ht ON g.home_team_id = ht.team_id
		JOIN nfl_teams at ON g.away_team_id = at.team_id
		WHERE 1=1`

	var args []interface{}
	argCount := 0

	if seasonYear != "" {
		argCount++
		query += " AND g.season_year = $" + strconv.Itoa(argCount)
		if year, err := strconv.Atoi(seasonYear); err == nil {
			args = append(args, year)
		}
	}

	if week != "" {
		argCount++
		query += " AND g.week = $" + strconv.Itoa(argCount)
		if weekNum, err := strconv.Atoi(week); err == nil {
			args = append(args, weekNum)
		}
	}

	if status != "" {
		argCount++
		query += " AND g.status = $" + strconv.Itoa(argCount)
		args = append(args, status)
	}

	query += " ORDER BY g.game_date"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		h.logger.Error("Failed to query games", "error", err)
		response.InternalServerError(c, "query_failed", "Failed to fetch games")
		return
	}
	defer rows.Close()

	var games []models.GameWithTeams
	for rows.Next() {
		var game models.GameWithTeams
		err := rows.Scan(
			&game.GameID, &game.ExternalID, &game.SeasonYear, &game.Week, &game.GameType,
			&game.HomeTeamID, &game.AwayTeamID, &game.GameDate, &game.HomeScore, &game.AwayScore,
			&game.Status, &game.Quarter, &game.TimeRemaining, &game.LastUpdated, &game.CreatedAt,
			&game.HomeTeam.TeamName, &game.HomeTeam.TeamAbbreviation, &game.HomeTeam.City,
			&game.HomeTeam.Conference, &game.HomeTeam.Division, &game.HomeTeam.LogoURL,
			&game.HomeTeam.PrimaryColor, &game.HomeTeam.SecondaryColor,
			&game.AwayTeam.TeamName, &game.AwayTeam.TeamAbbreviation, &game.AwayTeam.City,
			&game.AwayTeam.Conference, &game.AwayTeam.Division, &game.AwayTeam.LogoURL,
			&game.AwayTeam.PrimaryColor, &game.AwayTeam.SecondaryColor,
		)
		if err != nil {
			h.logger.Error("Failed to scan game", "error", err)
			continue
		}

		game.HomeTeam.TeamID = game.HomeTeamID
		game.AwayTeam.TeamID = game.AwayTeamID

		games = append(games, game)
	}

	response.Success(c, games)
}

// Get returns a specific game by ID
func (h *GameHandler) Get(c *gin.Context) {
	gameID := c.Param("id")

	var game models.GameWithTeams
	err := h.db.QueryRow(`
		SELECT g.game_id, g.external_id, g.season_year, g.week, g.game_type,
		       g.home_team_id, g.away_team_id, g.game_date, g.home_score, g.away_score,
		       g.status, g.quarter, g.time_remaining, g.last_updated, g.created_at,
		       ht.team_name as home_team_name, ht.team_abbreviation as home_team_abbr,
		       ht.city as home_team_city, ht.conference as home_team_conference,
		       ht.division as home_team_division, ht.logo_url as home_team_logo,
		       ht.primary_color as home_team_primary, ht.secondary_color as home_team_secondary,
		       at.team_name as away_team_name, at.team_abbreviation as away_team_abbr,
		       at.city as away_team_city, at.conference as away_team_conference,
		       at.division as away_team_division, at.logo_url as away_team_logo,
		       at.primary_color as away_team_primary, at.secondary_color as away_team_secondary
		FROM nfl_games g
		JOIN nfl_teams ht ON g.home_team_id = ht.team_id
		JOIN nfl_teams at ON g.away_team_id = at.team_id
		WHERE g.game_id = $1`,
		gameID,
	).Scan(
		&game.GameID, &game.ExternalID, &game.SeasonYear, &game.Week, &game.GameType,
		&game.HomeTeamID, &game.AwayTeamID, &game.GameDate, &game.HomeScore, &game.AwayScore,
		&game.Status, &game.Quarter, &game.TimeRemaining, &game.LastUpdated, &game.CreatedAt,
		&game.HomeTeam.TeamName, &game.HomeTeam.TeamAbbreviation, &game.HomeTeam.City,
		&game.HomeTeam.Conference, &game.HomeTeam.Division, &game.HomeTeam.LogoURL,
		&game.HomeTeam.PrimaryColor, &game.HomeTeam.SecondaryColor,
		&game.AwayTeam.TeamName, &game.AwayTeam.TeamAbbreviation, &game.AwayTeam.City,
		&game.AwayTeam.Conference, &game.AwayTeam.Division, &game.AwayTeam.LogoURL,
		&game.AwayTeam.PrimaryColor, &game.AwayTeam.SecondaryColor,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			response.NotFound(c, "game_not_found", "Game not found")
			return
		}
		h.logger.Error("Failed to query game", "error", err)
		response.InternalServerError(c, "query_failed", "Failed to fetch game")
		return
	}

	game.HomeTeam.TeamID = game.HomeTeamID
	game.AwayTeam.TeamID = game.AwayTeamID

	response.Success(c, game)
}

// GetByWeek returns games for a specific week
func (h *GameHandler) GetByWeek(c *gin.Context) {
	week := c.Param("week")
	seasonYear := c.Query("season_year")

	if seasonYear == "" {
		seasonYear = strconv.Itoa(h.config.NFLSeasonYear)
	}

	c.Set("week", week)
	c.Set("season_year", seasonYear)
	h.List(c)
}

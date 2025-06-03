package handlers

import (
	"database/sql"

	"touchdown-tally/internal/config"
	"touchdown-tally/internal/models"
	"touchdown-tally/pkg/logger"
	"touchdown-tally/pkg/response"

	"github.com/gin-gonic/gin"
)

// TeamHandler handles team-related requests
type TeamHandler struct {
	db     *sql.DB
	config *config.Config
	logger *logger.Logger
}

// NewTeamHandler creates a new TeamHandler
func NewTeamHandler(db *sql.DB, cfg *config.Config, logger *logger.Logger) *TeamHandler {
	return &TeamHandler{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

// List returns all NFL teams
func (h *TeamHandler) List(c *gin.Context) {
	conference := c.Query("conference") // Optional filter by conference
	division := c.Query("division")     // Optional filter by division

	query := `
		SELECT team_id, team_name, team_abbreviation, city, conference, division, 
		       logo_url, primary_color, secondary_color, created_at
		FROM nfl_teams
		WHERE 1=1`
	
	var args []interface{}
	argCount := 0

	if conference != "" {
		argCount++
		query += " AND conference = $" + string(rune('0'+argCount))
		args = append(args, conference)
	}

	if division != "" {
		argCount++
		query += " AND division = $" + string(rune('0'+argCount))
		args = append(args, division)
	}

	query += " ORDER BY conference, division, team_name"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		h.logger.Error("Failed to query teams", "error", err)
		response.InternalServerError(c, "query_failed", "Failed to fetch teams")
		return
	}
	defer rows.Close()

	var teams []models.NFLTeam
	for rows.Next() {
		var team models.NFLTeam
		err := rows.Scan(
			&team.TeamID, &team.TeamName, &team.TeamAbbreviation, &team.City,
			&team.Conference, &team.Division, &team.LogoURL, &team.PrimaryColor,
			&team.SecondaryColor, &team.CreatedAt,
		)
		if err != nil {
			h.logger.Error("Failed to scan team", "error", err)
			continue
		}
		teams = append(teams, team)
	}

	response.Success(c, teams)
}

// Get returns a specific team by ID
func (h *TeamHandler) Get(c *gin.Context) {
	teamID := c.Param("id")

	var team models.NFLTeam
	err := h.db.QueryRow(`
		SELECT team_id, team_name, team_abbreviation, city, conference, division,
		       logo_url, primary_color, secondary_color, created_at
		FROM nfl_teams
		WHERE team_id = $1`,
		teamID,
	).Scan(
		&team.TeamID, &team.TeamName, &team.TeamAbbreviation, &team.City,
		&team.Conference, &team.Division, &team.LogoURL, &team.PrimaryColor,
		&team.SecondaryColor, &team.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			response.NotFound(c, "team_not_found", "Team not found")
			return
		}
		h.logger.Error("Failed to query team", "error", err)
		response.InternalServerError(c, "query_failed", "Failed to fetch team")
		return
	}

	response.Success(c, team)
}

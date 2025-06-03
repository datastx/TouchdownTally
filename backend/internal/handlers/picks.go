package handlers

import (
	"database/sql"

	"touchdown-tally/internal/config"
	"touchdown-tally/internal/models"
	"touchdown-tally/pkg/logger"
	"touchdown-tally/pkg/response"

	"github.com/gin-gonic/gin"
)

// PickHandler handles pick-related requests
type PickHandler struct {
	db     *sql.DB
	config *config.Config
	logger *logger.Logger
}

// NewPickHandler creates a new PickHandler
func NewPickHandler(db *sql.DB, cfg *config.Config, logger *logger.Logger) *PickHandler {
	return &PickHandler{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

// GetByPool returns all picks for a specific pool
func (h *PickHandler) GetByPool(c *gin.Context) {
	poolID := c.Param("pool_id")
	userID, _ := c.Get("user_id")

	// Verify user is a member of the pool
	var membershipExists bool
	err := h.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM pool_memberships WHERE pool_id = $1 AND user_id = $2)",
		poolID, userID,
	).Scan(&membershipExists)

	if err != nil {
		h.logger.Error("Failed to check pool membership", "error", err)
		response.InternalServerError(c, "membership_check_failed", "Failed to verify pool membership")
		return
	}

	if !membershipExists {
		response.Forbidden(c, "not_pool_member", "You are not a member of this pool")
		return
	}

	rows, err := h.db.Query(`
		SELECT p.pick_id, p.pool_id, p.user_id, p.team_id, p.pick_order,
		       p.points_scored, p.is_eliminated, p.elimination_week, p.created_at, p.updated_at,
		       t.team_name, t.team_abbreviation, t.city, t.conference, t.division,
		       t.logo_url, t.primary_color, t.secondary_color,
		       u.username, u.display_name
		FROM season_picks p
		JOIN nfl_teams t ON p.team_id = t.team_id
		JOIN user_profiles u ON p.user_id = u.user_id
		WHERE p.pool_id = $1
		ORDER BY u.display_name, p.pick_order`,
		poolID,
	)

	if err != nil {
		h.logger.Error("Failed to query picks", "error", err)
		response.InternalServerError(c, "query_failed", "Failed to fetch picks")
		return
	}
	defer rows.Close()

	var picks []models.PickWithTeam
	for rows.Next() {
		var pick models.PickWithTeam
		var username, displayName string
		err := rows.Scan(
			&pick.PickID, &pick.PoolID, &pick.UserID, &pick.TeamID, &pick.PickOrder,
			&pick.PointsScored, &pick.IsEliminated, &pick.EliminationWeek, &pick.CreatedAt, &pick.UpdatedAt,
			&pick.Team.TeamName, &pick.Team.TeamAbbreviation, &pick.Team.City,
			&pick.Team.Conference, &pick.Team.Division, &pick.Team.LogoURL,
			&pick.Team.PrimaryColor, &pick.Team.SecondaryColor,
			&username, &displayName,
		)
		if err != nil {
			h.logger.Error("Failed to scan pick", "error", err)
			continue
		}

		pick.Team.TeamID = pick.TeamID
		picks = append(picks, pick)
	}

	response.Success(c, picks)
}

// Create creates a new pick
func (h *PickHandler) Create(c *gin.Context) {
	var req models.CreatePickRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid_request", err.Error())
		return
	}

	userID, _ := c.Get("user_id")

	// Verify user is a member of the pool
	var membershipExists bool
	err := h.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM pool_memberships WHERE pool_id = $1 AND user_id = $2)",
		req.PoolID, userID,
	).Scan(&membershipExists)

	if err != nil {
		h.logger.Error("Failed to check pool membership", "error", err)
		response.InternalServerError(c, "membership_check_failed", "Failed to verify pool membership")
		return
	}

	if !membershipExists {
		response.Forbidden(c, "not_pool_member", "You are not a member of this pool")
		return
	}

	// Check if team is already picked in this pool
	var teamTaken bool
	err = h.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM season_picks WHERE pool_id = $1 AND team_id = $2)",
		req.PoolID, req.TeamID,
	).Scan(&teamTaken)

	if err != nil {
		h.logger.Error("Failed to check team availability", "error", err)
		response.InternalServerError(c, "team_check_failed", "Failed to verify team availability")
		return
	}

	if teamTaken {
		response.Conflict(c, "team_already_picked", "This team has already been picked by another user")
		return
	}

	// Check if user already has a pick for this order
	var existingPickID int
	err = h.db.QueryRow(
		"SELECT pick_id FROM season_picks WHERE pool_id = $1 AND user_id = $2 AND pick_order = $3",
		req.PoolID, userID, req.PickOrder,
	).Scan(&existingPickID)

	if err != nil && err != sql.ErrNoRows {
		h.logger.Error("Failed to check existing pick", "error", err)
		response.InternalServerError(c, "pick_check_failed", "Failed to check existing picks")
		return
	}

	if existingPickID > 0 {
		response.Conflict(c, "pick_order_taken", "You already have a pick for this order")
		return
	}

	// Create the pick
	var pickID int
	err = h.db.QueryRow(`
		INSERT INTO season_picks (pool_id, user_id, team_id, pick_order)
		VALUES ($1, $2, $3, $4)
		RETURNING pick_id`,
		req.PoolID, userID, req.TeamID, req.PickOrder,
	).Scan(&pickID)

	if err != nil {
		h.logger.Error("Failed to create pick", "error", err)
		response.InternalServerError(c, "pick_creation_failed", "Failed to create pick")
		return
	}

	// Fetch the created pick with team details
	var pick models.PickWithTeam
	err = h.db.QueryRow(`
		SELECT p.pick_id, p.pool_id, p.user_id, p.team_id, p.pick_order,
		       p.points_scored, p.is_eliminated, p.elimination_week, p.created_at, p.updated_at,
		       t.team_name, t.team_abbreviation, t.city, t.conference, t.division,
		       t.logo_url, t.primary_color, t.secondary_color
		FROM season_picks p
		JOIN nfl_teams t ON p.team_id = t.team_id
		WHERE p.pick_id = $1`,
		pickID,
	).Scan(
		&pick.PickID, &pick.PoolID, &pick.UserID, &pick.TeamID, &pick.PickOrder,
		&pick.PointsScored, &pick.IsEliminated, &pick.EliminationWeek, &pick.CreatedAt, &pick.UpdatedAt,
		&pick.Team.TeamName, &pick.Team.TeamAbbreviation, &pick.Team.City,
		&pick.Team.Conference, &pick.Team.Division, &pick.Team.LogoURL,
		&pick.Team.PrimaryColor, &pick.Team.SecondaryColor,
	)

	if err != nil {
		h.logger.Error("Failed to fetch created pick", "error", err)
		response.InternalServerError(c, "pick_fetch_failed", "Pick created but failed to fetch details")
		return
	}

	pick.Team.TeamID = pick.TeamID

	h.logger.Info("Pick created successfully", "pick_id", pickID, "user_id", userID, "pool_id", req.PoolID)
	response.Created(c, pick, "Pick created successfully")
}

// Update updates an existing pick
func (h *PickHandler) Update(c *gin.Context) {
	pickID := c.Param("id")
	userID, _ := c.Get("user_id")

	var req models.CreatePickRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid_request", err.Error())
		return
	}

	// Verify user owns this pick
	var currentPoolID int
	err := h.db.QueryRow(
		"SELECT pool_id FROM season_picks WHERE pick_id = $1 AND user_id = $2",
		pickID, userID,
	).Scan(&currentPoolID)

	if err != nil {
		if err == sql.ErrNoRows {
			response.NotFound(c, "pick_not_found", "Pick not found or you don't have permission to update it")
			return
		}
		h.logger.Error("Failed to verify pick ownership", "error", err)
		response.InternalServerError(c, "pick_verification_failed", "Failed to verify pick ownership")
		return
	}

	// Check if new team is available (excluding current pick)
	var teamTaken bool
	err = h.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM season_picks WHERE pool_id = $1 AND team_id = $2 AND pick_id != $3)",
		currentPoolID, req.TeamID, pickID,
	).Scan(&teamTaken)

	if err != nil {
		h.logger.Error("Failed to check team availability", "error", err)
		response.InternalServerError(c, "team_check_failed", "Failed to verify team availability")
		return
	}

	if teamTaken {
		response.Conflict(c, "team_already_picked", "This team has already been picked by another user")
		return
	}

	// Update the pick
	_, err = h.db.Exec(`
		UPDATE season_picks 
		SET team_id = $1, pick_order = $2, updated_at = CURRENT_TIMESTAMP
		WHERE pick_id = $3`,
		req.TeamID, req.PickOrder, pickID,
	)

	if err != nil {
		h.logger.Error("Failed to update pick", "error", err)
		response.InternalServerError(c, "pick_update_failed", "Failed to update pick")
		return
	}

	// Fetch updated pick
	var pick models.PickWithTeam
	err = h.db.QueryRow(`
		SELECT p.pick_id, p.pool_id, p.user_id, p.team_id, p.pick_order,
		       p.points_scored, p.is_eliminated, p.elimination_week, p.created_at, p.updated_at,
		       t.team_name, t.team_abbreviation, t.city, t.conference, t.division,
		       t.logo_url, t.primary_color, t.secondary_color
		FROM season_picks p
		JOIN nfl_teams t ON p.team_id = t.team_id
		WHERE p.pick_id = $1`,
		pickID,
	).Scan(
		&pick.PickID, &pick.PoolID, &pick.UserID, &pick.TeamID, &pick.PickOrder,
		&pick.PointsScored, &pick.IsEliminated, &pick.EliminationWeek, &pick.CreatedAt, &pick.UpdatedAt,
		&pick.Team.TeamName, &pick.Team.TeamAbbreviation, &pick.Team.City,
		&pick.Team.Conference, &pick.Team.Division, &pick.Team.LogoURL,
		&pick.Team.PrimaryColor, &pick.Team.SecondaryColor,
	)

	if err != nil {
		h.logger.Error("Failed to fetch updated pick", "error", err)
		response.InternalServerError(c, "pick_fetch_failed", "Pick updated but failed to fetch details")
		return
	}

	pick.Team.TeamID = pick.TeamID

	h.logger.Info("Pick updated successfully", "pick_id", pickID, "user_id", userID)
	response.Success(c, pick, "Pick updated successfully")
}

// Delete deletes a pick
func (h *PickHandler) Delete(c *gin.Context) {
	pickID := c.Param("id")
	userID, _ := c.Get("user_id")

	// Verify user owns this pick
	var exists bool
	err := h.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM season_picks WHERE pick_id = $1 AND user_id = $2)",
		pickID, userID,
	).Scan(&exists)

	if err != nil {
		h.logger.Error("Failed to verify pick ownership", "error", err)
		response.InternalServerError(c, "pick_verification_failed", "Failed to verify pick ownership")
		return
	}

	if !exists {
		response.NotFound(c, "pick_not_found", "Pick not found or you don't have permission to delete it")
		return
	}

	// Delete the pick
	_, err = h.db.Exec("DELETE FROM season_picks WHERE pick_id = $1", pickID)
	if err != nil {
		h.logger.Error("Failed to delete pick", "error", err)
		response.InternalServerError(c, "pick_deletion_failed", "Failed to delete pick")
		return
	}

	h.logger.Info("Pick deleted successfully", "pick_id", pickID, "user_id", userID)
	response.Success(c, nil, "Pick deleted successfully")
}

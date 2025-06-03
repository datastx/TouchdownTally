package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"touchdown-tally/internal/config"
	"touchdown-tally/internal/models"
	"touchdown-tally/pkg/logger"
	"touchdown-tally/pkg/response"
)

type StandingHandler struct {
	db     *sql.DB
	config *config.Config
	logger *logger.Logger
}

func NewStandingHandler(db *sql.DB, config *config.Config, logger *logger.Logger) *StandingHandler {
	return &StandingHandler{
		db:     db,
		config: config,
		logger: logger,
	}
}

// GetPoolStandings returns the standings for a specific pool
func (h *StandingHandler) GetPoolStandings(c *gin.Context) {
	poolID := c.Param("id")
	userID := c.GetString("user_id")

	// Verify user has access to this pool
	var memberRole string
	err := h.db.QueryRow(`
		SELECT role FROM pool_memberships 
		WHERE pool_id = $1 AND user_id = $2 AND is_active = true
	`, poolID, userID).Scan(&memberRole)

	if err == sql.ErrNoRows {
		response.Error(c, http.StatusForbidden, "Access denied to this pool")
		return
	}
	if err != nil {
		h.logger.Error("Failed to check pool membership", "pool_id", poolID, "user_id", userID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to validate access")
		return
	}

	// Get optional week parameter for weekly standings
	weekStr := c.Query("week")
	var week *int
	if weekStr != "" {
		if w, err := strconv.Atoi(weekStr); err == nil {
			week = &w
		}
	}

	// Get pool type to determine standings calculation
	var poolType string
	var season int
	err = h.db.QueryRow(`
		SELECT pool_type, season FROM pools WHERE id = $1
	`, poolID).Scan(&poolType, &season)

	if err != nil {
		h.logger.Error("Failed to get pool info", "pool_id", poolID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve pool information")
		return
	}

	var standings []models.Standing
	if poolType == "season" {
		standings, err = h.getSeasonStandings(poolID, season)
	} else {
		if week == nil {
			// Get current week if not specified
			currentWeek, err := h.getCurrentWeek(season)
			if err != nil {
				h.logger.Error("Failed to get current week", "season", season, "error", err)
				response.Error(c, http.StatusInternalServerError, "Failed to determine current week")
				return
			}
			week = &currentWeek
		}
		standings, err = h.getWeeklyStandings(poolID, season, *week)
	}

	if err != nil {
		h.logger.Error("Failed to get standings", "pool_id", poolID, "pool_type", poolType, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve standings")
		return
	}

	response.Success(c, gin.H{
		"pool_id":   poolID,
		"pool_type": poolType,
		"season":    season,
		"week":      week,
		"standings": standings,
	})
}

// GetUserStats returns detailed statistics for a user in a pool
func (h *StandingHandler) GetUserStats(c *gin.Context) {
	poolID := c.Param("id")
	targetUserID := c.Param("userId")
	userID := c.GetString("user_id")

	// Verify requesting user has access to this pool
	var memberRole string
	err := h.db.QueryRow(`
		SELECT role FROM pool_memberships 
		WHERE pool_id = $1 AND user_id = $2 AND is_active = true
	`, poolID, userID).Scan(&memberRole)

	if err == sql.ErrNoRows {
		response.Error(c, http.StatusForbidden, "Access denied to this pool")
		return
	}
	if err != nil {
		h.logger.Error("Failed to check pool membership", "pool_id", poolID, "user_id", userID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to validate access")
		return
	}

	// Verify target user is also in the pool
	var targetExists bool
	err = h.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM pool_memberships 
		WHERE pool_id = $1 AND user_id = $2 AND is_active = true)
	`, poolID, targetUserID).Scan(&targetExists)

	if err != nil {
		h.logger.Error("Failed to check target user membership", "pool_id", poolID, "target_user_id", targetUserID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to validate target user")
		return
	}

	if !targetExists {
		response.Error(c, http.StatusNotFound, "User not found in this pool")
		return
	}

	// Get pool info
	var poolType string
	var season int
	err = h.db.QueryRow(`
		SELECT pool_type, season FROM pools WHERE id = $1
	`, poolID).Scan(&poolType, &season)

	if err != nil {
		h.logger.Error("Failed to get pool info", "pool_id", poolID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve pool information")
		return
	}

	stats, err := h.getUserDetailedStats(poolID, targetUserID, season, poolType)
	if err != nil {
		h.logger.Error("Failed to get user stats", "pool_id", poolID, "user_id", targetUserID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve user statistics")
		return
	}

	response.Success(c, stats)
}

// Helper functions

func (h *StandingHandler) getSeasonStandings(poolID string, season int) ([]models.Standing, error) {
	rows, err := h.db.Query(`
		WITH user_scores AS (
			SELECT 
				pm.user_id,
				up.display_name,
				COUNT(sp.id) as total_picks,
				SUM(CASE 
					WHEN g.status = 'final' AND 
					     ((sp.picked_team_id = g.home_team_id AND g.home_score > g.away_score) OR
					      (sp.picked_team_id = g.away_team_id AND g.away_score > g.home_score))
					THEN 1 ELSE 0 
				END) as correct_picks,
				SUM(CASE 
					WHEN g.status = 'final' AND 
					     ((sp.picked_team_id = g.home_team_id AND g.home_score < g.away_score) OR
					      (sp.picked_team_id = g.away_team_id AND g.away_score < g.home_score))
					THEN 1 ELSE 0 
				END) as incorrect_picks,
				ROUND(
					CASE 
						WHEN COUNT(sp.id) > 0 
						THEN (SUM(CASE 
							WHEN g.status = 'final' AND 
							     ((sp.picked_team_id = g.home_team_id AND g.home_score > g.away_score) OR
							      (sp.picked_team_id = g.away_team_id AND g.away_score > g.home_score))
							THEN 1.0 ELSE 0.0 
						END) / COUNT(sp.id)) * 100
						ELSE 0 
					END, 2
				) as win_percentage
			FROM pool_memberships pm
			JOIN user_profiles up ON pm.user_id = up.id
			LEFT JOIN season_picks sp ON pm.user_id = sp.user_id AND pm.pool_id = sp.pool_id
			LEFT JOIN nfl_games g ON sp.game_id = g.id
			WHERE pm.pool_id = $1 AND pm.is_active = true AND g.season = $2
			GROUP BY pm.user_id, up.display_name
		)
		SELECT 
			user_id, display_name, total_picks, correct_picks, 
			incorrect_picks, win_percentage
		FROM user_scores
		ORDER BY correct_picks DESC, win_percentage DESC, display_name ASC
	`, poolID, season)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var standings []models.Standing
	position := 1
	
	for rows.Next() {
		var standing models.Standing
		err := rows.Scan(
			&standing.UserID, &standing.DisplayName, &standing.TotalPicks,
			&standing.CorrectPicks, &standing.IncorrectPicks, &standing.WinPercentage,
		)
		if err != nil {
			return nil, err
		}
		
		standing.Position = position
		standings = append(standings, standing)
		position++
	}

	return standings, rows.Err()
}

func (h *StandingHandler) getWeeklyStandings(poolID string, season, week int) ([]models.Standing, error) {
	rows, err := h.db.Query(`
		WITH user_scores AS (
			SELECT 
				pm.user_id,
				up.display_name,
				COUNT(sp.id) as total_picks,
				SUM(CASE 
					WHEN g.status = 'final' AND 
					     ((sp.picked_team_id = g.home_team_id AND g.home_score > g.away_score) OR
					      (sp.picked_team_id = g.away_team_id AND g.away_score > g.home_score))
					THEN 1 ELSE 0 
				END) as correct_picks,
				SUM(CASE 
					WHEN g.status = 'final' AND 
					     ((sp.picked_team_id = g.home_team_id AND g.home_score < g.away_score) OR
					      (sp.picked_team_id = g.away_team_id AND g.away_score < g.home_score))
					THEN 1 ELSE 0 
				END) as incorrect_picks,
				ROUND(
					CASE 
						WHEN COUNT(sp.id) > 0 
						THEN (SUM(CASE 
							WHEN g.status = 'final' AND 
							     ((sp.picked_team_id = g.home_team_id AND g.home_score > g.away_score) OR
							      (sp.picked_team_id = g.away_team_id AND g.away_score > g.home_score))
							THEN 1.0 ELSE 0.0 
						END) / COUNT(sp.id)) * 100
						ELSE 0 
					END, 2
				) as win_percentage
			FROM pool_memberships pm
			JOIN user_profiles up ON pm.user_id = up.id
			LEFT JOIN season_picks sp ON pm.user_id = sp.user_id AND pm.pool_id = sp.pool_id
			LEFT JOIN nfl_games g ON sp.game_id = g.id
			WHERE pm.pool_id = $1 AND pm.is_active = true 
			AND g.season = $2 AND g.week = $3
			GROUP BY pm.user_id, up.display_name
		)
		SELECT 
			user_id, display_name, total_picks, correct_picks, 
			incorrect_picks, win_percentage
		FROM user_scores
		ORDER BY correct_picks DESC, win_percentage DESC, display_name ASC
	`, poolID, season, week)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var standings []models.Standing
	position := 1
	
	for rows.Next() {
		var standing models.Standing
		err := rows.Scan(
			&standing.UserID, &standing.DisplayName, &standing.TotalPicks,
			&standing.CorrectPicks, &standing.IncorrectPicks, &standing.WinPercentage,
		)
		if err != nil {
			return nil, err
		}
		
		standing.Position = position
		standing.Week = &week
		standings = append(standings, standing)
		position++
	}

	return standings, rows.Err()
}

func (h *StandingHandler) getCurrentWeek(season int) (int, error) {
	var week int
	err := h.db.QueryRow(`
		SELECT COALESCE(MIN(week), 1) 
		FROM nfl_games 
		WHERE season = $1 AND status = 'scheduled'
		ORDER BY week ASC
	`, season).Scan(&week)
	
	if err == sql.ErrNoRows {
		// If no scheduled games, return the latest week
		err = h.db.QueryRow(`
			SELECT COALESCE(MAX(week), 1) 
			FROM nfl_games 
			WHERE season = $1
		`, season).Scan(&week)
	}
	
	return week, err
}

func (h *StandingHandler) getUserDetailedStats(poolID, userID string, season int, poolType string) (*models.UserStats, error) {
	stats := &models.UserStats{
		UserID:   userID,
		PoolID:   poolID,
		Season:   season,
		PoolType: poolType,
	}

	// Get user display name
	err := h.db.QueryRow(`
		SELECT display_name FROM user_profiles WHERE id = $1
	`, userID).Scan(&stats.DisplayName)
	if err != nil {
		return nil, err
	}

	// Get overall stats
	err = h.db.QueryRow(`
		SELECT 
			COUNT(sp.id) as total_picks,
			SUM(CASE 
				WHEN g.status = 'final' AND 
				     ((sp.picked_team_id = g.home_team_id AND g.home_score > g.away_score) OR
				      (sp.picked_team_id = g.away_team_id AND g.away_score > g.home_score))
				THEN 1 ELSE 0 
			END) as correct_picks,
			SUM(CASE 
				WHEN g.status = 'final' AND 
				     ((sp.picked_team_id = g.home_team_id AND g.home_score < g.away_score) OR
				      (sp.picked_team_id = g.away_team_id AND g.away_score < g.home_score))
				THEN 1 ELSE 0 
			END) as incorrect_picks,
			ROUND(
				CASE 
					WHEN COUNT(sp.id) > 0 
					THEN (SUM(CASE 
						WHEN g.status = 'final' AND 
						     ((sp.picked_team_id = g.home_team_id AND g.home_score > g.away_score) OR
						      (sp.picked_team_id = g.away_team_id AND g.away_score > g.home_score))
						THEN 1.0 ELSE 0.0 
					END) / COUNT(sp.id)) * 100
					ELSE 0 
				END, 2
			) as win_percentage
		FROM season_picks sp
		JOIN nfl_games g ON sp.game_id = g.id
		WHERE sp.pool_id = $1 AND sp.user_id = $2 AND g.season = $3
	`, poolID, userID, season).Scan(
		&stats.TotalPicks, &stats.CorrectPicks, 
		&stats.IncorrectPicks, &stats.WinPercentage,
	)
	if err != nil {
		return nil, err
	}

	// Get weekly breakdown if it's a season pool
	if poolType == "season" {
		weeklyStats, err := h.getUserWeeklyStats(poolID, userID, season)
		if err != nil {
			return nil, err
		}
		stats.WeeklyStats = weeklyStats
	}

	// Get recent picks
	recentPicks, err := h.getUserRecentPicks(poolID, userID, season, 10)
	if err != nil {
		return nil, err
	}
	stats.RecentPicks = recentPicks

	return stats, nil
}

func (h *StandingHandler) getUserWeeklyStats(poolID, userID string, season int) ([]models.WeeklyStats, error) {
	rows, err := h.db.Query(`
		SELECT 
			g.week,
			COUNT(sp.id) as total_picks,
			SUM(CASE 
				WHEN g.status = 'final' AND 
				     ((sp.picked_team_id = g.home_team_id AND g.home_score > g.away_score) OR
				      (sp.picked_team_id = g.away_team_id AND g.away_score > g.home_score))
				THEN 1 ELSE 0 
			END) as correct_picks
		FROM nfl_games g
		LEFT JOIN season_picks sp ON g.id = sp.game_id 
			AND sp.pool_id = $1 AND sp.user_id = $2
		WHERE g.season = $3
		GROUP BY g.week
		ORDER BY g.week ASC
	`, poolID, userID, season)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var weeklyStats []models.WeeklyStats
	for rows.Next() {
		var ws models.WeeklyStats
		err := rows.Scan(&ws.Week, &ws.TotalPicks, &ws.CorrectPicks)
		if err != nil {
			return nil, err
		}
		ws.IncorrectPicks = ws.TotalPicks - ws.CorrectPicks
		if ws.TotalPicks > 0 {
			ws.WinPercentage = float64(ws.CorrectPicks) / float64(ws.TotalPicks) * 100
		}
		weeklyStats = append(weeklyStats, ws)
	}

	return weeklyStats, rows.Err()
}

func (h *StandingHandler) getUserRecentPicks(poolID, userID string, season, limit int) ([]models.RecentPick, error) {
	rows, err := h.db.Query(`
		SELECT 
			g.id, g.week, g.game_date, g.status,
			ht.abbreviation as home_team, ht.name as home_team_name,
			at.abbreviation as away_team, at.name as away_team_name,
			g.home_score, g.away_score,
			pt.abbreviation as picked_team, pt.name as picked_team_name,
			sp.confidence,
			CASE 
				WHEN g.status = 'final' AND 
				     ((sp.picked_team_id = g.home_team_id AND g.home_score > g.away_score) OR
				      (sp.picked_team_id = g.away_team_id AND g.away_score > g.home_score))
				THEN true
				WHEN g.status = 'final'
				THEN false
				ELSE null
			END as is_correct
		FROM season_picks sp
		JOIN nfl_games g ON sp.game_id = g.id
		JOIN nfl_teams ht ON g.home_team_id = ht.id
		JOIN nfl_teams at ON g.away_team_id = at.id
		JOIN nfl_teams pt ON sp.picked_team_id = pt.id
		WHERE sp.pool_id = $1 AND sp.user_id = $2 AND g.season = $3
		ORDER BY g.game_date DESC, g.week DESC
		LIMIT $4
	`, poolID, userID, season, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recentPicks []models.RecentPick
	for rows.Next() {
		var rp models.RecentPick
		err := rows.Scan(
			&rp.GameID, &rp.Week, &rp.GameDate, &rp.Status,
			&rp.HomeTeam, &rp.HomeTeamName, &rp.AwayTeam, &rp.AwayTeamName,
			&rp.HomeScore, &rp.AwayScore, &rp.PickedTeam, &rp.PickedTeamName,
			&rp.Confidence, &rp.IsCorrect,
		)
		if err != nil {
			return nil, err
		}
		recentPicks = append(recentPicks, rp)
	}

	return recentPicks, rows.Err()
}

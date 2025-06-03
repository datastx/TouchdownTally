package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"touchdown-tally/internal/config"
	"touchdown-tally/internal/models"
	"touchdown-tally/pkg/logger"
	"touchdown-tally/pkg/response"
)

type PoolHandler struct {
	db     *sql.DB
	config *config.Config
	logger *logger.Logger
}

func NewPoolHandler(db *sql.DB, config *config.Config, logger *logger.Logger) *PoolHandler {
	return &PoolHandler{
		db:     db,
		config: config,
		logger: logger,
	}
}

// CreatePool creates a new pool
func (h *PoolHandler) CreatePool(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	
	userID, ok := userIDInterface.(int)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Invalid user context")
		return
	}

	var req models.CreatePoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	// For now, allow any authenticated user to create pools
	// In the future, we can add role-based permissions

	// Create pool
	var poolID int64
	
	// Marshal JSON fields
	prizeStructureJSON, err := json.Marshal(req.PrizeStructure)
	if err != nil {
		h.logger.Error("Failed to marshal prize structure", "user_id", userID, "error", err)
		response.Error(c, http.StatusBadRequest, "Invalid prize structure")
		return
	}
	
	settingsJSON, err := json.Marshal(req.Settings)
	if err != nil {
		h.logger.Error("Failed to marshal settings", "user_id", userID, "error", err)
		response.Error(c, http.StatusBadRequest, "Invalid settings")
		return
	}
	
	result, err := h.db.Exec(`
		INSERT INTO pools (pool_name, commissioner_id, season_year, max_members, entry_fee, prize_structure, pool_type, settings)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, req.PoolName, userID, req.SeasonYear, req.MaxMembers, req.EntryFee, 
		string(prizeStructureJSON), "survivor", string(settingsJSON))

	if err != nil {
		h.logger.Error("Failed to create pool", "user_id", userID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to create pool")
		return
	}

	poolID, err = result.LastInsertId()
	if err != nil {
		h.logger.Error("Failed to get pool ID", "user_id", userID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to create pool")
		return
	}

	// Add creator as pool commissioner (role_id=1 for commissioner)
	_, err = h.db.Exec(`
		INSERT INTO pool_memberships (pool_id, user_id, role_id, joined_at)
		VALUES (?, ?, 1, CURRENT_TIMESTAMP)
	`, poolID, userID)

	if err != nil {
		h.logger.Error("Failed to add creator to pool", "pool_id", poolID, "user_id", userID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to complete pool creation")
		return
	}

	// Get the created pool
	pool, err := h.getPoolByID(poolID)
	if err != nil {
		h.logger.Error("Failed to retrieve created pool", "pool_id", poolID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Pool created but failed to retrieve details")
		return
	}

	response.Success(c, pool)
}

// GetPools returns pools the user is a member of or can join
func (h *PoolHandler) GetPools(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	
	userID, ok := userIDInterface.(int)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Invalid user context")
		return
	}

	// Get pools user is a member of
	memberPools, err := h.getUserPools(userID)
	if err != nil {
		h.logger.Error("Failed to get user pools", "user_id", userID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve pools")
		return
	}

	// Get available pools (not full, user not already a member)
	availablePools, err := h.getAvailablePools(userID)
	if err != nil {
		h.logger.Error("Failed to get available pools", "user_id", userID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve available pools")
		return
	}

	response.Success(c, gin.H{
		"member_pools":    memberPools,
		"available_pools": availablePools,
	})
}

// GetPool returns details for a specific pool
func (h *PoolHandler) GetPool(c *gin.Context) {
	poolID := c.Param("id")
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	
	userID, ok := userIDInterface.(int)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Invalid user context")
		return
	}

	// Check if user has access to this pool
	var memberRole string
	err := h.db.QueryRow(`
		SELECT r.role_name FROM pool_memberships pm 
		JOIN roles r ON pm.role_id = r.role_id
		WHERE pm.pool_id = ? AND pm.user_id = ?
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

	poolIDInt64, err := strconv.ParseInt(poolID, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid pool ID")
		return
	}

	pool, err := h.getPoolByID(poolIDInt64)
	if err != nil {
		h.logger.Error("Failed to get pool", "pool_id", poolID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve pool")
		return
	}

	// Get pool members
	members, err := h.getPoolMembers(poolID)
	if err != nil {
		h.logger.Error("Failed to get pool members", "pool_id", poolID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to retrieve pool members")
		return
	}

	pool.Members = members
	pool.UserRole = memberRole

	response.Success(c, pool)
}

// JoinPool allows a user to join a pool
func (h *PoolHandler) JoinPool(c *gin.Context) {
	poolID := c.Param("id")
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	
	userID, ok := userIDInterface.(int)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Invalid user context")
		return
	}

	// Check if pool exists and is not full
	var currentMembers, maxMembers int
	var status string
	err := h.db.QueryRow(`
		SELECT 
			(SELECT COUNT(*) FROM pool_memberships WHERE pool_id = p.pool_id),
			p.max_members,
			p.status
		FROM pools p WHERE p.pool_id = ?
	`, poolID).Scan(&currentMembers, &maxMembers, &status)

	if err == sql.ErrNoRows {
		response.Error(c, http.StatusNotFound, "Pool not found")
		return
	}
	if err != nil {
		h.logger.Error("Failed to check pool capacity", "pool_id", poolID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to validate pool")
		return
	}

	if status != "active" {
		response.Error(c, http.StatusBadRequest, "Pool is not active")
		return
	}

	if currentMembers >= maxMembers {
		response.Error(c, http.StatusBadRequest, "Pool is full")
		return
	}

	// Check if user is already a member
	var existingMember bool
	err = h.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM pool_memberships 
		WHERE pool_id = ? AND user_id = ?)
	`, poolID, userID).Scan(&existingMember)

	if err != nil {
		h.logger.Error("Failed to check existing membership", "pool_id", poolID, "user_id", userID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to validate membership")
		return
	}

	if existingMember {
		response.Error(c, http.StatusBadRequest, "Already a member of this pool")
		return
	}

	// Add user to pool (role_id 2 = member)
	_, err = h.db.Exec(`
		INSERT INTO pool_memberships (pool_id, user_id, role_id, joined_at)
		VALUES (?, ?, 2, CURRENT_TIMESTAMP)
	`, poolID, userID)

	if err != nil {
		h.logger.Error("Failed to join pool", "pool_id", poolID, "user_id", userID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to join pool")
		return
	}

	response.Success(c, gin.H{"message": "Successfully joined pool"})
}

// LeavePool allows a user to leave a pool
func (h *PoolHandler) LeavePool(c *gin.Context) {
	poolID := c.Param("id")
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	
	userID, ok := userIDInterface.(int)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "Invalid user context")
		return
	}

	// Check if user is a member and get their role
	var roleID int
	err := h.db.QueryRow(`
		SELECT role_id FROM pool_memberships 
		WHERE pool_id = ? AND user_id = ?
	`, poolID, userID).Scan(&roleID)

	if err == sql.ErrNoRows {
		response.Error(c, http.StatusBadRequest, "Not a member of this pool")
		return
	}
	if err != nil {
		h.logger.Error("Failed to check membership", "pool_id", poolID, "user_id", userID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to validate membership")
		return
	}

	// Check if this is the only commissioner (role_id 1 = commissioner)
	if roleID == 1 {
		var commissionerCount int
		err = h.db.QueryRow(`
			SELECT COUNT(*) FROM pool_memberships 
			WHERE pool_id = ? AND role_id = 1
		`, poolID).Scan(&commissionerCount)

		if err != nil {
			h.logger.Error("Failed to count commissioners", "pool_id", poolID, "error", err)
			response.Error(c, http.StatusInternalServerError, "Failed to validate commissioner status")
			return
		}

		if commissionerCount <= 1 {
			response.Error(c, http.StatusBadRequest, "Cannot leave pool as the only commissioner")
			return
		}
	}

	// Remove user from pool
	_, err = h.db.Exec(`
		DELETE FROM pool_memberships 
		WHERE pool_id = ? AND user_id = ?
	`, poolID, userID)

	if err != nil {
		h.logger.Error("Failed to leave pool", "pool_id", poolID, "user_id", userID, "error", err)
		response.Error(c, http.StatusInternalServerError, "Failed to leave pool")
		return
	}

	response.Success(c, gin.H{"message": "Successfully left pool"})
}

// Helper functions

func (h *PoolHandler) getPoolByID(poolID int64) (*models.Pool, error) {
	pool := &models.Pool{}
	
	err := h.db.QueryRow(`
		SELECT 
			p.pool_id, p.pool_name, COALESCE(p.description, ''), p.max_members, p.season_year, p.pool_type,
			p.entry_fee, p.status, p.created_at, p.updated_at,
			up.display_name as creator_name,
			(SELECT COUNT(*) FROM pool_memberships WHERE pool_id = p.pool_id) as current_members
		FROM pools p
		JOIN user_profiles up ON p.commissioner_id = up.user_id
		WHERE p.pool_id = ?
	`, poolID).Scan(
		&pool.ID, &pool.Name, &pool.Description, &pool.MaxPlayers, &pool.Season,
		&pool.PoolType, &pool.EntryFee, &pool.IsActive, &pool.CreatedAt, &pool.UpdatedAt,
		&pool.CreatorName, &pool.CurrentMembers,
	)

	return pool, err
}

func (h *PoolHandler) getUserPools(userID int) ([]models.Pool, error) {
	rows, err := h.db.Query(`
		SELECT 
			p.pool_id, p.pool_name, COALESCE(p.description, ''), p.max_members, p.season_year, p.pool_type,
			p.entry_fee, p.status, p.created_at, p.updated_at,
			up.display_name as creator_name,
			r.role_name as user_role,
			(SELECT COUNT(*) FROM pool_memberships WHERE pool_id = p.pool_id) as current_members
		FROM pools p
		JOIN pool_memberships pm ON p.pool_id = pm.pool_id
		JOIN user_profiles up ON p.commissioner_id = up.user_id
		JOIN roles r ON pm.role_id = r.role_id
		WHERE pm.user_id = ?
		ORDER BY p.created_at DESC
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pools []models.Pool
	for rows.Next() {
		var pool models.Pool
		err := rows.Scan(
			&pool.ID, &pool.Name, &pool.Description, &pool.MaxPlayers, &pool.Season,
			&pool.PoolType, &pool.EntryFee, &pool.IsActive, &pool.CreatedAt, &pool.UpdatedAt,
			&pool.CreatorName, &pool.UserRole, &pool.CurrentMembers,
		)
		if err != nil {
			return nil, err
		}
		pools = append(pools, pool)
	}

	return pools, rows.Err()
}

func (h *PoolHandler) getAvailablePools(userID int) ([]models.Pool, error) {
	rows, err := h.db.Query(`
		SELECT 
			p.pool_id, p.pool_name, COALESCE(p.description, ''), p.max_members, p.season_year, p.pool_type,
			p.entry_fee, p.status, p.created_at, p.updated_at,
			up.display_name as creator_name,
			(SELECT COUNT(*) FROM pool_memberships WHERE pool_id = p.pool_id) as current_members
		FROM pools p
		JOIN user_profiles up ON p.commissioner_id = up.user_id
		WHERE p.status = 'active'
		AND p.pool_id NOT IN (
			SELECT pool_id FROM pool_memberships 
			WHERE user_id = ?
		)
		AND (SELECT COUNT(*) FROM pool_memberships WHERE pool_id = p.pool_id) < p.max_members
		ORDER BY p.created_at DESC
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pools []models.Pool
	for rows.Next() {
		var pool models.Pool
		err := rows.Scan(
			&pool.ID, &pool.Name, &pool.Description, &pool.MaxPlayers, &pool.Season,
			&pool.PoolType, &pool.EntryFee, &pool.IsActive, &pool.CreatedAt, &pool.UpdatedAt,
			&pool.CreatorName, &pool.CurrentMembers,
		)
		if err != nil {
			return nil, err
		}
		pools = append(pools, pool)
	}

	return pools, rows.Err()
}

func (h *PoolHandler) getPoolMembers(poolID string) ([]models.PoolMember, error) {
	rows, err := h.db.Query(`
		SELECT 
			pm.user_id, r.role_name, pm.joined_at,
			up.display_name, up.username
		FROM pool_memberships pm
		JOIN user_profiles up ON pm.user_id = up.user_id
		JOIN roles r ON pm.role_id = r.role_id
		WHERE pm.pool_id = ?
		ORDER BY pm.joined_at ASC
	`, poolID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.PoolMember
	for rows.Next() {
		var member models.PoolMember
		err := rows.Scan(
			&member.UserID, &member.Role, &member.JoinedAt,
			&member.DisplayName, &member.Email,
		)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, rows.Err()
}

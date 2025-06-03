package handlers

import (
	"database/sql"

	"touchdown-tally/internal/auth"
	"touchdown-tally/internal/config"
	"touchdown-tally/internal/models"
	"touchdown-tally/pkg/logger"
	"touchdown-tally/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	db     *sql.DB
	config *config.Config
	logger *logger.Logger
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(db *sql.DB, cfg *config.Config, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid_request", err.Error())
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		h.logger.Error("Failed to hash password", "error", err)
		response.InternalServerError(c, "password_hash_failed", "Failed to process password")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		h.logger.Error("Failed to begin transaction", "error", err)
		response.InternalServerError(c, "transaction_failed", "Failed to start registration process")
		return
	}
	defer tx.Rollback()

	// Create email account
	var emailID int
	err = tx.QueryRow(
		"INSERT INTO email_accounts (email_address, password_hash) VALUES ($1, $2) RETURNING email_id",
		req.Email, hashedPassword,
	).Scan(&emailID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" { // unique violation
			response.Conflict(c, "email_exists", "Email address already registered")
			return
		}
		h.logger.Error("Failed to create email account", "error", err)
		response.InternalServerError(c, "account_creation_failed", "Failed to create account")
		return
	}

	// Create user profile
	var userID int
	err = tx.QueryRow(
		"INSERT INTO user_profiles (email_id, username, display_name) VALUES ($1, $2, $3) RETURNING user_id",
		emailID, req.Username, req.DisplayName,
	).Scan(&userID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" { // unique violation
			response.Conflict(c, "username_exists", "Username already exists for this email")
			return
		}
		h.logger.Error("Failed to create user profile", "error", err)
		response.InternalServerError(c, "profile_creation_failed", "Failed to create user profile")
		return
	}

	if err = tx.Commit(); err != nil {
		h.logger.Error("Failed to commit transaction", "error", err)
		response.InternalServerError(c, "registration_failed", "Failed to complete registration")
		return
	}

	// Create user object for response
	user := models.UserProfile{
		UserID:      userID,
		EmailID:     emailID,
		Username:    req.Username,
		DisplayName: req.DisplayName,
	}

	// Generate JWT token
	token, err := auth.GenerateJWT(user, h.config.JWTSecret)
	if err != nil {
		h.logger.Error("Failed to generate JWT token", "error", err)
		response.InternalServerError(c, "token_generation_failed", "Failed to generate authentication token")
		return
	}

	// Get all profiles for this email (for now just the one we created)
	profiles := []models.UserProfile{user}

	loginResponse := models.LoginResponse{
		Token:    token,
		User:     user,
		Profiles: profiles,
	}

	h.logger.Info("User registered successfully", "user_id", userID, "email", req.Email)
	response.Created(c, loginResponse, "User registered successfully")
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid_request", err.Error())
		return
	}

	// Get email account
	var account models.EmailAccount
	err := h.db.QueryRow(
		"SELECT email_id, email_address, password_hash FROM email_accounts WHERE email_address = $1",
		req.Email,
	).Scan(&account.EmailID, &account.EmailAddress, &account.PasswordHash)

	if err != nil {
		if err == sql.ErrNoRows {
			response.Unauthorized(c, "invalid_credentials", "Invalid email or password")
			return
		}
		h.logger.Error("Failed to query email account", "error", err)
		response.InternalServerError(c, "login_failed", "Failed to process login")
		return
	}

	// Check password
	if err := auth.CheckPassword(req.Password, account.PasswordHash); err != nil {
		response.Unauthorized(c, "invalid_credentials", "Invalid email or password")
		return
	}

	// Get all profiles for this email
	rows, err := h.db.Query(
		"SELECT user_id, username, display_name, created_at FROM user_profiles WHERE email_id = $1 ORDER BY created_at",
		account.EmailID,
	)
	if err != nil {
		h.logger.Error("Failed to query user profiles", "error", err)
		response.InternalServerError(c, "login_failed", "Failed to load user profiles")
		return
	}
	defer rows.Close()

	var profiles []models.UserProfile
	for rows.Next() {
		var profile models.UserProfile
		err := rows.Scan(&profile.UserID, &profile.Username, &profile.DisplayName, &profile.CreatedAt)
		if err != nil {
			h.logger.Error("Failed to scan user profile", "error", err)
			continue
		}
		profile.EmailID = account.EmailID
		profiles = append(profiles, profile)
	}

	if len(profiles) == 0 {
		response.InternalServerError(c, "no_profiles", "No user profiles found")
		return
	}

	// Use the first profile as the default
	defaultProfile := profiles[0]

	// Generate JWT token
	token, err := auth.GenerateJWT(defaultProfile, h.config.JWTSecret)
	if err != nil {
		h.logger.Error("Failed to generate JWT token", "error", err)
		response.InternalServerError(c, "token_generation_failed", "Failed to generate authentication token")
		return
	}

	loginResponse := models.LoginResponse{
		Token:    token,
		User:     defaultProfile,
		Profiles: profiles,
	}

	h.logger.Info("User logged in successfully", "user_id", defaultProfile.UserID, "email", req.Email)
	response.Success(c, loginResponse, "Login successful")
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT system, logout is typically handled client-side
	// by removing the token from storage. However, we can log the event.
	
	if userID, exists := c.Get("user_id"); exists {
		h.logger.Info("User logged out", "user_id", userID)
	}

	response.Success(c, nil, "Logout successful")
}

// GetProfile returns the current user's profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "authentication_required", "User must be authenticated")
		return
	}

	var profile models.UserProfile
	err := h.db.QueryRow(
		"SELECT user_id, email_id, username, display_name, created_at FROM user_profiles WHERE user_id = $1",
		userID,
	).Scan(&profile.UserID, &profile.EmailID, &profile.Username, &profile.DisplayName, &profile.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			response.NotFound(c, "user_not_found", "User profile not found")
			return
		}
		h.logger.Error("Failed to query user profile", "error", err)
		response.InternalServerError(c, "profile_fetch_failed", "Failed to fetch user profile")
		return
	}

	response.Success(c, profile)
}

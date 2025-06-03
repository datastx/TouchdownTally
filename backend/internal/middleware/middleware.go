package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"touchdown-tally/internal/config"
	"touchdown-tally/pkg/logger"
	"touchdown-tally/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Logger returns a gin.HandlerFunc that logs requests
func Logger(logger *logger.Logger) gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			logger.HTTP(
				param.Method,
				param.Path,
				param.StatusCode,
				param.Latency.String(),
				"client_ip", param.ClientIP,
				"user_agent", param.Request.UserAgent(),
			)
			return ""
		},
	})
}

// CORS returns a gin.HandlerFunc that handles CORS
func CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin || allowedOrigin == "*" {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequireAuth returns a gin.HandlerFunc that requires JWT authentication
func RequireAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "authorization_required", "Authorization header is required")
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			response.Unauthorized(c, "invalid_authorization_format", "Authorization header must be in format 'Bearer <token>'")
			c.Abort()
			return
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			response.Unauthorized(c, "invalid_token", "Invalid or expired token")
			c.Abort()
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Set user context
			if emailID, ok := claims["email_id"]; ok {
				if emailIDFloat, ok := emailID.(float64); ok {
					c.Set("email_id", int(emailIDFloat))
				}
			}

			if userID, ok := claims["user_id"]; ok {
				if userIDFloat, ok := userID.(float64); ok {
					c.Set("user_id", int(userIDFloat))
				}
			}

			if username, ok := claims["username"]; ok {
				if usernameStr, ok := username.(string); ok {
					c.Set("username", usernameStr)
				}
			}
		}

		c.Next()
	}
}

// RequirePoolMember returns a gin.HandlerFunc that requires pool membership
func RequirePoolMember() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This middleware should be used after RequireAuth
		userID, exists := c.Get("user_id")
		if !exists {
			response.Unauthorized(c, "authentication_required", "User must be authenticated")
			c.Abort()
			return
		}

		// Get pool ID from URL parameter
		poolIDStr := c.Param("pool_id")
		if poolIDStr == "" {
			poolIDStr = c.Param("id") // Alternative parameter name
		}

		if poolIDStr == "" {
			response.BadRequest(c, "pool_id_required", "Pool ID is required")
			c.Abort()
			return
		}

		poolID, err := strconv.Atoi(poolIDStr)
		if err != nil {
			response.BadRequest(c, "invalid_pool_id", "Pool ID must be a valid integer")
			c.Abort()
			return
		}

		// TODO: Check if user is a member of the pool in the database
		// This would require a database query, which we'll implement in the handlers

		// Store pool ID in context for handlers to use
		c.Set("pool_id", poolID)
		c.Set("requesting_user_id", userID)

		c.Next()
	}
}

// RequirePoolAdmin returns a gin.HandlerFunc that requires pool admin rights
func RequirePoolAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This middleware should be used after RequireAuth and RequirePoolMember
		// TODO: Check if user has admin rights in the pool
		// This would require a database query to check the user's role

		c.Next()
	}
}

// RateLimiter returns a gin.HandlerFunc that implements rate limiting
func RateLimiter(requests int, window time.Duration) gin.HandlerFunc {
	// This is a simple in-memory rate limiter
	// For production, consider using Redis-based rate limiting
	clients := make(map[string][]time.Time)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// Clean old entries
		if timestamps, exists := clients[clientIP]; exists {
			var validTimestamps []time.Time
			for _, timestamp := range timestamps {
				if now.Sub(timestamp) < window {
					validTimestamps = append(validTimestamps, timestamp)
				}
			}
			clients[clientIP] = validTimestamps
		}

		// Check if limit exceeded
		if len(clients[clientIP]) >= requests {
			response.Error(c, http.StatusTooManyRequests, "rate_limit_exceeded", "Too many requests")
			c.Abort()
			return
		}

		// Add current request
		clients[clientIP] = append(clients[clientIP], now)

		c.Next()
	}
}

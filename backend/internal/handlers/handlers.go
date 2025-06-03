package handlers

import (
	"database/sql"

	"touchdown-tally/internal/config"
	"touchdown-tally/pkg/logger"
)

// Handlers aggregates all handler groups
type Handlers struct {
	Auth      *AuthHandler
	Pools     *PoolHandler
	Picks     *PickHandler
	Games     *GameHandler
	Teams     *TeamHandler
	Standings *StandingHandler
	Chat      *ChatHandler
}

// New creates a new Handlers instance with all handler groups
func New(db *sql.DB, cfg *config.Config, logger *logger.Logger) *Handlers {
	return &Handlers{
		Auth:      NewAuthHandler(db, cfg, logger),
		Pools:     NewPoolHandler(db, cfg, logger),
		Picks:     NewPickHandler(db, cfg, logger),
		Games:     NewGameHandler(db, cfg, logger),
		Teams:     NewTeamHandler(db, cfg, logger),
		Standings: NewStandingHandler(db, cfg, logger),
		Chat:      NewChatHandler(db, cfg, logger),
	}
}

package database

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Connect establishes a connection to the database (PostgreSQL or SQLite)
func Connect(databaseURL string) (*sql.DB, error) {
	var driverName, dataSourceName string
	
	if strings.HasPrefix(databaseURL, "sqlite://") {
		driverName = "sqlite3"
		dataSourceName = strings.TrimPrefix(databaseURL, "sqlite://")
	} else {
		driverName = "postgres"
		dataSourceName = databaseURL
	}

	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

// Migrate runs database migrations
func Migrate(db *sql.DB) error {
	// Detect database type
	driverName := db.Driver()
	isSQLite := fmt.Sprintf("%T", driverName) == "*sqlite3.SQLiteDriver"
	
	var migrations []string
	if isSQLite {
		migrations = []string{
			createEmailAccountsTableSQLite,
			createUserProfilesTableSQLite,
			createRolesTableSQLite,
			createNFLTeamsTableSQLite,
			createPoolsTableSQLite,
			createPoolMembershipsTableSQLite,
			createNFLGamesTableSQLite,
			createSeasonPicksTableSQLite,
			createChatMessagesTableSQLite,
			insertRoles,
			insertNFLTeams,
		}
	} else {
		migrations = []string{
			createEmailAccountsTable,
			createUserProfilesTable,
			createRolesTable,
			createNFLTeamsTable,
			createPoolsTable,
			createPoolMembershipsTable,
			createNFLGamesTable,
			createSeasonPicksTable,
			createChatMessagesTable,
			insertRoles,
			insertNFLTeams,
		}
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	return nil
}

// Migration SQL statements
const (
	createEmailAccountsTable = `
		CREATE TABLE IF NOT EXISTS email_accounts (
			email_id SERIAL PRIMARY KEY,
			email_address VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`

	createUserProfilesTable = `
		CREATE TABLE IF NOT EXISTS user_profiles (
			user_id SERIAL PRIMARY KEY,
			email_id INTEGER REFERENCES email_accounts(email_id) ON DELETE CASCADE,
			username VARCHAR(50) NOT NULL,
			display_name VARCHAR(100) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(email_id, username)
		);
	`

	createRolesTable = `
		CREATE TABLE IF NOT EXISTS roles (
			role_id SERIAL PRIMARY KEY,
			role_name VARCHAR(50) UNIQUE NOT NULL,
			description TEXT
		);
	`

	createNFLTeamsTable = `
		CREATE TABLE IF NOT EXISTS nfl_teams (
			team_id SERIAL PRIMARY KEY,
			team_name VARCHAR(100) NOT NULL,
			team_abbreviation VARCHAR(10) UNIQUE NOT NULL,
			city VARCHAR(100) NOT NULL,
			conference VARCHAR(10) NOT NULL CHECK (conference IN ('NFC', 'AFC')),
			division VARCHAR(10) NOT NULL,
			logo_url TEXT,
			primary_color VARCHAR(7),
			secondary_color VARCHAR(7),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`

	createPoolsTable = `
		CREATE TABLE IF NOT EXISTS pools (
			pool_id SERIAL PRIMARY KEY,
			pool_name VARCHAR(100) NOT NULL,
			pool_code VARCHAR(20) UNIQUE NOT NULL,
			season_year INTEGER NOT NULL,
			created_by INTEGER REFERENCES user_profiles(user_id) ON DELETE SET NULL,
			max_members INTEGER DEFAULT 50,
			entry_fee DECIMAL(10,2) DEFAULT 0.00,
			prize_structure JSONB,
			settings JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`

	createPoolMembershipsTable = `
		CREATE TABLE IF NOT EXISTS pool_memberships (
			membership_id SERIAL PRIMARY KEY,
			pool_id INTEGER REFERENCES pools(pool_id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES user_profiles(user_id) ON DELETE CASCADE,
			role_id INTEGER REFERENCES roles(role_id) DEFAULT 2, -- Default to member role
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(pool_id, user_id)
		);
	`

	createNFLGamesTable = `
		CREATE TABLE IF NOT EXISTS nfl_games (
			game_id SERIAL PRIMARY KEY,
			external_id VARCHAR(50) UNIQUE, -- MySportsFeeds game ID
			season_year INTEGER NOT NULL,
			week INTEGER NOT NULL,
			game_type VARCHAR(20) DEFAULT 'regular', -- regular, playoff, superbowl
			home_team_id INTEGER REFERENCES nfl_teams(team_id),
			away_team_id INTEGER REFERENCES nfl_teams(team_id),
			game_date TIMESTAMP NOT NULL,
			home_score INTEGER DEFAULT 0,
			away_score INTEGER DEFAULT 0,
			status VARCHAR(20) DEFAULT 'scheduled', -- scheduled, in_progress, completed, postponed
			quarter INTEGER DEFAULT 0,
			time_remaining VARCHAR(10),
			last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`

	createSeasonPicksTable = `
		CREATE TABLE IF NOT EXISTS season_picks (
			pick_id SERIAL PRIMARY KEY,
			pool_id INTEGER REFERENCES pools(pool_id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES user_profiles(user_id) ON DELETE CASCADE,
			team_id INTEGER REFERENCES nfl_teams(team_id),
			pick_order INTEGER CHECK (pick_order BETWEEN 1 AND 4),
			points_scored INTEGER DEFAULT 0,
			is_eliminated BOOLEAN DEFAULT FALSE,
			elimination_week INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(pool_id, user_id, pick_order),
			UNIQUE(pool_id, team_id) -- Each team can only be picked once per pool
		);
	`

	createChatMessagesTable = `
		CREATE TABLE IF NOT EXISTS chat_messages (
			message_id SERIAL PRIMARY KEY,
			pool_id INTEGER REFERENCES pools(pool_id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES user_profiles(user_id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			message_type VARCHAR(20) DEFAULT 'user_message', -- user_message, system_message, moderation_action
			is_deleted BOOLEAN DEFAULT FALSE,
			deleted_by INTEGER REFERENCES user_profiles(user_id),
			deleted_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`

	insertRoles = `
		INSERT INTO roles (role_name, description) VALUES 
		('commissioner', 'Pool commissioner with full administrative rights'),
		('member', 'Regular pool member'),
		('moderator', 'Chat moderator with limited administrative rights')
		ON CONFLICT (role_name) DO NOTHING;
	`

	insertNFLTeams = `
		INSERT INTO nfl_teams (team_name, team_abbreviation, city, conference, division, primary_color, secondary_color) VALUES
		('Cardinals', 'ARI', 'Arizona', 'NFC', 'West', '#97233F', '#000000'),
		('Falcons', 'ATL', 'Atlanta', 'NFC', 'South', '#A71930', '#000000'),
		('Ravens', 'BAL', 'Baltimore', 'AFC', 'North', '#241773', '#9E7C0C'),
		('Bills', 'BUF', 'Buffalo', 'AFC', 'East', '#00338D', '#C60C30'),
		('Panthers', 'CAR', 'Carolina', 'NFC', 'South', '#0085CA', '#101820'),
		('Bears', 'CHI', 'Chicago', 'NFC', 'North', '#0B162A', '#C83803'),
		('Bengals', 'CIN', 'Cincinnati', 'AFC', 'North', '#FB4F14', '#000000'),
		('Browns', 'CLE', 'Cleveland', 'AFC', 'North', '#311D00', '#FF3C00'),
		('Cowboys', 'DAL', 'Dallas', 'NFC', 'East', '#003594', '#041E42'),
		('Broncos', 'DEN', 'Denver', 'AFC', 'West', '#FB4F14', '#002244'),
		('Lions', 'DET', 'Detroit', 'NFC', 'North', '#0076B6', '#B0B7BC'),
		('Packers', 'GB', 'Green Bay', 'NFC', 'North', '#203731', '#FFB612'),
		('Texans', 'HOU', 'Houston', 'AFC', 'South', '#03202F', '#A71930'),
		('Colts', 'IND', 'Indianapolis', 'AFC', 'South', '#002C5F', '#A2AAAD'),
		('Jaguars', 'JAX', 'Jacksonville', 'AFC', 'South', '#006778', '#9F792C'),
		('Chiefs', 'KC', 'Kansas City', 'AFC', 'West', '#E31837', '#FFB81C'),
		('Raiders', 'LV', 'Las Vegas', 'AFC', 'West', '#000000', '#A5ACAF'),
		('Chargers', 'LAC', 'Los Angeles', 'AFC', 'West', '#0080C6', '#FFC20E'),
		('Rams', 'LAR', 'Los Angeles', 'NFC', 'West', '#003594', '#FFA300'),
		('Dolphins', 'MIA', 'Miami', 'AFC', 'East', '#008E97', '#FC4C02'),
		('Vikings', 'MIN', 'Minnesota', 'NFC', 'North', '#4F2683', '#FFC62F'),
		('Patriots', 'NE', 'New England', 'AFC', 'East', '#002244', '#C60C30'),
		('Saints', 'NO', 'New Orleans', 'NFC', 'South', '#D3BC8D', '#101820'),
		('Giants', 'NYG', 'New York', 'NFC', 'East', '#0B2265', '#A71930'),
		('Jets', 'NYJ', 'New York', 'AFC', 'East', '#125740', '#000000'),
		('Eagles', 'PHI', 'Philadelphia', 'NFC', 'East', '#004C54', '#A5ACAF'),
		('Steelers', 'PIT', 'Pittsburgh', 'AFC', 'North', '#FFB612', '#101820'),
		('49ers', 'SF', 'San Francisco', 'NFC', 'West', '#AA0000', '#B3995D'),
		('Seahawks', 'SEA', 'Seattle', 'NFC', 'West', '#002244', '#69BE28'),
		('Buccaneers', 'TB', 'Tampa Bay', 'NFC', 'South', '#D50A0A', '#FF7900'),
		('Titans', 'TEN', 'Tennessee', 'AFC', 'South', '#0C2340', '#4B92DB'),
		('Commanders', 'WAS', 'Washington', 'NFC', 'East', '#5A1414', '#FFB612')
		ON CONFLICT (team_abbreviation) DO NOTHING;
	`

	// SQLite-compatible migration SQL statements
	createEmailAccountsTableSQLite = `
		CREATE TABLE IF NOT EXISTS email_accounts (
			email_id INTEGER PRIMARY KEY AUTOINCREMENT,
			email_address TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	createUserProfilesTableSQLite = `
		CREATE TABLE IF NOT EXISTS user_profiles (
			user_id INTEGER PRIMARY KEY AUTOINCREMENT,
			email_id INTEGER REFERENCES email_accounts(email_id) ON DELETE CASCADE,
			username TEXT NOT NULL,
			display_name TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(email_id, username)
		);
	`

	createRolesTableSQLite = `
		CREATE TABLE IF NOT EXISTS roles (
			role_id INTEGER PRIMARY KEY AUTOINCREMENT,
			role_name TEXT UNIQUE NOT NULL,
			description TEXT
		);
	`

	createNFLTeamsTableSQLite = `
		CREATE TABLE IF NOT EXISTS nfl_teams (
			team_id INTEGER PRIMARY KEY AUTOINCREMENT,
			team_name TEXT NOT NULL,
			team_abbreviation TEXT UNIQUE NOT NULL,
			city TEXT NOT NULL,
			conference TEXT NOT NULL,
			division TEXT NOT NULL,
			primary_color TEXT,
			secondary_color TEXT,
			logo_url TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	createPoolsTableSQLite = `
		CREATE TABLE IF NOT EXISTS pools (
			pool_id INTEGER PRIMARY KEY AUTOINCREMENT,
			pool_name TEXT NOT NULL,
			description TEXT,
			commissioner_id INTEGER REFERENCES user_profiles(user_id),
			season_year INTEGER NOT NULL,
			max_members INTEGER DEFAULT 50,
			entry_fee REAL DEFAULT 0.00,
			prize_structure TEXT,
			pool_type TEXT DEFAULT 'survivor',
			status TEXT DEFAULT 'active',
			settings TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	createPoolMembershipsTableSQLite = `
		CREATE TABLE IF NOT EXISTS pool_memberships (
			membership_id INTEGER PRIMARY KEY AUTOINCREMENT,
			pool_id INTEGER REFERENCES pools(pool_id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES user_profiles(user_id) ON DELETE CASCADE,
			role_id INTEGER REFERENCES roles(role_id) DEFAULT 2,
			joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(pool_id, user_id)
		);
	`

	createNFLGamesTableSQLite = `
		CREATE TABLE IF NOT EXISTS nfl_games (
			game_id INTEGER PRIMARY KEY AUTOINCREMENT,
			external_id TEXT UNIQUE,
			season_year INTEGER NOT NULL,
			week INTEGER NOT NULL,
			game_type TEXT DEFAULT 'regular',
			home_team_id INTEGER REFERENCES nfl_teams(team_id),
			away_team_id INTEGER REFERENCES nfl_teams(team_id),
			game_date DATETIME NOT NULL,
			home_score INTEGER DEFAULT 0,
			away_score INTEGER DEFAULT 0,
			status TEXT DEFAULT 'scheduled',
			quarter INTEGER DEFAULT 0,
			time_remaining TEXT,
			last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	createSeasonPicksTableSQLite = `
		CREATE TABLE IF NOT EXISTS season_picks (
			pick_id INTEGER PRIMARY KEY AUTOINCREMENT,
			pool_id INTEGER REFERENCES pools(pool_id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES user_profiles(user_id) ON DELETE CASCADE,
			team_id INTEGER REFERENCES nfl_teams(team_id),
			pick_order INTEGER CHECK (pick_order BETWEEN 1 AND 4),
			points_scored INTEGER DEFAULT 0,
			is_eliminated INTEGER DEFAULT 0,
			elimination_week INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(pool_id, user_id, pick_order),
			UNIQUE(pool_id, team_id)
		);
	`

	createChatMessagesTableSQLite = `
		CREATE TABLE IF NOT EXISTS chat_messages (
			message_id INTEGER PRIMARY KEY AUTOINCREMENT,
			pool_id INTEGER REFERENCES pools(pool_id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES user_profiles(user_id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			message_type TEXT DEFAULT 'user_message',
			is_deleted INTEGER DEFAULT 0,
			deleted_by INTEGER REFERENCES user_profiles(user_id),
			deleted_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`
)

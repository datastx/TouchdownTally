-- Initialize the database with required extensions and basic structure
-- This file is automatically executed when the PostgreSQL container starts

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create basic tables structure (can be expanded with migrations)
-- Note: Full schema will be created through Go migrations

-- Create a simple health check table
CREATE TABLE IF NOT EXISTS health_check (
    id SERIAL PRIMARY KEY,
    status VARCHAR(50) DEFAULT 'healthy',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert initial health check record
INSERT INTO health_check (status) VALUES ('initialized');

-- Grant permissions to the application user
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO touchdown_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO touchdown_user;

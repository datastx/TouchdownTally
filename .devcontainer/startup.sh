#!/bin/bash

# TouchdownTally Development Environment Startup Script
# This script helps set up the development environment within the dev container

set -e

echo "🏈 Setting up TouchdownTally development environment..."

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not available. Make sure Docker-in-Docker feature is enabled."
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose is not available. Installing..."
    curl -L "https://github.com/docker/compose/releases/download/v2.23.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
fi

echo "✅ Docker and Docker Compose are available"

# Create .env file if it doesn't exist
if [ ! -f ".env" ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo "✅ Created .env file - please update with your API keys"
fi

# Start database services
echo "🚀 Starting database services..."
docker-compose up -d postgres redis adminer

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
    if docker-compose exec -T postgres pg_isready -U touchdown_user -d touchdown_tally; then
        echo "✅ PostgreSQL is ready"
        break
    fi
    sleep 2
done

echo "🎉 Development environment is ready!"
echo ""
echo "Available services:"
echo "  - PostgreSQL: localhost:5432"
echo "  - Redis: localhost:6379"
echo "  - Adminer (DB Admin): http://localhost:8081"
echo ""
echo "Next steps:"
echo "  1. Run 'make setup' to initialize the project"
echo "  2. Run 'make dev' to start development servers"
echo ""

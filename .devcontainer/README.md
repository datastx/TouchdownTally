# TouchdownTally Dev Container Setup

This directory contains the VS Code dev container configuration for TouchdownTally, providing a complete development environment with Docker-in-Docker support for running all services within the container.

## What's Included

### Development Tools
- **Go 1.21** with all essential tools (gopls, golangci-lint, delve debugger)
- **Node.js 18** with npm and Vue CLI (via dev container features)
- **Docker-in-Docker** for running PostgreSQL, Redis, and other services
- **Python 3** with useful packages for data analysis

### VS Code Extensions
- Go language support with debugging
- Vue.js development tools
- PostgreSQL database tools
- REST client for API testing
- Prettier for code formatting
- GitHub Copilot (if available)

### Services (via Docker Compose within container)
- **PostgreSQL**: Available on port 5432
- **Redis**: Available on port 6379  
- **Adminer**: Database admin UI on port 8081
- **Backend**: Go server on port 8080
- **Frontend**: Vue.js dev server on port 3000

## Quick Start

1. **Open in Container**
   - Install the "Remote - Containers" extension in VS Code
   - Open this project in VS Code
   - Click "Reopen in Container" when prompted (or use Command Palette: "Remote-Containers: Reopen in Container")

2. **Initial Setup**
   ```bash
   make quickstart
   ```
   This will:
   - Set up the Go backend with dependencies
   - Create the Vue.js frontend
   - Start database services using Docker Compose within the container
   - Run initial migrations
   - Seed with sample data

3. **Start Development**
   ```bash
   make dev
   ```
   This starts both backend and frontend development servers.

## Docker-in-Docker Architecture

The dev container now uses Docker-in-Docker, which means:
- The dev container itself runs as a Docker container
- Inside this container, Docker is available to run additional services
- PostgreSQL, Redis, and Adminer run as separate containers within the dev container
- All services are isolated and can be managed independently
- No need for complex docker-compose orchestration at the host level

## Common Commands

### Development
```bash
make dev              # Start both backend and frontend
make dev-backend      # Start only Go backend
make dev-frontend     # Start only Vue.js frontend
```

### Database
```bash
make db-up           # Start database services
make db-down         # Stop database services
make db-reset        # Reset database (destroys all data)
make db-migrate      # Run database migrations
make db-console      # Connect to PostgreSQL console
```

### Testing & Quality
```bash
make test            # Run all tests
make test-coverage   # Run tests with coverage
make lint            # Run all linters
make format          # Format all code
```

### Building
```bash
make build           # Build for production
make build-backend   # Build Go binary
make build-frontend  # Build Vue.js app
```

## Database Access

### Via Adminer (Web Interface)
- URL: http://localhost:8081
- Server: `postgres`
- Username: `touchdown_user`
- Password: `touchdown_dev`
- Database: `touchdown_tally`

### Via Command Line
```bash
make db-console
```

### Via VS Code
Use the SQLTools extension (included) to connect:
- Host: `localhost`
- Port: `5432`
- Database: `touchdown_tally`
- Username: `touchdown_user`
- Password: `touchdown_dev`

## Environment Variables

Copy `.env.example` to `.env` and update as needed:
```bash
cp .env.example .env
```

Key variables:
- `MYSPORTSFEEDS_API_KEY`: Your MySportsFeeds API key
- `JWT_SECRET`: JWT signing secret (change in production)
- `DATABASE_URL`: PostgreSQL connection string

## Project Structure

The dev container expects this structure:
```
TouchdownTally/
├── backend/           # Go backend code
│   ├── cmd/          # Command line tools
│   ├── internal/     # Internal packages
│   └── go.mod        # Go module file
├── frontend/         # Vue.js frontend code
├── .devcontainer/    # Dev container configuration
├── .vscode/          # VS Code settings
└── Makefile          # Build and development commands
```

## Debugging

### Go Backend
- Use VS Code's built-in Go debugging
- Pre-configured launch configurations available
- Set breakpoints and debug normally

### Vue.js Frontend
- Use browser dev tools
- Vue.js devtools extension recommended
- Hot reload enabled for rapid development

## Tips

1. **Performance**: The dev container mounts your source code as a volume for fast file access
2. **Extensions**: Additional VS Code extensions can be added to `.devcontainer/devcontainer.json`
3. **Database**: Database data persists between container restarts via Docker volumes
4. **Ports**: All necessary ports are automatically forwarded to your host machine
5. **Shell**: The container uses bash by default with useful aliases

## Troubleshooting

### Container Won't Start
- Ensure Docker is running
- Check for port conflicts on 3000, 5432, 6379, 8080, 8081
- Try rebuilding: Command Palette → "Remote-Containers: Rebuild Container"

### Database Connection Issues
- Ensure database is running: `make db-up`
- Check connection in Adminer: http://localhost:8081
- Verify environment variables in `.env`

### Go Module Issues
- Run `make setup-backend` to reinitialize Go modules
- Clear module cache: `go clean -modcache`

### Frontend Issues
- Delete `node_modules` and run `make setup-frontend`
- Check Node.js version: `node --version`

## Getting Help

- Run `make help` for available commands
- Check service health: `make health`
- View environment info: `make info`
- Check logs: `make logs`

For more detailed information, see the main project documentation.

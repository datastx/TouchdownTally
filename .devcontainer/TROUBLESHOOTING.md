# Dev Container Troubleshooting Guide

## Common Issues and Solutions

### 1. Container Build Failures

**Issue**: The dev container fails to build or shows errors during the build process.

**Common Causes**:
- Docker daemon not running
- Insufficient disk space
- Network connectivity issues during package installation
- Conflicting Docker configurations

**Solutions**:

#### Option A: Use the Simplified Configuration
Replace your current `devcontainer.json` with the simplified version:

```bash
cp .devcontainer/devcontainer-simple.json .devcontainer/devcontainer.json
```

#### Option B: Manual Troubleshooting Steps

1. **Check Docker Desktop is running**:
   ```bash
   docker --version
   docker info
   ```

2. **Clear Docker cache**:
   ```bash
   docker system prune -a
   ```

3. **Rebuild without cache**:
   - Open Command Palette (Cmd+Shift+P)
   - Run "Remote-Containers: Rebuild Container Without Cache"

4. **Check disk space**:
   ```bash
   df -h
   docker system df
   ```

### 2. Feature Installation Issues

**Issue**: Docker-in-Docker or Node.js features fail to install.

**Solutions**:

1. **Use base container approach**:
   ```json
   {
     "image": "mcr.microsoft.com/vscode/devcontainers/go:1.21",
     "features": {
       "ghcr.io/devcontainers/features/docker-in-docker:2": {}
     }
   }
   ```

2. **Install tools manually in Dockerfile**:
   - Remove problematic features from devcontainer.json
   - Add installation commands to Dockerfile

### 3. Port Forwarding Issues

**Issue**: Services not accessible on forwarded ports.

**Solutions**:
1. Check if ports are already in use on host
2. Restart VS Code
3. Use different ports

### 4. Permission Issues

**Issue**: Permission denied errors when running commands.

**Solutions**:
1. Check if running as correct user (`vscode`)
2. Add user to docker group:
   ```dockerfile
   RUN usermod -aG docker vscode
   ```

## Alternative: Minimal Setup

If you continue having issues, try this minimal approach:

### Step 1: Use Minimal devcontainer.json
```json
{
  "name": "TouchdownTally",
  "image": "mcr.microsoft.com/vscode/devcontainers/go:1.21",
  "features": {
    "ghcr.io/devcontainers/features/docker-in-docker:2": {},
    "ghcr.io/devcontainers/features/node:1": {"version": "18"}
  },
  "forwardPorts": [8080, 3000, 5432],
  "remoteUser": "vscode"
}
```

### Step 2: Install Tools After Container Creation
```bash
# Install additional tools manually
sudo apt-get update
sudo apt-get install postgresql-client redis-tools

# Install Go tools
go install golang.org/x/tools/gopls@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Debugging Container Issues

### Check Container Logs
```bash
# In VS Code terminal
docker logs $(docker ps -q)

# Check dev container logs
code --log debug
```

### Test Docker-in-Docker
```bash
# Inside the dev container
docker --version
docker-compose --version
docker run hello-world
```

### Common Error Messages and Solutions

1. **"Failed to connect to Docker daemon"**
   - Ensure Docker-in-Docker feature is properly installed
   - Restart the container

2. **"Permission denied" accessing docker.sock**
   - Check user permissions
   - Ensure user is in docker group

3. **"No space left on device"**
   - Clean up Docker: `docker system prune -a`
   - Check disk space: `df -h`

4. **Network timeout during build**
   - Check internet connection
   - Try building again (some downloads may be temporary failures)

## Quick Recovery Commands

```bash
# Reset everything
make clean
docker system prune -a

# Rebuild container
# Command Palette: "Remote-Containers: Rebuild Container"

# Start fresh with database
make db-reset
make quickstart
```

If none of these solutions work, please share the specific error messages from the VS Code logs, and I can provide more targeted assistance.

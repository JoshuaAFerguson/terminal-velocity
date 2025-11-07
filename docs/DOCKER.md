# Docker Deployment Guide

This guide covers deploying Terminal Velocity using Docker and Docker Compose.

## Quick Start

### Prerequisites

- Docker 20.10+ ([Install Docker](https://docs.docker.com/get-docker/))
- Docker Compose 2.0+ ([Install Docker Compose](https://docs.docker.com/compose/install/))

### 1. Clone the Repository

```bash
git clone https://github.com/JoshuaAFerguson/terminal-velocity.git
cd terminal-velocity
```

### 2. Configure Environment

```bash
# Copy example environment file
cp .env.example .env

# Edit with your settings
nano .env
```

**Important**: Change the `DB_PASSWORD` in `.env` to a secure password!

### 3. Start the Stack

```bash
# Start server and database
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f
```

### 4. Connect to the Game

```bash
ssh -p 2222 username@localhost
```

That's it! The server is running with PostgreSQL.

---

## Architecture

The Docker stack consists of:

```
┌─────────────────────────────────────┐
│   Terminal Velocity Server          │
│   - Go application                   │
│   - SSH server on port 2222          │
│   - Game engine                      │
└─────────────┬───────────────────────┘
              │
              │ PostgreSQL protocol
              │
┌─────────────▼───────────────────────┐
│   PostgreSQL Database                │
│   - User data                        │
│   - Universe state                   │
│   - Game data                        │
└──────────────────────────────────────┘
```

### Services

**1. `postgres`** - PostgreSQL 16 Database
- Container: `terminal-velocity-db`
- Port: `5432` (mapped to host)
- Volume: `postgres_data` (persistent)
- Auto-initializes with schema

**2. `server`** - Terminal Velocity Game Server
- Container: `terminal-velocity-server`
- Port: `2222` (SSH)
- Volumes: `server_logs`, `server_data`
- Depends on: `postgres`

**3. `pgadmin`** - Database Admin UI (Optional)
- Container: `terminal-velocity-pgadmin`
- Port: `5050` (HTTP)
- Profile: `tools` (not started by default)

---

## Configuration

### Environment Variables

Create `.env` file from `.env.example`:

```bash
# Database
DB_PASSWORD=your_secure_password

# Server
SSH_PORT=2222
MAX_PLAYERS=100

# Game
UNIVERSE_SEED=0        # 0 = random, or set seed for reproducible universe
NUM_SYSTEMS=100        # Number of star systems
```

### Custom Configuration

Edit `configs/config.yaml` for advanced settings:

```yaml
server:
  port: 2222
  max_players: 100

database:
  url: "postgres://terminal_velocity:password@postgres:5432/terminal_velocity"

game:
  num_systems: 100
  starting_credits: 10000
```

**Note**: When running in Docker, use `postgres` as the hostname (not `localhost`).

---

## Usage

### Start Services

```bash
# Start in background
docker-compose up -d

# Start with logs visible
docker-compose up

# Start specific service
docker-compose up -d postgres
```

### Stop Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (WARNING: deletes all data!)
docker-compose down -v
```

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f server
docker-compose logs -f postgres

# Last 100 lines
docker-compose logs --tail=100 server
```

### Service Status

```bash
# Check status
docker-compose ps

# Check health
docker-compose ps server
```

### Restart Services

```bash
# Restart all
docker-compose restart

# Restart specific service
docker-compose restart server
```

---

## Database Management

### PgAdmin (Web UI)

Start PgAdmin for database management:

```bash
# Start with pgadmin
docker-compose --profile tools up -d

# Access at http://localhost:5050
# Login with credentials from .env
```

**Add Server in PgAdmin**:
- Host: `postgres`
- Port: `5432`
- Database: `terminal_velocity`
- Username: `terminal_velocity`
- Password: (from `.env`)

### Direct Database Access

```bash
# Connect to database
docker-compose exec postgres psql -U terminal_velocity -d terminal_velocity

# Run SQL file
docker-compose exec -T postgres psql -U terminal_velocity -d terminal_velocity < backup.sql
```

### Backup Database

```bash
# Create backup
docker-compose exec -T postgres pg_dump -U terminal_velocity terminal_velocity > backup.sql

# Restore backup
docker-compose exec -T postgres psql -U terminal_velocity terminal_velocity < backup.sql
```

---

## Volumes and Data Persistence

### Persistent Volumes

Data is stored in Docker volumes:

```bash
# List volumes
docker volume ls | grep terminal-velocity

# Inspect volume
docker volume inspect terminal-velocity_postgres_data

# Backup volume
docker run --rm -v terminal-velocity_postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres_backup.tar.gz /data
```

### Volume Locations

- `postgres_data` - Database files
- `server_logs` - Application logs
- `server_data` - Game data
- `pgadmin_data` - PgAdmin settings

### Clean Start

```bash
# Stop services
docker-compose down

# Remove volumes (WARNING: deletes all data!)
docker volume rm terminal-velocity_postgres_data
docker volume rm terminal-velocity_server_logs
docker volume rm terminal-velocity_server_data

# Start fresh
docker-compose up -d
```

---

## Building Custom Images

### Build Server Image

```bash
# Build image
docker-compose build server

# Build with version info
VERSION=0.1.0 COMMIT=$(git rev-parse --short HEAD) docker-compose build server

# Build without cache
docker-compose build --no-cache server
```

### Multi-Platform Builds

```bash
# Build for multiple architectures
docker buildx build --platform linux/amd64,linux/arm64 -t terminal-velocity:latest .
```

---

## Networking

### Container Network

Containers communicate via `terminal-velocity-net` bridge network.

```bash
# Inspect network
docker network inspect terminal-velocity_terminal-velocity-net

# See connected containers
docker network ls
```

### Port Mapping

Default ports:
- `2222` - SSH server (game access)
- `5432` - PostgreSQL (database)
- `5050` - PgAdmin (with --profile tools)

Change ports in `docker-compose.yml` or override:

```bash
# Use different SSH port
SSH_PORT=2223 docker-compose up -d
```

---

## Production Deployment

### Security Hardening

**1. Change Default Passwords**
```bash
# Generate strong password
openssl rand -base64 32 > .env.secret
```

**2. Use Secrets Management**
```yaml
# docker-compose.yml
secrets:
  db_password:
    file: .env.secret
```

**3. Enable SSL/TLS**
- Use reverse proxy (nginx, Caddy)
- Terminate SSL at proxy
- Forward to SSH port

**4. Firewall Rules**
```bash
# Only expose SSH port
ufw allow 2222/tcp
ufw deny 5432/tcp  # Don't expose database
```

### Resource Limits

Add to `docker-compose.yml`:

```yaml
services:
  server:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G
```

### Monitoring

**Health Checks**:
```bash
# Check container health
docker-compose ps

# Manual health check
docker exec terminal-velocity-server nc -z localhost 2222
```

**Logs**:
```bash
# Set up log rotation
docker-compose logs --tail=1000 server > logs/server.log
```

### Backups

**Automated Backup Script**:
```bash
#!/bin/bash
# backup.sh
DATE=$(date +%Y%m%d_%H%M%S)
docker-compose exec -T postgres pg_dump -U terminal_velocity terminal_velocity | gzip > backups/backup_$DATE.sql.gz
find backups/ -name "backup_*.sql.gz" -mtime +7 -delete  # Keep 7 days
```

**Cron Job**:
```bash
# Daily backup at 2 AM
0 2 * * * /path/to/backup.sh
```

---

## Troubleshooting

### Server Won't Start

```bash
# Check logs
docker-compose logs server

# Check if database is ready
docker-compose exec postgres pg_isready -U terminal_velocity

# Restart with fresh build
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

### Database Connection Failed

```bash
# Verify database is running
docker-compose ps postgres

# Check database logs
docker-compose logs postgres

# Test connection
docker-compose exec server nc -zv postgres 5432
```

### Permission Denied Errors

```bash
# Fix volume permissions
docker-compose exec server chown -R terminalvelocity:terminalvelocity /app
```

### Can't Connect via SSH

```bash
# Check if port is open
nc -zv localhost 2222

# Check server health
docker-compose exec server nc -z localhost 2222

# Check firewall
ufw status
```

### Out of Memory

```bash
# Check resource usage
docker stats terminal-velocity-server

# Increase memory limit in docker-compose.yml
```

---

## Development

### Hot Reload Setup

For development with automatic rebuild:

```bash
# Install air (hot reload for Go)
go install github.com/cosmtrek/air@latest

# Create air config (outside container)
air init

# Run with volume mount
docker-compose -f docker-compose.dev.yml up
```

### Debug Mode

```bash
# Run server with debug output
docker-compose exec server /app/terminal-velocity --debug

# Interactive shell
docker-compose exec server sh
```

### Testing

```bash
# Run tests in container
docker-compose exec server go test ./...

# Build and test
docker-compose build server
docker-compose run --rm server go test -v ./...
```

---

## Migration from Local Setup

### Export Local Data

```bash
# Export local database
pg_dump -U postgres terminal_velocity > export.sql
```

### Import to Docker

```bash
# Start docker containers
docker-compose up -d

# Import data
docker-compose exec -T postgres psql -U terminal_velocity terminal_velocity < export.sql
```

---

## Advanced Configuration

### Custom Dockerfile

Create `Dockerfile.custom`:

```dockerfile
FROM terminal-velocity:latest

# Add custom configuration
COPY my-config.yaml /app/configs/config.yaml

# Add plugins or modifications
RUN apk add --no-cache custom-package
```

Build:
```bash
docker build -f Dockerfile.custom -t terminal-velocity:custom .
```

### Multiple Environments

```bash
# Production
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Development
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d
```

---

## Scaling

### Multiple Server Instances

```bash
# Scale server instances
docker-compose up -d --scale server=3
```

**Note**: Requires load balancer and shared session storage.

### Database Replication

Set up PostgreSQL streaming replication for high availability.

---

## CI/CD Integration

### GitHub Actions

```yaml
# .github/workflows/docker.yml
- name: Build Docker image
  run: docker-compose build

- name: Push to registry
  run: docker-compose push
```

### Registry

```bash
# Tag and push
docker tag terminal-velocity:latest ghcr.io/joshuaaferguson/terminal-velocity:latest
docker push ghcr.io/joshuaaferguson/terminal-velocity:latest
```

---

## Performance Tuning

### PostgreSQL

```yaml
# docker-compose.yml
postgres:
  command:
    - "postgres"
    - "-c"
    - "max_connections=200"
    - "-c"
    - "shared_buffers=256MB"
    - "-c"
    - "effective_cache_size=1GB"
```

### Server

```yaml
server:
  environment:
    GOMAXPROCS: 4
    GOMEMLIMIT: 2GiB
```

---

## Support

- **Issues**: https://github.com/JoshuaAFerguson/terminal-velocity/issues
- **Documentation**: https://github.com/JoshuaAFerguson/terminal-velocity/docs
- **Email**: contact@joshua-ferguson.com

---

**Last Updated**: 2025-11-06

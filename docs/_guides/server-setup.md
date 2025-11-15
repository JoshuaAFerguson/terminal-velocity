---
layout: guide
title: Server Setup
description: Complete guide to installing, configuring, and running your own Terminal Velocity server
---

# Server Setup Guide

This comprehensive guide will walk you through setting up your own Terminal Velocity server, from prerequisites to production deployment.

---

## Prerequisites

### System Requirements

**Minimum**:
- CPU: 2 cores
- RAM: 2GB
- Storage: 10GB
- OS: Linux (Ubuntu 20.04+, Debian 11+, CentOS 8+) or macOS

**Recommended** (for 50+ concurrent players):
- CPU: 4 cores
- RAM: 4GB
- Storage: 20GB SSD
- OS: Linux (Ubuntu 22.04 LTS)

**Network**:
- Open port 2222 for SSH connections
- Open port 8080 for metrics (optional, internal network only)
- Static IP or domain name recommended

### Required Software

**1. Go 1.24 or later**

Check if installed:
```bash
go version
```

Install Go:
```bash
# Ubuntu/Debian
wget https://go.dev/dl/go1.24.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# macOS (with Homebrew)
brew install go
```

**2. PostgreSQL 12 or later**

Check if installed:
```bash
psql --version
```

Install PostgreSQL:
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install postgresql postgresql-contrib

# macOS (with Homebrew)
brew install postgresql@15
brew services start postgresql@15
```

**3. Git**

```bash
# Ubuntu/Debian
sudo apt install git

# macOS
brew install git
```

**4. Make** (usually pre-installed)

```bash
# Ubuntu/Debian
sudo apt install build-essential

# macOS
xcode-select --install
```

### Optional Software

**Docker & Docker Compose** (for containerized deployment):
```bash
# Ubuntu/Debian
sudo apt install docker.io docker-compose
sudo systemctl enable docker
sudo systemctl start docker
sudo usermod -aG docker $USER  # Re-login after this

# macOS
brew install docker docker-compose
```

**entr** (for auto-rebuild during development):
```bash
# Ubuntu/Debian
sudo apt install entr

# macOS
brew install entr
```

---

## Quick Start (Recommended)

The fastest way to get a server running using the initialization script.

### 1. Clone Repository

```bash
git clone https://github.com/JoshuaAFerguson/terminal-velocity.git
cd terminal-velocity
```

### 2. Build Tools

```bash
make build-tools
```

This builds:
- `genmap` - Universe generation tool
- `accounts` - Account management tool

### 3. Run Initialization Script

```bash
./scripts/init-server.sh
```

The script will:
1. Check for required dependencies
2. Create PostgreSQL database and user
3. Initialize database schema
4. Generate and populate universe (100 systems)
5. Display connection instructions

**Follow the prompts** and provide:
- Database password (you'll need this later)
- PostgreSQL superuser password (usually your system password)

### 4. Create Admin Account

```bash
./accounts create admin admin@example.com
```

You'll be prompted to set a password for this account.

### 5. Start the Server

```bash
make run
```

The server will start on port 2222.

### 6. Connect and Play!

```bash
ssh -p 2222 admin@localhost
```

**Success!** You now have a running Terminal Velocity server.

---

## Manual Setup

For more control over the setup process or troubleshooting.

### 1. Clone and Build

```bash
# Clone repository
git clone https://github.com/JoshuaAFerguson/terminal-velocity.git
cd terminal-velocity

# Download dependencies
go mod download

# Build server
make build

# Build tools
make build-tools
```

### 2. Database Setup

#### Create Database and User

```bash
# Switch to postgres user
sudo -u postgres psql

# In psql:
CREATE DATABASE terminal_velocity;
CREATE USER terminal_velocity WITH PASSWORD 'your_secure_password_here';
GRANT ALL PRIVILEGES ON DATABASE terminal_velocity TO terminal_velocity;
\q
```

#### Initialize Schema

```bash
psql -U terminal_velocity -d terminal_velocity -f scripts/schema.sql
```

You'll be prompted for the password you set above.

#### Verify Database

```bash
psql -U terminal_velocity -d terminal_velocity -c "\dt"
```

You should see a list of tables (players, star_systems, etc.).

### 3. Generate Universe

Create the game universe (star systems, planets, jump routes):

```bash
./genmap -systems 100 -save \
  -db-host localhost \
  -db-port 5432 \
  -db-user terminal_velocity \
  -db-password your_secure_password_here \
  -db-name terminal_velocity
```

**Parameters**:
- `-systems 100`: Generate 100 star systems
- `-save`: Save to database (otherwise just preview)
- `-db-*`: Database connection parameters

**Options**:
- `-systems N`: Number of systems (default: 100, recommended: 50-200)
- `-stats`: Show statistics about generated universe
- `-preview`: Display ASCII preview of galaxy

**Example with preview**:
```bash
./genmap -systems 50 -stats -preview
```

### 4. Configure Server

Copy example configuration:
```bash
cp configs/config.example.yaml configs/config.yaml
```

Edit `configs/config.yaml`:

```yaml
server:
  host: "0.0.0.0"          # Listen on all interfaces
  port: 2222                # SSH port
  host_key_path: "configs/ssh_host_key"  # SSH host key

database:
  host: "localhost"
  port: 5432
  user: "terminal_velocity"
  password: "your_secure_password_here"
  database: "terminal_velocity"
  max_connections: 25

game:
  allow_registration: true   # Allow new player registration
  starting_credits: 10000    # Starting credits for new players
  starting_system: "Sol"     # Starting system name
  auto_save_interval: 30     # Auto-save interval in seconds

metrics:
  enabled: true
  port: 8080                 # Metrics HTTP server port

security:
  rate_limit_enabled: true
  max_connections_per_ip: 5
  max_auth_attempts: 5
  auth_lockout_time: 900     # 15 minutes in seconds
  autoban_threshold: 20
  autoban_duration: 86400    # 24 hours in seconds
```

**Important**: Change the database password to match what you set earlier!

### 5. Generate SSH Host Key

```bash
ssh-keygen -t ed25519 -f configs/ssh_host_key -N ""
```

This creates a persistent SSH host key to prevent "host key changed" warnings.

### 6. Create User Accounts

Create your first user account (admin):

```bash
./accounts create admin admin@example.com
```

Create additional player accounts:

```bash
./accounts create player1 player1@example.com
./accounts create player2 player2@example.com
```

**Add SSH public keys** (optional, for passwordless login):

```bash
./accounts add-key admin ~/.ssh/id_ed25519.pub
```

### 7. Start the Server

```bash
./server -config configs/config.yaml
```

Or use Make:
```bash
make run
```

**Server is running!** You should see:
```
Terminal Velocity SSH Server Starting...
Listening on 0.0.0.0:2222
Metrics server running on :8080
```

### 8. Test Connection

From the same machine:
```bash
ssh -p 2222 admin@localhost
```

From another machine:
```bash
ssh -p 2222 admin@your-server-ip
```

---

## Docker Setup

For containerized deployment with Docker Compose.

### 1. Clone Repository

```bash
git clone https://github.com/JoshuaAFerguson/terminal-velocity.git
cd terminal-velocity
```

### 2. Configure Environment

Create `.env` file:
```bash
cp .env.example .env
```

Edit `.env`:
```
DB_PASSWORD=your_secure_password_here
DB_USER=terminal_velocity
DB_NAME=terminal_velocity
DB_HOST=postgres
DB_PORT=5432

SSH_PORT=2222
METRICS_PORT=8080
```

### 3. Start Stack

```bash
docker compose up -d
```

This starts:
- PostgreSQL database
- Terminal Velocity server

**Check logs**:
```bash
docker compose logs -f
```

### 4. Initialize Database

```bash
# Run schema initialization
docker compose exec server sh -c "psql -h postgres -U terminal_velocity -d terminal_velocity -f scripts/schema.sql"

# Generate universe (inside container)
docker compose exec server ./genmap -systems 100 -save \
  -db-host postgres \
  -db-user terminal_velocity \
  -db-password your_secure_password_here \
  -db-name terminal_velocity
```

### 5. Create Accounts

```bash
docker compose exec server ./accounts create admin admin@example.com
```

### 6. Connect

```bash
ssh -p 2222 admin@localhost
```

### Docker Management Commands

```bash
# Start stack
docker compose up -d

# Stop stack
docker compose down

# View logs
docker compose logs -f server

# Restart services
docker compose restart

# Remove everything (including volumes)
docker compose down -v

# Update and rebuild
git pull
docker compose build
docker compose up -d
```

---

## Server Administration

### Account Management

**Create account**:
```bash
./accounts create <username> <email>
```

**Add SSH key**:
```bash
./accounts add-key <username> <path-to-public-key>
```

**Delete account**:
```bash
# From PostgreSQL
psql -U terminal_velocity -d terminal_velocity -c \
  "DELETE FROM players WHERE username = 'badplayer';"
```

**Reset password** (future feature):
```bash
./accounts reset-password <username>
```

### Database Management

**Backup database**:
```bash
./scripts/backup.sh
```

**Backup with custom options**:
```bash
./scripts/backup.sh -d /var/backups/terminal-velocity -r 30 -c 50
```
- `-d`: Backup directory
- `-r`: Retention in days
- `-c`: Maximum backup count

**List backups**:
```bash
./scripts/restore.sh --list
```

**Restore from backup**:
```bash
./scripts/restore.sh /path/to/backup.sql.gz
```

**Automated backups with cron**:
```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * /path/to/terminal-velocity/scripts/backup.sh
```

### Monitoring

**Metrics endpoints**:
- `http://localhost:8080/metrics` - Prometheus format
- `http://localhost:8080/stats` - HTML dashboard
- `http://localhost:8080/stats/enhanced` - Enhanced metrics
- `http://localhost:8080/health` - Health check

**Check server health**:
```bash
curl http://localhost:8080/health
```

**View stats in browser**:
```
http://your-server-ip:8080/stats
```

**Prometheus integration**:
Add to `prometheus.yml`:
```yaml
scrape_configs:
  - job_name: 'terminal-velocity'
    static_configs:
      - targets: ['localhost:8080']
```

### Security

**Rate limiting** is enabled by default:
- 5 concurrent connections per IP
- 20 connections per minute per IP
- 5 authentication attempts before 15-minute lockout
- 20 failed attempts = 24-hour automatic ban

**View banned IPs** (in database):
```sql
SELECT * FROM ip_bans WHERE expires_at > NOW();
```

**Manually ban IP**:
```sql
INSERT INTO ip_bans (ip_address, reason, banned_at, expires_at)
VALUES ('1.2.3.4', 'Abuse', NOW(), NOW() + INTERVAL '24 hours');
```

**Unban IP**:
```sql
DELETE FROM ip_bans WHERE ip_address = '1.2.3.4';
```

### Server Settings

**Edit game settings** via database:
```sql
-- Disable registration
UPDATE server_settings SET value = 'false' WHERE key = 'allow_registration';

-- Change starting credits
UPDATE server_settings SET value = '20000' WHERE key = 'starting_credits';
```

Or edit `configs/config.yaml` and restart server.

---

## Production Deployment

### System Setup

**1. Create dedicated user**:
```bash
sudo useradd -r -m -s /bin/bash terminal-velocity
sudo -u terminal-velocity -i
```

**2. Clone and build** (as terminal-velocity user):
```bash
cd ~
git clone https://github.com/JoshuaAFerguson/terminal-velocity.git
cd terminal-velocity
make build
make build-tools
```

**3. Set up database** (as postgres user):
```bash
sudo -u postgres createuser terminal-velocity
sudo -u postgres createdb -O terminal-velocity terminal_velocity
```

**4. Initialize**:
```bash
./scripts/init-server.sh
```

### Systemd Service

Create `/etc/systemd/system/terminal-velocity.service`:

```ini
[Unit]
Description=Terminal Velocity SSH Game Server
After=network.target postgresql.service
Requires=postgresql.service

[Service]
Type=simple
User=terminal-velocity
Group=terminal-velocity
WorkingDirectory=/home/terminal-velocity/terminal-velocity
ExecStart=/home/terminal-velocity/terminal-velocity/server -config /home/terminal-velocity/terminal-velocity/configs/config.yaml
Restart=always
RestartSec=10

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/home/terminal-velocity/terminal-velocity/data
ReadWritePaths=/home/terminal-velocity/terminal-velocity/logs

[Install]
WantedBy=multi-user.target
```

**Enable and start**:
```bash
sudo systemctl daemon-reload
sudo systemctl enable terminal-velocity
sudo systemctl start terminal-velocity
```

**Check status**:
```bash
sudo systemctl status terminal-velocity
```

**View logs**:
```bash
sudo journalctl -u terminal-velocity -f
```

### Firewall Configuration

**UFW** (Ubuntu):
```bash
sudo ufw allow 2222/tcp comment 'Terminal Velocity SSH'
sudo ufw enable
```

**firewalld** (CentOS/RHEL):
```bash
sudo firewall-cmd --permanent --add-port=2222/tcp
sudo firewall-cmd --reload
```

**iptables**:
```bash
sudo iptables -A INPUT -p tcp --dport 2222 -j ACCEPT
sudo iptables-save > /etc/iptables/rules.v4
```

### Reverse Proxy (Optional)

If you want to serve metrics dashboard via HTTPS:

**Nginx**:
```nginx
server {
    listen 80;
    server_name stats.yourdomain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;

        # Restrict access
        allow 10.0.0.0/8;   # Your internal network
        deny all;
    }
}
```

### Automated Backups

**Cron job** (as terminal-velocity user):
```bash
crontab -e
```

Add:
```
# Daily backup at 2 AM
0 2 * * * /home/terminal-velocity/terminal-velocity/scripts/backup.sh -d /var/backups/terminal-velocity -r 30

# Weekly cleanup of old logs
0 3 * * 0 find /home/terminal-velocity/terminal-velocity/logs -name "*.log" -mtime +30 -delete
```

### Monitoring Setup

**Health check script** (`/usr/local/bin/check-tv-health.sh`):
```bash
#!/bin/bash
HEALTH_URL="http://localhost:8080/health"
STATUS=$(curl -s "$HEALTH_URL" | jq -r '.status')

if [ "$STATUS" != "healthy" ]; then
    echo "Terminal Velocity unhealthy: $STATUS"
    # Send alert (email, Slack, PagerDuty, etc.)
    systemctl restart terminal-velocity
fi
```

**Cron** (check every 5 minutes):
```
*/5 * * * * /usr/local/bin/check-tv-health.sh
```

---

## Troubleshooting

### Common Issues

**Server won't start**:
```bash
# Check logs
./server -config configs/config.yaml

# Common causes:
# 1. Port 2222 already in use
sudo lsof -i :2222

# 2. Database connection failed
psql -U terminal_velocity -d terminal_velocity -c "SELECT 1;"

# 3. Missing SSH host key
ssh-keygen -t ed25519 -f configs/ssh_host_key -N ""
```

**Can't connect via SSH**:
```bash
# Verify server is listening
netstat -tlnp | grep 2222

# Test connection with verbose output
ssh -vvv -p 2222 username@localhost

# Check firewall
sudo ufw status
sudo iptables -L -n
```

**Database errors**:
```bash
# Check PostgreSQL is running
sudo systemctl status postgresql

# Check connection
psql -U terminal_velocity -d terminal_velocity

# Reinitialize schema
psql -U terminal_velocity -d terminal_velocity -f scripts/schema.sql
```

**Universe not generated**:
```bash
# Check if systems exist
psql -U terminal_velocity -d terminal_velocity -c \
  "SELECT COUNT(*) FROM star_systems;"

# Regenerate universe
./genmap -systems 100 -save \
  -db-host localhost \
  -db-user terminal_velocity \
  -db-password your_password \
  -db-name terminal_velocity
```

**Performance issues**:
```bash
# Check metrics
curl http://localhost:8080/stats/enhanced

# Database performance
psql -U terminal_velocity -d terminal_velocity -c \
  "SELECT schemaname, tablename, idx_scan, seq_scan FROM pg_stat_user_tables;"

# Add indexes (already in schema.sql, but verify)
# See scripts/schema.sql for index definitions
```

### Error Messages

**"connection refused"**:
- Server not running
- Wrong port
- Firewall blocking

**"permission denied"**:
- Wrong username/password
- Account not created
- SSH key mismatch

**"too many connections"**:
- Database connection pool exhausted
- Increase `max_connections` in config.yaml

**"database locked"**:
- PostgreSQL, not SQLite - shouldn't happen
- Check for long-running queries

---

## Upgrading

### Upgrading the Server

```bash
# Stop server
sudo systemctl stop terminal-velocity
# or Ctrl+C if running in foreground

# Backup database
./scripts/backup.sh

# Pull latest code
git pull origin main

# Update dependencies
go mod download

# Rebuild
make build
make build-tools

# Run migrations (if any)
./scripts/migrate.sh up

# Restart server
sudo systemctl start terminal-velocity
```

### Database Migrations

Check migration status:
```bash
./scripts/migrate.sh status
```

Apply pending migrations:
```bash
./scripts/migrate.sh up
```

Rollback last migration:
```bash
./scripts/migrate.sh down
```

### Backup Before Upgrading

**Always backup before upgrading!**
```bash
./scripts/backup.sh -d /var/backups/pre-upgrade-$(date +%Y%m%d)
```

---

## Performance Tuning

### PostgreSQL Tuning

Edit `/etc/postgresql/15/main/postgresql.conf`:

```ini
# Memory (for 4GB RAM server)
shared_buffers = 1GB
effective_cache_size = 3GB
maintenance_work_mem = 256MB
work_mem = 16MB

# Connections
max_connections = 100

# Performance
random_page_cost = 1.1  # For SSD
effective_io_concurrency = 200

# Write-Ahead Log
wal_buffers = 16MB
checkpoint_completion_target = 0.9
```

Restart PostgreSQL:
```bash
sudo systemctl restart postgresql
```

### Server Tuning

Edit `configs/config.yaml`:

```yaml
database:
  max_connections: 25        # Adjust based on player count
  connection_timeout: 10     # Seconds
  max_idle_connections: 5
  max_connection_lifetime: 3600

game:
  auto_save_interval: 30     # Lower = more saves, more DB load
  max_active_missions: 5
  event_check_interval: 60   # Seconds between event checks

metrics:
  collection_interval: 10    # Seconds between metric updates
```

### System Limits

Edit `/etc/security/limits.conf`:

```
terminal-velocity soft nofile 65536
terminal-velocity hard nofile 65536
```

### Load Testing

Test with multiple concurrent connections:

```bash
# Install siege
sudo apt install siege

# Test connection handling
siege -c 50 -t 30S ssh://player@localhost:2222
```

Monitor performance:
```bash
# Real-time stats
watch -n 1 'curl -s http://localhost:8080/stats/enhanced'

# Database connections
watch -n 1 "psql -U terminal_velocity -d terminal_velocity -c \
  'SELECT COUNT(*) FROM pg_stat_activity;'"
```

---

## Security Best Practices

### System Hardening

1. **Use a firewall** (UFW, firewalld, or iptables)
2. **Disable password SSH for system** (use keys only for admin access)
3. **Keep system updated**: `sudo apt update && sudo apt upgrade`
4. **Use strong database passwords** (20+ characters, mixed case, symbols)
5. **Enable automatic security updates**

### Application Security

1. **Enable rate limiting** (default: enabled)
2. **Monitor audit logs** regularly
3. **Review banned IPs** periodically
4. **Use 2FA for admin accounts** (when implemented)
5. **Restrict metrics port** to internal network only

### Monitoring

1. **Set up health checks** (every 5 minutes)
2. **Configure alerts** for:
   - Server down
   - High error rate
   - Database connection issues
   - Disk space low
3. **Review metrics** daily
4. **Check logs** for suspicious activity

---

## Next Steps

Now that your server is running:

1. **[Create content]({{ site.baseurl }}/documentation)**:
   - Add custom quests
   - Create events
   - Design storylines

2. **[Customize settings]({{ site.baseurl }}/ADMIN_SYSTEM)**:
   - Adjust economy balance
   - Configure game difficulty
   - Set server rules

3. **[Monitor performance]({{ site.baseurl }}/METRICS_MONITORING)**:
   - Watch metrics dashboard
   - Optimize database queries
   - Tune server settings

4. **[Invite players]({{ site.baseurl }}/guides/getting-started)**:
   - Share connection details
   - Create welcome documentation
   - Build community

5. **[Contribute](https://github.com/JoshuaAFerguson/terminal-velocity)**:
   - Report bugs
   - Suggest features
   - Submit pull requests

---

## Support

**Documentation**:
- [Technical Docs]({{ site.baseurl }}/documentation)
- [CLAUDE.md](https://github.com/JoshuaAFerguson/terminal-velocity/blob/main/CLAUDE.md) - Complete reference

**Community**:
- [GitHub Discussions](https://github.com/JoshuaAFerguson/terminal-velocity/discussions)
- [Issues](https://github.com/JoshuaAFerguson/terminal-velocity/issues)

**Commercial Support** (coming soon):
- Managed hosting
- Custom development
- Priority support

---

## Conclusion

Congratulations! You now have a fully functional Terminal Velocity server. Whether you're running it for friends, a community, or the public, you're now part of the Terminal Velocity universe.

**Remember**:
- Backup regularly
- Monitor health metrics
- Keep software updated
- Engage with your community

**May your server thrive and your universe prosper!** 🚀

---

**Quick Reference**:
```bash
# Start server
make run
# or
sudo systemctl start terminal-velocity

# Stop server
Ctrl+C
# or
sudo systemctl stop terminal-velocity

# Backup database
./scripts/backup.sh

# Create account
./accounts create <username> <email>

# Check health
curl http://localhost:8080/health

# View stats
curl http://localhost:8080/stats
```

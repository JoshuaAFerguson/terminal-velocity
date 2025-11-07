---
name: Deployment/Configuration Issue
about: Report issues with deployment, setup, or configuration
title: '[DEPLOY] '
labels: deployment
assignees: ''
---

## Issue Type
- [ ] Docker deployment problem
- [ ] Database setup issue
- [ ] SSH configuration problem
- [ ] Environment configuration
- [ ] Build/compilation issue
- [ ] Dependency problem
- [ ] Performance/scaling issue
- [ ] Other (specify below)

## Description
Clear description of the deployment or configuration issue.

## Environment
- **OS**: [e.g., Ubuntu 22.04, macOS 14.0, Windows 11]
- **Deployment Method**: [e.g., Docker, Docker Compose, bare metal]
- **Go Version**: [e.g., 1.23.0]
- **PostgreSQL Version**: [e.g., 16]
- **Architecture**: [e.g., amd64, arm64]
- **Cloud Provider** (if applicable): [e.g., AWS, GCP, Azure, DigitalOcean]

## Steps Taken
What steps did you take that led to this issue?
1. Step 1
2. Step 2
3. Step 3

## Expected Result
What did you expect to happen?

## Actual Result
What actually happened?

## Error Messages/Logs
```
Paste relevant error messages, logs, or stack traces
```

## Configuration Files
Relevant configuration (sanitize any secrets!):

**docker-compose.yml** (if applicable):
```yaml
# paste relevant sections
```

**.env** (sanitized):
```
# paste relevant variables (remove actual secrets)
```

**config.yaml** (sanitized):
```yaml
# paste relevant sections
```

## System Resources
- **CPU**: [e.g., 4 cores]
- **RAM**: [e.g., 8GB]
- **Disk Space**: [e.g., 50GB available]

## Network Configuration
- **Ports**: [e.g., 2222 for SSH, 5432 for PostgreSQL]
- **Firewall**: [any relevant rules]
- **Reverse Proxy**: [if using nginx, caddy, etc.]

## Docker Information (if applicable)
```bash
# Output of: docker --version
# Output of: docker-compose --version
# Output of: docker ps
# Output of: docker logs <container>
```

## Build Information (if applicable)
```bash
# Output of: go version
# Output of: go env
# Output of: make build
```

## Have you tried?
- [ ] Checking the documentation
- [ ] Searching existing issues
- [ ] Fresh install/rebuild
- [ ] Clearing Docker volumes
- [ ] Checking firewall/port settings
- [ ] Reviewing logs

## Additional Context
Any other relevant information about the deployment environment or issue.

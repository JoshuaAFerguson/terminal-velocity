# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Currently supported versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

**Note**: Terminal Velocity is currently in early development (Phase 1). Security updates will be provided for the current development version.

## Reporting a Vulnerability

The Terminal Velocity team takes security bugs seriously. We appreciate your efforts to responsibly disclose your findings.

### How to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to:

**contact@joshua-ferguson.com**

Include the following information in your report:

- **Type of vulnerability** (e.g., SQL injection, authentication bypass, etc.)
- **Full paths of source file(s)** related to the vulnerability
- **Location of the affected source code** (tag/branch/commit or direct URL)
- **Step-by-step instructions to reproduce** the issue
- **Proof-of-concept or exploit code** (if possible)
- **Impact of the issue**, including how an attacker might exploit it

This information will help us triage your report more quickly.

### What to Expect

When you report a security issue, you can expect:

1. **Acknowledgment**: We will acknowledge receipt of your vulnerability report within **48 hours**.

2. **Communication**: We will send you regular updates about our progress addressing the issue.

3. **Timeline**: We aim to:
   - Confirm the problem and determine affected versions within **7 days**
   - Release a fix within **30 days** of confirmation
   - Credit you in the security advisory (if desired)

4. **Disclosure**: We follow a coordinated disclosure process:
   - We will work with you to understand and resolve the issue
   - We will not disclose the issue until a fix is available
   - We will credit you in our release notes (unless you prefer to remain anonymous)

### Preferred Languages

We prefer all communications to be in English.

## Security Best Practices for Deployments

If you're running a Terminal Velocity server, we recommend:

### SSH Server Security

- **Use strong host keys**: Generate new ED25519 host keys, don't use default keys
- **Limit connections**: Configure firewall rules to restrict SSH access
- **Monitor logs**: Regularly review SSH connection logs for suspicious activity
- **Rate limiting**: Implement connection rate limiting to prevent brute force attacks
- **Keep updated**: Always run the latest version with security patches

### Database Security

- **Use strong passwords**: Generate secure passwords for PostgreSQL users
- **Network isolation**: Run PostgreSQL on localhost or private network only
- **Regular backups**: Backup your database regularly and test restoration
- **Least privilege**: Grant minimal required permissions to database users
- **Update regularly**: Keep PostgreSQL updated with security patches

### Application Security

- **Environment variables**: Never commit sensitive credentials to Git
- **Input validation**: The game validates all user input, but review logs for anomalies
- **Session management**: Session tokens are cryptographically secure
- **Dependency updates**: Keep Go modules updated (use `go get -u ./...`)

### Server Hardening

```bash
# Example: Run server with limited permissions
useradd -r -s /bin/false terminalvelocity
chown terminalvelocity:terminalvelocity /opt/terminal-velocity
sudo -u terminalvelocity ./terminal-velocity

# Use systemd service for automatic restart and logging
systemctl enable terminal-velocity
systemctl start terminal-velocity
```

### Network Security

- **Use a firewall**: Only expose SSH port (default 2222)
- **Consider VPN**: For private servers, use a VPN for access control
- **DDoS protection**: Use services like Cloudflare or similar if running publicly
- **Monitor traffic**: Watch for unusual connection patterns

## Known Security Considerations

### Current Development Phase

Terminal Velocity is in **Phase 1** development. Some security features are not yet implemented:

- ⚠️ Password reset functionality (planned for Phase 2)
- ⚠️ Two-factor authentication (planned for future)
- ⚠️ Advanced rate limiting (basic implementation only)
- ⚠️ Audit logging (planned for Phase 2)
- ⚠️ Encrypted player data at rest (planned for future)

### Security Features Implemented

- ✅ Bcrypt password hashing
- ✅ Secure session token generation
- ✅ SQL injection prevention (parameterized queries)
- ✅ Input validation and sanitization
- ✅ Protection against path traversal
- ✅ Safe concurrency primitives (mutexes, channels)

### Future Security Enhancements

Planned security improvements:

1. **OAuth/SSO integration** - Phase 3+
2. **End-to-end encryption** for player communications - Phase 3+
3. **Advanced intrusion detection** - Phase 3+
4. **Automated security scanning** in CI/CD - In progress
5. **Penetration testing** - Before public release
6. **Bug bounty program** - After v1.0 release

## Security Updates

Security updates will be announced through:

1. **GitHub Security Advisories**: https://github.com/JoshuaAFerguson/terminal-velocity/security/advisories
2. **Release Notes**: Security fixes will be clearly marked in release notes
3. **Git Tags**: Security releases will be tagged with patch version bumps

Subscribe to repository notifications to stay informed about security updates.

## Vulnerability Disclosure Policy

We follow these principles:

- **Coordinated Disclosure**: We work with reporters to validate and fix issues before public disclosure
- **Transparency**: Once fixed, we publicly disclose vulnerabilities with credit to reporters
- **No Legal Action**: We will not pursue legal action against security researchers who follow this policy
- **Recognition**: We maintain a security hall of fame for researchers who help us

## Security Hall of Fame

Contributors who have responsibly disclosed security issues:

*No entries yet - be the first to help secure Terminal Velocity!*

## Scope

The following are **in scope** for security reports:

- SSH authentication bypass
- SQL injection vulnerabilities
- Remote code execution
- Privilege escalation
- Session hijacking
- Denial of service (critical only)
- Information disclosure
- Cross-site scripting (if web interface is added)

The following are **out of scope**:

- Social engineering
- Physical attacks
- Issues in third-party dependencies (report to the dependency maintainers)
- Denial of service through resource exhaustion (expected in multiplayer games)
- Issues requiring physical access to the server
- Reports from automated tools without validation

## Contact

For security concerns, contact:

**Email**: contact@joshua-ferguson.com
**PGP Key**: (To be added in future release)

For general questions, use GitHub Issues.

## Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Best Practices](https://github.com/golang/go/wiki/Security)
- [PostgreSQL Security](https://www.postgresql.org/docs/current/security.html)
- [SSH Hardening Guide](https://www.ssh.com/academy/ssh/server)

---

**Last Updated**: 2025-11-06
**Version**: 0.1.0 (Phase 1 Development)

Thank you for helping keep Terminal Velocity secure!

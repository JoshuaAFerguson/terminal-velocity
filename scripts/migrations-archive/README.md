# Migration Files Archive

**Date Archived:** 2025-11-15

These migration files have been **consolidated into** `scripts/schema.sql`.

Since there is no production database to migrate, all changes from these migration files have been applied directly to the main schema file.

## Archived Migrations

- `000_create_migrations_table.sql` - Schema migrations tracking
- `001_add_email_and_ssh_keys.sql` - Email and SSH key authentication
- `002_add_player_coordinates.sql` - Player X/Y coordinates
- `003_add_admin_tables.sql` - Admin users, bans, mutes, audit log
- `004_add_player_mail.sql` - Player mail system
- `005_add_shared_loadouts.sql` - Ship loadout sharing
- `010_performance_indexes.sql` - Performance optimization indexes
- `010_security_v2_tables.sql` - Security V2 tables (2FA, login history, etc.)
- `011_social_features.sql` - Friends, blocks, notifications
- `012_add_player_progression_fields.sql` - Player progression tracking
- `013_add_capture_mining_stats.sql` - Capture and mining statistics
- `014_add_weapon_ammo_tracking.sql` - Weapon ammo tracking
- `015_add_planet_coordinates.sql` - Planet coordinates
- `016_add_resources_mined_tracking.sql` - Resources mined tracking
- `017_add_crafting_skill_tracking.sql` - Crafting skill tracking
- `018_add_research_points.sql` - Research points system

## Changes Consolidated

All tables, columns, indexes, and constraints from these migrations are now in `scripts/schema.sql`:

### Tables Added
- `schema_migrations` - Migration tracking
- `player_friends` - Friend relationships
- `friend_requests` - Pending friend requests
- `player_blocks` - Blocked players
- `player_notifications` - Notification system
- `account_events` - Security event tracking
- `player_two_factor` - 2FA configuration
- `password_reset_tokens` - Password reset system
- `admin_ip_whitelist` - Admin IP restrictions
- `login_history` - Login history for anomaly detection
- `player_sessions` - Session management
- `honeypot_attempts` - Honeypot tracking
- `rate_limit_tracking` - Rate limiting
- `player_security_settings` - Security preferences
- `trusted_devices` - Device trust management

### Columns Added to `players` Table
- `crafting_skill_metalwork`, `crafting_skill_electronics`, `crafting_skill_weapons`, `crafting_skill_propulsion`
- `research_points`
- `bio`, `join_date`, `total_playtime`, `profile_privacy`

### Indexes Added
- Performance indexes for market, cargo, player location, etc.
- Social feature indexes
- Security feature indexes

## Future Migrations

If production deployment occurs in the future, new migrations should be created in `scripts/migrations/` directory.

For now, all schema changes should be made directly to `scripts/schema.sql`.

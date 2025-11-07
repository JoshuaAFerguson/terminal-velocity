## Description
Brief description of what this PR does and why.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Refactoring (no functional changes)
- [ ] Performance improvement
- [ ] Tests
- [ ] Database migration
- [ ] Configuration change

## Related Issue
Fixes #(issue number)
Related to #(issue number)

## Roadmap Phase
- [ ] Phase 1: Foundation & Navigation
- [ ] Phase 2: Core Economy
- [ ] Phase 3: Ship Progression
- [ ] Phase 4: Combat System
- [ ] Phase 5: Missions & Progression
- [ ] Phase 6: Multiplayer Features
- [ ] Phase 7: Polish & Content
- [ ] Phase 8: Advanced Features
- [ ] Not part of roadmap

## Changes Made
### Code Changes
- Change 1
- Change 2
- Change 3

### Database Changes
- [ ] Schema migrations included
- [ ] Data migrations needed
- [ ] No database changes

### API Changes
- [ ] New endpoints added
- [ ] Existing endpoints modified
- [ ] Breaking API changes
- [ ] No API changes

## Game Design Impact
How does this change affect gameplay?
- **Trading**: [no impact / describe impact]
- **Combat**: [no impact / describe impact]
- **Progression**: [no impact / describe impact]
- **Economy**: [no impact / describe impact]
- **Multiplayer**: [no impact / describe impact]
- **Balance**: [any balance concerns or adjustments]

## Testing
### Automated Tests
- [ ] Unit tests pass (`go test ./...`)
- [ ] Integration tests pass
- [ ] All existing tests still pass
- [ ] Added new tests for this change
- [ ] Test coverage maintained or improved

### Manual Testing
Describe the manual testing you performed:
- [ ] Tested locally with Docker
- [ ] Tested SSH connection
- [ ] Tested gameplay functionality
- [ ] Tested with multiplayer (if applicable)
- [ ] Tested database migrations
- [ ] Tested on multiple platforms: [list platforms]

### Performance Testing
- [ ] No performance impact expected
- [ ] Performance tested and acceptable
- [ ] Performance improvement measured: [describe]

## Code Quality Checklist
- [ ] My code follows the project's style guidelines
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] My code passes `go vet` and `golangci-lint`
- [ ] My changes generate no new warnings
- [ ] I have kept functions small and focused
- [ ] I have avoided code duplication

## Documentation Checklist
- [ ] I have made corresponding changes to the documentation
- [ ] Code comments added/updated
- [ ] API documentation updated (if applicable)
- [ ] README updated (if needed)
- [ ] CHANGELOG.md updated
- [ ] Migration guide provided (if breaking change)

## Database Migration (if applicable)
- [ ] Migration script included in `scripts/migrations/`
- [ ] Rollback script provided
- [ ] Migration tested on clean database
- [ ] Migration tested on existing data
- [ ] Data backup recommended before applying

## Dependencies
- [ ] No new dependencies added
- [ ] New dependencies added (justify below)
- [ ] Dependencies updated (list below)
- [ ] `go.mod` and `go.sum` updated

**New/Updated Dependencies:**
```
List any new or updated dependencies and why they're needed
```

## Deployment Notes
Any special deployment considerations?
- [ ] Requires environment variables update
- [ ] Requires configuration changes
- [ ] Requires database migration
- [ ] Requires service restart
- [ ] Breaking change for existing deployments

**Deployment Instructions:**
```
Special steps needed for deployment, if any
```

## Screenshots/Terminal Output
Add screenshots or terminal output to demonstrate the changes.

**Before:**
```
[Screenshot or terminal output of old behavior]
```

**After:**
```
[Screenshot or terminal output of new behavior]
```

## Security Considerations
- [ ] No security implications
- [ ] Security review completed
- [ ] No credentials or secrets in code
- [ ] Input validation added
- [ ] SQL injection prevention verified
- [ ] XSS prevention verified (if applicable)

## Backwards Compatibility
- [ ] Fully backwards compatible
- [ ] Requires migration but compatible
- [ ] Breaking change (documented above)

## Additional Notes
Any additional information that reviewers should know:
- Known issues or limitations
- Future improvements planned
- Alternative approaches considered
- Performance benchmarks

## Reviewer Guidelines
Please review:
1. Code quality and style
2. Test coverage
3. Documentation completeness
4. Game balance implications
5. Security considerations
6. Performance impact

## Post-Merge Tasks
- [ ] Update project board
- [ ] Close related issues
- [ ] Notify in Discord/chat (if applicable)
- [ ] Create follow-up issues (if needed)
- [ ] Tag version (if release)

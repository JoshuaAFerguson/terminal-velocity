# Authentication Screens

This document covers all authentication-related UI screens in Terminal Velocity.

## Overview

**Screens**: 2
- Login Screen
- Registration Screen

**Purpose**: Handle user authentication via password or SSH key, and new account creation.

**Source Files**:
- `internal/tui/login.go` - Login screen implementation
- `internal/tui/registration.go` - Registration screen implementation
- `internal/server/server.go` - SSH authentication handlers

---

## Login Screen

### Source File
`internal/tui/login.go`

### Purpose
Primary authentication screen displayed when users connect via SSH. Supports both password and SSH key authentication.

### ASCII Prototype

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃                                                                              ┃
┃                                                                              ┃
┃                         ████████╗███████╗██████╗ ███╗   ███╗               ┃
┃                         ╚══██╔══╝██╔════╝██╔══██╗████╗ ████║               ┃
┃                            ██║   █████╗  ██████╔╝██╔████╔██║               ┃
┃                            ██║   ██╔══╝  ██╔══██╗██║╚██╔╝██║               ┃
┃                            ██║   ███████╗██║  ██║██║ ╚═╝ ██║               ┃
┃                            ╚═╝   ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝               ┃
┃                                                                              ┃
┃                        ██╗   ██╗███████╗██╗      ██████╗  ██████╗██╗████████╗██╗   ██╗
┃                        ██║   ██║██╔════╝██║     ██╔═══██╗██╔════╝██║╚══██╔══╝╚██╗ ██╔╝
┃                        ██║   ██║█████╗  ██║     ██║   ██║██║     ██║   ██║    ╚████╔╝
┃                        ╚██╗ ██╔╝██╔══╝  ██║     ██║   ██║██║     ██║   ██║     ╚██╔╝
┃                         ╚████╔╝ ███████╗███████╗╚██████╔╝╚██████╗██║   ██║      ██║
┃                          ╚═══╝  ╚══════╝╚══════╝ ╚═════╝  ╚═════╝╚═╝   ╚═╝      ╚═╝
┃                                                                              ┃
┃                           A Multiplayer Space Trading Game                  ┃
┃                                                                              ┃
┃                                                                              ┃
┃                  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓              ┃
┃                  ┃           LOGIN TO YOUR ACCOUNT           ┃              ┃
┃                  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫              ┃
┃                  ┃                                           ┃              ┃
┃                  ┃  Username: [___________________________]  ┃              ┃
┃                  ┃                                           ┃              ┃
┃                  ┃  Password: [***************************]  ┃              ┃
┃                  ┃                                           ┃              ┃
┃                  ┃           [ Login with Password ]         ┃              ┃
┃                  ┃           [ Login with SSH Key  ]         ┃              ┃
┃                  ┃                                           ┃              ┃
┃                  ┃  ─────────────── OR ───────────────────   ┃              ┃
┃                  ┃                                           ┃              ┃
┃                  ┃       [ Create New Account ]              ┃              ┃
┃                  ┃                                           ┃              ┃
┃                  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛              ┃
┃                                                                              ┃
┃                                                                              ┃
┃              Connect via SSH: ssh username@terminal-velocity.io:2222        ┃
┃                                                                              ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [Tab] Next Field  [Enter] Submit  [R]egister  [Q]uit                        ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Components
- **Title Banner**: ASCII art logo
- **Input Fields**: Username and password fields with masking
- **Button List**: Login options and registration link
- **Footer**: Key binding hints

### Key Bindings
- `Tab` - Navigate between fields
- `Enter` - Submit login or select button
- `R` - Jump to registration screen
- `Q` / `Ctrl+C` - Quit application

### State Management

**Model Structure** (`loginModel`):
```go
type loginModel struct {
    username     string
    password     string
    focusedField int          // 0=username, 1=password, 2=buttons
    selectedBtn  int          // Button selection index
    error        string       // Error message to display
    width        int
    height       int
}
```

**Messages**:
- `tea.KeyMsg` - Keyboard input
- `loginSuccessMsg` - Authentication succeeded
- `loginErrorMsg` - Authentication failed

### Data Flow
1. User enters credentials
2. On submit, sends authentication request
3. Server validates against database (`internal/server/server.go`)
4. Success: Transition to Main Menu
5. Failure: Display error message, clear password field

### Authentication Methods

**Password Authentication**:
- Username and password validated against `players` table
- Password hashed with bcrypt/scrypt
- Rate limited: 5 attempts before 15min lockout
- 20 failures = 24h IP ban

**SSH Key Authentication**:
- Public key fingerprint compared against `player_ssh_keys` table
- SHA256 fingerprint matching
- Automatic login if key matches

### Security Features
- Password masking in UI (displays asterisks)
- Rate limiting via `internal/ratelimit/`
- Auto-ban for brute force attempts
- Session tokens in SSH permissions

### Related Screens
- **Next**: Main Menu (`ScreenMainMenu`)
- **Alt**: Registration (`ScreenRegistration`)

---

## Registration Screen

### Source File
`internal/tui/registration.go`

### Purpose
New account creation with username, email, and password validation.

### ASCII Prototype

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃                                                                              ┃
┃                         ████████╗███████╗██████╗ ███╗   ███╗               ┃
┃                         ╚══██╔══╝██╔════╝██╔══██╗████╗ ████║               ┃
┃                            ██║   █████╗  ██████╔╝██╔████╔██║               ┃
┃                            ██║   ██╔══╝  ██╔══██╗██║╚██╔╝██║               ┃
┃                            ██║   ███████╗██║  ██║██║ ╚═╝ ██║               ┃
┃                            ╚═╝   ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝               ┃
┃                                                                              ┃
┃                        ██╗   ██╗███████╗██╗      ██████╗  ██████╗██╗████████╗██╗   ██╗
┃                        ██║   ██║██╔════╝██║     ██╔═══██╗██╔════╝██║╚══██╔══╝╚██╗ ██╔╝
┃                        ██║   ██║█████╗  ██║     ██║   ██║██║     ██║   ██║    ╚████╔╝
┃                        ╚██╗ ██╔╝██╔══╝  ██║     ██║   ██║██║     ██║   ██║     ╚██╔╝
┃                         ╚████╔╝ ███████╗███████╗╚██████╔╝╚██████╗██║   ██║      ██║
┃                          ╚═══╝  ╚══════╝╚══════╝ ╚═════╝  ╚═════╝╚═╝   ╚═╝      ╚═╝
┃                                                                              ┃
┃                           Create Your Account                                ┃
┃                                                                              ┃
┃                  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓              ┃
┃                  ┃         CREATE NEW ACCOUNT                ┃              ┃
┃                  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫              ┃
┃                  ┃                                           ┃              ┃
┃                  ┃  Username: [___________________________]  ┃              ┃
┃                  ┃  (3-20 characters, alphanumeric + _)      ┃              ┃
┃                  ┃                                           ┃              ┃
┃                  ┃  Email: [______________________________]  ┃              ┃
┃                  ┃  (Valid email address required)           ┃              ┃
┃                  ┃                                           ┃              ┃
┃                  ┃  Password: [***************************]  ┃              ┃
┃                  ┃  (Min 8 characters, mix of types)         ┃              ┃
┃                  ┃                                           ┃              ┃
┃                  ┃  Confirm: [****************************]  ┃              ┃
┃                  ┃                                           ┃              ┃
┃                  ┃           [ Create Account ]              ┃              ┃
┃                  ┃           [ Back to Login ]               ┃              ┃
┃                  ┃                                           ┃              ┃
┃                  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛              ┃
┃                                                                              ┃
┃              ℹ️ You will start with a Shuttle at a random starting system   ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [Tab] Next Field  [Enter] Submit  [ESC] Back to Login                       ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Components
- **Title Banner**: ASCII art logo
- **Input Fields**: Username, email, password, confirm password
- **Validation Hints**: Real-time feedback below each field
- **Button List**: Create account or return to login
- **Info Box**: Starting conditions message

### Key Bindings
- `Tab` / `Shift+Tab` - Navigate between fields
- `Enter` - Submit registration or select button
- `ESC` - Return to login screen

### State Management

**Model Structure** (`registrationModel`):
```go
type registrationModel struct {
    username        string
    email           string
    password        string
    confirmPassword string
    focusedField    int     // 0-3 for fields, 4+ for buttons
    selectedBtn     int
    validationErrors map[string]string
    error           string
    width           int
    height          int
}
```

**Messages**:
- `tea.KeyMsg` - Keyboard input
- `registrationSuccessMsg` - Account created
- `registrationErrorMsg` - Validation failed
- `fieldValidationMsg` - Real-time field validation

### Data Flow
1. User fills out registration form
2. Real-time validation on field changes
3. On submit, validate all fields
4. Create player in database (`database.PlayerRepository`)
5. Assign starter ship (Shuttle)
6. Set random starting location
7. Success: Transition to Tutorial or Main Menu
8. Failure: Display specific validation errors

### Validation Rules

**Username**:
- 3-20 characters
- Alphanumeric plus underscore only
- Must be unique (check database)
- Cannot contain profanity (word filter)

**Email**:
- Valid email format (regex validation)
- Must be unique (check database)
- Optional but recommended

**Password**:
- Minimum 8 characters
- Must contain: uppercase, lowercase, number
- Optional: special character for stronger security
- Must match confirmation field

### Server Configuration
Registration can be disabled server-side via config:
```yaml
server:
  allow_registration: true  # Set to false to disable
```

### New Player Defaults
Created in `internal/database/player_repository.go`:
- **Ship**: Shuttle (basic starter ship)
- **Credits**: 1,000 cr
- **Starting System**: Random low-tech system
- **Reputation**: 0 with all factions
- **Tutorial State**: Enabled by default

### Related Screens
- **Previous**: Login (`ScreenLogin`)
- **Next**: Tutorial (`ScreenTutorial`) or Main Menu (`ScreenMainMenu`)

---

## Implementation Notes

### Database Integration
Both screens interact with these repositories:
- `database.PlayerRepository` - Player account management
- `database.SSHKeyRepository` - SSH key storage and lookup
- `database.ShipRepository` - Starter ship creation
- `database.SystemRepository` - Random starting location

### Error Handling
Common error scenarios:
- Invalid credentials (login)
- Duplicate username/email (registration)
- Database connection failure
- Rate limit exceeded
- Validation failures

All errors display user-friendly messages in the UI and log technical details server-side.

### Testing
Test files:
- `internal/tui/integration_test.go` - End-to-end authentication flow
- `internal/tui/input_validation_test.go` - Field validation testing
- `internal/server/auth_test.go` - Server-side authentication

### Security Considerations
- Passwords never stored in plaintext
- Password hashing uses bcrypt (cost factor 10)
- SSH keys stored as SHA256 fingerprints
- Rate limiting prevents brute force
- Auto-ban for persistent attackers
- Session tokens prevent replay attacks

---

**Last Updated**: 2025-11-15
**Document Version**: 1.0.0

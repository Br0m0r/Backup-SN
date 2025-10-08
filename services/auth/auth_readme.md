# 🛡️ AUTH SERVICE - The Security Guard

Think of the auth service as the **bouncer at a nightclub** - it decides who gets in and who doesn't!

## 🏗️ FOLDER STRUCTURE

```
services/auth/
├── 📄 main.go             # Starting point - puts everything together
├── 📁 handlers/           # The "waiters" - handle HTTP requests
├── 📁 middleware/         # Security checks that run first
├── 📁 models/            # Data blueprints (User, LoginRequest, etc.)
├── 📁 services/          # The "brain" - business logic
├── 📁 db/                # Database operations
└── 📁 utils/             # Helper tools (password hashing, responses)
```

---

## 📁 KEY FOLDERS EXPLAINED

### 📄 **MAIN.GO** - *The Orchestra Conductor*
**What it does:** Starts everything up
1. Connects to database
2. Creates all handlers and services  
3. Sets up routes (/register, /login, /logout, /session)
4. Adds security middleware
5. Starts server on port 8081

### 📁 **HANDLERS** - *The Waiters*
**What they do:** Take HTTP requests and respond

- **`auth.go`** - Registration & Login Desk
  - `Register()` - Creates new accounts
  - `Login()` - Verifies credentials  
  - `Logout()` - Destroys sessions
- **`token.go`** - ID Card Checker
  - `VerifyToken()` - Checks if tokens are valid
- **`health.go`** - Health Check
  - `HealthHandler()` - "Are you alive?" checker

### 📁 **MIDDLEWARE** - *Security Checkpoints*
**What they do:** Run BEFORE handlers (like airport security)

- **`cors.go`** - Allows browsers from other websites to talk to us
- **`ratelimit.go`** - Prevents spam (max 10 requests/minute per IP)
- **`logging.go`** - Records all requests for debugging

### 📁 **MODELS** - *The Blueprints*
**What they define:** Data structures

- **`User`** struct - User profile info (ID, username, email, etc.)
- **`LoginRequest`** - Email + password for login
- **`RegisterRequest`** - All info needed to create account
- **`AuthResponse`** - What we send back (user info + token)
- **`Validate()`** functions - Check if data is correct

### 📁 **SERVICES** - *The Brain*
**What they do:** Smart business logic

- **`auth_service.go`** - Main auth logic
  - `Register()` - Complete signup process
  - `Login()` - Complete login process
  - `VerifyToken()` - Check if token is real
  - `Logout()` - Destroy session
- **`token_service.go`** - Session management
  - `GenerateToken()` - Create new session tokens
  - `ValidateToken()` - Check token validity
  - Auto-cleanup of expired sessions

### 📁 **DB** - *The Librarian*
**What it does:** Talks to the database

- **`queries.go`** - All database operations
  - `CreateUser()` - Save new user to database
  - `GetUserByEmail()` - Find user by email
  - `GetUserByID()` - Find user by ID
  - `UserExistsByEmail()` - Check if email is taken
  - `UserExistsByUsername()` - Check if username is taken

### 📁 **UTILS** - *The Toolbox*
**What they provide:** Helper functions

- **`password.go`** - Password security
  - `HashPassword()` - Scramble passwords with bcrypt
  - `CheckPassword()` - Verify password matches hash
- **`response.go`** - Response formatting
  - `SuccessResponse()` - Send success JSON
  - `ErrorResponse()` - Send error JSON

---

## 🔄 HOW IT ALL WORKS TOGETHER

```
1. 👤 User: "I want to sign up!"

2. 🌐 Request: POST /register {username, email, password}

3. 🛡️ Middleware: Security checks ✓

4. 🍽️ Handler: "Registration request received"

5. 🧠 Service: 
   - Validates data
   - Checks if username/email exists
   - Hashes password
   - Saves to database
   - Creates session token

6. 📚 Database: Saves new user

7. 📦 Response: {user: {...}, token: "abc123"}

8. 👤 User: Gets account + session token
```

## 🎯 WHAT THE AUTH SERVICE DOES

- ✅ **User Registration** - Create new accounts
- ✅ **User Login** - Verify credentials and create sessions
- ✅ **Token Management** - Generate and validate session tokens
- ✅ **User Logout** - Destroy sessions
- ✅ **Security** - Rate limiting, CORS, password hashing
- ✅ **Health Checks** - Monitor service status

**Bottom line:** It's your app's security department that handles who can join and who can access what! 🚀
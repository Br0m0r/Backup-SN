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
3. Sets up routes:
   - Public: /register, /login, /logout, /session
   - Internal: /internal/verify-token, /internal/user/:id
4. Adds security middleware
5. Starts server on port 8081

### 📁 **HANDLERS** - *The Waiters*
**What they do:** Take HTTP requests and respond

- **`auth.go`** - Registration & Login Desk
  - `Register()` - Creates new accounts
  - `Login()` - Verifies credentials  
  - `Logout()` - Destroys sessions
- **`token.go`** - Session & Token Management
  - `GetSession()` - Get current user session (for frontend)
  - `VerifyToken()` - Verify tokens (for microservices)
  - `GetUserByID()` - Get user info (for microservices)
- **`health.go`** - Health Check
  - `HealthHandler()` - "Are you alive?" checker

### 📁 **MIDDLEWARE** - *Security Checkpoints*
**What they do:** Run BEFORE handlers (like airport security)

- **`cors.go`** - Allows browsers from other websites to talk to us
- **`ratelimit.go`** - Prevents spam (max 10 requests/minute per IP)
- **`logging.go`** - Records all requests for debugging

### 📁 **MODELS** - *The Blueprints*
**What they define:** Data structures (no business logic)

- **`User`** struct - User profile info (ID, username, email, etc.)
- **`LoginRequest`** - Email + password for login
- **`RegisterRequest`** - All info needed to create account
- **`AuthResponse`** - What we send back (user info + token)

### 📁 **SERVICES** - *The Brain*
**What they do:** Smart business logic

- **`auth_service.go`** - Main auth logic
  - `Register()` - Complete signup process with validation
  - `Login()` - Complete login process
  - `VerifyToken()` - Check if token is real
  - `Logout()` - Destroy session from database
  - `GetUserByID()` - Get user info (for microservices)
- **`token_service.go`** - Database-backed session management
  - `GenerateToken()` - Create & store session tokens in DB
  - `ValidateToken()` - Check token validity from DB
  - Auto-cleanup of expired sessions from database

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
- **`validation.go`** - Input validation
  - `ValidateRegisterRequest()` - Check registration data
  - `ValidateLoginRequest()` - Check login data

---

## 🔄 HOW IT ALL WORKS TOGETHER

```
1. 👤 User: "I want to sign up!"

2. 🌐 Request: POST /register {username, email, password}

3. 🛡️ Middleware: Security checks ✓

4. 🍽️ Handler: "Registration request received"

5. 🧠 Service: 
   - Validates data (using utils.ValidateRegisterRequest)
   - Checks if username/email exists
   - Hashes password
   - Saves to database
   - Creates session token (stored in sessions table)

6. 📚 Database: Saves new user + session token

7. 📦 Response: {user: {...}, token: "abc123..."}

8. 👤 User: Gets account + session token
```

## 🎯 WHAT THE AUTH SERVICE DOES

- ✅ **User Registration** - Create new accounts with validation
- ✅ **User Login** - Verify credentials and create database-backed sessions
- ✅ **Token Management** - Generate & validate session tokens (stored in DB)
- ✅ **User Logout** - Destroy sessions from database
- ✅ **Microservice Support** - Internal endpoints for other services
- ✅ **Security** - Rate limiting, CORS, password hashing (bcrypt)
- ✅ **Health Checks** - Monitor service status

**Bottom line:** It's your app's security department that handles who can join and who can access what! 🚀
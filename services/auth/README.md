# Authentication Service - Complete Guide

**A comprehensive walkthrough of the authentication system with sessions, tokens, and password security**

---

## Table of Contents

1. [Overview - The Big Picture](#overview---the-big-picture)
2. [Go Methods vs Functions](#go-methods-vs-functions)
3. [Service Architecture](#service-architecture)
4. [Constructors in Go](#constructors-in-go)
5. [Password Security](#password-security)
6. [Sessions and Tokens](#sessions-and-tokens)
7. [The Complete Flow - User Registration](#the-complete-flow---user-registration)
8. [The Complete Flow - User Login](#the-complete-flow---user-login)
9. [Token Verification](#token-verification)
10. [Database Operations](#database-operations)
11. [HTTP REST Endpoints](#http-rest-endpoints)
12. [Error Handling](#error-handling)

---

## Overview - The Big Picture

### What Does This Service Do?

The authentication service is the **gatekeeper** of your social network. It:
- Creates new user accounts (registration)
- Verifies user identity (login)
- Issues authentication tokens (sessions)
- Validates tokens for other services
- Manages user logout

### Architecture at a Glance

```
Browser                Auth Service              Database
   |                        |                        |
   |--[Register]----------->|                        |
   |                        |--[Create User]-------->|
   |                        |<-[User Created]--------|
   |                        |--[Create Session]----->|
   |                        |<-[Session Token]-------|
   |<-[Token + User Info]---|                        |
   |                        |                        |
   |--[Request with Token]->|                        |
   |                        |--[Verify Token]------->|
   |                        |<-[Valid/Invalid]-------|
   |<-[Allow/Deny]----------|                        |
```

### Key Components

1. **Handlers**: Receive HTTP requests and send responses
2. **Services**: Contain business logic (validation, token generation)
3. **Database**: Store users, sessions, hashed passwords
4. **Middleware**: Verify tokens on protected routes
5. **Utils**: Helper functions (password hashing, validation)

[Back to Top](#table-of-contents)

---

## Go Methods vs Functions

### What's the Difference?

In Go, there are **functions** and **methods**. They look similar but serve different purposes.

### Regular Function

A function is **standalone** and doesn't belong to any specific type:

```go
// Regular function - standalone
func Add(a int, b int) int {
    return a + b
}

// Call it anywhere
result := Add(5, 3)  // result = 8
```

### Method (Function with Receiver)

A method is a function that **belongs to a specific type** (like a struct):

```go
// Define a struct
type Calculator struct {
    memory int
}

// Method with receiver - belongs to Calculator
func (c *Calculator) Add(a int, b int) int {
    result := a + b
    c.memory = result  // Can access and modify struct fields
    return result
}

// Usage
calc := Calculator{memory: 0}
result := calc.Add(5, 3)  // Call method on calc instance
// calc.memory is now 8
```

**Key Difference:**
- `(c *Calculator)` before the function name makes it a **method**
- `c` is called the **receiver**
- Methods can access the struct's fields (`c.memory`)

### Real Example from Auth Service

```go
// AuthService struct holds state (database, tokenService)
type AuthService struct {
    database     *sql.DB
    tokenService *TokenService
}

// Method - belongs to AuthService
func (s *AuthService) Login(req *models.LoginRequest) (*models.AuthResponse, error) {
    // 's' is the receiver - we can access s.database, s.tokenService
    user, err := db.GetUserByEmail(s.database, req.Email)
    token, err := s.tokenService.GenerateToken(user.ID, user.Username, user.Email)
    return &models.AuthResponse{User: user, Token: token}, nil
}
```

**Why use methods?**
- **Encapsulation**: Data (fields) and behavior (methods) stay together
- **State access**: Methods can use and modify struct fields
- **Organization**: Groups related functionality

**Think of it like this:**
- **Function**: "Make coffee" (standalone action)
- **Method**: "coffeMachine.MakeCoffee()" (action specific to coffee machine, uses machine's water, beans, etc.)

### The `*` in Receivers

```go
func (s *AuthService) Login(...)  // Pointer receiver
func (s AuthService) Login(...)   // Value receiver
```

**Pointer receiver `(s *AuthService)`:**
- Can **modify** the struct's fields
- More efficient (doesn't copy the entire struct)
- **Most common** for services

**Value receiver `(s AuthService)`:**
- **Cannot** modify the original struct
- Makes a copy of the struct
- Used for small, immutable types

[Back to Top](#table-of-contents)

---

## Service Architecture

### What is a Service?

A **service** is a struct that groups related business logic. Think of it as a "worker" that knows how to do specific tasks.

### Layers in Our Architecture

```
┌─────────────────────────────────────────┐
│         HTTP Handler Layer              │  ← Receives requests, sends responses
│    (handlers/auth.go)                   │
└──────────────┬──────────────────────────┘
               │
               ↓
┌─────────────────────────────────────────┐
│         Service Layer                   │  ← Business logic, validation
│    (services/auth_service.go)          │
│    (services/token_service.go)         │
└──────────────┬──────────────────────────┘
               │
               ↓
┌─────────────────────────────────────────┐
│         Database Layer                  │  ← Database queries
│    (db/queries.go)                     │
└─────────────────────────────────────────┘
```

### Why This Structure?

**Separation of Concerns** - Each layer has ONE job:

1. **Handlers**: "I handle HTTP" (parse JSON, send responses)
2. **Services**: "I handle business logic" (validate, hash passwords)
3. **Database**: "I handle data storage" (SQL queries)

**Example: User Registration**

```go
// Handler Layer - HTTP concerns
func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
    // Parse JSON request
    var req models.RegisterRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // Call service (business logic)
    response, err := h.authService.Register(&req)
    
    // Send HTTP response
    json.NewEncoder(w).Encode(response)
}

// Service Layer - Business logic
func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
    // Validate data
    if err := utils.ValidateRegisterRequest(req); err != nil {
        return nil, err
    }
    
    // Hash password
    hashedPassword, _ := utils.HashPassword(req.Password)
    
    // Call database layer
    user, err := db.CreateUser(s.database, req.Username, req.Email, hashedPassword, ...)
    
    return response, nil
}

// Database Layer - SQL queries
func CreateUser(db *sql.DB, username, email, passwordHash string, ...) (*User, error) {
    query := "INSERT INTO users (username, email, password_hash, ...) VALUES (?, ?, ?, ...)"
    result, err := db.Exec(query, username, email, passwordHash, ...)
    return user, nil
}
```

### Benefits of This Architecture

1. **Testability**: Test each layer independently
2. **Reusability**: Services can be called from different handlers
3. **Maintainability**: Change one layer without affecting others
4. **Clarity**: Easy to understand where logic belongs

[Back to Top](#table-of-contents)

---

## Constructors in Go

### What is a Constructor?

Go doesn't have built-in constructors like Java or C++. Instead, we use **factory functions** (functions that create and return instances).

### Convention: `New<TypeName>`

```go
// Constructor pattern - creates and returns a new instance
func NewAuthService(database *sql.DB) *AuthService {
    return &AuthService{
        database:     database,
        tokenService: NewTokenService(database),
    }
}
```

### Why Use Constructors?

**Without Constructor:**
```go
// Manual initialization - error-prone
service := AuthService{
    database:     db,
    tokenService: nil,  // Oops! Forgot to initialize
}
```

**With Constructor:**
```go
// Guaranteed proper initialization
service := NewAuthService(db)
// tokenService is automatically created
```

### Constructor Benefits

1. **Guaranteed initialization**: All fields are set correctly
2. **Encapsulation**: Hide complex setup logic
3. **Consistency**: Always create objects the same way
4. **Dependencies**: Clearly show what's required

### Real Example from Auth Service

```go
type AuthService struct {
    database     *sql.DB        // Needs database connection
    tokenService *TokenService  // Needs token service
}

// Constructor ensures both are properly initialized
func NewAuthService(database *sql.DB) *AuthService {
    return &AuthService{
        database:     database,                    // Store database
        tokenService: NewTokenService(database),  // Create token service
    }
}
```

**Usage:**
```go
// In main.go
db, _ := OpenDB("social_network.db")
authService := NewAuthService(db)  // One line, everything ready!

// Now use it
response, err := authService.Login(loginRequest)
```

### Nested Constructors

Notice how `NewAuthService` calls `NewTokenService`:

```go
func NewAuthService(database *sql.DB) *AuthService {
    return &AuthService{
        database:     database,
        tokenService: NewTokenService(database),  // Another constructor!
    }
}

func NewTokenService(database *sql.DB) *TokenService {
    return &TokenService{
        database: database,
    }
}
```

This is **dependency injection** - passing dependencies (database) to objects that need them.

[Back to Top](#table-of-contents)

---

## Password Security

### Never Store Plain Passwords!

**Wrong:**
```sql
INSERT INTO users (email, password) VALUES ('user@email.com', 'MyPassword123');
```

**If database is stolen, attacker has all passwords!**

### Hashing - One-Way Encryption

A **hash** is like a shredder:
- You can shred a document (hash a password)
- You **cannot** unshred it (cannot reverse a hash)
- Same document always shreds the same way (same password = same hash)

```go
password := "MyPassword123"
hash := HashPassword(password)
// hash = "$2a$10$N9qo8uLOickgx2ZMRZoMy.EhFf9..."

// Cannot reverse it!
original := UnhashPassword(hash)  // IMPOSSIBLE!
```

### bcrypt - Industry Standard

We use **bcrypt** for password hashing:

```go
import "golang.org/x/crypto/bcrypt"

// Hash password (when registering)
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// Check password (when logging in)
func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

### Why bcrypt?

1. **Slow by design**: Makes brute-force attacks impractical
2. **Salt included**: Each hash is unique even for same password
3. **Future-proof**: Can increase difficulty as computers get faster

### How It Works

**Registration:**
```
User enters: "MyPassword123"
     ↓
Hash it: "$2a$10$N9qo8uLOickgx2ZMRZoMy.EhFf9..."
     ↓
Store hash in database (NOT original password)
```

**Login:**
```
User enters: "MyPassword123"
     ↓
Get stored hash: "$2a$10$N9qo8uLOickgx2ZMRZoMy.EhFf9..."
     ↓
Compare: bcrypt.CompareHashAndPassword(storedHash, enteredPassword)
     ↓
Match? → Login success!
No match? → Invalid password
```

### Salt - Extra Security

bcrypt automatically adds a **salt** (random data):

```
Password: "MyPassword123"
Salt:     "random_xyz_789"
Hash:     Hash("MyPassword123" + "random_xyz_789")
```

**Why?**
- Two users with same password get **different hashes**
- Prevents "rainbow table" attacks (pre-computed hash databases)

[Back to Top](#table-of-contents)

---

## Sessions and Tokens

### What is a Session?

HTTP is **stateless** - each request is independent. The server doesn't remember you!

```
Browser: "Show me my profile"
Server: "Who are you?" ← Doesn't remember previous requests!
```

**Solution: Sessions with Tokens**

```
Browser: "Here's my token: abc123xyz"
Server: "Ah! You're User #5. Here's your profile."
```

### Token Structure

```go
type Session struct {
    ID        string    // Unique session ID (the token)
    UserID    int       // Which user owns this session
    Username  string    // User's username (for quick access)
    Email     string    // User's email
    ExpiresAt time.Time // When token expires
    CreatedAt time.Time // When created
}
```

### How Tokens Work

**1. Generate Token (Login)**
```go
func GenerateToken(userID int, username, email string) (string, error) {
    // Create unique token
    token := uuid.New().String()
    // token = "550e8400-e29b-41d4-a716-446655440000"
    
    // Store in database
    session := &Session{
        ID:        token,
        UserID:    userID,
        Username:  username,
        Email:     email,
        ExpiresAt: time.Now().Add(24 * time.Hour),  // Valid for 24 hours
        CreatedAt: time.Now(),
    }
    
    // Save to sessions table
    SaveSession(database, session)
    
    return token, nil
}
```

**2. Validate Token (Every Request)**
```go
func ValidateToken(token string) (*Session, error) {
    // Look up token in database
    session, err := GetSessionByToken(database, token)
    if err != nil {
        return nil, errors.New("invalid token")
    }
    
    // Check if expired
    if time.Now().After(session.ExpiresAt) {
        DeleteSession(database, token)
        return nil, errors.New("token expired")
    }
    
    return session, nil
}
```

**3. Invalidate Token (Logout)**
```go
func InvalidateToken(token string) error {
    // Delete from database
    return DeleteSession(database, token)
}
```

### Token Flow Diagram

```
┌─────────────────────────────────────────────────────────┐
│  1. Login: User enters email + password                 │
│     ↓                                                    │
│  2. Validate: Check password hash                        │
│     ↓                                                    │
│  3. Generate Token: Create UUID                          │
│     ↓                                                    │
│  4. Store: Save to sessions table                        │
│     ↓                                                    │
│  5. Return: Send token to browser                        │
│     ↓                                                    │
│  6. Browser: Store token (localStorage/cookie)           │
│     ↓                                                    │
│  7. Future Requests: Send token in Authorization header  │
│     ↓                                                    │
│  8. Middleware: Validate token before allowing access    │
└─────────────────────────────────────────────────────────┘
```

### Database Table

```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,              -- The token itself
    user_id INTEGER NOT NULL,         -- Which user
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    expires_at DATETIME NOT NULL,     -- When it expires
    created_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### Token in HTTP Requests

**Browser sends token in header:**
```http
GET /api/profile HTTP/1.1
Host: localhost:8081
Authorization: Bearer 550e8400-e29b-41d4-a716-446655440000
```

**Server validates:**
```go
token := r.Header.Get("Authorization")  // "Bearer 550e8400..."
token = strings.TrimPrefix(token, "Bearer ")

session, err := tokenService.ValidateToken(token)
if err != nil {
    http.Error(w, "Unauthorized", 401)
    return
}

// Token valid! User is session.UserID
```

[Back to Top](#table-of-contents)

---

## The Complete Flow - User Registration

Let's follow what happens when a new user signs up with email "alice@example.com" and password "SecurePass123".

---

### Step 1: Browser Sends Registration Request

```javascript
// Frontend JavaScript
const registerData = {
    username: "alice",
    email: "alice@example.com",
    password: "SecurePass123",
    first_name: "Alice",
    last_name: "Smith",
    date_of_birth: "1995-05-15"
};

fetch('http://localhost:8081/register', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(registerData)
});
```

---

### Step 2: Handler Receives Request

```go
// handlers/auth.go
func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
    // 1. Parse JSON from request body
    var req models.RegisterRequest
    err := json.NewDecoder(r.Body).Decode(&req)
    // req.Email = "alice@example.com"
    // req.Password = "SecurePass123"
    // req.FirstName = "Alice"
    
    // 2. Call service to handle business logic
    response, err := h.authService.Register(&req)
    
    // 3. Send response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

---

### Step 3: Service Validates Request

```go
// services/auth_service.go
func (s *AuthService) Register(req *models.RegisterRequest) (*models.AuthResponse, error) {
    // 1. Validate all fields
    if err := utils.ValidateRegisterRequest(req); err != nil {
        return nil, err  // "email is invalid", "password too short", etc.
    }
    
    // Checks:
    // - Email format valid?
    // - Password at least 8 characters?
    // - Required fields not empty?
```

---

### Step 4: Check if User Already Exists

```go
    // 2. Check if username taken
    exists, _ := db.UserExistsByUsername(s.database, "alice")
    if exists {
        return nil, errors.New("username already exists")
    }
    
    // 3. Check if email taken
    exists, _ := db.UserExistsByEmail(s.database, "alice@example.com")
    if exists {
        return nil, errors.New("email already exists")
    }
```

**Database Query:**
```sql
SELECT COUNT(*) FROM users WHERE email = 'alice@example.com';
-- Result: 0 (user doesn't exist, good to proceed)
```

---

### Step 5: Hash Password

```go
    // 4. Hash password (NEVER store plain password!)
    hashedPassword, _ := utils.HashPassword("SecurePass123")
    // hashedPassword = "$2a$10$N9qo8uLOickgx2ZMRZoMy.EhFf9YvJp..."
```

**What happens in HashPassword:**
```go
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}
```

---

### Step 6: Create User in Database

```go
    // 5. Create user
    user, err := db.CreateUser(
        s.database,
        "alice",                    // username
        "alice@example.com",        // email
        hashedPassword,             // hashed password
        "Alice",                    // first_name
        "Smith"                     // last_name
    )
    // user.ID = 42 (auto-generated)
```

**Database Query:**
```sql
INSERT INTO users (
    username, email, password_hash, first_name, last_name, 
    date_of_birth, is_public_profile, created_at
) VALUES (
    'alice', 'alice@example.com', '$2a$10$N9qo8u...', 
    'Alice', 'Smith', '1995-05-15', 1, datetime('now')
);
-- Returns ID: 42
```

---

### Step 7: Generate Authentication Token

```go
    // 6. Generate token (auto-login after registration)
    token, err := s.tokenService.GenerateToken(
        user.ID,       // 42
        user.Username, // "alice"
        user.Email     // "alice@example.com"
    )
    // token = "550e8400-e29b-41d4-a716-446655440000"
```

**What happens in GenerateToken:**
```go
func (t *TokenService) GenerateToken(userID int, username, email string) (string, error) {
    // Create unique UUID
    token := uuid.New().String()
    
    // Create session
    session := &models.Session{
        ID:        token,
        UserID:    userID,
        Username:  username,
        Email:     email,
        ExpiresAt: time.Now().Add(24 * time.Hour),
        CreatedAt: time.Now(),
    }
    
    // Save to database
    db.SaveSession(t.database, session)
    
    return token, nil
}
```

**Database Query:**
```sql
INSERT INTO sessions (
    id, user_id, username, email, expires_at, created_at
) VALUES (
    '550e8400-e29b-41d4-a716-446655440000',
    42,
    'alice',
    'alice@example.com',
    datetime('now', '+24 hours'),
    datetime('now')
);
```

---

### Step 8: Return Response to Browser

```go
    // 7. Return response
    return &models.AuthResponse{
        User:  user,   // User info (without password!)
        Token: token,  // Authentication token
    }, nil
}
```

**JSON Response:**
```json
{
    "user": {
        "id": 42,
        "username": "alice",
        "email": "alice@example.com",
        "first_name": "Alice",
        "last_name": "Smith",
        "date_of_birth": "1995-05-15",
        "is_public_profile": true,
        "created_at": "2025-10-16T15:30:00Z"
    },
    "token": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

### Step 9: Browser Stores Token

```javascript
// Frontend receives response
const response = await fetch('/register', {...});
const data = await response.json();

// Store token for future requests
localStorage.setItem('session_token', data.token);

// Redirect to homepage
window.location.href = '/home';
```

---

### Complete Timeline

```
[0.0s] User fills form and clicks "Register"
[0.1s] Browser sends POST /register with JSON
[0.2s] Handler receives request, parses JSON
[0.3s] Service validates fields (email format, password length)
[0.4s] Check database: Does username exist? → No
[0.5s] Check database: Does email exist? → No
[0.6s] Hash password with bcrypt (takes ~100ms)
[0.7s] Insert user into users table → User ID = 42
[0.8s] Generate UUID token
[0.9s] Insert session into sessions table
[1.0s] Return response with user info + token
[1.1s] Browser stores token in localStorage
[1.2s] User is now logged in!
```

**Total time: ~1.2 seconds**

[Back to Top](#table-of-contents)

---

## The Complete Flow - User Login

Let's follow what happens when Alice logs in with email "alice@example.com" and password "SecurePass123".

---

### Step 1: Browser Sends Login Request

```javascript
const loginData = {
    email: "alice@example.com",
    password: "SecurePass123"
};

fetch('http://localhost:8081/login', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(loginData)
});
```

---

### Step 2: Handler Receives Request

```go
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
    var req models.LoginRequest
    json.NewDecoder(r.Body).Decode(&req)
    // req.Email = "alice@example.com"
    // req.Password = "SecurePass123"
    
    response, err := h.authService.Login(&req)
    
    json.NewEncoder(w).Encode(response)
}
```

---

### Step 3: Service Validates Request

```go
func (s *AuthService) Login(req *models.LoginRequest) (*models.AuthResponse, error) {
    // 1. Validate fields
    if err := utils.ValidateLoginRequest(req); err != nil {
        return nil, err
    }
```

---

### Step 4: Get User from Database

```go
    // 2. Get user by email
    user, err := db.GetUserByEmail(s.database, "alice@example.com")
    if err != nil {
        return nil, errors.New("invalid email or password")
    }
    // user.ID = 42
    // user.PasswordHash = "$2a$10$N9qo8uLOickgx2ZMRZoMy..."
```

**Database Query:**
```sql
SELECT id, username, email, password_hash, first_name, last_name, ...
FROM users
WHERE email = 'alice@example.com';
```

**Important:** Even if email not found, we return generic "invalid email or password" (don't reveal which field is wrong - security!)

---

### Step 5: Verify Password

```go
    // 3. Check password
    if !utils.CheckPassword("SecurePass123", user.PasswordHash) {
        return nil, errors.New("invalid email or password")
    }
```

**What happens in CheckPassword:**
```go
func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

**bcrypt does:**
1. Extract salt from stored hash
2. Hash entered password with same salt
3. Compare results
4. Match? → Password correct!

---

### Step 6: Generate New Token

```go
    // 4. Generate new session token
    token, err := s.tokenService.GenerateToken(
        user.ID,       // 42
        user.Username, // "alice"
        user.Email     // "alice@example.com"
    )
    // token = "7c9e6679-7425-40de-944b-e07fc1f90ae7"
```

**Database Query:**
```sql
INSERT INTO sessions (id, user_id, username, email, expires_at, created_at)
VALUES (
    '7c9e6679-7425-40de-944b-e07fc1f90ae7',
    42,
    'alice',
    'alice@example.com',
    datetime('now', '+24 hours'),
    datetime('now')
);
```

---

### Step 7: Return Response

```go
    return &models.AuthResponse{
        User:  user,
        Token: token,
    }, nil
}
```

**JSON Response:**
```json
{
    "user": {
        "id": 42,
        "username": "alice",
        "email": "alice@example.com",
        "first_name": "Alice",
        "last_name": "Smith"
    },
    "token": "7c9e6679-7425-40de-944b-e07fc1f90ae7"
}
```

---

### Step 8: Browser Uses Token for Future Requests

```javascript
// Store token
localStorage.setItem('session_token', data.token);

// All future requests include token
fetch('/api/posts', {
    headers: {
        'Authorization': 'Bearer 7c9e6679-7425-40de-944b-e07fc1f90ae7'
    }
});
```

---

### Complete Timeline

```
[0.0s] User enters email + password, clicks "Login"
[0.1s] Browser sends POST /login
[0.2s] Handler parses JSON
[0.3s] Service validates fields
[0.4s] Query database for user by email
[0.5s] User found: ID=42, get password_hash
[0.6s] bcrypt.CompareHashAndPassword (takes ~100ms)
[0.7s] Password correct!
[0.8s] Generate new UUID token
[0.9s] Insert session into sessions table
[1.0s] Return user + token
[1.1s] Browser stores token
[1.2s] User logged in!
```

**Total time: ~1.2 seconds**

[Back to Top](#table-of-contents)

---

## Token Verification

### Why Verify Tokens?

Every protected endpoint needs to verify the user is authenticated:

```
User: "Show me my private posts"
Server: "Prove you're logged in" ← Needs token verification
```

### Middleware - Automatic Verification

Instead of checking tokens in every handler, we use **middleware**:

```go
// Middleware runs BEFORE handler
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. Extract token from header
        token := r.Header.Get("Authorization")
        token = strings.TrimPrefix(token, "Bearer ")
        
        // 2. Validate token
        session, err := tokenService.ValidateToken(token)
        if err != nil {
            http.Error(w, "Unauthorized", 401)
            return  // Stop here, don't call next handler
        }
        
        // 3. Add user ID to context
        ctx := context.WithValue(r.Context(), "userID", session.UserID)
        r = r.WithContext(ctx)
        
        // 4. Call next handler (token is valid!)
        next.ServeHTTP(w, r)
    })
}
```

### How It Works

```
Request → Middleware → Handler
   ↓           ↓            ↓
 Token?    Validate    Use userID
            Valid? ──→ Allow
            Invalid?   Deny (401)
```

### ValidateToken Function

```go
func (t *TokenService) ValidateToken(token string) (*models.Session, error) {
    // 1. Query database
    session, err := db.GetSessionByToken(t.database, token)
    if err != nil {
        return nil, errors.New("invalid token")
    }
    
    // 2. Check expiration
    if time.Now().After(session.ExpiresAt) {
        // Token expired, delete it
        db.DeleteSession(t.database, token)
        return nil, errors.New("token expired")
    }
    
    // 3. Token valid!
    return session, nil
}
```

**Database Query:**
```sql
SELECT id, user_id, username, email, expires_at, created_at
FROM sessions
WHERE id = '7c9e6679-7425-40de-944b-e07fc1f90ae7';
```

### Protected Route Example

```go
// main.go
mux := http.NewServeMux()

// Public routes (no auth needed)
mux.HandleFunc("/register", authHandlers.Register)
mux.HandleFunc("/login", authHandlers.Login)
mux.HandleFunc("/health", healthCheck)

// Protected routes (auth required)
mux.Handle("/profile", authMiddleware(http.HandlerFunc(getProfile)))
mux.Handle("/posts", authMiddleware(http.HandlerFunc(createPost)))
```

### Request Flow with Middleware

```
[Request] GET /profile
    ↓
[Middleware] Extract token from header
    ↓
[Middleware] Query sessions table
    ↓
[Middleware] Token valid? → YES
    ↓
[Middleware] Add userID to context
    ↓
[Handler] GetProfile
    ↓
[Handler] userID := context.Value("userID")
    ↓
[Handler] Query user's profile data
    ↓
[Response] Return profile JSON
```

### Context - Passing Data Between Middleware and Handlers

```go
// Middleware sets userID in context
ctx := context.WithValue(r.Context(), "userID", 42)
r = r.WithContext(ctx)

// Handler retrieves userID from context
func GetProfile(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("userID").(int)  // userID = 42
    // Now we know which user is making the request!
}
```

[Back to Top](#table-of-contents)

---

## Database Operations

### Users Table

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    date_of_birth DATE,
    avatar_url TEXT,
    nickname TEXT,
    about_me TEXT,
    is_public_profile BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);
```

### Sessions Table

```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### Key Database Functions

**CreateUser**
```go
func CreateUser(db *sql.DB, username, email, passwordHash, firstName, lastName string) (*User, error) {
    query := `
        INSERT INTO users (username, email, password_hash, first_name, last_name, created_at)
        VALUES (?, ?, ?, ?, ?, datetime('now'))
    `
    result, _ := db.Exec(query, username, email, passwordHash, firstName, lastName)
    id, _ := result.LastInsertId()
    
    return GetUserByID(db, int(id))
}
```

**GetUserByEmail**
```go
func GetUserByEmail(db *sql.DB, email string) (*User, error) {
    query := `
        SELECT id, username, email, password_hash, first_name, last_name, 
               date_of_birth, avatar_url, nickname, about_me, is_public_profile, created_at
        FROM users
        WHERE email = ?
    `
    var user User
    err := db.QueryRow(query, email).Scan(
        &user.ID, &user.Username, &user.Email, &user.PasswordHash,
        &user.FirstName, &user.LastName, &user.DateOfBirth, &user.AvatarURL,
        &user.Nickname, &user.AboutMe, &user.IsPublicProfile, &user.CreatedAt,
    )
    return &user, err
}
```

**SaveSession**
```go
func SaveSession(db *sql.DB, session *Session) error {
    query := `
        INSERT INTO sessions (id, user_id, username, email, expires_at, created_at)
        VALUES (?, ?, ?, ?, ?, ?)
    `
    _, err := db.Exec(query, 
        session.ID, 
        session.UserID, 
        session.Username, 
        session.Email,
        session.ExpiresAt, 
        session.CreatedAt,
    )
    return err
}
```

**GetSessionByToken**
```go
func GetSessionByToken(db *sql.DB, token string) (*Session, error) {
    query := `
        SELECT id, user_id, username, email, expires_at, created_at
        FROM sessions
        WHERE id = ?
    `
    var session Session
    err := db.QueryRow(query, token).Scan(
        &session.ID,
        &session.UserID,
        &session.Username,
        &session.Email,
        &session.ExpiresAt,
        &session.CreatedAt,
    )
    return &session, err
}
```

**DeleteSession (Logout)**
```go
func DeleteSession(db *sql.DB, token string) error {
    query := "DELETE FROM sessions WHERE id = ?"
    _, err := db.Exec(query, token)
    return err
}
```

[Back to Top](#table-of-contents)

---

## HTTP REST Endpoints

### Available Endpoints

| Method | Endpoint | Purpose | Auth Required |
|--------|----------|---------|---------------|
| GET | `/health` | Health check | No |
| POST | `/register` | Create new account | No |
| POST | `/login` | Authenticate user | No |
| POST | `/logout` | Invalidate token | Yes |
| GET | `/session` | Verify token (for other services) | Yes |
| GET | `/users/:id` | Get user by ID (internal) | Yes |

---

### POST /register

**Request:**
```http
POST /register HTTP/1.1
Content-Type: application/json

{
    "username": "alice",
    "email": "alice@example.com",
    "password": "SecurePass123",
    "first_name": "Alice",
    "last_name": "Smith",
    "date_of_birth": "1995-05-15"
}
```

**Response (Success):**
```json
{
    "user": {
        "id": 42,
        "username": "alice",
        "email": "alice@example.com",
        "first_name": "Alice",
        "last_name": "Smith",
        "is_public_profile": true,
        "created_at": "2025-10-16T15:30:00Z"
    },
    "token": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response (Error):**
```json
{
    "error": "email already exists"
}
```

---

### POST /login

**Request:**
```http
POST /login HTTP/1.1
Content-Type: application/json

{
    "email": "alice@example.com",
    "password": "SecurePass123"
}
```

**Response:**
```json
{
    "user": {
        "id": 42,
        "username": "alice",
        "email": "alice@example.com",
        "first_name": "Alice",
        "last_name": "Smith"
    },
    "token": "7c9e6679-7425-40de-944b-e07fc1f90ae7"
}
```

---

### POST /logout

**Request:**
```http
POST /logout HTTP/1.1
Authorization: Bearer 7c9e6679-7425-40de-944b-e07fc1f90ae7
```

**Response:**
```json
{
    "message": "Logged out successfully"
}
```

---

### GET /session

**Purpose:** Other services call this to verify tokens

**Request:**
```http
GET /session HTTP/1.1
Authorization: Bearer 7c9e6679-7425-40de-944b-e07fc1f90ae7
```

**Response:**
```json
{
    "success": true,
    "data": {
        "user": {
            "id": 42,
            "username": "alice",
            "email": "alice@example.com"
        }
    }
}
```

**Usage by Other Services:**
```go
// Chat service verifying a token
func AuthMiddleware(authServiceURL string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")
            
            // Call auth service
            req, _ := http.NewRequest("GET", authServiceURL+"/session", nil)
            req.Header.Set("Authorization", token)
            
            resp, _ := http.DefaultClient.Do(req)
            if resp.StatusCode != 200 {
                http.Error(w, "Unauthorized", 401)
                return
            }
            
            // Token valid, continue
            next.ServeHTTP(w, r)
        })
    }
}
```

[Back to Top](#table-of-contents)

---

## Error Handling

### Generic Error Messages (Security)

**Don't reveal too much information:**

**Bad:**
```json
{"error": "User with email alice@example.com not found"}
```
Attacker now knows this email doesn't have an account!

**Good:**
```json
{"error": "invalid email or password"}
```
Attacker doesn't know if email or password is wrong.

### Common Error Responses

**Validation Errors:**
```json
{
    "error": "email is required"
}
{
    "error": "password must be at least 8 characters"
}
```

**Authentication Errors:**
```json
{
    "error": "invalid email or password"
}
{
    "error": "unauthorized"
}
{
    "error": "token expired"
}
```

**Conflict Errors:**
```json
{
    "error": "username already exists"
}
{
    "error": "email already exists"
}
```

### HTTP Status Codes

| Code | Meaning | When to Use |
|------|---------|-------------|
| 200 | OK | Successful request |
| 201 | Created | User registered successfully |
| 400 | Bad Request | Invalid input (validation failed) |
| 401 | Unauthorized | Invalid token or credentials |
| 409 | Conflict | Username/email already exists |
| 500 | Internal Server Error | Database error, unexpected error |

### Error Handling in Code

```go
// Handler layer
func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
    response, err := h.authService.Register(&req)
    if err != nil {
        // Determine appropriate status code
        statusCode := http.StatusInternalServerError
        
        if strings.Contains(err.Error(), "already exists") {
            statusCode = http.StatusConflict  // 409
        } else if strings.Contains(err.Error(), "invalid") || 
                  strings.Contains(err.Error(), "required") {
            statusCode = http.StatusBadRequest  // 400
        }
        
        http.Error(w, err.Error(), statusCode)
        return
    }
    
    // Success
    w.WriteHeader(http.StatusCreated)  // 201
    json.NewEncoder(w).Encode(response)
}
```

[Back to Top](#table-of-contents)

---

## Final Summary

### Key Concepts Recap

1. **Methods vs Functions**: Methods belong to types, have receivers, can access struct fields
2. **Service Architecture**: Handlers → Services → Database (separation of concerns)
3. **Constructors**: `NewXxx()` functions ensure proper initialization
4. **Password Security**: Always hash with bcrypt, never store plain passwords
5. **Sessions/Tokens**: UUID tokens stored in database, expire after 24 hours
6. **Middleware**: Automatic token verification before handler execution

### The Complete Picture

```
                    ┌──────────────────────────────────────┐
                    │      Auth Service (Port 8081)       │
                    └──────────────────────────────────────┘
                                    │
                    ┌───────────────┼───────────────┐
                    │               │               │
            ┌───────▼──────┐ ┌─────▼──────┐ ┌─────▼──────┐
            │   Handlers   │ │  Services  │ │  Database  │
            │              │ │            │ │            │
            │ • Register   │ │ • Validate │ │ • users    │
            │ • Login      │ │ • Hash     │ │ • sessions │
            │ • Logout     │ │ • Generate │ │            │
            │ • Session    │ │   Token    │ │            │
            └──────────────┘ └────────────┘ └────────────┘
                    │               │               │
                    │               │               │
            ┌───────▼──────────────────────────────▼───────┐
            │         Other Services (Chat, Posts, etc)    │
            │    Call /session to verify tokens            │
            └──────────────────────────────────────────────┘
```

### Authentication Flow Summary

**Registration:**
```
User submits form → Validate → Check duplicates → Hash password 
→ Insert user → Generate token → Store session → Return token
```

**Login:**
```
User submits credentials → Get user → Compare password hash 
→ Generate new token → Store session → Return token
```

**Authenticated Request:**
```
Request with token → Middleware extracts token → Validate token 
→ Check expiration → Add userID to context → Call handler
```

---

**Congratulations! You now understand the complete authentication system!**

[Back to Top](#table-of-contents)

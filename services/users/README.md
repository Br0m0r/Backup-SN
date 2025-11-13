# Users Service - Comprehensive Guide

A complete teaching guide for the Users microservice, covering profile management, follow relationships, privacy controls, and the middleware pattern.

## Table of Contents

- [Overview](#overview)
- [The Middleware Pattern - Matryoshka Functions](#the-middleware-pattern---matryoshka-functions)
  - [What is Middleware?](#what-is-middleware)
  - [The Three-Layer Pattern Explained](#the-three-layer-pattern-explained)
  - [Why This Pattern Exists](#why-this-pattern-exists)
  - [Step-by-Step Execution Flow](#step-by-step-execution-flow)
- [Service Architecture](#service-architecture)
- [Privacy System - Public vs Private Profiles](#privacy-system---public-vs-private-profiles)
- [Follow System - Followers and Following](#follow-system---followers-and-following)
- [The Context Pattern - Passing Data Between Middleware and Handlers](#the-context-pattern---passing-data-between-middleware-and-handlers)
- [The Complete Flow - Getting a User Profile](#the-complete-flow---getting-a-user-profile)
- [The Complete Flow - Following a User](#the-complete-flow---following-a-user)
- [Database Operations](#database-operations)
- [HTTP REST Endpoints](#http-rest-endpoints)
- [Error Handling and Privacy](#error-handling-and-privacy)
- [Summary](#summary)

[Back to Top](#table-of-contents)

---

## Overview

The **Users Service** is responsible for managing user profiles, follow relationships, and profile privacy settings.

**Port**: 8082  
**Database**: SQLite (shared with other services)  
**Dependencies**: 
- Auth Service (port 8081) - for token verification
- Posts Service (indirectly) - for displaying user posts on profile

**Core Responsibilities**:
1. **Profile Management** - View and update user profiles (name, bio, avatar, privacy settings)
2. **Follow System** - Follow/unfollow users, manage follow requests (pending/accepted)
3. **Privacy Controls** - Public profiles (anyone can view) vs Private profiles (followers only)
4. **Profile Access** - Determine what data a viewer can see based on relationship and privacy settings
5. **Search** - Find users by username

**Key Concept**: This service handles social relationships and privacy boundaries. Not all users can see everything about other users - it depends on follow status and privacy settings.

[Back to Top](#table-of-contents)

---

## The Middleware Pattern - Matryoshka Functions

This is probably the most confusing pattern in Go web development. Let's break it down step by step.

### What is Middleware?

**Middleware** is code that runs BEFORE your actual handler executes.

Think of it like security checkpoints at an airport:
1. **Check passport** (middleware - verify token)
2. **Scan bags** (middleware - rate limiting)
3. **Board plane** (your actual handler - get profile data)

If you fail at step 1 (no passport), you never reach step 3. Same with middleware - if token is invalid, your handler never executes.

[Back to Top](#table-of-contents)

### The Three-Layer Pattern Explained

Here's the confusing code:

```go
func AuthMiddleware(authServiceURL string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Token verification code here
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

This is called a **"matryoshka"** pattern (like Russian nesting dolls). There are THREE functions nested inside each other.

**Layer 1: Configuration Function**
```go
func AuthMiddleware(authServiceURL string) func(http.Handler) http.Handler
```

- **What it does**: Takes configuration (the URL of auth service)
- **Returns**: A function that can wrap handlers
- **Why**: We need to configure WHERE to verify tokens (auth service URL)
- **When it runs**: When you set up your routes in `main.go`

**Layer 2: The Wrapper Function**
```go
func(next http.Handler) http.Handler
```

- **What it does**: Takes the "next handler" (the actual route handler like GetProfile)
- **Returns**: A new handler that wraps it
- **Why**: This is the standard middleware signature in Go
- **When it runs**: When you call `authMiddleware(http.HandlerFunc(...))`

**Layer 3: The Actual Request Handler**
```go
http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // This code runs for EVERY request
    // 1. Verify token
    // 2. Add user to context
    // 3. Call next handler
})
```

- **What it does**: This is the code that runs for every HTTP request
- **Returns**: Nothing (it either calls `next` or sends an error response)
- **Why**: This is where the actual authentication logic lives
- **When it runs**: Every time a request comes in to a protected route

[Back to Top](#table-of-contents)

### Why This Pattern Exists

**Question**: Why not just have one function?

**Answer**: Because of Go's type system and HTTP handler interface.

**The Problem**: We need to:
1. Configure middleware once (with auth service URL)
2. Apply it to multiple routes
3. Have it execute on every request
4. Keep the standard `http.Handler` interface

**One-Function Approach (doesn't work)**:
```go
// ❌ This doesn't work
func BadMiddleware(authServiceURL string, w http.ResponseWriter, r *http.Request) {
    // Problem: Can't apply this to http.Handler
    // Problem: Where does authServiceURL come from on every request?
}
```

**The Solution**: Three layers

```go
// ✅ Layer 1: Configuration (runs once)
authMiddleware := middleware.AuthMiddleware("http://auth-service:8081")

// ✅ Layer 2: Wrap a handler (runs once per route)
protectedHandler := authMiddleware(http.HandlerFunc(GetProfile))

// ✅ Layer 3: Execute on request (runs on every HTTP request)
// This is automatic when HTTP request comes in
```

[Back to Top](#table-of-contents)

### Step-by-Step Execution Flow

Let's trace what happens when code executes:

**Setup Phase (When Server Starts)**:

```go
// main.go

// Step 1: Create middleware with configuration
authMiddleware := middleware.AuthMiddleware("http://auth-service:8081")
// What happened: Layer 1 executed, returned Layer 2 function
// authMiddleware is now: func(http.Handler) http.Handler

// Step 2: Wrap your handler
mux.Handle("/profile", authMiddleware(http.HandlerFunc(GetProfile)))
// What happened: Layer 2 executed, returned Layer 3 handler
// The route is now: Layer 3 wrapper → GetProfile
```

**Request Phase (When User Makes HTTP Request)**:

```
User → HTTP Request → Layer 3 Handler → GetProfile
```

**Timeline**:

```
0.0s - Request arrives at server
0.1s - Layer 3 executes: Extract Authorization header
0.2s - Layer 3: Make HTTP call to auth service
0.5s - Auth service responds: token valid, userID = 42
0.6s - Layer 3: Add userID to context
0.7s - Layer 3: Call next.ServeHTTP(w, r.WithContext(ctx))
0.8s - GetProfile executes: Access context to get userID
1.0s - GetProfile: Query database for user 42
1.5s - Response sent back to user
```

**Key Point**: Layer 3 runs BEFORE GetProfile. If token is invalid, GetProfile never executes.

[Back to Top](#table-of-contents)

### Visual Diagram

```
┌─────────────────────────────────────────────────────┐
│  Layer 1: AuthMiddleware(authServiceURL)            │
│  Purpose: Store configuration                       │
│  Runs: Once at server startup                       │
│  Returns: Function that takes http.Handler          │
└───────────────┬─────────────────────────────────────┘
                │
                │ returns
                ▼
┌─────────────────────────────────────────────────────┐
│  Layer 2: func(next http.Handler) http.Handler      │
│  Purpose: Wrap the actual route handler            │
│  Runs: Once per route during setup                 │
│  Returns: New handler (Layer 3) that wraps next    │
└───────────────┬─────────────────────────────────────┘
                │
                │ returns
                ▼
┌─────────────────────────────────────────────────────┐
│  Layer 3: http.HandlerFunc(func(w, r))             │
│  Purpose: Execute on every request                 │
│  Runs: Every time HTTP request comes in           │
│  Does:                                             │
│    1. Extract token from Authorization header     │
│    2. Call auth service to verify token           │
│    3. If valid, add userID to context             │
│    4. Call next.ServeHTTP(w, r.WithContext(ctx))  │
│    5. If invalid, return 401 error                │
└───────────────┬─────────────────────────────────────┘
                │
                │ if token valid
                ▼
┌─────────────────────────────────────────────────────┐
│  Your Handler: GetProfile(w, r)                    │
│  Runs: Only if middleware passes                   │
│  Has access to: userID from context                │
└─────────────────────────────────────────────────────┘
```

**Real-World Analogy**:

- **Layer 1**: Hiring a security company (give them office address once)
- **Layer 2**: Assigning guards to specific doors (setup phase)
- **Layer 3**: Guard checking ID every time someone approaches (runtime)
- **Your Handler**: The actual room/resource they're trying to access

[Back to Top](#table-of-contents)

### Why Not Simpler?

**Could we do this?**

```go
func SimpleAuthMiddleware(w http.ResponseWriter, r *http.Request, next http.Handler) {
    // Check token
    next.ServeHTTP(w, r)
}
```

**No, because**:
1. Go's `http.Handler` interface requires `ServeHTTP(w, r)` - only 2 parameters
2. We need configuration (auth service URL) but don't want to hardcode it
3. We need to wrap multiple handlers, not call them directly

**The three-layer pattern solves**:
- ✅ Configuration injection (Layer 1)
- ✅ Handler wrapping (Layer 2)
- ✅ Request execution (Layer 3)
- ✅ Compatible with Go's `http.Handler` interface

[Back to Top](#table-of-contents)

---

## Service Architecture

Like the Auth service, the Users service follows a **three-layer architecture**:

```
HTTP Request → Handlers → Services → Database
```

**Layer 1: Handlers** (`handlers/user.go`)
- Receive HTTP requests
- Extract data from URL/body/headers
- Call service methods
- Send HTTP responses

**Layer 2: Services** (`services/user_service.go`)
- Business logic
- Privacy checks
- Follow rules (can't follow yourself, check if already following)
- Orchestrate multiple database calls

**Layer 3: Database** (`db/queries.go`)
- SQL queries
- Data retrieval and persistence
- No business logic

**Why Separate?**
- **Testability**: Can test business logic without HTTP server
- **Reusability**: Service methods can be called from different handlers
- **Clarity**: Each layer has single responsibility

**Example Flow**:

```go
// Handler Layer
func (h *UserHandlers) FollowUser(w http.ResponseWriter, r *http.Request) {
    userID := middleware.GetUserIDFromContext(r) // From middleware
    targetID := parseBodyForTargetID(r)          // From HTTP body
    
    err := h.userService.FollowUser(userID, targetID) // Call service
    
    if err != nil {
        utils.ErrorResponse(w, err.Error(), 400) // HTTP response
    }
}

// Service Layer
func (s *UserService) FollowUser(followerID, followingID int) error {
    // Business logic
    if followerID == followingID {
        return errors.New("cannot follow yourself")
    }
    
    // Check if already following
    status, _ := db.CheckFollowStatus(s.database, followerID, followingID)
    if status == "accepted" {
        return errors.New("already following")
    }
    
    // Check target privacy
    targetUser, _ := db.GetUserByID(s.database, followingID)
    followStatus := "accepted"
    if !targetUser.IsPublicProfile {
        followStatus = "pending"
    }
    
    // Database call
    return db.CreateFollow(s.database, followerID, followingID, followStatus)
}
```

Notice how:
- **Handler**: Deals with HTTP (parsing, responses)
- **Service**: Deals with business rules (can't follow yourself, privacy checks)
- **Database**: Deals with SQL (insert into follows table)

[Back to Top](#table-of-contents)

---

## Privacy System - Public vs Private Profiles

Every user has a privacy setting: `is_public_profile` (boolean)

**Public Profile** (`is_public_profile = true`):
- Anyone can view full profile
- Anyone can see posts
- Anyone can see followers/following lists
- Follow requests are **automatically accepted**

**Private Profile** (`is_public_profile = false`):
- Only followers can view full profile
- Non-followers see limited data (username, avatar)
- Posts are hidden from non-followers
- Follow requests go to **pending** status

**Database**:

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    is_public_profile BOOLEAN DEFAULT 0,  -- 0 = private, 1 = public
    -- ... other fields
);
```

**Access Control Logic** (`CheckProfileAccess`):

```go
func CheckProfileAccess(db *sql.DB, profileUserID, viewerID int) (bool, error) {
    // Rule 1: You can always view your own profile
    if profileUserID == viewerID {
        return true, nil
    }
    
    // Rule 2: Check if profile is public
    user, _ := GetUserByID(db, profileUserID)
    if user.IsPublicProfile {
        return true, nil  // Public = anyone can view
    }
    
    // Rule 3: For private profiles, must be accepted follower
    followStatus, _ := CheckFollowStatus(db, viewerID, profileUserID)
    if followStatus == "accepted" {
        return true, nil  // Viewer is following (accepted)
    }
    
    // Rule 4: Otherwise, no access
    return false, nil
}
```

**What Viewers See**:

| Viewer Status | Public Profile | Private Profile |
|--------------|----------------|-----------------|
| Owner (self) | Everything | Everything |
| Accepted Follower | Everything | Everything |
| Pending/Not Following | Everything | Limited (username, avatar only) |

**Limited Profile** (for non-followers of private profiles):

```go
func (u *User) PublicProfile() *User {
    return &User{
        ID:              u.ID,
        Username:        u.Username,
        FirstName:       u.FirstName,
        LastName:        u.LastName,
        AvatarPath:      u.AvatarPath,
        Nickname:        u.Nickname,
        AboutMe:         u.AboutMe,
        IsPublicProfile: u.IsPublicProfile,
        CreatedAt:       u.CreatedAt,
        // Email omitted (sensitive)
        // DateOfBirth omitted (sensitive)
    }
}
```

[Back to Top](#table-of-contents)

---

## Follow System - Followers and Following

The follow system has two statuses: **pending** and **accepted**.

**Database**:

```sql
CREATE TABLE follows (
    id INTEGER PRIMARY KEY,
    follower_id INTEGER NOT NULL,      -- User who is following
    following_id INTEGER NOT NULL,     -- User being followed
    status TEXT NOT NULL,              -- 'pending' or 'accepted'
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (follower_id) REFERENCES users(id),
    FOREIGN KEY (following_id) REFERENCES users(id),
    UNIQUE(follower_id, following_id)
);
```

**Follow Status Flow**:

```
User A wants to follow User B

                    ┌─────────────────┐
                    │  Is User B      │
                    │  Public?        │
                    └────────┬────────┘
                             │
              ┌──────────────┴──────────────┐
              │                             │
              ▼                             ▼
       ┌──────────────┐            ┌──────────────┐
       │   YES        │            │    NO        │
       │   Public     │            │   Private    │
       └──────┬───────┘            └──────┬───────┘
              │                           │
              ▼                           ▼
    ┌──────────────────┐        ┌──────────────────┐
    │ INSERT follows   │        │ INSERT follows   │
    │ status='accepted'│        │ status='pending' │
    └──────────────────┘        └──────────────────┘
              │                           │
              ▼                           ▼
    User A is now              User B must approve
    following User B           via /follow/respond
```

**Follow Request Approval**:

Private profiles require manual approval:

```go
func (s *UserService) RespondToFollowRequest(followerID, followingID int, accept bool) error {
    return db.RespondToFollowRequest(s.database, followerID, followingID, accept)
}
```

Database operation:

```go
func RespondToFollowRequest(db *sql.DB, followerID, followingID int, accept bool) error {
    if accept {
        // Accept: Change status from 'pending' to 'accepted'
        query := `UPDATE follows SET status = 'accepted' WHERE follower_id = ? AND following_id = ?`
        _, err := db.Exec(query, followerID, followingID)
        return err
    } else {
        // Reject: Delete the follow request
        query := `DELETE FROM follows WHERE follower_id = ? AND following_id = ?`
        _, err := db.Exec(query, followerID, followingID)
        return err
    }
}
```

**Business Rules** (enforced in service layer):

1. **Can't follow yourself**:
```go
if followerID == followingID {
    return errors.New("cannot follow yourself")
}
```

2. **Can't follow twice**:
```go
status, _ := db.CheckFollowStatus(db, followerID, followingID)
if status == "accepted" || status == "pending" {
    return errors.New("already following or request pending")
}
```

3. **Public profiles auto-accept**:
```go
targetUser, _ := db.GetUserByID(db, followingID)
followStatus := "accepted"
if !targetUser.IsPublicProfile {
    followStatus = "pending"
}
```

[Back to Top](#table-of-contents)

---

## The Context Pattern - Passing Data Between Middleware and Handlers

After middleware verifies the token, how does the handler know WHO the user is?

**Answer**: The `context` package.

### The Problem

Middleware verifies token and gets `userID = 42` from auth service.  
Handler needs `userID` to fetch profile.  
But HTTP handlers only have `(w, r)` parameters - no extra parameters allowed.

### The Solution: Context

**Context** is like a backpack attached to the request. Middleware puts things in, handlers take things out.

**Step 1: Middleware Adds to Context**

```go
// middleware/auth.go

// After verifying token with auth service, we get user info
var authResp struct {
    User struct {
        ID       int    `json:"id"`
        Username string `json:"username"`
    } `json:"user"`
}

// Add userID to context
ctx := context.WithValue(r.Context(), "userID", authResp.User.ID)
ctx = context.WithValue(ctx, "username", authResp.User.Username)

// Pass request WITH new context to next handler
next.ServeHTTP(w, r.WithContext(ctx))
```

**What happened**:
1. Original request `r` had empty context
2. We created new context `ctx` with `userID = 42` inside
3. We created new request `r.WithContext(ctx)` with this context
4. We passed new request to next handler

**Step 2: Handler Extracts from Context**

```go
// handlers/user.go

func (h *UserHandlers) GetCurrentUserProfile(w http.ResponseWriter, r *http.Request) {
    // Extract userID from context
    userID, ok := middleware.GetUserIDFromContext(r)
    if !ok {
        // Context doesn't have userID - should never happen if middleware ran
        utils.ErrorResponse(w, "Unauthorized", 401)
        return
    }
    
    // Now we know WHO is making the request
    user, err := h.userService.GetProfile(userID)
    // ...
}
```

**Helper Function**:

```go
// middleware/auth.go

func GetUserIDFromContext(r *http.Request) (int, bool) {
    userID, ok := r.Context().Value("userID").(int)
    return userID, ok
}
```

**What's happening**:
- `r.Context()` - Get the context from request
- `.Value("userID")` - Get the value with key "userID"
- `.(int)` - Type assertion (convert interface{} to int)
- Returns `(userID, ok)` - value and whether it exists

### Why This Pattern?

**Alternative (doesn't work)**:

```go
// ❌ Can't do this - handlers must match signature
func GetProfile(w http.ResponseWriter, r *http.Request, userID int) {
    // Extra parameter not allowed by http.Handler interface
}
```

**Context solves**:
- ✅ Pass data between middleware and handlers
- ✅ Maintain standard `http.Handler` signature
- ✅ Type-safe with helper functions
- ✅ Multiple middleware can add data

### Context Keys Best Practice

**Problem**: Context keys are `interface{}` - could have collisions

```go
// Bad: String keys could conflict
ctx = context.WithValue(ctx, "userID", 42)
ctx = context.WithValue(ctx, "userID", 99) // Overwrites!
```

**Better**: Use custom type

```go
type contextKey string

const (
    userIDKey contextKey = "userID"
    usernameKey contextKey = "username"
)

ctx = context.WithValue(r.Context(), userIDKey, 42)
```

Our code uses strings for simplicity, but production code should use typed keys.

[Back to Top](#table-of-contents)

---

## The Complete Flow - Getting a User Profile

Let's trace what happens when User A (ID=1) views User B's profile (ID=2).

**HTTP Request**:

```http
GET /users/2/profile HTTP/1.1
Host: localhost:8082
Authorization: Bearer abc123token
```

**Timeline**:

```
0.0s - Request arrives at Users Service (port 8082)

0.1s - CORS Middleware executes
       - Adds CORS headers
       - Calls next handler

0.2s - Logging Middleware executes
       - Logs: "GET /users/2/profile"
       - Calls next handler

0.3s - Auth Middleware executes (Layer 3 from matryoshka)
       - Extracts token: "abc123token"
       - Makes HTTP request to Auth Service (port 8081)

0.5s - Auth Service responds
       - Token valid
       - User ID: 1, Username: "alice"

0.6s - Auth Middleware continues
       - Adds userID=1 to context
       - Calls next handler: next.ServeHTTP(w, r.WithContext(ctx))

0.7s - GetUserProfileByID Handler executes
       - Extracts from URL: profileUserID = 2
       - Extracts from context: viewerID = 1
       - Calls service: GetUserProfile(userID=2, viewerID=1)

0.8s - UserService.GetUserProfile executes
       - Calls: CheckProfileAccess(userID=2, viewerID=1)

0.9s - CheckProfileAccess query
       Query: SELECT is_public_profile FROM users WHERE id = 2
       Result: is_public_profile = false (private profile)
       
1.0s - CheckProfileAccess query
       Query: SELECT status FROM follows WHERE follower_id=1 AND following_id=2
       Result: status = 'accepted'
       Decision: viewerID=1 is following userID=2 → canView = true

1.1s - UserService continues (canView=true)
       - Query: SELECT * FROM users WHERE id = 2
       - Query: SELECT * FROM posts WHERE user_id = 2
       - Query: SELECT * FROM follows WHERE following_id = 2 (followers)
       - Query: SELECT * FROM follows WHERE follower_id = 2 (following)

1.5s - UserService builds response
       - User profile (full data, viewer is follower)
       - Posts: 5 posts
       - Followers: 12 users
       - Following: 8 users
       - CanView: true

1.6s - Handler sends response
       Status: 200 OK
       Body: JSON with profile data

1.7s - Response reaches User A's browser
```

**Code Flow**:

```go
// 1. main.go - Route setup
mux.Handle("/users/", authMiddleware(http.HandlerFunc(userHandlers.GetUserProfileByID)))

// 2. Middleware - Auth check
func (Layer3) ServeHTTP(w, r) {
    token := extractToken(r)
    userInfo := callAuthService(token) // userID = 1
    ctx := context.WithValue(r.Context(), "userID", userInfo.ID)
    next.ServeHTTP(w, r.WithContext(ctx))
}

// 3. Handler - Extract IDs
func (h *UserHandlers) GetUserProfileByID(w, r) {
    profileUserID := extractFromURL(r) // 2
    viewerID := middleware.GetUserIDFromContext(r) // 1
    profile, err := h.userService.GetUserProfile(profileUserID, viewerID)
    utils.SuccessResponse(w, profile)
}

// 4. Service - Check access and fetch data
func (s *UserService) GetUserProfile(userID, viewerID int) (*ProfileResponse, error) {
    canView := db.CheckProfileAccess(userID, viewerID)
    
    if !canView {
        return &ProfileResponse{
            User: user.PublicProfile(), // Limited data
            CanView: false,
        }
    }
    
    user := db.GetUserByID(userID)
    posts := db.GetUserPosts(userID)
    followers := db.GetFollowers(userID)
    following := db.GetFollowing(userID)
    
    return &ProfileResponse{
        User: user,
        Posts: posts,
        Followers: followers,
        Following: following,
        CanView: true,
    }
}
```

**Key Decisions in Flow**:

| Check | Condition | Result |
|-------|-----------|--------|
| Is viewer the owner? | viewerID (1) == userID (2) | No |
| Is profile public? | is_public_profile = false | No |
| Is viewer following? | status = 'accepted' | Yes ✓ |
| **Final Decision** | | **Full access granted** |

[Back to Top](#table-of-contents)

---

## The Complete Flow - Following a User

Let's trace what happens when User A (ID=1) follows User B (ID=2), where User B has a private profile.

**HTTP Request**:

```http
POST /follow HTTP/1.1
Host: localhost:8082
Authorization: Bearer abc123token
Content-Type: application/json

{
  "user_id": 2
}
```

**Timeline**:

```
0.0s - Request arrives at Users Service

0.1s - Middleware chain executes
       - CORS middleware
       - Logging middleware
       - Auth middleware (verifies token, adds userID=1 to context)
       - Rate limiter (checks if user exceeded follow requests)

0.2s - FollowUser Handler executes
       - Extracts followerID from context: 1
       - Extracts followingID from body: 2
       - Calls: h.userService.FollowUser(1, 2)

0.3s - UserService.FollowUser executes
       Business Rule Check #1: Can't follow yourself
       if followerID (1) == followingID (2) → false ✓

0.4s - Business Rule Check #2: Check if already following
       Query: SELECT status FROM follows WHERE follower_id=1 AND following_id=2
       Result: No rows (not following)

0.5s - Business Rule Check #3: Get target user's privacy setting
       Query: SELECT * FROM users WHERE id = 2
       Result: User B, is_public_profile = false (private)

0.6s - Determine follow status
       Logic: Profile is private → followStatus = "pending"
       (If public, would be "accepted" immediately)

0.7s - Create follow relationship
       Query: INSERT INTO follows (follower_id, following_id, status, created_at)
              VALUES (1, 2, 'pending', datetime('now'))
       Result: Follow request created with status='pending'

0.8s - Response sent
       Status: 200 OK
       Body: {"message": "Follow request sent"}

0.9s - User A sees: "Follow request sent to User B"
       User B will see: "User A wants to follow you" in pending requests
```

**Code Flow**:

```go
// 1. Handler - Extract data
func (h *UserHandlers) FollowUser(w http.ResponseWriter, r *http.Request) {
    followerID, _ := middleware.GetUserIDFromContext(r) // 1
    
    var req models.FollowRequest
    json.NewDecoder(r.Body).Decode(&req)
    followingID := req.UserID // 2
    
    err := h.userService.FollowUser(followerID, followingID)
    
    if err != nil {
        utils.ErrorResponse(w, err.Error(), 400)
        return
    }
    
    utils.SuccessResponse(w, map[string]string{
        "message": "Follow request sent",
    })
}

// 2. Service - Business logic
func (s *UserService) FollowUser(followerID, followingID int) error {
    // Check #1: Can't follow yourself
    if followerID == followingID {
        return errors.New("cannot follow yourself")
    }
    
    // Check #2: Already following?
    status, _ := db.CheckFollowStatus(s.database, followerID, followingID)
    if status == "accepted" || status == "pending" {
        return errors.New("already following or request pending")
    }
    
    // Check #3: Target user's privacy
    targetUser, err := db.GetUserByID(s.database, followingID)
    if err != nil {
        return errors.New("user not found")
    }
    
    // Determine status based on privacy
    followStatus := "accepted"
    if !targetUser.IsPublicProfile {
        followStatus = "pending"
    }
    
    // Create follow relationship
    return db.CreateFollow(s.database, followerID, followingID, followStatus)
}

// 3. Database - Insert
func CreateFollow(db *sql.DB, followerID, followingID int, status string) error {
    query := `
        INSERT INTO follows (follower_id, following_id, status, created_at)
        VALUES (?, ?, ?, datetime('now'))
    `
    _, err := db.Exec(query, followerID, followingID, status)
    return err
}
```

**Follow Status Scenarios**:

| Target Profile | Initial Status | What Happens |
|---------------|----------------|--------------|
| Public | `accepted` | Immediately following, can see all content |
| Private | `pending` | Must wait for approval, can't see content yet |

**User B's Approval Flow** (later):

```http
POST /follow/respond HTTP/1.1
Authorization: Bearer <User B's token>

{
  "follower_id": 1,
  "accept": true
}
```

**What happens**:
1. Extract User B's ID from token (ID=2)
2. Verify: User B (2) is the one being followed (following_id=2)
3. Update: `UPDATE follows SET status='accepted' WHERE follower_id=1 AND following_id=2`
4. Now User A can see User B's profile and posts

[Back to Top](#table-of-contents)

---

## Database Operations

All SQL queries are in `db/queries.go`.

### User Queries

**GetUserByID** - Retrieve user profile:

```go
func GetUserByID(db *sql.DB, userID int) (*models.User, error) {
    query := `
        SELECT id, username, email, first_name, last_name, date_of_birth, 
               avatar_path, nickname, about_me, is_public_profile, created_at
        FROM users 
        WHERE id = ?
    `
    // Scan and return user
}
```

**UpdateUserProfile** - Update profile fields:

```go
func UpdateUserProfile(db *sql.DB, userID int, req *models.UpdateProfileRequest) error {
    query := `
        UPDATE users 
        SET first_name = COALESCE(?, first_name),
            last_name = COALESCE(?, last_name),
            date_of_birth = COALESCE(?, date_of_birth),
            nickname = COALESCE(?, nickname),
            about_me = COALESCE(?, about_me),
            is_public_profile = COALESCE(?, is_public_profile)
        WHERE id = ?
    `
    // COALESCE: Use new value if provided, otherwise keep existing
}
```

**SearchUsers** - Find users by username:

```go
func SearchUsers(db *sql.DB, searchTerm string) ([]*models.User, error) {
    query := `
        SELECT id, username, email, first_name, last_name, avatar_path, 
               nickname, about_me, is_public_profile, created_at
        FROM users 
        WHERE username LIKE ? OR first_name LIKE ? OR last_name LIKE ?
        LIMIT 50
    `
    // Search with LIKE '%term%'
}
```

### Follow Queries

**CreateFollow** - Create follow relationship:

```go
func CreateFollow(db *sql.DB, followerID, followingID int, status string) error {
    query := `
        INSERT INTO follows (follower_id, following_id, status, created_at)
        VALUES (?, ?, ?, datetime('now'))
    `
    // status = 'pending' or 'accepted'
}
```

**DeleteFollow** - Unfollow user:

```go
func DeleteFollow(db *sql.DB, followerID, followingID int) error {
    query := `
        DELETE FROM follows 
        WHERE follower_id = ? AND following_id = ?
    `
}
```

**CheckFollowStatus** - Get relationship status:

```go
func CheckFollowStatus(db *sql.DB, followerID, followingID int) (string, error) {
    query := `
        SELECT status 
        FROM follows 
        WHERE follower_id = ? AND following_id = ?
    `
    // Returns: "" (not following), "pending", or "accepted"
}
```

**GetFollowers** - Get all followers (accepted only):

```go
func GetFollowers(db *sql.DB, userID int) ([]*models.User, error) {
    query := `
        SELECT u.* 
        FROM users u
        INNER JOIN follows f ON u.id = f.follower_id
        WHERE f.following_id = ? AND f.status = 'accepted'
        ORDER BY f.created_at DESC
    `
    // Returns users who are following this user
}
```

**GetFollowing** - Get all following (accepted only):

```go
func GetFollowing(db *sql.DB, userID int) ([]*models.User, error) {
    query := `
        SELECT u.* 
        FROM users u
        INNER JOIN follows f ON u.id = f.following_id
        WHERE f.follower_id = ? AND f.status = 'accepted'
        ORDER BY f.created_at DESC
    `
    // Returns users that this user is following
}
```

**GetPendingFollowRequests** - Get users waiting for approval:

```go
func GetPendingFollowRequests(db *sql.DB, userID int) ([]*models.User, error) {
    query := `
        SELECT u.* 
        FROM users u
        INNER JOIN follows f ON u.id = f.follower_id
        WHERE f.following_id = ? AND f.status = 'pending'
        ORDER BY f.created_at DESC
    `
    // Returns users who want to follow (for User B to approve)
}
```

**RespondToFollowRequest** - Accept or reject:

```go
func RespondToFollowRequest(db *sql.DB, followerID, followingID int, accept bool) error {
    if accept {
        // Accept: Change to 'accepted'
        query := `
            UPDATE follows 
            SET status = 'accepted' 
            WHERE follower_id = ? AND following_id = ?
        `
    } else {
        // Reject: Delete request
        query := `
            DELETE FROM follows 
            WHERE follower_id = ? AND following_id = ?
        `
    }
}
```

### Privacy Queries

**CheckProfileAccess** - Determine if viewer can see profile:

```go
func CheckProfileAccess(db *sql.DB, profileUserID, viewerID int) (bool, error) {
    // Rule 1: Own profile
    if profileUserID == viewerID {
        return true, nil
    }
    
    // Rule 2: Public profile
    user, _ := GetUserByID(db, profileUserID)
    if user.IsPublicProfile {
        return true, nil
    }
    
    // Rule 3: Accepted follower
    status, _ := CheckFollowStatus(db, viewerID, profileUserID)
    if status == "accepted" {
        return true, nil
    }
    
    // Rule 4: No access
    return false, nil
}
```

[Back to Top](#table-of-contents)

---

## HTTP REST Endpoints

All endpoints require authentication (except `/health`).

### Profile Endpoints

**GET /profile** - Get current user's profile

```http
GET /profile HTTP/1.1
Authorization: Bearer <token>

Response 200 OK:
{
  "user": {
    "id": 1,
    "username": "alice",
    "email": "alice@example.com",
    "first_name": "Alice",
    "last_name": "Smith",
    "avatar_path": "/uploads/alice.jpg",
    "nickname": "Ali",
    "about_me": "Software developer",
    "is_public_profile": true,
    "created_at": "2024-01-15T10:00:00Z"
  }
}
```

**PUT /profile** - Update current user's profile

```http
PUT /profile HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "first_name": "Alice",
  "last_name": "Smith",
  "nickname": "Ali",
  "about_me": "Full-stack developer",
  "is_public_profile": false
}

Response 200 OK:
{
  "user": { /* updated profile */ }
}
```

**GET /users/:id/profile** - Get another user's profile

```http
GET /users/2/profile HTTP/1.1
Authorization: Bearer <token>

Response 200 OK (if viewer has access):
{
  "user": { /* profile data */ },
  "posts": [ /* user's posts */ ],
  "followers": [ /* follower list */ ],
  "following": [ /* following list */ ],
  "follower_count": 12,
  "following_count": 8,
  "post_count": 5,
  "can_view": true
}

Response 200 OK (if viewer doesn't have access):
{
  "user": { /* limited: username, avatar only */ },
  "posts": [],
  "followers": [],
  "following": [],
  "follower_count": 0,
  "following_count": 0,
  "post_count": 0,
  "can_view": false
}
```

### Follow Endpoints

**POST /follow** - Follow a user

```http
POST /follow HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "user_id": 2
}

Response 200 OK:
{
  "message": "Follow request sent"
}

Response 400 Bad Request (if error):
{
  "error": "already following or request pending"
}
```

**DELETE /follow** - Unfollow a user

```http
DELETE /follow HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "user_id": 2
}

Response 200 OK:
{
  "message": "Unfollowed successfully"
}
```

**GET /followers** - Get followers of current user

```http
GET /followers HTTP/1.1
Authorization: Bearer <token>

Response 200 OK:
{
  "followers": [
    {
      "id": 3,
      "username": "bob",
      "first_name": "Bob",
      "avatar_path": "/uploads/bob.jpg"
    },
    // ... more followers
  ]
}
```

**GET /following** - Get users current user is following

```http
GET /following HTTP/1.1
Authorization: Bearer <token>

Response 200 OK:
{
  "following": [ /* list of users */ ]
}
```

**GET /follow/status/:id** - Check follow status with specific user

```http
GET /follow/status/2 HTTP/1.1
Authorization: Bearer <token>

Response 200 OK:
{
  "status": "accepted"  // or "pending" or "" (not following)
}
```

**GET /follow/requests** - Get pending follow requests

```http
GET /follow/requests HTTP/1.1
Authorization: Bearer <token>

Response 200 OK:
{
  "requests": [
    {
      "id": 5,
      "username": "charlie",
      "first_name": "Charlie",
      "avatar_path": "/uploads/charlie.jpg",
      "created_at": "2024-10-16T14:30:00Z"
    }
  ]
}
```

**POST /follow/respond** - Accept or reject follow request

```http
POST /follow/respond HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "follower_id": 5,
  "accept": true
}

Response 200 OK:
{
  "message": "Follow request accepted"
}
```

### Search Endpoint

**GET /search?q=alice** - Search for users

```http
GET /search?q=alice HTTP/1.1
Authorization: Bearer <token>

Response 200 OK:
{
  "users": [
    {
      "id": 1,
      "username": "alice",
      "first_name": "Alice",
      "last_name": "Smith",
      "avatar_path": "/uploads/alice.jpg"
    },
    {
      "id": 12,
      "username": "alice_jones",
      "first_name": "Alice",
      "last_name": "Jones",
      "avatar_path": null
    }
  ]
}
```

[Back to Top](#table-of-contents)

---

## Error Handling and Privacy

### Security Considerations

**Don't leak information**:

```go
// ❌ Bad - Reveals whether user exists
if user not found {
    return error: "user not found"
}
if cannot access {
    return error: "you don't have access to this profile"
}

// ✅ Good - Same response for both cases
if user not found OR cannot access {
    return limited profile with can_view=false
}
```

**Hide sensitive data**:

```go
// Always use PublicProfile() for non-owners
if userID != viewerID {
    profile.User = user.PublicProfile() // Hides email, date of birth
}
```

**Rate limiting**:

Follow requests are rate-limited to prevent spam:

```go
// main.go
rateLimiter := middleware.NewRateLimiter()
mux.Handle("/follow", authMiddleware(rateLimiter.RateLimit(handler)))
```

### Error Response Format

All errors use consistent format:

```go
func ErrorResponse(w http.ResponseWriter, message string, statusCode int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(map[string]string{
        "error": message,
    })
}
```

**Common Status Codes**:
- `200 OK` - Success
- `400 Bad Request` - Invalid input (e.g., can't follow yourself)
- `401 Unauthorized` - No token or invalid token
- `404 Not Found` - Resource doesn't exist
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

[Back to Top](#table-of-contents)

---

## Summary

**Users Service Core Concepts**:

1. **Middleware Pattern** (Matryoshka)
   - Three nested functions: Configuration → Wrapper → Request Handler
   - Layer 1: Takes config (auth service URL), returns wrapper
   - Layer 2: Takes next handler, returns request handler
   - Layer 3: Executes on every request (verify token, add to context)

2. **Context Pattern**
   - Middleware adds data: `context.WithValue(r.Context(), "userID", 42)`
   - Handlers extract data: `userID := r.Context().Value("userID").(int)`
   - Enables passing data without changing function signatures

3. **Privacy System**
   - Public profiles: Anyone can view everything
   - Private profiles: Only accepted followers can view
   - Limited profiles: Non-followers see username/avatar only

4. **Follow System**
   - Public profiles: Follow requests auto-accepted
   - Private profiles: Follow requests go to pending, require approval
   - Business rules: Can't follow yourself, can't follow twice

5. **Service Architecture**
   - Handlers: HTTP request/response
   - Services: Business logic and privacy checks
   - Database: SQL queries

**Why These Patterns?**

- **Middleware**: Reusable authentication for all routes
- **Context**: Pass data between middleware and handlers
- **Service Layer**: Separate business logic from HTTP concerns
- **Privacy Checks**: Protect user data based on relationships

**Key Takeaway**: The Users service manages social relationships and privacy boundaries. The middleware pattern ensures every request is authenticated, the context pattern passes user identity, and the privacy system determines what each user can see based on follow status and profile settings.

[Back to Top](#table-of-contents)

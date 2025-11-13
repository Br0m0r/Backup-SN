# Posts Service - Comprehensive Guide

A complete teaching guide for the Posts microservice, covering post creation, privacy levels, feed generation, comments, and access control.

## Table of Contents

- [Overview](#overview)
- [Privacy Levels - The Three Types Explained](#privacy-levels---the-three-types-explained)
  - [Public Posts](#public-posts)
  - [Private Posts](#private-posts)
  - [Almost Private Posts](#almost-private-posts)
  - [Privacy Level Comparison](#privacy-level-comparison)
- [Service Architecture](#service-architecture)
- [The Feed Algorithm - What You See](#the-feed-algorithm---what-you-see)
- [Access Control - Who Can See What](#access-control---who-can-see-what)
- [Post Viewers - The "Almost Private" Whitelist](#post-viewers---the-almost-private-whitelist)
- [The Complete Flow - Creating a Post](#the-complete-flow---creating-a-post)
- [The Complete Flow - Viewing the Feed](#the-complete-flow---viewing-the-feed)
- [The Complete Flow - Creating a Comment](#the-complete-flow---creating-a-comment)
- [Database Operations](#database-operations)
- [HTTP REST Endpoints](#http-rest-endpoints)
- [Error Handling and Security](#error-handling-and-security)
- [Summary](#summary)

[Back to Top](#table-of-contents)

---

## Overview

The **Posts Service** is responsible for managing user posts, comments, privacy settings, and feed generation.

**Port**: 8083  
**Database**: SQLite (shared with other services)  
**Dependencies**: 
- Auth Service (port 8081) - for token verification
- Users Service (indirectly) - for follow relationships

**Core Responsibilities**:
1. **Post Management** - Create, read, update, delete posts
2. **Privacy Control** - Three privacy levels (public, private, almost_private)
3. **Feed Generation** - Smart algorithm to show relevant posts
4. **Comment System** - Comment on posts with access control
5. **Access Control** - Determine who can view specific posts

**Key Concept**: This service implements a sophisticated privacy system with THREE levels. Not all posts are visible to everyone. The "almost_private" level is unique - it allows you to select specific users who can see your post.

[Back to Top](#table-of-contents)

---

## Privacy Levels - The Three Types Explained

Every post has a `privacy_level` field with one of three values:

### Public Posts

**privacy_level = "public"**

- **Visible to**: Everyone (all users, even if not following)
- **Use case**: Announcements, general content, public updates
- **Feed behavior**: Shows up in everyone's feed
- **Example**: "Check out my new blog post about Go programming!"

**Database**:
```sql
INSERT INTO posts (user_id, content, privacy_level)
VALUES (1, 'Hello world!', 'public')
```

**Who can see**:
- ✅ User's followers
- ✅ Non-followers
- ✅ Anyone logged into the platform

### Private Posts

**privacy_level = "private"**

- **Visible to**: Only users who are **following** the author (with status='accepted')
- **Use case**: Personal updates, content for friends/followers only
- **Feed behavior**: Only appears in followers' feeds
- **Example**: "Had a great day with friends today" (only for followers)

**Database**:
```sql
INSERT INTO posts (user_id, content, privacy_level)
VALUES (1, 'Personal update', 'private')
```

**Who can see**:
- ✅ User's accepted followers
- ✅ Post author (always sees own posts)
- ❌ Non-followers (access denied)

**Access Check Logic**:
```go
// For private posts
if post.PrivacyLevel == "private" {
    // Check if viewer is following the author
    query := `
        SELECT 1 FROM follows 
        WHERE follower_id = ? AND following_id = ? AND status = 'accepted'
    `
    // If no result → access denied
}
```

### Almost Private Posts

**privacy_level = "almost_private"**

This is the most interesting and complex privacy level.

- **Visible to**: ONLY specific users selected by the author (whitelist)
- **Use case**: Sharing with a small group, targeted content, selective sharing
- **Feed behavior**: Only appears in selected users' feeds
- **Example**: "Planning surprise party for Alice" (visible to: Bob, Charlie, David only)

**How it works**:
1. Author creates post with `privacy_level = "almost_private"`
2. Author provides list of user IDs: `[5, 8, 12]` (the "viewers")
3. System stores these in `post_viewers` table
4. Only users 5, 8, and 12 (plus the author) can see this post

**Database**:
```sql
-- Insert the post
INSERT INTO posts (user_id, content, privacy_level)
VALUES (1, 'Secret content', 'almost_private')

-- Add viewers (whitelist)
INSERT INTO post_viewers (post_id, user_id) VALUES (123, 5)
INSERT INTO post_viewers (post_id, user_id) VALUES (123, 8)
INSERT INTO post_viewers (post_id, user_id) VALUES (123, 12)
```

**Who can see**:
- ✅ Post author (always sees own posts)
- ✅ User ID 5 (in whitelist)
- ✅ User ID 8 (in whitelist)
- ✅ User ID 12 (in whitelist)
- ❌ Everyone else (even followers!)

**Access Check Logic**:
```go
// For almost_private posts
if post.PrivacyLevel == "almost_private" {
    // Check if viewer is in the whitelist
    query := `
        SELECT 1 FROM post_viewers 
        WHERE post_id = ? AND user_id = ?
    `
    // If no result → access denied
}
```

### Privacy Level Comparison

| Privacy Level | Who Can See | Use Case | Follow Required? | Whitelist? |
|--------------|-------------|----------|------------------|------------|
| **public** | Everyone | General content | No | No |
| **private** | Followers only | Personal updates | Yes (accepted) | No |
| **almost_private** | Selected users only | Targeted sharing | No | Yes |

**Visual Diagram**:

```
┌─────────────────────────────────────────────────────────────┐
│                         ALL USERS                           │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                PUBLIC POSTS VISIBLE                    │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │        FOLLOWERS (of Author)                    │  │  │
│  │  │  ┌───────────────────────────────────────────┐  │  │  │
│  │  │  │   PRIVATE POSTS VISIBLE                   │  │  │  │
│  │  │  │                                           │  │  │  │
│  │  │  └───────────────────────────────────────────┘  │  │  │
│  │  │                                                 │  │  │
│  │  │  ┌───────────────────────────────────────────┐  │  │  │
│  │  │  │   SELECTED USERS (Whitelist)             │  │  │  │
│  │  │  │   ALMOST_PRIVATE POSTS VISIBLE           │  │  │  │
│  │  │  │   (Even if not following!)               │  │  │  │
│  │  │  └───────────────────────────────────────────┘  │  │  │
│  │  │                                                 │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

**Key Insight**: 
- `public` → Biggest audience (everyone)
- `private` → Medium audience (followers)
- `almost_private` → Smallest audience (hand-picked users)

[Back to Top](#table-of-contents)

---

## Service Architecture

The Posts service follows the standard **three-layer architecture**:

```
HTTP Request → Handlers → Services → Database
```

**Layer 1: Handlers** (`handlers/post.go`)
- Parse HTTP requests
- Extract userID from context (added by auth middleware)
- Call service methods
- Send JSON responses

**Layer 2: Services** (`services/post_service.go`)
- Business logic
- Validation (content required, valid privacy level)
- Access control (ownership checks)
- Privacy enforcement

**Layer 3: Database** (`db/queries.go`)
- SQL queries
- Data retrieval and persistence
- Complex feed queries

**Example Flow**:

```go
// Layer 1: Handler
func (h *PostHandlers) CreatePost(w http.ResponseWriter, r *http.Request) {
    userID := middleware.GetUserIDFromContext(r) // From auth middleware
    
    var req models.CreatePostRequest
    json.NewDecoder(r.Body).Decode(&req) // Parse JSON body
    
    post, err := h.postService.CreatePost(&req, userID) // Call service
    
    utils.SuccessResponse(w, map[string]interface{}{"post": post}) // JSON response
}

// Layer 2: Service
func (s *PostService) CreatePost(req *CreatePostRequest, userID int) (*Post, error) {
    // Validation
    if req.PrivacyLevel != "public" && req.PrivacyLevel != "private" && req.PrivacyLevel != "almost_private" {
        return nil, errors.New("invalid privacy level")
    }
    
    if req.Content == "" {
        return nil, errors.New("content is required")
    }
    
    // Create post
    post := &Post{
        UserID:       userID,
        Content:      req.Content,
        PrivacyLevel: req.PrivacyLevel,
        CreatedAt:    time.Now(),
    }
    
    // Database call
    db.CreatePost(s.database, post)
    
    // If almost_private, add viewers
    if req.PrivacyLevel == "almost_private" && len(req.Viewers) > 0 {
        db.AddPostViewers(s.database, post.ID, req.Viewers)
    }
    
    return post, nil
}

// Layer 3: Database
func CreatePost(db *sql.DB, post *Post) error {
    query := `
        INSERT INTO posts (user_id, content, privacy_level, created_at)
        VALUES (?, ?, ?, ?)
    `
    result, _ := db.Exec(query, post.UserID, post.Content, post.PrivacyLevel, post.CreatedAt)
    
    id, _ := result.LastInsertId()
    post.ID = int(id)
    return nil
}
```

[Back to Top](#table-of-contents)

---

## The Feed Algorithm - What You See

When you open the app and view your feed, what posts do you see? This is determined by the **feed algorithm**.

**GetFeedPosts** - The SQL Query:

```sql
SELECT DISTINCT p.id, p.user_id, p.content, p.privacy_level, p.created_at
FROM posts p
LEFT JOIN follows f ON p.user_id = f.following_id 
                    AND f.follower_id = ? 
                    AND f.status = 'accepted'
LEFT JOIN post_viewers pv ON p.id = pv.post_id 
                          AND pv.user_id = ?
WHERE 
    p.privacy_level = 'public' OR              -- Rule 1: All public posts
    p.user_id = ? OR                           -- Rule 2: Your own posts
    (p.privacy_level = 'almost_private'        -- Rule 3: Almost private posts where you're in whitelist
     AND pv.user_id IS NOT NULL) OR
    (p.privacy_level = 'private'               -- Rule 4: Private posts from users you follow
     AND f.follower_id IS NOT NULL)
ORDER BY p.created_at DESC
```

**Feed Rules** (in plain English):

You see a post if ANY of these conditions are true:

1. **Public Posts**: Post has `privacy_level = 'public'` → Everyone sees it
2. **Your Own Posts**: Post's `user_id` equals your ID → You always see your own posts
3. **Almost Private Whitelist**: Post is `almost_private` AND you're in the `post_viewers` table for that post
4. **Private from Following**: Post is `private` AND you're following the author (status='accepted')

**Visual Example**:

```
User A (ID=1) opens their feed. What do they see?

Database Posts:
┌────┬─────────┬─────────────────┬───────────────────┬─────────┐
│ ID │ UserID  │ Content         │ Privacy Level     │ Viewer  │
├────┼─────────┼─────────────────┼───────────────────┼─────────┤
│ 10 │ 1       │ "My post"       │ private           │ -       │  ✅ Rule 2: Own post
│ 11 │ 2       │ "Public update" │ public            │ -       │  ✅ Rule 1: Public
│ 12 │ 3       │ "Private post"  │ private           │ -       │  ❌ Not following User 3
│ 13 │ 4       │ "Secret"        │ almost_private    │ [1,5]   │  ✅ Rule 3: User 1 in whitelist
│ 14 │ 5       │ "Followers only"│ private           │ -       │  ✅ Rule 4: Following User 5
│ 15 │ 6       │ "Another secret"│ almost_private    │ [7,8]   │  ❌ User 1 NOT in whitelist
└────┴─────────┴─────────────────┴───────────────────┴─────────┘

User A's Feed (posts they see):
- Post 10: "My post" (own post)
- Post 11: "Public update" (public)
- Post 13: "Secret" (in whitelist)
- Post 14: "Followers only" (following User 5)

Total: 4 posts in feed
```

**Feed vs Individual Post Access**:

| Action | Behavior |
|--------|----------|
| **GET /posts** (feed) | Returns all posts user CAN see based on rules |
| **GET /posts/123** | Returns specific post IF user has access, else 403 Forbidden |

The feed query is optimized with `LEFT JOIN` to check relationships efficiently. Without proper indexing, this query could be slow for large databases.

[Back to Top](#table-of-contents)

---

## Access Control - Who Can See What

Before showing any post, the system checks **CheckPostAccess**.

**The Access Control Function**:

```go
func CheckPostAccess(db *sql.DB, postID, userID int) (bool, error) {
    query := `
        SELECT 
            CASE 
                WHEN p.user_id = ? THEN 1                    -- Rule 1: Owner
                WHEN p.privacy_level = 'public' THEN 1       -- Rule 2: Public
                WHEN p.privacy_level = 'private' AND EXISTS (
                    SELECT 1 FROM follows 
                    WHERE follower_id = ? 
                      AND following_id = p.user_id 
                      AND status = 'accepted'
                ) THEN 1                                     -- Rule 3: Private + Following
                WHEN p.privacy_level = 'almost_private' AND EXISTS (
                    SELECT 1 FROM post_viewers 
                    WHERE post_id = ? 
                      AND user_id = ?
                ) THEN 1                                     -- Rule 4: Almost Private + Whitelist
                ELSE 0
            END as has_access
        FROM posts p
        WHERE p.id = ?
    `
    
    var hasAccess int
    db.QueryRow(query, userID, userID, postID, userID, postID).Scan(&hasAccess)
    
    return hasAccess == 1, nil
}
```

**Decision Tree**:

```
Is userID the post author?
├─ YES → ✅ Access Granted
└─ NO
   └─ Is privacy level "public"?
      ├─ YES → ✅ Access Granted
      └─ NO
         └─ Is privacy level "private"?
            ├─ YES
            │  └─ Is userID following author (accepted)?
            │     ├─ YES → ✅ Access Granted
            │     └─ NO → ❌ Access Denied
            └─ NO (must be "almost_private")
               └─ Is userID in post_viewers table?
                  ├─ YES → ✅ Access Granted
                  └─ NO → ❌ Access Denied
```

**Where This is Used**:

1. **GetPost** - Viewing a single post
```go
func (s *PostService) GetPost(postID, userID int) (*Post, error) {
    hasAccess, _ := db.CheckPostAccess(s.database, postID, userID)
    if !hasAccess {
        return nil, errors.New("access denied")
    }
    return db.GetPostByID(s.database, postID)
}
```

2. **CreateComment** - Commenting on a post
```go
func (s *PostService) CreateComment(req *CreateCommentRequest, userID int) (*Comment, error) {
    hasAccess, _ := db.CheckPostAccess(s.database, req.PostID, userID)
    if !hasAccess {
        return nil, errors.New("access denied: cannot comment on this post")
    }
    // Create comment...
}
```

3. **GetComments** - Viewing comments on a post
```go
func (s *PostService) GetComments(postID, userID int) ([]*Comment, error) {
    hasAccess, _ := db.CheckPostAccess(s.database, postID, userID)
    if !hasAccess {
        return nil, errors.New("access denied: cannot view comments on this post")
    }
    return db.GetCommentsByPostID(s.database, postID)
}
```

**Key Point**: Access control is checked BEFORE every operation. If you can't see the post, you can't comment on it or see its comments.

[Back to Top](#table-of-contents)

---

## Post Viewers - The "Almost Private" Whitelist

The `post_viewers` table is the key to the "almost_private" privacy level.

**Database Schema**:

```sql
CREATE TABLE post_viewers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(post_id, user_id)
);
```

**How It Works**:

1. **Creating an "almost_private" post**:

```http
POST /posts
Authorization: Bearer <token>

{
  "content": "Secret project details",
  "privacy_level": "almost_private",
  "viewers": [5, 8, 12]  ← Selected user IDs
}
```

2. **Backend processes this**:

```go
// Step 1: Create the post
post := &Post{
    UserID:       1,  // From token
    Content:      "Secret project details",
    PrivacyLevel: "almost_private",
}
db.CreatePost(database, post)  // Returns post.ID = 123

// Step 2: Add viewers to whitelist
db.AddPostViewers(database, 123, [5, 8, 12])

// Step 3: Database inserts
// INSERT INTO post_viewers (post_id, user_id) VALUES (123, 5)
// INSERT INTO post_viewers (post_id, user_id) VALUES (123, 8)
// INSERT INTO post_viewers (post_id, user_id) VALUES (123, 12)
```

3. **Resulting database state**:

```
posts table:
┌────┬─────────┬──────────────────────┬───────────────────┐
│ id │ user_id │ content              │ privacy_level     │
├────┼─────────┼──────────────────────┼───────────────────┤
│123 │ 1       │ Secret project...    │ almost_private    │
└────┴─────────┴──────────────────────┴───────────────────┘

post_viewers table:
┌────┬─────────┬─────────┐
│ id │ post_id │ user_id │
├────┼─────────┼─────────┤
│ 1  │ 123     │ 5       │  ← User 5 can see post 123
│ 2  │ 123     │ 8       │  ← User 8 can see post 123
│ 3  │ 123     │ 12      │  ← User 12 can see post 123
└────┴─────────┴─────────┘
```

**Updating Viewers**:

When you update a post and change the viewers list:

```go
func AddPostViewers(db *sql.DB, postID int, userIDs []int) error {
    // Step 1: Clear existing viewers
    db.Exec(`DELETE FROM post_viewers WHERE post_id = ?`, postID)
    
    // Step 2: Insert new viewers
    for _, userID := range userIDs {
        db.Exec(`INSERT INTO post_viewers (post_id, user_id) VALUES (?, ?)`, postID, userID)
    }
}
```

This ensures:
- If you remove someone from viewers list, they lose access immediately
- If you add someone, they gain access immediately
- Changing from "almost_private" to "public" or "private" clears all viewers

**Cascade Delete**:

```sql
FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
```

When a post is deleted, all entries in `post_viewers` for that post are automatically deleted. This prevents orphaned records.

[Back to Top](#table-of-contents)

---

## The Complete Flow - Creating a Post

Let's trace what happens when User A (ID=1) creates an "almost_private" post.

**HTTP Request**:

```http
POST /posts HTTP/1.1
Host: localhost:8083
Authorization: Bearer abc123token
Content-Type: application/json

{
  "title": "Project Update",
  "content": "The new feature is ready for review",
  "privacy_level": "almost_private",
  "viewers": [5, 8, 12]
}
```

**Timeline**:

```
0.0s - Request arrives at Posts Service (port 8083)

0.1s - Middleware chain executes
       - CORS middleware (add headers)
       - Logging middleware (log request)
       - Auth middleware (verify token with auth service)

0.3s - Auth Service responds
       - Token valid
       - User ID: 1
       - Adds userID=1 to request context

0.4s - Rate Limiter middleware
       - Check if user exceeded post creation limit
       - User has made 2 posts in last hour (under limit of 10)
       - Allow request

0.5s - CreatePost Handler executes
       - Extract userID from context: 1
       - Parse JSON body into CreatePostRequest
       - Call: h.postService.CreatePost(&req, userID=1)

0.6s - PostService.CreatePost executes
       Validation #1: Check privacy level
       - req.PrivacyLevel = "almost_private" ✓ (valid)
       
       Validation #2: Check content
       - req.Content = "The new feature is ready for review" ✓ (not empty)

0.7s - Create Post object
       post := &Post{
           UserID:       1,
           Title:        "Project Update",
           Content:      "The new feature is ready for review",
           PrivacyLevel: "almost_private",
           CreatedAt:    2024-10-16 15:30:00,
       }

0.8s - Database INSERT
       Query: INSERT INTO posts (user_id, title, content, privacy_level, created_at)
              VALUES (1, 'Project Update', '...', 'almost_private', '2024-10-16 15:30:00')
       Result: post.ID = 123

0.9s - Add Post Viewers (privacy is "almost_private" and viewers=[5,8,12])
       Call: db.AddPostViewers(database, postID=123, viewers=[5,8,12])

1.0s - Clear existing viewers
       Query: DELETE FROM post_viewers WHERE post_id = 123
       Result: 0 rows affected (new post, no existing viewers)

1.1s - Insert viewer #1
       Query: INSERT INTO post_viewers (post_id, user_id) VALUES (123, 5)

1.2s - Insert viewer #2
       Query: INSERT INTO post_viewers (post_id, user_id) VALUES (123, 8)

1.3s - Insert viewer #3
       Query: INSERT INTO post_viewers (post_id, user_id) VALUES (123, 12)

1.4s - Service returns post object to handler

1.5s - Handler sends response
       Status: 200 OK
       Body: {
           "post": {
               "id": 123,
               "user_id": 1,
               "title": "Project Update",
               "content": "The new feature is ready for review",
               "privacy_level": "almost_private",
               "created_at": "2024-10-16T15:30:00Z"
           }
       }

1.6s - Response reaches User A's browser
```

**Database State After Creation**:

```
posts table:
┌────┬─────────┬─────────────────┬───────────────────────────┬───────────────────┐
│ id │ user_id │ title           │ content                   │ privacy_level     │
├────┼─────────┼─────────────────┼───────────────────────────┼───────────────────┤
│123 │ 1       │ Project Update  │ The new feature is...     │ almost_private    │
└────┴─────────┴─────────────────┴───────────────────────────┴───────────────────┘

post_viewers table:
┌────┬─────────┬─────────┐
│ id │ post_id │ user_id │
├────┼─────────┼─────────┤
│ 1  │ 123     │ 5       │
│ 2  │ 123     │ 8       │
│ 3  │ 123     │ 12      │
└────┴─────────┴─────────┘
```

**Who Can Now See This Post**:
- ✅ User 1 (author)
- ✅ User 5 (in viewers list)
- ✅ User 8 (in viewers list)
- ✅ User 12 (in viewers list)
- ❌ Everyone else (even User 1's followers!)

[Back to Top](#table-of-contents)

---

## The Complete Flow - Viewing the Feed

Let's trace what happens when User A (ID=1) opens their feed.

**HTTP Request**:

```http
GET /posts HTTP/1.1
Host: localhost:8083
Authorization: Bearer abc123token
```

**Timeline**:

```
0.0s - Request arrives at Posts Service

0.1s - Middleware chain executes (CORS, Logging, Auth)

0.3s - Auth Service responds: userID = 1

0.4s - GetFeed Handler executes
       - Extract userID from context: 1
       - Call: h.postService.GetFeed(userID=1)

0.5s - PostService.GetFeed executes
       - Call: db.GetFeedPosts(database, userID=1)

0.6s - Complex SQL query executes

Database has these posts:
┌────┬─────────┬──────────────────────┬───────────────────┐
│ id │ user_id │ content              │ privacy_level     │
├────┼─────────┼──────────────────────┼───────────────────┤
│ 10 │ 1       │ "My own post"        │ private           │
│ 11 │ 2       │ "Everyone see this"  │ public            │
│ 12 │ 3       │ "Private update"     │ private           │
│ 13 │ 4       │ "Secret project"     │ almost_private    │
│ 14 │ 5       │ "For my followers"   │ private           │
└────┴─────────┴──────────────────────┴───────────────────┘

User 1's relationships:
- Following: User 5 (status='accepted')
- Not following: User 2, 3, 4
- In post_viewers: Post 13 (User 4's almost_private post)

SQL Query processes each post:

Post 10: privacy='private', user_id=1
  → Rule 2: p.user_id = 1 → TRUE ✅ (own post)
  
Post 11: privacy='public', user_id=2
  → Rule 1: p.privacy_level = 'public' → TRUE ✅
  
Post 12: privacy='private', user_id=3
  → Rule 2: p.user_id = 1? → FALSE
  → Rule 1: public? → FALSE
  → Rule 4: private + following user 3? → FALSE ❌ (not following)
  
Post 13: privacy='almost_private', user_id=4
  → Rule 2: p.user_id = 1? → FALSE
  → Rule 1: public? → FALSE
  → Rule 3: almost_private + in post_viewers? → TRUE ✅ (user 1 in whitelist)
  
Post 14: privacy='private', user_id=5
  → Rule 2: p.user_id = 1? → FALSE
  → Rule 1: public? → FALSE
  → Rule 4: private + following user 5 (accepted)? → TRUE ✅

0.9s - Query returns results (posts 10, 11, 13, 14)

1.0s - Service returns array of posts to handler

1.1s - Handler sends response
       Status: 200 OK
       Body: {
           "posts": [
               { "id": 14, "user_id": 5, "content": "For my followers", ... },
               { "id": 13, "user_id": 4, "content": "Secret project", ... },
               { "id": 11, "user_id": 2, "content": "Everyone see this", ... },
               { "id": 10, "user_id": 1, "content": "My own post", ... }
           ]
       }
       (Ordered by created_at DESC, newest first)

1.2s - Response reaches User A's browser
       Feed displays 4 posts
```

**Why Post 12 was excluded**:
- Post 12 is `privacy='private'` by User 3
- User 1 is NOT following User 3
- Therefore, User 1 cannot see it

**Why Post 13 was included**:
- Post 13 is `privacy='almost_private'` by User 4
- User 1 is in the `post_viewers` table for Post 13
- Therefore, User 1 CAN see it (even though not following User 4)

This demonstrates the power of "almost_private" - it bypasses the follow relationship entirely.

[Back to Top](#table-of-contents)

---

## The Complete Flow - Creating a Comment

Let's trace what happens when User A (ID=1) comments on a post.

**HTTP Request**:

```http
POST /comments HTTP/1.1
Host: localhost:8083
Authorization: Bearer abc123token
Content-Type: application/json

{
  "post_id": 123,
  "content": "Great post! I agree."
}
```

**Timeline**:

```
0.0s - Request arrives at Posts Service

0.1s - Middleware chain executes (CORS, Logging, Auth, Rate Limiter)

0.4s - CreateComment Handler executes
       - Extract userID from context: 1
       - Parse JSON body: postID=123, content="Great post! I agree."
       - Call: h.postService.CreateComment(&req, userID=1)

0.5s - PostService.CreateComment executes
       
       Access Check: Can user 1 comment on post 123?
       - Call: db.CheckPostAccess(database, postID=123, userID=1)

0.6s - CheckPostAccess query executes
       Query: SELECT 
                  CASE 
                      WHEN p.user_id = 1 THEN 1          -- Is User 1 the author?
                      WHEN p.privacy_level = 'public' THEN 1
                      WHEN p.privacy_level = 'private' AND EXISTS (...)
                      WHEN p.privacy_level = 'almost_private' AND EXISTS (...)
                      ELSE 0
                  END
              FROM posts p WHERE p.id = 123
       
       Post 123 details:
       - user_id: 4 (not User 1)
       - privacy_level: 'almost_private'
       - User 1 is in post_viewers (from earlier)
       
       Result: has_access = 1 (TRUE) ✅

0.7s - Access granted, continue with comment creation
       
       Validation: Check content
       - req.Content = "Great post! I agree." ✓ (not empty)

0.8s - Create Comment object
       comment := &Comment{
           PostID:    123,
           UserID:    1,
           Content:   "Great post! I agree.",
           CreatedAt: 2024-10-16 15:35:00,
       }

0.9s - Database INSERT
       Query: INSERT INTO comments (post_id, user_id, content, created_at)
              VALUES (123, 1, 'Great post! I agree.', '2024-10-16 15:35:00')
       Result: comment.ID = 45

1.0s - Service returns comment object to handler

1.1s - Handler sends response
       Status: 200 OK
       Body: {
           "comment": {
               "id": 45,
               "post_id": 123,
               "user_id": 1,
               "content": "Great post! I agree.",
               "created_at": "2024-10-16T15:35:00Z"
           }
       }

1.2s - Response reaches User A's browser
```

**If Access Was Denied**:

Let's say User B (ID=2) tries to comment on the same post 123:

```
User B (ID=2) → POST /comments {"post_id": 123, "content": "Nice!"}

0.5s - CheckPostAccess(postID=123, userID=2)
       
       Post 123:
       - user_id: 4 (not User 2)
       - privacy_level: 'almost_private'
       - User 2 NOT in post_viewers (only [1, 5, 8, 12])
       
       Result: has_access = 0 (FALSE) ❌

0.6s - Service returns error: "access denied: cannot comment on this post"

0.7s - Handler sends response:
       Status: 403 Forbidden
       Body: {"error": "access denied: cannot comment on this post"}
```

**Key Point**: You must be able to VIEW a post to COMMENT on it. The same access control rules apply to both viewing and commenting.

[Back to Top](#table-of-contents)

---

## Database Operations

All SQL queries are in `db/queries.go`.

### Post Queries

**CreatePost** - Insert new post:

```go
func CreatePost(db *sql.DB, post *Post) error {
    query := `
        INSERT INTO posts (user_id, title, content, image_path, privacy_level, created_at)
        VALUES (?, ?, ?, ?, ?, ?)
    `
    result, _ := db.Exec(query, post.UserID, post.Title, post.Content, 
                         post.ImagePath, post.PrivacyLevel, post.CreatedAt)
    
    id, _ := result.LastInsertId()
    post.ID = int(id)  // Populate post ID from database
    return nil
}
```

**GetPostByID** - Retrieve single post:

```go
func GetPostByID(db *sql.DB, postID int) (*Post, error) {
    query := `
        SELECT id, user_id, title, content, image_path, privacy_level, created_at
        FROM posts
        WHERE id = ?
    `
    post := &Post{}
    err := db.QueryRow(query, postID).Scan(&post.ID, &post.UserID, ...)
    return post, err
}
```

**UpdatePost** - Update existing post:

```go
func UpdatePost(db *sql.DB, post *Post) error {
    query := `
        UPDATE posts
        SET content = ?, image_path = ?, privacy_level = ?
        WHERE id = ?
    `
    _, err := db.Exec(query, post.Content, post.ImagePath, post.PrivacyLevel, post.ID)
    return err
}
```

**DeletePost** - Delete post (cascade deletes comments and viewers):

```go
func DeletePost(db *sql.DB, postID int) error {
    query := `DELETE FROM posts WHERE id = ?`
    _, err := db.Exec(query, postID)
    return err
    // CASCADE will automatically delete:
    // - All comments on this post
    // - All post_viewers entries for this post
}
```

**GetFeedPosts** - Complex feed query (see [Feed Algorithm](#the-feed-algorithm---what-you-see)):

```go
func GetFeedPosts(db *sql.DB, userID int) ([]*Post, error) {
    query := `
        SELECT DISTINCT p.id, p.user_id, p.title, p.content, p.image_path, 
                        p.privacy_level, p.created_at
        FROM posts p
        LEFT JOIN follows f ON p.user_id = f.following_id 
                            AND f.follower_id = ? 
                            AND f.status = 'accepted'
        LEFT JOIN post_viewers pv ON p.id = pv.post_id 
                                  AND pv.user_id = ?
        WHERE 
            p.privacy_level = 'public' OR
            p.user_id = ? OR
            (p.privacy_level = 'almost_private' AND pv.user_id IS NOT NULL) OR
            (p.privacy_level = 'private' AND f.follower_id IS NOT NULL)
        ORDER BY p.created_at DESC
    `
    rows, _ := db.Query(query, userID, userID, userID)
    // Scan rows into Post objects...
}
```

### Post Viewer Queries

**AddPostViewers** - Set whitelist for almost_private post:

```go
func AddPostViewers(db *sql.DB, postID int, userIDs []int) error {
    // Clear existing viewers
    db.Exec(`DELETE FROM post_viewers WHERE post_id = ?`, postID)
    
    // Insert new viewers
    stmt, _ := db.Prepare(`INSERT INTO post_viewers (post_id, user_id) VALUES (?, ?)`)
    defer stmt.Close()
    
    for _, userID := range userIDs {
        stmt.Exec(postID, userID)
    }
    return nil
}
```

### Comment Queries

**CreateComment** - Insert new comment:

```go
func CreateComment(db *sql.DB, comment *Comment) error {
    query := `
        INSERT INTO comments (post_id, user_id, content, image_path, created_at)
        VALUES (?, ?, ?, ?, ?)
    `
    result, _ := db.Exec(query, comment.PostID, comment.UserID, 
                         comment.Content, comment.ImagePath, comment.CreatedAt)
    
    id, _ := result.LastInsertId()
    comment.ID = int(id)
    return nil
}
```

**GetCommentsByPostID** - Get all comments for a post:

```go
func GetCommentsByPostID(db *sql.DB, postID int) ([]*Comment, error) {
    query := `
        SELECT id, post_id, user_id, content, image_path, created_at
        FROM comments
        WHERE post_id = ?
        ORDER BY created_at ASC  -- Chronological order (oldest first)
    `
    rows, _ := db.Query(query, postID)
    defer rows.Close()
    
    var comments []*Comment
    for rows.Next() {
        comment := &Comment{}
        rows.Scan(&comment.ID, &comment.PostID, ...)
        comments = append(comments, comment)
    }
    return comments, nil
}
```

### Access Control Query

**CheckPostAccess** - Determine if user can view post:

```go
func CheckPostAccess(db *sql.DB, postID, userID int) (bool, error) {
    query := `
        SELECT 
            CASE 
                WHEN p.user_id = ? THEN 1
                WHEN p.privacy_level = 'public' THEN 1
                WHEN p.privacy_level = 'private' AND EXISTS (
                    SELECT 1 FROM follows 
                    WHERE follower_id = ? AND following_id = p.user_id AND status = 'accepted'
                ) THEN 1
                WHEN p.privacy_level = 'almost_private' AND EXISTS (
                    SELECT 1 FROM post_viewers 
                    WHERE post_id = ? AND user_id = ?
                ) THEN 1
                ELSE 0
            END as has_access
        FROM posts p
        WHERE p.id = ?
    `
    var hasAccess int
    err := db.QueryRow(query, userID, userID, postID, userID, postID).Scan(&hasAccess)
    return hasAccess == 1, err
}
```

[Back to Top](#table-of-contents)

---

## HTTP REST Endpoints

All endpoints require authentication.

### Post Endpoints

**POST /posts** - Create a new post

```http
POST /posts HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "My First Post",
  "content": "This is the content of my post",
  "privacy_level": "public",
  "image_path": "/uploads/image.jpg",
  "viewers": []
}

Privacy levels:
- "public" → Everyone can see
- "private" → Only followers can see
- "almost_private" → Only specified viewers can see (provide viewers array)

Response 200 OK:
{
  "post": {
    "id": 123,
    "user_id": 1,
    "title": "My First Post",
    "content": "This is the content of my post",
    "privacy_level": "public",
    "created_at": "2024-10-16T15:30:00Z"
  }
}
```

**GET /posts** - Get user's feed

```http
GET /posts HTTP/1.1
Authorization: Bearer <token>

Response 200 OK:
{
  "posts": [
    {
      "id": 123,
      "user_id": 1,
      "title": "Latest Post",
      "content": "Content here",
      "privacy_level": "public",
      "created_at": "2024-10-16T15:30:00Z"
    },
    // ... more posts
  ]
}

Feed includes:
- All public posts
- Your own posts
- Private posts from users you follow
- Almost private posts where you're in the viewers list
```

**GET /posts/:id** - Get specific post

```http
GET /posts/123 HTTP/1.1
Authorization: Bearer <token>

Response 200 OK (if you have access):
{
  "post": {
    "id": 123,
    "user_id": 5,
    "content": "Post content",
    "privacy_level": "private",
    "created_at": "2024-10-16T15:00:00Z"
  }
}

Response 403 Forbidden (if no access):
{
  "error": "access denied"
}
```

**PUT /posts/:id** - Update post

```http
PUT /posts/123 HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "content": "Updated content",
  "privacy_level": "almost_private",
  "viewers": [5, 8, 12]
}

Response 200 OK (if you're the owner):
{
  "post": { /* updated post */ }
}

Response 403 Forbidden (if not owner):
{
  "error": "unauthorized: you can only update your own posts"
}
```

**DELETE /posts/:id** - Delete post

```http
DELETE /posts/123 HTTP/1.1
Authorization: Bearer <token>

Response 200 OK (if you're the owner):
{
  "message": "Post deleted successfully"
}

Response 403 Forbidden (if not owner):
{
  "error": "unauthorized: you can only delete your own posts"
}
```

### Comment Endpoints

**POST /comments** - Create a comment

```http
POST /comments HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "post_id": 123,
  "content": "Great post!",
  "image_path": "/uploads/comment_image.jpg"
}

Response 200 OK (if you can view the post):
{
  "comment": {
    "id": 45,
    "post_id": 123,
    "user_id": 1,
    "content": "Great post!",
    "created_at": "2024-10-16T15:35:00Z"
  }
}

Response 403 Forbidden (if you can't view the post):
{
  "error": "access denied: cannot comment on this post"
}
```

**GET /comments?post_id=123** - Get comments on a post

```http
GET /comments?post_id=123 HTTP/1.1
Authorization: Bearer <token>

Response 200 OK (if you can view the post):
{
  "comments": [
    {
      "id": 45,
      "post_id": 123,
      "user_id": 1,
      "content": "Great post!",
      "created_at": "2024-10-16T15:35:00Z"
    },
    {
      "id": 46,
      "post_id": 123,
      "user_id": 5,
      "content": "I agree!",
      "created_at": "2024-10-16T15:40:00Z"
    }
  ]
}

Response 403 Forbidden (if you can't view the post):
{
  "error": "access denied: cannot view comments on this post"
}
```

[Back to Top](#table-of-contents)

---

## Error Handling and Security

### Ownership Checks

Always verify ownership before allowing updates or deletes:

```go
// Get existing post
post, _ := db.GetPostByID(s.database, postID)

// Check ownership
if post.UserID != userID {
    return errors.New("unauthorized: you can only update your own posts")
}
```

### Access Control Enforcement

Every view/comment operation checks access first:

```go
hasAccess, _ := db.CheckPostAccess(s.database, postID, userID)
if !hasAccess {
    return nil, errors.New("access denied")
}
```

### Rate Limiting

Post creation and comments are rate-limited:

```go
// main.go
rateLimiter := middleware.NewRateLimiter()
mux.Handle("/posts", authMiddleware(rateLimiter.RateLimit(handler)))
mux.Handle("/comments", authMiddleware(rateLimiter.RateLimit(handler)))
```

This prevents spam and abuse.

### Validation

**Content validation**:
```go
if req.Content == "" {
    return errors.New("content is required")
}
```

**Privacy level validation**:
```go
if req.PrivacyLevel != "public" && req.PrivacyLevel != "private" && req.PrivacyLevel != "almost_private" {
    return errors.New("invalid privacy level")
}
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
- `400 Bad Request` - Invalid input
- `401 Unauthorized` - No token or invalid token
- `403 Forbidden` - Access denied (valid auth but no permission)
- `404 Not Found` - Resource doesn't exist
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

**Difference between 401 and 403**:
- `401 Unauthorized` - "Who are you?" (authentication failed)
- `403 Forbidden` - "I know who you are, but you can't do this" (authorization failed)

[Back to Top](#table-of-contents)

---

## Summary

**Posts Service Core Concepts**:

1. **Three Privacy Levels**
   - `public` → Everyone sees it
   - `private` → Only followers see it
   - `almost_private` → Only selected users see it (whitelist in post_viewers table)

2. **Feed Algorithm**
   - Shows posts based on privacy level, follow relationships, and viewer whitelist
   - Complex SQL query with LEFT JOINs for efficient access checking
   - Four rules: public posts, own posts, almost_private whitelist, private from following

3. **Access Control**
   - CheckPostAccess verifies if user can view a post
   - Enforced before viewing, commenting, or viewing comments
   - Uses CASE statement with EXISTS subqueries for efficiency

4. **Post Viewers System**
   - `post_viewers` table stores whitelist for almost_private posts
   - Updated when post privacy changes
   - Cascade delete when post is deleted

5. **Service Architecture**
   - Handlers: HTTP parsing and responses
   - Services: Business logic and validation
   - Database: SQL queries and data access

6. **Ownership and Authorization**
   - Users can only update/delete their own posts
   - 403 Forbidden vs 401 Unauthorized
   - Rate limiting on creation endpoints

**Why These Patterns?**

- **Three privacy levels**: Flexibility for users to control audience
- **Whitelist system**: Fine-grained control beyond just "public" or "followers"
- **Access checks**: Prevent unauthorized access to private content
- **Service layer**: Business logic separated from HTTP and database concerns

**Key Takeaway**: The Posts service implements a sophisticated privacy system with three distinct levels. The "almost_private" level is the most unique, allowing users to hand-pick specific viewers using the `post_viewers` table. Access control is enforced at every operation (view, comment, view comments) to ensure privacy boundaries are respected.

[Back to Top](#table-of-contents)

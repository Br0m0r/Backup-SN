# Quick Start Guide

> **Quick Navigation:** Use `Ctrl+F` to search, or click any link in the Table of Contents below.

## Table of Contents
- [Prerequisites](#prerequisites)
- [Database Setup](#database-setup)
- [Start Services](#start-services)
- [Test the API](#test-the-api)
- [Database Inspection (Docker)](#database-inspection-docker)
  - [View Users](#view-users)
  - [View Posts & Comments](#view-posts--comments)
  - [View Sessions](#view-sessions)
  - [View Relationships](#view-relationships-follows)
  - [View Messages](#view-messages)
  - [Advanced Queries](#advanced-queries)
- [When to Rebuild/Restart](#when-to-rebuildrestart)
- [Troubleshooting](#troubleshooting)
- [Useful Commands](#useful-commands)
- [Adding New Migrations](#adding-new-migrations)
- [Health Check](#health-check)
- [Port Reference](#port-reference)
- [Quick Command Reference](#quick-command-reference)

---

## Prerequisites
- Docker & Docker Compose
- SQLite3
- Web browser

---

## Database Setup

```bash
cd /home/mfoteino/Zone/social-network

# Create database with DELETE journal mode (simpler, no WAL files)
sqlite3 social_network.db "PRAGMA journal_mode=DELETE;"

# Run migrations
for file in db/migrations/*.up.sql; do 
    sqlite3 social_network.db < "$file"
done

# Verify (should show 13 tables)
sqlite3 social_network.db ".tables"
```

**Note:** We use `DELETE` journal mode instead of `WAL` for simplicity. This means:
- ✅ Single database file (no .db-wal or .db-shm)
- ✅ Immediate data visibility
- ⚠️ Database locked while containers access it (use Docker exec for queries)

**[Back to Top](#quick-start-guide)**

---

## Start Services

```bash
# Build and start all containers
sudo docker compose up --build -d

# Check containers are running
sudo docker ps
# Should show: auth-service, user-service, post-service

# View logs
sudo docker compose logs -f
# Look for: "Connected to database" and "Service starting on port"
```

**[Back to Top](#quick-start-guide)**

---

## Test the API

### Using Frontend (Simple)
```bash
firefox frontend/index.html
```
1. Register an account
2. Login
3. Update profile (Profile tab)
4. Create posts (Posts tab)
5. Search users (Users tab)

### Using cURL (Advanced)
```bash
# Health check
curl http://localhost:8081/health

# Register
curl -X POST http://localhost:8081/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@mail.com","password":"pass123","first_name":"Test","last_name":"User"}'

# Login (save token)
curl -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@mail.com","password":"pass123"}'

# Get profile (use token from login)
curl http://localhost:8082/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**[Back to Top](#quick-start-guide)**

---

## Database Inspection (Docker)

**Why Docker exec?** Since we use DELETE journal mode, the database is locked when containers are running. Query through any container to access the shared database.

### Quick Reference: Container Names
- `auth-service` - Port 8081
- `user-service` - Port 8082
- `post-service` - Port 8083

All containers share `/app/social_network.db`

---

### View Users

```bash
# List all users
sudo docker exec auth-service sqlite3 /app/social_network.db "SELECT id, username, email, first_name, last_name, created_at FROM users;"

# Pretty format
sudo docker exec auth-service sqlite3 /app/social_network.db \
  "SELECT printf('ID: %d | %s (%s) | Name: %s %s | Created: %s', 
    id, username, email, first_name, last_name, created_at) 
   FROM users;"

# Check specific user
sudo docker exec auth-service sqlite3 /app/social_network.db \
  "SELECT * FROM users WHERE username='tom';"

# Count users
sudo docker exec auth-service sqlite3 /app/social_network.db \
  "SELECT COUNT(*) as total_users FROM users;"

# Public vs private profiles
sudo docker exec auth-service sqlite3 /app/social_network.db \
  "SELECT is_public_profile, COUNT(*) FROM users GROUP BY is_public_profile;"
```

**[Back to Top](#quick-start-guide)**

---

### View Posts & Comments

```bash
# List all posts with author
sudo docker exec post-service sqlite3 /app/social_network.db \
  "SELECT p.id, u.username, p.content, p.privacy, p.created_at 
   FROM posts p 
   JOIN users u ON p.user_id = u.id 
   ORDER BY p.created_at DESC;"

# Posts by user
sudo docker exec post-service sqlite3 /app/social_network.db \
  "SELECT id, content, privacy, created_at 
   FROM posts WHERE user_id=1 ORDER BY created_at DESC;"

# Count posts by privacy level
sudo docker exec post-service sqlite3 /app/social_network.db \
  "SELECT privacy, COUNT(*) as count FROM posts GROUP BY privacy;"

# View comments with post and author
sudo docker exec post-service sqlite3 /app/social_network.db \
  "SELECT c.id, u.username, c.content, c.created_at, p.content as post_content
   FROM comments c
   JOIN users u ON c.user_id = u.id
   JOIN posts p ON c.post_id = p.id
   ORDER BY c.created_at DESC;"

# Comments on specific post
sudo docker exec post-service sqlite3 /app/social_network.db \
  "SELECT u.username, c.content, c.created_at 
   FROM comments c 
   JOIN users u ON c.user_id = u.id 
   WHERE c.post_id=1;"
```

**[Back to Top](#quick-start-guide)**

---

### View Sessions

```bash
# Active sessions
sudo docker exec auth-service sqlite3 /app/social_network.db \
  "SELECT s.id, u.username, s.token, s.expires_at 
   FROM sessions s 
   JOIN users u ON s.user_id = u.id 
   WHERE datetime(s.expires_at) > datetime('now');"

# All sessions with expiry status
sudo docker exec auth-service sqlite3 /app/social_network.db \
  "SELECT u.username, s.created_at, s.expires_at,
   CASE WHEN datetime(s.expires_at) > datetime('now') THEN 'ACTIVE' ELSE 'EXPIRED' END as status
   FROM sessions s 
   JOIN users u ON s.user_id = u.id;"

# Count sessions per user
sudo docker exec auth-service sqlite3 /app/social_network.db \
  "SELECT u.username, COUNT(s.id) as session_count 
   FROM users u 
   LEFT JOIN sessions s ON u.id = s.user_id 
   GROUP BY u.id;"
```

**[Back to Top](#quick-start-guide)**

---

### View Relationships (Follows)

```bash
# Who follows who
sudo docker exec user-service sqlite3 /app/social_network.db \
  "SELECT u1.username as follower, u2.username as following, f.created_at
   FROM follows f
   JOIN users u1 ON f.follower_id = u1.id
   JOIN users u2 ON f.followed_id = u2.id;"

# User's followers
sudo docker exec user-service sqlite3 /app/social_network.db \
  "SELECT u.username, COUNT(*) as follower_count
   FROM follows f
   JOIN users u ON f.followed_id = u.id
   WHERE u.username='tom';"

# User's following
sudo docker exec user-service sqlite3 /app/social_network.db \
  "SELECT u.username, COUNT(*) as following_count
   FROM follows f
   JOIN users u ON f.follower_id = u.id
   WHERE u.username='tom';"

# Popular users (most followers)
sudo docker exec user-service sqlite3 /app/social_network.db \
  "SELECT u.username, COUNT(f.follower_id) as followers
   FROM users u
   LEFT JOIN follows f ON u.id = f.followed_id
   GROUP BY u.id
   ORDER BY followers DESC
   LIMIT 10;"
```

**[Back to Top](#quick-start-guide)**

---

### View Messages

```bash
# Direct messages between users
sudo docker exec user-service sqlite3 /app/social_network.db \
  "SELECT u1.username as from_user, u2.username as to_user, m.content, m.created_at
   FROM messages m
   JOIN users u1 ON m.sender_id = u1.id
   JOIN users u2 ON m.receiver_id = u2.id
   ORDER BY m.created_at DESC
   LIMIT 20;"

# Conversation between two users
sudo docker exec user-service sqlite3 /app/social_network.db \
  "SELECT u.username, m.content, m.created_at
   FROM messages m
   JOIN users u ON m.sender_id = u.id
   WHERE (m.sender_id=1 AND m.receiver_id=2) OR (m.sender_id=2 AND m.receiver_id=1)
   ORDER BY m.created_at;"

# Group messages
sudo docker exec user-service sqlite3 /app/social_network.db \
  "SELECT g.name as group_name, u.username, gm.content, gm.created_at
   FROM group_messages gm
   JOIN users u ON gm.user_id = u.id
   JOIN groups g ON gm.group_id = g.id
   ORDER BY gm.created_at DESC
   LIMIT 20;"
```

**[Back to Top](#quick-start-guide)**

---

### Advanced Queries

```bash
# User activity summary
sudo docker exec user-service sqlite3 /app/social_network.db \
  "SELECT 
    u.username,
    COUNT(DISTINCT p.id) as posts_count,
    COUNT(DISTINCT c.id) as comments_count,
    COUNT(DISTINCT f1.follower_id) as followers,
    COUNT(DISTINCT f2.followed_id) as following
   FROM users u
   LEFT JOIN posts p ON u.id = p.user_id
   LEFT JOIN comments c ON u.id = c.user_id
   LEFT JOIN follows f1 ON u.id = f1.followed_id
   LEFT JOIN follows f2 ON u.id = f2.follower_id
   GROUP BY u.id;"

# Most active commenters
sudo docker exec post-service sqlite3 /app/social_network.db \
  "SELECT u.username, COUNT(c.id) as comment_count
   FROM users u
   JOIN comments c ON u.id = c.user_id
   GROUP BY u.id
   ORDER BY comment_count DESC
   LIMIT 10;"

# Recent activity feed
sudo docker exec post-service sqlite3 /app/social_network.db \
  "SELECT 'post' as type, u.username, p.content, p.created_at
   FROM posts p
   JOIN users u ON p.user_id = u.id
   UNION ALL
   SELECT 'comment' as type, u.username, c.content, c.created_at
   FROM comments c
   JOIN users u ON c.user_id = u.id
   ORDER BY created_at DESC
   LIMIT 20;"

# Database schema info
sudo docker exec auth-service sqlite3 /app/social_network.db ".schema users"
sudo docker exec auth-service sqlite3 /app/social_network.db ".tables"

# Database size and record counts
sudo docker exec auth-service sqlite3 /app/social_network.db \
  "SELECT 
    (SELECT COUNT(*) FROM users) as users,
    (SELECT COUNT(*) FROM posts) as posts,
    (SELECT COUNT(*) FROM comments) as comments,
    (SELECT COUNT(*) FROM sessions) as sessions,
    (SELECT COUNT(*) FROM follows) as follows,
    (SELECT COUNT(*) FROM messages) as messages;"
```

**Pro Tip:** Add `.mode column` and `.headers on` for better formatting:
```bash
sudo docker exec auth-service sqlite3 /app/social_network.db \
  ".mode column" ".headers on" "SELECT * FROM users LIMIT 3;"
```

**[Back to Top](#quick-start-guide)**

---

## When to Rebuild/Restart

| Change Type | Command | Why |
|-------------|---------|-----|
| **Go code** (handlers, models) | `sudo docker compose up --build -d` | Code compiled into image |
| **Frontend** (HTML/CSS/JS) | Just refresh browser | Not in containers |
| **Database schema** | Stop → Run migration → Start | Need new tables |
| **docker-compose.yml** | `sudo docker compose down && up -d` | Config reload |

### Quick Rebuild Single Service
```bash
sudo docker compose up --build -d user-service
```

**[Back to Top](#quick-start-guide)**

---

## Troubleshooting

### Services won't start
```bash
# Check logs for errors
sudo docker logs user-service --tail 50

# Common fix: restart everything
sudo docker compose down
sudo docker compose up --build -d
```

### "no such table: users"
```bash
# Run migrations
for file in db/migrations/*.up.sql; do 
    sqlite3 social_network.db < "$file"
done
```

### 404/500 errors
```bash
# Check if services are running
sudo docker ps

# Check database has tables
sqlite3 social_network.db ".tables"

# View real-time logs
sudo docker compose logs -f
```

### Start completely fresh
```bash
# Stop containers
sudo docker compose down

# Delete database
rm social_network.db

# Recreate with DELETE journal mode
sqlite3 social_network.db "PRAGMA journal_mode=DELETE;"

# Run all migrations
for f in db/migrations/*.up.sql; do sqlite3 social_network.db < "$f"; done

# Rebuild and start
sudo docker compose up --build -d
```

**[Back to Top](#quick-start-guide)**

---

## Useful Commands

### Container Management
```bash
# View logs
sudo docker logs auth-service --tail 50
sudo docker logs user-service --tail 50
sudo docker logs post-service --tail 50

# Follow logs live
sudo docker compose logs -f

# Stop services
sudo docker compose down

# Container status
sudo docker ps
sudo docker stats

# Restart single service
sudo docker compose restart user-service

# Remove stopped containers
sudo docker compose rm
```

### Database Queries
**⚠️ Important:** With DELETE journal mode, the database is locked when containers run.

**While containers are running:** Use Docker exec (see [Database Inspection](#database-inspection-docker))
```bash
# Example: List users via Docker
sudo docker exec auth-service sqlite3 /app/social_network.db "SELECT * FROM users;"
```

**With containers stopped:** Use local sqlite3
```bash
# Stop containers first
sudo docker compose down

# Then query locally
sqlite3 social_network.db "SELECT * FROM users;"
sqlite3 social_network.db "SELECT * FROM posts;"
sqlite3 social_network.db ".tables"
```

**[Back to Top](#quick-start-guide)**

---

## Adding New Migrations

```bash
# 1. Create migration files
touch db/migrations/000014_NewFeature.up.sql
touch db/migrations/000014_NewFeature.down.sql

# 2. Write SQL in .up.sql

# 3. Stop containers
sudo docker compose down

# 4. Apply migration
sqlite3 social_network.db < db/migrations/000014_NewFeature.up.sql

# 5. Restart
sudo docker compose up -d
```

**[Back to Top](#quick-start-guide)**

---

## Health Check

Services are working when:
- ✅ `sudo docker ps` shows 3 running containers
- ✅ `curl http://localhost:8081/health` returns `{"status":"healthy"}`
- ✅ `curl http://localhost:8082/health` returns `{"status":"healthy"}`
- ✅ `curl http://localhost:8083/health` returns `{"status":"healthy"}`
- ✅ Frontend loads without console errors
- ✅ Can register and login successfully

**[Back to Top](#quick-start-guide)**

---

## Port Reference

| Service | Port | Health Check |
|---------|------|--------------|
| Auth    | 8081 | `curl http://localhost:8081/health` |
| User    | 8082 | `curl http://localhost:8082/health` |
| Post    | 8083 | `curl http://localhost:8083/health` |

**[Back to Top](#quick-start-guide)**

---

## Quick Command Reference

### Most Common Commands
```bash
# Start everything
sudo docker compose up -d

# View logs
sudo docker compose logs -f

# Check database (containers running)
sudo docker exec auth-service sqlite3 /app/social_network.db "SELECT * FROM users;"

# Rebuild after code changes
sudo docker compose up --build -d

# Stop everything
sudo docker compose down
```

### First Time Setup
1. Create database: `sqlite3 social_network.db "PRAGMA journal_mode=DELETE;"`
2. Run migrations: `for f in db/migrations/*.up.sql; do sqlite3 social_network.db < "$f"; done`
3. Start services: `sudo docker compose up --build -d`
4. Test: Open `frontend/index.html` in browser

### Common Workflows

**After changing Go code:**
```bash
sudo docker compose up --build -d
```

**After changing frontend (HTML/CSS/JS):**
```
Just refresh browser (no rebuild needed)
```

**To inspect database while services run:**
```bash
# See Database Inspection section above
sudo docker exec auth-service sqlite3 /app/social_network.db "SELECT * FROM users;"
```

**To debug 404/500 errors:**
1. Check logs: `sudo docker compose logs -f`
2. Verify tables exist: `sudo docker exec auth-service sqlite3 /app/social_network.db ".tables"`
3. Check services running: `sudo docker ps`

---

**Need more details?** Check `notes/todo.md` or service README files in `services/*/`

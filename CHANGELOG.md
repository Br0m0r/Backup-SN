# Developer Changelog

> Brief log of all changes made to the project. Each entry is max 2 sentences.

---

## 2025-10-10

### Database Migrations Update
**Updated all database migrations to match flowchart schema.** Added 5 new tables (PostViewers, GroupMembers, GroupMessages, Events, EventResponses) and modified existing tables to include missing fields like `status`, `is_read`, `nickname`, and `privacy_level`.

### Auth Service Query Fixes
**Fixed queries.go and user.go model to match updated database schema.** Changed `Avatar_url` to `avatar_path` and added missing `nickname` field in all user queries.

### User Service Implementation
**Built complete user microservice with profile, follow system, and search functionality.** Includes GET/PUT /profile, POST/DELETE /follow, GET /followers, GET /following, GET /search endpoints per flowchart, with auth middleware calling auth-service for token verification.

### Standardized API Response Format
**Updated auth service response.go to match user service pattern.** All services now use consistent Response struct with `success`, `data`, and `error` fields for uniform API responses across microservices.

### Rate Limiting and Connection Pooling
**Added selective rate limiting to user service /follow endpoint and connection pool settings to auth service.** Rate limiting prevents spam following (10 requests/sec per IP), and connection pooling (MaxOpenConns=25, MaxIdleConns=5) improves database performance for both services.

### Email Privacy Implementation
**Added privacy controls to user profile endpoint to hide sensitive data.** GetProfile handler now returns full profile (including email and DOB) only to profile owner, while other users receive public profile via PublicProfile() method.

### Handler Endpoint Comments Standardization
**Added consistent endpoint comments to all auth service handlers.** All handler functions now clearly specify which HTTP method and endpoint they handle (e.g., "handles POST /register requests").

### Post Service Implementation
**Built complete post microservice with CRUD operations, comments, and three-tier privacy system.** Includes POST/GET/PUT/DELETE /posts, POST/GET /comments endpoints with privacy levels (public, private, almost_private), access control logic, rate limiting on post/comment creation, and integration with auth service for token verification on port 8083.

---

## Template for Future Entries

```markdown
### [Feature/Fix Name]
**[Brief description in 1-2 sentences.]** [Impact or what was changed.]
```

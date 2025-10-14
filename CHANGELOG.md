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

### Test Frontend Implementation
**Created single-page HTML/JavaScript test client for all microservices.** Simple browser-based UI with no build tools required, featuring auth (register/login/logout), profile management (view/update/search users/follow system), and posts (create/view feed/comments/delete) with tabbed interface and JSON response viewers for debugging.

### Post Service
- **FIXED**: Added database connection initialization in main.go. Post service now properly connects to SQLite database with connection pooling.
- **FIXED**: Added HealthHandler to post handlers for health check endpoint.

### Frontend Fixes
**Fixed frontend to match actual API requirements and separated code into multiple files.** Removed DOB field from registration (not in API), fixed login to use email instead of identifier, separated HTML/CSS/JS into index.html, style.css, and app.js for better maintainability.

### Database Migration Script and Fixes
**Created automated migration script and fixed missing tables issue.** Added migrate.sh to apply all pending migrations in order, manually applied missing posts/comments/sessions tables, all 13 migrations now properly executed in database.

### Database Path Standardization
**Fixed environment variable inconsistency across all services.** Changed user service from DB_PATH to DATABASE_PATH to match auth and post services, ensuring all three microservices correctly reference /app/social_network.db in Docker containers.

### Comprehensive Logging Implementation
**Added detailed logging with safe pointer dereferencing to user service.** Implemented helper functions (getStrValue, getBoolValue) in user.go handler for logging actual values instead of memory addresses, added route handler logging to debug 404 errors, and enhanced UpdateUserProfile with field-by-field logging.

### Docker Volume Mount Fixes
**Added explicit read-write flags to all database volume mounts.** Updated docker-compose.yml with :rw flags on all three services' volume mounts to ensure proper shared database access across containers.

### SQLite Journal Mode Migration
**Switched from WAL to DELETE journal mode for simplified database operations.** Removed WAL autocheckpoint code from all three service main.go files, updated database initialization to use DELETE mode, eliminating .db-wal and .db-shm files for immediate data visibility and single-file simplicity.

---

## 2025-10-14

### Group Service Implementation
**Built complete group microservice with full CRUD operations, member management, and event system.** Implemented groups, group_members, group_messages, events, and event_responses tables with POST/GET/PUT /groups, POST /groups/:id/invite, POST /groups/:id/request, POST/GET /groups/:id/requests/respond, POST/GET /events, POST /events/:id/respond endpoints per exercise requirements, running on port 8084 with proper auth middleware integration.

### Frontend Groups Tab Implementation
**Added comprehensive Groups UI to test frontend with all group management features.** Created groups tab with create group form, browse/my groups lists, group detail views, member management, event creation/RSVP system, join request handling, and invite system with proper styling matching the purple gradient theme.

### API Response Handling Fix
**Fixed frontend to properly access nested API response structure.** Updated all 11 group-related JavaScript functions to correctly access result.data.data instead of result.data, resolving "failed to create/load" errors caused by incorrect response object traversal.

### Error Handling Improvements
**Added user-friendly error messages for duplicate group names and database constraints.** Backend now returns HTTP 409 Conflict with clear messages like "A group with this name already exists" instead of generic 500 errors for UNIQUE constraint violations.

### UX Enhancement for Join Requests
**Implemented has_pending_request field to prevent duplicate join request errors.** Backend queries now check for pending membership status, frontend displays disabled "Request Pending" button with yellow dashed badge instead of allowing users to click "Request to Join" multiple times.

### Frontend Code Modularization
**Refactored monolithic app.js into 7 separate module files for better maintainability.** Split into config.js (global state/URLs), utils.js (shared functions), auth.js, users.js, posts.js, groups.js, and events.js modules with proper namespacing and global exports for onclick handlers.

### DOM API Naming Conflict Resolution
**Fixed createEvent() function conflict with native Document.createEvent() method.** Renamed function to createGroupEvent() to prevent browser from calling native DOM API instead of custom event creation function, resolving "At least 1 argument required" TypeError.



---

## Template for Future Entries

```markdown
### [Feature/Fix Name]
**[Brief description in 1-2 sentences.]** [Impact or what was changed.]
```

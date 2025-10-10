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

---

## Template for Future Entries

```markdown
### [Feature/Fix Name]
**[Brief description in 1-2 sentences.]** [Impact or what was changed.]
```

# Database Migrations

This directory contains database migrations for the social network application.

## Running Migrations

### Prerequisites

Install the golang-migrate tool:

**Windows:**
```powershell
scoop install migrate
# or
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

**macOS/Linux:**
```bash
brew install golang-migrate
# or
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Run All Migrations

**Windows (PowerShell):**
```powershell
cd social-network
.\db\migrate.ps1
```

**Linux/macOS:**
```bash
cd social-network
chmod +x db/migrate.sh
./db/migrate.sh
```

### Manual Migration Commands

**Migrate to latest version:**
```bash
migrate -path ./db/migrations -database "sqlite3://./social_network.db" up
```

**Check current version:**
```bash
migrate -path ./db/migrations -database "sqlite3://./social_network.db" version
```

**Rollback one migration:**
```bash
migrate -path ./db/migrations -database "sqlite3://./social_network.db" down 1
```

**Go to specific version:**
```bash
migrate -path ./db/migrations -database "sqlite3://./social_network.db" goto VERSION
```

**Force version (if migrations are dirty):**
```bash
migrate -path ./db/migrations -database "sqlite3://./social_network.db" force VERSION
```

## Migration Files

Migrations are numbered sequentially and come in pairs:
- `XXXXXX_Description.up.sql` - Applied when migrating up
- `XXXXXX_Description.down.sql` - Applied when rolling back

Current migrations:
1. Users table
2. Posts table
3. Comments table
4. Follows table
5. Messages table
6. Groups table
7. Notifications table
8. Sessions table
9. PostViewers table
10. GroupMembers table
11. GroupMessages table
12. Events table
13. EventResponses table
14. AddTitleToPosts
15. AddUserSearchIndex
16. AddImagePathToMessages
17. AddGroupIdToPosts
18. AddInvitedStatusToGroupMembers

## Creating New Migrations

To create a new migration:

```bash
migrate create -ext sql -dir ./db/migrations -seq description_of_migration
```

This will create two files:
- `XXXXXX_description_of_migration.up.sql`
- `XXXXXX_description_of_migration.down.sql`

Edit both files to define your schema changes.

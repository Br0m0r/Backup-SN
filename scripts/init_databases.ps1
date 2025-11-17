# Initialize separate databases for each service (PowerShell)

Write-Host "Creating separate databases for each service..." -ForegroundColor Blue

# Create data directory if it doesn't exist
if (-not (Test-Path -Path "data")) {
    New-Item -ItemType Directory -Path "data" | Out-Null
}

# Auth Service Database
Write-Host "Creating auth_service.db..." -ForegroundColor Green
Get-Content "db\migrations_per_service\auth\001_create_users.sql" | sqlite3 "data\auth_service.db"
Write-Host "Auth service database created."

# User Service Database
Write-Host "Creating user_service.db..." -ForegroundColor Green
Get-Content "db\migrations_per_service\users\001_create_user_profiles.sql" | sqlite3 "data\user_service.db"
Write-Host "User service database created."

# Post Service Database
Write-Host "Creating post_service.db..." -ForegroundColor Green
Get-Content "db\migrations_per_service\posts\001_create_posts.sql" | sqlite3 "data\post_service.db"
Write-Host "Post service database created."

# Group Service Database
Write-Host "Creating group_service.db..." -ForegroundColor Green
Get-Content "db\migrations_per_service\groups\001_create_groups.sql" | sqlite3 "data\group_service.db"
Write-Host "Group service database created."

# Chat Service Database
Write-Host "Creating chat_service.db..." -ForegroundColor Green
Get-Content "db\migrations_per_service\chat\001_create_messages.sql" | sqlite3 "data\chat_service.db"
Write-Host "Chat service database created."

# Notification Service Database
Write-Host "Creating notif_service.db..." -ForegroundColor Green
Get-Content "db\migrations_per_service\notifications\001_create_notifications.sql" | sqlite3 "data\notif_service.db"
Write-Host "Notification service database created."

Write-Host "`nAll databases created successfully!" -ForegroundColor Blue
Write-Host "`nDatabase files created in .\data\ directory:"
Get-ChildItem "data\*.db" | Format-Table Name, Length -AutoSize

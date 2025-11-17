#!/bin/bash
# Initialize separate databases for each service

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Creating separate databases for each service...${NC}"

# Create data directory if it doesn't exist
mkdir -p data

# Auth Service Database
echo -e "${GREEN}Creating auth_service.db...${NC}"
sqlite3 data/auth_service.db < db/migrations_per_service/auth/001_create_users.sql
echo "Auth service database created."

# User Service Database
echo -e "${GREEN}Creating user_service.db...${NC}"
sqlite3 data/user_service.db < db/migrations_per_service/users/001_create_user_profiles.sql
echo "User service database created."

# Post Service Database
echo -e "${GREEN}Creating post_service.db...${NC}"
sqlite3 data/post_service.db < db/migrations_per_service/posts/001_create_posts.sql
echo "Post service database created."

# Group Service Database
echo -e "${GREEN}Creating group_service.db...${NC}"
sqlite3 data/group_service.db < db/migrations_per_service/groups/001_create_groups.sql
echo "Group service database created."

# Chat Service Database
echo -e "${GREEN}Creating chat_service.db...${NC}"
sqlite3 data/chat_service.db < db/migrations_per_service/chat/001_create_messages.sql
echo "Chat service database created."

# Notification Service Database
echo -e "${GREEN}Creating notif_service.db...${NC}"
sqlite3 data/notif_service.db < db/migrations_per_service/notifications/001_create_notifications.sql
echo "Notification service database created."

echo -e "${BLUE}All databases created successfully!${NC}"
echo ""
echo "Database files created in ./data/ directory:"
ls -lh data/*.db

# Group Service

**Port:** 8084

## Overview

The Group Service handles all group-related functionality including group creation, membership management, events, and group chat.

## Features

### Groups
- Create groups with name, description, and optional image
- Browse all groups
- View group details with member count
- Update group details (creator only)
- Automatic creator becomes admin member

### Membership Management
- **Invite Users:** Any group member can invite others (creates pending invitation)
- **Join Requests:** Users can request to join a group (creates pending request)
- **Accept/Reject:** Group creator can approve or reject pending requests
- Two types of members: `admin` (creator) and `member`
- Status tracking: `pending` or `accepted`

### Events
- Create events within a group (members only)
- Event details: title, description, date/time
- RSVP system with three options:
  - `going`
  - `not_going`
  - `interested`
- View event with response counts
- View all group events

### Group Chat
- Post messages in group chat (members only)
- Retrieve chat history with configurable limit
- Messages stored with sender ID and timestamp

## API Endpoints

### Group Management

```bash
# Create a group
POST /groups
Authorization: Bearer <token>
{
  "name": "Tech Enthusiasts",
  "description": "A group for tech lovers",
  "image_url": "https://example.com/image.jpg"  # optional
}

# Get all groups (browse)
GET /groups
Authorization: Bearer <token>

# Get specific group
GET /groups/:id
Authorization: Bearer <token>

# Update group (creator only)
PUT /groups/:id
Authorization: Bearer <token>
{
  "name": "Updated Name",
  "description": "Updated description"
}
```

### Membership

```bash
# Invite a user to group (members can invite)
POST /groups/:id/invite
Authorization: Bearer <token>
{
  "user_id": 5
}

# Request to join a group
POST /groups/:id/request
Authorization: Bearer <token>

# Get pending requests (creator only)
GET /groups/:id/requests
Authorization: Bearer <token>

# Respond to join request (creator only)
POST /groups/:id/requests/respond
Authorization: Bearer <token>
{
  "member_id": 123,
  "accept": true
}

# Get group members (members only)
GET /groups/:id/members
Authorization: Bearer <token>
```

### Events

```bash
# Create an event (members only)
POST /events
Authorization: Bearer <token>
{
  "group_id": 1,
  "title": "Weekly Meetup",
  "description": "Discuss latest tech trends",
  "event_time": "2025-10-20T18:00:00Z"  # ISO 8601 format
}

# Get event details
GET /events/:id
Authorization: Bearer <token>

# Get all events for a group (members only)
GET /groups/:id/events
Authorization: Bearer <token>

# Respond to event (RSVP)
POST /events/respond
Authorization: Bearer <token>
{
  "event_id": 1,
  "response": "going"  # or "not_going", "interested"
}
```

### Group Chat

```bash
# Post a message (members only)
POST /groups/:id/messages
Authorization: Bearer <token>
{
  "content": "Hello everyone!"
}

# Get chat history (members only)
GET /groups/:id/messages?limit=50
Authorization: Bearer <token>
```

## Database Schema

### Tables Used
- `groups` - Group information
- `group_members` - Membership tracking (with role and status)
- `events` - Group events
- `event_responses` - RSVP responses
- `group_messages` - Chat messages

## Permissions

| Action | Who Can Do It |
|--------|--------------|
| Create group | Any authenticated user |
| View group details | Any authenticated user (for browsing) |
| Update group | Creator only |
| Invite members | Any accepted member |
| Request to join | Any authenticated user |
| Accept/reject requests | Creator only |
| Create events | Any accepted member |
| RSVP to events | Any accepted member |
| Post messages | Any accepted member |
| View messages | Any accepted member |

## Response Format

All responses follow the standard format:

```json
{
  "success": true,
  "data": { ... }
}
```

Or for errors:

```json
{
  "success": false,
  "error": "Error message"
}
```

## Group Model

```json
{
  "id": 1,
  "name": "Tech Enthusiasts",
  "description": "A group for tech lovers",
  "image_url": "https://example.com/image.jpg",
  "creator_id": 1,
  "created_at": "2025-10-10T12:00:00Z",
  "member_count": 15,
  "is_member": true,
  "is_creator": false
}
```

## Event Model

```json
{
  "id": 1,
  "group_id": 1,
  "creator_id": 1,
  "title": "Weekly Meetup",
  "description": "Discuss latest tech trends",
  "event_time": "2025-10-20T18:00:00Z",
  "created_at": "2025-10-10T12:00:00Z",
  "going_count": 10,
  "not_going_count": 2,
  "interested_count": 5,
  "user_response": "going"
}
```

## Testing

```bash
# Health check
curl http://localhost:8084/health

# Create a group (need valid token)
curl -X POST http://localhost:8084/groups \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Group","description":"A test group"}'

# Browse groups
curl http://localhost:8084/groups \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Integration with Other Services

- **Auth Service (8081):** Token verification via `/internal/verify-token`
- **Database:** Shared SQLite database at `/app/social_network.db`
- **Frontend:** Will connect to `http://localhost:8084` for group features

## Next Steps

1. Build and deploy: `docker-compose up --build -d`
2. Test endpoints with cURL or Postman
3. Integrate with frontend for UI
4. Consider WebSocket for real-time group chat notifications

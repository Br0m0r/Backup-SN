# Notifications Service API Documentation

Base URL: `http://localhost:8086`

## Overview

The Notifications Service manages real-time notifications for user activities across the platform. Supports WebSocket for instant delivery and HTTP endpoints for notification management.

---

## WebSocket Endpoint

### Connect to Notifications WebSocket
- **Protocol**: WebSocket
- **Path**: `/ws`
- **Auth Required**: Yes
- **Description**: Real-time notification updates

#### Authentication
Include token via query parameter:
```
ws://localhost:8086/ws?token=<bearer_token>
```

#### Incoming Messages (Server → Client)
```json
{
  "id": 123,
  "user_id": 456,
  "type": "follow",
  "title": "New Follower",
  "message": "John Doe started following you",
  "link": "/profile/789",
  "is_read": false,
  "created_at": "2025-11-08T10:00:00Z"
}
```

---

## Notification Endpoints

### Create Notification (Internal)
- **Method**: `POST`
- **Path**: `/notifications`
- **Auth Required**: No (service-to-service only)
- **Description**: Create a new notification (called by other services)
- **Request Body**:
  ```json
  {
    "user_id": 123,
    "type": "follow",
    "title": "New Follower",
    "message": "John Doe started following you",
    "link": "/profile/456"
  }
  ```
- **Notification Types**:
  - `follow`: User followed you
  - `follow_request`: User requested to follow you
  - `group_invite`: Invited to group
  - `group_request`: User requested to join your group
  - `event`: Group event notification
  - `message`: New message
  - `comment`: Comment on your post
  - `post`: New post from followed user
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "id": 123,
      "user_id": 456,
      "type": "follow",
      "title": "New Follower",
      "message": "John Doe started following you",
      "link": "/profile/789",
      "is_read": false,
      "created_at": "2025-11-08T10:00:00Z"
    }
  }
  ```
- **Note**: Also broadcasts via WebSocket to online users

---

### Get Notifications
- **Method**: `GET`
- **Path**: `/notifications/list`
- **Auth Required**: Yes
- **Description**: Get user's notifications (paginated)
- **Query Parameters**:
  - `limit`: Number of notifications (default: 20)
  - `offset`: Pagination offset (default: 0)
  - `unread`: Set to "true" to get only unread notifications
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "notifications": [
        {
          "id": 123,
          "user_id": 456,
          "type": "follow",
          "title": "New Follower",
          "message": "John Doe started following you",
          "link": "/profile/789",
          "is_read": false,
          "created_at": "2025-11-08T10:00:00Z"
        }
      ],
      "count": 1
    }
  }
  ```

#### Get Unread Only
```
GET /notifications/list?unread=true
```

#### Pagination
```
GET /notifications/list?limit=10&offset=20
```

---

### Get Unread Count
- **Method**: `GET`
- **Path**: `/notifications/unread-count`
- **Auth Required**: Yes
- **Description**: Get count of unread notifications
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "unread_count": 5
    }
  }
  ```

---

### Mark as Read
- **Method**: `PUT` or `POST`
- **Path**: `/notifications/read/:id`
- **Auth Required**: Yes
- **Description**: Mark a specific notification as read
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "marked": true
    }
  }
  ```
- **Error Responses**:
  - `400`: Invalid notification ID
  - `404`: Notification not found or not owned by user

---

### Mark All as Read
- **Method**: `POST`
- **Path**: `/notifications/read-all`
- **Auth Required**: Yes
- **Description**: Mark all notifications as read
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "marked": true
    }
  }
  ```

---

### Delete Notification
- **Method**: `DELETE`
- **Path**: `/notifications/delete/:id`
- **Auth Required**: Yes
- **Description**: Delete a specific notification
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "deleted": true
    }
  }
  ```
- **Error Responses**:
  - `400`: Invalid notification ID
  - `404`: Notification not found or not owned by user

---

### Health Check
- **Method**: `GET`
- **Path**: `/health`
- **Description**: Service health status
- **Success Response**:
  ```json
  {
    "status": "healthy",
    "service": "notifications"
  }
  ```

---

## Notification Types

### Follow Notifications

#### New Follower
```json
{
  "type": "follow",
  "title": "New Follower",
  "message": "John Doe started following you",
  "link": "/profile/123"
}
```

#### Follow Request
```json
{
  "type": "follow_request",
  "title": "Follow Request",
  "message": "Jane Smith wants to follow you",
  "link": "/profile/456"
}
```

---

### Group Notifications

#### Group Invite
```json
{
  "type": "group_invite",
  "title": "Group Invitation",
  "message": "You've been invited to join 'Tech Enthusiasts'",
  "link": "/groups/789"
}
```

#### Join Request
```json
{
  "type": "group_request",
  "title": "Join Request",
  "message": "Bob Jones wants to join your group 'Tech Enthusiasts'",
  "link": "/groups/789/requests"
}
```

---

### Event Notifications

```json
{
  "type": "event",
  "title": "New Event",
  "message": "New event: 'Team Meeting' in 'Work Group'",
  "link": "/events/111"
}
```

---

### Engagement Notifications

#### New Comment
```json
{
  "type": "comment",
  "title": "New Comment",
  "message": "Alice commented on your post: 'Great content!'",
  "link": "/posts/222"
}
```

#### New Post
```json
{
  "type": "post",
  "title": "New Post",
  "message": "John Doe shared a new post",
  "link": "/posts/333"
}
```

---

### Message Notifications

```json
{
  "type": "message",
  "title": "New Message",
  "message": "You have a new message from Sarah",
  "link": "/chat/444"
}
```

---

## WebSocket Connection Management

### Connection Lifecycle
1. **Connect**: Client establishes WebSocket with auth token
2. **Active**: Receives real-time notifications
3. **Disconnect**: Connection closes

### Real-time Delivery
- Notifications are broadcast immediately to online users
- Offline users receive notifications when they reconnect
- All notifications also retrievable via REST endpoints

---

## Authentication

All endpoints (except `/notifications` POST and `/health`) require authentication:

```
Authorization: Bearer <token>
```

For WebSocket:
```
ws://localhost:8086/ws?token=<bearer_token>
```

---

## Service-to-Service Integration

### Creating Notifications from Other Services

Other microservices create notifications by calling:
```
POST http://notifications-service:8086/notifications
```

**No authentication required** for this internal endpoint (should be secured via network policies in production).

### Notification Triggers

**Users Service** triggers:
- New follower
- Follow request

**Groups Service** triggers:
- Group invitation
- Join request
- New event

**Posts Service** triggers:
- New comment
- New post from followed user

**Chat Service** triggers:
- New message (optional, can use separate mechanism)

---

## Read State Management

- Notifications start as `is_read: false`
- User can mark individual notifications as read
- User can mark all notifications as read at once
- Read state persists across sessions

---

## Pagination

For large notification lists, use pagination:
```
GET /notifications/list?limit=20&offset=0
```

- `limit`: Results per page (default: 20)
- `offset`: Skip this many results (default: 0)

---

## Error Handling

### HTTP Errors
```json
{
  "success": false,
  "error": "Error message description"
}
```

### WebSocket Errors
Disconnection occurs on authentication failure or server error. Client should reconnect with exponential backoff.

---

**Service Port**: 8086  
**Last Updated**: November 8, 2025

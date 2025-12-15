# Groups Service API Documentation

Base URL: `http://localhost:8084`

## Overview

The Groups Service manages group creation, membership, group messaging, and group events. Groups support admin-based management with invite and request-to-join functionality.

---

## Group Endpoints

### Create Group
- **Method**: `POST`
- **Path**: `/groups`
- **Auth Required**: Yes
- **Description**: Create a new group (creator becomes admin)
- **Request Body**:
  ```json
  {
    "name": "Group Name",
    "description": "Group description"
  }
  ```
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "id": 123,
      "name": "Group Name",
      "description": "Group description",
      "creator_id": 456,
      "created_at": "2025-11-08T10:00:00Z"
    }
  }
  ```
- **Error Responses**:
  - `409`: Group name already exists

---

### Get All Groups
- **Method**: `GET`
- **Path**: `/groups`
- **Auth Required**: Yes
- **Description**: Get all groups (shows membership status for user)
- **Success Response**:
  ```json
  {
    "success": true,
    "data": [
      {
        "id": 123,
        "name": "Group Name",
        "description": "Group description",
        "creator_id": 456,
        "member_count": 10,
        "is_member": true,
        "is_admin": false,
        "created_at": "2025-11-08T10:00:00Z"
      }
    ]
  }
  ```

---

### Get Single Group
- **Method**: `GET`
- **Path**: `/groups/:id`
- **Auth Required**: Yes
- **Description**: Get detailed information about a specific group
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "id": 123,
      "name": "Group Name",
      "description": "Group description",
      "creator_id": 456,
      "member_count": 10,
      "is_member": true,
      "is_admin": false,
      "created_at": "2025-11-08T10:00:00Z"
    }
  }
  ```

---

### Update Group
- **Method**: `PUT`
- **Path**: `/groups/:id`
- **Auth Required**: Yes (Admin only)
- **Description**: Update group information
- **Request Body**:
  ```json
  {
    "name": "New Name",
    "description": "New description"
  }
  ```
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "message": "Group updated successfully"
    }
  }
  ```
- **Error Responses**:
  - `403`: Not authorized (not admin)

---

## Group Membership Endpoints

### Invite Member
- **Method**: `POST`
- **Path**: `/groups/:id/invite`
- **Auth Required**: Yes (Admin only)
- **Description**: Invite a user to join the group
- **Request Body**:
  ```json
  {
    "user_id": 123
  }
  ```
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "message": "Invitation sent successfully"
    }
  }
  ```
- **Behavior**: Creates a notification for the invited user

---

### Request to Join
- **Method**: `POST`
- **Path**: `/groups/:id/request`
- **Auth Required**: Yes
- **Description**: Request to join a group
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "message": "Join request sent successfully"
    }
  }
  ```
- **Behavior**: Creates a notification for group admins

---

### Get Pending Requests
- **Method**: `GET`
- **Path**: `/groups/:id/requests`
- **Auth Required**: Yes (Admin only)
- **Description**: Get pending join requests for a group
- **Success Response**:
  ```json
  {
    "success": true,
    "data": [
      {
        "user_id": 789,
        "username": "newuser",
        "first_name": "Bob",
        "last_name": "Jones",
        "avatar_url": "https://...",
        "requested_at": "2025-11-08T10:00:00Z"
      }
    ]
  }
  ```

---

### Respond to Request
- **Method**: `POST`
- **Path**: `/groups/:id/requests/respond`
- **Auth Required**: Yes (Admin only)
- **Description**: Accept or reject a join request
- **Request Body**:
  ```json
  {
    "member_id": 123,
    "accept": true
  }
  ```
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "message": "Request accepted"
    }
  }
  ```

---

### Get Group Members
- **Method**: `GET`
- **Path**: `/groups/:id/members`
- **Auth Required**: Yes
- **Description**: Get list of group members
- **Success Response**:
  ```json
  {
    "success": true,
    "data": [
      {
        "user_id": 456,
        "username": "johndoe",
        "first_name": "John",
        "last_name": "Doe",
        "avatar_url": "https://...",
        "is_admin": true,
        "joined_at": "2025-11-08T10:00:00Z"
      }
    ]
  }
  ```

---

## Group Messaging Endpoints

### Create Group Message
- **Method**: `POST`
- **Path**: `/groups/:id/messages`
- **Auth Required**: Yes (Members only)
- **Description**: Post a message in group
- **Request Body**:
  ```json
  {
    "content": "Message content"
  }
  ```
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "id": 789,
      "group_id": 123,
      "sender_id": 456,
      "content": "Message content",
      "created_at": "2025-11-08T10:00:00Z"
    }
  }
  ```

---

### Get Group Messages
- **Method**: `GET`
- **Path**: `/groups/:id/messages`
- **Auth Required**: Yes (Members only)
- **Description**: Get group messages
- **Query Parameters**:
  - `limit`: Number of messages to retrieve (default: 50)
- **Success Response**:
  ```json
  {
    "success": true,
    "data": [
      {
        "id": 789,
        "group_id": 123,
        "sender_id": 456,
        "sender_username": "johndoe",
        "content": "Message content",
        "created_at": "2025-11-08T10:00:00Z"
      }
    ]
  }
  ```

---

## Event Endpoints

### Create Event
- **Method**: `POST`
- **Path**: `/events`
- **Auth Required**: Yes (Group members only)
- **Description**: Create a group event
- **Request Body**:
  ```json
  {
    "group_id": 123,
    "title": "Event Title",
    "description": "Event description",
    "event_time": "2025-12-31T18:00:00Z"
  }
  ```
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "id": 456,
      "group_id": 123,
      "creator_id": 789,
      "title": "Event Title",
      "description": "Event description",
      "event_time": "2025-12-31T18:00:00Z",
      "created_at": "2025-11-08T10:00:00Z"
    }
  }
  ```
- **Behavior**: Creates notifications for all group members

---

### Get Event
- **Method**: `GET`
- **Path**: `/events/:id`
- **Auth Required**: Yes (Group members only)
- **Description**: Get event details
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "id": 456,
      "group_id": 123,
      "creator_id": 789,
      "title": "Event Title",
      "description": "Event description",
      "event_time": "2025-12-31T18:00:00Z",
      "created_at": "2025-11-08T10:00:00Z",
      "responses": [
        {
          "user_id": 111,
          "username": "user1",
          "response": "going"
        }
      ]
    }
  }
  ```

---

### Get Group Events
- **Method**: `GET`
- **Path**: `/groups/:id/events`
- **Auth Required**: Yes (Members only)
- **Description**: Get all events for a specific group
- **Success Response**:
  ```json
  {
    "success": true,
    "data": [
      {
        "id": 456,
        "group_id": 123,
        "title": "Event Title",
        "description": "Event description",
        "event_time": "2025-12-31T18:00:00Z",
        "created_at": "2025-11-08T10:00:00Z"
      }
    ]
  }
  ```

---

### Respond to Event
- **Method**: `POST`
- **Path**: `/events/:id/respond`
- **Auth Required**: Yes (Group members only)
- **Description**: RSVP to an event
- **Request Body**:
  ```json
  {
    "event_id": 123,
    "response": "going"
  }
  ```
- **Response Options**:
  - `"going"`: Attending
  - `"not_going"`: Not attending
  - `"interested"`: Interested but not confirmed
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "message": "Response recorded successfully"
    }
  }
  ```

---

### Health Check
- **Method**: `GET`
- **Path**: `/health`
- **Description**: Service health status
- **Success Response**:
  ```json
  {
    "status": "healthy"
  }
  ```

---

## Permissions

### Admin Permissions
- Update group information
- Invite members
- Accept/reject join requests
- All member permissions

### Member Permissions
- View group details
- View and send group messages
- Create events
- Respond to events
- View other members

---

## Authentication

All endpoints (except `/health`) require authentication:

```
Authorization: Bearer <token>
```

---

## Notifications

The Groups Service triggers notifications for:
- Group invitations
- Join requests (to admins)
- Join request responses (to requester)
- New events (to all members)

---

**Service Port**: 8084  
**Last Updated**: November 8, 2025

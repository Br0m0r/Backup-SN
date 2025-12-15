# Chat Service API Documentation

Base URL: `http://localhost:8085`

## Overview

The Chat Service provides real-time messaging capabilities using WebSockets for instant communication and HTTP endpoints as fallback. Supports both direct messages and group chats.

---

## WebSocket Endpoint

### Connect to Chat WebSocket
- **Protocol**: WebSocket
- **Path**: `/ws`
- **Auth Required**: Yes
- **Description**: Real-time bidirectional chat communication

#### Authentication
Include token via query parameter or header:
```
ws://localhost:8085/ws?token=<bearer_token>
```
Or via header:
```
Authorization: Bearer <token>
```

#### Message Types

**Outgoing (Client → Server)**:
```json
{
  "type": "message",
  "receiver_id": 123,
  "content": "Hello!"
}
```

```json
{
  "type": "group_message",
  "group_id": 456,
  "content": "Hello group!"
}
```

```json
{
  "type": "typing",
  "receiver_id": 123
}
```

**Incoming (Server → Client)**:
```json
{
  "type": "message",
  "message_id": 789,
  "sender_id": 123,
  "receiver_id": 456,
  "content": "Hello!",
  "timestamp": "2025-11-08T10:00:00Z"
}
```

```json
{
  "type": "group_message",
  "message_id": 789,
  "sender_id": 123,
  "group_id": 456,
  "content": "Hello group!",
  "timestamp": "2025-11-08T10:00:00Z"
}
```

```json
{
  "type": "read_receipt",
  "message_id": 789,
  "reader_id": 123,
  "read_at": "2025-11-08T10:05:00Z"
}
```

---

## REST Endpoints (HTTP Fallback)

### Get Conversations
- **Method**: `GET`
- **Path**: `/chat/conversations`
- **Auth Required**: Yes
- **Description**: Get list of all conversations with last message and online status
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "conversations": [
        {
          "user_id": 123,
          "username": "johndoe",
          "first_name": "John",
          "last_name": "Doe",
          "avatar_url": "https://...",
          "last_message": "Hello!",
          "last_message_time": "2025-11-08T10:00:00Z",
          "unread_count": 3,
          "is_online": true
        }
      ],
      "count": 1
    }
  }
  ```

---

### Get Chat History
- **Method**: `GET`
- **Path**: `/chat/history/:userId`
- **Auth Required**: Yes
- **Description**: Get message history with a specific user
- **Query Parameters**:
  - `limit`: Number of messages (default: 50, max: 200)
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "messages": [
        {
          "id": 789,
          "sender_id": 123,
          "receiver_id": 456,
          "content": "Hello!",
          "is_read": true,
          "created_at": "2025-11-08T10:00:00Z"
        }
      ],
      "count": 1
    }
  }
  ```
- **Permission**: Only users who are mutual followers can access chat history

---

### Send Message
- **Method**: `POST`
- **Path**: `/chat/send`
- **Auth Required**: Yes
- **Description**: Send a message via HTTP (alternative to WebSocket)
- **Request Body**:
  ```json
  {
    "receiver_id": 123,
    "content": "Message text"
  }
  ```
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "message": {
        "id": 789,
        "sender_id": 456,
        "receiver_id": 123,
        "content": "Message text",
        "is_read": false,
        "created_at": "2025-11-08T10:00:00Z"
      }
    }
  }
  ```
- **Error Responses**:
  - `403`: Cannot send messages to this user (not mutual followers)

---

### Mark Messages as Read
- **Method**: `POST`
- **Path**: `/chat/read/:userId`
- **Auth Required**: Yes
- **Description**: Mark all messages from a user as read
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "message": "Messages marked as read"
    }
  }
  ```

---

### Get Unread Count
- **Method**: `GET`
- **Path**: `/chat/unread`
- **Auth Required**: Yes
- **Description**: Get total count of unread messages
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

## Group Chat Endpoints

### Get Group Chat History
- **Method**: `GET`
- **Path**: `/chat/groups/:groupId/history`
- **Auth Required**: Yes (Members only)
- **Description**: Get message history for a group
- **Query Parameters**:
  - `limit`: Number of messages (default: 50, max: 200)
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "messages": [
        {
          "id": 789,
          "group_id": 123,
          "sender_id": 456,
          "sender_username": "johndoe",
          "content": "Hello group!",
          "created_at": "2025-11-08T10:00:00Z"
        }
      ],
      "count": 1
    }
  }
  ```

---

### Send Group Message
- **Method**: `POST`
- **Path**: `/chat/groups/:groupId/messages`
- **Auth Required**: Yes (Members only)
- **Description**: Send a message to group via HTTP
- **Request Body**:
  ```json
  {
    "content": "Message text"
  }
  ```
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "message": {
        "id": 789,
        "group_id": 123,
        "sender_id": 456,
        "content": "Message text",
        "created_at": "2025-11-08T10:00:00Z"
      }
    }
  }
  ```
- **Error Responses**:
  - `403`: Not a member of this group

---

### Health Check
- **Method**: `GET`
- **Path**: `/health`
- **Description**: Service health status
- **Success Response**:
  ```json
  {
    "status": "healthy",
    "service": "chat-service"
  }
  ```

---

## Chat Permissions

### Direct Messages
Users can chat if:
- They are mutual followers (both follow each other)

### Group Messages
Users can chat if:
- They are members of the group

---

## WebSocket Connection Management

### Connection Lifecycle
1. **Connect**: Client establishes WebSocket connection with auth token
2. **Active**: Client can send/receive messages in real-time
3. **Disconnect**: Connection closes (manual or timeout)

### Online Status
- Users appear online when WebSocket is connected
- Online status visible in conversations list
- Status updates broadcast to relevant conversations

### Heartbeat/Ping
WebSocket connections include automatic ping/pong to maintain connection.

---

## Authentication

All endpoints require authentication:

```
Authorization: Bearer <token>
```

For WebSocket:
```
ws://localhost:8085/ws?token=<bearer_token>
```

---

## Message Delivery

### Delivery Guarantees
- **Online Users**: Messages delivered immediately via WebSocket
- **Offline Users**: Messages stored in database, delivered on next connection
- **HTTP Fallback**: All messages accessible via REST endpoints

### Read Receipts
- Automatic read receipt when user marks messages as read
- Visible to sender via WebSocket `read_receipt` message

---

## Rate Limiting

No explicit rate limiting on chat endpoints, but WebSocket connections are monitored for abuse.

---

## Error Handling

### WebSocket Errors
```json
{
  "type": "error",
  "message": "Error description",
  "code": "ERROR_CODE"
}
```

### HTTP Errors
Standard HTTP error codes with JSON response:
```json
{
  "success": false,
  "error": "Error message"
}
```

---

**Service Port**: 8085  
**Last Updated**: November 8, 2025

# Auth Service API Documentation

Base URL: `http://localhost:8081`

## Overview

The Auth Service handles user authentication, registration, session management, and token verification for the entire social network platform.

---

## Public Endpoints

### Register User
- **Method**: `POST`
- **Path**: `/register`
- **Auth Required**: No
- **Rate Limited**: Yes
- **Description**: Create a new user account
- **Request Body**:
  ```json
  {
    "email": "user@example.com",
    "username": "username",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe",
    "date_of_birth": "1990-01-01"
  }
  ```
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "user": { ... },
      "token": "session_token_here"
    }
  }
  ```
- **Error Responses**:
  - `400`: Invalid request body or validation error
  - `409`: Username or email already exists

---

### Login
- **Method**: `POST`
- **Path**: `/login`
- **Auth Required**: No
- **Rate Limited**: Yes
- **Description**: Authenticate user and create session
- **Request Body**:
  ```json
  {
    "identifier": "username or email",
    "password": "password123"
  }
  ```
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "user": { ... },
      "token": "session_token_here"
    }
  }
  ```
- **Error Responses**:
  - `400`: Invalid request body
  - `401`: Invalid credentials

---

### Logout
- **Method**: `POST`
- **Path**: `/logout`
- **Auth Required**: Yes (Bearer token)
- **Description**: Invalidate current session token
- **Headers**:
  ```
  Authorization: Bearer <token>
  ```
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "message": "Logged out successfully"
    }
  }
  ```
- **Error Responses**:
  - `400`: Missing Authorization header
  - `500`: Server error

---

### Get Session
- **Method**: `GET`
- **Path**: `/session`
- **Auth Required**: Yes (via session cookie or header)
- **Description**: Retrieve current session information
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "user_id": 123,
      "username": "username",
      "session_id": "..."
    }
  }
  ```

---

## Internal Endpoints (Service-to-Service)

### Verify Token
- **Method**: `GET`
- **Path**: `/internal/verify-token`
- **Description**: Verify if a token is valid (used by other services)
- **Headers**:
  ```
  Authorization: Bearer <token>
  ```
- **Success Response**:
  ```json
  {
    "valid": true,
    "user_id": 123,
    "username": "username"
  }
  ```

---

### Get User By ID
- **Method**: `GET`
- **Path**: `/internal/user/:id`
- **Description**: Get user information by ID (internal use)
- **Success Response**:
  ```json
  {
    "id": 123,
    "username": "username",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe"
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

## Authentication

Tokens are returned from login/register endpoints and should be included in subsequent requests:

```
Authorization: Bearer <token>
```

Session tokens are stored in the database and validated on each request.

---

## Rate Limiting

- `/register`: Limited to prevent spam accounts
- `/login`: Limited to prevent brute force attacks

Default: 100 requests per minute per IP address

---

## Error Response Format

```json
{
  "success": false,
  "error": "Error message description"
}
```

---

**Service Port**: 8081  
**Last Updated**: November 8, 2025

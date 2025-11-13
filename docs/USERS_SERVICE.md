# Users Service API Documentation

Base URL: `http://localhost:8082`

## Overview

The Users Service manages user profiles, follow relationships, user search, and privacy settings.

---

## Profile Endpoints

### Get Current User Profile
- **Method**: `GET`
- **Path**: `/profile`
- **Auth Required**: Yes
- **Description**: Get authenticated user's full profile (includes email, DOB)
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "user": {
        "id": 123,
        "username": "johndoe",
        "email": "john@example.com",
        "first_name": "John",
        "last_name": "Doe",
        "date_of_birth": "1990-01-01",
        "nickname": "JD",
        "about_me": "Bio text",
        "avatar_url": "https://...",
        "is_public_profile": true
      }
    }
  }
  ```

---

### Get User Profile by ID
- **Method**: `GET`
- **Path**: `/profile/:id`
- **Auth Required**: Yes
- **Description**: Get another user's public profile (no email, no DOB)
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "user": {
        "id": 456,
        "username": "janedoe",
        "first_name": "Jane",
        "last_name": "Doe",
        "nickname": "Jane",
        "about_me": "Bio text",
        "avatar_url": "https://...",
        "is_public_profile": true
      }
    }
  }
  ```

---

### Get Comprehensive User Profile
- **Method**: `GET`
- **Path**: `/users/:id/profile`
- **Auth Required**: Yes
- **Description**: Get detailed profile including posts, followers, following counts. Respects privacy settings.
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "profile": {
        "user": { ... },
        "post_count": 42,
        "follower_count": 150,
        "following_count": 200,
        "can_view": true,
        "is_following": true,
        "follow_status": "accepted"
      }
    }
  }
  ```
- **Privacy Note**: If `can_view` is false, limited information is returned with a message about private profile

---

### Update Profile
- **Method**: `PUT`
- **Path**: `/profile`
- **Auth Required**: Yes
- **Description**: Update current user's profile information
- **Request Body**:
  ```json
  {
    "nickname": "Cool User",
    "about_me": "Bio text here",
    "is_public_profile": true,
    "avatar_url": "https://..."
  }
  ```
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "user": { ... },
      "message": "Profile updated successfully"
    }
  }
  ```

---

## Follow Endpoints

### Follow User
- **Method**: `POST`
- **Path**: `/follow`
- **Auth Required**: Yes
- **Rate Limited**: Yes
- **Description**: Follow a user or send follow request if profile is private
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
      "message": "Follow request sent successfully"
    }
  }
  ```
- **Notes**:
  - Public profiles: Instant follow
  - Private profiles: Creates pending request

---

### Unfollow User
- **Method**: `DELETE`
- **Path**: `/follow`
- **Auth Required**: Yes
- **Rate Limited**: Yes
- **Description**: Unfollow a user
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
      "message": "Unfollowed successfully"
    }
  }
  ```

---

### Get Followers
- **Method**: `GET`
- **Path**: `/followers`
- **Auth Required**: Yes
- **Description**: Get list of users following you
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "followers": [
        {
          "id": 123,
          "username": "user1",
          "first_name": "John",
          "last_name": "Doe",
          "avatar_url": "https://..."
        }
      ],
      "count": 1
    }
  }
  ```

---

### Get Following
- **Method**: `GET`
- **Path**: `/following`
- **Auth Required**: Yes
- **Description**: Get list of users you're following
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "following": [
        {
          "id": 456,
          "username": "user2",
          "first_name": "Jane",
          "last_name": "Smith",
          "avatar_url": "https://..."
        }
      ],
      "count": 1
    }
  }
  ```

---

### Get Follow Status
- **Method**: `GET`
- **Path**: `/follow/status/:id`
- **Auth Required**: Yes
- **Description**: Check follow status with another user
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "status": "accepted"
    }
  }
  ```
- **Status Values**:
  - `"none"`: Not following
  - `"pending"`: Follow request sent
  - `"accepted"`: Following

---

### Get Pending Follow Requests
- **Method**: `GET`
- **Path**: `/follow/requests`
- **Auth Required**: Yes
- **Description**: Get list of pending follow requests to your account
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "requests": [
        {
          "id": 789,
          "username": "newuser",
          "first_name": "Bob",
          "last_name": "Jones",
          "avatar_url": "https://...",
          "requested_at": "2025-11-08T10:00:00Z"
        }
      ],
      "count": 1
    }
  }
  ```

---

### Respond to Follow Request
- **Method**: `POST`
- **Path**: `/follow/respond`
- **Auth Required**: Yes
- **Description**: Accept or reject a follow request
- **Request Body**:
  ```json
  {
    "follower_id": 123,
    "accept": true
  }
  ```
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "message": "Follow request accepted"
    }
  }
  ```

---

## Search Endpoints

### Search Users
- **Method**: `GET`
- **Path**: `/search?q=searchterm`
- **Auth Required**: Yes
- **Description**: Search for users by username, first name, or last name
- **Query Parameters**:
  - `q`: Search term (required)
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "users": [
        {
          "id": 123,
          "username": "johndoe",
          "first_name": "John",
          "last_name": "Doe",
          "avatar_url": "https://..."
        }
      ],
      "count": 1
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
    "status": "healthy",
    "service": "user",
    "message": "User service is running"
  }
  ```

---

## Privacy Settings

- **Public Profile**: Anyone can follow and view content
- **Private Profile**: Follow requests must be accepted before viewing content

---

## Authentication

All endpoints (except `/health`) require authentication:

```
Authorization: Bearer <token>
```

---

## Rate Limiting

Follow/Unfollow endpoints are rate limited to prevent spam:
- Limit: 100 requests per minute per user

---

**Service Port**: 8082  
**Last Updated**: November 8, 2025

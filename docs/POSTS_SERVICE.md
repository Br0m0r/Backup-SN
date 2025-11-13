# Posts Service API Documentation

Base URL: `http://localhost:8083`

## Overview

The Posts Service manages post creation, retrieval, updates, deletion, and comments. It supports privacy controls including public, private, and custom viewer lists.

---

## Post Endpoints

### Create Post
- **Method**: `POST`
- **Path**: `/posts`
- **Auth Required**: Yes
- **Rate Limited**: Yes
- **Description**: Create a new post
- **Request Body**:
  ```json
  {
    "title": "Post title",
    "content": "Post content",
    "image_url": "https://...",
    "privacy": "public",
    "allowed_users": [1, 2, 3]
  }
  ```
- **Privacy Options**:
  - `"public"`: Visible to all users
  - `"private"`: Visible only to followers
  - `"almost_private"`: Visible only to users in `allowed_users` array
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "post": {
        "id": 123,
        "user_id": 456,
        "title": "Post title",
        "content": "Post content",
        "image_url": "https://...",
        "privacy": "public",
        "created_at": "2025-11-08T10:00:00Z",
        "updated_at": "2025-11-08T10:00:00Z"
      }
    }
  }
  ```

---

### Get Feed
- **Method**: `GET`
- **Path**: `/posts`
- **Auth Required**: Yes
- **Rate Limited**: Yes
- **Description**: Get personalized feed of posts (from followed users and public posts)
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "posts": [
        {
          "id": 123,
          "user_id": 456,
          "username": "johndoe",
          "title": "Post title",
          "content": "Post content",
          "image_url": "https://...",
          "privacy": "public",
          "created_at": "2025-11-08T10:00:00Z",
          "comment_count": 5
        }
      ]
    }
  }
  ```
- **Feed Logic**:
  - Posts from users you follow
  - Public posts from all users
  - Posts where you're in the `allowed_users` list
  - Respects privacy settings

---

### Get Single Post
- **Method**: `GET`
- **Path**: `/posts/:id`
- **Auth Required**: Yes
- **Description**: Get a specific post by ID (respects privacy settings)
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "post": {
        "id": 123,
        "user_id": 456,
        "username": "johndoe",
        "title": "Post title",
        "content": "Post content",
        "image_url": "https://...",
        "privacy": "public",
        "created_at": "2025-11-08T10:00:00Z",
        "updated_at": "2025-11-08T10:00:00Z",
        "comment_count": 5
      }
    }
  }
  ```
- **Error Responses**:
  - `403`: Access denied (privacy restriction)
  - `404`: Post not found

---

### Update Post
- **Method**: `PUT`
- **Path**: `/posts/:id`
- **Auth Required**: Yes
- **Description**: Update an existing post (must be post owner)
- **Request Body**:
  ```json
  {
    "title": "Updated title",
    "content": "Updated content",
    "image_url": "https://...",
    "privacy": "public"
  }
  ```
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "post": { ... }
    }
  }
  ```
- **Error Responses**:
  - `403`: Unauthorized (not post owner)
  - `404`: Post not found

---

### Delete Post
- **Method**: `DELETE`
- **Path**: `/posts/:id`
- **Auth Required**: Yes
- **Description**: Delete a post (must be post owner)
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "message": "Post deleted successfully"
    }
  }
  ```
- **Error Responses**:
  - `403`: Unauthorized (not post owner)
  - `404`: Post not found

---

## Comment Endpoints

### Create Comment
- **Method**: `POST`
- **Path**: `/comments`
- **Auth Required**: Yes
- **Rate Limited**: Yes
- **Description**: Add a comment to a post
- **Request Body**:
  ```json
  {
    "post_id": 123,
    "content": "Comment text"
  }
  ```
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "comment": {
        "id": 789,
        "post_id": 123,
        "user_id": 456,
        "username": "johndoe",
        "content": "Comment text",
        "created_at": "2025-11-08T10:00:00Z"
      }
    }
  }
  ```
- **Error Responses**:
  - `403`: Access denied (cannot comment on this post due to privacy)
  - `404`: Post not found

---

### Get Comments
- **Method**: `GET`
- **Path**: `/comments?post_id=123`
- **Auth Required**: Yes
- **Rate Limited**: Yes
- **Description**: Get all comments for a specific post
- **Query Parameters**:
  - `post_id`: Post ID (required)
- **Success Response**:
  ```json
  {
    "success": true,
    "data": {
      "comments": [
        {
          "id": 789,
          "post_id": 123,
          "user_id": 456,
          "username": "johndoe",
          "content": "Comment text",
          "created_at": "2025-11-08T10:00:00Z"
        }
      ]
    }
  }
  ```
- **Error Responses**:
  - `403`: Access denied (cannot view comments)
  - `404`: Post not found

---

### Health Check
- **Method**: `GET`
- **Path**: `/health`
- **Description**: Service health status
- **Success Response**:
  ```json
  {
    "status": "healthy",
    "service": "post",
    "message": "Post service is running"
  }
  ```

---

## Privacy Behavior

### Post Visibility Rules

1. **Public Posts**: Visible to all authenticated users
2. **Private Posts**: Visible only to:
   - Post owner
   - Users following the post owner
3. **Almost Private Posts**: Visible only to:
   - Post owner
   - Users in the `allowed_users` list

### Comment Access Rules

Users can comment on a post if they can view it.

---

## Authentication

All endpoints (except `/health`) require authentication:

```
Authorization: Bearer <token>
```

---

## Rate Limiting

The following endpoints are rate limited:
- `POST /posts`: 100 requests per minute
- `POST /comments`: 100 requests per minute
- `GET /posts`: 100 requests per minute
- `GET /comments`: 100 requests per minute

---

## Notifications

The Posts Service triggers notifications to the Notifications Service when:
- A user comments on a post
- A followed user creates a new post

---

**Service Port**: 8083  
**Last Updated**: November 8, 2025

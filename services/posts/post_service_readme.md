# Post Service

Post microservice for social network - handles posts, comments, and privacy controls.

## Port
- **8083**

## Dependencies
- **Auth Service** (`http://auth-service:8081`) - Token verification via `/internal/verify-token`
- **Database** - Shared SQLite database at `./social_network.db`

## Database Tables
- `posts` - User posts with privacy levels
- `comments` - Comments on posts
- `post_viewers` - Users who can view "almost_private" posts

## Endpoints

### Posts
- `POST /posts` - Create new post (auth required, rate limited)
- `GET /posts` - Get user feed (auth required)
- `GET /posts/:id` - Get single post (auth required, privacy check)
- `PUT /posts/:id` - Update post (auth required, owner only)
- `DELETE /posts/:id` - Delete post (auth required, owner only)

### Comments
- `POST /comments` - Create comment (auth required, rate limited)
- `GET /comments?post_id=:id` - Get comments for a post (auth required, privacy check)

## Privacy Levels
- **public** - Anyone can see
- **private** - Only followers (with accepted status) can see
- **almost_private** - Only specific users (in post_viewers table) can see

## Middleware Chain
1. **CORS** - Handles cross-origin requests
2. **Logging** - Logs all requests with timing
3. **Auth** - Validates token with auth-service (all routes protected)
4. **RateLimit** - Applied to POST /posts and POST /comments (10 req/sec per IP)

## Connection Pool Settings
- MaxOpenConns: 25
- MaxIdleConns: 5

## Request/Response Examples

### Create Post
```json
POST /posts
{
  "content": "Hello world!",
  "image_path": "/uploads/image.jpg",
  "privacy_level": "public",
  "viewers": [] // Only for "almost_private" posts
}

Response:
{
  "success": true,
  "data": {
    "post": {
      "id": 1,
      "user_id": 1,
      "content": "Hello world!",
      "image_path": "/uploads/image.jpg",
      "privacy_level": "public",
      "created_at": "2025-10-10T12:00:00Z"
    }
  }
}
```

### Get Feed
```json
GET /posts

Response:
{
  "success": true,
  "data": {
    "posts": [...]
  }
}
```

### Create Comment
```json
POST /comments
{
  "post_id": 1,
  "content": "Great post!",
  "image_path": null
}

Response:
{
  "success": true,
  "data": {
    "comment": {
      "id": 1,
      "post_id": 1,
      "user_id": 2,
      "content": "Great post!",
      "created_at": "2025-10-10T12:05:00Z"
    }
  }
}
```

## Access Control Logic
The feed query (`GetFeedPosts`) returns posts where:
- Post is public, OR
- User is the post owner, OR
- Post is "almost_private" AND user is in post_viewers table, OR
- Post is "private" AND user follows the post owner (status = 'accepted')

## Notes
- All endpoints require authentication token in `Authorization: Bearer <token>` header
- Rate limiting prevents spam (post/comment creation)
- Cascade delete: Deleting a post also deletes its comments and viewers
- Update post allows changing privacy level and updating viewers list

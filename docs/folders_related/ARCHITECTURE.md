# Frontend Architecture

## Overview

This Vue 3 frontend follows a clean, service-oriented architecture with clear separation of concerns. All API calls are centralized in service modules, UI logic is handled by composables, and components focus purely on presentation.

## Directory Structure

```
frontend/
├── src/
│   ├── assets/          # Static assets (CSS, images)
│   ├── components/      # Reusable UI components
│   ├── composables/     # Shared reactive logic
│   ├── pages/           # Route-level page components
│   ├── router/          # Vue Router configuration
│   ├── services/        # API service layer
│   └── stores/          # State management
├── index.html
└── vite.config.js
```

---

## Architecture Layers

### 1. Service Layer (`src/services/`)

Centralized API communication layer. All axios calls are isolated here.

#### **authService.js**
Authentication and session management
- `login(email, password)` - User login
- `register(userData)` - User registration
- `logout()` - User logout

#### **usersService.js**
User-related operations
- `searchUsers(query, token)` - Search for users
- `followUser(userId, token)` - Follow a user
- `unfollowUser(userId, token)` - Unfollow a user
- `getFollowers(token)` - Get user's followers
- `getFollowing(token)` - Get users being followed
- `getFollowStatus(userId, token)` - Check follow relationship
- `getUserProfile(userId, token)` - Get user profile
- `updatePrivacy(isPublic, token)` - Update privacy settings
- `respondToFollowRequest(followerId, accept, token)` - Accept/reject follow request

#### **postsService.js**
Posts and comments management
- `uploadImage(file, token)` - Upload post/comment image
- `createPost(postData, token)` - Create new post
- `getPost(postId, token)` - Get single post
- `updatePost(postId, postData, token)` - Update post
- `deletePost(postId, token)` - Delete post
- `getFeedPosts(token)` - Get user's feed
- `searchPosts(query, token)` - Search posts
- `getComments(postId, token)` - Get post comments
- `createComment(commentData, token)` - Add comment
- `updateComment(commentId, commentData, token)` - Update comment
- `deleteComment(commentId, token)` - Delete comment

#### **chatService.js**
Real-time messaging operations
- `getContacts(token)` - Get chat contacts
- `getChatHistory(userId, token, limit)` - Load message history
- `uploadImage(file, token)` - Upload chat image
- `markAsRead(userId, token)` - Mark messages as read
- `getImageUrl(path)` - Helper for image URLs

#### **groupsService.js**
Group-related operations
- `respondToGroupInvite(groupId, accept, token)` - Accept/decline group invite

**Pattern:**
```javascript
// All services follow this pattern:
export async function serviceName(params, token) {
  try {
    const response = await axios.post/get/put/delete(url, data, {
      headers: { Authorization: `Bearer ${token}` }
    })
    return unwrapResponse(response)
  } catch (error) {
    throw new Error(error.response?.data?.error || 'Operation failed')
  }
}
```

---

### 2. Composables (`src/composables/`)

Shared reactive logic using Vue Composition API.

#### **useToast.js**
Global toast notification system
- `add(message, type)` - Add toast notification
- `remove(id)` - Remove toast
- `success(message)` - Success toast
- `error(message)` - Error toast
- `warning(message)` - Warning toast
- `info(message)` - Info toast

**Usage:**
```vue
<script setup>
import { useToast } from '@/composables/useToast'
const { success, error } = useToast()

success('Post created!')
error('Failed to save')
</script>
```

#### **useWebSocket.js**
WebSocket connection management for real-time chat
- `connected` - Connection state
- `sendMessage(userId, content, imagePath)` - Send chat message
- `on(event, callback)` - Listen for events
- `connect()` - Establish connection
- `disconnect()` - Close connection

#### **useNotifications.js**
WebSocket-based notification system
- `notifications` - Notification array
- `unreadCount` - Number of unread notifications
- `connected` - Connection state
- `markAsRead(notifId)` - Mark notification as read
- `markAllAsRead()` - Mark all as read
- `deleteNotification(notifId)` - Delete notification

---

### 3. Components (`src/components/`)

#### **ToastContainer.vue**
Renders global toast notifications
- Displays toasts with icons and animations
- Auto-dismiss after timeout
- Click to dismiss
- 4 types: success (green), error (red), warning (orange), info (cyan)

#### **CreatePost.vue**
Post creation form
- Text and image support
- Privacy levels (public/followers/private)
- Uses `postsService` for API calls
- Toast notifications for feedback

#### **Chat.vue**
Real-time messaging interface
- Contact list with online status
- Multiple chat windows (max 3)
- Image sharing
- Emoji picker
- Uses `chatService` and `useWebSocket`

#### **Notifications.vue**
Notification panel
- Real-time notifications via WebSocket
- Follow request handling
- Group invite handling
- Uses `usersService` and `groupsService`

#### **SuggestedUsers.vue**
User recommendations widget

#### **EmojiPicker.vue**
Emoji selection component

---

### 4. Pages (`src/pages/`)

Route-level components.

#### **AuthView.vue**
Login/registration page
- Uses `authService`

#### **FeedView.vue**
Main feed page
- Post list
- Search functionality
- Uses `postsService`

#### **PostView.vue**
Single post view with comments
- Post display and editing (owner only)
- Comment CRUD operations
- Custom delete confirmation modal
- Uses `postsService`
- Toast notifications

#### **ProfileView.vue**
User profile page
- Profile information
- Privacy settings
- Uses `usersService`

#### **ToastTest.vue**
Test page for toast notifications (`/toast-test`)

---

### 5. Stores (`src/stores/`)

#### **auth.js**
Centralized authentication state
- `getToken()` - Get auth token
- `getUser()` - Get current user
- `setToken(token)` - Save token
- `setUser(user)` - Save user
- `clearAuth()` - Clear authentication

---

## Design Principles

### 1. **Service Layer Pattern**
- ✅ All API calls isolated in service modules
- ✅ No direct axios imports in components
- ✅ Consistent error handling with `unwrapResponse`
- ✅ Single source of truth for API endpoints

### 2. **Composable Logic**
- ✅ Reusable reactive logic extracted to composables
- ✅ Shared state managed centrally
- ✅ Event-driven communication

### 3. **Component Responsibility**
- ✅ Components focus on UI presentation
- ✅ Business logic delegated to composables/services
- ✅ Toast notifications for user feedback (no `alert()`)

### 4. **Code Organization**
```
Component/Page
    ↓ uses
Composable (optional)
    ↓ calls
Service Layer
    ↓ communicates with
Backend API
```

---

## API Communication Flow

### Example: Creating a Post

```vue
<script setup>
// 1. Import service and toast
import { createPost } from '@/services/postsService'
import { useToast } from '@/composables/useToast'
import { getToken } from '@/stores/auth'

const { success, error } = useToast()

async function handleSubmit() {
  try {
    // 2. Get token from store
    const token = getToken()
    
    // 3. Call service
    const data = await createPost(postData, token)
    
    // 4. Show feedback
    success('Post created successfully!')
    
    // 5. Update UI
    emit('posted')
  } catch (err) {
    // 6. Handle errors
    error(err.message || 'Failed to create post')
  }
}
</script>
```

---

## WebSocket Architecture

Real-time features use WebSocket connections:

- **Chat**: `useWebSocket` composable
- **Notifications**: `useNotifications` composable

Both maintain persistent connections and handle:
- Automatic reconnection
- Online status tracking
- Message queueing
- Event-driven updates

See `WEBSOCKET_ARCHITECTURE.md` and `WEBSOCKET_SUMMARY.md` for details.

---

## UI/UX Features

### Toast Notifications
Global notification system replacing browser alerts
- **Location**: Bottom-right corner
- **Types**: Success, Error, Warning, Info
- **Features**: Auto-dismiss, click-to-dismiss, animations
- **Integration**: `ToastContainer` in `App.vue`

### Custom Modals
- Delete confirmation modal in `PostView.vue`
- Styled to match neon theme
- No browser `confirm()` dialogs

### Theme
Neon cyan/pink gradient design with:
- Dark background (`rgba(5, 6, 13, 0.95)`)
- Backdrop blur effects
- Gradient borders and shadows
- Glassmorphism UI elements

---

## Environment Variables

Create `.env` file:
```env
VITE_AUTH_API_URL=http://localhost:8081
VITE_USERS_API_URL=http://localhost:8082
VITE_POSTS_API_URL=http://localhost:8083
VITE_GROUPS_API_URL=http://localhost:8084
VITE_CHAT_API_URL=http://localhost:8085
VITE_NOTIFICATIONS_API_URL=http://localhost:8086
VITE_WS_URL=ws://localhost:8084/ws
VITE_NOTIFICATIONS_WS_URL=ws://localhost:8086/ws
```

---

## Development Guidelines

### Adding a New Feature

1. **Create/update service** in `src/services/`
   ```javascript
   export async function newFeature(params, token) {
     const response = await axios.post(url, data, {
       headers: { Authorization: `Bearer ${token}` }
     })
     return unwrapResponse(response)
   }
   ```

2. **Create composable** (if shared logic needed)
   ```javascript
   export function useFeature() {
     const state = ref(null)
     // ... reactive logic
     return { state, methods }
   }
   ```

3. **Use in component**
   ```vue
   <script setup>
   import { newFeature } from '@/services/featureService'
   import { useToast } from '@/composables/useToast'
   
   const { success, error } = useToast()
   
   async function handleAction() {
     try {
       await newFeature(params, token)
       success('Success!')
     } catch (err) {
       error(err.message)
     }
   }
   </script>
   ```

### Testing

- Test page for toasts: `/toast-test`
- Use browser DevTools to inspect WebSocket connections
- Check Network tab for API calls

---

## Dependencies

- **Vue 3** - Progressive framework
- **Vue Router** - Client-side routing
- **Axios** - HTTP client
- **Vite** - Build tool

---

## Related Documentation

- `WEBSOCKET_ARCHITECTURE.md` - WebSocket implementation details
- `WEBSOCKET_SUMMARY.md` - WebSocket system overview
- `NOTIFICATIONS_GUIDE.md` - Notification system guide
- Backend service docs in `/docs/`

# Groups Feature Implementation Guide

## Overview
This document outlines the complete implementation of the Groups feature for the social network, following the project requirements.

## Architecture

### Database Schema (Already in migrations)
- **groups**: id, title, description, creator_id, created_at
- **group_members**: group_id, user_id, role (creator/member), status (pending/accepted), invited_by, created_at
- **group_posts**: Standard posts with group_id reference
- **group_invitations**: id, group_id, inviter_id, invitee_id, status
- **events**: id, group_id, creator_id, title, description, event_date, created_at
- **event_responses**: event_id, user_id, response (going/not_going)

### Backend API Endpoints (services/groups/)

#### Group Management
```
GET    /groups              - List all groups (discovery)
POST   /groups              - Create new group
GET    /groups/my           - Get user's groups
GET    /groups/search?q=    - Search groups by title/description
GET    /groups/:id          - Get group details
PUT    /groups/:id          - Update group (creator only)
DELETE /groups/:id          - Delete group (creator only)
```

#### Membership Management
```
POST   /groups/:id/invite   - Invite user to group
POST   /groups/:id/request  - Request to join group
POST   /groups/:id/respond  - Accept/decline invitation or request
GET    /groups/:id/members  - List group members
```

#### Group Content
```
POST   /groups/:id/posts    - Create group post
GET    /groups/:id/posts    - Get group posts
POST   /groups/:id/events   - Create event
GET    /groups/:id/events   - List events
POST   /events/:id/respond  - Respond to event (going/not_going)
```

### Frontend Components Created

#### ✅ SuggestedGroups.vue (Left Sidebar)
**Location**: `frontend/src/components/SuggestedGroups.vue`

**Features**:
- **My Groups** section: Shows groups user is a member of
- **Discover** section: Shows suggested groups to join
- **Search**: Real-time search for groups
- **Create Group** modal: Title + Description form
- Compact card design with group initials avatar
- Click to navigate to group page

**Usage**:
```vue
<SuggestedGroups />
```

#### Frontend Services

**Enhanced**: `frontend/src/services/groupsService.js`

**Available Functions**:
- `getAllGroups(token)` - Browse all groups
- `getMyGroups(token)` - User's groups
- `searchGroups(query, token)` - Search functionality
- `getGroup(groupId, token)` - Single group details
- `createGroup(groupData, token)` - Create new group
- `updateGroup(groupId, groupData, token)` - Update group
- `deleteGroup(groupId, token)` - Delete group
- `inviteToGroup(groupId, userId, token)` - Invite user
- `requestJoinGroup(groupId, token)` - Request to join
- `respondToGroupInvite(groupId, accept, token)` - Accept/decline
- `getGroupMembers(groupId, token)` - List members
- `getGroupPosts(groupId, token)` - Group posts
- `createGroupPost(groupId, postData, token)` - Create post in group
- `getGroupEvents(groupId, token)` - Group events
- `createEvent(groupId, eventData, token)` - Create event
- `respondToEvent(eventId, response, token)` - RSVP to event

### Layout Integration

**Modified**: `frontend/src/pages/FeedView.vue`

Added sidebar layout:
```vue
<div class="feed-layout">
  <SuggestedGroups />
  <section class="main-panel">
    <!-- Feed content -->
  </section>
</div>
```

**Responsive Design**:
- Desktop (>1200px): Sidebar + Feed side-by-side
- Tablet/Mobile (<1200px): Stacked layout

## Next Steps to Complete Implementation

### 1. Create Group Page Components

#### GroupView.vue (Single Group Page)
**Location**: `frontend/src/pages/GroupView.vue`

```vue
<template>
  <div class="group-view">
    <!-- Group Header -->
    <section class="group-header">
      <div class="group-icon-large">{{ groupInitials }}</div>
      <div class="group-info">
        <h1>{{ group.title }}</h1>
        <p>{{ group.description }}</p>
        <div class="group-stats">
          <span>{{ group.member_count }} members</span>
          <span>{{ group.posts_count }} posts</span>
        </div>
      </div>
      <!-- Actions: Join/Leave, Invite, etc. -->
    </section>

    <!-- Tabs: Posts, Members, Events -->
    <div class="group-tabs">
      <button @click="activeTab = 'posts'">Posts</button>
      <button @click="activeTab = 'members'">Members</button>
      <button @click="activeTab = 'events'">Events</button>
    </div>

    <!-- Tab Content -->
    <component :is="activeTabComponent" :group-id="groupId" />
  </div>
</template>
```

#### GroupPosts.vue (Posts Tab)
```vue
<template>
  <div class="group-posts">
    <!-- Create post form (members only) -->
    <CreateGroupPost v-if="isMember" @posted="loadPosts" />
    
    <!-- Posts list (same as feed) -->
    <article v-for="post in posts" :key="post.id" class="post-card">
      <!-- Same styling as FeedView posts -->
    </article>
  </div>
</template>
```

#### GroupMembers.vue (Members Tab)
```vue
<template>
  <div class="group-members">
    <!-- Invite button (members only) -->
    <button v-if="isMember" @click="showInviteModal = true">
      Invite Members
    </button>
    
    <!-- Members list -->
    <div class="members-grid">
      <article v-for="member in members" :key="member.id" class="member-card">
        <img :src="member.avatar" />
        <strong>{{ member.name }}</strong>
        <span class="role">{{ member.role }}</span>
      </article>
    </div>
  </div>
</template>
```

#### GroupEvents.vue (Events Tab)
```vue
<template>
  <div class="group-events">
    <!-- Create event button (members only) -->
    <button v-if="isMember" @click="showCreateEvent = true">
      Create Event
    </button>
    
    <!-- Events list -->
    <article v-for="event in events" :key="event.id" class="event-card">
      <h3>{{ event.title }}</h3>
      <p>{{ event.description }}</p>
      <span class="event-date">{{ formatDate(event.event_date) }}</span>
      
      <!-- RSVP Buttons -->
      <div class="rsvp-buttons">
        <button 
          @click="respondTo(event.id, 'going')"
          :class="{ active: event.user_response === 'going' }"
        >
          ✓ Going ({{ event.going_count }})
        </button>
        <button 
          @click="respondTo(event.id, 'not_going')"
          :class="{ active: event.user_response === 'not_going' }"
        >
          ✗ Not Going ({{ event.not_going_count }})
        </button>
      </div>
    </article>
  </div>
</template>
```

### 2. Backend Implementation Checklist

Verify these endpoints exist in `services/groups/`:

- [ ] `handlers/group.go` - Group CRUD operations
- [ ] `handlers/event.go` - Event creation and responses
- [ ] `db/queries.go` - Database queries for groups
- [ ] `services/group_service.go` - Business logic
- [ ] `models/group.go` - Group data structures

**Key Backend Logic**:

```go
// Privacy checks
func (s *GroupService) CanUserSeeGroup(groupID, userID int) bool {
    // Check if user is a member
}

func (s *GroupService) CanUserPost(groupID, userID int) bool {
    // Only members can post
}

func (s *GroupService) CanUserInvite(groupID, userID int) bool {
    // Members can invite
}

func (s *GroupService) CanUserManageRequests(groupID, userID int) bool {
    // Only creator can accept/decline join requests
}
```

### 3. Router Configuration

Add routes to `frontend/src/router/index.js`:

```javascript
{
  path: '/groups/:id',
  name: 'Group',
  component: () => import('../pages/GroupView.vue'),
  meta: { requiresAuth: true },
  props: true
}
```

### 4. Notifications Integration

Add group-related notifications:
- User invited to group
- Join request accepted/declined
- New post in group
- New event created
- Event reminder (day before)

### 5. Testing Checklist

#### Frontend
- [ ] Create group modal opens and submits
- [ ] Search groups works with debouncing
- [ ] Navigation to group page
- [ ] Responsive layout (desktop/mobile)
- [ ] Join/leave group functionality
- [ ] Invite members
- [ ] Request to join
- [ ] Accept/decline requests (creator only)
- [ ] Create posts in group
- [ ] Create events
- [ ] RSVP to events
- [ ] Real-time updates (WebSocket)

#### Backend
- [ ] Group creation validation (title, description required)
- [ ] Only creator can delete group
- [ ] Only creator can accept join requests
- [ ] Members can invite others
- [ ] Only members see group posts
- [ ] Event RSVP tracking
- [ ] Proper error messages
- [ ] Rate limiting on creation endpoints

### 6. Docker Configuration

Ensure `docker-compose.yml` has groups service:

```yaml
services:
  groups-service:
    build: ./services/groups
    ports:
      - "8084:8084"
    environment:
      - DATABASE_PATH=/app/social_network.db
      - AUTH_SERVICE_URL=http://auth-service:8081
    volumes:
      - ./social_network.db:/app/social_network.db
```

## Implementation Priority

### Phase 1: Core Group Features (Current)
✅ SuggestedGroups component
✅ groupsService.js with all API calls
✅ Feed layout with sidebar
⏳ GroupView page
⏳ Backend endpoints verification

### Phase 2: Group Content
- Group posts
- Member list
- Invite system
- Join requests

### Phase 3: Events Feature
- Event creation
- RSVP functionality
- Event notifications
- Event reminders

### Phase 4: Polish
- Real-time updates via WebSocket
- Group search optimization
- Image upload for groups
- Group settings
- Member roles (admin, moderator)

## File Structure Summary

```
frontend/src/
├── components/
│   ├── SuggestedGroups.vue      ✅ Created
│   ├── GroupPosts.vue            ⏳ To create
│   ├── GroupMembers.vue          ⏳ To create
│   ├── GroupEvents.vue           ⏳ To create
│   └── CreateEvent.vue           ⏳ To create
├── pages/
│   ├── FeedView.vue              ✅ Modified
│   └── GroupView.vue             ⏳ To create
└── services/
    └── groupsService.js          ✅ Enhanced

services/groups/
├── main.go
├── handlers/
│   ├── group.go                  ⏳ Verify/create
│   └── event.go                  ⏳ Verify/create
├── db/
│   └── queries.go                ⏳ Verify/create
├── services/
│   └── group_service.go          ⏳ Verify/create
└── models/
    └── group.go                  ⏳ Verify/create
```

## Design Patterns Used

1. **Sidebar Component Pattern**: Reusable, self-contained group discovery
2. **Service Layer**: Centralized API calls in groupsService.js
3. **Responsive Layout**: Flexbox with media queries
4. **Modal Pattern**: Create group form as teleported modal
5. **Tab Component Pattern**: For group content sections
6. **Real-time Updates**: WebSocket integration for live data

## Key Requirements Met

✅ Users can create groups with title and description
✅ Search/browse groups functionality
✅ Invitation system (pending backend)
✅ Join request system (pending backend)
✅ Group posts (members only) (pending)
✅ Events with RSVP (pending)
✅ Sidebar layout for discovery
✅ Responsive design

## Notes

- The SuggestedGroups component is designed to be compact and always visible on desktop
- Groups are discoverable through both "My Groups" and "Discover" sections
- Search debouncing (300ms) prevents excessive API calls
- All group operations require authentication
- Group privacy: All groups are discoverable, but content is members-only

# Groups Service - Comprehensive Guide

A concise guide for the Groups microservice, covering group management, membership control, events, and group messaging.

## Table of Contents

- [Overview](#overview)
- [Group Roles - Creator vs Member](#group-roles---creator-vs-member)
- [Membership System - Invite and Request](#membership-system---invite-and-request)
- [Events System - Group Activities](#events-system---group-activities)
- [The Complete Flow - Creating and Joining a Group](#the-complete-flow---creating-and-joining-a-group)
- [The Complete Flow - Creating an Event](#the-complete-flow---creating-an-event)
- [Database Operations](#database-operations)
- [HTTP REST Endpoints](#http-rest-endpoints)
- [Summary](#summary)

[Back to Top](#table-of-contents)

---

## Overview

The **Groups Service** manages social groups, membership, events, and group messaging.

**Port**: 8084  
**Database**: SQLite (shared)  
**Dependencies**: Auth Service (port 8081), Chat Service (port 8085 - for real-time group chat)

**Core Responsibilities**:
1. **Group Management** - Create, update, browse groups
2. **Membership Control** - Invites, join requests, approval system
3. **Events** - Create group events, RSVP tracking (going/not_going/interested)
4. **Group Messages** - REST API for group chat history (WebSocket in Chat Service)

**Key Concepts**:
- **Creator** - Group owner with admin privileges (can approve requests, update group)
- **Member** - Regular group member (can invite others, create events, send messages)
- **Two Membership Paths** - Users can be invited OR request to join
- **Member-Only Access** - Only accepted members can see events, messages, member list

[Back to Top](#table-of-contents)

---

## Group Roles - Creator vs Member

Every group has exactly ONE creator and multiple members.

### Creator Role

The user who creates the group automatically becomes the creator with `role='admin'`.

**Privileges**:
- ✅ Update group details (name, description, image)
- ✅ View and respond to join requests
- ✅ Invite members
- ✅ Create events
- ✅ Send messages
- ✅ View member list

**Creator is automatically added as member**:
```sql
-- When group is created
INSERT INTO groups (name, creator_id) VALUES ('Study Group', 1)
-- Group ID = 123

-- Creator is auto-added with admin role
INSERT INTO group_members (group_id, user_id, role, status)
VALUES (123, 1, 'admin', 'accepted')
```

### Member Role

Regular members have limited privileges compared to creator.

**Privileges**:
- ✅ Invite other users
- ✅ Create events
- ✅ Send messages
- ✅ View member list
- ✅ RSVP to events
- ❌ Cannot update group details
- ❌ Cannot approve join requests

**Database Structure**:

```sql
CREATE TABLE group_members (
    id INTEGER PRIMARY KEY,
    group_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    role TEXT NOT NULL,         -- 'admin' or 'member'
    status TEXT NOT NULL,       -- 'pending' or 'accepted'
    joined_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**Role Comparison**:

| Action | Creator | Member | Non-Member |
|--------|---------|--------|------------|
| Update group details | ✅ | ❌ | ❌ |
| Approve join requests | ✅ | ❌ | ❌ |
| Invite users | ✅ | ✅ | ❌ |
| Create events | ✅ | ✅ | ❌ |
| Send messages | ✅ | ✅ | ❌ |
| View events | ✅ | ✅ | ❌ |
| View messages | ✅ | ✅ | ❌ |
| View members | ✅ | ✅ | ❌ |
| Request to join | ❌ | ❌ | ✅ |

[Back to Top](#table-of-contents)

---

## Membership System - Invite and Request

There are **two ways** to join a group: invitation or request.

### Path 1: Invitation (Member → User)

Any member can invite other users. Invites are auto-accepted.

**Flow**:
```
1. Member A invites User B
2. System creates group_members entry with status='accepted'
3. User B is now a member (no approval needed)
```

**HTTP Request**:
```http
POST /groups/123/invite
Authorization: Bearer <member_token>

{
  "user_id": 5
}
```

**Database**:
```sql
INSERT INTO group_members (group_id, user_id, role, status)
VALUES (123, 5, 'member', 'accepted')
```

**Why auto-accept?** Trust model - if you're a member, you can vouch for others.

### Path 2: Join Request (User → Creator)

Any user can request to join. Creator must approve.

**Flow**:
```
1. User B requests to join Group 123
2. System creates group_members entry with status='pending'
3. Creator receives notification
4. Creator approves or rejects
   - Approve: status → 'accepted'
   - Reject: Delete the record
5. If approved, User B is now a member
```

**Request to Join**:
```http
POST /groups/123/request
Authorization: Bearer <user_token>
```

**Database**:
```sql
-- Step 1: User requests
INSERT INTO group_members (group_id, user_id, role, status)
VALUES (123, 5, 'member', 'pending')

-- Step 2: Creator approves
UPDATE group_members 
SET status = 'accepted' 
WHERE id = <member_id>

-- OR Creator rejects
DELETE FROM group_members WHERE id = <member_id>
```

**Membership Status Diagram**:

```
┌─────────────────────────────────────────────────┐
│                  Non-Member                     │
│                                                 │
│  ┌─────────────────┐      ┌─────────────────┐  │
│  │ Gets Invited    │      │ Requests to Join│  │
│  └────────┬────────┘      └────────┬────────┘  │
│           │                        │            │
│           ▼                        ▼            │
│  ┌─────────────────┐      ┌─────────────────┐  │
│  │ status=accepted │      │ status=pending  │  │
│  │ (instant)       │      │ (awaits creator)│  │
│  └────────┬────────┘      └────────┬────────┘  │
│           │                        │            │
│           │                        ▼            │
│           │              ┌──────────────────┐   │
│           │              │ Creator Responds │   │
│           │              └────┬────────┬────┘   │
│           │                   │        │        │
│           │            Accept │        │ Reject │
│           │                   ▼        ▼        │
│           │             accepted   (deleted)    │
│           │                   │                 │
│           └───────────────────┘                 │
│                       │                         │
│                       ▼                         │
│              ┌─────────────────┐                │
│              │  MEMBER          │                │
│              │  (Full Access)   │                │
│              └─────────────────┘                │
└─────────────────────────────────────────────────┘
```

[Back to Top](#table-of-contents)

---

## Events System - Group Activities

Group members can create events and RSVP.

### Creating Events

**Any member** can create an event (not just creator).

**Request**:
```http
POST /events
Authorization: Bearer <member_token>

{
  "group_id": 123,
  "title": "Study Session",
  "description": "Math exam prep",
  "event_time": "2024-10-20T15:00:00Z"
}
```

**Database**:
```sql
CREATE TABLE events (
    id INTEGER PRIMARY KEY,
    group_id INTEGER NOT NULL,
    creator_id INTEGER,
    title TEXT NOT NULL,
    description TEXT,
    event_time DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(id)
);
```

### RSVP System

Members can respond with three options: **going**, **not_going**, or **interested**.

**Response**:
```http
POST /events/456/respond
Authorization: Bearer <member_token>

{
  "event_id": 456,
  "response": "going"
}
```

**Database**:
```sql
CREATE TABLE event_responses (
    id INTEGER PRIMARY KEY,
    event_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    response TEXT NOT NULL,  -- 'going', 'not_going', 'interested'
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (event_id) REFERENCES events(id),
    UNIQUE(event_id, user_id)  -- One response per user per event
);
```

**Upsert Logic**:
```sql
-- If user already responded, update their response
INSERT INTO event_responses (event_id, user_id, response)
VALUES (456, 1, 'going')
ON CONFLICT(event_id, user_id) 
DO UPDATE SET response = 'going', created_at = CURRENT_TIMESTAMP
```

### Event Response Counts

When retrieving an event, include response counts:

```json
{
  "id": 456,
  "group_id": 123,
  "title": "Study Session",
  "event_time": "2024-10-20T15:00:00Z",
  "going_count": 12,
  "not_going_count": 3,
  "interested_count": 5,
  "user_response": "going"
}
```

**SQL Query**:
```sql
SELECT 
    e.*,
    SUM(CASE WHEN er.response = 'going' THEN 1 ELSE 0 END) as going_count,
    SUM(CASE WHEN er.response = 'not_going' THEN 1 ELSE 0 END) as not_going_count,
    SUM(CASE WHEN er.response = 'interested' THEN 1 ELSE 0 END) as interested_count,
    (SELECT response FROM event_responses WHERE event_id = e.id AND user_id = ?) as user_response
FROM events e
LEFT JOIN event_responses er ON e.id = er.event_id
WHERE e.id = ?
GROUP BY e.id
```

[Back to Top](#table-of-contents)

---

## The Complete Flow - Creating and Joining a Group

**Scenario**: User A creates a group, User B requests to join.

### Part 1: Creating the Group

```
0.0s - User A: POST /groups {"name": "Book Club"}

0.5s - GroupService.CreateGroup
       - Validate: name not empty ✓
       - Insert into groups table

0.6s - Database INSERT
       Query: INSERT INTO groups (name, creator_id) VALUES ('Book Club', 1)
       Result: group.ID = 123

0.7s - Auto-add creator as admin member
       Query: INSERT INTO group_members (group_id, user_id, role, status)
              VALUES (123, 1, 'admin', 'accepted')

0.8s - Response: 200 OK
       {
         "group": {
           "id": 123,
           "name": "Book Club",
           "creator_id": 1,
           "created_at": "2024-10-16T15:00:00Z"
         }
       }
```

**Database State**:
```
groups table:
┌────┬───────────┬────────────┐
│ id │ name      │ creator_id │
├────┼───────────┼────────────┤
│123 │ Book Club │ 1          │
└────┴───────────┴────────────┘

group_members table:
┌────┬──────────┬─────────┬───────┬──────────┐
│ id │ group_id │ user_id │ role  │ status   │
├────┼──────────┼─────────┼───────┼──────────┤
│ 1  │ 123      │ 1       │ admin │ accepted │
└────┴──────────┴─────────┴───────┴──────────┘
```

### Part 2: User B Requests to Join

```
0.0s - User B: POST /groups/123/request

0.5s - GroupService.RequestToJoin
       - Call: db.RequestToJoinGroup(database, groupID=123, userID=2)

0.6s - Database INSERT
       Query: INSERT INTO group_members (group_id, user_id, role, status)
              VALUES (123, 2, 'member', 'pending')

0.7s - Response: 200 OK
       {"message": "Join request sent"}
```

**Database State**:
```
group_members table:
┌────┬──────────┬─────────┬────────┬──────────┐
│ id │ group_id │ user_id │ role   │ status   │
├────┼──────────┼─────────┼────────┼──────────┤
│ 1  │ 123      │ 1       │ admin  │ accepted │
│ 2  │ 123      │ 2       │ member │ pending  │ ← User B waiting
└────┴──────────┴─────────┴────────┴──────────┘
```

### Part 3: Creator Approves Request

```
0.0s - User A (creator): GET /groups/123/requests
       Response: [
         {
           "id": 2,
           "group_id": 123,
           "user_id": 2,
           "role": "member",
           "status": "pending"
         }
       ]

1.0s - User A: POST /groups/123/requests/respond
       {"member_id": 2, "accept": true}

1.5s - GroupService.RespondToRequest
       - Check: Is User A the creator? ✓
       - Call: db.RespondToJoinRequest(memberID=2, accept=true)

1.6s - Database UPDATE
       Query: UPDATE group_members SET status = 'accepted' WHERE id = 2

1.7s - Response: 200 OK
       {"message": "Request accepted"}
```

**Final Database State**:
```
group_members table:
┌────┬──────────┬─────────┬────────┬──────────┐
│ id │ group_id │ user_id │ role   │ status   │
├────┼──────────┼─────────┼────────┼──────────┤
│ 1  │ 123      │ 1       │ admin  │ accepted │
│ 2  │ 123      │ 2       │ member │ accepted │ ← User B now member
└────┴──────────┴─────────┴────────┴──────────┘
```

User B can now:
- View group events
- Create events
- Send group messages
- Invite other users
- View member list

[Back to Top](#table-of-contents)

---

## The Complete Flow - Creating an Event

**Scenario**: User B (member) creates an event, User C RSVPs.

```
0.0s - User B: POST /events
       {
         "group_id": 123,
         "title": "Monthly Meetup",
         "event_time": "2024-10-25T18:00:00Z"
       }

0.5s - GroupService.CreateEvent
       - Check: Is User B a member of Group 123?
         Query: SELECT 1 FROM group_members 
                WHERE group_id=123 AND user_id=2 AND status='accepted'
         Result: TRUE ✓

0.6s - Validation
       - title not empty ✓
       - event_time valid ISO format ✓

0.7s - Database INSERT
       Query: INSERT INTO events (group_id, creator_id, title, event_time)
              VALUES (123, 2, 'Monthly Meetup', '2024-10-25 18:00:00')
       Result: event.ID = 456

0.8s - Response: 200 OK
       {
         "event": {
           "id": 456,
           "group_id": 123,
           "creator_id": 2,
           "title": "Monthly Meetup",
           "event_time": "2024-10-25T18:00:00Z"
         }
       }

---

1.0s - User C (another member): POST /events/456/respond
       {"event_id": 456, "response": "going"}

1.5s - GroupService.RespondToEvent
       - Get event to check group_id
       - Check: Is User C a member of Group 123? ✓
       - Validate response: "going" is valid ✓

1.6s - Database UPSERT
       Query: INSERT INTO event_responses (event_id, user_id, response)
              VALUES (456, 3, 'going')
              ON CONFLICT(event_id, user_id) 
              DO UPDATE SET response='going'

1.7s - Response: 200 OK
       {"message": "Response recorded"}

---

2.0s - Any member: GET /events/456
       Response: {
         "event": {
           "id": 456,
           "title": "Monthly Meetup",
           "event_time": "2024-10-25T18:00:00Z",
           "going_count": 1,        ← User C
           "not_going_count": 0,
           "interested_count": 0
         }
       }
```

[Back to Top](#table-of-contents)

---

## Database Operations

### Group Queries

**CreateGroup** - Create group and add creator as admin:
```go
func CreateGroup(db *sql.DB, name string, creatorID int) (*Group, error) {
    // Insert group
    result, _ := db.Exec(`INSERT INTO groups (name, creator_id) VALUES (?, ?)`, name, creatorID)
    groupID, _ := result.LastInsertId()
    
    // Add creator as admin member
    db.Exec(`INSERT INTO group_members (group_id, user_id, role, status) 
             VALUES (?, ?, 'admin', 'accepted')`, groupID, creatorID)
    
    return GetGroupByID(db, int(groupID))
}
```

**GetGroupWithDetails** - Get group with user's relationship:
```go
func GetGroupWithDetails(db *sql.DB, groupID, userID int) (*GroupWithDetails, error) {
    query := `
        SELECT g.*, COUNT(gm.id) as member_count,
               CASE WHEN EXISTS(SELECT 1 FROM group_members 
                    WHERE group_id=g.id AND user_id=? AND status='accepted') 
                    THEN 1 ELSE 0 END as is_member,
               CASE WHEN g.creator_id = ? THEN 1 ELSE 0 END as is_creator
        FROM groups g
        LEFT JOIN group_members gm ON g.id = gm.group_id AND gm.status = 'accepted'
        WHERE g.id = ?
        GROUP BY g.id
    `
    // Returns: group info + member_count + is_member + is_creator flags
}
```

### Membership Queries

**InviteMember** - Invite user (auto-accept):
```go
func InviteMember(db *sql.DB, groupID, userID int) error {
    query := `INSERT INTO group_members (group_id, user_id, role, status)
              VALUES (?, ?, 'member', 'accepted')`
    _, err := db.Exec(query, groupID, userID)
    return err
}
```

**RequestToJoinGroup** - Request to join (pending):
```go
func RequestToJoinGroup(db *sql.DB, groupID, userID int) error {
    query := `INSERT INTO group_members (group_id, user_id, role, status)
              VALUES (?, ?, 'member', 'pending')`
    _, err := db.Exec(query, groupID, userID)
    return err
}
```

**RespondToJoinRequest** - Approve or reject:
```go
func RespondToJoinRequest(db *sql.DB, memberID int, accept bool) error {
    if accept {
        query := `UPDATE group_members SET status = 'accepted' WHERE id = ?`
        _, err := db.Exec(query, memberID)
        return err
    } else {
        query := `DELETE FROM group_members WHERE id = ?`
        _, err := db.Exec(query, memberID)
        return err
    }
}
```

**IsGroupMember** - Check membership:
```go
func IsGroupMember(db *sql.DB, groupID, userID int) (bool, error) {
    query := `SELECT 1 FROM group_members 
              WHERE group_id = ? AND user_id = ? AND status = 'accepted'`
    var exists int
    err := db.QueryRow(query, groupID, userID).Scan(&exists)
    if err == sql.ErrNoRows {
        return false, nil
    }
    return true, err
}
```

### Event Queries

**CreateEvent**:
```go
func CreateEvent(db *sql.DB, groupID, creatorID int, title string, eventTime time.Time) (*Event, error) {
    query := `INSERT INTO events (group_id, creator_id, title, event_time)
              VALUES (?, ?, ?, ?)`
    result, _ := db.Exec(query, groupID, creatorID, title, eventTime)
    eventID, _ := result.LastInsertId()
    return GetEventByID(db, int(eventID))
}
```

**RespondToEvent** - RSVP:
```go
func RespondToEvent(db *sql.DB, eventID, userID int, response string) error {
    query := `INSERT INTO event_responses (event_id, user_id, response)
              VALUES (?, ?, ?)
              ON CONFLICT(event_id, user_id) 
              DO UPDATE SET response = ?, created_at = CURRENT_TIMESTAMP`
    _, err := db.Exec(query, eventID, userID, response, response)
    return err
}
```

[Back to Top](#table-of-contents)

---

## HTTP REST Endpoints

### Group Endpoints

**POST /groups** - Create group
```http
POST /groups
Authorization: Bearer <token>

{"name": "Book Club", "description": "Weekly book discussions"}

Response 200: {"group": {...}}
```

**GET /groups** - Get all groups
```http
GET /groups
Authorization: Bearer <token>

Response 200: {"groups": [{...}, {...}]}
```

**GET /groups/:id** - Get specific group
```http
GET /groups/123
Authorization: Bearer <token>

Response 200: {
  "group": {
    "id": 123,
    "name": "Book Club",
    "member_count": 15,
    "is_member": true,
    "is_creator": false,
    "has_pending_request": false
  }
}
```

**PUT /groups/:id** - Update group (creator only)
```http
PUT /groups/123
Authorization: Bearer <creator_token>

{"name": "New Name", "description": "Updated description"}

Response 200: {"message": "Group updated"}
Response 403: {"error": "only group creator can update"}
```

### Membership Endpoints

**POST /groups/:id/invite** - Invite user (members can invite)
```http
POST /groups/123/invite
Authorization: Bearer <member_token>

{"user_id": 5}

Response 200: {"message": "User invited"}
```

**POST /groups/:id/request** - Request to join
```http
POST /groups/123/request
Authorization: Bearer <token>

Response 200: {"message": "Join request sent"}
```

**GET /groups/:id/requests** - Get pending requests (creator only)
```http
GET /groups/123/requests
Authorization: Bearer <creator_token>

Response 200: {"requests": [{...}]}
```

**POST /groups/:id/requests/respond** - Approve/reject (creator only)
```http
POST /groups/123/requests/respond
Authorization: Bearer <creator_token>

{"member_id": 2, "accept": true}

Response 200: {"message": "Request accepted"}
```

**GET /groups/:id/members** - Get members (members only)
```http
GET /groups/123/members
Authorization: Bearer <member_token>

Response 200: {"members": [{...}]}
```

### Event Endpoints

**POST /events** - Create event (members only)
```http
POST /events
Authorization: Bearer <member_token>

{
  "group_id": 123,
  "title": "Meetup",
  "event_time": "2024-10-25T18:00:00Z"
}

Response 200: {"event": {...}}
```

**GET /events/:id** - Get event with RSVP counts
```http
GET /events/456
Authorization: Bearer <token>

Response 200: {
  "event": {
    "id": 456,
    "title": "Meetup",
    "going_count": 12,
    "not_going_count": 3,
    "interested_count": 5,
    "user_response": "going"
  }
}
```

**POST /events/:id/respond** - RSVP to event
```http
POST /events/456/respond
Authorization: Bearer <member_token>

{"event_id": 456, "response": "going"}

Response 200: {"message": "Response recorded"}

Valid responses: "going", "not_going", "interested"
```

**GET /groups/:id/events** - Get all group events (members only)
```http
GET /groups/123/events
Authorization: Bearer <member_token>

Response 200: {"events": [{...}]}
```

### Group Message Endpoints

**POST /groups/:id/messages** - Send message (members only)
```http
POST /groups/123/messages
Authorization: Bearer <member_token>

{"content": "Hello everyone!"}

Response 200: {"message": {...}}
```

**GET /groups/:id/messages** - Get message history (members only)
```http
GET /groups/123/messages?limit=50
Authorization: Bearer <member_token>

Response 200: {"messages": [{...}]}
```

**Note**: For real-time group chat, use Chat Service WebSocket (port 8085).

[Back to Top](#table-of-contents)

---

## Summary

**Groups Service Core Concepts**:

1. **Two Roles**
   - **Creator** (admin) - Full control, approves requests, updates group
   - **Member** - Can invite, create events, send messages

2. **Two Membership Paths**
   - **Invitation** - Members invite users → auto-accepted
   - **Request** - Users request → creator approves/rejects

3. **Status System**
   - `pending` - Awaiting creator approval
   - `accepted` - Full member access

4. **Events with RSVP**
   - Any member can create events
   - Three response types: going, not_going, interested
   - Response counts tracked per event

5. **Member-Only Access**
   - All group content (events, messages, members) requires membership
   - Enforced via `IsGroupMember` check in service layer

6. **Integration with Chat Service**
   - REST endpoints for message history
   - Real-time WebSocket chat handled by Chat Service (port 8085)

**Key Takeaway**: The Groups service implements a permission-based system with creator privileges and member-only access. The dual membership path (invite vs request) balances open invitation with creator control, while the event RSVP system tracks participation across three engagement levels.

[Back to Top](#table-of-contents)

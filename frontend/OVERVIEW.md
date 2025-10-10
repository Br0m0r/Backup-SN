# 🎨 Test Frontend Overview

## Visual Layout

```
┌─────────────────────────────────────────────────────────┐
│              🚀 Social Network Test Client              │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│  🔐 Authentication                                       │
│  ┌────────────┬────────────┐                           │
│  │  Register  │   Login    │  ← Tabs                   │
│  └────────────┴────────────┘                           │
│                                                          │
│  Email:        [________________]                       │
│  Password:     [________________]                       │
│  First Name:   [________________]                       │
│  Last Name:    [________________]                       │
│  DOB:          [________________]                       │
│                                                          │
│  [Register]                                             │
└─────────────────────────────────────────────────────────┘

         ↓ After Login ↓

┌─────────────────────────────────────────────────────────┐
│  Logged in as: alice@example.com                        │
│  ID: 1  Name: Alice Smith                               │
│  [Logout] [Get Session]                                 │
│                                                          │
│  ┌─────────┬─────────┬─────────┐                       │
│  │ Profile │  Posts  │  Users  │  ← Main Tabs          │
│  └─────────┴─────────┴─────────┘                       │
│                                                          │
│  Profile Tab:                                           │
│  [Get My Profile] [Update Profile]                      │
│  ┌────────────────────────────────────────┐            │
│  │ {JSON Response}                        │            │
│  │ "user": {                              │            │
│  │   "id": 1,                             │            │
│  │   "email": "alice@example.com"         │            │
│  │ }                                      │            │
│  └────────────────────────────────────────┘            │
└─────────────────────────────────────────────────────────┘
```

## Features by Tab

### 👤 Profile Tab
```
┌─────────────────────────────────────┐
│ 👤 Profile Management               │
├─────────────────────────────────────┤
│ [Get My Profile]                    │
│ [Update Profile]                    │
│                                     │
│ ┌─────────────────────────────────┐ │
│ │ Update Form (when shown):       │ │
│ │ Nickname:    [_______________]  │ │
│ │ About Me:    [_______________]  │ │
│ │ Visibility:  [Public ▼]         │ │
│ │ [Save] [Cancel]                 │ │
│ └─────────────────────────────────┘ │
│                                     │
│ Response:                           │
│ ┌─────────────────────────────────┐ │
│ │ {                               │ │
│ │   "success": true,              │ │
│ │   "data": {...}                 │ │
│ │ }                               │ │
│ └─────────────────────────────────┘ │
└─────────────────────────────────────┘
```

### 📝 Posts Tab
```
┌─────────────────────────────────────┐
│ 📝 Posts & Feed                     │
├─────────────────────────────────────┤
│ [Create Post] [Get Feed]            │
│                                     │
│ ┌─────────────────────────────────┐ │
│ │ Post #1 (public)                │ │
│ │ Hello world!                    │ │
│ │ By User ID: 1 • 2025-10-10     │ │
│ │ [View Comments] [Delete]        │ │
│ └─────────────────────────────────┘ │
│                                     │
│ ┌─────────────────────────────────┐ │
│ │ Post #2 (private)               │ │
│ │ Private thoughts...             │ │
│ │ By User ID: 2 • 2025-10-10     │ │
│ │ [View Comments]                 │ │
│ └─────────────────────────────────┘ │
│                                     │
│ Response:                           │
│ ┌─────────────────────────────────┐ │
│ │ {"posts": [...]}                │ │
│ └─────────────────────────────────┘ │
└─────────────────────────────────────┘
```

### 👥 Users Tab
```
┌─────────────────────────────────────┐
│ 👥 Users & Social                   │
├─────────────────────────────────────┤
│ Search: [_______________] [Search]  │
│ [My Followers] [My Following]       │
│                                     │
│ ┌─────────────────────────────────┐ │
│ │ Bob Jones                       │ │
│ │ Username: bob123                │ │
│ │ About: Love coding!             │ │
│ │ [Follow]                        │ │
│ └─────────────────────────────────┘ │
│                                     │
│ ┌─────────────────────────────────┐ │
│ │ Charlie Brown                   │ │
│ │ Username: charlie               │ │
│ │ About: Coffee enthusiast        │ │
│ │ [Follow]                        │ │
│ └─────────────────────────────────┘ │
│                                     │
│ Response:                           │
│ ┌─────────────────────────────────┐ │
│ │ {"users": [...]}                │ │
│ └─────────────────────────────────┘ │
└─────────────────────────────────────┘
```

## Color Scheme

- **Primary Gradient**: Purple (#667eea) to Violet (#764ba2)
- **Background**: White cards with rounded corners
- **Success**: Green (#d4edda)
- **Error**: Red (#f8d7da)
- **Info**: Blue (#d1ecf1)

## Interactive Elements

### Buttons
- Hover: Lifts up with shadow
- Click: Pressed down effect
- Gradient background

### Forms
- Focus: Border changes to purple
- Validation: Red border on error

### Status Messages
- Auto-dismiss after 5 seconds
- Color-coded by type

### Response Boxes
- Scrollable JSON viewer
- Monospace font
- Gray background

## API Call Flow

```
┌──────────┐      ┌──────────┐      ┌──────────┐
│ Browser  │─────▶│  Auth    │      │  User    │
│ Frontend │      │ Service  │      │ Service  │
│          │◀─────│  :8081   │      │  :8082   │
└──────────┘      └──────────┘      └──────────┘
     │                                     │
     │  1. Register/Login                 │
     │  ──────────────▶                   │
     │  ◀── Token ────                    │
     │                                     │
     │  2. Get Profile (+ Token)          │
     │  ───────────────────────────────▶  │
     │  ◀─────── User Data ───────────────│
     │                                     │
     
┌──────────┐      ┌──────────┐
│ Browser  │      │  Post    │
│ Frontend │      │ Service  │
│          │      │  :8083   │
└──────────┘      └──────────┘
     │                  │
     │  3. Get Feed     │
     │  ──────────────▶ │
     │  ◀── Posts ───── │
```

## Storage

### LocalStorage
```javascript
{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

Token is automatically:
- Saved on login/register
- Included in all API calls
- Removed on logout
- Persisted across page refreshes

## Keyboard Shortcuts

- **Enter** in login/register forms → Submit
- **Escape** in forms → Cancel (if cancel button exists)
- **Tab** → Navigate between fields

## Mobile Responsiveness

The UI is responsive but optimized for desktop testing. On mobile:
- Cards stack vertically
- Buttons take full width
- Tabs remain horizontal scrollable

## Browser Compatibility

Tested on:
- ✅ Chrome/Chromium
- ✅ Firefox
- ✅ Safari
- ✅ Edge

Requires modern browser with:
- ES6+ JavaScript
- Fetch API
- LocalStorage
- CSS Grid/Flexbox

## Testing Tips

1. **Open DevTools (F12)** to see:
   - Network requests
   - Console logs
   - LocalStorage state

2. **Use Multiple Windows** to test:
   - Open normal window for User 1
   - Open incognito for User 2
   - Test interactions between users

3. **Watch Response Boxes** to see:
   - Actual API responses
   - Data structures
   - Error messages

4. **Check Status Messages** for:
   - Operation success/failure
   - Quick feedback
   - Auto-dismiss after 5s

## Quick Actions Reference

| Action | Location | Requires Auth |
|--------|----------|---------------|
| Register | Auth Tab | No |
| Login | Auth Tab | No |
| Logout | Top Bar | Yes |
| Get Session | Top Bar | Yes |
| View Profile | Profile Tab | Yes |
| Update Profile | Profile Tab | Yes |
| Search Users | Users Tab | Yes |
| Follow User | Users Tab | Yes |
| View Followers | Users Tab | Yes |
| View Following | Users Tab | Yes |
| Create Post | Posts Tab | Yes |
| Get Feed | Posts Tab | Yes |
| View Comments | Posts Tab | Yes |
| Delete Post | Posts Tab | Yes (owner) |

Happy Testing! 🎉

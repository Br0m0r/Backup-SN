import axios from 'axios'

const GROUPS_API_URL = import.meta.env.VITE_GROUPS_API_URL || 'http://localhost:8084'

const client = axios.create({
  baseURL: GROUPS_API_URL,
  headers: {
    'Content-Type': 'application/json'
  },
  withCredentials: false,
  timeout: 10000
})

// Helper to unwrap response data
function unwrapResponse(response) {
  if (response.data.success) {
    return response.data.data
  }
  throw new Error(response.data.error || 'Request failed')
}

// Get all groups (for discovery)
export async function getAllGroups(token) {
  const response = await client.get('/groups', {
    headers: { Authorization: `Bearer ${token}` }
  })
  const groups = unwrapResponse(response)
  return { groups: Array.isArray(groups) ? groups : [] }
}

// Get groups user is a member of (filtered client-side from getAllGroups)
export async function getMyGroups(token) {
  const response = await client.get('/groups', {
    headers: { Authorization: `Bearer ${token}` }
  })
  const groups = unwrapResponse(response)
  const groupsArray = Array.isArray(groups) ? groups : []
  // Filter to only groups where user is a member
  return { groups: groupsArray.filter(g => g.is_member) }
}

// Search groups
export async function searchGroups(query, token) {
  const response = await client.get('/groups/search', {
    params: { q: query },
    headers: { Authorization: `Bearer ${token}` }
  })
  return unwrapResponse(response)
}

// Get single group details
export async function getGroup(groupId, token) {
  const response = await client.get(`/groups/${groupId}`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return unwrapResponse(response)
}

// Create new group
export async function createGroup(formData, token) {
  const response = await client.post('/groups', formData, {
    headers: { 
      Authorization: `Bearer ${token}`,
      'Content-Type': 'multipart/form-data'
    }
  })
  return unwrapResponse(response)
}

// Update group
export async function updateGroup(groupId, groupData, token) {
  const response = await client.put(`/groups/${groupId}`, groupData, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return unwrapResponse(response)
}

// Update group image (owner only)
export async function updateGroupImage(groupId, imageFile, token) {
  const formData = new FormData()
  formData.append('image', imageFile)
  
  const response = await client.put(`/groups/${groupId}/image`, formData, {
    headers: { 
      Authorization: `Bearer ${token}`,
      'Content-Type': 'multipart/form-data'
    }
  })
  return unwrapResponse(response)
}

// Delete group
export async function deleteGroup(groupId, token) {
  const response = await client.delete(`/groups/${groupId}`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return unwrapResponse(response)
}

// Invite user to group
export async function inviteToGroup(groupId, userId, token) {
  const response = await client.post(`/groups/${groupId}/invite`, 
    { user_id: userId },
    { headers: { Authorization: `Bearer ${token}` } }
  )
  return unwrapResponse(response)
}

// Request to join group
export async function requestJoinGroup(groupId, token) {
  const response = await client.post(`/groups/${groupId}/request`, {},
    { headers: { Authorization: `Bearer ${token}` } }
  )
  return unwrapResponse(response)
}

// Get user's pending invitations
export async function getMyInvitations(token) {
  const response = await client.get('/invitations', {
    headers: { Authorization: `Bearer ${token}` }
  })
  return unwrapResponse(response)
}

// Respond to group invitation
export async function respondToInvitation(invitationId, accept, token) {
  const response = await client.post(`/invitations/${invitationId}/respond`,
    { accept },
    { headers: { Authorization: `Bearer ${token}` } }
  )
  return unwrapResponse(response)
}

// Respond to group invitation or join request (DEPRECATED - kept for backward compatibility)
export async function respondToGroupInvite(groupId, accept, token) {
  const response = await client.post(
    `${GROUPS_API_URL}/groups/${groupId}/respond`,
    { accept },
    {
      headers: {
        Authorization: `Bearer ${token}`
      }
    }
  )
  return unwrapResponse(response)
}

// Get group members
export async function getGroupMembers(groupId, token) {
  const response = await client.get(`/groups/${groupId}/members`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return unwrapResponse(response)
}

// Get group posts
export async function getGroupPosts(groupId, token) {
  const response = await client.get(`/groups/${groupId}/posts`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return unwrapResponse(response)
}

// Create group post
export async function createGroupPost(groupId, postData, token) {
  const response = await client.post(`/groups/${groupId}/posts`, postData, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return unwrapResponse(response)
}

// Get group events
export async function getGroupEvents(groupId, token) {
  const response = await client.get(`/groups/${groupId}/events`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return unwrapResponse(response)
}


export async function createEvent(groupId, eventData, token) {
  const response = await client.post('/events', 
    { ...eventData, group_id: groupId },
    { headers: { Authorization: `Bearer ${token}` } }
  )
  return unwrapResponse(response)
}

export async function respondToEvent(eventId, responseStatus, token) {
  const res = await client.post('/events/respond', 
    { event_id: eventId, response: responseStatus },
    { headers: { Authorization: `Bearer ${token}` } }
  )
  return unwrapResponse(res)
}


export async function getPendingRequests(groupId, token) {
  const response = await client.get(`/groups/${groupId}/requests`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return unwrapResponse(response)
}

// Respond to join request (creator only)
export async function respondToRequest(groupId, memberId, accept, token) {
  const response = await client.post(`/groups/${groupId}/requests/respond`,
    { member_id: memberId, accept },
    { headers: { Authorization: `Bearer ${token}` } }
  )
  return unwrapResponse(response)
}

// Get group chat messages
export async function getGroupMessages(groupId, token) {
  const response = await client.get(`/groups/${groupId}/messages`, {
    params: { limit: 50 },
    headers: { Authorization: `Bearer ${token}` }
  })
  return unwrapResponse(response)
}

// Send a chat message to the group
export async function createGroupMessage(groupId, content, token) {
  const response = await client.post(`/groups/${groupId}/messages`,
    { content },
    { headers: { Authorization: `Bearer ${token}` } }
  )
  return unwrapResponse(response)
}

export async function leaveGroup(groupId, token) {
  const response = await client.delete(`/groups/${groupId}/leave`, {
    headers: { Authorization: `Bearer ${token}` }
  })
  return unwrapResponse(response)
}

// Helper to construct group image URLs
export function getGroupImageUrl(imageUrl) {
  if (!imageUrl) return ''
  if (imageUrl.startsWith('http://') || imageUrl.startsWith('https://')) {
    return imageUrl
  }
  // Relative path, prepend groups service base URL
  return `${GROUPS_API_URL}/${imageUrl}`
}
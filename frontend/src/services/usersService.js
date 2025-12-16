import axios from 'axios'

const USERS_BASE_URL = import.meta.env.VITE_USERS_API_URL 

const client = axios.create({
  baseURL: USERS_BASE_URL,
  headers: {
    'Content-Type': 'application/json'
  },
  withCredentials: false,
  timeout: 10000
})

function unwrapResponse(response) {
  const payload = response?.data

  if (!payload) {
    throw new Error('No response from users service')
  }

  if (payload.success === false) {
    throw new Error(payload.error || 'Users service error')
  }

  if (payload.data !== undefined) {
    return payload.data
  }

  return payload
}

export async function searchUsers(searchTerm, token) {
  const response = await client.get('/search', {
    params: { q: searchTerm },
    headers: {
      Authorization: `Bearer ${token}`
    }
  })
  return unwrapResponse(response)
}

export async function searchUsersForGroup(searchTerm, groupId, token) {
  const response = await client.get('/search/group', {
    params: { q: searchTerm, group_id: groupId },
    headers: {
      Authorization: `Bearer ${token}`
    }
  })
  return unwrapResponse(response)
}

export async function followUser(userID, token) {
  const response = await client.post(
    '/follow',
    { user_id: userID },
    {
      headers: {
        Authorization: `Bearer ${token}`
      }
    }
  )
  return unwrapResponse(response)
}

export async function unfollowUser(userID, token) {
  const response = await client.delete('/follow', {
    data: { user_id: userID },
    headers: {
      Authorization: `Bearer ${token}`
    }
  })
  return unwrapResponse(response)
}

export async function getFollowers(token, userId) {
  // userId is required - always use the user-specific endpoint
  const response = await client.get(`/users/${userId}/followers`, {
    headers: {
      Authorization: `Bearer ${token}`
    }
  })
  return unwrapResponse(response)
}

export async function getFollowing(token, userId) {
  // userId is required - always use the user-specific endpoint
  const response = await client.get(`/users/${userId}/following`, {
    headers: {
      Authorization: `Bearer ${token}`
    }
  })
  return unwrapResponse(response)
}

export async function getFollowStatus(userID, token) {
  const response = await client.get(`/follow/status/${userID}`, {
    headers: {
      Authorization: `Bearer ${token}`
    }
  })
  return unwrapResponse(response)
}

export async function getUserProfile(userID, token) {
  const response = await client.get(`/users/${userID}/profile`, {
    headers: {
      Authorization: `Bearer ${token}`
    }
  })
  return unwrapResponse(response)
}

export async function updatePrivacy(isPublic, token) {
  const response = await client.put(
    '/users/me/privacy',
    { is_public: isPublic },
    {
      headers: {
        Authorization: `Bearer ${token}`
      }
    }
  )
  return unwrapResponse(response)
}

export async function respondToFollowRequest(followerId, accept, token) {
  const response = await client.post(
    '/follow/respond',
    { follower_id: followerId, accept },
    {
      headers: {
        Authorization: `Bearer ${token}`
      }
    }
  )
  return unwrapResponse(response)
}

export async function updateProfile(profileData, token) {
  const response = await client.put(
    '/users/me',
    profileData,
    {
      headers: {
        Authorization: `Bearer ${token}`
      }
    }
  )
  return unwrapResponse(response)
}

export async function uploadAvatar(file, token) {
  const formData = new FormData()
  formData.append('avatar', file)

  const response = await axios.post(`${USERS_BASE_URL}/upload/avatar`, formData, {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'multipart/form-data'
    }
  })
  return unwrapResponse(response)
}

export async function deleteAvatar(avatarPath, token) {
  const response = await client.delete('/upload/avatar', {
    params: { path: avatarPath },
    headers: {
      Authorization: `Bearer ${token}`
    }
  })
  return unwrapResponse(response)
}

/**
 * Get the full URL for an avatar
 * @param {string} avatarPath - The avatar path from the API
 * @returns {string} Full avatar URL or empty string
 */
export function getAvatarUrl(avatarPath) {
  if (!avatarPath) return ''
  
  // If it's already a full URL, return it
  if (avatarPath.startsWith('http')) return avatarPath
  
  // Otherwise, construct the full URL to the users service
  return `${USERS_BASE_URL}${avatarPath}`
}

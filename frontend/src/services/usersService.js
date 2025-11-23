import axios from 'axios'

const USERS_BASE_URL = import.meta.env.VITE_USERS_API_URL || 'http://localhost:8082'

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

export async function getFollowers(token) {
  const response = await client.get('/followers', {
    headers: {
      Authorization: `Bearer ${token}`
    }
  })
  return unwrapResponse(response)
}

export async function getFollowing(token) {
  const response = await client.get('/following', {
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

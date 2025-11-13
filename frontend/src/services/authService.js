import axios from 'axios'

const AUTH_BASE_URL = import.meta.env.VITE_AUTH_API_URL || 'http://localhost:8081'

const client = axios.create({
  baseURL: AUTH_BASE_URL,
  headers: {
    'Content-Type': 'application/json'
  },
  withCredentials: false,
  timeout: 10000
})

function unwrapResponse(response) {
  const payload = response?.data

  if (!payload) {
    throw new Error('No response from auth service')
  }

  if (payload.success === false) {
    throw new Error(payload.error || 'Auth service error')
  }

  if (payload.data !== undefined) {
    return payload.data
  }

  return payload
}

export async function registerUser(requestBody) {
  const response = await client.post('/register', requestBody)
  return unwrapResponse(response)
}

export async function loginUser(requestBody) {
  const response = await client.post('/login', requestBody)
  return unwrapResponse(response)
}

export async function logoutUser(token) {
  if (!token) return

  const response = await client.post(
    '/logout',
    null,
    {
      headers: {
        Authorization: `Bearer ${token}`
      }
    }
  )

  return unwrapResponse(response)
}

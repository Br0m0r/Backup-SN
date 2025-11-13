// Authentication Store
// This manages the current user's session and provides auth info to WebSocket connections

import { ref } from 'vue'

// Global reactive state for the authenticated user
const user = ref(null)
const token = ref(null)

const STORAGE_KEYS = {
  user: 'user',
  token: 'token'
}

/**
 * WHY: Store user session data globally
 * - WebSocket connections need user ID and token for authentication
 * - Multiple components need to access user info without prop drilling
 * - Centralized auth state makes logout/login easier
 */

/**
 * Get the current authenticated user
 * @returns {Object|null} User object with id, username, etc.
 */
export function getUser() {
  return user.value
}

/**
 * Get the current auth token
 * @returns {string|null} JWT token
 */
export function getToken() {
  return token.value
}

/**
 * Set user session after login
 * @param {Object} userData - User object from login response
 * @param {string} authToken - JWT token
 */
export function setUser(userData, authToken, options = {}) {
  const { persist = true } = options

  user.value = userData
  token.value = authToken
  
  if (!userData || !authToken) {
    clearUser()
    return
  }

  if (persist) {
    localStorage.setItem(STORAGE_KEYS.user, JSON.stringify(userData))
    localStorage.setItem(STORAGE_KEYS.token, authToken)
  } else {
    localStorage.removeItem(STORAGE_KEYS.user)
    localStorage.removeItem(STORAGE_KEYS.token)
  }
}

/**
 * Clear user session on logout
 */
export function clearUser() {
  user.value = null
  token.value = null
  localStorage.removeItem(STORAGE_KEYS.user)
  localStorage.removeItem(STORAGE_KEYS.token)
}

/**
 * Restore user session from localStorage (on app load)
 */
export function restoreSession() {
  const storedUser = localStorage.getItem(STORAGE_KEYS.user)
  const storedToken = localStorage.getItem(STORAGE_KEYS.token)
  
  if (storedUser && storedToken) {
    try {
      user.value = JSON.parse(storedUser)
      token.value = storedToken
      return true
    } catch (error) {
      console.error('Failed to restore session:', error)
      clearUser()
    }
  }
  
  return false
}

/**
 * Check if user is authenticated
 * @returns {boolean}
 */
export function isAuthenticated() {
  return !!(user.value && token.value)
}

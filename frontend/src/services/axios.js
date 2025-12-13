import axios from 'axios'
import { clearUser } from '@/stores/auth'
import router from '@/router'

/**
 * Global axios response interceptor to handle token expiration
 * Automatically logs out users when receiving 401 Unauthorized responses
 */
axios.interceptors.response.use(
  // Success handler - pass through
  response => response,
  
  // Error handler - check for 401
  error => {
    if (error.response?.status === 401) {
      // Token is invalid or expired
      clearUser() // Clear localStorage and memory
      
      // Redirect to login with expiration message
      router.push('/auth?expired=true')
      
      // Optionally show a toast notification
      // (requires toast system to be available globally)
      console.warn('Session expired. Please login again.')
    }
    
    return Promise.reject(error)
  }
)

export default axios

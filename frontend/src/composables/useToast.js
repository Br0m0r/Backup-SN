import { ref } from 'vue'

// Global toast state - shared across all components
const toasts = ref([])
let nextId = 0

export function useToast() {
  const add = (message, type = 'info', duration = 4000) => {
    const id = nextId++
    const toast = { id, message, type, visible: true }
    
    toasts.value.push(toast)
    
    // Auto-dismiss after duration
    if (duration > 0) {
      setTimeout(() => {
        remove(id)
      }, duration)
    }
    
    return id
  }
  
  const remove = (id) => {
    const index = toasts.value.findIndex(t => t.id === id)
    if (index !== -1) {
      // Fade out animation
      toasts.value[index].visible = false
      // Remove from DOM after animation
      setTimeout(() => {
        toasts.value = toasts.value.filter(t => t.id !== id)
      }, 300)
    }
  }
  
  const success = (message, duration) => add(message, 'success', duration)
  const error = (message, duration) => add(message, 'error', duration)
  const info = (message, duration) => add(message, 'info', duration)
  const warning = (message, duration) => add(message, 'warning', duration)
  
  return {
    toasts,
    success,
    error,
    info,
    warning,
    remove
  }
}

import { createRouter, createWebHistory } from 'vue-router'
import { isAuthenticated } from '../stores/auth'

const routes = [
  {
    path: '/auth',
    name: 'Auth',
    component: () => import('../pages/AuthView.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/feed',
    name: 'Feed',
    component: () => import('../pages/FeedView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/post/:id',
    name: 'Post',
    component: () => import('../pages/PostView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/',
    redirect: '/feed'
  },
  {
    path: '/profile/:id?',
    name: 'Profile',
    component: () => import('../pages/ProfileView.vue'),
    props: true,
    meta: { requiresAuth: true }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/feed'
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    }
    return { top: 0 }
  }
})

// Navigation guard for protected routes
router.beforeEach((to, from, next) => {
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth)
  const authenticated = isAuthenticated()

  if (requiresAuth && !authenticated) {
    // Redirect to auth if trying to access protected route while not logged in
    next({ name: 'Auth' })
  } else if (to.name === 'Auth' && authenticated) {
    // Redirect to feed if trying to access auth while logged in
    next({ name: 'Feed' })
  } else {
    next()
  }
})

export default router

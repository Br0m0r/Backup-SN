<template>
  <div v-if="isAuthenticated" class="neon-shell">
    <header class="neon-header">
      <div class="brand" role="button" tabindex="0" @click="goToFeed">
        <span class="pulse-dot"></span>
        <span>Neon Connex</span>
      </div>
      <div class="header-actions">
        <!-- Search Users Dropdown -->
        <div class="header-icon search-dropdown" @click="toggleSearch" role="button" tabindex="0">
          <span class="icon icon-search"></span>
          <SuggestedUsers v-if="searchOpen" @close="closeSearch" />
        </div>

        <!-- Notifications Widget -->
        <Notifications />

        <!-- Profile Menu -->
        <div class="header-icon avatar" @click="toggleProfile" role="button" tabindex="0">
          <img src="https://placehold.co/64x64/11121f/fff?text=ME" alt="profile" />
          <div v-if="profileOpen" class="dropdown profile-menu">
            <button class="profile-btn" @click="viewProfile">View Profile</button>
            <button class="ghost" @click="logout">Logout</button>
          </div>
        </div>
      </div>
    </header>

    <!-- Routed content when authenticated -->
    <main class="content">
      <router-view />
    </main>

    <!-- Chat Component - Fixed at bottom right -->
    <Chat />
    
    <!-- Toast Notifications - Global -->
    <ToastContainer />
  </div>

  <!-- Routed content when not authenticated -->
  <div v-else class="auth-gate">
    <router-view />
    
    <!-- Toast Notifications - Available in auth screens too -->
    <ToastContainer />
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import Chat from './components/Chat.vue';
import SuggestedUsers from './components/SuggestedUsers.vue';
import Notifications from './components/Notifications.vue';
import ToastContainer from './components/ToastContainer.vue';
import { clearUser, getToken, isAuthenticated as hasSession, restoreSession } from './stores/auth';
import { logoutUser } from './services/authService';

const router = useRouter();
const profileOpen = ref(false);
const searchOpen = ref(false);
const isAuthenticated = computed(() => hasSession());

onMounted(() => {
  restoreSession();
});

function toggleProfile() {
  profileOpen.value = !profileOpen.value;
  searchOpen.value = false;
}

function toggleSearch() {
  searchOpen.value = !searchOpen.value;
  profileOpen.value = false;
}

function closeSearch() {
  searchOpen.value = false;
}

function goToFeed() {
  router.push({ name: 'Feed' });
  profileOpen.value = false;
}

function viewProfile() {
  router.push({ name: 'Profile' });
  profileOpen.value = false;
}

async function logout() {
  try {
    const token = getToken();
    if (token) {
      await logoutUser(token);
    }
  } catch (error) {
    console.error('Failed to logout:', error);
  } finally {
    clearUser();
    profileOpen.value = false;
    router.push({ name: 'Auth' });
  }
}
</script>

<style scoped>
.neon-shell {
  min-height: 100vh;
  padding: 2rem clamp(1.25rem, 5vw, 3.5rem);
  display: flex;
  flex-direction: column;
  gap: 2rem;
  position: relative;
}

.neon-header {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  background: linear-gradient(135deg, rgba(0, 247, 255, 0.08), rgba(255, 0, 230, 0.08));
  border-radius: 1.25rem;
  box-shadow: 0 20px 50px rgba(2, 4, 12, 0.65);
  backdrop-filter: blur(16px);
}

.brand {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.05em;
  cursor: pointer;
}

.pulse-dot {
  width: 0.75rem;
  height: 0.75rem;
  border-radius: 999px;
  background: var(--neon-cyan);
  box-shadow: 0 0 12px var(--neon-cyan);
  animation: pulse 1.8s infinite;
}

.header-actions {
  display: flex;
  gap: 0.75rem;
  position: relative;
}

.header-actions {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.header-icon {
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 999px;
  display: grid;
  place-items: center;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.08);
  cursor: pointer;
  transition: border-color 0.2s ease, transform 0.2s ease;
  position: relative;
}

.header-icon:hover {
  border-color: var(--border-glow);
  transform: scale(1.05);
}

.header-icon.search-dropdown {
  position: relative;
}

.icon-search {
  display: inline-block;
  width: 1.2rem;
  height: 1.2rem;
  mask: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="white" d="M15.5 14h-.79l-.28-.27a6.5 6.5 0 0 0 1.48-5.34c-.47-2.78-2.79-5-5.59-5.34a6.505 6.505 0 0 0-7.27 7.27c.34 2.8 2.56 5.12 5.34 5.59a6.5 6.5 0 0 0 5.34-1.48l.27.28v.79l4.25 4.25c.41.41 1.08.41 1.49 0 .41-.41.41-1.08 0-1.49L15.5 14zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/></svg>') center/contain no-repeat;
  background: var(--neon-cyan);
}

.header-icon.avatar {
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 999px;
  display: grid;
  place-items: center;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.08);
  cursor: pointer;
  transition: border-color 0.2s ease, transform 0.2s ease;
  position: relative;
}

.header-icon:hover {
  border-color: var(--border-glow);
  transform: scale(1.05);
}

.header-icon.search-dropdown {
  position: relative;
}

.icon-search {
  display: inline-block;
  width: 1.2rem;
  height: 1.2rem;
  mask: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="white" d="M15.5 14h-.79l-.28-.27a6.5 6.5 0 0 0 1.48-5.34c-.47-2.78-2.79-5-5.59-5.34a6.505 6.505 0 0 0-7.27 7.27c.34 2.8 2.56 5.12 5.34 5.59a6.5 6.5 0 0 0 5.34-1.48l.27.28v.79l4.25 4.25c.41.41 1.08.41 1.49 0 .41-.41.41-1.08 0-1.49L15.5 14zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/></svg>') center/contain no-repeat;
  background: var(--neon-cyan);
}

.header-icon.avatar {
  position: relative;
  width: 3.25rem;
  height: 3.25rem;
  border-radius: 1rem;
  background: rgba(8, 10, 22, 0.85);
  border: 1px solid rgba(255, 255, 255, 0.08);
  display: grid;
  place-items: center;
  cursor: pointer;
  transition: transform 0.2s ease, border-color 0.2s ease;
}

.header-icon:hover {
  transform: translateY(-2px);
  border-color: var(--border-glow);
}

.header-icon.avatar {
  padding: 0.35rem;
}

.header-icon img {
  width: 100%;
  height: 100%;
  border-radius: inherit;
  object-fit: cover;
}

.badge {
  position: absolute;
  top: -0.35rem;
  right: -0.35rem;
  background: var(--neon-pink);
  color: #05060d;
  font-size: 0.65rem;
  font-weight: 700;
  border-radius: 999px;
  padding: 0.15rem 0.45rem;
  box-shadow: 0 0 14px rgba(255, 0, 230, 0.55);
}

.dropdown {
  position: absolute;
  right: 0;
  top: 110%;
  width: clamp(14rem, 25vw, 19rem);
  border-radius: 1rem;
  padding: 1rem;
  background: rgba(5, 6, 13, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.12);
  box-shadow: var(--shadow);
  backdrop-filter: blur(20px);
}

.dropdown.notifications ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.dropdown.notifications li {
  padding: 0.75rem;
  background: rgba(16, 18, 32, 0.6);
  border-radius: 0.85rem;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.dropdown.notifications .note-type {
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--neon-cyan);
}

.dropdown.notifications small {
  color: var(--text-muted);
}

.notifications-panel {
  top: 120%;
  right: 0;
  width: clamp(16rem, 30vw, 20rem);
}

.ghost.mini.full-width {
  width: 100%;
  justify-content: center;
}

.profile-menu {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.profile-menu button {
  width: 100%;
}

.profile-btn {
  background: transparent;
  border: 1px solid rgba(0, 247, 255, 0.3);
  color: var(--neon-cyan);
  border-radius: 0.5rem;
  padding: 0.6rem 1rem;
  cursor: pointer;
  transition: all 0.2s ease;
  font-weight: 500;
}

.profile-btn:hover {
  background: rgba(0, 247, 255, 0.1);
  border-color: var(--neon-cyan);
  box-shadow: 0 0 10px rgba(0, 247, 255, 0.2);
}

.content {
  display: flex;
  justify-content: center;
  width: 100%;
}

.profile-wrapper {
  width: min(1100px, 100%);
  margin: 0 auto;
}

.auth-gate {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: clamp(1.5rem, 5vw, 4rem);
  background: radial-gradient(circle at top, rgba(0, 247, 255, 0.08), rgba(5, 6, 13, 0.95));
}

.filter-btn {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  padding: 0.85rem 1rem;
  border-radius: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(7, 9, 20, 0.6);
  color: inherit;
  cursor: pointer;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  font-weight: 600;
  transition: background 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease;
}

.filter-btn.active {
  border-color: var(--border-glow);
  box-shadow: 0 0 22px rgba(0, 247, 255, 0.2);
  background: rgba(0, 247, 255, 0.1);
}

.main-panel {
  width: min(900px, 100%);
  border-radius: 1.5rem;
  padding: clamp(1rem, 3vw, 2rem);
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(5, 6, 13, 0.85);
  box-shadow: var(--shadow);
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.filters-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 0.85rem;
}

.post-stack {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.post-card {
  padding: 1.5rem;
  border-radius: 1.25rem;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(8, 10, 24, 0.8);
  box-shadow: 0 12px 25px rgba(0, 0, 0, 0.35);
  backdrop-filter: blur(12px);
}

.post-card header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.85rem;
}

.author {
  display: flex;
  gap: 0.85rem;
  align-items: center;
}

.author img {
  width: 2.75rem;
  height: 2.75rem;
  border-radius: 999px;
  border: 2px solid rgba(255, 255, 255, 0.08);
}

.author small {
  color: var(--text-muted);
}

.ghost {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.15);
  color: inherit;
  border-radius: 999px;
  padding: 0.35rem 0.85rem;
  cursor: pointer;
}

.search-pane {
  display: flex;
  flex-direction: column;
  gap: 1.2rem;
}

.search-input {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem 1.25rem;
  border-radius: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(0, 0, 0, 0.45);
}

.search-input input {
  background: transparent;
  border: none;
  flex: 1;
  color: inherit;
  font-size: 1rem;
}

.search-input input:focus {
  outline: none;
}

.suggestions {
  border-radius: 1.25rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  padding: 1.25rem;
  background: rgba(8, 10, 24, 0.75);
}

.suggestion-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 0.85rem;
  padding: 0.85rem;
  border-radius: 1rem;
  background: rgba(10, 12, 22, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.suggestion-card .user {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.suggestion-card img {
  width: 3rem;
  height: 3rem;
  border-radius: 999px;
}

.dropdown-title {
  margin: 0 0 0.5rem;
  font-weight: 600;
}

.icon {
  display: inline-block;
  width: 1.1rem;
  height: 1.1rem;
}

.icon-bell {
  background: radial-gradient(circle, var(--neon-cyan) 0%, rgba(0, 0, 0, 0) 70%);
  mask: url('data:image/svg+xml,<svg xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 24 24\"><path fill=\"white\" d=\"M18 8a6 6 0 0 0-12 0c0 7-3 9-3 9h18s-3-2-3-9zm-6 14a2.5 2.5 0 0 0 2.45-2h-4.9A2.5 2.5 0 0 0 12 22\"/></svg>')
    center / contain no-repeat;
}

.icon-globe {
  background: radial-gradient(circle, var(--neon-cyan), rgba(0, 0, 0, 0));
  mask: url('data:image/svg+xml,<svg xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 24 24\"><path fill=\"white\" d=\"M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm6.93 8h-3.54a16.06 16.06 0 0 0-1.05-4.29A8.026 8.026 0 0 1 18.93 10zM12 4c.83 0 2.36 2.17 2.82 6H9.18C9.64 6.17 11.17 4 12 4zm-3.34.71A16.06 16.06 0 0 0 7.61 10H4.07a8.026 8.026 0 0 1 4.59-5.29zM4.07 14h3.54c.21 1.49.67 2.98 1.39 4.29A8.026 8.026 0 0 1 4.07 14zm4.75 0h6.36c-.46 3.83-1.99 6-2.82 6s-2.36-2.17-2.82-6zm7.57 4.29c.72-1.31 1.18-2.8 1.39-4.29h3.54a8.026 8.026 0 0 1-4.93 4.29z\"/></svg>')
    center / contain no-repeat;
}

.icon-home {
  background: radial-gradient(circle, var(--neon-pink), rgba(0, 0, 0, 0));
  mask: url('data:image/svg+xml,<svg xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 24 24\"><path fill=\"white\" d=\"m12 3l9 8h-3v9h-5v-5H11v5H6v-9H3l9-8z\"/></svg>')
    center / contain no-repeat;
}

.icon-search {
  background: radial-gradient(circle, var(--neon-cyan), rgba(0, 0, 0, 0));
  mask: url('data:image/svg+xml,<svg xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 24 24\"><path fill=\"white\" d=\"m15.5 14h-.79l-.28-.27A6.471 6.471 0 0 0 16 9.5 6.5 6.5 0 1 0 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79L20 21.49 21.49 20 15.5 14zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z\"/></svg>')
    center / contain no-repeat;
}

@media (max-width: 960px) {
  .main-panel {
    width: 100%;
  }
}

@media (max-width: 640px) {
  .neon-shell {
    padding: 1.25rem;
  }

  .neon-header {
    flex-direction: column;
    gap: 1rem;
  }

  .filters-row {
    grid-template-columns: 1fr;
  }
}

@keyframes pulse {
  0% {
    transform: scale(0.9);
    opacity: 0.85;
  }
  50% {
    transform: scale(1.1);
    opacity: 1;
  }
  100% {
    transform: scale(0.9);
    opacity: 0.85;
  }
}
</style>

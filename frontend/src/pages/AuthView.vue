<template>
  <section class="auth-card">
    <header>
      <p class="eyebrow">@neonconnex</p>
      <h1>Log in or register</h1>
      <p class="intro">
        Claim your handle, sync devices, and keep your neon feed glowing. Backend wiring will plug
        in soon—this panel focuses on the experience.
      </p>
    </header>

    <div class="auth-toggle">
      <button :class="{ active: activeTab === 'login' }" @click="activeTab = 'login'">Login</button>
      <button :class="{ active: activeTab === 'register' }" @click="activeTab = 'register'">Register</button>
    </div>

    <Transition name="fade-slide" mode="out-in">
      <form v-if="activeTab === 'login'" key="login" class="auth-form" @submit.prevent="handleLogin">
        <label>
          <span>Email</span>
          <input
            type="email"
            v-model="loginForm.email"
            placeholder="pilot@neon.city"
            required
            autocomplete="email"
          />
        </label>
        <label>
          <span>Password</span>
          <input
            type="password"
            v-model="loginForm.password"
            placeholder="••••••••"
            required
            autocomplete="current-password"
            minlength="6"
          />
        </label>
        <div class="form-actions">
          <label class="remember">
            <input type="checkbox" v-model="loginForm.remember" />
            Keep me signed in
          </label>
          <button type="button" class="link-btn">Forgot?</button>
        </div>
        <button class="cta full" type="submit" :disabled="loginLoading">
          {{ loginLoading ? 'Entering...' : 'Enter feed' }}
        </button>
      </form>

      <form v-else key="register" class="auth-form" @submit.prevent="handleRegister">
        <label>
          <span>Email</span>
          <input type="email" v-model="registerForm.email" placeholder="you@neon.city" required autocomplete="email" />
        </label>
        <div class="split">
          <label>
            <span>Password</span>
            <input type="password" v-model="registerForm.password" placeholder="Create a passphrase" minlength="6" required autocomplete="new-password" />
          </label>
          <label>
            <span>Confirm</span>
            <input
              type="password"
              v-model="registerForm.confirm"
              placeholder="Repeat password"
              minlength="6"
              required
              autocomplete="new-password"
            />
          </label>
        </div>
        <div class="split">
          <label>
            <span>First name</span>
            <input type="text" v-model="registerForm.firstName" placeholder="Marina" required />
          </label>
          <label>
            <span>Last name</span>
            <input type="text" v-model="registerForm.lastName" placeholder="Pulse" required />
          </label>
        </div>
        <label>
          <span>Date of birth</span>
          <input type="date" v-model="registerForm.dateOfBirth" required />
        </label>
        <label>
          <span>Avatar/Image <span class="optional">(Optional)</span></span>
          <input type="file" @change="handleAvatarChange" accept="image/*" />
        </label>
        <label>
          <span>Nickname <span class="optional">(Optional)</span></span>
          <input type="text" v-model="registerForm.nickname" placeholder="CoolNickname" />
        </label>
        <label>
          <span>About me <span class="optional">(Optional)</span></span>
          <textarea v-model="registerForm.aboutMe" placeholder="Tell us about yourself..." rows="3"></textarea>
        </label>
        <button class="cta full" type="submit" :disabled="registerLoading">
          {{ registerLoading ? 'Creating account...' : 'Create account' }}
        </button>
      </form>
    </Transition>

    <p v-if="feedback.message" :class="['auth-message', feedback.variant]">{{ feedback.message }}</p>

    <button class="ghost full support" type="button">Need help? Contact support</button>
  </section>
</template>

<script setup>
import { reactive, ref, watch } from 'vue'
import { loginUser, registerUser } from '../services/authService'
import { setUser } from '../stores/auth'

const emit = defineEmits(['authenticated'])

const activeTab = ref('login')
const loginLoading = ref(false)
const registerLoading = ref(false)

const feedback = reactive({
  message: '',
  variant: 'info'
})

const loginForm = reactive({
  email: '',
  password: '',
  remember: true
})

const registerForm = reactive({
  firstName: '',
  lastName: '',
  email: '',
  password: '',
  confirm: '',
  dateOfBirth: '',
  avatar: null,
  nickname: '',
  aboutMe: ''
})

function setFeedback(message = '', variant = 'info') {
  feedback.message = message
  feedback.variant = variant
}

function sanitizeHandle(value) {
  return value.trim().replace(/^@+/, '').toLowerCase()
}

function handleAvatarChange(event) {
  const file = event.target.files?.[0]
  registerForm.avatar = file || null
}

watch(activeTab, () => setFeedback())

async function handleLogin() {
  if (loginLoading.value) return

  setFeedback()
  loginLoading.value = true

  try {
    const payload = {
      email: loginForm.email.trim().toLowerCase(),
      password: loginForm.password
    }

    const { user, token } = await loginUser(payload)
    setUser(user, token, { persist: loginForm.remember })

    const welcomeName = user.first_name || user.username || loginForm.email
    setFeedback(`Welcome back, ${welcomeName}! Redirecting to your feed...`, 'success')
    emit('authenticated', { mode: 'login', user })
  } catch (error) {
    const message = error.response?.data?.error || error.message || 'Failed to login. Please try again.'
    setFeedback(message, 'error')
  } finally {
    loginLoading.value = false
  }
}

async function handleRegister() {
  if (registerLoading.value) return

  setFeedback()

  if (registerForm.password !== registerForm.confirm) {
    setFeedback('Passwords do not match. Try again.', 'error')
    return
  }

  const firstName = registerForm.firstName.trim()
  const lastName = registerForm.lastName.trim()
  const email = registerForm.email.trim().toLowerCase()
  const dateOfBirth = registerForm.dateOfBirth

  if (!firstName || !lastName) {
    setFeedback('First and last name are required.', 'error')
    return
  }

  if (!email) {
    setFeedback('Email is required.', 'error')
    return
  }

  if (!dateOfBirth) {
    setFeedback('Date of birth is required.', 'error')
    return
  }

  registerLoading.value = true

  try {
    const payload = {
      username: '', // Backend will generate from email if empty
      email,
      password: registerForm.password,
      first_name: firstName,
      last_name: lastName,
      date_of_birth: dateOfBirth,
      nickname: registerForm.nickname.trim() || undefined,
      about_me: registerForm.aboutMe.trim() || undefined
    }

    const { user, token } = await registerUser(payload)
    setUser(user, token)

    const welcomeName = user.first_name || email
    setFeedback(`Welcome aboard, ${welcomeName}! Taking you to your feed...`, 'success')
    emit('authenticated', { mode: 'register', user })
  } catch (error) {
    const message = error.response?.data?.error || error.message || 'Unable to create your account right now.'
    setFeedback(message, 'error')
  } finally {
    registerLoading.value = false
  }
}
</script>

<style scoped>
.auth-card {
  width: min(520px, 100%);
  margin: 0 auto;
  padding: clamp(1.5rem, 4vw, 2.5rem);
  border-radius: 1.75rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(6, 8, 18, 0.9);
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  box-shadow: 0 30px 60px rgba(0, 0, 0, 0.45);
}

.auth-card header h1 {
  margin: 0.25rem 0;
  font-size: clamp(1.8rem, 3vw, 2.4rem);
}

.eyebrow {
  text-transform: uppercase;
  letter-spacing: 0.2em;
  font-size: 0.75rem;
  color: var(--neon-cyan);
}

.intro {
  color: var(--text-muted);
  max-width: 42ch;
}

.auth-toggle {
  display: grid;
  grid-template-columns: repeat(2, minmax(120px, 1fr));
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  padding: 0.35rem;
  background: rgba(0, 0, 0, 0.35);
}

.auth-toggle button {
  border: none;
  border-radius: 999px;
  padding: 0.65rem 1rem;
  background: transparent;
  color: inherit;
  cursor: pointer;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.auth-toggle button.active {
  background: linear-gradient(120deg, var(--neon-cyan), var(--neon-pink));
  color: #05060d;
  box-shadow: 0 8px 22px rgba(255, 0, 230, 0.28);
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.auth-form label {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.auth-form input {
  padding: 0.75rem 1rem;
  border-radius: 0.9rem;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(8, 10, 24, 0.85);
  color: inherit;
}

.auth-form textarea {
  padding: 0.75rem 1rem;
  border-radius: 0.9rem;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(8, 10, 24, 0.85);
  color: inherit;
  font-family: inherit;
  resize: vertical;
  min-height: 80px;
}

.auth-form input:focus,
.auth-form textarea:focus {
  outline: none;
  border-color: var(--neon-cyan);
  box-shadow: 0 0 14px rgba(0, 247, 255, 0.2);
}

.optional {
  color: var(--text-muted);
  font-size: 0.85em;
  font-weight: normal;
}

.cta {
  border: none;
  border-radius: 999px;
  padding: 0.75rem 1.4rem;
  background: linear-gradient(120deg, var(--neon-cyan), var(--neon-pink));
  color: #05060d;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 12px 30px rgba(255, 0, 230, 0.25);
}

.cta:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.ghost {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: inherit;
  border-radius: 999px;
  padding: 0.65rem 1.2rem;
  cursor: pointer;
}

.split {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 0.75rem;
}

.form-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  font-size: 0.9rem;
}

.remember {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.remember input {
  width: 1rem;
  height: 1rem;
}

.link-btn {
  border: none;
  background: transparent;
  color: var(--neon-cyan);
  cursor: pointer;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.cta.full,
.ghost.full {
  width: 100%;
  justify-content: center;
}

.auth-message {
  margin: 0;
  padding: 0.85rem 1rem;
  border-radius: 0.85rem;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(8, 10, 24, 0.85);
  font-size: 0.95rem;
  text-align: center;
}

.auth-message.success {
  border-color: rgba(0, 247, 255, 0.35);
  color: var(--neon-cyan);
  box-shadow: 0 0 14px rgba(0, 247, 255, 0.2);
}

.auth-message.error {
  border-color: rgba(255, 0, 230, 0.35);
  color: var(--neon-pink);
  box-shadow: 0 0 14px rgba(255, 0, 230, 0.15);
}

.auth-message.info {
  color: var(--text-muted);
}

.support {
  margin-top: 0.5rem;
}

.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.fade-slide-enter-from,
.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(12px);
}

@media (max-width: 640px) {
  .auth-card {
    width: 100%;
    border-radius: 1.5rem;
  }
}
</style>

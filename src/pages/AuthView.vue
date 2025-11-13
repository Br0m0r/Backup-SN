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
          <input type="email" v-model="loginForm.email" placeholder="pilot@neon.city" required autocomplete="email" />
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
        <button class="cta full" type="submit">Enter feed</button>
      </form>

      <form v-else key="register" class="auth-form" @submit.prevent="handleRegister">
        <label>
          <span>Display name</span>
          <input type="text" v-model="registerForm.name" placeholder="Marina Pulse" required />
        </label>
        <label>
          <span>Handle</span>
          <input type="text" v-model="registerForm.handle" placeholder="@neonpilot" required />
        </label>
        <label>
          <span>Email</span>
          <input type="email" v-model="registerForm.email" placeholder="you@neon.city" required autocomplete="email" />
        </label>
        <div class="split">
          <label>
            <span>Password</span>
            <input type="password" v-model="registerForm.password" placeholder="Create a passphrase" minlength="6" required />
          </label>
          <label>
            <span>Confirm</span>
            <input
              type="password"
              v-model="registerForm.confirm"
              placeholder="Repeat password"
              minlength="6"
              required
            />
          </label>
        </div>
        <button class="cta full" type="submit">Create account</button>
      </form>
    </Transition>

    <p v-if="authMessage" class="auth-message">{{ authMessage }}</p>

    <button class="ghost full support" type="button">Need help? Contact support</button>
  </section>
</template>

<script setup>
import { reactive, ref } from 'vue';

const emit = defineEmits(['authenticated']);

const activeTab = ref('login');
const authMessage = ref('');

const loginForm = reactive({
  email: '',
  password: '',
  remember: true,
});

const registerForm = reactive({
  name: '',
  handle: '',
  email: '',
  password: '',
  confirm: '',
});

function handleLogin() {
  authMessage.value = `Welcome back, ${loginForm.email || 'pilot'}! Redirecting to your feed...`;
  emit('authenticated', { mode: 'login', email: loginForm.email });
}

function handleRegister() {
  if (registerForm.password !== registerForm.confirm) {
    authMessage.value = 'Passwords do not match. Try again.';
    return;
  }
  authMessage.value = `Creating ${registerForm.handle || registerForm.email}... redirecting shortly.`;
  emit('authenticated', {
    mode: 'register',
    email: registerForm.email,
    handle: registerForm.handle,
  });
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

.auth-form input:focus {
  outline: none;
  border-color: var(--neon-cyan);
  box-shadow: 0 0 14px rgba(0, 247, 255, 0.2);
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
  padding: 0.85rem;
  border-radius: 0.85rem;
  background: rgba(0, 247, 255, 0.08);
  color: var(--neon-cyan);
  font-size: 0.9rem;
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

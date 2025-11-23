<template>
  <div class="toast-test">
    <div class="test-container">
      <h1>🎨 Toast Notification Tester</h1>
      <p>Click the buttons below to test different toast types</p>
      
      <div class="button-grid">
        <button @click="testSuccess" class="btn btn-success">
          ✓ Test Success Toast
        </button>
        
        <button @click="testError" class="btn btn-error">
          ✕ Test Error Toast
        </button>
        
        <button @click="testWarning" class="btn btn-warning">
          ⚠ Test Warning Toast
        </button>
        
        <button @click="testInfo" class="btn btn-info">
          ℹ Test Info Toast
        </button>
        
        <button @click="testMultiple" class="btn btn-multiple">
          🎉 Test Multiple Toasts
        </button>
        
        <button @click="testLongMessage" class="btn btn-long">
          📝 Test Long Message
        </button>
      </div>
      
      <div class="back-link">
        <button @click="$router.push('/feed')" class="btn-back">
          ← Back to Feed
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useToast } from '@/composables/useToast'

const { success, error, warning, info } = useToast()

function testSuccess() {
  success('Operation completed successfully!')
}

function testError() {
  error('Something went wrong. Please try again.')
}

function testWarning() {
  warning('This action cannot be undone!')
}

function testInfo() {
  info('Here is some useful information for you.')
}

function testMultiple() {
  success('First notification')
  setTimeout(() => error('Second notification'), 300)
  setTimeout(() => warning('Third notification'), 600)
  setTimeout(() => info('Fourth notification'), 900)
}

function testLongMessage() {
  error('This is a much longer error message to test how the toast handles text that spans multiple lines and contains quite a bit of content.')
}
</script>

<style scoped>
.toast-test {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
}

.test-container {
  max-width: 600px;
  width: 100%;
  background: rgba(8, 10, 24, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  padding: 3rem;
  text-align: center;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
}

h1 {
  font-size: 2rem;
  margin-bottom: 0.5rem;
  background: linear-gradient(135deg, var(--neon-cyan), var(--neon-pink));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

p {
  color: rgba(255, 255, 255, 0.7);
  margin-bottom: 2rem;
}

.button-grid {
  display: grid;
  gap: 1rem;
  margin-bottom: 2rem;
}

.btn {
  padding: 1rem 1.5rem;
  border: none;
  border-radius: 12px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  color: white;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.4);
}

.btn:active {
  transform: translateY(0);
}

.btn-success {
  background: linear-gradient(135deg, #10b981, #059669);
}

.btn-error {
  background: linear-gradient(135deg, #ef4444, #dc2626);
}

.btn-warning {
  background: linear-gradient(135deg, #f59e0b, #d97706);
}

.btn-info {
  background: linear-gradient(135deg, var(--neon-cyan), var(--neon-pink));
}

.btn-multiple {
  background: linear-gradient(135deg, #8b5cf6, #ec4899);
}

.btn-long {
  background: linear-gradient(135deg, #6366f1, #3b82f6);
}

.back-link {
  margin-top: 2rem;
  padding-top: 2rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.btn-back {
  padding: 0.75rem 1.5rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 10px;
  color: rgba(255, 255, 255, 0.9);
  font-size: 0.95rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-back:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: var(--neon-cyan);
  transform: translateX(-4px);
}
</style>

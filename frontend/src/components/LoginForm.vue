<template>
  <div class="card">
    <h2>Login</h2>
    <div v-if="errorMessage" style="background:#3b1d2a;color:#ffb4c0;border:1px solid #7a2a3e;padding:8px 12px;border-radius:8px;margin-bottom:8px;">
      {{ errorMessage }}
    </div>
    <div v-if="successMessage" style="background:#0f2a1a;color:#85e0a3;border:1px solid #2a5a3e;padding:8px 12px;border-radius:8px;margin-bottom:8px;">
      {{ successMessage }}
    </div>
    <form class="row" @submit.prevent="login" style="flex-direction:column; gap:12px;">
      <input class="text" v-model="email" placeholder="Email" type="email" required />
      <input class="text" v-model="password" placeholder="Password" type="password" required />
      <button class="primary" type="submit" :disabled="isSubmitting">
        {{ isSubmitting ? 'Logging in...' : 'Login' }}
      </button>
    </form>
    <p class="muted" style="margin-top:8px">API: {{ apiBase }}/api/login</p>
    
    <div v-if="user" style="margin-top:16px; padding:12px; background:var(--input-bg); border-radius:8px;">
      <h3>Welcome, {{ user.name }}!</h3>
      <p class="muted">{{ user.email }}</p>
      <button class="ghost" @click="logout" style="margin-top:8px;">Logout</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { ref } from 'vue'

type User = { id: number; email: string; name: string; created_at: string }

const apiBase = (import.meta as any).env.VITE_API_BASE || 'http://localhost:8000'
const http = axios.create({ baseURL: apiBase })

const email = ref('')
const password = ref('')
const errorMessage = ref('')
const successMessage = ref('')
const isSubmitting = ref(false)
const user = ref<User | null>(null)

async function login() {
  errorMessage.value = ''
  successMessage.value = ''
  
  const e = email.value.trim()
  const p = password.value.trim()
  
  if (!e || !p) {
    errorMessage.value = 'Email and password are required.'
    return
  }
  
  // Simple email validation
  if (!/^\S+@\S+\.\S+$/.test(e)) {
    errorMessage.value = 'Invalid email format.'
    return
  }
  
  isSubmitting.value = true
  try {
    const { data } = await http.post('/api/login', { email: e, password: p })
    if (data.success) {
      user.value = data
      successMessage.value = `Login successful! Welcome ${data.name}.`
      email.value = ''
      password.value = ''
      // Store login state in localStorage for persistence
      localStorage.setItem('user', JSON.stringify(data))
    }
  } catch (err: any) {
    errorMessage.value = extractError(err)
  } finally {
    isSubmitting.value = false
  }
}

function logout() {
  user.value = null
  successMessage.value = ''
  errorMessage.value = ''
  localStorage.removeItem('user')
}

// Check for existing login state on component mount
const savedUser = localStorage.getItem('user')
if (savedUser) {
  try {
    user.value = JSON.parse(savedUser)
  } catch (e) {
    localStorage.removeItem('user')
  }
}

function extractError(err: any): string {
  if (!err) return 'An unexpected error occurred.'
  const res = err.response
  if (res && res.data) {
    if (typeof res.data === 'string') return res.data
    if (res.data.error) return String(res.data.error)
    try { return JSON.stringify(res.data) } catch { /* noop */ }
  }
  if (err.message) return String(err.message)
  return 'An error occurred.'
}
</script>
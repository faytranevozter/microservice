<template>
  <div class="card">
    <h2>Users</h2>
    <div v-if="errorMessage" style="background:#3b1d2a;color:#ffb4c0;border:1px solid #7a2a3e;padding:8px 12px;border-radius:8px;margin-bottom:8px;">
      {{ errorMessage }}
    </div>
    <form class="row" @submit.prevent="addUser" style="flex-direction:column; gap:8px;">
      <div class="row">
        <input class="text" v-model="email" placeholder="email" />
        <input class="text" v-model="name" placeholder="name" />
      </div>
      <div class="row">
        <input class="text" v-model="password" placeholder="password (optional)" type="password" />
        <button class="primary" type="submit" :disabled="isSubmitting">{{ isSubmitting ? 'Adding...' : 'Add' }}</button>
      </div>
    </form>
    <p class="muted" style="margin-top:8px">API: {{ apiBase }}/api/users</p>
    <ul style="list-style:none; padding:0; margin-top:12px; display:flex; flex-direction:column; gap:8px;">
      <li v-for="u in users" :key="u.id" class="row" style="align-items:center;">
        <span style="flex:1">{{ u.name }} <span class="muted">&lt;{{ u.email }}&gt;</span></span>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { onMounted, ref } from 'vue'

type User = { id:number; email:string; name:string }

const apiBase = (import.meta as any).env.VITE_API_BASE || 'http://localhost:8000'
const http = axios.create({ baseURL: apiBase })

const users = ref<User[]>([])
const email = ref('')
const name = ref('')
const password = ref('')
const errorMessage = ref('')
const isSubmitting = ref(false)

async function load() {
  try {
    const { data } = await http.get('/api/users')
    users.value = data
  } catch (e: any) {
    errorMessage.value = extractError(e)
  }
}

async function addUser() {
  errorMessage.value = ''
  const e = email.value.trim()
  const n = name.value.trim()
  const p = password.value.trim()
  if (!e || !n) {
    errorMessage.value = 'Email dan nama wajib diisi.'
    return
  }
  // validasi email sederhana
  if (!/^\S+@\S+\.\S+$/.test(e)) {
    errorMessage.value = 'Format email tidak valid.'
    return
  }
  isSubmitting.value = true
  try {
    const payload: any = { email: e, name: n }
    if (p) {
      payload.password = p
    }
    await http.post('/api/users', payload)
    email.value = ''
    name.value = ''
    password.value = ''
    await load()
  } catch (er: any) {
    errorMessage.value = extractError(er)
  } finally {
    isSubmitting.value = false
  }
}

onMounted(load)

function extractError(err: any): string {
  if (!err) return 'Terjadi kesalahan tak terduga.'
  const res = err.response
  if (res && res.data) {
    if (typeof res.data === 'string') return res.data
    if (res.data.error) return String(res.data.error)
    try { return JSON.stringify(res.data) } catch { /* noop */ }
  }
  if (err.message) return String(err.message)
  return 'Terjadi kesalahan.'
}
</script>



<template>
  <div class="card">
    <h2>Users</h2>
    <form class="row" @submit.prevent="addUser">
      <input class="text" v-model="email" placeholder="email" />
      <input class="text" v-model="name" placeholder="name" />
      <button class="primary" type="submit">Add</button>
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

async function load() {
  const { data } = await http.get('/api/users')
  users.value = data
}

async function addUser() {
  const e = email.value.trim()
  const n = name.value.trim()
  if (!e || !n) return
  await http.post('/api/users', { email: e, name: n })
  email.value = ''
  name.value = ''
  await load()
}

onMounted(load)
</script>



<template>
  <div class="card">
    <h2>Todos</h2>
    <div v-if="errorMessage" style="background:#3b1d2a;color:#ffb4c0;border:1px solid #7a2a3e;padding:8px 12px;border-radius:8px;margin-bottom:8px;">
      {{ errorMessage }}
    </div>
    <form class="row" @submit.prevent="addTodo">
      <input class="text" v-model="newTitle" placeholder="What needs to be done?" />
      <input class="text" v-model.number="userId" placeholder="user id (opsional)" style="max-width:160px" />
      <button class="primary" type="submit" :disabled="isSubmitting">{{ isSubmitting ? 'Adding...' : 'Add' }}</button>
    </form>
    <p class="muted" style="margin-top:8px">API: {{ apiBase }}/api/todos</p>

    <ul class="todo-list">
      <li v-for="t in todos" :key="t.id" class="todo-item">
        <label class="todo-left">
          <input type="checkbox" :checked="t.completed" @change="toggle(t)" :disabled="isSubmitting" />
          <span class="todo-title" :class="{ completed: t.completed }">{{ t.title }}</span>
        </label>
        <div class="todo-meta">
          <span v-if="t.user" class="muted">{{ t.user.name }} &lt;{{ t.user.email }}&gt;</span>
          <span v-else-if="t.user_id" class="muted">user #{{ t.user_id }}</span>
          <button class="ghost" @click="remove(t)" :disabled="isSubmitting">Delete</button>
        </div>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { onMounted, ref } from 'vue'

type User = { id:number; name:string; email:string }
type Todo = { id:number; title:string; completed:boolean; user_id?: number; user?: User }

const apiBase = (import.meta as any).env.VITE_API_BASE || 'http://localhost:8000'
const http = axios.create({ baseURL: apiBase })

const todos = ref<Todo[]>([])
const newTitle = ref('')
const userId = ref<number | undefined>()
const errorMessage = ref('')
const isSubmitting = ref(false)

async function load() {
  try {
    const { data } = await http.get('/api/todos', { params: { include_user: true } })
    todos.value = data
  } catch (e: any) {
    errorMessage.value = extractError(e)
  }
}

async function addTodo() {
  errorMessage.value = ''
  const title = newTitle.value.trim()
  if (!title) {
    errorMessage.value = 'Title wajib diisi.'
    return
  }
  if (userId.value !== undefined && (isNaN(userId.value as number) || (userId.value as number) <= 0)) {
    errorMessage.value = 'User ID harus angka positif.'
    return
  }
  isSubmitting.value = true
  try {
    await http.post('/api/todos', { title, user_id: userId.value })
    newTitle.value = ''
    userId.value = undefined
    await load()
  } catch (e: any) {
    errorMessage.value = extractError(e)
  } finally {
    isSubmitting.value = false
  }
}

async function toggle(t: Todo) {
  errorMessage.value = ''
  isSubmitting.value = true
  try {
    await http.put(`/api/todos/${t.id}`, { completed: !t.completed })
    await load()
  } catch (e: any) {
    errorMessage.value = extractError(e)
  } finally {
    isSubmitting.value = false
  }
}

async function remove(t: Todo) {
  errorMessage.value = ''
  isSubmitting.value = true
  try {
    await http.delete(`/api/todos/${t.id}`)
    await load()
  } catch (e: any) {
    errorMessage.value = extractError(e)
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

<style scoped>
.todo-list { list-style: none; padding: 0; margin-top: 12px; display: flex; flex-direction: column; gap: 8px; }
.todo-item { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 10px 12px; border: 1px solid var(--panel-border); border-radius: 10px; background: var(--panel); }
.todo-left { display: flex; align-items: center; gap: 10px; }
.todo-title { font-weight: 500; }
.todo-meta { display: flex; align-items: center; gap: 8px; }
</style>



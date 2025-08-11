<template>
  <div class="card">
    <h2>Todos</h2>
    <form class="row" @submit.prevent="addTodo">
      <input class="text" v-model="newTitle" placeholder="What needs to be done?" />
      <input class="text" v-model.number="userId" placeholder="user id (opsional)" style="max-width:160px" />
      <button class="primary" type="submit">Add</button>
    </form>
    <p class="muted" style="margin-top:8px">API: {{ apiBase }}/api/todos</p>

    <ul style="list-style:none; padding:0; margin-top:12px; display:flex; flex-direction:column; gap:8px;">
      <li v-for="t in todos" :key="t.id" class="row" style="align-items:center;">
        <input type="checkbox" :checked="t.completed" @change="toggle(t)" />
        <span style="flex:1">
          {{ t.title }}
          <span v-if="t.user" class="muted"> — {{ t.user.name }} &lt;{{ t.user.email }}&gt;</span>
          <span v-else-if="t.user_id" class="muted"> — user #{{ t.user_id }}</span>
        </span>
        <button class="ghost" @click="remove(t)">Delete</button>
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

async function load() {
  const { data } = await http.get('/api/todos', { params: { include_user: true } })
  todos.value = data
}

async function addTodo() {
  const title = newTitle.value.trim()
  if (!title) return
  await http.post('/api/todos', { title, user_id: userId.value })
  newTitle.value = ''
  userId.value = undefined
  await load()
}

async function toggle(t: Todo) {
  await http.put(`/api/todos/${t.id}`, { completed: !t.completed })
  await load()
}

async function remove(t: Todo) {
  await http.delete(`/api/todos/${t.id}`)
  await load()
}

onMounted(load)
</script>



<template>
  <div class="app">
    <header class="nav">
      <div class="brand">Microservices Todo</div>
      <div class="spacer"></div>
      <nav class="tabs">
        <button class="tab" :class="{ active: activeTab === 'todos' }" @click="activeTab = 'todos'">Todos</button>
        <button class="tab" :class="{ active: activeTab === 'users' }" @click="activeTab = 'users'">Users</button>
        <button class="tab" :class="{ active: activeTab === 'login' }" @click="activeTab = 'login'">Login</button>
      </nav>
      <button class="theme-toggle" @click="toggleTheme">{{ theme === 'dark' ? 'Light' : 'Dark' }} mode</button>
    </header>

    <main class="container">
      <TodoList v-if="activeTab === 'todos'" />
      <UsersList v-else-if="activeTab === 'users'" />
      <LoginForm v-else-if="activeTab === 'login'" />
    </main>
  </div>
  
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import TodoList from './components/TodoList.vue'
import UsersList from './components/UsersList.vue'
import LoginForm from './components/LoginForm.vue'

const activeTab = ref<'todos' | 'users' | 'login'>('todos')
const theme = ref<'dark' | 'light'>('dark')

function applyTheme(t: 'dark' | 'light') {
  document.documentElement.setAttribute('data-theme', t)
}

function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
  localStorage.setItem('theme', theme.value)
  applyTheme(theme.value)
}

onMounted(() => {
  const saved = localStorage.getItem('theme')
  if (saved === 'light' || saved === 'dark') {
    theme.value = saved as 'light' | 'dark'
  }
  applyTheme(theme.value)
})
</script>

<style>
:root {
  --bg: #f6f8ff;
  --panel: #ffffff;
  --panel-border: #d8e0ff;
  --text: #0b1020;
  --muted: #4a5a8a;
  --primary: #315efb;
  --primary-contrast: #ffffff;
  --ghost-border: #c7d2fe;
  --input-bg: #ffffff;
  --input-border: #cbd5e1;
  --danger-bg: #ffe5ea;
  --danger-text: #b4233a;
  --danger-border: #ffb3c1;
}

[data-theme='dark'] {
  --bg: #0b1020;
  --panel: #121a33;
  --panel-border: #1e2a52;
  --text: #e8eefc;
  --muted: #9db2d7;
  --primary: #3b82f6;
  --primary-contrast: #ffffff;
  --ghost-border: #2b3a6c;
  --input-bg: #0a132b;
  --input-border: #1e2a52;
  --danger-bg: #3b1d2a;
  --danger-text: #ffb4c0;
  --danger-border: #7a2a3e;
}

* { box-sizing: border-box; }
html, body, #app { height: 100%; }
body { margin: 0; background: var(--bg); color: var(--text); font-family: system-ui, Arial, sans-serif; }

.app { min-height: 100%; display: flex; flex-direction: column; }
.nav { display: flex; align-items: center; gap: 12px; padding: 12px 16px; border-bottom: 1px solid var(--panel-border); position: sticky; top: 0; background: var(--panel); z-index: 10; }
.brand { font-weight: 700; }
.spacer { flex: 1; }
.tabs { display: flex; gap: 8px; }
.tab { background: transparent; border: 1px solid var(--ghost-border); color: var(--text); padding: 6px 10px; border-radius: 8px; cursor: pointer; }
.tab.active { background: var(--primary); border-color: var(--primary); color: var(--primary-contrast); }
.theme-toggle { margin-left: 8px; background: transparent; border: 1px solid var(--ghost-border); color: var(--text); padding: 6px 10px; border-radius: 8px; cursor: pointer; }

.container { max-width: 980px; margin: 24px auto; padding: 0 16px; width: 100%; }

/* Shared UI tokens used by child components */
.card { background: var(--panel); border: 1px solid var(--panel-border); padding: 16px; border-radius: 12px; box-shadow: 0 6px 18px rgba(0,0,0,0.06) inset; }
.muted { color: var(--muted); }
.row { display: flex; gap: 8px; }
button.primary { background: var(--primary); border: 0; color: var(--primary-contrast); border-radius: 8px; padding: 8px 12px; cursor: pointer; }
button.ghost { background: transparent; border: 1px solid var(--ghost-border); color: var(--text); border-radius: 8px; padding: 8px 12px; cursor: pointer; }
input.text { background: var(--input-bg); color: var(--text); border: 1px solid var(--input-border); border-radius: 8px; padding: 8px 12px; outline: none; width: 100%; }

/* Todos styling enhancements */
.completed { text-decoration: line-through; color: var(--muted); }
</style>



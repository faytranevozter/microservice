<template>
  <main style="max-width: 920px; margin: 24px auto; font-family: system-ui, Arial, sans-serif;">
    <h1>Microservices Todo (Go + MySQL + Kong + Vue)</h1>
    <section style="display:flex; gap:24px; align-items:flex-start;">
      <TodoList />
      <UsersList />
    </section>
  </main>
  
</template>

<script setup lang="ts">
import TodoList from './components/TodoList.vue'
import UsersList from './components/UsersList.vue'
</script>

<style>
body { margin: 0; background: #0b1020; color: #e8eefc; }
input, button { font-size: 16px; }
section > * { flex: 1; }
.card { background:#121a33; border:1px solid #1e2a52; padding:16px; border-radius:12px; box-shadow:0 6px 18px rgba(0,0,0,0.25) inset; }
.muted { color:#9db2d7; }
.row { display:flex; gap:8px; }
button.primary { background:#3b82f6; border:0; color:white; border-radius:8px; padding:8px 12px; }
button.ghost { background:transparent; border:1px solid #2b3a6c; color:#e8eefc; border-radius:8px; padding:8px 12px; }
input.text { background:#0a132b; color:#e8eefc; border:1px solid #1e2a52; border-radius:8px; padding:8px 12px; outline:none; width:100%; }
label { font-size:14px; }
</style>



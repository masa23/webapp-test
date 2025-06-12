<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { useAuth } from '@/stores/auth'
import { ref } from 'vue'

const isLoggedIn = ref<boolean>(false)

const auth = useAuth()
function logout() {
  auth.logout()
}

auth.getToken().then(token => {
  if (token) {
    console.log('User is logged in')
    isLoggedIn.value = true
  } else {
    console.log('User is not logged in')
    isLoggedIn.value = false
  }
})
</script>

<template>
  <!-- Header -->
  <header class="bg-gray-100 shadow p-4 flex items-center justify-between max-w-[900px]">
    <h1 class="bg-gray-100 text-2xl font-bold text-gray-800">VM Manager</h1>
    <nav class="space-x-4">
      <RouterLink to="/" class="text-gray-600 hover:text-gray-900">Home</RouterLink>
      <RouterLink to="/api" class="text-gray-600 hover:text-gray-900">API</RouterLink>
      <RouterLink to="/login" class="ml-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">ログイン</RouterLink>
    </nav>
  </header>
  <main class="max-w-[900px]">
    <RouterView />
  </main>
</template>

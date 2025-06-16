<script setup lang="ts">
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { useAuth } from '@/stores/auth'
import { ref, watch } from 'vue'
import router from './router'

const isLoggedIn = ref<boolean>(false)

const auth = useAuth()
function logout() {
  auth.logout()
}

const checkLogin = async () => {
  const token = await auth.getToken()
  isLoggedIn.value = !!token
}

checkLogin()
watch(() => router.currentRoute.value.fullPath, () => {
  checkLogin() 
})
</script>

<template>
  <!-- Header -->
  <header class="bg-gray-100 shadow p-4 flex items-center justify-between max-w-[900px]">
    <h1 class="bg-gray-100 text-2xl font-bold text-gray-800">VM Manager</h1>
    <nav class="space-x-4">
      <RouterLink to="/" class="text-gray-600 hover:text-gray-900">Home</RouterLink>
      <!--<RouterLink to="/apikey" class="text-gray-600 hover:text-gray-900">API</RouterLink>-->
      <RouterLink v-if="isLoggedIn===false" to="/login" class="ml-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">ログイン</RouterLink>
      <button v-if="isLoggedIn===true" @click="logout" class="ml-4 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700">ログアウト</button>
    </nav>
  </header>
  <main class="max-w-[900px]">
    <RouterView />
  </main>
</template>

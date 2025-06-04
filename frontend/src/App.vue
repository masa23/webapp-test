<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { useAuth } from '@/stores/auth'

const auth = useAuth()
function logout() {
  auth.logout()
}
</script>

<template>
  <div class="flex min-h-screen bg-gray-100 text-gray-800">
    <div class="w-full max-w-[900px] flex bg-white shadow-lg rounded-lg overflow-hidden mt-8">
      <aside class="w-64 bg-white shadow flex-shrink-0">
        <div class="p-6 text-2xl font-bold border-b">VM Manager</div>
        <nav class="p-4 flex flex-col gap-3">
          <RouterLink to="/" class="hover:text-blue-600">Dashboard</RouterLink>
          <RouterLink v-if="!auth.isAuthenticated" to="/login" class="text-blue-600 hover:underline">
            Login
          </RouterLink>
          <a v-if="auth.isAuthenticated" @click.prevent="logout" class="text-left text-red-600 hover:underline">
            Logout
          </a>
        </nav>
      </aside>

      <div class="flex-1 flex flex-col">
        <main class="flex-1 p-6">
          <RouterView />
        </main>
      </div>
    </div>
  </div>
</template>

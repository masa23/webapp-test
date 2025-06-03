<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { useAuth } from '@/stores/auth'
import router from './router'

const auth = useAuth()
function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <header>
    <div class="wrapper">
      <nav>
        <RouterLink to="/">Home</RouterLink>
        <RouterLink to="/about">About</RouterLink>
        <a v-if="auth.isAuthenticated" v-on:click="logout" href="#">Logout</a>
        <RouterLink v-if="!auth.isAuthenticated" to="/login">Login</RouterLink>
      </nav>
    </div>
  </header>

  <RouterView />
</template>

<style scoped>
</style>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAuth } from '@/stores/auth';

const router = useRouter();
const username = ref(<string>'');
const password = ref(<string>'');
const errorMessage = ref('');
const successMessage = ref('');

const auth = useAuth();

const login = async () => {
  try {
    await auth.login(username.value, password.value);
    successMessage.value = 'Login successful!';
    errorMessage.value = '';
    router.push('/'); // Redirect to home after successful login
  } catch (error) {
    errorMessage.value = 'Login failed. Please check your credentials.' + (error instanceof Error ? error.message : '');
    successMessage.value = '';
  }
};
</script>

<template>
  <main>
    <!-- Login Form -->
    <div class="login-container">
      <h1>Login</h1>
      <form @submit.prevent>
        <div class="form-group">
          <label for="username">Username</label>
          <input type="text" id="username" v-model="username" required />
        </div>
        <div class="form-group">
          <label for="password">Password</label>
          <input type="password" id="password" v-model="password" required />
        </div>
        <button type="submit" @click="login">Login</button>
      </form>
      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
      <p v-if="successMessage" class="success">{{ successMessage }}</p>
    </div>
  </main>
</template>

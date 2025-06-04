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
  <main class="min-h-screen flex justify-center bg-gray-100 pt-24">
    <div class="bg-white shadow-lg rounded-xl p-8 w-[300px]">
      <h1 class="text-2xl font-bold text-center mb-6 text-gray-800">Login</h1>

      <form @submit.prevent>
        <div class="mb-4">
          <label for="username" class="block text-sm font-medium text-gray-700">Username</label>
          <input type="text" id="username" v-model="username" required
            class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500" />
        </div>

        <div class="mb-4">
          <label for="password" class="block text-sm font-medium text-gray-700">Password</label>
          <input type="password" id="password" v-model="password" required
            class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500" />
        </div>

        <button type="submit" @click="login"
          class="w-full bg-blue-600 text-white py-2 rounded-md hover:bg-blue-700 transition">
          Login
        </button>
      </form>

      <p v-if="errorMessage" class="mt-4 text-sm text-red-600">{{ errorMessage }}</p>
      <p v-if="successMessage" class="mt-4 text-sm text-green-600">{{ successMessage }}</p>
    </div>
  </main>
</template>

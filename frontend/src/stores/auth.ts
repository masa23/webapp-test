import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import axios from 'axios'
import router from '@/router'

export const useAuth = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const username = ref<string | null>(null)

  const isAuthenticated = computed(() => !!token.value)

  const login = async (usernameInput: string, passwordInput: string) => {
    try {
      const res = await axios.post('/login', {
        username: usernameInput,
        password: passwordInput,
      })
      token.value = res.data.token
      if (!token.value) {
        throw new Error('トークンが取得できませんでした')
      }
      localStorage.setItem('token', token.value)
      await fetchProfile()
    } catch (err) {
      throw new Error('ログイン失敗')
    }
  }

  const fetchProfile = async () => {
    if (!token.value) return
    try {
      const res = await axios.get('/api/profile', {
        headers: {
          Authorization: `Bearer ${token.value}`,
        },
      })
      username.value = res.data.username
    } catch {
      logout()
    }
  }

  const logout = () => {
    token.value = null
    username.value = null
    localStorage.removeItem('token')
  }

  return {
    token,
    username,
    isAuthenticated,
    login,
    logout,
    fetchProfile,
  }
})

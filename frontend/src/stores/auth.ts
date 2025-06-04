import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { useRouter } from 'vue-router'
import axios from 'axios'


export const useAuth = defineStore('auth', () => {
  const router = useRouter()
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
      router.push('/')
    } catch (err) {
      const error = err as any
      throw new Error(error.response?.data?.message || 'ログイン失敗')
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
    router.push('/login')
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

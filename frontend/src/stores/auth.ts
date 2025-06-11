import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { useRouter } from 'vue-router'
import axios from 'axios'

type AccessToken = {
  access_token: string
  expires_at: number
}

export const useAuth = defineStore('auth', () => {
  const router = useRouter()
  const accessToken = ref<AccessToken | null>(null)
  const username = ref<string | null>(null)

  const login = async (usernameInput: string, passwordInput: string) => {
    try {
      const res = await axios.post('/auth/login', {
        username: usernameInput,
        password: passwordInput,
      })
      await fetchAccessToken()
      await fetchProfile()
      router.push('/')
    } catch (error) {
      throw new Error('Login failed')
    }
  }

  const fetchAccessToken = async () => {
    if (
      !accessToken.value ||
      !accessToken.value.access_token ||
      new Date().getTime() > (accessToken.value.expires_at*1000 - 10 * 1000)
    ) {
      try {
        const res = await axios.get('/auth/refresh')
        if (res.data && res.data.access_token) {
          accessToken.value = {
            access_token: res.data.access_token,
            expires_at: res.data.expires_at,
          }
        }
      } catch (error) {
        console.error('Failed to fetch access token', error)
        logout()
      }
    }
  }

  const fetchProfile = async () => {
    const t = await getToken()
    try {
      const res = await axios.get('/api/profile', {
        headers: {
          Authorization: `Bearer ${t}`,
        },
      })
      if (res.data && res.data.username) {
        username.value = res.data.username
      }
    } catch (error) {
      console.error('Failed to fetch profile', error)
      logout()
    }
  }

  const getToken = async () => {
    await fetchAccessToken()
    return accessToken.value?.access_token || null
  }

  const logout = () => {
    try {
      axios.post('/auth/logout')
    } catch (error) {
      console.error('Logout failed', error)
    }
    accessToken.value = null
    router.push('/login')
  }

  return {
    username,
    getToken,
    login,
    logout,
    fetchAccessToken,
  }
})

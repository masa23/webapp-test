import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import HomeView from '@/views/HomeView.vue'
import { useAuth } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
      meta: { requiresAuth: true },
    },
    {
      path: '/login',
      name: 'login',
      component: LoginView,
    },
  ],
})


// 認証チェック
router.beforeEach((to, _from, next) => {
  const auth = useAuth()
  if (to.meta.requiresAuth && auth.token === null) { 
    next('/login')
  } else {
    next()
  }
})

export default router

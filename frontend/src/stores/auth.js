import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as apiLogin, register as apiRegister } from '@/api/auth.js'

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null)
  const isLoggedIn = computed(() => !!user.value)

  function init() {
    const stored = localStorage.getItem('user')
    if (stored) user.value = JSON.parse(stored)
  }

  async function login(email, password) {
    const { data } = await apiLogin({ email, password })
    user.value = data.user
    localStorage.setItem('user', JSON.stringify(data.user))
    localStorage.setItem('access_token', data.tokens.access_token)
    localStorage.setItem('refresh_token', data.tokens.refresh_token)
  }

  async function register(payload) {
    const { data } = await apiRegister(payload)
    user.value = data.user
    localStorage.setItem('user', JSON.stringify(data.user))
    localStorage.setItem('access_token', data.tokens.access_token)
    localStorage.setItem('refresh_token', data.tokens.refresh_token)
  }

  function logout() {
    user.value = null
    localStorage.removeItem('user')
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  }

  return { user, isLoggedIn, init, login, register, logout }
})

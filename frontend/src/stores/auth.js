import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as apiLogin } from '@/api/auth.js'

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null)
  const isLoggedIn = computed(() => !!user.value)

  function init() {
    const stored = localStorage.getItem('user')
    if (stored) user.value = JSON.parse(stored)
  }

  // setSession is shared by both login and the email-verified sign-up flow,
  // since /register/verify also returns a {user, tokens} pair.
  function setSession(payload) {
    user.value = payload.user
    localStorage.setItem('user', JSON.stringify(payload.user))
    localStorage.setItem('access_token', payload.tokens.access_token)
    localStorage.setItem('refresh_token', payload.tokens.refresh_token)
  }

  async function login(email, password) {
    const { data } = await apiLogin({ email, password })
    setSession(data)
  }

  function logout() {
    user.value = null
    localStorage.removeItem('user')
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  }

  return { user, isLoggedIn, init, login, setSession, logout }
})

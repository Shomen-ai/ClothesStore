import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as apiLogin } from '@/api/auth.js'

export const useAuthStore = defineStore('auth', () => {
  // Authenticated user object (null when logged out); tokens live in localStorage.
  const user = ref(null)
  // Derived auth flag: true whenever a user is loaded into state.
  const isLoggedIn = computed(() => !!user.value)

  // Rehydrate the user from localStorage on app start so a refresh keeps the session.
  function init() {
    const stored = localStorage.getItem('user')
    if (stored) user.value = JSON.parse(stored)
  }

  // setSession is shared by both login and the email-verified sign-up flow,
  // since /register/verify also returns a {user, tokens} pair.
  function setSession(payload) {
    // Store the user in reactive state and persist user + token pair to localStorage.
    user.value = payload.user
    localStorage.setItem('user', JSON.stringify(payload.user))
    localStorage.setItem('access_token', payload.tokens.access_token)
    localStorage.setItem('refresh_token', payload.tokens.refresh_token)
  }

  // Authenticate via the API and open a session from the returned {user, tokens}.
  async function login(email, password) {
    const { data } = await apiLogin({ email, password })
    setSession(data)
  }

  // Clear reactive state and wipe all persisted session data from localStorage.
  function logout() {
    user.value = null
    localStorage.removeItem('user')
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  }

  return { user, isLoggedIn, init, login, setSession, logout }
})

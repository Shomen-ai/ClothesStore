// Shared axios instance for the whole app. All API modules import this so that
// auth-header injection and token-refresh handling are applied uniformly.
import axios from 'axios'

// All requests are prefixed with /api (proxied to the backend in dev/prod).
const api = axios.create({ baseURL: '/api' })

// Request interceptor: attach the current access token as a Bearer header
// so authenticated endpoints receive credentials automatically.
api.interceptors.request.use(config => {
  const token = localStorage.getItem('access_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// Response interceptor: transparently handle expired access tokens.
api.interceptors.response.use(
  res => res,
  async err => {
    // Only act on 401s, and only retry a given request once (_retry guard
    // prevents an infinite refresh/retry loop when refresh also returns 401).
    if (err.response?.status === 401 && !err.config._retry) {
      err.config._retry = true
      const refresh = localStorage.getItem('refresh_token')
      if (refresh) {
        try {
          // Exchange the refresh token for a fresh token pair. Uses bare axios
          // (not `api`) to bypass this interceptor and avoid recursion.
          const { data } = await axios.post('/api/auth/refresh', { refresh_token: refresh })
          // Persist the rotated tokens for subsequent requests.
          localStorage.setItem('access_token', data.access_token)
          localStorage.setItem('refresh_token', data.refresh_token)
          // Re-authorize and replay the original failed request.
          err.config.headers.Authorization = `Bearer ${data.access_token}`
          return api(err.config)
        } catch {
          // Refresh failed (expired/invalid) -> clear session and force re-login.
          localStorage.removeItem('access_token')
          localStorage.removeItem('refresh_token')
          window.location.href = '/login'
        }
      }
    }
    // Non-401 errors, or 401s without a refresh token, propagate to the caller.
    return Promise.reject(err)
  }
)

export default api

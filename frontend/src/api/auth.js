import api from './axios.js'
export const register = (data) => api.post('/auth/register', data)
export const login = (data) => api.post('/auth/login', data)
export const refresh = (token) => api.post('/auth/refresh', { refresh_token: token })

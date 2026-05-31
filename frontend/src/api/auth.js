import api from './axios.js'

export const login           = (data)  => api.post('/auth/login', data)
export const refresh         = (token) => api.post('/auth/refresh', { refresh_token: token })
export const registerStart   = (data)  => api.post('/auth/register/start',  data)
export const registerVerify  = (data)  => api.post('/auth/register/verify', data)
export const registerResend  = (email) => api.post('/auth/register/resend', { email })

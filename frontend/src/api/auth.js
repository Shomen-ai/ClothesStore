// Authentication endpoints: login, token refresh, and the email-verified
// two-step registration flow (start -> verify, with resend support).
import api from './axios.js'

// POST /auth/login — authenticate with credentials; returns {user, tokens}.
export const login           = (data)  => api.post('/auth/login', data)
// POST /auth/refresh — exchange a refresh token for a new token pair.
export const refresh         = (token) => api.post('/auth/refresh', { refresh_token: token })
// POST /auth/register/start — begin sign-up and trigger the email verification code.
export const registerStart   = (data)  => api.post('/auth/register/start',  data)
// POST /auth/register/verify — confirm the emailed code and finalize the account.
export const registerVerify  = (data)  => api.post('/auth/register/verify', data)
// POST /auth/register/resend — re-send the verification code to the email.
export const registerResend  = (email) => api.post('/auth/register/resend', { email })

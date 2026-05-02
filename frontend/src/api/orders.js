import api from './axios.js'
export const createOrder = (data) => api.post('/orders', data)
export const validatePromo = (code, total) => api.post('/promo/validate', { code, total })

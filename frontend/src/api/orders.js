import api from './axios.js'
export const createOrder     = (data) => api.post('/orders', data)
export const validatePromo   = (code, total) => api.post('/promo/validate', { code, total })
export const getOrder        = (id) => api.get(`/user/orders/${id}`)
export const confirmPayment  = (id) => api.post(`/orders/${id}/confirm-payment`)

// Checkout/order endpoints: place orders, validate promo codes, confirm payment.
import api from './axios.js'
// POST /orders — create a new order from the checkout payload.
export const createOrder     = (data) => api.post('/orders', data)
// POST /promo/validate — check a promo code against the cart total; returns discount.
export const validatePromo   = (code, total) => api.post('/promo/validate', { code, total })
// GET /user/orders/:id — fetch one of the current user's orders.
export const getOrder        = (id) => api.get(`/user/orders/${id}`)
// POST /orders/:id/confirm-payment — mark an order's payment as confirmed.
export const confirmPayment  = (id) => api.post(`/orders/${id}/confirm-payment`)

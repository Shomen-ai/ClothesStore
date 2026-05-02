import api from './axios.js'
// Products
export const adminListProducts = (p) => api.get('/admin/products', { params: p })
export const adminCreateProduct = (d) => api.post('/admin/products', d)
export const adminUpdateProduct = (id, d) => api.put(`/admin/products/${id}`, d)
export const adminDeleteProduct = (id) => api.delete(`/admin/products/${id}`)
export const adminUploadImages = (id, form) => api.post(`/admin/products/${id}/images`, form, { headers: { 'Content-Type': 'multipart/form-data' } })
export const adminDeleteImage = (pid, imgId) => api.delete(`/admin/products/${pid}/images/${imgId}`)
// Categories
export const adminGetCategories = () => api.get('/admin/categories')
export const adminCreateCategory = (d) => api.post('/admin/categories', d)
export const adminUpdateCategory = (id, d) => api.put(`/admin/categories/${id}`, d)
export const adminDeleteCategory = (id) => api.delete(`/admin/categories/${id}`)
// Orders
export const adminListOrders = (p) => api.get('/admin/orders', { params: p })
export const adminGetOrder = (id) => api.get(`/admin/orders/${id}`)
export const adminUpdateStatus = (id, status) => api.put(`/admin/orders/${id}/status`, { status })
// Promo
export const adminListPromos = () => api.get('/admin/promo-codes')
export const adminCreatePromo = (d) => api.post('/admin/promo-codes', d)
export const adminDeactivatePromo = (id) => api.put(`/admin/promo-codes/${id}/deactivate`)
export const adminDeletePromo = (id) => api.delete(`/admin/promo-codes/${id}`)
// Stats
export const getRevenueStats = (period) => api.get('/admin/stats/revenue', { params: { period } })
export const getOrderStats = (period) => api.get('/admin/stats/orders', { params: { period } })
export const getPromoStats = (period) => api.get('/admin/stats/promo-codes', { params: { period } })

// Admin-only API endpoints (catalogue CRUD, orders, promo codes, stats, reports,
// review moderation). All routes require an authenticated admin session.
import api from './axios.js'
// Products
// GET /admin/products — paginated/filtered product list (p = query params).
export const adminListProducts = (p) => api.get('/admin/products', { params: p })
// POST /admin/products — create a product (d = product payload).
export const adminCreateProduct = (d) => api.post('/admin/products', d)
// PUT /admin/products/:id — update an existing product.
export const adminUpdateProduct = (id, d) => api.put(`/admin/products/${id}`, d)
// DELETE /admin/products/:id — remove a product.
export const adminDeleteProduct = (id) => api.delete(`/admin/products/${id}`)
// POST /admin/products/:id/images — upload product images (multipart form-data).
export const adminUploadImages = (id, form) => api.post(`/admin/products/${id}/images`, form, { headers: { 'Content-Type': 'multipart/form-data' } })
// DELETE /admin/products/:pid/images/:imgId — delete a single product image.
export const adminDeleteImage = (pid, imgId) => api.delete(`/admin/products/${pid}/images/${imgId}`)
// Categories
// GET /admin/categories — list all categories.
export const adminGetCategories = () => api.get('/admin/categories')
// POST /admin/categories — create a category.
export const adminCreateCategory = (d) => api.post('/admin/categories', d)
// PUT /admin/categories/:id — update a category.
export const adminUpdateCategory = (id, d) => api.put(`/admin/categories/${id}`, d)
// DELETE /admin/categories/:id — delete a category.
export const adminDeleteCategory = (id) => api.delete(`/admin/categories/${id}`)
// Orders
// GET /admin/orders — paginated/filtered order list (p = query params).
export const adminListOrders = (p) => api.get('/admin/orders', { params: p })
// GET /admin/orders/:id — fetch one order with full detail.
export const adminGetOrder = (id) => api.get(`/admin/orders/${id}`)
// PUT /admin/orders/:id/status — change an order's fulfillment status.
export const adminUpdateStatus = (id, status) => api.put(`/admin/orders/${id}/status`, { status })
// Promo
// GET /admin/promo-codes — list all promo codes.
export const adminListPromos = () => api.get('/admin/promo-codes')
// POST /admin/promo-codes — create a promo code.
export const adminCreatePromo = (d) => api.post('/admin/promo-codes', d)
// PUT /admin/promo-codes/:id/deactivate — disable a promo code without deleting it.
export const adminDeactivatePromo = (id) => api.put(`/admin/promo-codes/${id}/deactivate`)
// DELETE /admin/promo-codes/:id — delete a promo code.
export const adminDeletePromo = (id) => api.delete(`/admin/promo-codes/${id}`)
// Stats
// GET /admin/stats/revenue — revenue figures for the given period.
export const getRevenueStats = (period) => api.get('/admin/stats/revenue', { params: { period } })
// GET /admin/stats/orders — order counts/metrics for the given period.
export const getOrderStats = (period) => api.get('/admin/stats/orders', { params: { period } })
// GET /admin/stats/promo-codes — promo-code usage stats for the given period.
export const getPromoStats = (period) => api.get('/admin/stats/promo-codes', { params: { period } })
// Reports
// GET /admin/reports/excel — download the Excel report as a binary blob.
export const downloadExcelReport = () => api.get('/admin/reports/excel', { responseType: 'blob' })
// Reviews
// GET /admin/reviews — list reviews; optional `hidden` flag filters by visibility (omitted when null).
export const adminListReviews    = (hidden) => api.get('/admin/reviews', { params: hidden == null ? {} : { hidden } })
// PATCH /admin/reviews/:id — toggle a review's hidden (moderation) state.
export const adminSetReviewHidden = (id, isHidden) => api.patch(`/admin/reviews/${id}`, { is_hidden: isHidden })
// DELETE /admin/reviews/:id — permanently delete a review.
export const adminDeleteReview   = (id) => api.delete(`/admin/reviews/${id}`)

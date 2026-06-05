// Product review endpoints for the storefront (read list, plus the current
// user's own review CRUD).
import api from './axios.js'

// GET /products/:productId/reviews — list visible reviews for a product.
export const getProductReviews = (productId) => api.get(`/products/${productId}/reviews`)
// GET /products/:productId/reviews/me — the current user's review state/eligibility for a product.
export const getMyReviewState  = (productId) => api.get(`/products/${productId}/reviews/me`)
// POST /products/:productId/reviews — create the user's review for a product.
export const createReview      = (productId, body) => api.post(`/products/${productId}/reviews`, body)
// PUT /products/:productId/reviews/:reviewId — edit the user's existing review.
export const updateReview      = (productId, reviewId, body) => api.put(`/products/${productId}/reviews/${reviewId}`, body)
// DELETE /products/:productId/reviews/:reviewId — delete the user's review.
export const deleteReview      = (productId, reviewId) => api.delete(`/products/${productId}/reviews/${reviewId}`)

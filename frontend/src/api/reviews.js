import api from './axios.js'

export const getProductReviews = (productId) => api.get(`/products/${productId}/reviews`)
export const getMyReviewState  = (productId) => api.get(`/products/${productId}/reviews/me`)
export const createReview      = (productId, body) => api.post(`/products/${productId}/reviews`, body)
export const updateReview      = (productId, reviewId, body) => api.put(`/products/${productId}/reviews/${reviewId}`, body)
export const deleteReview      = (productId, reviewId) => api.delete(`/products/${productId}/reviews/${reviewId}`)

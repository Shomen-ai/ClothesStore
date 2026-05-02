import api from './axios.js'
export const getCategories = () => api.get('/categories')
export const getProducts = (params) => api.get('/products', { params })
export const getProduct = (id) => api.get(`/products/${id}`)
export const getFeatured = () => api.get('/products/featured')

// Public storefront catalogue endpoints (categories and products).
import api from './axios.js'
// GET /categories — list all storefront categories.
export const getCategories = () => api.get('/categories')
// GET /products — list products with optional filter/sort/pagination params.
export const getProducts = (params) => api.get('/products', { params })
// GET /products/:id — fetch a single product's detail.
export const getProduct = (id) => api.get(`/products/${id}`)
// GET /products/featured — featured products for the homepage.
export const getFeatured = () => api.get('/products/featured')

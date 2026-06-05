// Authenticated user account endpoints: profile, addresses, wishlist, orders.
import api from './axios.js'
// GET /user/profile — fetch the current user's profile.
export const getProfile = () => api.get('/user/profile')
// PUT /user/profile — update the current user's profile.
export const updateProfile = (data) => api.put('/user/profile', data)
// GET /user/addresses — list the user's saved delivery addresses.
export const getAddresses = () => api.get('/user/addresses')
// POST /user/addresses — add a new delivery address.
export const createAddress = (data) => api.post('/user/addresses', data)
// PUT /user/addresses/:id — update a saved address.
export const updateAddress = (id, data) => api.put(`/user/addresses/${id}`, data)
// DELETE /user/addresses/:id — remove a saved address.
export const deleteAddress = (id) => api.delete(`/user/addresses/${id}`)
// GET /user/wishlist — fetch the user's wishlist products.
export const getWishlist = () => api.get('/user/wishlist')
// POST /user/wishlist/:id — add a product to the wishlist.
export const addToWishlist = (id) => api.post(`/user/wishlist/${id}`)
// DELETE /user/wishlist/:id — remove a product from the wishlist.
export const removeFromWishlist = (id) => api.delete(`/user/wishlist/${id}`)
// GET /user/orders — list the user's order history.
export const getUserOrders = () => api.get('/user/orders')
// GET /user/orders/:id — fetch one of the user's orders.
export const getUserOrder = (id) => api.get(`/user/orders/${id}`)

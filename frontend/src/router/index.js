import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth.js'

const routes = [
  { path: '/', component: () => import('@/views/HomeView.vue') },
  { path: '/catalogue', component: () => import('@/views/CatalogueView.vue') },
  { path: '/catalogue/:id', component: () => import('@/views/ProductView.vue') },
  { path: '/cart', component: () => import('@/views/CartView.vue') },
  { path: '/checkout-success', component: () => import('@/views/CheckoutSuccessView.vue') },
  { path: '/login', component: () => import('@/views/LoginView.vue') },
  { path: '/register', component: () => import('@/views/RegisterView.vue') },
  {
    path: '/account',
    component: () => import('@/views/account/AccountLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      { path: '', component: () => import('@/views/account/ProfileView.vue') },
      { path: 'orders', component: () => import('@/views/account/OrdersView.vue') },
      { path: 'orders/:id', component: () => import('@/views/account/OrderDetailView.vue') },
      { path: 'addresses', component: () => import('@/views/account/AddressesView.vue') },
      { path: 'wishlist', component: () => import('@/views/account/WishlistView.vue') },
    ]
  },
  {
    path: '/admin',
    component: () => import('@/views/admin/AdminLayout.vue'),
    meta: { requiresAdmin: true },
    children: [
      { path: '', component: () => import('@/views/admin/DashboardView.vue') },
      { path: 'products', component: () => import('@/views/admin/AdminProductsView.vue') },
      { path: 'orders', component: () => import('@/views/admin/AdminOrdersView.vue') },
      { path: 'promo-codes', component: () => import('@/views/admin/AdminPromoView.vue') },
    ]
  }
]

const router = createRouter({ history: createWebHistory(), routes })

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isLoggedIn) return '/login'
  if (to.meta.requiresAdmin && auth.user?.role !== 'admin') return '/'
})

export default router

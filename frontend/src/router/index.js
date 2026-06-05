// Application router: defines all route groups, lazy-loads every view via dynamic
// import() (so each page becomes its own code-split chunk), and enforces auth/admin
// gating through a single global beforeEach guard.
import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth.js'

const routes = [
  // --- Public storefront (no auth required) ---
  { path: '/', redirect: '/catalogue' },
  { path: '/catalogue', component: () => import('@/views/CatalogueView.vue') },
  { path: '/catalogue/:id', component: () => import('@/views/ProductView.vue') },   // product detail by id
  { path: '/cart', component: () => import('@/views/CartView.vue') },
  { path: '/checkout-success', component: () => import('@/views/CheckoutSuccessView.vue') },
  // Payment page is gated: only a logged-in user can pay for an order.
  { path: '/payment/:id', component: () => import('@/views/PaymentView.vue'), meta: { requiresAuth: true } },

  // --- Auth pages ---
  { path: '/login', component: () => import('@/views/LoginView.vue') },
  { path: '/register', component: () => import('@/views/RegisterView.vue') },

  // --- Customer account area (auth-gated parent; children inherit the guard) ---
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
  // --- Admin back office (admin-role-gated parent; children inherit the guard) ---
  {
    path: '/admin',
    component: () => import('@/views/admin/AdminLayout.vue'),
    meta: { requiresAdmin: true },
    children: [
      { path: '', component: () => import('@/views/admin/DashboardView.vue') },
      { path: 'products', component: () => import('@/views/admin/AdminProductsView.vue') },
      { path: 'orders', component: () => import('@/views/admin/AdminOrdersView.vue') },
      { path: 'promo-codes', component: () => import('@/views/admin/AdminPromoView.vue') },
      { path: 'reviews', component: () => import('@/views/admin/ReviewsView.vue') },
    ]
  }
]

// HTML5 history mode (clean URLs, no hash). Requires a server fallback to index.html.
const router = createRouter({ history: createWebHistory(), routes })

// Global navigation guard: client-side gating for protected routes.
// requiresAuth  -> must be logged in, otherwise redirect to /login.
// requiresAdmin -> must have the 'admin' role, otherwise redirect home.
// NOTE: this only controls which views render. It is purely cosmetic security —
// the role comes from localStorage-restored state and is not re-verified here, so
// the backend API must enforce authorization on every protected/admin endpoint.
router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isLoggedIn) return '/login'
  if (to.meta.requiresAdmin && auth.user?.role !== 'admin') return '/'
})

export default router

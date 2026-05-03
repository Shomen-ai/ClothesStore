<template>
  <header class="app-header">
    <div class="header-inner">
      <RouterLink to="/" class="logo">STORE</RouterLink>

      <nav class="nav-links">
        <RouterLink to="/catalogue">Каталог</RouterLink>
      </nav>

      <div class="header-actions">
        <RouterLink to="/cart" class="cart-btn">
          <el-icon><ShoppingCart /></el-icon>
          <span v-if="cart.count > 0" class="cart-badge">{{ cart.count }}</span>
        </RouterLink>

        <template v-if="auth.isLoggedIn">
          <RouterLink to="/account" class="icon-btn">
            <el-icon><User /></el-icon>
          </RouterLink>
          <RouterLink v-if="auth.user?.role === 'admin'" to="/admin" class="icon-btn admin-link">
            <el-icon><Setting /></el-icon>
          </RouterLink>
          <button class="icon-btn" @click="handleLogout">
            <el-icon><SwitchButton /></el-icon>
          </button>
        </template>
        <template v-else>
          <RouterLink to="/login" class="nav-link">Войти</RouterLink>
        </template>
      </div>
    </div>
  </header>
</template>

<script setup>
import { useAuthStore } from '@/stores/auth.js'
import { useCartStore } from '@/stores/cart.js'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const cart = useCartStore()
const router = useRouter()

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.app-header {
  position: fixed; top: 0; left: 0; right: 0; z-index: 100;
  background: rgba(10,10,10,0.95);
  border-bottom: 1px solid var(--color-border);
  backdrop-filter: blur(8px);
}
.header-inner {
  max-width: 1400px; margin: 0 auto; padding: 0 24px;
  height: 64px; display: flex; align-items: center; gap: 32px;
}
@media (max-width: 768px) {
  .header-inner { padding: 0 16px; gap: 16px; }
  .nav-links { gap: 12px; }
  .nav-links a { font-size: 12px; }
  .header-actions { gap: 10px; }
}
.logo {
  font-size: 20px; font-weight: 800; letter-spacing: 4px;
  color: var(--color-text); text-decoration: none;
}
.logo:hover { color: var(--color-red); }
.nav-links { display: flex; gap: 24px; flex: 1; }
.nav-links a { color: var(--color-text-muted); font-size: 13px; letter-spacing: 1px; text-transform: uppercase; }
.nav-links a:hover, .nav-links a.router-link-active { color: var(--color-text); }
.header-actions { display: flex; align-items: center; gap: 16px; }
.cart-btn { position: relative; color: var(--color-text); display: flex; font-size: 26px; }
.cart-badge {
  position: absolute; top: -6px; right: -6px;
  background: var(--color-red); color: white;
  border-radius: 50%; width: 16px; height: 16px;
  font-size: 10px; display: flex; align-items: center; justify-content: center;
}
.icon-btn { background: none; border: none; cursor: pointer; color: var(--color-text-muted); display: flex; font-size: 26px; }
.icon-btn:hover { color: var(--color-text); }
.admin-link { color: var(--color-red); }
.nav-link { color: var(--color-text-muted); font-size: 13px; }
</style>

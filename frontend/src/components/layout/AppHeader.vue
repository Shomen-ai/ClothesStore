<!--
  AppHeader: fixed top navigation bar. Holds the brand link, catalogue/sale nav,
  theme toggle, cart button (with item-count badge), and an auth-aware action cluster
  (account/admin/logout when signed in, otherwise a "login" link). Adds a `scrolled`
  class once the page is scrolled to switch to its elevated/bordered style.
-->
<template>
  <header class="app-header" :class="{ scrolled }">
    <div class="header-inner">
      <RouterLink to="/" class="brand">
        <span class="brand-mark">◆</span>
        <span class="brand-name">STORE</span>
        <span class="brand-tag">/ носи <span class="gothic-accent">громко</span></span>
      </RouterLink>

      <nav class="nav-links">
        <RouterLink to="/catalogue" class="nav-link">Каталог</RouterLink>
        <RouterLink to="/catalogue?sale=1" class="nav-link nav-sale">
          <span class="sale-dot"></span>Sale
        </RouterLink>
      </nav>

      <div class="header-actions">
        <button class="action theme-toggle" :class="{ 'is-dark': theme.theme === 'dark' }"
          @click="theme.toggle()"
          :aria-label="theme.theme === 'dark' ? 'Переключить на светлую тему' : 'Переключить на тёмную тему'">
          <svg class="theme-icon sun" width="20" height="20" viewBox="0 0 20 20" fill="none">
            <circle cx="10" cy="10" r="3.6" stroke="currentColor" stroke-width="1.4"/>
            <path d="M10 1.5v2.4M10 16.1v2.4M3.5 3.5l1.7 1.7M14.8 14.8l1.7 1.7M1.5 10h2.4M16.1 10h2.4M3.5 16.5l1.7-1.7M14.8 5.2l1.7-1.7"
              stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/>
          </svg>
          <svg class="theme-icon moon" width="20" height="20" viewBox="0 0 20 20" fill="none">
            <path d="M16 11.5A6.5 6.5 0 0 1 8.5 4a1 1 0 0 0-1.4-0.95A7.5 7.5 0 1 0 16.95 12.9 1 1 0 0 0 16 11.5z"
              stroke="currentColor" stroke-width="1.4" stroke-linejoin="round"/>
          </svg>
        </button>

        <RouterLink to="/cart" class="action cart-action" aria-label="Корзина">
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
            <path d="M3 5h2l1.5 9.5a1 1 0 0 0 1 .85h7a1 1 0 0 0 1-.85L17 7H6" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/>
            <circle cx="8" cy="17" r="1.2" fill="currentColor"/>
            <circle cx="15" cy="17" r="1.2" fill="currentColor"/>
          </svg>
          <!-- Item-count badge, hidden when the cart is empty -->
          <span v-if="cart.count > 0" class="cart-badge">{{ cart.count }}</span>
        </RouterLink>

        <!-- Signed-in actions: account, admin (admins only), logout -->
        <template v-if="auth.isLoggedIn">
          <RouterLink to="/account" class="action" aria-label="Аккаунт">
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
              <circle cx="10" cy="7" r="3.2" stroke="currentColor" stroke-width="1.4"/>
              <path d="M3.5 17c1-3 3.5-4.5 6.5-4.5s5.5 1.5 6.5 4.5" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/>
            </svg>
          </RouterLink>
          <!-- Admin entry shown only to admin-role users (UI gate only; backend must still authorize) -->
          <RouterLink v-if="auth.user?.role === 'admin'" to="/admin" class="action admin-action" aria-label="Админ">
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
              <path d="M10 1.5L17 5v6c0 4-3 6.5-7 8-4-1.5-7-4-7-8V5l7-3.5z" stroke="currentColor" stroke-width="1.4" stroke-linejoin="round"/>
            </svg>
          </RouterLink>
          <button class="action" @click="handleLogout" aria-label="Выход">
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
              <path d="M12 4h3a1 1 0 0 1 1 1v10a1 1 0 0 1-1 1h-3M8 7l-3 3 3 3M5 10h9" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </button>
        </template>
        <!-- Signed-out state: prompt to log in -->
        <RouterLink v-else to="/login" class="login-link">
          Войти
          <span class="login-arrow">→</span>
        </RouterLink>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useAuthStore } from '@/stores/auth.js'
import { useCartStore } from '@/stores/cart.js'
import { useThemeStore } from '@/stores/theme.js'
import { useRouter } from 'vue-router'

// Shared stores drive the auth-aware actions, cart badge, and theme toggle.
const auth = useAuthStore()
const cart = useCartStore()
const theme = useThemeStore()
const router = useRouter()
const scrolled = ref(false)

// Toggle the elevated header style after a few px of scroll.
function onScroll() { scrolled.value = window.scrollY > 8 }

// Track scroll position; passive listener since we never preventDefault. Run once on
// mount to set the initial state (e.g. when navigating in already scrolled).
onMounted(() => { onScroll(); window.addEventListener('scroll', onScroll, { passive: true }) })
onBeforeUnmount(() => window.removeEventListener('scroll', onScroll))

// Clear the session and send the user to the login page.
function handleLogout() { auth.logout(); router.push('/login') }
</script>

<style scoped>
.app-header {
  position: fixed; top: 0; left: 0; right: 0; z-index: 100;
  background: linear-gradient(180deg, var(--header-bg-from), var(--header-bg-to));
  backdrop-filter: blur(18px) saturate(140%);
  -webkit-backdrop-filter: blur(18px) saturate(140%);
  border-bottom: 1px solid transparent;
  transition: border-color .35s ease, background .35s ease, height .35s ease;
}
.app-header.scrolled {
  border-bottom-color: var(--border);
  background: linear-gradient(180deg, var(--header-bg-from-scrolled), var(--header-bg-to-scrolled));
}
.header-inner {
  max-width: 1440px; margin: 0 auto;
  padding: 0 32px;
  height: 72px;
  display: flex; align-items: center; gap: 40px;
}

.brand {
  display: inline-flex; align-items: baseline; gap: 8px;
  color: var(--text);
}
.brand-mark {
  display: inline-block;
  color: var(--accent);
  font-size: 16px;
  animation: text-glow 3.6s ease-in-out infinite;
}
.brand-name {
  font-weight: 700; letter-spacing: 0.28em; font-size: 14px;
}
.brand-tag {
  font-family: var(--font-serif); font-style: italic; font-weight: 300;
  font-size: 13px; color: var(--text-muted);
}

.nav-links { display: flex; gap: 28px; margin-left: auto; }
.nav-link {
  position: relative;
  color: var(--text-soft);
  font-size: 13px;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  padding: 6px 0;
  transition: color .2s ease;
}
.nav-link:hover { color: var(--text); }
.nav-link.router-link-active::after {
  content: ''; position: absolute; bottom: -2px; left: 0; right: 0;
  height: 1px; background: var(--accent);
  box-shadow: 0 0 8px var(--accent-glow);
}
.nav-sale { color: var(--sale-soft); display: inline-flex; align-items: center; gap: 6px; }
.nav-sale:hover { color: var(--sale-soft); }
.sale-dot {
  width: 6px; height: 6px; border-radius: 50%;
  background: var(--sale);
  box-shadow: 0 0 8px var(--sale-glow);
  animation: glow-pulse 1.6s ease-in-out infinite;
}

.header-actions { display: flex; align-items: center; gap: 4px; }
.action {
  position: relative;
  display: inline-flex; align-items: center; justify-content: center;
  width: 40px; height: 40px;
  border-radius: 50%;
  border: 1px solid transparent;
  background: transparent;
  color: var(--text-soft);
  cursor: pointer;
  transition: all .2s ease;
}
.action:hover { color: var(--text); border-color: var(--border); background: var(--bg-surface); }
.admin-action { color: var(--accent-soft); }
.admin-action:hover { color: var(--accent); }
.cart-badge {
  position: absolute; top: 4px; right: 4px;
  min-width: 16px; height: 16px;
  border-radius: 8px; padding: 0 4px;
  background: var(--accent);
  color: #fff;
  font: 600 10px/16px var(--font-mono);
  text-align: center;
  box-shadow: 0 0 0 2px var(--bg), 0 0 14px -2px var(--accent-glow);
}

/* Theme toggle — brand-style icon swap */
.theme-toggle {
  position: relative;
  overflow: hidden;
}
.theme-toggle::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: radial-gradient(circle at center, var(--accent-haze), transparent 70%);
  opacity: 0;
  transition: opacity .25s ease;
  pointer-events: none;
}
.theme-toggle:hover::before { opacity: 1; }
.theme-toggle .theme-icon {
  position: absolute;
  top: 50%; left: 50%;
  transform: translate(-50%, -50%) rotate(0deg) scale(1);
  transition: transform .45s cubic-bezier(.5,.1,.3,1.4), opacity .25s ease;
}
.theme-toggle .theme-icon.sun {
  opacity: 1;
  color: var(--accent);
}
.theme-toggle .theme-icon.moon {
  opacity: 0;
  color: var(--text);
  transform: translate(-50%, -50%) rotate(-90deg) scale(0.5);
}
.theme-toggle.is-dark .theme-icon.sun {
  opacity: 0;
  transform: translate(-50%, -50%) rotate(90deg) scale(0.5);
}
.theme-toggle.is-dark .theme-icon.moon {
  opacity: 1;
  color: var(--accent-soft);
  transform: translate(-50%, -50%) rotate(0deg) scale(1);
}

.login-link {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 8px 14px;
  font-size: 12px; letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--text);
  border: 1px solid var(--border-strong);
  transition: all .2s ease;
}
.login-link:hover { background: var(--bg-surface); border-color: var(--text); transform: translateX(2px); }
.login-arrow { font-family: var(--font-mono); }

@media (max-width: 768px) {
  .header-inner { padding: 0 16px; gap: 14px; height: 60px; }
  .brand-tag { display: none; }
  .nav-links { gap: 16px; margin-left: 0; }
  .nav-link { font-size: 11px; }
  .login-link { padding: 6px 10px; font-size: 11px; }
}
</style>

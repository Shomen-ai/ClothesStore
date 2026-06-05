<!--
  Root layout component: persistent header + footer wrapping the routed page.
  Rendered once and kept mounted for the whole app lifetime; only <RouterView />
  swaps as the route changes.
-->
<template>
  <AppHeader />
  <main class="app-main">
    <!-- Active route's view component renders here -->
    <RouterView />
  </main>
  <AppFooter />
</template>

<script setup>
import AppHeader from '@/components/layout/AppHeader.vue'
import AppFooter from '@/components/layout/AppFooter.vue'
import { useAuthStore } from '@/stores/auth.js'

// Restore the persisted session (user from localStorage) on app startup so the
// header and route guards see the logged-in state immediately after a reload.
const auth = useAuthStore()
auth.init()
</script>

<style>
.app-main {
  min-height: calc(100vh - 320px);
  padding-top: 72px;
}
@media (max-width: 768px) {
  .app-main { padding-top: 60px; }
}
</style>

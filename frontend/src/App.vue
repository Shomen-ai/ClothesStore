<template>
  <AppHeader />
  <main class="app-main">
    <RouterView />
  </main>
  <AppFooter />
</template>

<script setup>
import AppHeader from '@/components/layout/AppHeader.vue'
import AppFooter from '@/components/layout/AppFooter.vue'
import { useAuthStore } from '@/stores/auth.js'

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

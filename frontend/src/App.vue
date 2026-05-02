<template>
  <AppHeader />
  <main style="min-height: calc(100vh - 120px); padding-top: 64px;">
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

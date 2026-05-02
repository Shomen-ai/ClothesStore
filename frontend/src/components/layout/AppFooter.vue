<template>
  <footer class="app-footer">
    <div class="footer-inner">
      <span class="logo">STORE</span>
      <span class="copy">© {{ year }}</span>
    </div>
  </footer>
</template>

<script setup>
const year = new Date().getFullYear()
</script>

<style scoped>
.app-footer {
  border-top: 1px solid var(--color-border);
  padding: 24px;
  text-align: center;
}
.footer-inner { display: flex; justify-content: space-between; max-width: 1400px; margin: 0 auto; }
.logo { font-weight: 800; letter-spacing: 4px; font-size: 14px; }
.copy { color: var(--color-text-muted); font-size: 12px; }
</style>

<template>
  <div class="admin-page">
    <aside class="admin-sidebar">
      <p class="sidebar-title">ADMIN</p>
      <nav>
        <RouterLink to="/admin">Дашборд</RouterLink>
        <RouterLink to="/admin/products">Товары</RouterLink>
        <RouterLink to="/admin/orders">Заказы</RouterLink>
        <RouterLink to="/admin/promo-codes">Промокоды</RouterLink>
      </nav>
    </aside>
    <div class="admin-content">
      <RouterView />
    </div>
  </div>
</template>

<style scoped>
.admin-page { display: flex; min-height: calc(100vh - 64px); }
.admin-sidebar { width: 220px; border-right: 1px solid var(--color-border); padding: 32px 24px; flex-shrink: 0; }
.sidebar-title { font-size: 11px; letter-spacing: 3px; color: var(--color-red); margin-bottom: 24px; font-weight: 700; }
.admin-sidebar nav a { display: block; padding: 12px 0; color: var(--color-text-muted); font-size: 14px; border-bottom: 1px solid var(--color-border); }
.admin-sidebar nav a:hover, .admin-sidebar nav a.router-link-exact-active { color: var(--color-text); }
.admin-content { flex: 1; padding: 40px; overflow-y: auto; }
</style>

<!--
  AdminLayout — shell for the admin panel (route: /admin, meta.requiresAdmin).
  Renders the fixed sidebar navigation and a <RouterView> outlet for the admin child routes
  (dashboard / products / orders / promo-codes / reviews). Access is gated by the router
  guard which checks auth.user.role === 'admin'; this layout itself does no auth checks.
-->
<template>
  <div class="admin-page">
    <aside class="admin-sidebar">
      <div class="sidebar-brand">
        <span class="brand-mark">◆</span>
        <span class="brand-name">STORE</span>
        <span class="brand-tag">/ <span class="gothic-accent">админка</span></span>
      </div>

      <nav class="sidebar-nav">
        <RouterLink to="/admin" exact-active-class="active" class="nav-item">
          <span class="nav-bar"></span>
          <span class="nav-num">01</span>
          <span class="nav-label">Дашборд</span>
        </RouterLink>
        <RouterLink to="/admin/products" active-class="active" class="nav-item">
          <span class="nav-bar"></span>
          <span class="nav-num">02</span>
          <span class="nav-label">Товары</span>
        </RouterLink>
        <RouterLink to="/admin/orders" active-class="active" class="nav-item">
          <span class="nav-bar"></span>
          <span class="nav-num">03</span>
          <span class="nav-label">Заказы</span>
        </RouterLink>
        <RouterLink to="/admin/promo-codes" active-class="active" class="nav-item">
          <span class="nav-bar"></span>
          <span class="nav-num">04</span>
          <span class="nav-label">Промокоды</span>
        </RouterLink>
        <RouterLink to="/admin/reviews" active-class="active" class="nav-item">
          <span class="nav-bar"></span>
          <span class="nav-num">05</span>
          <span class="nav-label">Отзывы</span>
        </RouterLink>
      </nav>

      <div class="sidebar-footer">
        <RouterLink to="/" class="exit-link">
          <span class="exit-arrow">←</span>
          <span>На витрину</span>
        </RouterLink>
        <p class="sidebar-meta">v1.0 · made loud</p>
      </div>
    </aside>

    <div class="admin-content">
      <!-- Outlet for nested /admin/* child routes -->
      <RouterView />
    </div>
  </div>
</template>

<style scoped>
.admin-page {
  display: flex;
  min-height: calc(100vh - 64px);
}

.admin-sidebar {
  width: 240px;
  flex-shrink: 0;
  position: sticky;
  top: 72px;
  align-self: flex-start;
  height: calc(100vh - 72px);
  border-right: 1px solid var(--border);
  background: linear-gradient(180deg, var(--header-bg-from), var(--header-bg-to));
  backdrop-filter: blur(12px);
  padding: 32px 0 24px;
  display: flex;
  flex-direction: column;
}

.sidebar-brand {
  display: flex;
  align-items: baseline;
  gap: 8px;
  padding: 0 24px 28px;
  border-bottom: 1px solid var(--border);
  margin-bottom: 20px;
}
.sidebar-brand .brand-mark {
  color: var(--accent);
  font-size: 16px;
  filter: drop-shadow(0 0 12px var(--accent-glow));
}
.sidebar-brand .brand-name {
  font-weight: 700;
  letter-spacing: 0.28em;
  font-size: 14px;
}
.sidebar-brand .brand-tag {
  font-family: var(--font-serif);
  font-style: italic;
  font-size: 13px;
  color: var(--text-muted);
}
.sidebar-brand .brand-tag .gothic-accent {
  font-style: normal;
  font-size: 16px;
}

.sidebar-nav {
  display: flex;
  flex-direction: column;
  flex: 1;
}
.nav-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 24px;
  color: var(--text-soft);
  font-size: 14px;
  letter-spacing: 0.04em;
  border-bottom: 1px solid var(--border);
  transition: color .2s ease, background .2s ease, padding-left .2s ease;
}
.nav-item:hover {
  color: var(--text);
  background: var(--bg-elevated);
  padding-left: 30px;
}
.nav-item:hover .nav-bar { background: var(--text-muted); }
.nav-item:hover .nav-num { color: var(--text); }
.nav-item .nav-bar {
  width: 2px;
  height: 16px;
  background: transparent;
  transition: background .2s ease, box-shadow .2s ease;
}
.nav-item .nav-num {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-dim);
  letter-spacing: 0.1em;
}
.nav-item .nav-label {
  font-weight: 500;
}
.nav-item.active {
  color: var(--text);
  background: linear-gradient(90deg, var(--accent-haze), transparent);
  padding-left: 30px;
}
.nav-item.active .nav-bar {
  background: var(--accent);
  box-shadow: 0 0 12px var(--accent-glow);
}
.nav-item.active .nav-num {
  color: var(--accent-soft);
}

.sidebar-footer {
  padding: 20px 24px 0;
  border-top: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.exit-link {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--text-muted);
  font-size: 12px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  transition: color .2s ease;
}
.exit-link:hover { color: var(--text); }
.exit-arrow { font-family: var(--font-mono); transition: transform .2s ease; }
.exit-link:hover .exit-arrow { transform: translateX(-3px); }
.sidebar-meta {
  font-family: var(--font-mono);
  font-size: 10px;
  color: var(--text-dim);
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.admin-content {
  flex: 1;
  padding: 48px 56px;
  overflow-y: auto;
  min-width: 0;
}

@media (max-width: 900px) {
  .admin-page { flex-direction: column; }
  .admin-sidebar {
    width: 100%;
    height: auto;
    position: static;
    border-right: none;
    border-bottom: 1px solid var(--border);
    padding: 16px;
  }
  .sidebar-brand { padding: 0 0 14px; margin-bottom: 14px; }
  .sidebar-nav { flex-direction: row; overflow-x: auto; gap: 0; }
  .nav-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
    padding: 10px 14px;
    border-bottom: none;
    border-right: 1px solid var(--border);
    flex-shrink: 0;
  }
  .nav-item .nav-bar { display: none; }
  .nav-item.active { padding-left: 14px; }
  .sidebar-footer { flex-direction: row; justify-content: space-between; align-items: center; padding: 12px 0 0; margin-top: 12px; }
  .admin-content { padding: 24px 16px; }
}
</style>

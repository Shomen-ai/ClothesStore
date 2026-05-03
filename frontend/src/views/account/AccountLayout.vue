<template>
  <div class="account-page">
    <div class="account-layout">
      <nav class="account-nav">
        <RouterLink to="/account">Профиль</RouterLink>
        <RouterLink to="/account/orders">Заказы</RouterLink>
        <RouterLink to="/account/addresses">Адреса</RouterLink>
        <RouterLink to="/account/wishlist">Избранное</RouterLink>
      </nav>
      <div class="account-content">
        <RouterView />
      </div>
    </div>
  </div>
</template>

<style scoped>
.account-page { max-width: 1200px; margin: 40px auto; padding: 0 24px; }
.account-layout { display: grid; grid-template-columns: 220px 1fr; gap: 48px; }
.account-nav { border-right: 1px solid var(--color-border); padding-right: 32px; }
.account-nav a { display: block; padding: 12px 0; color: var(--color-text-muted); font-size: 14px; letter-spacing: 1px; border-bottom: 1px solid var(--color-border); }
.account-nav a:hover, .account-nav a.router-link-exact-active { color: var(--color-text); }

@media (max-width: 768px) {
  .account-page { margin: 16px auto; padding: 0 16px; }
  .account-layout { grid-template-columns: 1fr; gap: 16px; }
  .account-nav { border-right: none; padding-right: 0; border-bottom: 1px solid var(--color-border); padding-bottom: 8px; display: flex; gap: 0; overflow-x: auto; }
  .account-nav a { display: inline-block; padding: 10px 14px; border-bottom: none; border-right: 1px solid var(--color-border); white-space: nowrap; font-size: 13px; }
  .account-nav a:last-child { border-right: none; }
}
</style>

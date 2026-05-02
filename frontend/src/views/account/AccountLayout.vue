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
</style>

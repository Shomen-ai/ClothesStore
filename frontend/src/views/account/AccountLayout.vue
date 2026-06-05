<!--
  AccountLayout — shell for the customer account area (route: /account, requiresAuth).
  Renders the decorative header, the side navigation, and a <RouterView> outlet for the
  nested child routes (profile / orders / addresses / wishlist). Purely presentational:
  no data fetching here — each child view loads its own data.
-->
<template>
  <div class="account-page">
    <header class="account-head">
      <div class="head-marquee-wrap" aria-hidden="true">
        <div class="head-marquee">
          <div class="marquee-row row-1">
            <div class="marquee-track">
              <span>PROFILE · ORDERS · WISHLIST · ADDRESSES · HISTORY · YOU ·&nbsp;</span>
              <span>PROFILE · ORDERS · WISHLIST · ADDRESSES · HISTORY · YOU ·&nbsp;</span>
            </div>
          </div>
          <div class="marquee-row row-2">
            <div class="marquee-track">
              <span>аккаунт — твоя территория — drop history — signature — fits & shifts —&nbsp;</span>
              <span>аккаунт — твоя территория — drop history — signature — fits & shifts —&nbsp;</span>
            </div>
          </div>
          <div class="marquee-row row-3">
            <div class="marquee-track">
              <span>PERSONAL · DROPS · SAVED · ARCHIVE · YOURS · LOUD ·&nbsp;</span>
              <span>PERSONAL · DROPS · SAVED · ARCHIVE · YOURS · LOUD ·&nbsp;</span>
            </div>
          </div>
        </div>
      </div>

      <p class="eyebrow">Личный кабинет</p>
      <h1 class="title">
        Только <span class="gothic-accent">твоё</span>
      </h1>
    </header>

    <div class="account-layout">
      <nav class="account-nav">
        <RouterLink to="/account">Профиль</RouterLink>
        <RouterLink to="/account/orders">Заказы</RouterLink>
        <RouterLink to="/account/addresses">Адреса</RouterLink>
        <RouterLink to="/account/wishlist">Избранное</RouterLink>
      </nav>
      <div class="account-content">
        <!-- Outlet for nested /account/* child routes -->
        <RouterView />
      </div>
    </div>
  </div>
</template>

<style scoped>
.account-page { max-width: 1200px; margin: 0 auto; padding: 80px 24px 0; }

.account-head {
  position: relative;
  padding: 0 0 24px;
  margin-bottom: 40px;
  min-height: 180px;
  border-bottom: 1px solid var(--border);
  overflow: visible;
  isolation: isolate;
  display: flex; flex-direction: column; gap: 8px; justify-content: flex-end;
}
.account-head .eyebrow, .account-head .title { position: relative; z-index: 2; }
.account-head .title {
  font-size: clamp(36px, 5vw, 64px);
  font-weight: 700;
  letter-spacing: -0.03em;
  line-height: 1;
  color: var(--text);
}
.account-head .title .gothic-accent { font-size: 1.05em; }

.account-layout { display: grid; grid-template-columns: 220px 1fr; gap: 48px; }
.account-nav { border-right: 1px solid var(--color-border); padding-right: 32px; }
.account-nav a { display: block; padding: 12px 0; color: var(--color-text-muted); font-size: 14px; letter-spacing: 1px; border-bottom: 1px solid var(--color-border); }
.account-nav a:hover, .account-nav a.router-link-exact-active { color: var(--color-text); }

@media (max-width: 768px) {
  .account-page { padding: 56px 16px 0; }
  .account-head { min-height: 140px; padding-bottom: 16px; margin-bottom: 24px; }
  .account-layout { grid-template-columns: 1fr; gap: 16px; }
  .account-nav { border-right: none; padding-right: 0; border-bottom: 1px solid var(--color-border); padding-bottom: 8px; display: flex; gap: 0; overflow-x: auto; }
  .account-nav a { display: inline-block; padding: 10px 14px; border-bottom: none; border-right: 1px solid var(--color-border); white-space: nowrap; font-size: 13px; }
  .account-nav a:last-child { border-right: none; }
}
</style>

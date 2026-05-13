<template>
  <footer class="app-footer">
    <div class="footer-grid">
      <div class="brand-block">
        <div class="brand-line">
          <span class="brand-mark">◆</span>
          <span class="brand-name">STORE</span>
        </div>
        <p class="brand-poem">
          <span class="gothic-accent poem-accent">Громкие fits</span>
          <br />для тех, кто не как все.
        </p>
      </div>

      <div class="links-col">
        <p class="col-title">Каталог</p>
        <RouterLink to="/catalogue">Все товары</RouterLink>
        <RouterLink to="/catalogue?sale=1">Распродажа</RouterLink>
      </div>

      <div class="links-col">
        <p class="col-title">Аккаунт</p>
        <RouterLink to="/account">Профиль</RouterLink>
        <RouterLink to="/account/orders">Заказы</RouterLink>
        <RouterLink to="/account/wishlist">Избранное</RouterLink>
      </div>

      <div class="links-col">
        <p class="col-title">Контакт</p>
        <a href="mailto:hello@store.local">hello@store.local</a>
        <span class="muted">© {{ year }} STORE</span>
      </div>
    </div>

    <div class="strip">
      <span class="strip-item">made <span class="gothic-accent">loud</span></span>
      <span class="strip-dot">●</span>
      <span class="strip-item serif-italic">ships worldwide</span>
      <span class="strip-dot">●</span>
      <span class="strip-item">v1.0</span>
    </div>
  </footer>
</template>

<script setup>
const year = new Date().getFullYear()
</script>

<style scoped>
.app-footer {
  border-top: 1px solid var(--border);
  margin-top: 120px;
  padding: 80px 32px 40px;
  position: relative;
  overflow: hidden;
}
.app-footer::before {
  content: '';
  position: absolute; inset: 0;
  background: radial-gradient(60% 80% at 50% 0%, var(--accent-haze), transparent 70%);
  pointer-events: none;
}
.footer-grid {
  position: relative;
  max-width: 1440px; margin: 0 auto;
  display: grid; grid-template-columns: 2fr 1fr 1fr 1fr;
  gap: 64px;
  padding-bottom: 64px;
}
.brand-block { max-width: 320px; }
.brand-line { display: flex; align-items: baseline; gap: 8px; margin-bottom: 16px; }
.brand-mark { color: var(--accent); font-size: 16px; filter: drop-shadow(0 0 12px var(--accent-glow)); }
.brand-name { font-weight: 700; letter-spacing: 0.28em; font-size: 14px; }
.brand-poem { font-size: 15px; line-height: 1.6; color: var(--text-soft); }
.brand-poem .gothic-accent { font-size: 24px; line-height: 1.2; }

.links-col { display: flex; flex-direction: column; gap: 10px; }
.col-title {
  font-family: var(--font-mono);
  text-transform: uppercase;
  letter-spacing: 0.18em;
  font-size: 11px;
  color: var(--text-muted);
  margin-bottom: 6px;
}
.links-col a { color: var(--text-soft); font-size: 14px; }
.links-col a:hover { color: var(--accent-soft); }
.muted { color: var(--text-dim); font-size: 13px; }

.strip {
  position: relative;
  max-width: 1440px; margin: 0 auto;
  padding-top: 24px;
  border-top: 1px dashed var(--border);
  display: flex; gap: 12px; justify-content: center;
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--text-dim);
}
.strip .serif-italic { text-transform: none; font-family: var(--font-serif); letter-spacing: 0; font-size: 13px; color: var(--text-muted); }
.strip .gothic-accent { text-transform: none; font-size: 14px; letter-spacing: 0; }
.strip-dot { color: var(--accent); }

@media (max-width: 768px) {
  .app-footer { padding: 56px 16px 32px; margin-top: 64px; }
  .footer-grid { grid-template-columns: 1fr 1fr; gap: 40px; padding-bottom: 40px; }
  .brand-block { grid-column: span 2; }
}
</style>

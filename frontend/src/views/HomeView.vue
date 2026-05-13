<template>
  <div class="home">
    <!-- HERO -->
    <section class="hero">
      <div class="hero-grid">
        <div class="hero-left">
          <p class="eyebrow">— SS / 26 collection</p>
          <h1 class="hero-title">
            Носи
            <span class="serif-italic">громко</span>.
            <br />
            Не как все.
          </h1>
          <p class="hero-sub">
            Streetwear для тех, кто не вписывается в дресс-код. Громкие принты, дерзкие силуэты, ноль компромиссов.
          </p>
          <div class="hero-cta">
            <RouterLink to="/catalogue" class="btn btn-primary">
              <span>Смотреть каталог</span>
              <span class="arrow">→</span>
            </RouterLink>
            <RouterLink to="/catalogue?sale=1" class="hero-sale">
              <span class="dot"></span>
              <span>На распродаже</span>
            </RouterLink>
          </div>
        </div>

        <aside class="hero-right">
          <div class="hero-spec">
            <p class="spec-num">04</p>
            <p class="spec-label">сезона <br />в одной вещи</p>
          </div>
          <div class="hero-spec">
            <p class="spec-num">14</p>
            <p class="spec-label">категорий</p>
          </div>
          <div class="hero-spec spec-wide">
            <p class="spec-num"><span class="serif-italic">∞</span></p>
            <p class="spec-label">возможностей <br />комбинаций</p>
          </div>
        </aside>
      </div>
    </section>

    <!-- HITS -->
    <section v-if="hits.length" class="section">
      <header class="sec-head">
        <span class="eyebrow">— Что покупают</span>
        <h2 class="sec-title">Хиты <span class="serif-italic">продаж</span></h2>
        <RouterLink to="/catalogue" class="sec-link">Все товары →</RouterLink>
      </header>
      <div class="grid">
        <ProductCard v-for="(p, i) in hits" :key="p.id" :product="p"
          :style="{ animationDelay: `${i * 60}ms` }" />
      </div>
    </section>

    <!-- NEW -->
    <section v-if="newest.length" class="section">
      <header class="sec-head">
        <span class="eyebrow">— На полках</span>
        <h2 class="sec-title">Только <span class="serif-italic">приехало</span></h2>
        <RouterLink to="/catalogue" class="sec-link">Все товары →</RouterLink>
      </header>
      <div class="grid">
        <ProductCard v-for="(p, i) in newest" :key="p.id" :product="p"
          :style="{ animationDelay: `${i * 60}ms` }" />
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getFeatured } from '@/api/catalogue.js'
import { useWishlistStore } from '@/stores/wishlist.js'
import ProductCard from '@/components/catalogue/ProductCard.vue'

const hits = ref([])
const newest = ref([])
const wishlist = useWishlistStore()

onMounted(async () => {
  wishlist.load()
  const { data } = await getFeatured()
  hits.value = data.hits || []
  newest.value = data.new || []
})
</script>

<style scoped>
.home { max-width: 1440px; margin: 0 auto; padding: 0 32px; }

.hero {
  position: relative;
  padding: 96px 0 80px;
  overflow: hidden;
}
.hero::before {
  content: '';
  position: absolute; inset: 0;
  background:
    radial-gradient(60% 80% at 20% 30%, var(--accent-haze), transparent 60%),
    radial-gradient(40% 60% at 90% 80%, rgba(255, 120, 73, 0.07), transparent 60%);
  pointer-events: none;
  z-index: -1;
}
.hero-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.6fr) minmax(0, 1fr);
  gap: 56px;
  align-items: end;
}
.hero-left .eyebrow { margin-bottom: 24px; display: block; }
.hero-title {
  font-size: clamp(48px, 8vw, 120px);
  line-height: 0.92;
  letter-spacing: -0.04em;
  font-weight: 600;
  color: var(--text);
}
.hero-title .serif-italic { font-weight: 300; color: var(--accent-soft); font-size: 0.95em; }
.hero-sub {
  font-size: 16px;
  line-height: 1.65;
  color: var(--text-soft);
  max-width: 44ch;
  margin: 32px 0 40px;
}
.hero-cta { display: flex; align-items: center; gap: 24px; }
.btn-primary .arrow { font-family: var(--font-mono); margin-left: auto; }
.hero-sale {
  display: inline-flex; align-items: center; gap: 8px;
  color: var(--sale-soft);
  font-size: 13px;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}
.hero-sale .dot { width: 6px; height: 6px; border-radius: 50%; background: var(--sale); box-shadow: 0 0 8px var(--sale-glow); animation: glow-pulse 1.6s ease-in-out infinite; }
.hero-sale:hover { color: var(--sale); }

.hero-right {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}
.hero-spec {
  border: 1px solid var(--border);
  background: var(--bg-surface);
  padding: 24px;
  display: flex; flex-direction: column; gap: 12px;
}
.hero-spec.spec-wide { grid-column: span 2; }
.spec-num {
  font-family: var(--font-mono);
  font-size: 56px;
  font-weight: 300;
  letter-spacing: -0.02em;
  color: var(--text);
}
.spec-num .serif-italic { font-family: var(--font-serif); color: var(--accent-soft); }
.spec-label {
  font-size: 11px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--text-muted);
  font-family: var(--font-mono);
}

.section { padding: 80px 0 24px; }
.sec-head {
  display: flex; justify-content: space-between; align-items: flex-end;
  margin-bottom: 40px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border);
  gap: 24px;
  flex-wrap: wrap;
}
.sec-head .eyebrow { display: block; margin-bottom: 8px; }
.sec-title {
  font-size: clamp(28px, 4vw, 48px);
  font-weight: 600;
  letter-spacing: -0.02em;
  line-height: 1;
}
.sec-title .serif-italic { font-weight: 300; color: var(--accent-soft); }
.sec-link {
  color: var(--text-soft);
  font-family: var(--font-mono);
  font-size: 12px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}
.sec-link:hover { color: var(--accent-soft); }

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 32px 20px;
}
.grid > * {
  opacity: 0;
  animation: fade-up .6s ease forwards;
}

@media (max-width: 900px) {
  .home { padding: 0 16px; }
  .hero { padding: 56px 0 40px; }
  .hero-grid { grid-template-columns: 1fr; gap: 32px; }
  .hero-right { order: -1; }
  .spec-num { font-size: 36px; }
  .hero-sub { font-size: 15px; margin: 20px 0 28px; }
  .hero-cta { flex-direction: column; align-items: flex-start; gap: 16px; }
  .section { padding: 56px 0; }
  .grid { grid-template-columns: repeat(2, 1fr); gap: 20px 12px; }
}
</style>

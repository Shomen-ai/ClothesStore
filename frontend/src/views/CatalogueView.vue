<!--
  CatalogueView - product listing with search, filters and sort (route: /catalogue).
  Reads the optional ?sale=1 query, fetches categories + products, and re-queries
  the API whenever a filter changes (search input is debounced).
-->
<template>
  <div class="page">
    <header class="page-head">
      <div class="head-marquee-wrap" aria-hidden="true">
        <div class="head-marquee">
          <div class="marquee-row row-1">
            <div class="marquee-track">
              <span>OVERSIZE · DENIM · HOODIE · LEATHER · CARGO · FLEECE · KNIT · COAT · JACKET · TEE ·&nbsp;</span>
              <span>OVERSIZE · DENIM · HOODIE · LEATHER · CARGO · FLEECE · KNIT · COAT · JACKET · TEE ·&nbsp;</span>
            </div>
          </div>
          <div class="marquee-row row-2">
            <div class="marquee-track">
              <span>loud — raw — never quiet — anti — signal — void — drop — ritual — signature —&nbsp;</span>
              <span>loud — raw — never quiet — anti — signal — void — drop — ritual — signature —&nbsp;</span>
            </div>
          </div>
          <div class="marquee-row row-3">
            <div class="marquee-track">
              <span>не как все · громко · drop /26 · edition · made for noise · sale · beware ·&nbsp;</span>
              <span>не как все · громко · drop /26 · edition · made for noise · sale · beware ·&nbsp;</span>
            </div>
          </div>
        </div>
      </div>

      <div class="head-left">
        <p class="eyebrow">Каталог · SS / 26</p>
        <h1 class="head-title">
          Носи <span class="gothic-accent">громко</span>,
          <br />не как все.
        </h1>
      </div>
      <div class="head-meta">
        <span class="meta-num">{{ String(products.length).padStart(3, '0') }}</span>
        <span class="meta-label">видны</span>
        <span class="meta-sep">/</span>
        <span class="meta-num">14</span>
        <span class="meta-label">категорий</span>
      </div>
    </header>

    <div class="search-bar">
      <span class="search-icon">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <circle cx="6" cy="6" r="4.5" stroke="currentColor" stroke-width="1.2"/>
          <path d="M9.5 9.5 13 13" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/>
        </svg>
      </span>
      <input v-model="filters.q" placeholder="Поиск по названию…" @input="debouncedLoad" />
      <span v-if="filters.q" class="search-clear" @click="filters.q=''; load()">✕</span>
    </div>

    <button class="filters-toggle" @click="filtersOpen = !filtersOpen">
      <span>{{ filtersOpen ? 'Свернуть' : 'Фильтры' }}</span>
      <span class="dot"></span>
    </button>

    <div class="layout">
      <aside class="sidebar" :class="{ open: filtersOpen }">
        <ProductFilters
          :categories="categories"
          :model-category="filters.category"
          :model-size="filters.size"
          :model-sort="filters.sort"
          :model-price="filters.price"
          :model-sale="filters.sale"
          @update:category="v => { filters.category = v; load() }"
          @update:size="v => { filters.size = v; load() }"
          @update:sort="v => { filters.sort = v; load() }"
          @update:price="v => { filters.price = v; load() }"
          @update:sale="v => { filters.sale = v; load() }"
        />
      </aside>

      <section class="catalogue-main">
        <div v-if="loading" class="status status-mix">
          <span class="status-line">
            <span class="status-mono">load</span><span class="gothic-accent status-gothic">ing</span><span class="status-mono status-dots">...</span>
          </span>
        </div>
        <div v-else-if="products.length === 0" class="status status-mix">
          <span class="status-line">
            <span class="serif-italic status-italic">nothing</span>&nbsp;<span class="gothic-accent status-gothic">here</span>
          </span>
          <p class="status-hint">Попробуй другие фильтры или сбрось всё.</p>
        </div>
        <div v-else class="grid">
          <ProductCard v-for="(p, i) in products" :key="p.id" :product="p"
            :style="{ animationDelay: `${Math.min(i * 30, 600)}ms` }" />
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getCategories, getProducts } from '@/api/catalogue.js'
import { useWishlistStore } from '@/stores/wishlist.js'
import ProductCard from '@/components/catalogue/ProductCard.vue'
import ProductFilters from '@/components/catalogue/ProductFilters.vue'
import { priceBoundsFromIds } from '@/components/catalogue/priceRanges.js'

const products = ref([])
const categories = ref([])
const loading = ref(false)
const filters = ref({ category: null, size: '', sort: '', q: '', price: [], sale: false })
const filtersOpen = ref(false)
const wishlist = useWishlistStore()
const route = useRoute()

// Debounce search keystrokes so we don't fire a request on every character.
let debounceTimer = null
function debouncedLoad() {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(load, 400)
}

// Build the query from active filters and fetch matching products.
// Only non-empty filters are sent; price range ids map to min/max bounds.
async function load() {
  loading.value = true
  const params = { page: 1 }
  if (filters.value.category) params.category = filters.value.category
  if (filters.value.size) params.size = filters.value.size
  if (filters.value.sort) params.sort = filters.value.sort
  if (filters.value.q) params.q = filters.value.q
  if (filters.value.sale) params.sale = 1
  const { min, max } = priceBoundsFromIds(filters.value.price)
  if (min > 0) params.price_min = min
  if (max > 0) params.price_max = max
  const { data } = await getProducts(params)
  products.value = data || []
  loading.value = false
}

// Sync the "sale" filter from the URL so /catalogue?sale=1 deep-links work.
function applyQuery() {
  filters.value.sale = route.query.sale === '1' || route.query.sale === 'true'
}

// React to in-app navigation that only changes the query string.
watch(() => route.query, applyQuery)

onMounted(async () => {
  // Load wishlist (for heart state on cards), seed categories, then initial list.
  wishlist.load()
  applyQuery()
  const { data } = await getCategories()
  categories.value = data || []
  load()
})
</script>

<style scoped>
.page {
  max-width: 1440px;
  margin: 0 auto;
  padding: 80px 32px 0;
}

.page-head {
  position: relative;
  display: flex; justify-content: space-between; align-items: flex-end;
  gap: 32px;
  margin-bottom: 48px;
  padding: 0 0 24px;
  min-height: 180px;
  border-bottom: 1px solid var(--border);
  overflow: visible;
  isolation: isolate;
}
.head-left, .head-meta { position: relative; z-index: 2; }
.head-left .eyebrow { margin-bottom: 16px; }
.head-title {
  font-size: clamp(36px, 5vw, 64px);
  line-height: 0.95;
  font-weight: 700;
  letter-spacing: -0.03em;
  color: var(--text);
}
.head-title .gothic-accent { font-size: 1.05em; }

/* marquee styles live in main.css (reused on Cart / Account) */

.head-meta {
  display: flex; align-items: baseline; gap: 8px;
  font-family: var(--font-mono);
  color: var(--text-muted);
  font-size: 12px;
  flex-shrink: 0;
}
.meta-num { color: var(--text); font-size: 28px; font-weight: 300; }
.meta-label { letter-spacing: 0.1em; text-transform: uppercase; font-size: 10px; }
.meta-sep { color: var(--text-dim); margin: 0 4px; }

.search-bar {
  position: relative;
  margin-bottom: 32px;
  display: flex; align-items: center;
}
.search-bar input {
  flex: 1;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  color: var(--text);
  padding: 14px 44px;
  font-size: 15px;
  font-family: inherit;
  letter-spacing: 0.01em;
  outline: none;
  transition: border-color .2s ease;
}
.search-bar input::placeholder { color: var(--text-dim); }
.search-bar input:focus { border-color: var(--text-muted); }
.search-icon {
  position: absolute; left: 16px; color: var(--text-muted);
  pointer-events: none;
}
.search-clear {
  position: absolute; right: 16px;
  cursor: pointer; color: var(--text-muted); font-size: 13px;
}
.search-clear:hover { color: var(--text); }

.filters-toggle {
  display: none;
  margin-bottom: 20px;
}

.layout { display: grid; grid-template-columns: 220px 1fr; gap: 56px; }
.sidebar { position: sticky; top: 96px; align-self: start; }

.catalogue-main { min-height: 60vh; }
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 32px 20px;
}
.grid > * {
  opacity: 0;
  animation: fade-up .55s ease forwards;
}

.status {
  display: flex; flex-direction: column; align-items: center;
  padding: 120px 0;
  color: var(--text-muted);
  text-align: center;
  gap: 8px;
}
.status .serif-italic { font-size: 32px; color: var(--text-soft); }
.status-hint { font-size: 13px; color: var(--text-dim); margin-top: 8px; }
.ellipsis { font-family: var(--font-mono); color: var(--text-muted); }

.status-mix { padding: 140px 0; gap: 16px; }
.status-line {
  display: inline-flex; align-items: baseline; gap: 4px;
  font-size: clamp(40px, 5.4vw, 68px);
  line-height: 1;
}
.status-mono {
  font-family: var(--font-mono);
  font-weight: 300;
  letter-spacing: 0.04em;
  color: var(--text-soft);
  font-size: 0.72em;
}
.status-italic {
  font-family: var(--font-serif);
  font-style: italic;
  font-weight: 300;
  color: var(--text-soft);
  font-size: 0.85em;
}
.status-gothic {
  font-size: 1.15em;
  letter-spacing: 0.02em;
}
.status-dots {
  color: var(--accent);
  animation: dots-blink 1.4s steps(4, end) infinite;
}
@keyframes dots-blink {
  0%   { opacity: 0.2; transform: translateY(0); }
  50%  { opacity: 1;   transform: translateY(-2px); }
  100% { opacity: 0.2; transform: translateY(0); }
}

@media (max-width: 900px) {
  .page { padding: 56px 16px 0; }
  .page-head { flex-direction: column; align-items: flex-start; gap: 12px; margin-bottom: 32px; padding: 12px 0 16px; min-height: 150px; }
  .head-meta { font-size: 11px; }
  .meta-num { font-size: 22px; }

  .filters-toggle {
    display: inline-flex; align-items: center; gap: 10px;
    background: var(--bg-surface);
    border: 1px solid var(--border-strong);
    color: var(--text);
    padding: 10px 16px;
    font: 500 12px/1 var(--font-mono);
    letter-spacing: 0.16em;
    text-transform: uppercase;
    cursor: pointer;
  }
  .filters-toggle .dot { width: 6px; height: 6px; border-radius: 50%; background: var(--accent); box-shadow: 0 0 8px var(--accent-glow); }

  .layout { grid-template-columns: 1fr; gap: 24px; }
  .sidebar { display: none; position: static; padding: 20px; background: var(--bg-surface); border: 1px solid var(--border); }
  .sidebar.open { display: block; }
  .grid { grid-template-columns: repeat(2, 1fr); gap: 20px 12px; }
}
</style>

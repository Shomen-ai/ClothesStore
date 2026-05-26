<template>
  <div class="page" v-if="product">
    <nav class="breadcrumbs">
      <RouterLink to="/catalogue">Каталог</RouterLink>
      <span class="sep">/</span>
      <span class="muted">{{ product.name }}</span>
    </nav>

    <div class="layout">
      <!-- Gallery -->
      <div class="gallery">
        <div class="thumbnails" v-if="product.images?.length > 1">
          <button v-for="img in product.images" :key="img.id"
                  class="thumb"
                  :class="{ active: activeImage === img.image_path }"
                  @click="activeImage = img.image_path"
                  @mouseenter="activeImage = img.image_path">
            <img :src="img.image_path" :alt="product.name" loading="lazy" />
          </button>
        </div>

        <figure class="main-image">
          <Transition name="fade" mode="out-in">
            <img :key="activeImage" v-if="activeImage" :src="activeImage" :alt="product.name" />
          </Transition>
          <span v-if="product.is_on_sale" class="sale-tag">
            <span class="sale-dot"></span>on sale
          </span>
        </figure>
      </div>

      <!-- Info column (sticky) -->
      <div class="info">
        <span class="eyebrow">{{ categoryName }}</span>
        <h1 class="title">{{ product.name }}</h1>

        <div class="price-row">
          <span class="price">{{ formatPrice(product.price) }}</span>
          <span class="price-unit">RUB</span>
          <span v-if="product.is_on_sale" class="price-flag serif-italic">+ on sale</span>
        </div>

        <p class="desc" v-if="product.description">{{ product.description }}</p>

        <div class="size-block">
          <div class="size-head">
            <span class="eyebrow">Размер</span>
            <span v-if="selectedSize" class="size-pick">— {{ selectedSize.size }}</span>
          </div>
          <div class="sizes">
            <button v-for="s in product.sizes" :key="s.id"
                    class="size"
                    :class="{ selected: selectedSize?.id === s.id, gone: s.stock_qty === 0 }"
                    :disabled="s.stock_qty === 0"
                    @click="selectedSize = s">
              {{ s.size }}
              <span v-if="s.stock_qty === 0" class="strike"></span>
            </button>
          </div>
          <p v-if="!product.sizes?.length" class="no-sizes">Без размеров — одинарная позиция.</p>
        </div>

        <div class="actions">
          <button class="btn btn-primary cta" :disabled="!selectedSize && product.sizes?.length" @click="addToCart">
            <span>В корзину</span>
            <span class="cta-arrow">→</span>
          </button>
          <button class="wish-btn" :class="{ active: inWishlist }" @click="wishlist.toggle(product.id)" aria-label="В избранное">
            <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
              <path d="M9 16s-6-3.7-6-8.4A3.3 3.3 0 0 1 9 5.4a3.3 3.3 0 0 1 6 2.2C15 12.3 9 16 9 16Z"
                    stroke="currentColor" stroke-width="1.5"
                    :fill="inWishlist ? 'currentColor' : 'none'"/>
            </svg>
          </button>
        </div>

        <ul class="bullets">
          <li><span class="b-key">Доставка</span><span class="b-val">3–7 дней по России</span></li>
          <li><span class="b-key">Возврат</span><span class="b-val">14 дней без вопросов</span></li>
          <li><span class="b-key">Артикул</span><span class="b-val">#{{ String(product.id).padStart(4, '0') }}</span></li>
        </ul>
      </div>
    </div>
  </div>

  <div v-else class="page-loading">
    <span class="serif-italic">loading</span>
    <span class="ellipsis">...</span>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getProduct, getCategories } from '@/api/catalogue.js'
import { useCartStore } from '@/stores/cart.js'
import { useWishlistStore } from '@/stores/wishlist.js'
import { ElMessage } from 'element-plus'

const route = useRoute()
const cart = useCartStore()
const wishlist = useWishlistStore()

const product = ref(null)
const categories = ref([])
const selectedSize = ref(null)
const activeImage = ref('')

const inWishlist = computed(() => wishlist.has(product.value?.id))
const categoryName = computed(() => categories.value.find(c => c.id === product.value?.category_id)?.name || '—')

async function loadProduct(id) {
  const { data } = await getProduct(id)
  product.value = data
  activeImage.value = data.images?.find(i => i.is_primary)?.image_path || data.images?.[0]?.image_path || ''
  selectedSize.value = null
}

onMounted(async () => {
  wishlist.load()
  const [, catsRes] = await Promise.all([
    loadProduct(route.params.id),
    getCategories(),
  ])
  categories.value = catsRes.data || []
})

watch(() => route.params.id, (id) => { if (id) loadProduct(id) })

function addToCart() {
  if (product.value.sizes?.length && !selectedSize.value) return
  const sizeID = selectedSize.value?.id || 0
  const sizeLabel = selectedSize.value?.size || '—'
  cart.add(product.value, sizeID, sizeLabel)
  ElMessage({ message: '✓ Добавлено в корзину', type: 'success', duration: 1500 })
}

function formatPrice(p) {
  return new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 0 }).format(p)
}
</script>

<style scoped>
.page { max-width: 1440px; margin: 0 auto; padding: 64px 32px 0; }
.page-loading {
  min-height: 60vh;
  display: flex; align-items: center; justify-content: center; gap: 6px;
  color: var(--text-muted);
}
.page-loading .serif-italic { font-size: 28px; color: var(--text-soft); }
.ellipsis { font-family: var(--font-mono); }

.breadcrumbs {
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 28px;
}
.breadcrumbs a { color: var(--text-muted); }
.breadcrumbs a:hover { color: var(--text); }
.breadcrumbs .sep { margin: 0 8px; color: var(--text-dim); }
.breadcrumbs .muted { color: var(--text); }

.layout {
  display: grid;
  grid-template-columns: minmax(0, 1.15fr) minmax(0, 0.85fr);
  gap: 64px;
  align-items: start;
}

.gallery { display: grid; grid-template-columns: 72px 1fr; gap: 16px; }
.thumbnails {
  display: flex; flex-direction: column; gap: 8px;
  max-height: 80vh; overflow-y: auto;
  padding-right: 2px;
}
.thumbnails::-webkit-scrollbar { width: 4px; }
.thumbnails::-webkit-scrollbar-thumb { background: var(--border-strong); }
.thumb {
  width: 72px; aspect-ratio: 4/5;
  padding: 0; background: var(--bg-surface);
  border: 1px solid var(--border);
  cursor: pointer; overflow: hidden;
  transition: all .15s ease;
  opacity: 0.55;
}
.thumb img { width: 100%; height: 100%; object-fit: cover; }
.thumb:hover { opacity: 0.85; }
.thumb.active { opacity: 1; border-color: var(--accent); box-shadow: 0 0 0 1px var(--accent), 0 0 16px -4px var(--accent-glow); }

.main-image {
  position: relative;
  aspect-ratio: 4/5;
  background: linear-gradient(135deg, rgba(255,255,255,0.04), rgba(255,255,255,0.02));
  border: 1px solid var(--border);
  overflow: hidden;
}
.main-image img { width: 100%; height: 100%; object-fit: cover; display: block; }
.fade-enter-active, .fade-leave-active { transition: opacity .25s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

.sale-tag {
  position: absolute; top: 14px; left: 14px;
  display: inline-flex; align-items: center; gap: 6px;
  padding: 6px 10px;
  font-family: var(--font-mono);
  font-size: 10px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--sale-soft);
  background: var(--bg-overlay);
  border: 1px solid var(--sale);
  backdrop-filter: blur(8px);
  box-shadow: 0 0 20px -4px var(--sale-glow);
}
.sale-tag .sale-dot {
  width: 5px; height: 5px; border-radius: 50%;
  background: var(--sale); box-shadow: 0 0 6px var(--sale-glow);
}

/* Info column */
.info {
  position: sticky; top: 96px;
  display: flex; flex-direction: column; gap: 28px;
}
.info > .eyebrow { margin: 0; }
.title {
  font-size: clamp(28px, 3.5vw, 44px);
  font-weight: 600;
  letter-spacing: -0.02em;
  line-height: 1.05;
}

.price-row { display: flex; align-items: baseline; gap: 8px; }
.price {
  font-family: var(--font-mono);
  font-size: 30px;
  font-weight: 400;
  color: var(--text);
  letter-spacing: -0.01em;
}
.price-unit { color: var(--text-dim); font-family: var(--font-mono); font-size: 12px; letter-spacing: 0.12em; }
.price-flag { font-size: 16px; color: var(--sale); margin-left: 12px; }

.desc {
  font-size: 14.5px;
  line-height: 1.65;
  color: var(--text-soft);
  white-space: pre-line;
  max-width: 52ch;
}

.size-block { display: flex; flex-direction: column; gap: 12px; }
.size-head { display: flex; align-items: baseline; gap: 8px; }
.size-pick { color: var(--text-soft); font-size: 13px; }
.sizes { display: flex; flex-wrap: wrap; gap: 6px; }
.size {
  position: relative;
  min-width: 50px; height: 44px; padding: 0 12px;
  font: 500 12px/1 var(--font-mono);
  letter-spacing: 0.06em;
  color: var(--text-soft);
  background: transparent;
  border: 1px solid var(--border);
  cursor: pointer;
  transition: all .15s ease;
}
.size:hover:not(:disabled) { border-color: var(--text-muted); color: var(--text); }
.size.selected {
  border-color: var(--accent);
  color: var(--accent);
  background: var(--accent-haze);
  box-shadow: 0 0 0 1px var(--accent), 0 0 18px -4px var(--accent-glow);
}
.size.gone { color: var(--text-dim); cursor: not-allowed; }
.size .strike {
  position: absolute; inset: 0;
  background: linear-gradient(to top right, transparent calc(50% - 1px), var(--text-dim) 50%, transparent calc(50% + 1px));
  pointer-events: none;
}
.no-sizes { color: var(--text-muted); font-size: 13px; }

.actions { display: flex; gap: 12px; align-items: stretch; }
.cta { flex: 1; height: 56px; font-size: 13px; letter-spacing: 0.08em; }
.cta-arrow { font-family: var(--font-mono); margin-left: auto; }
.wish-btn {
  width: 56px; height: 56px;
  background: transparent;
  border: 1px solid var(--border-strong);
  color: var(--text-muted);
  display: inline-flex; align-items: center; justify-content: center;
  cursor: pointer;
  transition: all .15s ease;
}
.wish-btn:hover { color: var(--text); border-color: var(--text); }
.wish-btn.active { color: var(--sale); border-color: var(--sale); }

.bullets {
  list-style: none;
  margin-top: 12px;
  padding-top: 24px;
  border-top: 1px dashed var(--border);
  display: flex; flex-direction: column; gap: 10px;
}
.bullets li { display: flex; justify-content: space-between; font-size: 13px; }
.b-key { font-family: var(--font-mono); font-size: 11px; letter-spacing: 0.14em; text-transform: uppercase; color: var(--text-muted); }
.b-val { color: var(--text-soft); }

@media (max-width: 900px) {
  .page { padding: 32px 16px 0; }
  .layout { grid-template-columns: 1fr; gap: 32px; }
  .gallery { grid-template-columns: 1fr; }
  .thumbnails { flex-direction: row; max-height: none; overflow-x: auto; }
  .thumb { width: 64px; flex-shrink: 0; }
  .info { position: static; }
  .price { font-size: 24px; }
}
</style>

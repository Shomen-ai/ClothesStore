<template>
  <RouterLink :to="`/catalogue/${product.id}`" class="card" :class="{ 'has-sale': product.is_on_sale }">
    <div class="frame">
      <img v-if="primaryImage" :src="primaryImage" :alt="product.name" loading="lazy" />
      <img v-if="hoverImage" :src="hoverImage" class="hover-image" loading="lazy" />
      <div v-if="!primaryImage" class="no-image">
        <span class="serif-italic">no image</span>
      </div>

      <span v-if="product.is_on_sale" class="sale-tag">
        <span class="sale-dot"></span>
        on sale
      </span>

      <button class="wish-btn" :class="{ active: inWishlist }" @click.prevent="toggleWish" aria-label="В избранное">
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
          <path d="M8 14s-5-3.2-5-7.2A2.8 2.8 0 0 1 8 4.6a2.8 2.8 0 0 1 5 2.2C13 10.8 8 14 8 14Z"
                stroke="currentColor" stroke-width="1.4"
                :fill="inWishlist ? 'currentColor' : 'none'" />
        </svg>
      </button>
    </div>

    <div class="meta">
      <p class="name">{{ product.name }}</p>
      <p class="price">
        {{ formatPrice(product.price) }}
        <span class="price-unit">RUB</span>
      </p>
    </div>
  </RouterLink>
</template>

<script setup>
import { computed } from 'vue'
import { useWishlistStore } from '@/stores/wishlist.js'

const props = defineProps({ product: Object })
const wishlist = useWishlistStore()

const primaryImage = computed(() =>
  props.product.images?.find(i => i.is_primary)?.image_path ||
  props.product.images?.[0]?.image_path
)
const hoverImage = computed(() => props.product.images?.[1]?.image_path)
const inWishlist = computed(() => wishlist.has(props.product.id))

async function toggleWish() { await wishlist.toggle(props.product.id) }

function formatPrice(p) {
  return new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 0 }).format(p)
}
</script>

<style scoped>
.card {
  display: flex; flex-direction: column; gap: 14px;
  color: var(--text);
  transition: transform .3s ease;
}
.frame {
  position: relative;
  aspect-ratio: 4/5;
  background: linear-gradient(135deg, rgba(255,255,255,0.04), rgba(255,255,255,0.02));
  border: 1px solid var(--border);
  overflow: hidden;
}
.frame::after {
  content: '';
  position: absolute; inset: 0;
  background: linear-gradient(180deg, transparent 60%, rgba(8,8,12,0.55));
  pointer-events: none;
  opacity: 0;
  transition: opacity .4s ease;
}
.card:hover .frame::after { opacity: 1; }

.frame img {
  width: 100%; height: 100%; object-fit: cover;
  transition: transform .8s cubic-bezier(.2,.6,.2,1), opacity .35s ease;
}
.hover-image {
  position: absolute; inset: 0;
  opacity: 0;
}
.card:hover .frame > img:first-of-type { transform: scale(1.04); opacity: 0; }
.card:hover .hover-image { opacity: 1; transform: scale(1.04); }

.no-image {
  width: 100%; height: 100%;
  display: flex; align-items: center; justify-content: center;
  color: var(--text-dim);
  font-size: 18px;
}

.sale-tag {
  position: absolute; top: 12px; left: 12px;
  z-index: 2;
  display: inline-flex; align-items: center; gap: 6px;
  padding: 5px 10px 5px 8px;
  font-family: var(--font-mono);
  font-size: 10px;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--sale-soft);
  background: rgba(8,8,12,0.7);
  border: 1px solid var(--sale);
  backdrop-filter: blur(8px);
  box-shadow: 0 0 16px -4px var(--sale-glow);
}
.sale-tag .sale-dot {
  width: 5px; height: 5px; border-radius: 50%;
  background: var(--sale);
  box-shadow: 0 0 6px var(--sale-glow);
}

.wish-btn {
  position: absolute; top: 10px; right: 10px;
  z-index: 2;
  width: 32px; height: 32px; border-radius: 50%;
  background: rgba(8,8,12,0.55);
  backdrop-filter: blur(8px);
  border: 1px solid var(--border-strong);
  color: var(--text-muted);
  display: flex; align-items: center; justify-content: center;
  cursor: pointer;
  transition: all .2s ease;
}
.wish-btn:hover { color: var(--text); border-color: var(--text); transform: scale(1.06); }
.wish-btn.active { color: var(--sale); border-color: var(--sale); }

.meta {
  display: flex; justify-content: space-between; align-items: baseline;
  gap: 12px;
  padding: 0 2px;
}
.name {
  font-size: 13.5px;
  font-weight: 500;
  letter-spacing: 0.02em;
  color: var(--text);
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.price {
  font-family: var(--font-mono);
  font-size: 13px;
  font-weight: 400;
  color: var(--text-soft);
  flex-shrink: 0;
}
.price-unit {
  color: var(--text-dim);
  font-size: 10px;
  margin-left: 2px;
  letter-spacing: 0.1em;
}
</style>

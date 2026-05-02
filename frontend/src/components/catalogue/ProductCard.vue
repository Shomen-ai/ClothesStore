<template>
  <div class="product-card" @click="$router.push(`/catalogue/${product.id}`)">
    <div class="product-image">
      <img v-if="primaryImage" :src="primaryImage" :alt="product.name" />
      <div v-else class="no-image">NO IMAGE</div>
      <button class="wish-btn" @click.stop="toggleWish" :class="{ active: inWishlist }">
        <el-icon><Star /></el-icon>
      </button>
    </div>
    <div class="product-info">
      <p class="product-name">{{ product.name }}</p>
      <p class="product-price">{{ formatPrice(product.price) }}</p>
    </div>
  </div>
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
const inWishlist = computed(() => wishlist.has(props.product.id))

async function toggleWish() {
  await wishlist.toggle(props.product.id)
}

function formatPrice(p) {
  return new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 0 }).format(p)
}
</script>

<style scoped>
.product-card { cursor: pointer; }
.product-image { position: relative; aspect-ratio: 3/4; background: var(--color-bg-surface); border: 1px solid var(--color-border); overflow: hidden; }
.product-image img { width: 100%; height: 100%; object-fit: cover; transition: transform 0.3s; }
.product-card:hover .product-image img { transform: scale(1.03); }
.no-image { display: flex; align-items: center; justify-content: center; height: 100%; color: var(--color-text-muted); font-size: 12px; letter-spacing: 2px; }
.wish-btn { position: absolute; top: 10px; right: 10px; background: rgba(0,0,0,0.6); border: none; cursor: pointer; color: var(--color-text-muted); padding: 6px; border-radius: 4px; display: flex; }
.wish-btn.active { color: var(--color-red); }
.wish-btn:hover { color: var(--color-red); }
.product-info { padding: 12px 0; }
.product-name { font-size: 13px; color: var(--color-text); margin-bottom: 4px; }
.product-price { font-size: 14px; font-weight: 600; color: var(--color-text); }
</style>

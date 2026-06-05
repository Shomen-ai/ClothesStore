<!--
  WishlistView — the customer's saved products (route: /account/wishlist).
  Renders saved products as ProductCard grid; the cards read the wishlist store to show
  their toggled state, so the store is primed here in addition to fetching the full list.
-->
<template>
  <div>
    <h2 class="section-title">ИЗБРАННОЕ</h2>
    <div v-if="items.length === 0" class="empty">Список пуст</div>
    <div v-else class="product-grid">
      <ProductCard v-for="p in items" :key="p.id" :product="p" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getWishlist } from '@/api/user.js'
import { useWishlistStore } from '@/stores/wishlist.js'
import ProductCard from '@/components/catalogue/ProductCard.vue'

const items = ref([])
const wishlist = useWishlistStore()

onMounted(async () => {
  // Prime the store so each ProductCard's heart icon reflects saved state...
  await wishlist.load()
  // ...then fetch the full product objects to render the grid. Note: this hits the same
  // wishlist endpoint twice (store.load + getWishlist) — see review note.
  const { data } = await getWishlist()
  items.value = data || []
})
</script>

<style scoped>
.section-title { font-size: 20px; font-weight: 700; letter-spacing: 3px; margin-bottom: 32px; }
.empty { color: var(--color-text-muted); }
.product-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 24px; }
</style>

<template>
  <div class="catalogue-page">
    <div class="catalogue-layout">
      <aside class="sidebar">
        <ProductFilters
          :categories="categories"
          :model-category="filters.category"
          :model-size="filters.size"
          :model-sort="filters.sort"
          @update:category="v => { filters.category = v; load() }"
          @update:size="v => { filters.size = v; load() }"
          @update:sort="v => { filters.sort = v; load() }"
        />
      </aside>
      <div class="catalogue-main">
        <div class="search-row">
          <el-input v-model="filters.q" placeholder="Поиск..." clearable @input="debouncedLoad">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
        </div>
        <div v-if="loading" class="loading-state">Загрузка...</div>
        <div v-else-if="products.length === 0" class="empty-state">Ничего не найдено</div>
        <div v-else class="product-grid">
          <ProductCard v-for="p in products" :key="p.id" :product="p" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getCategories, getProducts } from '@/api/catalogue.js'
import { useWishlistStore } from '@/stores/wishlist.js'
import ProductCard from '@/components/catalogue/ProductCard.vue'
import ProductFilters from '@/components/catalogue/ProductFilters.vue'

const products = ref([])
const categories = ref([])
const loading = ref(false)
const filters = ref({ category: null, size: '', sort: '', q: '' })
const wishlist = useWishlistStore()

let debounceTimer = null
function debouncedLoad() {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(load, 400)
}

async function load() {
  loading.value = true
  const params = { page: 1 }
  if (filters.value.category) params.category = filters.value.category
  if (filters.value.size) params.size = filters.value.size
  if (filters.value.sort) params.sort = filters.value.sort
  if (filters.value.q) params.q = filters.value.q
  const { data } = await getProducts(params)
  products.value = data || []
  loading.value = false
}

onMounted(async () => {
  wishlist.load()
  const { data } = await getCategories()
  categories.value = data || []
  load()
})
</script>

<style scoped>
.catalogue-page { padding: 40px 24px; max-width: 1400px; margin: 0 auto; }
.catalogue-layout { display: grid; grid-template-columns: 240px 1fr; gap: 40px; }
.sidebar { border-right: 1px solid var(--color-border); padding-right: 32px; }
.search-row { margin-bottom: 24px; }
.product-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(260px, 1fr)); gap: 24px; }
.loading-state, .empty-state { color: var(--color-text-muted); padding: 80px 0; text-align: center; }
</style>

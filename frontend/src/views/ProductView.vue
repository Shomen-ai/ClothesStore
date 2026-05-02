<template>
  <div class="product-page" v-if="product">
    <div class="product-layout">
      <div class="gallery">
        <div class="main-image">
          <img v-if="activeImage" :src="activeImage" :alt="product.name" />
          <div v-else class="no-image">NO IMAGE</div>
        </div>
        <div class="thumbnails" v-if="product.images?.length > 1">
          <img v-for="img in product.images" :key="img.id" :src="img.image_path"
               :class="{ active: activeImage === img.image_path }"
               @click="activeImage = img.image_path" />
        </div>
      </div>

      <div class="product-info">
        <h1 class="product-name">{{ product.name }}</h1>
        <p class="product-price">{{ formatPrice(product.price) }}</p>
        <p class="product-desc">{{ product.description }}</p>

        <div class="size-section">
          <p class="size-label">РАЗМЕР</p>
          <div class="sizes">
            <button v-for="s in product.sizes" :key="s.id"
                    class="size-btn" :class="{ selected: selectedSize?.id === s.id, disabled: s.stock_qty === 0 }"
                    :disabled="s.stock_qty === 0" @click="selectedSize = s">
              {{ s.size }}
            </button>
          </div>
        </div>

        <div class="actions">
          <el-button type="primary" size="large" :disabled="!selectedSize" @click="addToCart" style="flex:1;letter-spacing:1px">
            В КОРЗИНУ
          </el-button>
          <button class="wish-btn" :class="{ active: inWishlist }" @click="wishlist.toggle(product.id)">
            <el-icon size="20"><Star /></el-icon>
          </button>
        </div>
      </div>
    </div>
  </div>
  <div v-else-if="loading" class="loading-state">Загрузка...</div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getProduct } from '@/api/catalogue.js'
import { useCartStore } from '@/stores/cart.js'
import { useWishlistStore } from '@/stores/wishlist.js'
import { ElMessage } from 'element-plus'

const route = useRoute()
const cart = useCartStore()
const wishlist = useWishlistStore()

const product = ref(null)
const loading = ref(true)
const selectedSize = ref(null)
const activeImage = ref('')

const inWishlist = computed(() => wishlist.has(product.value?.id))

onMounted(async () => {
  wishlist.load()
  const { data } = await getProduct(route.params.id)
  product.value = data
  activeImage.value = data.images?.find(i => i.is_primary)?.image_path || data.images?.[0]?.image_path || ''
  loading.value = false
})

function addToCart() {
  cart.add(product.value, selectedSize.value.id, selectedSize.value.size)
  ElMessage({ message: 'Добавлено в корзину', type: 'success', duration: 1500 })
}

function formatPrice(p) {
  return new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 0 }).format(p)
}
</script>

<style scoped>
.product-page { max-width: 1200px; margin: 40px auto; padding: 0 24px; }
.product-layout { display: grid; grid-template-columns: 1fr 1fr; gap: 64px; }
.main-image { aspect-ratio: 3/4; background: var(--color-bg-surface); border: 1px solid var(--color-border); overflow: hidden; margin-bottom: 12px; display: flex; align-items: center; justify-content: center; }
.main-image img { width: 100%; height: 100%; object-fit: cover; }
.no-image { color: var(--color-text-muted); font-size: 12px; letter-spacing: 2px; }
.thumbnails { display: flex; gap: 8px; }
.thumbnails img { width: 72px; height: 96px; object-fit: cover; border: 1px solid var(--color-border); cursor: pointer; opacity: 0.7; }
.thumbnails img.active { opacity: 1; border-color: var(--color-text); }
.product-name { font-size: 28px; font-weight: 700; letter-spacing: 2px; margin-bottom: 16px; }
.product-price { font-size: 24px; font-weight: 600; margin-bottom: 24px; }
.product-desc { color: var(--color-text-muted); line-height: 1.7; margin-bottom: 32px; }
.size-label { font-size: 11px; letter-spacing: 2px; color: var(--color-text-muted); margin-bottom: 12px; }
.sizes { display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 32px; }
.size-btn { border: 1px solid var(--color-border); background: none; color: var(--color-text); padding: 8px 16px; cursor: pointer; font-size: 14px; }
.size-btn:hover:not(.disabled) { border-color: var(--color-text); }
.size-btn.selected { border-color: var(--color-red); color: var(--color-red); }
.size-btn.disabled { opacity: 0.3; cursor: not-allowed; }
.actions { display: flex; gap: 16px; align-items: center; }
.wish-btn { background: none; border: 1px solid var(--color-border); cursor: pointer; color: var(--color-text-muted); padding: 12px; display: flex; }
.wish-btn.active { color: var(--color-red); border-color: var(--color-red); }
.loading-state { max-width: 1200px; margin: 80px auto; padding: 0 24px; color: var(--color-text-muted); }
</style>

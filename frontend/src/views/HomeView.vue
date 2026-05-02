<template>
  <div>
    <section class="hero">
      <div class="hero-content">
        <h1 class="hero-title">НОВАЯ КОЛЛЕКЦИЯ</h1>
        <p class="hero-sub">Одежда для тех, кто выбирает стиль</p>
        <RouterLink to="/catalogue">
          <el-button type="primary" size="large" style="letter-spacing:2px">СМОТРЕТЬ</el-button>
        </RouterLink>
      </div>
    </section>

    <section class="section" v-if="hits.length">
      <div class="container">
        <h2 class="section-title">ХИТЫ ПРОДАЖ</h2>
        <div class="product-grid">
          <ProductCard v-for="p in hits" :key="p.id" :product="p" />
        </div>
      </div>
    </section>

    <section class="section" v-if="newest.length">
      <div class="container">
        <h2 class="section-title">НОВИНКИ</h2>
        <div class="product-grid">
          <ProductCard v-for="p in newest" :key="p.id" :product="p" />
        </div>
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
.hero {
  height: 80vh; min-height: 500px;
  background: linear-gradient(135deg, #0a0a0a 0%, #1a0505 50%, #0a0a0a 100%);
  display: flex; align-items: center; justify-content: center; text-align: center;
  border-bottom: 1px solid var(--color-red-dark);
}
.hero-title { font-size: clamp(36px, 8vw, 96px); font-weight: 900; letter-spacing: 8px; margin-bottom: 16px; }
.hero-sub { font-size: 16px; color: var(--color-text-muted); margin-bottom: 40px; letter-spacing: 2px; }
.section { padding: 80px 0; }
.container { max-width: 1400px; margin: 0 auto; padding: 0 24px; }
.section-title { font-size: 24px; font-weight: 800; letter-spacing: 4px; margin-bottom: 40px; }
.product-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 24px; }
</style>

<template>
  <div class="cart-item">
    <div class="item-image">
      <img v-if="item.image" :src="item.image" :alt="item.name" />
      <div v-else class="no-img" />
    </div>
    <div class="item-details">
      <p class="item-name">{{ item.name }}</p>
      <p class="item-size">Размер: {{ item.size }}</p>
    </div>
    <div class="item-qty">
      <el-input-number v-model="qty" :min="1" :max="99" size="small" @change="update" />
    </div>
    <p class="item-price">{{ formatPrice(item.price * item.quantity) }}</p>
    <button class="remove-btn" @click="emit('remove', item.product_size_id)">
      <el-icon><Close /></el-icon>
    </button>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
const props = defineProps({ item: Object })
const emit = defineEmits(['remove', 'update-qty'])
const qty = ref(props.item.quantity)
watch(() => props.item.quantity, v => { qty.value = v })
function update() { emit('update-qty', props.item.product_size_id, qty.value) }
function formatPrice(p) {
  return new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 0 }).format(p)
}
</script>

<style scoped>
.cart-item { display: flex; align-items: center; gap: 16px; padding: 16px 0; border-bottom: 1px solid var(--color-border); }
.item-image { width: 80px; height: 100px; background: var(--color-bg-surface); flex-shrink: 0; overflow: hidden; }
.item-image img { width: 100%; height: 100%; object-fit: cover; }
.no-img { width: 100%; height: 100%; }
.item-details { flex: 1; }
.item-name { font-size: 14px; margin-bottom: 4px; }
.item-size { font-size: 12px; color: var(--color-text-muted); }
.item-price { font-weight: 600; min-width: 100px; text-align: right; }
.remove-btn { background: none; border: none; cursor: pointer; color: var(--color-text-muted); display: flex; }
.remove-btn:hover { color: var(--color-red); }
</style>

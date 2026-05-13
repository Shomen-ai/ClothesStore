<template>
  <div class="cart-item">
    <RouterLink :to="`/catalogue/${item.product_id}`" class="thumb">
      <img v-if="item.image" :src="item.image" :alt="item.name" />
      <div v-else class="no-img"></div>
    </RouterLink>

    <div class="info">
      <RouterLink :to="`/catalogue/${item.product_id}`" class="name">{{ item.name }}</RouterLink>
      <p class="meta">
        <span class="meta-key">размер</span>
        <span class="meta-val">{{ item.size }}</span>
        <span class="meta-sep">·</span>
        <span class="meta-key">арт.</span>
        <span class="meta-val">#{{ String(item.product_id).padStart(4, '0') }}</span>
      </p>
      <p class="unit-price">{{ formatPrice(item.price) }} <span class="unit">/ шт</span></p>
    </div>

    <div class="qty">
      <button class="qty-btn" @click="dec" :disabled="qty <= 1" aria-label="-">−</button>
      <span class="qty-val">{{ qty }}</span>
      <button class="qty-btn" @click="inc" :disabled="qty >= 99" aria-label="+">+</button>
    </div>

    <div class="line-price">
      <span class="num">{{ formatPrice(item.price * item.quantity) }}</span>
    </div>

    <button class="remove" @click="emit('remove', item.product_size_id)" aria-label="Удалить">
      <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
        <path d="M3 3l8 8M11 3l-8 8" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/>
      </svg>
    </button>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
const props = defineProps({ item: Object })
const emit = defineEmits(['remove', 'update-qty'])
const qty = ref(props.item.quantity)

watch(() => props.item.quantity, v => { qty.value = v })

function inc() { if (qty.value < 99) { qty.value++; emit('update-qty', props.item.product_size_id, qty.value) } }
function dec() { if (qty.value > 1) { qty.value--; emit('update-qty', props.item.product_size_id, qty.value) } }

function formatPrice(p) {
  return new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 0, style: 'currency', currency: 'RUB' }).format(p)
}
</script>

<style scoped>
.cart-item {
  display: grid;
  grid-template-columns: 88px 1fr auto auto auto;
  align-items: center;
  gap: 20px;
  padding: 20px 0;
  border-bottom: 1px solid var(--border);
  transition: background .2s ease;
}
.cart-item:first-child { padding-top: 0; }
.cart-item:last-child { border-bottom: 0; }
.cart-item:hover { background: rgba(255,255,255,0.015); }

.thumb {
  display: block;
  width: 88px;
  aspect-ratio: 4/5;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  overflow: hidden;
}
.thumb img { width: 100%; height: 100%; object-fit: cover; transition: transform .4s ease; }
.thumb:hover img { transform: scale(1.05); }
.no-img { width: 100%; height: 100%; background: linear-gradient(135deg, rgba(255,255,255,0.04), rgba(255,255,255,0.01)); }

.info { display: flex; flex-direction: column; gap: 4px; min-width: 0; }
.name { font-size: 15px; color: var(--text); font-weight: 500; }
.name:hover { color: var(--accent-soft); }
.meta { display: flex; flex-wrap: wrap; gap: 5px; align-items: baseline; font-size: 11px; }
.meta-key { font-family: var(--font-mono); letter-spacing: 0.14em; text-transform: uppercase; color: var(--text-dim); }
.meta-val { color: var(--text-soft); }
.meta-sep { color: var(--text-dim); }
.unit-price { font-family: var(--font-mono); font-size: 12px; color: var(--text-muted); margin-top: 2px; }
.unit-price .unit { color: var(--text-dim); }

.qty {
  display: inline-flex; align-items: center;
  border: 1px solid var(--border-strong);
  height: 36px;
}
.qty-btn {
  width: 32px; height: 100%;
  background: transparent;
  border: 0;
  color: var(--text-soft);
  font: 400 16px/1 var(--font-mono);
  cursor: pointer;
  transition: color .15s ease;
}
.qty-btn:hover:not(:disabled) { color: var(--text); background: var(--bg-surface); }
.qty-btn:disabled { color: var(--text-dim); cursor: not-allowed; }
.qty-val { min-width: 32px; text-align: center; font: 500 13px/1 var(--font-mono); color: var(--text); }

.line-price { min-width: 110px; text-align: right; }
.line-price .num { font-family: var(--font-mono); font-size: 15px; color: var(--text); font-weight: 500; }

.remove {
  width: 32px; height: 32px;
  background: transparent;
  border: 1px solid transparent;
  color: var(--text-dim);
  display: inline-flex; align-items: center; justify-content: center;
  cursor: pointer;
  transition: all .15s ease;
}
.remove:hover { color: var(--sale); border-color: var(--sale); }

@media (max-width: 700px) {
  .cart-item {
    grid-template-columns: 72px 1fr auto;
    grid-template-areas:
      "thumb info remove"
      "thumb qty   price";
    gap: 8px 14px;
    padding: 16px 0;
  }
  .thumb { grid-area: thumb; width: 72px; }
  .info { grid-area: info; }
  .qty { grid-area: qty; }
  .line-price { grid-area: price; text-align: right; min-width: 0; }
  .remove { grid-area: remove; align-self: start; }
  .name { font-size: 14px; }
}
</style>

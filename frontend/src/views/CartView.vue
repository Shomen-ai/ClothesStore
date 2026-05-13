<template>
  <div class="page">
    <header class="page-head">
      <div class="head-marquee-wrap" aria-hidden="true">
        <div class="head-marquee">
          <div class="marquee-row row-1">
            <div class="marquee-track">
              <span>CHECKOUT · PACKED · SEALED · SHIP · DROP · RECEIPT · WRAPPED · LOUD ·&nbsp;</span>
              <span>CHECKOUT · PACKED · SEALED · SHIP · DROP · RECEIPT · WRAPPED · LOUD ·&nbsp;</span>
            </div>
          </div>
          <div class="marquee-row row-2">
            <div class="marquee-track">
              <span>к оплате — close to home — final — твоё — почти готово — secured —&nbsp;</span>
              <span>к оплате — close to home — final — твоё — почти готово — secured —&nbsp;</span>
            </div>
          </div>
          <div class="marquee-row row-3">
            <div class="marquee-track">
              <span>FITS · RITUAL · DELIVER · SIGNATURE · KEPT · PROMO · YOURS ·&nbsp;</span>
              <span>FITS · RITUAL · DELIVER · SIGNATURE · KEPT · PROMO · YOURS ·&nbsp;</span>
            </div>
          </div>
        </div>
      </div>

      <p class="eyebrow">Корзина</p>
      <h1 class="title">
        Уже почти <span class="gothic-accent">твоё</span>
      </h1>
      <p class="count"><span>{{ cart.items.length }}</span> позиций</p>
    </header>

    <div v-if="cart.items.length === 0" class="empty">
      <p class="empty-line">
        <span class="gothic-accent">пусто</span>
        <br />
        <span class="serif-italic">но не страшно</span>
      </p>
      <p class="empty-sub">Загляни в каталог — найдём что-нибудь под настроение.</p>
      <RouterLink to="/catalogue" class="btn btn-primary empty-cta">
        <span>В каталог</span><span class="arrow">→</span>
      </RouterLink>
    </div>

    <div v-else class="layout">
      <section class="items">
        <CartItem v-for="item in cart.items" :key="item.product_size_id" :item="item"
          @remove="cart.remove" @update-qty="cart.updateQty" />
      </section>

      <aside class="summary">
        <h2 class="sum-title">Итого</h2>

        <div class="sum-row">
          <span>Товары</span><span class="num">{{ formatPrice(cart.total) }}</span>
        </div>
        <div v-if="cart.discount > 0" class="sum-row discount">
          <span>
            Промокод
            <span class="code">{{ cart.promoCode }}</span>
          </span>
          <span class="num">−{{ formatPrice(cart.discount) }}</span>
        </div>
        <div class="sum-row total">
          <span>К оплате</span><span class="num big">{{ formatPrice(cart.finalTotal) }}</span>
        </div>

        <div class="promo">
          <input v-model="promoInput" placeholder="Промокод" :disabled="cart.discount > 0" />
          <button class="apply" :disabled="!promoInput || cart.discount > 0 || promoLoading" @click="applyPromo">
            {{ promoLoading ? '…' : 'Применить' }}
          </button>
        </div>
        <p v-if="promoError" class="promo-error">{{ promoError }}</p>

        <div v-if="auth.isLoggedIn" class="address">
          <p class="address-label eyebrow">Адрес доставки</p>
          <select v-model="selectedAddressID" class="address-select">
            <option :value="null" disabled>Выберите адрес</option>
            <option v-for="a in addresses" :key="a.id" :value="a.id">
              {{ a.city }}, {{ a.street }} {{ a.house }}{{ a.apartment ? `, кв.${a.apartment}` : '' }}
            </option>
          </select>
        </div>

        <button class="btn btn-primary checkout" :disabled="!canOrder || ordering" @click="placeOrder">
          <span>{{ ordering ? 'Оформляем…' : 'Оформить заказ' }}</span>
          <span class="arrow">→</span>
        </button>
        <p v-if="!auth.isLoggedIn" class="login-hint">
          <RouterLink to="/login">Войдите</RouterLink>, чтобы оформить заказ
        </p>
      </aside>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useCartStore } from '@/stores/cart.js'
import { useAuthStore } from '@/stores/auth.js'
import { getAddresses } from '@/api/user.js'
import { validatePromo as apiValidatePromo, createOrder } from '@/api/orders.js'
import CartItem from '@/components/cart/CartItem.vue'

const cart = useCartStore()
const auth = useAuthStore()
const router = useRouter()

const addresses = ref([])
const selectedAddressID = ref(null)
const promoInput = ref('')
const promoError = ref('')
const promoLoading = ref(false)
const ordering = ref(false)

const canOrder = computed(() => auth.isLoggedIn && selectedAddressID.value && cart.items.length > 0)

onMounted(async () => {
  if (auth.isLoggedIn) {
    const { data } = await getAddresses()
    addresses.value = data || []
    const def = addresses.value.find(a => a.is_default)
    if (def) selectedAddressID.value = def.id
  }
})

async function applyPromo() {
  promoError.value = ''
  promoLoading.value = true
  try {
    const { data } = await apiValidatePromo(promoInput.value, cart.total)
    cart.applyPromo(promoInput.value, data.discount_amount)
  } catch (e) {
    promoError.value = e.response?.data?.error || 'Промокод недействителен'
  } finally {
    promoLoading.value = false
  }
}

async function placeOrder() {
  ordering.value = true
  try {
    await createOrder({
      address_id: selectedAddressID.value,
      promo_code: cart.promoCode || undefined,
      items: cart.items.map(i => ({ product_size_id: i.product_size_id, quantity: i.quantity }))
    })
    cart.clear()
    router.push('/checkout-success')
  } catch (e) {
    console.error(e)
  } finally {
    ordering.value = false
  }
}

function formatPrice(p) {
  return new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 0, style: 'currency', currency: 'RUB' }).format(p)
}
</script>

<style scoped>
.page { max-width: 1440px; margin: 0 auto; padding: 80px 32px 0; }

.page-head {
  position: relative;
  display: flex; flex-direction: column; gap: 8px;
  margin-bottom: 48px;
  padding: 0 0 24px;
  min-height: 180px;
  border-bottom: 1px solid var(--border);
  overflow: visible;
  isolation: isolate;
}
.page-head .eyebrow, .page-head .title, .page-head .count { position: relative; z-index: 2; }
.title {
  font-size: clamp(36px, 5vw, 64px);
  font-weight: 700;
  letter-spacing: -0.03em;
  line-height: 1;
}
.title .gothic-accent { font-size: 1.05em; }
.count { color: var(--text-muted); font-family: var(--font-mono); font-size: 12px; letter-spacing: 0.12em; text-transform: uppercase; }
.count span { color: var(--text); font-size: 14px; margin-right: 4px; }

.empty {
  display: flex; flex-direction: column; align-items: center; gap: 12px;
  padding: 120px 0;
  text-align: center;
}
.empty-line { font-size: 40px; color: var(--text-soft); }
.empty-sub { color: var(--text-muted); max-width: 36ch; }
.empty-cta { margin-top: 16px; }
.empty-cta .arrow { font-family: var(--font-mono); }

.layout { display: grid; grid-template-columns: 1fr 360px; gap: 64px; align-items: start; }

.items { display: flex; flex-direction: column; }

/* Summary */
.summary {
  position: sticky; top: 96px;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  padding: 32px;
  display: flex; flex-direction: column; gap: 14px;
  backdrop-filter: blur(12px);
}
.sum-title {
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--text-muted);
  margin-bottom: 6px;
  padding-bottom: 12px;
  border-bottom: 1px dashed var(--border);
}
.sum-row { display: flex; justify-content: space-between; align-items: baseline; font-size: 14px; color: var(--text-soft); }
.sum-row .num { font-family: var(--font-mono); color: var(--text); }
.sum-row.discount { color: var(--sale-soft); }
.sum-row.discount .code { font-family: var(--font-mono); font-size: 11px; color: var(--sale); margin-left: 4px; padding: 2px 5px; background: rgba(255, 120, 73, 0.10); }
.sum-row.discount .num { color: var(--sale); }
.sum-row.total { padding-top: 12px; border-top: 1px dashed var(--border); color: var(--text); }
.sum-row.total .num.big { font-size: 22px; }

.promo { display: flex; gap: 0; margin-top: 12px; }
.promo input {
  flex: 1; background: transparent;
  border: 1px solid var(--border-strong); border-right: 0;
  color: var(--text);
  padding: 10px 12px;
  font: 500 12px/1 var(--font-mono);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  outline: none;
}
.promo input::placeholder { color: var(--text-dim); }
.promo input:focus { border-color: var(--accent); }
.promo input:disabled { opacity: 0.4; }
.apply {
  background: var(--text); color: #0a0a0c;
  border: 1px solid var(--text);
  padding: 0 16px;
  font: 600 11px/1 var(--font-mono);
  letter-spacing: 0.08em;
  text-transform: uppercase;
  cursor: pointer;
}
.apply:hover:not(:disabled) { background: var(--accent); border-color: var(--accent); color: #fff; }
.apply:disabled { opacity: 0.3; cursor: not-allowed; }
.promo-error { color: var(--sale); font-size: 12px; }

.address { display: flex; flex-direction: column; gap: 8px; padding-top: 12px; border-top: 1px dashed var(--border); }
.address-label { margin: 0; }
.address-select {
  background: transparent;
  border: 1px solid var(--border-strong);
  color: var(--text);
  padding: 10px 12px;
  font-family: inherit;
  font-size: 13px;
  outline: none;
}
.address-select:focus { border-color: var(--accent); }

.checkout { margin-top: 12px; height: 56px; }
.checkout .arrow { font-family: var(--font-mono); margin-left: auto; }

.login-hint { text-align: center; font-size: 13px; color: var(--text-muted); }
.login-hint a { color: var(--accent-soft); text-decoration: underline; text-decoration-color: var(--accent-haze); }

@media (max-width: 900px) {
  .page { padding: 32px 16px 0; }
  .layout { grid-template-columns: 1fr; gap: 24px; }
  .summary { position: static; padding: 20px; }
}
</style>

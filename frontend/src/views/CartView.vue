<!--
  CartView - shopping-cart page (route: /cart).
  Lists cart line items, applies a promo code, lets a logged-in user pick a
  delivery address, then creates a pending order and hands off to /payment.
-->
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
        <div class="sum-row">
          <span>Доставка</span><span class="num">{{ deliveryCost ? formatPrice(deliveryCost) : 'бесплатно' }}</span>
        </div>
        <div class="sum-row total">
          <span>К оплате</span><span class="num big">{{ formatPrice(grandTotal) }}</span>
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

        <div v-if="auth.isLoggedIn" class="address">
          <p class="address-label eyebrow">Способ доставки</p>
          <select v-model="deliveryMethod" class="address-select">
            <option v-for="d in DELIVERY" :key="d.value" :value="d.value">
              {{ d.label }} — {{ d.cost ? d.cost + ' ₽' : 'бесплатно' }}
            </option>
          </select>
        </div>

        <div v-if="auth.isLoggedIn" class="address">
          <p class="address-label eyebrow">Способ оплаты</p>
          <select v-model="paymentMethod" class="address-select">
            <option v-for="p in PAYMENTS" :key="p.value" :value="p.value">{{ p.label }}</option>
          </select>
        </div>

        <div v-if="auth.isLoggedIn" class="address">
          <p class="address-label eyebrow">Получатель</p>
          <input v-model="recipient" class="address-select" placeholder="ФИО получателя" />
        </div>

        <button class="btn btn-primary checkout" :disabled="!canOrder || ordering" @click="placeOrder">
          <span>{{ ordering ? 'Оформляем…' : 'Оформить заказ' }}</span>
          <span class="arrow">→</span>
        </button>
        <p v-if="!auth.isLoggedIn" class="login-hint">
          <RouterLink to="/login">Войдите</RouterLink>
          или <RouterLink to="/register" class="register-link">зарегистрируйтесь</RouterLink>,
          чтобы оформить заказ
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
import { getAddresses, getProfile } from '@/api/user.js'
import { validatePromo as apiValidatePromo, createOrder } from '@/api/orders.js'
import CartItem from '@/components/cart/CartItem.vue'

// Cart state (items, totals, promo) lives in the Pinia cart store (persisted to
// localStorage); auth gates checkout and address loading.
const cart = useCartStore()
const auth = useAuthStore()
const router = useRouter()

const addresses = ref([])
const selectedAddressID = ref(null)
const promoInput = ref('')
const promoError = ref('')
const promoLoading = ref(false)
const ordering = ref(false)

// Delivery and payment options. Delivery costs mirror the server (which is the
// source of truth for the charged amount); these drive the UI preview + total.
const DELIVERY = [
  { value: 'courier', label: 'Курьерская доставка', cost: 500 },
  { value: 'post',    label: 'Почта России',        cost: 350 },
  { value: 'pickup',  label: 'Самовывоз',           cost: 0 },
]
const PAYMENTS = [
  { value: 'card_online', label: 'Картой онлайн' },
  { value: 'on_delivery', label: 'При получении' },
]
const deliveryMethod = ref('courier')
const paymentMethod = ref('card_online')
const recipient = ref('')

const deliveryCost = computed(() => DELIVERY.find(d => d.value === deliveryMethod.value)?.cost || 0)
const grandTotal = computed(() => cart.finalTotal + deliveryCost.value)

// Checkout requires login, an address, delivery+payment method, a recipient, and a non-empty cart.
const canOrder = computed(() =>
  auth.isLoggedIn && selectedAddressID.value && deliveryMethod.value &&
  paymentMethod.value && recipient.value.trim() && cart.items.length > 0)

// Preload saved addresses (preselecting the default) and prefill the recipient from the profile.
onMounted(async () => {
  if (!auth.isLoggedIn) return
  const { data } = await getAddresses()
  addresses.value = data || []
  const def = addresses.value.find(a => a.is_default)
  if (def) selectedAddressID.value = def.id
  try {
    const prof = await getProfile()
    recipient.value = prof.data?.name || ''
  } catch { /* profile is optional for prefilling */ }
})

// Validate the promo against the current subtotal on the server; on success the
// returned discount amount is stored in the cart store and applied to finalTotal.
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

// Create the order from cart contents + chosen address/promo, clear the cart,
// then route to the payment stub. The server recomputes prices server-side.
async function placeOrder() {
  ordering.value = true
  try {
    const { data } = await createOrder({
      address_id: selectedAddressID.value,
      promo_code: cart.promoCode || undefined,
      delivery_method: deliveryMethod.value,
      payment_method: paymentMethod.value,
      recipient_name: recipient.value.trim(),
      items: cart.items.map(i => ({ product_size_id: i.product_size_id, quantity: i.quantity }))
    })
    cart.clear()
    // Card payments go through the payment stub; pay-on-delivery orders are already
    // confirmed server-side, so jump straight to the order page.
    if (paymentMethod.value === 'card_online') router.push(`/payment/${data.id}`)
    else router.push(`/account/orders/${data.id}`)
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

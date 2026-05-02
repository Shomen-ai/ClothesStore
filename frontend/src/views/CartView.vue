<template>
  <div class="cart-page">
    <h1 class="page-title">КОРЗИНА</h1>
    <div v-if="cart.items.length === 0" class="empty-cart">
      <p>Корзина пуста</p>
      <RouterLink to="/catalogue"><el-button>Перейти в каталог</el-button></RouterLink>
    </div>
    <div v-else class="cart-layout">
      <div class="cart-items">
        <CartItem v-for="item in cart.items" :key="item.product_size_id" :item="item"
          @remove="cart.remove" @update-qty="cart.updateQty" />
      </div>
      <div class="cart-summary">
        <h3 class="summary-title">ИТОГО</h3>
        <div class="summary-row">
          <span>Товары</span><span>{{ formatPrice(cart.total) }}</span>
        </div>
        <div v-if="cart.discount > 0" class="summary-row discount">
          <span>Скидка ({{ cart.promoCode }})</span>
          <span>-{{ formatPrice(cart.discount) }}</span>
        </div>
        <div class="summary-row total">
          <span>К оплате</span><span>{{ formatPrice(cart.finalTotal) }}</span>
        </div>

        <div class="promo-section">
          <el-input v-model="promoInput" placeholder="Промокод" :disabled="cart.discount > 0">
            <template #append>
              <el-button @click="applyPromo" :loading="promoLoading">Применить</el-button>
            </template>
          </el-input>
          <el-alert v-if="promoError" :title="promoError" type="error" :closable="false" style="margin-top:8px" />
        </div>

        <div class="address-section" v-if="auth.isLoggedIn">
          <p class="section-label">АДРЕС ДОСТАВКИ</p>
          <el-select v-model="selectedAddressID" placeholder="Выберите адрес" style="width:100%">
            <el-option v-for="a in addresses" :key="a.id" :value="a.id"
              :label="`${a.city}, ${a.street} ${a.house}${a.apartment ? ', кв.' + a.apartment : ''}`" />
          </el-select>
        </div>

        <el-button type="primary" size="large" style="width:100%;margin-top:24px;letter-spacing:2px"
          :disabled="!canOrder" :loading="ordering" @click="placeOrder">
          ОФОРМИТЬ ЗАКАЗ
        </el-button>
        <p v-if="!auth.isLoggedIn" class="login-hint">
          <RouterLink to="/login">Войдите</RouterLink>, чтобы оформить заказ
        </p>
      </div>
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
  return new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 0 }).format(p)
}
</script>

<style scoped>
.cart-page { max-width: 1200px; margin: 40px auto; padding: 0 24px; }
.page-title { font-size: 28px; font-weight: 800; letter-spacing: 4px; margin-bottom: 40px; }
.empty-cart { text-align: center; padding: 80px 0; color: var(--color-text-muted); }
.empty-cart p { margin-bottom: 24px; font-size: 18px; }
.cart-layout { display: grid; grid-template-columns: 1fr 360px; gap: 48px; align-items: start; }
.cart-summary { background: var(--color-bg-surface); border: 1px solid var(--color-border); padding: 32px; position: sticky; top: 80px; }
.summary-title { font-size: 16px; font-weight: 700; letter-spacing: 2px; margin-bottom: 24px; }
.summary-row { display: flex; justify-content: space-between; margin-bottom: 12px; font-size: 14px; }
.summary-row.discount { color: var(--color-red); }
.summary-row.total { font-size: 18px; font-weight: 700; border-top: 1px solid var(--color-border); padding-top: 12px; margin-top: 12px; }
.promo-section { margin: 24px 0; }
.address-section { margin-top: 16px; }
.section-label { font-size: 11px; letter-spacing: 2px; color: var(--color-text-muted); margin-bottom: 8px; }
.login-hint { margin-top: 12px; text-align: center; font-size: 13px; color: var(--color-text-muted); }
</style>

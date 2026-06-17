<!--
  OrderDetailView — a single customer order (route: /account/orders/:id).
  Detailed card: header (number/date/status/payment), the goods, payment method
  and the delivery block (type, address, recipient, shipping cost).
-->
<template>
  <div v-if="order" class="order-card">
    <header class="card-head">
      <div>
        <h2 class="order-no">Заказ №{{ order.id }} <span class="order-date">от {{ formatDate(order.created_at) }}</span></h2>
        <p class="order-sum">Сумма: <strong>{{ formatPrice(order.total_price) }}</strong></p>
      </div>
      <div class="tags">
        <el-tag :type="statusType(order.status)" size="large">{{ statusLabel(order.status) }}</el-tag>
        <el-tag :type="payType(order.payment_status)" size="large" effect="plain">{{ payLabel(order.payment_status) }}</el-tag>
      </div>
    </header>

    <div class="card-body">
      <!-- Goods -->
      <section class="goods">
        <h3 class="block-title">Товар</h3>
        <div v-for="item in order.items" :key="item.id" class="goods-item">
          <img v-if="item.image_path" :src="item.image_path" :alt="item.product_name" class="goods-img" />
          <div v-else class="goods-img placeholder" />
          <div class="goods-info">
            <p class="goods-name">{{ itemTitle(item) }}</p>
            <p class="goods-art">Арт: {{ article(item) }}</p>
          </div>
          <div class="goods-qty">{{ item.quantity }} шт.</div>
          <div class="goods-price">{{ formatPrice(item.price_at_order) }}</div>
        </div>
      </section>

      <!-- Payment + delivery -->
      <aside class="meta">
        <div class="meta-block">
          <p class="meta-line"><span class="meta-key">Способ оплаты:</span> {{ payMethodLabel(order.payment_method) }}</p>
        </div>
        <div class="meta-block">
          <h3 class="block-title">Способ получения</h3>
          <p class="meta-line"><span class="meta-key">Тип:</span> {{ deliveryLabel(order.delivery_method) }}</p>
          <p class="meta-line"><span class="meta-key">Адрес доставки:</span> {{ addressLine }}</p>
          <p class="meta-line"><span class="meta-key">Получатель:</span> {{ order.recipient_name || '—' }}</p>
          <p class="meta-line"><span class="meta-key">Стоимость доставки:</span> {{ order.delivery_cost ? formatPrice(order.delivery_cost) : 'бесплатно' }}</p>
        </div>
      </aside>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getUserOrder } from '@/api/user.js'

const route = useRoute()
const order = ref(null)
onMounted(async () => { const { data } = await getUserOrder(route.params.id); order.value = data })

const statusMap = { pending:'Ожидает', confirmed:'Подтверждён', shipped:'Отправлен', delivered:'Доставлен', cancelled:'Отменён' }
const typeMap = { pending:'info', confirmed:'warning', shipped:'', delivered:'success', cancelled:'danger' }
const payStatusMap = { unpaid:'Не оплачен', paid:'Оплачен', on_delivery:'Оплата при получении' }
const payStatusType = { unpaid:'danger', paid:'success', on_delivery:'info' }
const payMethodMap = { card_online:'Картой онлайн', on_delivery:'Наличный расчёт при получении' }
const deliveryMap = { courier:'Курьерская доставка', post:'Почта России', pickup:'Самовывоз' }

function statusLabel(s) { return statusMap[s] || s }
function statusType(s) { return typeMap[s] || '' }
function payLabel(s) { return payStatusMap[s] || '—' }
function payType(s) { return payStatusType[s] || 'info' }
function payMethodLabel(s) { return payMethodMap[s] || '—' }
function deliveryLabel(s) { return deliveryMap[s] || '—' }

// "Тип Название (размер)" for an order line.
function itemTitle(i) {
  const base = [i.type_name, i.product_name].filter(Boolean).join(' ') || `Товар #${i.product_id}`
  return i.size ? `${base} (${i.size})` : base
}
function article(i) {
  return String(i.product_id).padStart(4, '0') + (i.size ? ` ${i.size}` : '')
}

const addressLine = computed(() => {
  const a = order.value?.address
  if (!a) return '—'
  return `${a.city}, ${a.street} ${a.house}${a.apartment ? `, кв. ${a.apartment}` : ''}`
})

function formatDate(d) { return new Date(d).toLocaleDateString('ru-RU') }
function formatPrice(p) { return new Intl.NumberFormat('ru-RU',{style:'currency',currency:'RUB',maximumFractionDigits:0}).format(p) }
</script>

<style scoped>
.order-card { border: 1px solid var(--color-border); }
.card-head {
  display: flex; justify-content: space-between; align-items: flex-start; gap: 16px;
  padding: 24px; border-bottom: 1px solid var(--color-border);
}
.order-no { font-size: 22px; font-weight: 700; letter-spacing: 0.5px; }
.order-date { font-size: 14px; font-weight: 400; color: var(--color-text-muted); }
.order-sum { margin-top: 8px; font-size: 15px; color: var(--color-text-muted); }
.order-sum strong { color: var(--color-text); font-size: 18px; }
.tags { display: flex; flex-direction: column; gap: 8px; align-items: flex-end; }

.card-body { display: grid; grid-template-columns: 1.2fr 1fr; gap: 32px; padding: 24px; }
.block-title { font-size: 13px; letter-spacing: 1.5px; text-transform: uppercase; color: var(--color-text-muted); margin-bottom: 16px; }

.goods-item { display: grid; grid-template-columns: 64px 1fr auto auto; gap: 16px; align-items: center; padding: 12px 0; border-bottom: 1px solid var(--color-border); }
.goods-img { width: 64px; height: 80px; object-fit: cover; }
.goods-img.placeholder { background: var(--color-bg-elevated); }
.goods-name { font-size: 14px; font-weight: 500; }
.goods-art { font-size: 12px; color: var(--color-text-muted); margin-top: 4px; }
.goods-qty { font-size: 13px; color: var(--color-text-muted); }
.goods-price { font-weight: 600; }

.meta-block { margin-bottom: 20px; }
.meta-line { font-size: 14px; line-height: 1.9; }
.meta-key { color: var(--color-text-muted); }

@media (max-width: 760px) {
  .card-body { grid-template-columns: 1fr; gap: 24px; }
  .card-head { flex-direction: column; }
  .tags { flex-direction: row; align-items: flex-start; }
}
</style>

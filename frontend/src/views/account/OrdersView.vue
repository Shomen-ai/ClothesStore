<!--
  OrdersView — list of the customer's orders (route: /account/orders).
  Each row links to its detail page; order status and payment status are shown as tags.
-->
<template>
  <div>
    <h2 class="section-title">МОИ ЗАКАЗЫ</h2>
    <div v-if="orders.length === 0" class="empty">Заказов пока нет</div>
    <div v-else class="orders-list">
      <div class="order-row head">
        <div>Номер заказа</div>
        <div>Дата заказа</div>
        <div>Сумма</div>
        <div>Статус заказа</div>
        <div>Статус платежа</div>
      </div>
      <div v-for="o in orders" :key="o.id" class="order-row" @click="$router.push(`/account/orders/${o.id}`)">
        <div><span class="order-id">№{{ o.id }}</span></div>
        <div class="order-date">{{ formatDate(o.created_at) }}</div>
        <div class="order-total">{{ formatPrice(o.total_price) }}</div>
        <div><el-tag :type="statusType(o.status)" size="small">{{ statusLabel(o.status) }}</el-tag></div>
        <div><el-tag :type="payType(o.payment_status)" size="small" effect="plain">{{ payLabel(o.payment_status) }}</el-tag></div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getUserOrders } from '@/api/user.js'

const orders = ref([])
onMounted(async () => { const { data } = await getUserOrders(); orders.value = data || [] })

const statusMap = { pending:'Ожидает', confirmed:'Подтверждён', shipped:'Отправлен', delivered:'Доставлен', cancelled:'Отменён' }
const typeMap = { pending:'info', confirmed:'warning', shipped:'', delivered:'success', cancelled:'danger' }
const payStatusMap = { unpaid:'Не оплачен', paid:'Оплачен', on_delivery:'При получении' }
const payStatusType = { unpaid:'danger', paid:'success', on_delivery:'info' }
function statusLabel(s) { return statusMap[s] || s }
function statusType(s) { return typeMap[s] || '' }
function payLabel(s) { return payStatusMap[s] || '—' }
function payType(s) { return payStatusType[s] || 'info' }
function formatDate(d) { return new Date(d).toLocaleDateString('ru-RU') }
function formatPrice(p) { return new Intl.NumberFormat('ru-RU',{style:'currency',currency:'RUB',maximumFractionDigits:0}).format(p) }
</script>

<style scoped>
.section-title { font-size: 20px; font-weight: 700; letter-spacing: 3px; margin-bottom: 32px; }
.empty { color: var(--color-text-muted); }
.order-row { display: grid; grid-template-columns: 110px 120px 120px 1fr 1fr; align-items: center; gap: 12px; padding: 16px; border: 1px solid var(--color-border); margin-bottom: 8px; cursor: pointer; }
.order-row:hover { border-color: var(--color-text-muted); }
.order-row.head { cursor: default; border: none; margin-bottom: 4px; padding: 0 16px; font-size: 11px; letter-spacing: 1px; text-transform: uppercase; color: var(--color-text-muted); }
.order-row.head:hover { border: none; }
.order-id { font-weight: 600; }
.order-date { color: var(--color-text-muted); font-size: 13px; }
.order-total { font-weight: 600; }

@media (max-width: 760px) {
  .order-row { grid-template-columns: 1fr 1fr; }
  .order-row.head { display: none; }
}
</style>

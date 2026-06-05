<!--
  OrdersView — list of the customer's orders (route: /account/orders).
  Each row links to its detail page; status is shown via a colored Element Plus tag.
-->
<template>
  <div>
    <h2 class="section-title">МОИ ЗАКАЗЫ</h2>
    <div v-if="orders.length === 0" class="empty">Заказов пока нет</div>
    <div v-else class="orders-list">
      <div v-for="o in orders" :key="o.id" class="order-row" @click="$router.push(`/account/orders/${o.id}`)">
        <div><span class="order-id">#{{ o.id }}</span></div>
        <div class="order-date">{{ formatDate(o.created_at) }}</div>
        <div><el-tag :type="statusType(o.status)" size="small">{{ statusLabel(o.status) }}</el-tag></div>
        <div class="order-total">{{ formatPrice(o.total_price) }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getUserOrders } from '@/api/user.js'

const orders = ref([])
// Load all of the current user's orders on mount.
onMounted(async () => { const { data } = await getUserOrders(); orders.value = data || [] })

// Status code -> Russian label and -> Element Plus tag color (duplicated across order views).
const statusMap = { pending:'Ожидает', confirmed:'Подтверждён', shipped:'Отправлен', delivered:'Доставлен', cancelled:'Отменён' }
const typeMap = { pending:'info', confirmed:'warning', shipped:'', delivered:'success', cancelled:'danger' }
function statusLabel(s) { return statusMap[s] || s }
function statusType(s) { return typeMap[s] || '' }
function formatDate(d) { return new Date(d).toLocaleDateString('ru-RU') }
function formatPrice(p) { return new Intl.NumberFormat('ru-RU',{style:'currency',currency:'RUB',maximumFractionDigits:0}).format(p) }
</script>

<style scoped>
.section-title { font-size: 20px; font-weight: 700; letter-spacing: 3px; margin-bottom: 32px; }
.empty { color: var(--color-text-muted); }
.order-row { display: grid; grid-template-columns: 80px 1fr 140px 140px; align-items: center; padding: 16px; border: 1px solid var(--color-border); margin-bottom: 8px; cursor: pointer; }
.order-row:hover { border-color: var(--color-text-muted); }
.order-id { font-weight: 600; }
.order-date { color: var(--color-text-muted); font-size: 13px; }
.order-total { text-align: right; font-weight: 600; }
</style>

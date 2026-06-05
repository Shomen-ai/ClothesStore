<!--
  OrderDetailView — a single customer order (route: /account/orders/:id).
  Loads one order by the :id route param and renders its status, totals, and line items.
-->
<template>
  <!-- Render nothing until the order has loaded (order starts as null) -->
  <div v-if="order">
    <div class="detail-header">
      <h2 class="section-title">ЗАКАЗ #{{ order.id }}</h2>
      <el-tag :type="statusType(order.status)">{{ statusLabel(order.status) }}</el-tag>
    </div>
    <div class="order-meta">
      <p>Дата: {{ formatDate(order.created_at) }}</p>
      <p>Сумма: {{ formatPrice(order.total_price) }}</p>
      <p v-if="order.discount_amount > 0">Скидка: -{{ formatPrice(order.discount_amount) }}</p>
    </div>
    <h3 style="margin:24px 0 16px;font-size:14px;letter-spacing:2px">СОСТАВ ЗАКАЗА</h3>
    <div v-for="item in order.items" :key="item.id" class="order-item">
      <span>Товар #{{ item.product_id }} — размер #{{ item.product_size_id }}</span>
      <span>{{ item.quantity }} шт. × {{ formatPrice(item.price_at_order) }}</span>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getUserOrder } from '@/api/user.js'

const route = useRoute()
const order = ref(null)
// Fetch the order keyed by the URL :id param (server enforces it belongs to this user).
onMounted(async () => { const { data } = await getUserOrder(route.params.id); order.value = data })

// Map raw status codes to Russian labels and to Element Plus tag color types.
const statusMap = { pending:'Ожидает', confirmed:'Подтверждён', shipped:'Отправлен', delivered:'Доставлен', cancelled:'Отменён' }
const typeMap = { pending:'info', confirmed:'warning', shipped:'', delivered:'success', cancelled:'danger' }
function statusLabel(s) { return statusMap[s] || s }
function statusType(s) { return typeMap[s] || '' }
function formatDate(d) { return new Date(d).toLocaleDateString('ru-RU') }
function formatPrice(p) { return new Intl.NumberFormat('ru-RU',{style:'currency',currency:'RUB',maximumFractionDigits:0}).format(p) }
</script>

<style scoped>
.section-title { font-size: 20px; font-weight: 700; letter-spacing: 3px; }
.detail-header { display: flex; align-items: center; gap: 16px; margin-bottom: 24px; }
.order-meta { color: var(--color-text-muted); font-size: 14px; line-height: 2; margin-bottom: 8px; }
.order-item { display: flex; justify-content: space-between; padding: 12px 0; border-bottom: 1px solid var(--color-border); font-size: 14px; }
</style>

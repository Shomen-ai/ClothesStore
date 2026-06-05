<!--
  AdminOrdersView — admin order management (route: /admin/orders).
  Lists all orders with server-side filtering by status and date range, and lets the admin
  change an order's status inline via a per-row select. Mutations rely on the backend
  enforcing the admin role (the route guard only hides the UI, see review note).
-->
<template>
  <div>
    <header class="admin-view-head">
      <div>
        <p class="eyebrow">Админка / Операции</p>
        <h1 class="head-title">Заказы <span class="gothic-accent">поток</span></h1>
      </div>
    </header>
    <div class="filters-row">
      <el-select v-model="statusFilter" placeholder="Статус" clearable @change="load" style="width:160px">
        <el-option v-for="s in statuses" :key="s.value" :value="s.value" :label="s.label" />
      </el-select>
      <el-date-picker v-model="dateRange" type="daterange" start-placeholder="От" end-placeholder="До"
        value-format="YYYY-MM-DD" @change="load" style="width:280px" />
    </div>

    <el-table :data="orders">
      <el-table-column prop="id" label="#" width="80" />
      <el-table-column prop="user_id" label="Клиент" width="100" />
      <el-table-column label="Дата" width="120">
        <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="Сумма" width="140">
        <template #default="{ row }">{{ formatPrice(row.total_price) }}</template>
      </el-table-column>
      <el-table-column label="Статус" width="160">
        <template #default="{ row }">
          <!-- Inline status editor: :model-value (one-way) + @change so the value only
               updates after the server call succeeds in changeStatus -->
          <el-select :model-value="row.status" size="small" @change="changeStatus(row, $event)" style="width:140px">
            <el-option v-for="s in statuses" :key="s.value" :value="s.value" :label="s.label" />
          </el-select>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { adminListOrders, adminUpdateStatus } from '@/api/admin.js'

const orders = ref([])
const statusFilter = ref('')
const dateRange = ref([])

const statuses = [
  { value:'pending', label:'Ожидает' },
  { value:'confirmed', label:'Подтверждён' },
  { value:'shipped', label:'Отправлен' },
  { value:'delivered', label:'Доставлен' },
  { value:'cancelled', label:'Отменён' },
]

onMounted(load)

// Build query params from the active filters and (re)fetch. Called on mount and on every
// filter change, so filtering is fully server-side.
async function load() {
  const params = {}
  if (statusFilter.value) params.status = statusFilter.value
  if (dateRange.value?.length) { params.date_from = dateRange.value[0]; params.date_to = dateRange.value[1] }
  const { data } = await adminListOrders(params)
  orders.value = data || []
}

// Persist the new status, then patch the local row so the UI stays in sync without a refetch.
async function changeStatus(row, status) {
  await adminUpdateStatus(row.id, status)
  row.status = status
}

function formatDate(d) { return new Date(d).toLocaleDateString('ru-RU') }
function formatPrice(p) { return new Intl.NumberFormat('ru-RU',{style:'currency',currency:'RUB',maximumFractionDigits:0}).format(p) }
</script>

<style scoped>
.admin-view-head {
  display: flex; justify-content: space-between; align-items: flex-end; gap: 24px;
  padding-bottom: 24px;
  margin-bottom: 28px;
  border-bottom: 1px solid var(--border);
}
.head-title {
  font-size: clamp(32px, 4vw, 48px);
  font-weight: 700;
  letter-spacing: -0.02em;
  line-height: 1;
  margin-top: 12px;
}
.head-title .gothic-accent { font-size: 1.05em; }
.filters-row { display: flex; gap: 16px; margin-bottom: 24px; }
</style>

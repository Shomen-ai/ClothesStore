<template>
  <div>
    <h2 class="page-title">ЗАКАЗЫ</h2>
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

async function load() {
  const params = {}
  if (statusFilter.value) params.status = statusFilter.value
  if (dateRange.value?.length) { params.date_from = dateRange.value[0]; params.date_to = dateRange.value[1] }
  const { data } = await adminListOrders(params)
  orders.value = data || []
}

async function changeStatus(row, status) {
  await adminUpdateStatus(row.id, status)
  row.status = status
}

function formatDate(d) { return new Date(d).toLocaleDateString('ru-RU') }
function formatPrice(p) { return new Intl.NumberFormat('ru-RU',{style:'currency',currency:'RUB',maximumFractionDigits:0}).format(p) }
</script>

<style scoped>
.page-title { font-size: 20px; font-weight: 800; letter-spacing: 3px; margin-bottom: 24px; }
.filters-row { display: flex; gap: 16px; margin-bottom: 24px; }
</style>

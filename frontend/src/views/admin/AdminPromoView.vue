<template>
  <div>
    <div class="view-header">
      <h2 class="page-title">ПРОМОКОДЫ</h2>
      <el-button type="primary" @click="dialogOpen = true; resetForm()">+ Создать</el-button>
    </div>

    <el-table :data="promos">
      <el-table-column prop="code" label="Код" width="140">
        <template #default="{ row }"><span style="font-family:monospace;color:var(--color-red)">{{ row.code }}</span></template>
      </el-table-column>
      <el-table-column label="Скидка" width="120">
        <template #default="{ row }">
          {{ row.discount_type === 'percent' ? row.discount_value + '%' : formatPrice(row.discount_value) }}
        </template>
      </el-table-column>
      <el-table-column label="Активаций" width="140">
        <template #default="{ row }">
          {{ row.activations_count }} / {{ row.max_activations ?? '∞' }}
        </template>
      </el-table-column>
      <el-table-column label="Истекает" width="140">
        <template #default="{ row }">{{ row.expires_at ? formatDate(row.expires_at) : 'Бессрочно' }}</template>
      </el-table-column>
      <el-table-column label="Статус" width="120">
        <template #default="{ row }">
          <el-tag :type="row.is_active ? 'success' : 'danger'" size="small">
            {{ row.is_active ? 'Активен' : 'Неактивен' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="Действия" width="180">
        <template #default="{ row }">
          <el-button v-if="row.is_active" size="small" @click="deactivate(row)">Деакт.</el-button>
          <el-button size="small" type="danger" @click="remove(row.id)">Удалить</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogOpen" title="Новый промокод" width="480px">
      <el-form :model="form" label-position="top">
        <el-form-item label="Код (пусто = авто)"><el-input v-model="form.code" placeholder="AUTO" /></el-form-item>
        <el-form-item label="Тип скидки">
          <el-radio-group v-model="form.discount_type">
            <el-radio value="percent">Процент %</el-radio>
            <el-radio value="fixed">Фиксированная сумма ₽</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="Значение скидки">
          <el-input-number v-model="form.discount_value" :min="0" />
        </el-form-item>
        <el-form-item label="Макс. активаций (пусто = без лимита)">
          <el-input-number v-model="form.max_activations" :min="1" :controls="false" placeholder="∞" />
        </el-form-item>
        <el-form-item label="Срок действия до (пусто = бессрочно)">
          <el-date-picker v-model="form.expires_at" type="date" value-format="YYYY-MM-DDTHH:mm:ss[Z]" style="width:100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogOpen = false">Отмена</el-button>
        <el-button type="primary" :loading="saving" @click="create">Создать</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { adminListPromos, adminCreatePromo, adminDeactivatePromo, adminDeletePromo } from '@/api/admin.js'

const promos = ref([])
const dialogOpen = ref(false)
const saving = ref(false)
const form = ref({})

onMounted(load)

async function load() { const { data } = await adminListPromos(); promos.value = data || [] }

function resetForm() {
  form.value = { code:'', discount_type:'percent', discount_value:10, max_activations:null, expires_at:null }
}

async function create() {
  saving.value = true
  const payload = { ...form.value }
  if (!payload.max_activations) delete payload.max_activations
  if (!payload.expires_at) delete payload.expires_at
  try { await adminCreatePromo(payload); dialogOpen.value = false; load() }
  finally { saving.value = false }
}

async function deactivate(row) { await adminDeactivatePromo(row.id); row.is_active = false }
async function remove(id) { await adminDeletePromo(id); promos.value = promos.value.filter(p => p.id !== id) }

function formatDate(d) { return new Date(d).toLocaleDateString('ru-RU') }
function formatPrice(p) { return new Intl.NumberFormat('ru-RU',{style:'currency',currency:'RUB',maximumFractionDigits:0}).format(p) }
</script>

<style scoped>
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-title { font-size: 20px; font-weight: 800; letter-spacing: 3px; }
</style>

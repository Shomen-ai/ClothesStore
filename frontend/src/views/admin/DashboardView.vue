<template>
  <div>
    <div class="dash-header">
      <h2 class="page-title">ДАШБОРД</h2>
      <el-radio-group v-model="period" @change="loadAll">
        <el-radio-button value="day">День</el-radio-button>
        <el-radio-button value="week">Неделя</el-radio-button>
        <el-radio-button value="month">Месяц</el-radio-button>
        <el-radio-button value="quarter">Квартал</el-radio-button>
        <el-radio-button value="all">Всё время</el-radio-button>
      </el-radio-group>
    </div>

    <div class="stats-grid">
      <el-card class="stat-card wide">
        <div class="card-title">ВЫРУЧКА</div>
        <div class="big-number">{{ formatPrice(totalRevenue) }}</div>
        <StatsChart :points="revenuePoints" />
      </el-card>

      <el-card class="stat-card">
        <div class="card-title">ЗАКАЗЫ ПО СТАТУСАМ</div>
        <div v-for="s in orderStats" :key="s.status" class="status-row">
          <span class="status-label">{{ statusLabel(s.status) }}</span>
          <span class="status-count">{{ s.count }}</span>
        </div>
        <div v-if="!orderStats.length" class="empty-stat">Нет данных</div>
      </el-card>

      <el-card class="stat-card">
        <div class="card-title">ТОП ТОВАРОВ</div>
        <div v-for="(p, i) in topProducts" :key="p.product_id" class="top-row">
          <span class="top-rank">{{ i + 1 }}</span>
          <span class="top-name">{{ p.name }}</span>
          <span class="top-sold">{{ p.sold }} шт.</span>
        </div>
        <div v-if="!topProducts.length" class="empty-stat">Нет данных</div>
      </el-card>

      <el-card class="stat-card">
        <div class="card-title">АКТИВАЦИИ ПРОМОКОДОВ</div>
        <div v-for="p in promoStats" :key="p.code" class="promo-stat-row">
          <span class="promo-code">{{ p.code }}</span>
          <span class="promo-count">{{ p.activations }}</span>
        </div>
        <div v-if="!promoStats.length" class="empty-stat">Нет данных</div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getRevenueStats, getOrderStats, getPromoStats } from '@/api/admin.js'
import StatsChart from '@/components/admin/StatsChart.vue'

const period = ref('month')
const revenuePoints = ref([])
const orderStats = ref([])
const topProducts = ref([])
const promoStats = ref([])

const totalRevenue = computed(() => revenuePoints.value.reduce((s, p) => s + p.revenue, 0))

const statusMap = { pending:'Ожидает', confirmed:'Подтверждён', shipped:'Отправлен', delivered:'Доставлен', cancelled:'Отменён' }
function statusLabel(s) { return statusMap[s] || s }

async function loadAll() {
  const [rev, ord, promo] = await Promise.all([
    getRevenueStats(period.value),
    getOrderStats(period.value),
    getPromoStats(period.value),
  ])
  revenuePoints.value = rev.data || []
  orderStats.value = ord.data?.by_status || []
  topProducts.value = ord.data?.top_products || []
  promoStats.value = promo.data || []
}

onMounted(loadAll)

function formatPrice(p) {
  return new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 0 }).format(p)
}
</script>

<style scoped>
.dash-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 32px; flex-wrap: wrap; gap: 16px; }
.page-title { font-size: 20px; font-weight: 800; letter-spacing: 3px; }
.stats-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 24px; }
.stat-card.wide { grid-column: 1 / -1; }
.card-title { font-size: 11px; letter-spacing: 2px; color: var(--color-text-muted); margin-bottom: 16px; }
.big-number { font-size: 36px; font-weight: 800; margin-bottom: 24px; }
.status-row, .top-row, .promo-stat-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid var(--color-border); font-size: 14px; }
.top-rank { color: var(--color-red); font-weight: 700; width: 24px; }
.top-name { flex: 1; }
.promo-code { font-family: monospace; color: var(--color-red); }
.empty-stat { color: var(--color-text-muted); font-size: 13px; }
</style>

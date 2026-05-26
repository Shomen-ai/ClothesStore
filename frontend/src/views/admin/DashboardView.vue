<template>
  <div class="dashboard">
    <header class="view-head">
      <div class="head-marquee-wrap" aria-hidden="true">
        <div class="head-marquee">
          <div class="marquee-row row-1">
            <div class="marquee-track">
              <span>METRICS · REVENUE · ORDERS · DROP · SIGNAL · LOUD · STATS ·&nbsp;</span>
              <span>METRICS · REVENUE · ORDERS · DROP · SIGNAL · LOUD · STATS ·&nbsp;</span>
            </div>
          </div>
          <div class="marquee-row row-2">
            <div class="marquee-track">
              <span>цифры — выручка — кассы — pulse — данные — pattern —&nbsp;</span>
              <span>цифры — выручка — кассы — pulse — данные — pattern —&nbsp;</span>
            </div>
          </div>
          <div class="marquee-row row-3">
            <div class="marquee-track">
              <span>TRACK · YIELD · RUN · GROW · CONVERT · RHYTHM · INTAKE ·&nbsp;</span>
              <span>TRACK · YIELD · RUN · GROW · CONVERT · RHYTHM · INTAKE ·&nbsp;</span>
            </div>
          </div>
        </div>
      </div>
      <div class="head-text">
        <p class="eyebrow">Админка / Дашборд</p>
        <h1 class="head-title">Цифры <span class="gothic-accent">громко</span></h1>
      </div>
      <button class="cta-add report-cta" :disabled="reportLoading" @click="downloadReport">
        <span class="cta-mark">⇩</span>
        <span class="cta-label">{{ reportLoading ? 'Готовим…' : 'Скачать отчёт' }}</span>
        <span class="cta-arrow">.xlsx</span>
      </button>
    </header>

    <div class="period-bar">
      <span class="period-label">Период</span>
      <div class="period-switch">
        <button v-for="opt in periodOptions" :key="opt.value"
          class="period-btn" :class="{ active: period === opt.value }"
          @click="period = opt.value; loadAll()">
          {{ opt.label }}
        </button>
      </div>
    </div>

    <div class="stats-grid">
      <article class="stat-card wide">
        <div class="card-head">
          <span class="card-eyebrow">Выручка</span>
          <span class="card-period">/ {{ currentPeriodLabel }}</span>
        </div>
        <div class="big-number">{{ formatPrice(totalRevenue) }}</div>
        <StatsChart :points="revenuePoints" :key="`${period}-${theme.theme}`" />
      </article>

      <article class="stat-card">
        <div class="card-head">
          <span class="card-eyebrow">Заказы по статусам</span>
        </div>
        <ul class="rows">
          <li v-for="s in orderStats" :key="s.status" class="row-line">
            <span class="row-label">
              <span class="row-dot" :class="`dot-${s.status}`"></span>
              {{ statusLabel(s.status) }}
            </span>
            <span class="row-value">{{ s.count }}</span>
          </li>
          <li v-if="!orderStats.length" class="row-empty">пока пусто</li>
        </ul>
      </article>

      <article class="stat-card">
        <div class="card-head">
          <span class="card-eyebrow">Топ товаров</span>
        </div>
        <ul class="rows">
          <li v-for="(p, i) in topProducts" :key="p.product_id" class="row-line">
            <span class="row-rank">{{ String(i + 1).padStart(2, '0') }}</span>
            <span class="row-name">{{ p.name }}</span>
            <span class="row-value">{{ p.sold }}<span class="row-unit"> шт</span></span>
          </li>
          <li v-if="!topProducts.length" class="row-empty">пока пусто</li>
        </ul>
      </article>

      <article class="stat-card">
        <div class="card-head">
          <span class="card-eyebrow">Промокоды</span>
        </div>
        <ul class="rows">
          <li v-for="p in promoStats" :key="p.code" class="row-line">
            <span class="row-code">{{ p.code }}</span>
            <span class="row-value">{{ p.activations }}<span class="row-unit"> акт.</span></span>
          </li>
          <li v-if="!promoStats.length" class="row-empty">пока пусто</li>
        </ul>
      </article>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getRevenueStats, getOrderStats, getPromoStats, downloadExcelReport } from '@/api/admin.js'
import StatsChart from '@/components/admin/StatsChart.vue'
import { useThemeStore } from '@/stores/theme.js'

const theme = useThemeStore()

const period = ref('month')
const revenuePoints = ref([])
const orderStats = ref([])
const topProducts = ref([])
const promoStats = ref([])
const reportLoading = ref(false)

async function downloadReport() {
  reportLoading.value = true
  try {
    const res = await downloadExcelReport()
    const blob = new Blob([res.data], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    const now = new Date()
    const pad = n => String(n).padStart(2, '0')
    a.href = url
    a.download = `report_${now.getFullYear()}-${pad(now.getMonth()+1)}-${pad(now.getDate())}_${pad(now.getHours())}${pad(now.getMinutes())}.xlsx`
    document.body.appendChild(a)
    a.click()
    a.remove()
    URL.revokeObjectURL(url)
  } catch (e) {
    console.error(e)
  } finally {
    reportLoading.value = false
  }
}

const periodOptions = [
  { value: 'day',     label: 'День' },
  { value: 'week',    label: 'Неделя' },
  { value: 'month',   label: 'Месяц' },
  { value: 'quarter', label: 'Квартал' },
  { value: 'all',     label: 'Всё время' },
]
const currentPeriodLabel = computed(() => periodOptions.find(o => o.value === period.value)?.label || '')

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
.dashboard { min-width: 0; }

.view-head {
  position: relative;
  margin-bottom: 28px;
  padding: 0 0 28px;
  min-height: 180px;
  border-bottom: 1px solid var(--border);
  overflow: visible;
  isolation: isolate;
  display: flex;
  flex-direction: row;
  align-items: flex-end;
  justify-content: space-between;
  gap: 24px;
}
.head-text { position: relative; z-index: 2; flex: 1; }
.report-cta { position: relative; z-index: 2; flex-shrink: 0; }
.report-cta:disabled { opacity: 0.7; cursor: progress; }
.report-cta .cta-arrow { font-family: var(--font-mono); font-size: 10px; letter-spacing: 0.18em; }
@media (max-width: 720px) {
  .view-head { flex-direction: column; align-items: stretch; }
  .report-cta { align-self: flex-start; }
}
.head-title {
  font-size: clamp(36px, 5vw, 56px);
  font-weight: 700;
  letter-spacing: -0.02em;
  line-height: 1;
  color: var(--text);
  margin-top: 12px;
}
.head-title .gothic-accent { font-size: 1.05em; }

.period-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}
.period-label {
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--text-muted);
}
.period-switch {
  display: inline-flex;
  border: 1px solid var(--border-strong);
  background: var(--bg-surface);
}
.period-btn {
  background: transparent;
  border: 0;
  color: var(--text-muted);
  font: 500 11px/1 var(--font-mono);
  letter-spacing: 0.14em;
  text-transform: uppercase;
  padding: 10px 14px;
  cursor: pointer;
  border-right: 1px solid var(--border);
  transition: color .15s ease, background .15s ease;
}
.period-btn:last-child { border-right: 0; }
.period-btn:hover { color: var(--text); }
.period-btn.active {
  color: var(--text);
  background: var(--accent-haze);
  box-shadow: inset 0 -2px 0 0 var(--accent);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}
.stat-card {
  position: relative;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  padding: 24px 24px 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  transition: border-color .2s ease, transform .2s ease;
}
.stat-card:hover { border-color: var(--border-strong); }
.stat-card.wide { grid-column: 1 / -1; }

.card-head { display: flex; align-items: baseline; justify-content: space-between; }
.card-eyebrow {
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--text-muted);
}
.card-period {
  font-family: var(--font-mono);
  font-size: 10px;
  color: var(--text-dim);
  letter-spacing: 0.1em;
}

.big-number {
  font-size: clamp(36px, 4.5vw, 54px);
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--text);
}

.rows { list-style: none; display: flex; flex-direction: column; }
.row-line {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 0;
  border-bottom: 1px dashed var(--border);
  font-size: 14px;
  color: var(--text-soft);
}
.row-line:last-child { border-bottom: 0; }
.row-label { display: inline-flex; align-items: center; gap: 10px; flex: 1; }
.row-dot { width: 8px; height: 8px; border-radius: 50%; background: var(--text-dim); flex-shrink: 0; }
.row-dot.dot-pending   { background: #eab308; }
.row-dot.dot-confirmed { background: #3b82f6; }
.row-dot.dot-shipped   { background: #8b5cf6; }
.row-dot.dot-delivered { background: #22c55e; box-shadow: 0 0 8px rgba(34,197,94,0.5); }
.row-dot.dot-cancelled { background: #ef4848; }

.row-rank {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--accent);
  letter-spacing: 0.04em;
  width: 22px;
}
.row-name { flex: 1; color: var(--text); }
.row-code {
  flex: 1;
  font-family: var(--font-mono);
  color: var(--accent-soft);
  letter-spacing: 0.08em;
  font-size: 13px;
}
.row-value {
  font-weight: 600;
  color: var(--text);
  font-variant-numeric: tabular-nums;
}
.row-unit { color: var(--text-dim); font-weight: 400; font-size: 11px; margin-left: 2px; }
.row-empty {
  font-family: var(--font-serif);
  font-style: italic;
  color: var(--text-dim);
  padding: 12px 0;
}

@media (max-width: 1100px) {
  .stats-grid { grid-template-columns: 1fr 1fr; }
}
@media (max-width: 720px) {
  .stats-grid { grid-template-columns: 1fr; }
  .view-head { min-height: 140px; }
  .head-title { font-size: 36px; }
  .period-btn { padding: 8px 10px; font-size: 10px; }
}
</style>

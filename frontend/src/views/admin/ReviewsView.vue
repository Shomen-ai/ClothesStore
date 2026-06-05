<!--
  ReviewsView — admin moderation of product reviews (route: /admin/reviews).
  Lists reviews with a server-side visibility filter, and lets the admin hide/show or delete
  each review. The filter value is the string the API expects for ?hidden= (see load()).
  Admin role is enforced server-side.
-->
<template>
  <div>
    <header class="admin-view-head">
      <div>
        <p class="eyebrow">Админка / Отзывы</p>
        <h1 class="head-title">Голоса <span class="gothic-accent">клиентов</span></h1>
      </div>
      <div class="filter-row">
        <!-- Filter maps to ?hidden=: '' = all, 'false' = visible, 'true' = hidden -->
        <el-radio-group v-model="filter" size="small" @change="load">
          <el-radio-button :value="''">Все</el-radio-button>
          <el-radio-button :value="'false'">Видимые</el-radio-button>
          <el-radio-button :value="'true'">Скрытые</el-radio-button>
        </el-radio-group>
      </div>
    </header>

    <el-table :data="reviews" stripe v-loading="loading">
      <el-table-column label="Товар" min-width="240">
        <template #default="{ row }">
          <RouterLink :to="`/catalogue/${row.product_id}`" class="product-link">{{ row.product_name }}</RouterLink>
        </template>
      </el-table-column>

      <el-table-column label="Автор" min-width="200">
        <template #default="{ row }">
          <div class="author">
            <span class="author-name">{{ row.user_name || '—' }}</span>
            <span class="author-email">{{ row.user_email }}</span>
          </div>
        </template>
      </el-table-column>

      <el-table-column label="Оценка" width="130">
        <template #default="{ row }">
          <div class="rating-cell">
            <span class="rate-num">{{ row.rating }}</span>
            <span class="stars">
              <span v-for="i in 5" :key="i" :class="{ on: i <= row.rating }">★</span>
            </span>
          </div>
        </template>
      </el-table-column>

      <el-table-column label="Отзыв" min-width="320">
        <template #default="{ row }">
          <p class="rv-text" v-if="row.text">{{ row.text }}</p>
          <span v-else class="muted">—</span>
        </template>
      </el-table-column>

      <el-table-column label="Дата" width="120">
        <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
      </el-table-column>

      <el-table-column label="Статус" width="110">
        <template #default="{ row }">
          <el-tag :type="row.is_hidden ? 'info' : 'success'" size="small">
            {{ row.is_hidden ? 'Скрыт' : 'Виден' }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column label="Действия" width="170">
        <template #default="{ row }">
          <div class="row-actions">
            <el-button size="small" @click="toggleHidden(row)">
              {{ row.is_hidden ? 'Показать' : 'Скрыть' }}
            </el-button>
            <el-button size="small" type="danger" @click="remove(row)">Удалить</el-button>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <p v-if="!loading && reviews.length === 0" class="empty">Пока отзывов нет.</p>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { adminListReviews, adminSetReviewHidden, adminDeleteReview } from '@/api/admin.js'
import { ElMessage, ElMessageBox } from 'element-plus'

const reviews = ref([])
const loading = ref(false)
const filter = ref('')   // '', 'false', 'true'

onMounted(load)

// Load reviews, passing the visibility filter as the ?hidden= param. Empty filter ('') is
// converted to undefined so the API omits the param and returns all reviews.
async function load() {
  loading.value = true
  try {
    const param = filter.value === '' ? undefined : filter.value
    const { data } = await adminListReviews(param)
    reviews.value = data || []
  } catch (e) {
    ElMessage({ message: e.response?.data?.error || 'Ошибка загрузки', type: 'error' })
  } finally { loading.value = false }
}

// Flip the review's hidden flag on the server, then patch the local row to match.
async function toggleHidden(row) {
  try {
    await adminSetReviewHidden(row.id, !row.is_hidden)
    row.is_hidden = !row.is_hidden
    ElMessage({ message: row.is_hidden ? 'Скрыто' : 'Показано', duration: 1400 })
  } catch (e) {
    ElMessage({ message: e.response?.data?.error || 'Ошибка', type: 'error' })
  }
}

async function remove(row) {
  try {
    // Confirm dialog; a rejected promise (user cancelled) short-circuits via the catch.
    await ElMessageBox.confirm(`Удалить отзыв "${row.product_name}"?`, 'Подтверждение', {
      type: 'warning', confirmButtonText: 'Удалить', cancelButtonText: 'Отмена',
    })
  } catch { return }
  try {
    await adminDeleteReview(row.id)
    reviews.value = reviews.value.filter(r => r.id !== row.id)
    ElMessage({ message: 'Удалено', duration: 1400 })
  } catch (e) {
    ElMessage({ message: e.response?.data?.error || 'Ошибка', type: 'error' })
  }
}

function formatDate(d) { return new Date(d).toLocaleDateString('ru-RU') }
</script>

<style scoped>
.admin-view-head {
  display: flex; justify-content: space-between; align-items: flex-end; gap: 24px;
  padding-bottom: 24px;
  margin-bottom: 32px;
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

.filter-row :deep(.el-radio-button__inner) {
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.08em;
}

.product-link {
  color: var(--text);
  font-weight: 500;
  text-decoration: none;
  transition: color .15s ease;
}
.product-link:hover { color: var(--accent); }

.author {
  display: flex; flex-direction: column; gap: 2px;
}
.author-name { color: var(--text); font-weight: 500; font-size: 13px; }
.author-email { color: var(--text-muted); font-size: 11px; font-family: var(--font-mono); }

.rating-cell {
  display: flex; align-items: center; gap: 6px;
}
.rate-num {
  font-family: var(--font-mono);
  font-weight: 600;
  color: var(--text);
}
.stars span { color: var(--text-dim); font-size: 12px; }
.stars span.on { color: var(--accent); }

.rv-text {
  font-size: 13px;
  line-height: 1.4;
  color: var(--text-soft);
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.muted { color: var(--text-dim); }

.row-actions { display: flex; flex-direction: column; gap: 6px; }
.row-actions :deep(.el-button) { width: 100%; margin-left: 0 !important; }

.empty {
  text-align: center;
  padding: 64px 0;
  color: var(--text-muted);
  font-family: var(--font-mono);
  font-size: 12px;
  letter-spacing: 0.1em;
}
</style>

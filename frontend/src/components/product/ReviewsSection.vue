<template>
  <section class="reviews">
    <header class="rv-head">
      <div class="rv-head-text">
        <p class="eyebrow">Отзывы</p>
        <h2 class="rv-title">Что говорят <span class="serif-italic">люди</span></h2>
      </div>

      <div class="rv-summary" v-if="summary.count > 0">
        <div class="rv-avg">{{ formatAvg(summary.avgRating) }}</div>
        <div class="rv-avg-meta">
          <div class="rv-stars" :aria-label="`Средняя ${summary.avgRating} из 5`">
            <span v-for="i in 5" :key="i" class="star" :class="{ on: i <= Math.round(summary.avgRating) }">★</span>
          </div>
          <p class="rv-avg-count">{{ summary.count }} {{ pluralize(summary.count) }}</p>
        </div>
      </div>
      <div class="rv-summary rv-summary--empty" v-else>
        <p>Отзывов пока нет</p>
        <p class="muted">Будь первым, кто их оставит</p>
      </div>
    </header>

    <!-- My review state -->
    <div class="rv-mine" v-if="auth.isLoggedIn && !loadingMine">
      <!-- own review exists -->
      <div v-if="mine.own && !editing" class="rv-card rv-card--mine">
        <div class="rv-card-head">
          <div class="rv-stars">
            <span v-for="i in 5" :key="i" class="star" :class="{ on: i <= mine.own.rating }">★</span>
          </div>
          <span class="rv-tag">твой отзыв</span>
        </div>
        <p class="rv-text" v-if="mine.own.text">{{ mine.own.text }}</p>
        <div class="rv-card-foot">
          <span class="rv-date">{{ formatDate(mine.own.updated_at) }}</span>
          <div class="rv-actions">
            <button class="rv-btn-ghost" @click="startEdit">Изменить</button>
            <button class="rv-btn-ghost rv-btn-danger" @click="onDelete">Удалить</button>
          </div>
        </div>
      </div>

      <!-- form: new or editing -->
      <form v-else-if="(mine.eligible && !mine.own) || editing" class="rv-form" @submit.prevent="onSubmit">
        <p class="eyebrow">{{ editing ? 'Редактирование' : 'Оставить отзыв' }}</p>
        <div class="rv-stars rv-stars--picker" role="radiogroup" aria-label="Оценка">
          <button v-for="i in 5" :key="i" type="button"
                  class="star star--btn"
                  :class="{ on: i <= form.rating || i <= hoverRating, hovering: hoverRating > 0 }"
                  @mouseenter="hoverRating = i"
                  @mouseleave="hoverRating = 0"
                  @click="form.rating = i"
                  :aria-label="`${i} из 5`">★</button>
          <span class="rv-rating-hint">{{ form.rating ? form.rating + ' / 5' : 'выбери оценку' }}</span>
        </div>
        <textarea v-model="form.text" rows="4" placeholder="Несколько слов о товаре (необязательно)" maxlength="2000"></textarea>
        <div class="rv-form-foot">
          <span class="rv-form-error" v-if="error">{{ error }}</span>
          <div class="rv-form-actions">
            <button v-if="editing" type="button" class="rv-btn-ghost" @click="cancelEdit">Отмена</button>
            <button type="submit" class="rv-btn-primary" :disabled="!form.rating || saving">
              <span>{{ saving ? 'Отправка…' : (editing ? 'Сохранить' : 'Опубликовать') }}</span>
              <span class="arrow" v-if="!saving">→</span>
            </button>
          </div>
        </div>
      </form>

      <!-- not eligible -->
      <div v-else class="rv-hint">
        <span class="rv-hint-mark">·</span>
        <span>Отзыв доступен после получения товара.</span>
      </div>
    </div>

    <!-- not logged in -->
    <div v-else-if="!auth.isLoggedIn" class="rv-hint">
      <span class="rv-hint-mark">·</span>
      <RouterLink :to="`/login?redirect=/catalogue/${productId}`">Войди</RouterLink>, чтобы оставить отзыв.
    </div>

    <!-- Reviews list -->
    <div class="rv-list" v-if="reviews.length">
      <article v-for="rv in reviews" :key="rv.id" class="rv-card">
        <div class="rv-card-head">
          <span class="rv-author">{{ rv.user_name || 'Покупатель' }}</span>
          <div class="rv-stars">
            <span v-for="i in 5" :key="i" class="star" :class="{ on: i <= rv.rating }">★</span>
          </div>
        </div>
        <p class="rv-text" v-if="rv.text">{{ rv.text }}</p>
        <div class="rv-card-foot">
          <span class="rv-date">{{ formatDate(rv.created_at) }}</span>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { useAuthStore } from '@/stores/auth.js'
import { getProductReviews, getMyReviewState, createReview, updateReview, deleteReview } from '@/api/reviews.js'
import { ElMessage, ElMessageBox } from 'element-plus'

const props = defineProps({ productId: { type: [Number, String], required: true } })
const auth = useAuthStore()

const reviews = ref([])
const summary = reactive({ avgRating: 0, count: 0 })
const mine = reactive({ eligible: false, own: null })
const loadingMine = ref(false)
const editing = ref(false)
const saving = ref(false)
const error = ref('')
const hoverRating = ref(0)
const form = reactive({ rating: 0, text: '' })

async function loadList() {
  try {
    const { data } = await getProductReviews(props.productId)
    reviews.value = data.reviews || []
    summary.avgRating = data.avg_rating || 0
    summary.count = data.count || 0
  } catch (e) { console.error(e) }
}

async function loadMine() {
  if (!auth.isLoggedIn) { mine.eligible = false; mine.own = null; return }
  loadingMine.value = true
  try {
    const { data } = await getMyReviewState(props.productId)
    mine.eligible = !!data.eligible
    mine.own = data.own || null
  } catch (e) { console.error(e) } finally { loadingMine.value = false }
}

onMounted(async () => {
  await Promise.all([loadList(), loadMine()])
})

watch(() => props.productId, async () => {
  reviews.value = []; summary.avgRating = 0; summary.count = 0
  mine.eligible = false; mine.own = null
  editing.value = false; resetForm()
  await Promise.all([loadList(), loadMine()])
})

function resetForm() { form.rating = 0; form.text = ''; error.value = '' }

function startEdit() {
  editing.value = true
  form.rating = mine.own.rating
  form.text = mine.own.text || ''
  error.value = ''
}
function cancelEdit() { editing.value = false; resetForm() }

async function onSubmit() {
  if (!form.rating) { error.value = 'Поставь оценку'; return }
  saving.value = true; error.value = ''
  try {
    if (editing.value && mine.own) {
      const { data } = await updateReview(props.productId, mine.own.id, { rating: form.rating, text: form.text })
      mine.own = data
      editing.value = false
      ElMessage({ message: '✓ Отзыв обновлён', type: 'success', duration: 1600 })
    } else {
      const { data } = await createReview(props.productId, { rating: form.rating, text: form.text })
      mine.own = data
      ElMessage({ message: '✓ Спасибо за отзыв', type: 'success', duration: 1600 })
    }
    resetForm()
    await loadList()
  } catch (e) {
    error.value = e.response?.data?.error || 'Не получилось отправить'
  } finally { saving.value = false }
}

async function onDelete() {
  try {
    await ElMessageBox.confirm('Удалить отзыв?', 'Подтверждение', { type: 'warning', confirmButtonText: 'Удалить', cancelButtonText: 'Отмена' })
  } catch { return }
  try {
    await deleteReview(props.productId, mine.own.id)
    mine.own = null
    await loadList()
    ElMessage({ message: 'Удалено', duration: 1400 })
  } catch (e) {
    ElMessage({ message: e.response?.data?.error || 'Ошибка', type: 'error' })
  }
}

function formatAvg(v) { return Number(v).toFixed(1).replace('.', ',') }
function formatDate(d) {
  return new Date(d).toLocaleDateString('ru-RU', { day: '2-digit', month: 'short', year: 'numeric' })
}
function pluralize(n) {
  const mod10 = n % 10, mod100 = n % 100
  if (mod10 === 1 && mod100 !== 11) return 'отзыв'
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20)) return 'отзыва'
  return 'отзывов'
}
</script>

<style scoped>
.reviews {
  margin-top: 96px;
  padding-top: 48px;
  border-top: 1px dashed var(--border);
}

.rv-head {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 32px;
  align-items: end;
  margin-bottom: 32px;
  padding-bottom: 24px;
  border-bottom: 1px solid var(--border);
}
.rv-head-text .eyebrow {
  font-family: var(--font-mono);
  font-size: 11px; letter-spacing: 0.18em; text-transform: uppercase;
  color: var(--text-muted); margin: 0 0 8px;
}
.rv-title {
  font-size: clamp(28px, 3.5vw, 40px);
  font-weight: 700;
  letter-spacing: -0.02em;
  line-height: 1;
  margin: 0;
}
.rv-title .serif-italic {
  font-family: var(--font-serif);
  font-style: italic;
  font-weight: 500;
  color: var(--accent);
}

.rv-summary {
  display: flex;
  align-items: center;
  gap: 18px;
  text-align: right;
}
.rv-avg {
  font-family: var(--font-mono);
  font-size: 56px;
  font-weight: 500;
  color: var(--text);
  letter-spacing: -0.03em;
  line-height: 1;
}
.rv-avg-meta { text-align: left; }
.rv-avg-count {
  font-family: var(--font-mono);
  font-size: 11px; letter-spacing: 0.12em; text-transform: uppercase;
  color: var(--text-muted);
  margin: 6px 0 0;
}
.rv-summary--empty {
  text-align: right;
  color: var(--text-muted);
  font-size: 14px;
}
.rv-summary--empty .muted { font-family: var(--font-mono); font-size: 11px; letter-spacing: 0.1em; text-transform: uppercase; color: var(--text-dim); margin: 4px 0 0; }

.rv-stars { display: inline-flex; gap: 2px; color: var(--text-dim); font-size: 16px; letter-spacing: 1px; }
.rv-stars .star { transition: color .15s ease, transform .15s ease; }
.rv-stars .star.on { color: var(--accent); text-shadow: 0 0 8px var(--accent-glow); }
.rv-stars--picker {
  display: flex; align-items: center; gap: 6px;
  font-size: 28px;
  margin: 10px 0 4px;
}
.rv-stars--picker .star--btn {
  background: transparent; border: none; cursor: pointer;
  color: var(--text-dim);
  padding: 0 2px;
}
.rv-stars--picker .star--btn:hover { transform: translateY(-1px); }
.rv-stars--picker .star--btn.on { color: var(--accent); text-shadow: 0 0 10px var(--accent-glow); }
.rv-rating-hint {
  margin-left: 12px;
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.08em;
  color: var(--text-muted);
}

/* mine area */
.rv-mine { margin-bottom: 32px; }

.rv-form {
  background: var(--bg-elevated);
  border: 1px solid var(--border-strong);
  padding: 22px 24px;
  display: flex; flex-direction: column; gap: 12px;
  position: relative;
}
.rv-form::before {
  content: '';
  position: absolute; left: -1px; top: -1px; bottom: -1px; width: 3px;
  background: var(--accent);
}
.rv-form .eyebrow {
  font-family: var(--font-mono); font-size: 10.5px; letter-spacing: 0.2em;
  text-transform: uppercase; color: var(--text-muted); margin: 0;
}
.rv-form textarea {
  width: 100%;
  background: transparent;
  border: 1px solid var(--border);
  padding: 12px 14px;
  color: var(--text);
  font: 14px/1.5 var(--font-sans);
  resize: vertical;
  min-height: 96px;
  transition: border-color .15s ease;
}
.rv-form textarea:focus { outline: none; border-color: var(--text-soft); }
.rv-form-foot {
  display: flex; justify-content: space-between; align-items: center; gap: 12px;
}
.rv-form-error {
  color: var(--accent);
  font-family: var(--font-mono);
  font-size: 11px; letter-spacing: 0.06em;
}
.rv-form-actions { display: flex; gap: 8px; margin-left: auto; }
.rv-btn-primary {
  display: inline-flex; align-items: center; gap: 8px;
  padding: 10px 18px;
  background: var(--accent); color: #fff;
  border: 1px solid var(--accent);
  font: 600 12px/1 var(--font-sans);
  letter-spacing: 0.08em; text-transform: uppercase;
  cursor: pointer;
  transition: filter .15s ease, transform .15s ease;
}
.rv-btn-primary:hover:not(:disabled) { filter: brightness(1.08); transform: translateY(-1px); }
.rv-btn-primary:disabled { opacity: 0.45; cursor: not-allowed; }
.rv-btn-primary .arrow { font-family: var(--font-mono); }
.rv-btn-ghost {
  background: transparent;
  border: 1px solid var(--border-strong);
  color: var(--text-soft);
  padding: 8px 14px;
  font: 500 12px/1 var(--font-mono);
  letter-spacing: 0.08em; text-transform: uppercase;
  cursor: pointer;
  transition: color .15s ease, border-color .15s ease;
}
.rv-btn-ghost:hover { color: var(--text); border-color: var(--text); }
.rv-btn-danger:hover { color: var(--accent); border-color: var(--accent); }

.rv-hint {
  display: flex; align-items: center; gap: 8px;
  padding: 16px 18px;
  font-size: 13.5px;
  color: var(--text-muted);
  border: 1px dashed var(--border-strong);
  background: var(--bg-elevated);
  margin-bottom: 32px;
}
.rv-hint a { color: var(--accent); }
.rv-hint-mark {
  display: inline-block;
  width: 6px; height: 6px;
  background: var(--accent);
  border-radius: 50%;
  box-shadow: 0 0 8px var(--accent-glow);
  flex-shrink: 0;
}

/* list */
.rv-list { display: grid; gap: 18px; }
.rv-card {
  border: 1px solid var(--border);
  padding: 18px 20px;
  background: var(--bg-elevated);
  display: flex; flex-direction: column; gap: 8px;
  transition: border-color .15s ease;
}
.rv-card:hover { border-color: var(--border-strong); }
.rv-card--mine {
  border-color: var(--accent-soft);
  background: linear-gradient(180deg, var(--accent-haze), transparent 80%), var(--bg-elevated);
}
.rv-card-head {
  display: flex; align-items: center; justify-content: space-between;
  gap: 12px;
}
.rv-author {
  font-family: var(--font-mono);
  font-size: 11px; letter-spacing: 0.12em; text-transform: uppercase;
  color: var(--text-soft);
}
.rv-tag {
  font-family: var(--font-mono);
  font-size: 10px; letter-spacing: 0.18em; text-transform: uppercase;
  color: var(--accent);
  padding: 2px 8px;
  border: 1px solid var(--accent-soft);
}
.rv-text {
  margin: 0;
  font-size: 14.5px; line-height: 1.55;
  color: var(--text-soft);
  white-space: pre-line;
}
.rv-card-foot {
  display: flex; justify-content: space-between; align-items: center;
  gap: 12px;
  margin-top: 4px;
}
.rv-date {
  font-family: var(--font-mono);
  font-size: 10.5px; letter-spacing: 0.1em; text-transform: uppercase;
  color: var(--text-dim);
}
.rv-actions { display: flex; gap: 6px; }

@media (max-width: 720px) {
  .rv-head { grid-template-columns: 1fr; gap: 16px; }
  .rv-summary { justify-content: flex-start; text-align: left; }
  .rv-summary--empty { text-align: left; }
  .rv-avg { font-size: 44px; }
}
</style>

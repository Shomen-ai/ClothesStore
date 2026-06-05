<!--
  PaymentView - card-payment stub for a pending order (route: /payment/:id).
  Loads the order, runs client-side card validation (Luhn, expiry, CVV, holder),
  fakes a bank delay, then confirms payment on the server and redirects to
  /checkout-success. No real card data is ever sent to the backend.
-->
<template>
  <div class="pay-page" v-if="order">
    <!-- background marquee, keeps the brand identity on a "system" page -->
    <div class="head-marquee-wrap" aria-hidden="true">
      <div class="head-marquee">
        <div class="marquee-row row-1"><div class="marquee-track"><span>PAYMENT · SECURE · TLS · CARD · CVV · APPROVED · LOUD ·&nbsp;</span><span>PAYMENT · SECURE · TLS · CARD · CVV · APPROVED · LOUD ·&nbsp;</span></div></div>
        <div class="marquee-row row-2"><div class="marquee-track"><span>оплата — почти всё — последний шаг — your turn — secured —&nbsp;</span><span>оплата — почти всё — последний шаг — your turn — secured —&nbsp;</span></div></div>
        <div class="marquee-row row-3"><div class="marquee-track"><span>SHIP · TRACK · RECEIPT · WRAPPED · KEPT · SIGNATURE ·&nbsp;</span><span>SHIP · TRACK · RECEIPT · WRAPPED · KEPT · SIGNATURE ·&nbsp;</span></div></div>
      </div>
    </div>

    <header class="pay-head">
      <p class="eyebrow">Оплата заказа #{{ String(order.id).padStart(4,'0') }}</p>
      <h1 class="pay-title">Последний <span class="gothic-accent">штрих</span></h1>
      <div class="pay-meta">
        <span class="lock">
          <svg width="12" height="14" viewBox="0 0 12 14" fill="none">
            <rect x="1.5" y="6" width="9" height="7" rx="1" stroke="currentColor" stroke-width="1.2"/>
            <path d="M3.5 6V4a2.5 2.5 0 0 1 5 0v2" stroke="currentColor" stroke-width="1.2"/>
          </svg>
          Защищённое соединение
        </span>
        <span class="dot">·</span>
        <span>TLS / PCI-DSS demo</span>
      </div>
    </header>

    <div class="pay-grid">
      <!-- ── Summary ──────────────────────────────────────────────── -->
      <aside class="summary">
        <h2 class="sum-title">К оплате</h2>
        <div class="big-total">
          <span class="big-num">{{ formatPrice(order.total_price) }}</span>
          <span class="big-cur">RUB</span>
        </div>
        <ul class="items">
          <li v-for="it in order.items" :key="it.id">
            <span class="it-name">{{ productName(it.product_id) }}</span>
            <span class="it-qty">× {{ it.quantity }}</span>
            <span class="it-price">{{ formatPrice(it.price_at_order * it.quantity) }}</span>
          </li>
        </ul>
        <div v-if="order.discount_amount > 0" class="sum-row discount">
          <span>Скидка</span>
          <span>−{{ formatPrice(order.discount_amount) }}</span>
        </div>
        <div class="sum-row total">
          <span>Итого</span>
          <span>{{ formatPrice(order.total_price) }}</span>
        </div>
      </aside>

      <!-- ── Form ────────────────────────────────────────────────── -->
      <section class="card-form">
        <p class="eyebrow">Данные карты</p>

        <label class="field">
          <span class="field-label">Номер карты</span>
          <div class="card-number-wrap">
            <input ref="cardNumberRef"
                   v-model="display.cardNumber"
                   inputmode="numeric"
                   autocomplete="cc-number"
                   placeholder="0000 0000 0000 0000"
                   maxlength="19"
                   :class="{ invalid: cardBlurred && !luhnValid }"
                   @input="onCardInput"
                   @blur="cardBlurred = true" />
            <span class="brand-badge" :class="brand">{{ brandLabel }}</span>
          </div>
          <span class="hint" v-if="cardBlurred && !luhnValid">Неверный номер карты</span>
        </label>

        <div class="row-2">
          <label class="field">
            <span class="field-label">Срок (MM/YY)</span>
            <input v-model="display.expiry"
                   inputmode="numeric"
                   autocomplete="cc-exp"
                   placeholder="12/27"
                   maxlength="5"
                   :class="{ invalid: expiryBlurred && !expiryValid }"
                   @input="onExpiryInput"
                   @blur="expiryBlurred = true" />
            <span class="hint" v-if="expiryBlurred && !expiryValid">Проверь срок</span>
          </label>
          <label class="field">
            <span class="field-label">CVV</span>
            <input v-model="form.cvv"
                   inputmode="numeric"
                   autocomplete="cc-csc"
                   placeholder="•••"
                   maxlength="3"
                   :class="{ invalid: cvvBlurred && !cvvValid }"
                   @input="onCvvInput"
                   @blur="cvvBlurred = true" />
          </label>
        </div>

        <label class="field">
          <span class="field-label">Имя на карте</span>
          <input v-model="form.holder"
                 autocomplete="cc-name"
                 placeholder="IVAN IVANOV"
                 @input="form.holder = form.holder.toUpperCase()" />
        </label>

        <p v-if="error" class="form-error">{{ error }}</p>

        <button class="pay-btn" :disabled="!formValid || processing" @click="onPay">
          <span v-if="!processing">Оплатить {{ formatPrice(order.total_price) }}</span>
          <span v-else class="pay-btn-loading">
            <span class="spinner"></span>
            Обработка платежа…
          </span>
        </button>

        <p class="pay-footnote">
          <span>Тест-картa для отказа: <code>4000 0000 0000 0002</code></span>
        </p>
      </section>
    </div>

    <!-- ── Processing overlay (full-screen scrim) ───────────────── -->
    <Transition name="overlay">
      <div v-if="processing" class="overlay">
        <div class="overlay-card">
          <span class="ring">
            <span class="ring-dot"></span>
          </span>
          <p class="overlay-title">Обработка платежа</p>
          <p class="overlay-sub">не закрывай окно — мы держим карту в безопасности</p>
        </div>
      </div>
    </Transition>
  </div>

  <div v-else class="page-loading">
    <span class="serif-italic">loading</span>
    <span class="ellipsis">...</span>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getOrder, confirmPayment } from '@/api/orders.js'

const route = useRoute()
const router = useRouter()

const order = ref(null)
const form = reactive({ cvv: '', holder: '' })
const display = reactive({ cardNumber: '', expiry: '' })
const cardBlurred = ref(false)
const expiryBlurred = ref(false)
const cvvBlurred = ref(false)
const processing = ref(false)
const error = ref('')

// `form` holds fields submitted conceptually; `display` holds the formatted
// (spaced/slashed) strings bound to the inputs. Card number/expiry are never
// posted - confirmPayment only needs the order id.
onMounted(async () => {
  try {
    const { data } = await getOrder(route.params.id)
    order.value = data
  } catch (e) {
    error.value = 'Заказ не найден'
  }
})

// ─── card formatting + brand detection ─────────────────────────

const digitsOnly = (s) => (s || '').replace(/\D/g, '')

const cardDigits = computed(() => digitsOnly(display.cardNumber))

// Detect card brand from the leading digits to drive the badge styling.
const brand = computed(() => {
  const d = cardDigits.value
  if (!d) return 'unknown'
  if (/^4/.test(d))                              return 'visa'
  if (/^(5[1-5]|2[2-7])/.test(d))                return 'mc'
  if (/^220[0-4]/.test(d))                       return 'mir'
  return 'unknown'
})

const brandLabel = computed(() => ({
  visa: 'VISA', mc: 'MC', mir: 'МИР', unknown: '••••',
}[brand.value]))

// Strip non-digits, cap at 16, then re-group into 4-digit blocks for display.
function onCardInput() {
  const d = digitsOnly(display.cardNumber).slice(0, 16)
  display.cardNumber = d.replace(/(.{4})/g, '$1 ').trim()
}
function onExpiryInput() {
  const d = digitsOnly(display.expiry).slice(0, 4)
  display.expiry = d.length > 2 ? `${d.slice(0,2)}/${d.slice(2)}` : d
}
function onCvvInput() { form.cvv = digitsOnly(form.cvv).slice(0, 3) }

// ─── validators ───────────────────────────────────────────────

// Standard Luhn checksum: double every second digit from the right, subtract 9
// if >9, and the total must be divisible by 10.
function luhn(d) {
  if (!d || d.length < 13) return false
  let sum = 0, alt = false
  for (let i = d.length - 1; i >= 0; i--) {
    let n = +d[i]
    if (alt) { n *= 2; if (n > 9) n -= 9 }
    sum += n
    alt = !alt
  }
  return sum % 10 === 0
}
const luhnValid = computed(() => luhn(cardDigits.value))

// Expiry is valid only if MM is 1-12 and the last day of that month is not past.
const expiryValid = computed(() => {
  const m = display.expiry.match(/^(\d{2})\/(\d{2})$/)
  if (!m) return false
  const mm = +m[1], yy = +m[2]
  if (mm < 1 || mm > 12) return false
  const now = new Date()
  const exp = new Date(2000 + yy, mm, 0, 23, 59, 59) // last day of expiry month
  return exp >= now
})
const cvvValid = computed(() => form.cvv.length === 3)
const holderValid = computed(() => form.holder.trim().length >= 3)

// Enable the pay button only when every card field passes validation.
const formValid = computed(() => luhnValid.value && expiryValid.value && cvvValid.value && holderValid.value)

// ─── pay ──────────────────────────────────────────────────────

// Payment flow: guard -> fake 3s bank delay -> client-side decline check ->
// server confirm -> redirect. Uses router.replace so Back doesn't reopen pay.
async function onPay() {
  if (!formValid.value || processing.value) return
  error.value = ''
  processing.value = true

  await new Promise(r => setTimeout(r, 3000)) // realism: bank thinks for 3s

  // Test decline card — bank "declined" without backend roundtrip.
  if (cardDigits.value === '4000000000000002') {
    processing.value = false
    error.value = 'Карта отклонена банком'
    return
  }

  try {
    await confirmPayment(order.value.id)
    router.replace('/checkout-success')
  } catch (e) {
    error.value = e.response?.data?.error || 'Ошибка платежа'
  } finally {
    processing.value = false
  }
}

// ─── presentation ─────────────────────────────────────────────

function productName(_id) {
  // The /user/orders/:id payload doesn't join products; we surface a generic
  // line — the summary's job is "how much" not "what". Refactor later if a
  // joined endpoint appears.
  return 'Позиция заказа'
}
function formatPrice(p) {
  return new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 0, style: 'currency', currency: 'RUB' }).format(p)
}
</script>

<style scoped>
.pay-page {
  max-width: 1280px;
  margin: 0 auto;
  padding: 96px 32px 64px;
  position: relative;
}
.page-loading {
  min-height: 60vh;
  display: flex; align-items: center; justify-content: center;
  gap: 6px; color: var(--text-muted);
}
.page-loading .serif-italic { font-size: 28px; color: var(--text-soft); }
.ellipsis { font-family: var(--font-mono); }

/* head */
.pay-head { position: relative; margin-bottom: 40px; padding-bottom: 24px; border-bottom: 1px solid var(--border); isolation: isolate; }
.pay-head .eyebrow, .pay-head .pay-title, .pay-head .pay-meta { position: relative; z-index: 2; }
.eyebrow {
  font-family: var(--font-mono);
  font-size: 11px; letter-spacing: 0.18em; text-transform: uppercase;
  color: var(--text-muted);
  margin: 0 0 8px;
}
.pay-title {
  font-size: clamp(36px, 4.5vw, 56px);
  font-weight: 700;
  letter-spacing: -0.025em;
  line-height: 1;
  margin: 0 0 12px;
}
.pay-title .gothic-accent { font-size: 1.05em; }
.pay-meta {
  display: inline-flex; align-items: center; gap: 8px;
  font-family: var(--font-mono);
  font-size: 11px; letter-spacing: 0.1em; text-transform: uppercase;
  color: var(--text-muted);
}
.pay-meta .lock { display: inline-flex; align-items: center; gap: 6px; color: var(--accent); }
.pay-meta .dot { color: var(--text-dim); }

/* grid */
.pay-grid {
  display: grid;
  grid-template-columns: minmax(0, 0.8fr) minmax(0, 1fr);
  gap: 48px;
  align-items: start;
}

/* summary */
.summary {
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  padding: 28px 28px 24px;
  display: flex; flex-direction: column; gap: 18px;
  position: sticky; top: 96px;
}
.sum-title {
  font-family: var(--font-mono);
  font-size: 11px; letter-spacing: 0.18em; text-transform: uppercase;
  color: var(--text-muted);
  margin: 0; padding-bottom: 8px;
  border-bottom: 1px dashed var(--border);
}
.big-total {
  display: flex; align-items: baseline; gap: 6px;
  margin: 4px 0 8px;
}
.big-num {
  font-family: var(--font-mono);
  font-size: 44px;
  font-weight: 500;
  color: var(--text);
  letter-spacing: -0.02em;
  line-height: 1;
}
.big-cur {
  font-family: var(--font-mono);
  font-size: 12px;
  letter-spacing: 0.16em;
  color: var(--text-muted);
}
.items {
  list-style: none; padding: 0; margin: 0;
  display: flex; flex-direction: column; gap: 8px;
  border-top: 1px dashed var(--border);
  padding-top: 14px;
}
.items li {
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: 12px;
  font-size: 13px;
  color: var(--text-soft);
}
.items .it-qty { font-family: var(--font-mono); color: var(--text-dim); font-size: 11px; align-self: center; }
.items .it-price { font-family: var(--font-mono); color: var(--text); }

.sum-row { display: flex; justify-content: space-between; font-size: 13.5px; }
.sum-row.discount { color: var(--accent); }
.sum-row.total {
  font-size: 15px;
  font-weight: 600;
  color: var(--text);
  border-top: 1px dashed var(--border);
  padding-top: 14px;
  margin-top: 6px;
}

/* form */
.card-form {
  background: var(--bg-card, var(--bg-elevated));
  border: 1px solid var(--border-strong);
  padding: 32px;
  display: flex; flex-direction: column; gap: 18px;
  box-shadow: 0 1px 0 rgba(0,0,0,0.02), 0 22px 60px -28px rgba(0,0,0,0.18);
  position: relative;
}
.card-form::before {
  content: ''; position: absolute; top: 0; left: 0; right: 0; height: 3px;
  background: linear-gradient(90deg, var(--accent), transparent 80%);
}
.card-form .eyebrow { margin: 0 0 4px; }

.field { display: flex; flex-direction: column; gap: 6px; }
.field-label {
  font-family: var(--font-mono);
  font-size: 10.5px; letter-spacing: 0.18em; text-transform: uppercase;
  color: var(--text-muted);
}
.field input {
  width: 100%;
  background: transparent;
  border: 1px solid var(--border-strong);
  padding: 14px 16px;
  font: 500 16px/1 var(--font-mono);
  letter-spacing: 0.04em;
  color: var(--text);
  outline: none;
  transition: border-color .15s ease, box-shadow .15s ease;
}
.field input:focus { border-color: var(--accent); box-shadow: 0 0 0 3px var(--accent-haze); }
.field input.invalid { border-color: var(--accent); }
.hint { font-size: 11px; color: var(--accent); font-family: var(--font-mono); letter-spacing: 0.06em; }

.card-number-wrap { position: relative; }
.card-number-wrap input { padding-right: 64px; }
.brand-badge {
  position: absolute; top: 50%; right: 12px; transform: translateY(-50%);
  font-family: var(--font-mono);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.12em;
  padding: 4px 8px;
  border: 1px solid var(--border-strong);
  color: var(--text-dim);
  background: var(--bg-soft, transparent);
  transition: all .2s ease;
}
.brand-badge.visa    { color: #fff; background: #1a1f71; border-color: #1a1f71; }
.brand-badge.mc      { color: #fff; background: #eb001b; border-color: #eb001b; }
.brand-badge.mir     { color: #fff; background: #0f754e; border-color: #0f754e; }

.row-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }

.form-error {
  margin: 0;
  font-family: var(--font-mono);
  font-size: 12px;
  letter-spacing: 0.04em;
  color: var(--accent);
  padding: 10px 12px;
  background: var(--accent-haze);
  border: 1px dashed var(--accent);
}

.pay-btn {
  margin-top: 6px;
  height: 56px;
  background: var(--accent); color: #fff;
  border: 1px solid var(--accent);
  font: 600 13px/1 var(--font-sans);
  letter-spacing: 0.1em;
  text-transform: uppercase;
  cursor: pointer;
  transition: filter .15s ease, transform .15s ease;
  display: inline-flex; align-items: center; justify-content: center; gap: 10px;
}
.pay-btn:hover:not(:disabled) { filter: brightness(1.05); transform: translateY(-1px); }
.pay-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.pay-btn-loading { display: inline-flex; align-items: center; gap: 10px; }
.spinner {
  width: 14px; height: 14px;
  border: 2px solid rgba(255,255,255,0.35);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.pay-footnote {
  font-family: var(--font-mono);
  font-size: 10.5px; letter-spacing: 0.06em;
  color: var(--text-dim);
  margin: 0; text-align: center;
}
.pay-footnote code { color: var(--text-muted); }

/* overlay */
.overlay {
  position: fixed; inset: 0; z-index: 200;
  background: rgba(8,8,12,0.55);
  backdrop-filter: blur(8px);
  display: flex; align-items: center; justify-content: center;
}
.overlay-card {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-strong);
  padding: 36px 44px;
  display: flex; flex-direction: column; align-items: center; gap: 14px;
  box-shadow: 0 40px 80px -20px rgba(0,0,0,0.4);
  min-width: 320px;
}
.overlay-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text);
  margin: 0;
}
.overlay-sub {
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.06em;
  color: var(--text-muted);
  margin: 0;
  max-width: 28ch;
  text-align: center;
}
.ring {
  width: 54px; height: 54px;
  border: 2px solid var(--border-strong);
  border-top-color: var(--accent);
  border-radius: 50%;
  display: inline-block;
  position: relative;
  animation: spin 1.1s linear infinite;
}
.ring-dot {
  position: absolute; top: -4px; left: 50%;
  width: 6px; height: 6px;
  background: var(--accent);
  border-radius: 50%;
  transform: translateX(-50%);
  box-shadow: 0 0 10px var(--accent-glow);
}

.overlay-enter-active, .overlay-leave-active { transition: opacity .2s ease; }
.overlay-enter-from, .overlay-leave-to { opacity: 0; }

@media (max-width: 920px) {
  .pay-page { padding: 56px 16px 32px; }
  .pay-grid { grid-template-columns: 1fr; gap: 24px; }
  .summary { position: static; }
}
</style>

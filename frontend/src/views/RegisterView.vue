<!--
  RegisterView - two-step sign-up with email verification (route: /register).
  Step 1 collects name/email/phone/password and requests a 6-digit code
  (/auth/register/start). Step 2 enters the code (/auth/register/verify), which
  returns a session and logs the user in. A resend link is rate-limited locally.
-->
<template>
  <div class="auth-page">
    <div class="auth-card">
      <p class="eyebrow">Аккаунт / шаг {{ step }} из 2</p>

      <!-- step 1: form -->
      <template v-if="step === 1">
        <h1 class="auth-title">Стань <span class="gothic-accent">своим</span></h1>
        <p class="auth-sub">5 секунд — и ты в листе drop-уведомлений и закрытых распродаж.</p>

        <form class="auth-form" @submit.prevent="goToCode">
          <label class="field">
            <span class="field-label">Имя</span>
            <input v-model="form.name" placeholder="Как тебя зовут" required />
          </label>
          <label class="field">
            <span class="field-label">Email</span>
            <input v-model="form.email" type="email" placeholder="your@email.com" autocomplete="email" required />
          </label>
          <label class="field">
            <span class="field-label">Телефон</span>
            <input v-model="form.phone" placeholder="+7 (999) 999-99-99" autocomplete="tel" />
          </label>
          <label class="field">
            <span class="field-label">Пароль</span>
            <input v-model="form.password" type="password" placeholder="Минимум 6 символов" autocomplete="new-password" required minlength="6" />
          </label>

          <p v-if="error" class="auth-error">{{ error }}</p>

          <button class="auth-submit" type="submit" :disabled="loading">
            <span>{{ loading ? 'Отправляем код…' : 'Отправить код' }}</span>
            <span class="arrow" v-if="!loading">→</span>
          </button>
        </form>
      </template>

      <!-- step 2: email verification -->
      <template v-else>
        <h1 class="auth-title">Проверь <span class="gothic-accent">почту</span></h1>
        <p class="auth-sub">
          Мы отправили код на
          <span class="email-pill">{{ form.email }}</span>
          — введи 6 цифр, чтобы подтвердить email.
        </p>

        <form class="auth-form" @submit.prevent="confirmCode">
          <!-- Six single-char inputs; refs are collected so focus can hop between cells -->
          <div class="code-grid">
            <input v-for="(_, i) in 6" :key="i"
              :ref="el => codeInputs[i] = el"
              v-model="code[i]"
              class="code-cell"
              inputmode="numeric" maxlength="1"
              @input="onCodeInput(i, $event)"
              @keydown.backspace="onBackspace(i, $event)"
              @paste.prevent="onPaste"
            />
          </div>

          <p v-if="error" class="auth-error">{{ error }}</p>

          <button class="auth-submit" type="submit" :disabled="loading || codeStr.length !== 6">
            <span>{{ loading ? 'Создаём…' : 'Подтвердить и войти' }}</span>
            <span class="arrow">→</span>
          </button>
        </form>

        <p class="auth-hint">
          <button class="link-btn" type="button" @click="step = 1">← Изменить данные</button>
          <button class="link-btn link-btn--resend"
                  type="button"
                  :disabled="resendCountdown > 0 || resending"
                  @click="resendCode">
            {{ resendCountdown > 0
                ? `Отправить ещё раз через ${resendCountdown} с`
                : (resending ? 'Отправка…' : 'Отправить ещё раз') }}
          </button>
        </p>
      </template>

      <p class="auth-link">
        Уже есть аккаунт?
        <RouterLink to="/login">Войти</RouterLink>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, nextTick, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth.js'
import { registerStart, registerVerify, registerResend } from '@/api/auth.js'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const step = ref(1)
const form = ref({ name: '', email: '', phone: '', password: '' })
const code = ref(['', '', '', '', '', ''])
const codeInputs = []
const error = ref('')
const loading = ref(false)
const resending = ref(false)
const resendCountdown = ref(0)
let resendTimer = null

// The six digit-cells joined into one string for submission/validation.
const codeStr = computed(() => code.value.join(''))

// Local cooldown that disables the "resend" button after each send.
function startResendCountdown(seconds = 60) {
  resendCountdown.value = seconds
  clearInterval(resendTimer)
  resendTimer = setInterval(() => {
    resendCountdown.value -= 1
    if (resendCountdown.value <= 0) {
      clearInterval(resendTimer); resendTimer = null
    }
  }, 1000)
}

// Avoid a leaked interval if the user leaves mid-countdown.
onUnmounted(() => { if (resendTimer) clearInterval(resendTimer) })

// Step 1 submit: validate locally, ask the server to email a code, then advance
// to step 2 (reset cells, start cooldown, focus the first cell).
async function goToCode() {
  error.value = ''
  if (form.value.password.length < 6) {
    error.value = 'Пароль должен быть не короче 6 символов'
    return
  }
  if (!form.value.name.trim()) {
    error.value = 'Укажи имя'
    return
  }
  loading.value = true
  try {
    await registerStart({
      name: form.value.name.trim(),
      email: form.value.email.trim().toLowerCase(),
      phone: form.value.phone.trim(),
      password: form.value.password,
    })
    step.value = 2
    code.value = ['', '', '', '', '', '']
    startResendCountdown(60)
    nextTick(() => codeInputs[0]?.focus())
  } catch (e) {
    error.value = e.response?.data?.error || 'Не удалось отправить код'
  } finally {
    loading.value = false
  }
}

// Keep only the last typed digit per cell and auto-advance focus.
function onCodeInput(i, e) {
  const v = e.target.value.replace(/\D/g, '').slice(-1)
  code.value[i] = v
  if (v && i < 5) codeInputs[i + 1]?.focus()
}

// Backspace on an empty cell clears and focuses the previous one.
function onBackspace(i, e) {
  if (!code.value[i] && i > 0) {
    codeInputs[i - 1]?.focus()
    code.value[i - 1] = ''
    e.preventDefault()
  }
}

// Paste a full code at once: spread digits across the six cells.
function onPaste(e) {
  const text = (e.clipboardData?.getData('text') || '').replace(/\D/g, '').slice(0, 6)
  if (!text) return
  for (let i = 0; i < 6; i++) code.value[i] = text[i] || ''
  const last = Math.min(text.length, 5)
  codeInputs[last]?.focus()
}

// Step 2 submit: verify the code; on success the server returns {user, tokens},
// which setSession persists, then we redirect (honoring ?redirect=).
async function confirmCode() {
  error.value = ''
  if (codeStr.value.length !== 6) {
    error.value = 'Введи 6 цифр'
    return
  }
  loading.value = true
  try {
    const { data } = await registerVerify({
      email: form.value.email.trim().toLowerCase(),
      code: codeStr.value,
    })
    auth.setSession(data)
    const redirect = route.query.redirect
    router.push(redirect || '/account')
  } catch (e) {
    error.value = e.response?.data?.error || 'Не удалось завершить регистрацию'
  } finally {
    loading.value = false
  }
}

// Request a fresh code, respecting the local cooldown, and restart it.
async function resendCode() {
  if (resendCountdown.value > 0 || resending.value) return
  resending.value = true
  error.value = ''
  try {
    await registerResend(form.value.email.trim().toLowerCase())
    startResendCountdown(60)
  } catch (e) {
    error.value = e.response?.data?.error || 'Не удалось отправить код'
  } finally {
    resending.value = false
  }
}
</script>

<style scoped>
.auth-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: calc(100vh - 80px);
  padding: 96px 24px 64px;
}
.auth-card {
  width: 100%;
  max-width: 460px;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  padding: 44px 40px 36px;
  backdrop-filter: blur(8px);
  position: relative;
}
.auth-card::before {
  content: '';
  position: absolute; top: -1px; left: -1px; width: 28px; height: 28px;
  border-top: 2px solid var(--accent);
  border-left: 2px solid var(--accent);
  pointer-events: none;
}
.auth-card::after {
  content: '';
  position: absolute; bottom: -1px; right: -1px; width: 28px; height: 28px;
  border-bottom: 2px solid var(--accent);
  border-right: 2px solid var(--accent);
  pointer-events: none;
}

.auth-title {
  font-size: clamp(36px, 4.5vw, 48px);
  font-weight: 700;
  letter-spacing: -0.02em;
  line-height: 1;
  margin: 12px 0 12px;
}
.auth-title .gothic-accent { font-size: 1.05em; }
.auth-sub {
  color: var(--text-muted);
  font-size: 14px;
  line-height: 1.55;
  margin-bottom: 24px;
}
.email-pill {
  display: inline-block;
  font-family: var(--font-mono);
  font-size: 12px;
  letter-spacing: 0.04em;
  color: var(--text);
  background: var(--bg-elevated);
  padding: 2px 8px;
  border: 1px solid var(--border-strong);
  margin: 0 2px;
}

.auth-form { display: flex; flex-direction: column; gap: 16px; }
.field { display: flex; flex-direction: column; gap: 6px; }
.field-label {
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--text-muted);
}
.field input {
  width: 100%;
  background: transparent;
  border: 0;
  border-bottom: 1px solid var(--border-strong);
  color: var(--text);
  padding: 10px 0;
  font-size: 15px;
  font-family: var(--font-sans);
  outline: none;
  transition: border-color .2s ease;
}
.field input:focus { border-bottom-color: var(--accent); }
.field input::placeholder { color: var(--text-dim); }

.code-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 10px;
  margin-bottom: 4px;
}
.code-cell {
  width: 100%;
  height: 56px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-strong);
  color: var(--text);
  font: 600 22px/1 var(--font-mono);
  text-align: center;
  outline: none;
  transition: border-color .15s ease, box-shadow .15s ease, background .15s ease;
}
.code-cell:focus {
  border-color: var(--accent);
  background: var(--bg-surface);
  box-shadow: inset 0 -2px 0 var(--accent);
}

.auth-error {
  font-size: 13px;
  color: var(--danger);
  background: rgba(239, 72, 72, 0.08);
  border: 1px solid rgba(239, 72, 72, 0.3);
  padding: 8px 12px;
  margin-top: -4px;
}

.auth-submit {
  margin-top: 12px;
  display: inline-flex; align-items: center; justify-content: center; gap: 12px;
  background: var(--accent);
  color: #fff;
  border: 0;
  padding: 14px 20px;
  font: 500 12px/1 var(--font-mono);
  letter-spacing: 0.18em;
  text-transform: uppercase;
  cursor: pointer;
  transition: background .2s ease, transform .15s ease, box-shadow .25s ease;
}
.auth-submit:hover:not(:disabled) {
  background: var(--accent-soft);
  box-shadow: 0 14px 32px -10px var(--accent-glow);
  transform: translateY(-1px);
}
.auth-submit:disabled { opacity: 0.5; cursor: not-allowed; }
.auth-submit .arrow { font-family: var(--font-mono); }

.auth-hint {
  margin-top: 18px;
  display: flex; justify-content: space-between; align-items: center;
  font-size: 12px;
  color: var(--text-muted);
}
.link-btn {
  background: transparent; border: 0; padding: 0; cursor: pointer;
  color: var(--text-muted); font-family: var(--font-mono); font-size: 11px;
  letter-spacing: 0.1em; text-transform: uppercase;
}
.link-btn:hover { color: var(--text); }
.hint-meta {
  font-family: var(--font-mono); font-size: 11px;
  letter-spacing: 0.06em; color: var(--text-dim);
}
.hint-meta code {
  color: var(--accent); background: var(--bg-elevated);
  padding: 2px 6px; border: 1px solid var(--border-strong);
}

.auth-link {
  margin-top: 24px;
  text-align: center;
  color: var(--text-muted);
  font-size: 13px;
}
.auth-link a {
  color: var(--accent);
  font-weight: 500;
  border-bottom: 1px solid currentColor;
  padding-bottom: 1px;
}
.auth-link a:hover { color: var(--accent-soft); }
</style>

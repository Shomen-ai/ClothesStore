<template>
  <div class="auth-page">
    <div class="auth-card">
      <p class="eyebrow">Аккаунт</p>
      <h1 class="auth-title">Войди <span class="gothic-accent">громко</span></h1>
      <p class="auth-sub">Только участники получают drop-уведомления раньше всех.</p>

      <form class="auth-form" @submit.prevent="handleLogin">
        <label class="field">
          <span class="field-label">Email</span>
          <input v-model="form.email" type="email" placeholder="your@email.com" autocomplete="email" required />
        </label>
        <label class="field">
          <span class="field-label">Пароль</span>
          <input v-model="form.password" type="password" placeholder="••••••" autocomplete="current-password" required />
        </label>

        <p v-if="error" class="auth-error">{{ error }}</p>

        <button class="auth-submit" type="submit" :disabled="loading">
          <span>{{ loading ? 'Входим…' : 'Войти' }}</span>
          <span class="arrow">→</span>
        </button>
      </form>

      <p class="auth-link">
        Нет аккаунта?
        <RouterLink to="/register">Регистрация</RouterLink>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth.js'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const form = ref({ email: '', password: '' })
const error = ref('')
const loading = ref(false)

async function handleLogin() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(form.value.email, form.value.password)
    const redirect = route.query.redirect
    if (redirect) router.push(redirect)
    else router.push(auth.user?.role === 'admin' ? '/admin' : '/account')
  } catch (e) {
    error.value = e.response?.data?.error || 'Неверный email или пароль'
  } finally {
    loading.value = false
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
  max-width: 440px;
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
  margin-bottom: 28px;
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
.auth-submit:hover {
  background: var(--accent-soft);
  box-shadow: 0 14px 32px -10px var(--accent-glow);
  transform: translateY(-1px);
}
.auth-submit:disabled { opacity: 0.7; cursor: progress; }
.auth-submit .arrow { font-family: var(--font-mono); }

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

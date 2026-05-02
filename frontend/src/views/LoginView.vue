<template>
  <div class="auth-page">
    <div class="auth-card">
      <h1 class="auth-title">Вход</h1>
      <el-form :model="form" @submit.prevent="handleLogin" label-position="top">
        <el-form-item label="Email">
          <el-input v-model="form.email" type="email" placeholder="your@email.com" />
        </el-form-item>
        <el-form-item label="Пароль">
          <el-input v-model="form.password" type="password" placeholder="••••••" show-password />
        </el-form-item>
        <el-alert v-if="error" :title="error" type="error" :closable="false" style="margin-bottom:16px" />
        <el-button type="primary" native-type="submit" :loading="loading" style="width:100%">
          Войти
        </el-button>
      </el-form>
      <p class="auth-link">Нет аккаунта? <RouterLink to="/register">Зарегистрироваться</RouterLink></p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth.js'

const auth = useAuthStore()
const router = useRouter()
const form = ref({ email: '', password: '' })
const error = ref('')
const loading = ref(false)

async function handleLogin() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(form.value.email, form.value.password)
    router.push(auth.user?.role === 'admin' ? '/admin' : '/account')
  } catch (e) {
    error.value = e.response?.data?.error || 'Ошибка входа'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page { display: flex; align-items: center; justify-content: center; min-height: 80vh; padding: 24px; }
.auth-card { width: 100%; max-width: 400px; background: var(--color-bg-surface); border: 1px solid var(--color-border); border-radius: 8px; padding: 40px; }
.auth-title { font-size: 24px; font-weight: 700; margin-bottom: 32px; }
.auth-link { margin-top: 20px; text-align: center; color: var(--color-text-muted); font-size: 14px; }
</style>

<template>
  <div class="auth-page">
    <div class="auth-card">
      <h1 class="auth-title">Регистрация</h1>
      <el-form :model="form" @submit.prevent="handleRegister" label-position="top">
        <el-form-item label="Имя">
          <el-input v-model="form.name" placeholder="Ваше имя" />
        </el-form-item>
        <el-form-item label="Email">
          <el-input v-model="form.email" type="email" placeholder="your@email.com" />
        </el-form-item>
        <el-form-item label="Телефон">
          <el-input v-model="form.phone" placeholder="+7 (999) 999-99-99" />
        </el-form-item>
        <el-form-item label="Пароль">
          <el-input v-model="form.password" type="password" show-password placeholder="Минимум 6 символов" />
        </el-form-item>
        <el-alert v-if="error" :title="error" type="error" :closable="false" style="margin-bottom:16px" />
        <el-button type="primary" native-type="submit" :loading="loading" style="width:100%">
          Создать аккаунт
        </el-button>
      </el-form>
      <p class="auth-link">Уже есть аккаунт? <RouterLink to="/login">Войти</RouterLink></p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth.js'

const auth = useAuthStore()
const router = useRouter()
const form = ref({ name: '', email: '', phone: '', password: '' })
const error = ref('')
const loading = ref(false)

async function handleRegister() {
  error.value = ''
  loading.value = true
  try {
    await auth.register(form.value)
    router.push('/account')
  } catch (e) {
    error.value = e.response?.data?.error || 'Ошибка регистрации'
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

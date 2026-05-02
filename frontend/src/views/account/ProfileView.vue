<template>
  <div>
    <h2 class="section-title">ПРОФИЛЬ</h2>
    <el-form :model="form" label-position="top" style="max-width:480px">
      <el-form-item label="Имя"><el-input v-model="form.name" /></el-form-item>
      <el-form-item label="Email"><el-input :value="auth.user?.email" disabled /></el-form-item>
      <el-form-item label="Телефон"><el-input v-model="form.phone" /></el-form-item>
      <el-form-item label="Новый пароль (оставьте пустым, если не меняете)">
        <el-input v-model="form.password" type="password" show-password />
      </el-form-item>
      <el-button type="primary" :loading="saving" @click="save">Сохранить</el-button>
    </el-form>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getProfile, updateProfile } from '@/api/user.js'
import { useAuthStore } from '@/stores/auth.js'
import { ElMessage } from 'element-plus'

const auth = useAuthStore()
const form = ref({ name: '', phone: '', password: '' })
const saving = ref(false)

onMounted(async () => {
  const { data } = await getProfile()
  form.value.name = data.name || ''
  form.value.phone = data.phone || ''
})

async function save() {
  saving.value = true
  try {
    await updateProfile(form.value)
    ElMessage({ message: 'Сохранено', type: 'success' })
  } catch {
    ElMessage.error('Ошибка сохранения')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.section-title { font-size: 20px; font-weight: 700; letter-spacing: 3px; margin-bottom: 32px; }
</style>

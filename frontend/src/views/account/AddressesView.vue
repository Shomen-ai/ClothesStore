<!--
  AddressesView — customer's saved shipping addresses (route: /account/addresses).
  Lists addresses and provides a single dialog reused for both create and edit, plus delete.
-->
<template>
  <div>
    <div class="view-header">
      <h2 class="section-title">АДРЕСА</h2>
      <el-button type="primary" @click="openForm(null)">+ Добавить</el-button>
    </div>

    <div v-for="a in addresses" :key="a.id" class="address-card">
      <div class="address-text">
        <span v-if="a.is_default" class="default-tag">По умолчанию</span>
        {{ a.zip_code }}, г. {{ a.city }}, {{ a.street }}, д. {{ a.house }}
        <span v-if="a.apartment">, кв. {{ a.apartment }}</span>
      </div>
      <div class="address-actions">
        <el-button size="small" @click="openForm(a)">Изменить</el-button>
        <el-button size="small" type="danger" @click="remove(a.id)">Удалить</el-button>
      </div>
    </div>

    <el-dialog v-model="dialogOpen" :title="editingAddr ? 'Редактировать адрес' : 'Новый адрес'" width="480px">
      <el-form :model="form" label-position="top">
        <el-form-item label="Город"><el-input v-model="form.city" /></el-form-item>
        <el-form-item label="Улица"><el-input v-model="form.street" /></el-form-item>
        <el-form-item label="Дом"><el-input v-model="form.house" /></el-form-item>
        <el-form-item label="Квартира"><el-input v-model="form.apartment" /></el-form-item>
        <el-form-item label="Индекс"><el-input v-model="form.zip_code" /></el-form-item>
        <el-form-item>
          <el-checkbox v-model="form.is_default">Сделать адресом по умолчанию</el-checkbox>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogOpen = false">Отмена</el-button>
        <el-button type="primary" @click="save">Сохранить</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getAddresses, createAddress, updateAddress, deleteAddress } from '@/api/user.js'

const addresses = ref([])
const dialogOpen = ref(false)
const editingAddr = ref(null)
const form = ref({ city:'', street:'', house:'', apartment:'', zip_code:'', is_default: false })

onMounted(load)

// Fetch the current user's addresses; the server scopes them to the authenticated user.
async function load() { const { data } = await getAddresses(); addresses.value = data || [] }

// Open the shared dialog. addr=null => create mode; otherwise edit a clone of the row
// (spread copy so editing the form doesn't mutate the list before save).
function openForm(addr) {
  editingAddr.value = addr
  form.value = addr ? { ...addr } : { city:'', street:'', house:'', apartment:'', zip_code:'', is_default: false }
  dialogOpen.value = true
}

// Same handler for create and update, branching on editingAddr; re-fetch on success.
async function save() {
  if (editingAddr.value) await updateAddress(editingAddr.value.id, form.value)
  else await createAddress(form.value)
  dialogOpen.value = false
  load()
}

async function remove(id) {
  await deleteAddress(id)
  load()
}
</script>

<style scoped>
.section-title { font-size: 20px; font-weight: 700; letter-spacing: 3px; }
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 32px; }
.address-card { display: flex; justify-content: space-between; align-items: center; padding: 16px; border: 1px solid var(--color-border); margin-bottom: 8px; }
.default-tag { background: var(--color-red-dark); color: white; font-size: 10px; padding: 2px 8px; margin-right: 8px; }
.address-actions { display: flex; gap: 8px; }
</style>

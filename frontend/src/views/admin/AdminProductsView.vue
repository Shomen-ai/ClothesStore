<template>
  <div>
    <div class="view-header">
      <h2 class="page-title">ТОВАРЫ</h2>
      <el-button type="primary" @click="openCreate">+ Добавить товар</el-button>
    </div>

    <el-table :data="products" style="width:100%">
      <el-table-column width="80">
        <template #default="{ row }">
          <img v-if="row.images?.[0]" :src="row.images[0].image_path" style="width:48px;height:64px;object-fit:cover" />
          <div v-else style="width:48px;height:64px;background:var(--color-bg-elevated);" />
        </template>
      </el-table-column>
      <el-table-column prop="name" label="Название" />
      <el-table-column label="Цена" width="120">
        <template #default="{ row }">{{ formatPrice(row.price) }}</template>
      </el-table-column>
      <el-table-column label="Активен" width="100">
        <template #default="{ row }">
          <el-switch :model-value="row.is_active" @change="toggleActive(row)" />
        </template>
      </el-table-column>
      <el-table-column label="Действия" width="180">
        <template #default="{ row }">
          <el-button size="small" @click="openEdit(row)">Изменить</el-button>
          <el-button size="small" type="danger" @click="remove(row.id)">Удалить</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogOpen" :title="editingId ? 'Редактировать товар' : 'Новый товар'" width="600px">
      <ProductForm v-model="formData" :categories="categories" :product-id="editingId"
        :images="editingImages" @files-changed="pendingFiles = $event" />
      <template #footer>
        <el-button @click="dialogOpen = false">Отмена</el-button>
        <el-button type="primary" :loading="saving" @click="save">Сохранить</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { adminListProducts, adminCreateProduct, adminUpdateProduct, adminDeleteProduct, adminUploadImages, adminGetCategories } from '@/api/admin.js'
import ProductForm from '@/components/admin/ProductForm.vue'
import { ElMessageBox } from 'element-plus'

const products = ref([])
const categories = ref([])
const dialogOpen = ref(false)
const editingId = ref(null)
const editingImages = ref([])
const saving = ref(false)
const pendingFiles = ref([])
const emptyForm = () => ({ name:'', description:'', category_id:null, price:0, is_active:true, sizes:[] })
const formData = ref(emptyForm())

onMounted(async () => {
  const [p, c] = await Promise.all([adminListProducts(), adminGetCategories()])
  products.value = p.data || []
  categories.value = c.data || []
})

function openCreate() {
  editingId.value = null; editingImages.value = []; pendingFiles.value = []
  formData.value = emptyForm()
  dialogOpen.value = true
}

function openEdit(p) {
  editingId.value = p.id; editingImages.value = p.images || []; pendingFiles.value = []
  formData.value = { ...p }
  dialogOpen.value = true
}

async function save() {
  saving.value = true
  try {
    let id = editingId.value
    if (id) {
      await adminUpdateProduct(id, formData.value)
    } else {
      const { data } = await adminCreateProduct(formData.value)
      id = data.id
    }
    if (pendingFiles.value.length) {
      const form = new FormData()
      pendingFiles.value.forEach(f => form.append('images', f))
      await adminUploadImages(id, form)
    }
    dialogOpen.value = false
    const { data } = await adminListProducts()
    products.value = data || []
  } finally {
    saving.value = false
  }
}

async function toggleActive(row) {
  await adminUpdateProduct(row.id, { ...row, is_active: !row.is_active })
  row.is_active = !row.is_active
}

async function remove(id) {
  await ElMessageBox.confirm('Удалить товар?', 'Подтверждение', { type: 'warning' })
  await adminDeleteProduct(id)
  products.value = products.value.filter(p => p.id !== id)
}

function formatPrice(p) {
  return new Intl.NumberFormat('ru-RU',{style:'currency',currency:'RUB',maximumFractionDigits:0}).format(p)
}
</script>

<style scoped>
.view-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-title { font-size: 20px; font-weight: 800; letter-spacing: 3px; }
</style>

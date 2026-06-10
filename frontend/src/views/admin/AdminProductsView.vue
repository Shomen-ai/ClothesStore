<!--
  AdminProductsView — admin product catalog CRUD (route: /admin/products).
  Table of products with inline active-toggle and sale flag, plus a dialog (reused for
  create/edit) wrapping ProductForm. On save, the product is created/updated and any
  newly picked images are uploaded as multipart. Admin role is enforced server-side.
-->
<template>
  <div>
    <header class="admin-view-head">
      <div>
        <p class="eyebrow">Админка / Каталог</p>
        <h1 class="head-title">Товары <span class="gothic-accent">витрина</span></h1>
      </div>
      <button class="cta-add" @click="openCreate">
        <span class="cta-mark">+</span>
        <span class="cta-label">Новый товар</span>
        <span class="cta-arrow">→</span>
      </button>
    </header>

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
          <!-- One-way bind + @change so the toggle reflects state only after the PUT lands -->
          <el-switch :model-value="row.is_active" @change="toggleActive(row)" />
        </template>
      </el-table-column>
      <el-table-column label="Sale" width="80">
        <template #default="{ row }">
          <el-tag v-if="row.is_on_sale" type="danger" size="small">SALE</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="Действия" width="140">
        <template #default="{ row }">
          <div class="row-actions">
            <el-button size="small" @click="openEdit(row)">Изменить</el-button>
            <el-button size="small" type="danger" @click="remove(row.id)">Удалить</el-button>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <!-- Single dialog reused for create (editingId=null) and edit. ProductForm emits
         files-changed with the list of newly picked File objects to upload on save. -->
    <el-dialog v-model="dialogOpen" :title="editingId ? 'Редактировать товар' : 'Новый товар'" width="600px">
      <!-- :key forces a fresh ProductForm instance per open so its internal state
           is re-seeded cleanly and never echoes modelValue back into itself. -->
      <ProductForm :key="formKey" v-model="formData" :categories="categories" :product-id="editingId"
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
// Bumped on every dialog open to remount ProductForm with fresh internal state.
const formKey = ref(0)
// Factory for a blank product form (used for create and when resetting the dialog).
const emptyForm = () => ({ name:'', description:'', category_id:null, price:0, is_active:true, is_on_sale:false, sizes:[] })
const formData = ref(emptyForm())

onMounted(async () => {
  // Fetch products and categories in parallel for the table and the form's category select.
  const [p, c] = await Promise.all([adminListProducts(), adminGetCategories()])
  products.value = p.data || []
  categories.value = c.data || []
})

// Open the dialog in create mode: clear edit id, existing images, and pending uploads.
function openCreate() {
  editingId.value = null; editingImages.value = []; pendingFiles.value = []
  formData.value = emptyForm()
  formKey.value++
  dialogOpen.value = true
}

// Open the dialog in edit mode: seed it with a clone of the row (so live edits don't
// mutate the table) and pass through the product's existing images.
function openEdit(p) {
  editingId.value = p.id; editingImages.value = p.images || []; pendingFiles.value = []
  formData.value = { ...p }
  formKey.value++
  dialogOpen.value = true
}

// Save flow: update or create the product first to obtain an id, then (if any new files
// were picked) upload them as multipart against that id, and finally re-fetch the list.
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

// Inline active toggle: PUT the whole row with the flipped flag, then update local state.
async function toggleActive(row) {
  await adminUpdateProduct(row.id, { ...row, is_active: !row.is_active })
  row.is_active = !row.is_active
}

// Confirm, delete on the server, then optimistically drop the row from the list.
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
.row-actions {
  display: flex;
  flex-direction: column;
  gap: 6px;
  align-items: stretch;
}
.row-actions :deep(.el-button + .el-button) { margin-left: 0; }
.row-actions :deep(.el-button) { width: 100%; }
</style>

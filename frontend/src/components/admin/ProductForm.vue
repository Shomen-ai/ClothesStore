<!--
  ProductForm: admin create/edit form for a product (name, description, category,
  price, active/sale flags, per-size stock, and images). Acts as a v-model component:
  reads the product via `modelValue` and emits `update:modelValue` with the assembled
  product (including a derived `sizes` array). Image upload is deferred — selected files
  are emitted via `files-changed` for the parent to upload on save; existing images are
  deletable inline (only available once `productId` exists, i.e. in edit mode).
-->
<template>
  <el-form :model="form" label-position="top">
    <el-form-item label="Название"><el-input v-model="form.name" /></el-form-item>
    <el-form-item label="Описание"><el-input v-model="form.description" type="textarea" :rows="4" /></el-form-item>
    <el-form-item label="Категория">
      <el-select v-model="form.category_id" style="width:100%">
        <el-option v-for="c in categories" :key="c.id" :value="c.id" :label="c.name" />
      </el-select>
    </el-form-item>
    <el-form-item label="Цена (₽)"><el-input-number v-model="form.price" :min="0" /></el-form-item>
    <el-form-item label="Активен"><el-switch v-model="form.is_active" /></el-form-item>
    <el-form-item label="Распродажа (SALE)"><el-switch v-model="form.is_on_sale" /></el-form-item>

    <el-form-item label="Размеры и остатки">
      <div class="size-grid">
        <div v-for="s in sizeOrder" :key="s" class="size-row">
          <span class="size-label">{{ s }}</span>
          <el-input-number v-model="sizeQty[s]" :min="0" size="small" />
        </div>
      </div>
    </el-form-item>

    <!-- Photos only after the product exists (edit mode), so uploads/deletes have a target id -->
    <el-form-item v-if="productId" label="Фотографии">
      <div class="images-grid">
        <div v-for="img in localImages" :key="img.id" class="img-thumb">
          <img :src="img.image_path" />
          <button class="del-img" @click="deleteImg(img.id)"><el-icon><Close /></el-icon></button>
        </div>
      </div>
      <!-- auto-upload disabled: files are collected and handed to the parent to upload on save -->
      <el-upload multiple :auto-upload="false" :on-change="onFileChange" accept="image/*">
        <el-button>Загрузить фото</el-button>
      </el-upload>
    </el-form-item>
  </el-form>
</template>

<script setup>
import { ref, watch } from 'vue'
import { adminDeleteImage } from '@/api/admin.js'

// modelValue: the product being edited; categories: select options;
// productId: present only in edit mode (enables image management);
// images: existing product images to render/delete.
const props = defineProps({ modelValue: Object, categories: Array, productId: Number, images: Array })
const emit = defineEmits(['update:modelValue', 'files-changed'])

// Standard size set offered for every product (note: the data uses "2XL", not "XXL").
const STD_SIZES = ['XS', 'S', 'M', 'L', 'XL', '2XL']
// Local working copy of the product fields (spread so we don't mutate the prop object).
const form = ref({ ...props.modelValue })
// Stock quantity per size, kept as a flat {size: qty} map for easy binding to the inputs.
const sizeQty = ref({})
// Ordered list of size rows: standard sizes first, then any non-standard sizes the
// product actually has (e.g. "42-44" for socks), so real stock is always editable.
const sizeOrder = ref([...STD_SIZES])
// Sizes the product already had on open. These are always sent on save (even at 0)
// so the admin can set a size out of stock — the backend upserts (never deletes),
// and filtering 0 out would otherwise leave the old stock untouched.
const originalSizes = ref(new Set())
const localImages = ref(props.images || [])

// Seed the size inputs from the product once, at creation. The parent remounts
// this component (via :key) whenever a different product/create is opened, so
// this runs fresh for each one.
//
// We intentionally do NOT watch props.modelValue to re-seed `form`: the parent
// binds it with v-model and the watchers below emit our edits back up, so
// re-seeding `form` from that echo created an infinite update loop — every edit
// bounced modelValue→form→emit→modelValue→… until Vue threw "Maximum recursive
// updates exceeded" and the page froze.
syncSizes(props.modelValue?.sizes)

watch(() => props.images, v => { localImages.value = v || [] })

// Any edit to a field or a size quantity re-emits the full product (with the derived
// `sizes` array) so the parent's bound model stays current. Two watchers because `form`
// and `sizeQty` are separate refs feeding the same emitted payload.
watch(form, v => emit('update:modelValue', { ...v, sizes: buildSizes() }), { deep: true })
watch(sizeQty, () => emit('update:modelValue', { ...form.value, sizes: buildSizes() }), { deep: true })

// Reset all size inputs to 0, then fill from the product's existing sizes (ignoring any
// unknown size keys not in allSizes).
function syncSizes(sizes) {
  const map = {}
  STD_SIZES.forEach(s => { map[s] = 0 })
  ;(sizes || []).forEach(s => { map[s.size] = s.stock_qty }) // include every actual size, even non-standard
  sizeQty.value = map
  originalSizes.value = new Set((sizes || []).map(s => s.size))
  const extra = (sizes || []).map(s => s.size).filter(s => !STD_SIZES.includes(s))
  sizeOrder.value = [...STD_SIZES, ...new Set(extra)]
}

// Collapse the {size: qty} map into the API's [{size, stock_qty}] shape, dropping zero-stock sizes.
function buildSizes() {
  // Send a size if it has stock OR it already existed on the product (so setting it
  // to 0 persists as out-of-stock instead of being silently dropped). New standard
  // sizes left at 0 are skipped so we don't create phantom out-of-stock rows.
  return Object.entries(sizeQty.value)
    .filter(([size, qty]) => qty > 0 || originalSizes.value.has(size))
    .map(([size, stock_qty]) => ({ size, stock_qty }))
}

// el-upload fires per file; emit the full current selection (raw File objects) each time.
function onFileChange(file, fileList) {
  emit('files-changed', fileList.map(f => f.raw))
}

// Delete an existing image immediately via the admin API, then drop it from the local list.
async function deleteImg(imgId) {
  await adminDeleteImage(props.productId, imgId)
  localImages.value = localImages.value.filter(i => i.id !== imgId)
}
</script>

<style scoped>
.size-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); gap: 10px 16px; width: 100%; }
.size-row { display: flex; align-items: center; gap: 10px; }
.size-label { width: 40px; flex-shrink: 0; font-size: 13px; }
.size-row :deep(.el-input-number) { width: 100%; }
.images-grid { display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 12px; }
.img-thumb { position: relative; width: 80px; height: 100px; }
.img-thumb img { width: 100%; height: 100%; object-fit: cover; }
.del-img { position: absolute; top: 2px; right: 2px; background: rgba(0,0,0,0.7); border: none; cursor: pointer; color: white; padding: 2px; display: flex; }
</style>

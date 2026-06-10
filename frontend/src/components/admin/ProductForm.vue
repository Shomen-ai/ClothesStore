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
      <div v-for="s in allSizes" :key="s" class="size-row">
        <span class="size-label">{{ s }}</span>
        <el-input-number v-model="sizeQty[s]" :min="0" size="small" />
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

const allSizes = ['XS', 'S', 'M', 'L', 'XL', 'XXL']
// Local working copy of the product fields (spread so we don't mutate the prop object).
const form = ref({ ...props.modelValue })
// Stock quantity per size, kept as a flat {size: qty} map for easy binding to the inputs.
const sizeQty = ref(Object.fromEntries(allSizes.map(s => [s, 0])))
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
  allSizes.forEach(s => { sizeQty.value[s] = 0 })
  if (!sizes) return
  sizes.forEach(s => { if (s.size in sizeQty.value) sizeQty.value[s.size] = s.stock_qty })
}

// Collapse the {size: qty} map into the API's [{size, stock_qty}] shape, dropping zero-stock sizes.
function buildSizes() {
  return Object.entries(sizeQty.value)
    .filter(([, qty]) => qty > 0)
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
.size-row { display: flex; align-items: center; gap: 12px; margin-bottom: 8px; }
.size-label { width: 36px; font-size: 13px; }
.images-grid { display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 12px; }
.img-thumb { position: relative; width: 80px; height: 100px; }
.img-thumb img { width: 100%; height: 100%; object-fit: cover; }
.del-img { position: absolute; top: 2px; right: 2px; background: rgba(0,0,0,0.7); border: none; cursor: pointer; color: white; padding: 2px; display: flex; }
</style>

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

    <el-form-item label="Размеры и остатки">
      <div v-for="s in allSizes" :key="s" class="size-row">
        <span class="size-label">{{ s }}</span>
        <el-input-number v-model="sizeQty[s]" :min="0" size="small" />
      </div>
    </el-form-item>

    <el-form-item v-if="productId" label="Фотографии">
      <div class="images-grid">
        <div v-for="img in localImages" :key="img.id" class="img-thumb">
          <img :src="img.image_path" />
          <button class="del-img" @click="deleteImg(img.id)"><el-icon><Close /></el-icon></button>
        </div>
      </div>
      <el-upload multiple :auto-upload="false" :on-change="onFileChange" accept="image/*">
        <el-button>Загрузить фото</el-button>
      </el-upload>
    </el-form-item>
  </el-form>
</template>

<script setup>
import { ref, watch } from 'vue'
import { adminDeleteImage } from '@/api/admin.js'

const props = defineProps({ modelValue: Object, categories: Array, productId: Number, images: Array })
const emit = defineEmits(['update:modelValue', 'files-changed'])

const allSizes = ['XS', 'S', 'M', 'L', 'XL', 'XXL']
const form = ref({ ...props.modelValue })
const sizeQty = ref(Object.fromEntries(allSizes.map(s => [s, 0])))
const localImages = ref(props.images || [])

watch(() => props.modelValue, v => {
  form.value = { ...v }
  syncSizes(v?.sizes)
}, { deep: true })

watch(() => props.images, v => { localImages.value = v || [] })

watch(form, v => emit('update:modelValue', { ...v, sizes: buildSizes() }), { deep: true })
watch(sizeQty, () => emit('update:modelValue', { ...form.value, sizes: buildSizes() }), { deep: true })

function syncSizes(sizes) {
  allSizes.forEach(s => { sizeQty.value[s] = 0 })
  if (!sizes) return
  sizes.forEach(s => { if (s.size in sizeQty.value) sizeQty.value[s.size] = s.stock_qty })
}

function buildSizes() {
  return Object.entries(sizeQty.value)
    .filter(([, qty]) => qty > 0)
    .map(([size, stock_qty]) => ({ size, stock_qty }))
}

function onFileChange(file, fileList) {
  emit('files-changed', fileList.map(f => f.raw))
}

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

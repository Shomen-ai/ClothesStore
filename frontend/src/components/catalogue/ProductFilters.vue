<template>
  <div class="filters">
    <div class="filter-group">
      <p class="filter-label">КАТЕГОРИЯ</p>
      <div class="filter-item">
        <label>
          <input type="radio" :value="null" v-model="localCat" @change="emit('update:category', null)" />
          Все
        </label>
      </div>
      <div v-for="cat in categories" :key="cat.id" class="filter-item">
        <label>
          <input type="radio" :value="cat.id" v-model="localCat" @change="emit('update:category', cat.id)" />
          {{ cat.name }}
        </label>
      </div>
    </div>
    <div class="filter-group">
      <p class="filter-label">РАЗМЕР</p>
      <div class="sizes">
        <div v-for="s in sizes" :key="s" class="size-chip"
             :class="{ active: localSize === s }"
             @click="toggleSize(s)">{{ s }}</div>
      </div>
    </div>
    <div class="filter-group">
      <p class="filter-label">СОРТИРОВКА</p>
      <el-select v-model="localSort" @change="emit('update:sort', localSort)" style="width:100%">
        <el-option label="Новинки" value="" />
        <el-option label="Цена: по возрастанию" value="price_asc" />
        <el-option label="Цена: по убыванию" value="price_desc" />
      </el-select>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({ categories: Array, modelCategory: Number, modelSize: String, modelSort: String })
const emit = defineEmits(['update:category', 'update:size', 'update:sort'])

const sizes = ['XS', 'S', 'M', 'L', 'XL', 'XXL']
const localCat = ref(props.modelCategory)
const localSize = ref(props.modelSize || '')
const localSort = ref(props.modelSort || '')

function toggleSize(s) {
  localSize.value = localSize.value === s ? '' : s
  emit('update:size', localSize.value)
}
</script>

<style scoped>
.filters { padding: 24px 0; }
.filter-group { margin-bottom: 32px; }
.filter-label { font-size: 11px; letter-spacing: 2px; color: var(--color-text-muted); margin-bottom: 12px; }
.filter-item { margin-bottom: 8px; font-size: 14px; }
.filter-item label { display: flex; align-items: center; gap: 8px; cursor: pointer; }
.sizes { display: flex; flex-wrap: wrap; gap: 6px; }
.size-chip { border: 1px solid var(--color-border); padding: 4px 12px; cursor: pointer; font-size: 12px; }
.size-chip.active { border-color: var(--color-red); color: var(--color-red); }
.size-chip:hover { border-color: var(--color-text-muted); }
</style>

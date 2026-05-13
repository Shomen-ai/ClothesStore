<template>
  <div class="filters">
    <section class="group">
      <p class="group-title">Категория</p>
      <div class="chips">
        <button class="chip" :class="{ active: !localCat }" @click="setCat(null)">Все</button>
        <button v-for="cat in categories" :key="cat.id" class="chip"
                :class="{ active: localCat === cat.id }"
                @click="setCat(cat.id)">{{ cat.name }}</button>
      </div>
    </section>

    <section class="group">
      <p class="group-title">Размер</p>
      <div class="size-row">
        <button v-for="s in sizes" :key="s" class="size"
                :class="{ active: localSize === s }"
                @click="toggleSize(s)">{{ s }}</button>
      </div>
    </section>

    <section class="group">
      <p class="group-title">Цена</p>
      <label v-for="r in priceRanges" :key="r.id" class="check">
        <input type="checkbox" :value="r.id" v-model="localPrice" @change="onPriceChange" />
        <span class="check-box"></span>
        <span class="check-label">{{ r.label }}</span>
      </label>
    </section>

    <section class="group">
      <p class="group-title">Распродажа</p>
      <label class="check check-sale">
        <input type="checkbox" v-model="localSale" @change="emit('update:sale', localSale)" />
        <span class="check-box"></span>
        <span class="check-label">Только со скидкой</span>
        <span v-if="localSale" class="sale-pulse"></span>
      </label>
    </section>

    <section class="group">
      <p class="group-title">Сортировка</p>
      <div class="sort">
        <button v-for="o in sortOptions" :key="o.value" class="sort-opt"
                :class="{ active: localSort === o.value }"
                @click="setSort(o.value)">{{ o.label }}</button>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  categories: Array,
  modelCategory: Number,
  modelSize: String,
  modelSort: String,
  modelPrice: { type: Array, default: () => [] },
  modelSale: Boolean,
})
const emit = defineEmits(['update:category', 'update:size', 'update:sort', 'update:price', 'update:sale'])

const sizes = ['XS', 'S', 'M', 'L', 'XL', 'XXL']
const priceRanges = [
  { id: 'p1', label: 'до 3 000 ₽' },
  { id: 'p2', label: '3 000 — 5 000 ₽' },
  { id: 'p3', label: '5 000 — 7 000 ₽' },
  { id: 'p4', label: 'от 7 000 ₽' },
]
const sortOptions = [
  { value: '',           label: 'Новинки' },
  { value: 'price_asc',  label: 'Сначала дешевле' },
  { value: 'price_desc', label: 'Сначала дороже' },
]

const localCat   = ref(props.modelCategory)
const localSize  = ref(props.modelSize || '')
const localSort  = ref(props.modelSort || '')
const localPrice = ref([...(props.modelPrice || [])])
const localSale  = ref(!!props.modelSale)

watch(() => props.modelSale,  v => { localSale.value = !!v })
watch(() => props.modelPrice, v => { localPrice.value = [...(v || [])] })
watch(() => props.modelCategory, v => { localCat.value = v })

function setCat(id) { localCat.value = id; emit('update:category', id) }
function setSort(v) { localSort.value = v; emit('update:sort', v) }
function toggleSize(s) {
  localSize.value = localSize.value === s ? '' : s
  emit('update:size', localSize.value)
}
function onPriceChange() { emit('update:price', [...localPrice.value]) }
</script>

<style scoped>
.filters {
  display: flex; flex-direction: column; gap: 36px;
  padding: 8px 0 16px;
}
.group { display: flex; flex-direction: column; gap: 12px; }
.group-title {
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--text-muted);
  padding-bottom: 8px;
  border-bottom: 1px dashed var(--border);
}

/* Category chips */
.chips { display: flex; flex-direction: column; gap: 2px; }
.chip {
  text-align: left;
  padding: 8px 12px;
  background: transparent;
  border: 1px solid transparent;
  color: var(--text-soft);
  font-size: 13.5px;
  font-family: inherit;
  cursor: pointer;
  transition: all .15s ease;
  border-radius: 2px;
}
.chip:hover { color: var(--text); background: var(--bg-surface); }
.chip.active {
  color: var(--accent);
  background: var(--accent-haze);
  border-color: var(--accent-haze);
}
.chip.active::before {
  content: '— ';
  font-family: var(--font-mono);
  margin-right: 2px;
}

/* Sizes */
.size-row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 6px; }
.size {
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text-soft);
  font: 500 12px/1 var(--font-mono);
  padding: 10px 0;
  cursor: pointer;
  transition: all .15s ease;
}
.size:hover { border-color: var(--text-muted); color: var(--text); }
.size.active {
  border-color: var(--accent);
  color: var(--accent);
  background: var(--accent-haze);
}

/* Checkboxes */
.check {
  display: flex; align-items: center; gap: 10px;
  cursor: pointer;
  font-size: 13.5px;
  color: var(--text-soft);
  padding: 4px 0;
  user-select: none;
  position: relative;
}
.check input { position: absolute; opacity: 0; pointer-events: none; }
.check-box {
  width: 14px; height: 14px;
  border: 1px solid var(--border-strong);
  display: inline-flex; align-items: center; justify-content: center;
  transition: all .15s ease;
}
.check input:checked + .check-box {
  background: var(--accent);
  border-color: var(--accent);
}
.check input:checked + .check-box::after {
  content: '';
  width: 4px; height: 8px;
  border: solid #fff;
  border-width: 0 1.5px 1.5px 0;
  transform: rotate(45deg);
  margin-bottom: 2px;
}
.check:hover .check-label { color: var(--text); }

.check-sale .check input:checked + .check-box {
  background: var(--sale);
  border-color: var(--sale);
}
.check-sale .check input:checked + .check-box::after {
  border-color: #fff;
}
.sale-pulse {
  width: 6px; height: 6px; border-radius: 50%;
  background: var(--sale);
  box-shadow: 0 0 8px var(--sale-glow);
  margin-left: auto;
  animation: glow-pulse 1.6s ease-in-out infinite;
}

/* Sort */
.sort { display: flex; flex-direction: column; gap: 2px; }
.sort-opt {
  text-align: left;
  padding: 8px 0;
  background: none;
  border: none;
  color: var(--text-soft);
  font-size: 13.5px;
  font-family: inherit;
  cursor: pointer;
  position: relative;
  transition: color .15s ease;
}
.sort-opt::before {
  content: '○';
  font-family: var(--font-mono);
  color: var(--text-dim);
  margin-right: 8px;
  transition: color .15s ease;
}
.sort-opt:hover { color: var(--text); }
.sort-opt.active { color: var(--accent); }
.sort-opt.active::before { content: '●'; color: var(--accent); }
</style>

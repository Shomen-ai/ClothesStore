import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useCartStore = defineStore('cart', () => {
  const items = ref(JSON.parse(localStorage.getItem('cart') || '[]'))
  const promoCode = ref('')
  const discount = ref(0)

  const total = computed(() => items.value.reduce((s, i) => s + i.price * i.quantity, 0))
  const finalTotal = computed(() => Math.max(0, total.value - discount.value))
  const count = computed(() => items.value.reduce((s, i) => s + i.quantity, 0))

  function save() { localStorage.setItem('cart', JSON.stringify(items.value)) }

  function add(product, sizeID, sizeName) {
    const existing = items.value.find(i => i.product_size_id === sizeID)
    if (existing) {
      existing.quantity++
    } else {
      items.value.push({
        product_id: product.id,
        product_size_id: sizeID,
        name: product.name,
        size: sizeName,
        price: product.price,
        quantity: 1,
        image: product.images?.[0]?.image_path || ''
      })
    }
    save()
  }

  function remove(sizeID) {
    items.value = items.value.filter(i => i.product_size_id !== sizeID)
    save()
  }

  function updateQty(sizeID, qty) {
    const item = items.value.find(i => i.product_size_id === sizeID)
    if (item) { item.quantity = qty; save() }
  }

  function applyPromo(code, amount) { promoCode.value = code; discount.value = amount }
  function clearPromo() { promoCode.value = ''; discount.value = 0 }
  function clear() { items.value = []; promoCode.value = ''; discount.value = 0; save() }

  return { items, promoCode, discount, total, finalTotal, count, add, remove, updateQty, applyPromo, clearPromo, clear }
})

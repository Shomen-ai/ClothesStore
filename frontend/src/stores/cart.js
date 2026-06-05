import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useCartStore = defineStore('cart', () => {
  // Cart line items, rehydrated from localStorage so the cart survives reloads.
  // Each line is keyed by product_size_id (product + chosen size variant).
  const items = ref(JSON.parse(localStorage.getItem('cart') || '[]'))
  // Currently applied promo code and its resolved discount amount.
  const promoCode = ref('')
  const discount = ref(0)

  // Subtotal before discount: sum of price * quantity across all line items.
  const total = computed(() => items.value.reduce((s, i) => s + i.price * i.quantity, 0))
  // Total after promo discount, floored at 0 so it can never go negative.
  const finalTotal = computed(() => Math.max(0, total.value - discount.value))
  // Total number of units across the cart (for the header badge).
  const count = computed(() => items.value.reduce((s, i) => s + i.quantity, 0))

  // Persist the current cart contents to localStorage.
  function save() { localStorage.setItem('cart', JSON.stringify(items.value)) }

  // Add a product/size to the cart: bump quantity if the size variant is already
  // present, otherwise push a new line snapshotting name/size/price/thumbnail.
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

  // Drop a line item by its size variant id and persist.
  function remove(sizeID) {
    items.value = items.value.filter(i => i.product_size_id !== sizeID)
    save()
  }

  // Set an explicit quantity for a given size variant and persist.
  function updateQty(sizeID, qty) {
    const item = items.value.find(i => i.product_size_id === sizeID)
    if (item) { item.quantity = qty; save() }
  }

  // Apply a validated promo code + discount; clear it; or empty the whole cart.
  function applyPromo(code, amount) { promoCode.value = code; discount.value = amount }
  function clearPromo() { promoCode.value = ''; discount.value = 0 }
  function clear() { items.value = []; promoCode.value = ''; discount.value = 0; save() }

  return { items, promoCode, discount, total, finalTotal, count, add, remove, updateQty, applyPromo, clearPromo, clear }
})

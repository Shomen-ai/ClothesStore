import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getWishlist, addToWishlist, removeFromWishlist } from '@/api/user.js'
import { useAuthStore } from './auth.js'

export const useWishlistStore = defineStore('wishlist', () => {
  const ids = ref(new Set())

  async function load() {
    const auth = useAuthStore()
    if (!auth.isLoggedIn) return
    const { data } = await getWishlist()
    ids.value = new Set(data.map(p => p.id))
  }

  function has(id) { return ids.value.has(id) }

  async function toggle(id) {
    if (ids.value.has(id)) {
      await removeFromWishlist(id)
      ids.value.delete(id)
    } else {
      await addToWishlist(id)
      ids.value.add(id)
    }
  }

  return { ids, load, has, toggle }
})

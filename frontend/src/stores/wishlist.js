import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getWishlist, addToWishlist, removeFromWishlist } from '@/api/user.js'
import { useAuthStore } from './auth.js'

export const useWishlistStore = defineStore('wishlist', () => {
  // Set of wishlisted product ids — fast membership checks for the heart toggle.
  const ids = ref(new Set())

  // Fetch the server-side wishlist and rebuild the id Set (no-op when logged out).
  async function load() {
    const auth = useAuthStore()
    if (!auth.isLoggedIn) return
    const { data } = await getWishlist()
    ids.value = new Set(data.map(p => p.id))
  }

  // Whether a product is currently in the wishlist.
  function has(id) { return ids.value.has(id) }

  // Toggle wishlist membership, syncing the change to the API and the local Set.
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

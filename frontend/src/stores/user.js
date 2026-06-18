import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useUserStore = defineStore('user', () => {
  const currentUser = ref({
    id: 1,
    username: 'admin',
    role: 'admin',
    email: 'admin@test.com'
  })

  const isAdmin = computed(() => currentUser.value.role === 'admin')

  function setUser(user) {
    currentUser.value = user
  }

  return { currentUser, isAdmin, setUser }
})

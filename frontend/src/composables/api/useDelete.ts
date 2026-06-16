import { ref } from 'vue'
import axios from 'axios'

export function useDelete() {
  const loading = ref(false)
  const error = ref<string | null>(null)

  const deleteData = async (url: string) => {
    loading.value = true
    error.value = null
    try {
      await axios.delete(url)
      return true
    } catch (err) {
      if (axios.isAxiosError(err)) {
        error.value = err.response?.data?.message || err.message
      } else {
        error.value = 'Something went wrong'
      }
      return false
    } finally {
      loading.value = false
    }
  }

  return { loading, error, deleteData }
}

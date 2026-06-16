import { ref } from 'vue'
import axios from 'axios'

export function usePut() {
  const loading = ref(false)
  const error = ref<string | null>(null)

  const putData = async (url: string, body: unknown) => {
    loading.value = true
    error.value = null
    try {
      await axios.put(url, body)
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

  return { loading, error, putData }
}

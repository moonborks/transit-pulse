import { ref } from 'vue'
import axios from 'axios'

export function usePost<T, R = T>(transform?: (data: T) => R) {
  const data = ref<T | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const postData = async (url: string, body: unknown) => {
    loading.value = true
    error.value = null
    try {
      const response = await axios.post<T>(url, body)
      data.value = transform ? transform(response.data) : (response.data as unknown as R)
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

  return { data, loading, error, postData }
}

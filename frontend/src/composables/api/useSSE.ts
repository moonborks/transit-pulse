import { endpoints } from '@/api/endpoints'
import { ref } from 'vue'

const tripEvent = ref<string | null>(null)

const eventSource = new EventSource(endpoints.mta.trips.getMessages)
eventSource.onopen = () => console.debug('SSE connected')
eventSource.onerror = (e) => console.error('SSE error', { e })
eventSource.onmessage = (event) => {
  tripEvent.value = event.data
  console.debug(`SSE trip event received, ${tripEvent.value}`)
}

export function useTripSSE() {
  return { tripEvent }
}

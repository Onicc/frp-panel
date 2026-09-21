import { onBeforeUnmount, onMounted, ref } from 'vue'

export function useAutoRefresh(callback: () => Promise<void>, interval = 5000) {
  // The first request starts on mount. Keep consumers in a loading state until
  // it completes so empty-state copy does not flash before data arrives.
  const loading = ref(true)
  const refreshing = ref(false)
  let timer: number | undefined
  let inFlight = false

  const refresh = async (silent = false) => {
    if (inFlight) return
    inFlight = true
    if (silent) refreshing.value = true
    else loading.value = true
    try {
      await callback()
    } finally {
      inFlight = false
      if (silent) refreshing.value = false
      else loading.value = false
    }
  }

  onMounted(() => {
    void refresh()
    timer = window.setInterval(() => void refresh(true), interval)
  })

  onBeforeUnmount(() => {
    if (timer !== undefined) window.clearInterval(timer)
  })

  return { loading, refreshing, refresh }
}

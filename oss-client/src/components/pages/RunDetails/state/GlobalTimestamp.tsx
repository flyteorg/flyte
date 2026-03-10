import { create } from 'zustand'
import { useEffect } from 'react'

const INTERVAL_MS = 153

export function GlobalNowTicker() {
  useEffect(() => {
    const tick = () => {
      useGlobalNow.setState((prev) => {
        const current = Date.now()
        if (prev.now === current) return prev // avoid unnecessary updates
        return { now: current }
      })
    }

    let interval: ReturnType<typeof setInterval> | null = null
    const start = () => {
      interval = setInterval(tick, INTERVAL_MS)
    }
    const stop = () => {
      if (interval) {
        clearInterval(interval)
        interval = null
      }
    }

    const handleVisibility = () => {
      if (document.hidden) {
        stop()
      } else {
        start()
      }
    }

    document.addEventListener('visibilitychange', handleVisibility)
    start()

    return () => {
      stop()
      document.removeEventListener('visibilitychange', handleVisibility)
    }
  }, [])

  return null
}

type State = {
  now: number
}

export const useGlobalNow = create<State>(() => ({
  now: Date.now(),
}))

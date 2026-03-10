import { useEffect, useState } from 'react'

export function useDelayedVisibility(value: number, delayMs: number) {
  const [visible, setVisible] = useState(value > 0)

  useEffect(() => {
    if (value > 0) {
      setVisible(true)
      return
    }
    const timer = setTimeout(() => setVisible(false), delayMs)
    return () => clearTimeout(timer)
  }, [value, delayMs])

  return visible
}

export function useDelayedShow(value: number, delayMs: number) {
  const [show, setShow] = useState(value > 0)
  useEffect(() => {
    if (value > 0) {
      const timer = setTimeout(() => setShow(true), delayMs)
      return () => clearTimeout(timer)
    } else {
      setShow(false)
    }
  }, [value, delayMs])
  return show
}

export function useDebouncedNumber(value: number, delayMs: number) {
  const [debounced, setDebounced] = useState(value)
  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delayMs)
    return () => clearTimeout(timer)
  }, [value, delayMs])
  return debounced
}

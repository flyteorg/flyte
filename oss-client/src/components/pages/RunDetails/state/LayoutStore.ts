import { create } from 'zustand'

export type LayoutMode =
  | 'default' // action log approx 500px, main content fills
  // | 'mini-action-log' // not-implemented: action log approx 20px, main content fills
  | 'no-action-log' // main content fills
  | 'full-action-log' // action log fullscreen, main hidden

interface LayoutState {
  mode: LayoutMode
  setMode: (mode: LayoutMode) => void
}

export const useLayoutStore = create<LayoutState>((set) => ({
  mode: 'default',
  setMode: (mode) => set({ mode }),
}))

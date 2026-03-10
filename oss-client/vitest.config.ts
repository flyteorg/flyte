import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    setupFiles: ['./vitest.setup.ts'],
    globals: true,
    exclude: ['**/e2e/**', '**/node_modules/**'],
  },
  resolve: {
    alias: {
      '@/gen': path.resolve(__dirname, './gen'),
      '@': path.resolve(__dirname, './src'),
    },
  },
})

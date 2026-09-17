import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: '../cmd/frpp/out',
    emptyOutDir: true,
    sourcemap: false,
    target: 'es2022',
  },
  server: {
    proxy: {
      '/api': 'http://127.0.0.1:9000',
    },
  },
  test: {
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
    exclude: ['e2e/**', 'node_modules/**'],
  },
})

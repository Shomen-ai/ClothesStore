import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) }
  },
  server: {
    // Dev proxy target. Defaults to the production server so the frontend can run
    // locally without a local backend. Override with VITE_API_TARGET=http://localhost:8080
    // to point at a locally-running backend instead.
    proxy: {
      '/api': { target: process.env.VITE_API_TARGET || 'http://31.130.151.134', changeOrigin: true },
      '/uploads': { target: process.env.VITE_API_TARGET || 'http://31.130.151.134', changeOrigin: true }
    }
  }
})

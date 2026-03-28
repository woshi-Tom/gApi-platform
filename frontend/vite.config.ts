import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

const API_TARGET = process.env.VITE_API_URL || 'http://localhost:8080'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      '/api': {
        target: API_TARGET,
        changeOrigin: true,
        rewrite: (path) => path
      }
    }
  },
  preview: {
    host: '0.0.0.0',
    port: 5173
  },
  build: {
    outDir: 'dist',
    chunkSizeWarningLimit: 600,
    rollupOptions: {
      input: resolve(__dirname, 'index.html')
    }
  },
  define: {
    'import.meta.env.VITE_API_URL': JSON.stringify(API_TARGET)
  }
})

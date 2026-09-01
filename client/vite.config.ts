import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
  ],
  server: {
    port: 5173,
    host: true,
    proxy: {
      '/todo.v1.TodoService': {
        target: 'https://localhost:8085',
        secure: false, // self-signed cert allow karta hai
        changeOrigin: true,
      },
    },
  },
})

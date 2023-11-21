import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173
  },
  resolve: {
    alias: {
      "@s": path.resolve(__dirname, "src"),
      "@ipc": path.resolve(__dirname, "ipc/index.js"),
      "@utils": path.resolve(__dirname, "utils"),
    },
  }
})

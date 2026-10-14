import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// 本地开发：5173 端口，/api 与 /media 代理到本机 Go 服务（:9090）
// 生产环境：由 Nginx 容器反代，见 nginx.conf
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:9090',
      '/media': 'http://localhost:9090',
    },
  },
})

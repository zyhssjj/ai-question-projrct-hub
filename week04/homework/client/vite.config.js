import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  // 配置开发服务器
  server: {
    port: 5173, // 前端开发服务器端口
    open: true, // 自动打开浏览器
    proxy: {
      // 将所有以/api开头的请求转发到后端
      '/api': {
        target: 'http://localhost:8080', // 后端地址
        changeOrigin: true, // 改变请求头中的Origin
        secure: false,
        // rewrite: (path) => path.replace(/^\/api/, ''), // 去掉/api前缀
      },
    },
  },
  // 构建配置
  build: {
    outDir: 'dist', // 输出目录
    sourcemap: false, // 不生成sourcemap
    rollupOptions: {
      output: {
        // 静态资源分类
        chunkFileNames: 'assets/js/[name]-[hash].js',
        entryFileNames: 'assets/js/[name]-[hash].js',
        assetFileNames: 'assets/[ext]/[name]-[hash].[ext]',
      },
    },
  },
});
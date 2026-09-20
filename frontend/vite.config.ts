import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 18941,
    proxy: { '/api': { target: 'http://localhost:19941', changeOrigin: true } }
  }
});

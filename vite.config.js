import { defineConfig } from 'vite';
import { resolve } from 'node:path';

export default defineConfig({
  root: 'frontend',
  base: './',
  define: {
    __VUE_OPTIONS_API__: true,
    __VUE_PROD_DEVTOOLS__: false,
    __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: false,
  },
  build: {
    outDir: '../web',
    license: { fileName: 'assets/third-party-licenses.txt' },
    emptyOutDir: true,
    rollupOptions: {
      input: {
        scalar: resolve('frontend/index.html'),
        elements: resolve('frontend/elements.html'),
        swagger: resolve('frontend/swagger.html'),
      },
    },
  },
});

import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

const workspaceRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  '..'
);

export default defineConfig({
  plugins: [vue()],
  server: {
    host: '0.0.0.0',
    port: 4175,
    fs: {
      allow: [workspaceRoot]
    }
  },
  preview: {
    host: '0.0.0.0',
    port: 4175
  }
});

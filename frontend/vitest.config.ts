import { defineConfig } from 'vitest/config';
import path from 'path';

export default defineConfig({
  resolve: {
    alias: [
      // Match any relative depth: ../wailsjs, ../../wailsjs, ../../../wailsjs, etc.
      {
        find: /.*\/wailsjs\/go\/main\/App/,
        replacement: path.resolve(__dirname, 'src/test/wailsMock.ts'),
      },
      {
        find: /.*\/wailsjs\/go\/models/,
        replacement: path.resolve(__dirname, 'src/test/modelsMock.ts'),
      },
    ],
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
      reportsDirectory: './coverage',
      include: ['src/**/*.{ts,tsx}'],
      exclude: ['src/test/**', 'src/vite-env.d.ts', 'src/main.tsx'],
    },
  },
});

import { defineConfig } from 'vitest/config';
import path from 'node:path';

const wailsMock = path.resolve(__dirname, 'src/test/wailsMock.ts');
const modelsMock = path.resolve(__dirname, 'src/test/modelsMock.ts');
const runtimeMock = path.resolve(__dirname, 'src/test/wailsRuntimeMock.ts');

export default defineConfig({
  resolve: {
    alias: [
      {
        find: /wailsjs\/go\/main\/App/,
        replacement: wailsMock,
        customResolver: () => wailsMock,
      },
      {
        find: /wailsjs\/go\/models/,
        replacement: modelsMock,
        customResolver: () => modelsMock,
      },
      {
        find: /wailsjs\/runtime\/runtime/,
        replacement: runtimeMock,
        customResolver: () => runtimeMock,
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
      all: false,
      include: ['src/**/*.{ts,tsx}'],
      exclude: ['src/test/**', 'src/vite-env.d.ts', 'src/main.tsx'],
    },
  },
});

// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
import '@testing-library/jest-dom/vitest';

const values = new Map();
const storage =
  window.localStorage ??
  {
    getItem: (key) => values.get(String(key)) ?? null,
    setItem: (key, value) => values.set(String(key), String(value)),
    removeItem: (key) => values.delete(String(key)),
    clear: () => values.clear(),
    key: (index) => [...values.keys()][index] ?? null,
    get length() {
      return values.size;
    },
  };

Object.defineProperty(globalThis, 'localStorage', {
  configurable: true,
  value: storage,
});

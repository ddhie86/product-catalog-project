/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx}'],
  theme: {
    extend: {
      colors: {
        ink: { DEFAULT: '#1B2733', soft: '#566675' },
        paper: '#F4F6F8',
        line: '#D9E0E6',
        brand: { DEFAULT: '#1D4E89', dark: '#163B68', tint: '#E6EEF7' },
        signal: { DEFAULT: '#9A5200', tint: '#FBEFD9' }, // pending, stok menipis
        ok: { DEFAULT: '#256A41', tint: '#E3F3E9' },
        danger: { DEFAULT: '#A5322A', tint: '#F9E4E1' },
      },
      fontFamily: {
        sans: ['"Plus Jakarta Sans Variable"', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
};

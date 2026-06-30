/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        mono: ['"Courier New"', 'Courier', 'monospace'],
      },
      colors: {
        ink: '#14110f',
        panel: '#1f1a14',
        parchment: '#e9dcc0',
        accent: '#c9a24b',
        good: '#4caf6b',
        evil: '#b5483c',
      },
    },
  },
  plugins: [],
}

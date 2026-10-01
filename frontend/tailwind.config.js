/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        ink: '#111827',
        page: '#f4f6f9',
        brand: '#5146f0',
        'brand-soft': '#eef0ff',
        success: '#0f8a64',
        warning: '#c34a08',
      },
      boxShadow: {
        card: '0 1px 2px rgb(15 23 42 / 0.04)',
      },
    },
  },
  plugins: [],
}

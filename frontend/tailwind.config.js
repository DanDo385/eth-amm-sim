/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './app/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        // Trading theme colors
        'bull': '#22c55e',
        'bear': '#ef4444',
        'neutral': '#6b7280',
        'background': '#0f1419',
        'surface': '#1a1f26',
        'surface-light': '#242b35',
        'border': '#2d3748',
      },
    },
  },
  plugins: [],
}

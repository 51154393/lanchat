/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: '#07c160',
        'primary-hover': '#06ad56',
        'primary-light': '#e8f8ee',
        danger: '#ff4d4f',
        'danger-hover': '#d9363e',
        warning: '#fa8c16',
        'bg-chat': '#f5f5f5',
        'bg-bubble-self': '#95ec69',
        'bg-bubble-other': '#fff',
        'text-primary': '#191919',
        'text-secondary': '#666666',
        'text-tertiary': '#999999',
        'border-light': '#e7e7e7',
        'border-lighter': '#f0f0f0',
      },
      fontFamily: {
        sans: ['-apple-system', 'BlinkMacSystemFont', '"Segoe UI"', 'Roboto',
          '"Helvetica Neue"', 'Arial', '"Noto Sans"', 'sans-serif'],
      },
      boxShadow: {
        'panel': '0 8px 32px rgba(0, 0, 0, 0.2)',
        'bubble': '0 1px 2px rgba(0, 0, 0, 0.06)',
      },
    },
  },
  plugins: [],
}

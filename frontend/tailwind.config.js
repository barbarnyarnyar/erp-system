/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx,html}"
  ],
  theme: {
    extend: {
      colors: {
        'obsidian-base': '#0B0F19',
        'slate-elevated': '#1E293B',
        'confirmed-blue': '#3B82F6',
        'credit-crimson': '#EF4444',
        'telemetry-dark': '#111827',
      },
    },
  },
  plugins: [],
}



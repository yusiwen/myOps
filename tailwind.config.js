/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./web/templates/**/*.html"],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        surface: {
          DEFAULT: "#EEF0F8",
          dark: "#1a1a23",
        },
        sidebar: {
          DEFAULT: "#ffffff",
          dark: "#23232d",
        },
        topbar: {
          DEFAULT: "#ffffff",
          dark: "#1e1e28",
        },
      },
    },
  },
  plugins: [],
};

/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {},
  },
  // La clave está en estas dos secciones:
  plugins: [
    require('daisyui'),
  ],
  // Aquí le decimos a Tailwind que clases NO debe eliminar
  safelist: [
    'alert-success',
    'alert-error',
    'alert-info',
    'alert-warning', // Agregamos warning por si lo usas en el futuro
  ],
  daisyui: {
    themes: ["synthwave"], // Asegúrate de que tu tema esté aquí
  },
}
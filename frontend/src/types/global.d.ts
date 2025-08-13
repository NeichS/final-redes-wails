// src/types/globals.d.ts

// CSS global (no-modular)
declare module '*.css';

// CSS Modules (si usás archivos *.module.css)
declare module '*.module.css' {
  const classes: { readonly [key: string]: string };
  export default classes;
}

// Opcional: por si usás otros assets
declare module '*.scss';
declare module '*.sass';
declare module '*.png';
declare module '*.jpg';
declare module '*.svg';

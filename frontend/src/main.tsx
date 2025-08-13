import React from 'react'
import { createRoot } from 'react-dom/client'
import './styles/App.css'
import App from './pages/App.js'

const container = document.getElementById('root')
if (!container) throw new Error('Root container not found')

createRoot(container).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
)
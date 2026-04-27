import React from 'react'
import {createRoot} from 'react-dom/client'
import './styles/theme.css'
import './styles/global.css'
import App from './App'
import { LogFromFrontend } from '../wailsjs/go/main/App'

// Intercept console methods and mirror them to the Go application log
// so frontend activity appears in the Logs modal alongside backend logs.
const origLog = console.log;
const origWarn = console.warn;
const origError = console.error;
const origDebug = console.debug;

function forward(level: string, args: unknown[]) {
  const msg = args.map(a => (typeof a === 'string' ? a : JSON.stringify(a))).join(' ');
  LogFromFrontend(level, 'frontend', msg).catch(() => {/* backend not ready yet */});
}

console.log = (...args: unknown[]) => { origLog(...args); forward('info', args); };
console.warn = (...args: unknown[]) => { origWarn(...args); forward('warn', args); };
console.error = (...args: unknown[]) => { origError(...args); forward('error', args); };
console.debug = (...args: unknown[]) => { origDebug(...args); forward('debug', args); };

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <React.StrictMode>
        <App/>
    </React.StrictMode>
)

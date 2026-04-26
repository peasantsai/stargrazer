# Frontend Architecture

## Stack

- **React 18** with TypeScript
- **Vite** for bundling and hot reload
- **Plain CSS** with custom properties (no CSS-in-JS)
- **Wails bindings** for Go backend calls

## Views

All components live in a single `App.tsx` file:

| View | Description |
|------|-------------|
| **Chat** | Upload form, message log, browser start/stop |
| **Schedules** | Job list, create/detail modals |
| **Settings** | Social connections, browser config, flags, logs |

## Styling

Two CSS files:

- `styles/theme.css` — CSS variables for `[data-theme="dark"]` and `[data-theme="light"]`
- `styles/global.css` — All component styles consuming theme variables

## Component Pattern

Components follow a consistent pattern:

```tsx
function MyPanel({ prop1, prop2 }: { prop1: Type; prop2: Type }) {
  const [state, setState] = useState<Type>(initial);
  // ... handlers
  return (<div className="my-panel">...</div>);
}
```

Modals use the overlay pattern with click-outside dismissal:

```tsx
<div className="modal-overlay" onClick={onClose}>
  <div className="modal-content" onClick={e => e.stopPropagation()}>
    <div className="modal-header">...</div>
    <div className="modal-body">...</div>
    <div className="modal-footer">...</div>
  </div>
</div>
```

## Wails Bindings

Auto-generated in `frontend/wailsjs/go/main/App.js`. Import and call:

```tsx
import { StartBrowser, GetPlatforms } from '../wailsjs/go/main/App';
const res = await StartBrowser();
```

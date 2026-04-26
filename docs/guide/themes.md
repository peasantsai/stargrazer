# Themes

Stargrazer supports dark and light themes.

## Switching Themes

Use the theme toggle at the bottom of the sidebar. The selection persists across sessions.

## CSS Variables

All colors are defined as CSS custom properties in `src/styles/theme.css`:

```css
[data-theme="dark"] {
  --bg-primary: #0f0f0f;
  --text-primary: #e8e8e8;
  --accent: #7c5cfc;
  /* ... 30+ variables */
}

[data-theme="light"] {
  --bg-primary: #f5f5f5;
  --text-primary: #1a1a1a;
  --accent: #6d4aed;
}
```

Components in `global.css` consume these variables exclusively — no hardcoded colors.

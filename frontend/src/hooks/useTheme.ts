import { useState, useEffect } from 'react';
import type { Theme } from '../types';

export function useTheme(): [Theme, (t: Theme) => void] {
  const [theme, setTheme] = useState<Theme>(
    () => (localStorage.getItem('stargrazer-theme') as Theme) || 'dark'
  );

  const applyTheme = (t: Theme) => {
    setTheme(t);
    localStorage.setItem('stargrazer-theme', t);
    document.documentElement.dataset.theme = t;
  };

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  return [theme, applyTheme];
}

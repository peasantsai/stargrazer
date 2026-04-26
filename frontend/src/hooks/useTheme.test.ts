import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useTheme } from './useTheme';

beforeEach(() => {
  localStorage.clear();
  delete document.documentElement.dataset.theme;
  vi.restoreAllMocks();
});

describe('useTheme – initial state', () => {
  it('defaults to dark theme when localStorage is empty', () => {
    const { result } = renderHook(() => useTheme());
    const [theme] = result.current;
    expect(theme).toBe('dark');
  });

  it('loads saved theme from localStorage on mount', () => {
    localStorage.setItem('stargrazer-theme', 'light');
    const { result } = renderHook(() => useTheme());
    const [theme] = result.current;
    expect(theme).toBe('light');
  });

  it('applies theme to document.documentElement.dataset.theme on mount', () => {
    const { result } = renderHook(() => useTheme());
    expect(document.documentElement.dataset.theme).toBe(result.current[0]);
  });

  it('applies saved light theme to dataset.theme on mount', () => {
    localStorage.setItem('stargrazer-theme', 'light');
    renderHook(() => useTheme());
    expect(document.documentElement.dataset.theme).toBe('light');
  });
});

describe('useTheme – applyTheme', () => {
  it('returns a tuple of [Theme, function]', () => {
    const { result } = renderHook(() => useTheme());
    expect(Array.isArray(result.current)).toBe(true);
    expect(result.current).toHaveLength(2);
    expect(typeof result.current[1]).toBe('function');
  });

  it('updates theme state when applyTheme is called', () => {
    const { result } = renderHook(() => useTheme());
    act(() => { result.current[1]('light'); });
    expect(result.current[0]).toBe('light');
  });

  it('persists theme to localStorage', () => {
    const { result } = renderHook(() => useTheme());
    act(() => { result.current[1]('light'); });
    expect(localStorage.getItem('stargrazer-theme')).toBe('light');
  });

  it('updates document.documentElement.dataset.theme', () => {
    const { result } = renderHook(() => useTheme());
    act(() => { result.current[1]('light'); });
    expect(document.documentElement.dataset.theme).toBe('light');
  });

  it('can switch back from light to dark', () => {
    localStorage.setItem('stargrazer-theme', 'light');
    const { result } = renderHook(() => useTheme());
    act(() => { result.current[1]('dark'); });
    expect(result.current[0]).toBe('dark');
    expect(localStorage.getItem('stargrazer-theme')).toBe('dark');
    expect(document.documentElement.dataset.theme).toBe('dark');
  });
});

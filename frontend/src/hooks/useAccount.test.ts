import { describe, it, expect, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useAccount } from './useAccount';

const STORAGE_KEY = 'stargrazer-account';

beforeEach(() => {
  localStorage.clear();
});

describe('useAccount – initial state', () => {
  it('returns default account when localStorage is empty', () => {
    const { result } = renderHook(() => useAccount());
    const [account] = result.current;
    expect(account.name).toBe('User');
    expect(account.email).toBe('');
    expect(account.avatarUrl).toBe('');
  });

  it('loads saved account from localStorage on mount', () => {
    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify({ name: 'Alice', email: 'alice@example.com', avatarUrl: 'https://example.com/avatar.png' }),
    );
    const { result } = renderHook(() => useAccount());
    const [account] = result.current;
    expect(account.name).toBe('Alice');
    expect(account.email).toBe('alice@example.com');
    expect(account.avatarUrl).toBe('https://example.com/avatar.png');
  });

  it('falls back to defaults when localStorage contains non-JSON', () => {
    localStorage.setItem(STORAGE_KEY, 'not-valid-json{{');
    const { result } = renderHook(() => useAccount());
    const [account] = result.current;
    expect(account.name).toBe('User');
  });

  it('falls back to defaults when localStorage value is null JSON', () => {
    localStorage.setItem(STORAGE_KEY, 'null');
    const { result } = renderHook(() => useAccount());
    const [account] = result.current;
    expect(account.name).toBe('User');
  });

  it('falls back to defaults when localStorage value is a JSON array', () => {
    localStorage.setItem(STORAGE_KEY, '[]');
    const { result } = renderHook(() => useAccount());
    const [account] = result.current;
    expect(account.name).toBe('User');
  });

  it('uses defaults for missing individual fields', () => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ name: 'Bob' }));
    const { result } = renderHook(() => useAccount());
    const [account] = result.current;
    expect(account.name).toBe('Bob');
    expect(account.email).toBe('');
    expect(account.avatarUrl).toBe('');
  });

  it('ignores non-string field values', () => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ name: 42, email: null, avatarUrl: true }));
    const { result } = renderHook(() => useAccount());
    const [account] = result.current;
    expect(account.name).toBe('User');
    expect(account.email).toBe('');
    expect(account.avatarUrl).toBe('');
  });
});

describe('useAccount – updateAccount', () => {
  it('updates name and persists to localStorage', () => {
    const { result } = renderHook(() => useAccount());
    act(() => {
      result.current[1]({ name: 'Charlie' });
    });
    expect(result.current[0].name).toBe('Charlie');
    const stored = JSON.parse(localStorage.getItem(STORAGE_KEY)!);
    expect(stored.name).toBe('Charlie');
  });

  it('updates email and persists to localStorage', () => {
    const { result } = renderHook(() => useAccount());
    act(() => {
      result.current[1]({ email: 'charlie@example.com' });
    });
    expect(result.current[0].email).toBe('charlie@example.com');
    const stored = JSON.parse(localStorage.getItem(STORAGE_KEY)!);
    expect(stored.email).toBe('charlie@example.com');
  });

  it('updates avatarUrl and persists to localStorage', () => {
    const { result } = renderHook(() => useAccount());
    act(() => {
      result.current[1]({ avatarUrl: 'https://example.com/new.png' });
    });
    expect(result.current[0].avatarUrl).toBe('https://example.com/new.png');
  });

  it('updates multiple fields at once', () => {
    const { result } = renderHook(() => useAccount());
    act(() => {
      result.current[1]({ name: 'Dana', email: 'dana@example.com', avatarUrl: 'https://example.com/dana.png' });
    });
    const [account] = result.current;
    expect(account.name).toBe('Dana');
    expect(account.email).toBe('dana@example.com');
    expect(account.avatarUrl).toBe('https://example.com/dana.png');
  });

  it('preserves unchanged fields when only one field is updated', () => {
    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify({ name: 'Eve', email: 'eve@example.com', avatarUrl: '' }),
    );
    const { result } = renderHook(() => useAccount());
    act(() => {
      result.current[1]({ name: 'Eve Updated' });
    });
    expect(result.current[0].email).toBe('eve@example.com');
    expect(result.current[0].avatarUrl).toBe('');
  });

  it('ignores non-string partial values and keeps existing', () => {
    const { result } = renderHook(() => useAccount());
    act(() => {
      result.current[1]({ name: 'Frank' });
    });
    act(() => {
      // @ts-expect-error intentional bad type for test
      result.current[1]({ name: 123 });
    });
    expect(result.current[0].name).toBe('Frank');
  });
});

describe('useAccount – sanitization', () => {
  it('trims leading and trailing whitespace from name', () => {
    const { result } = renderHook(() => useAccount());
    act(() => {
      result.current[1]({ name: '  Grace  ' });
    });
    expect(result.current[0].name).toBe('Grace');
  });

  it('trims whitespace from email', () => {
    const { result } = renderHook(() => useAccount());
    act(() => {
      result.current[1]({ email: '  grace@example.com  ' });
    });
    expect(result.current[0].email).toBe('grace@example.com');
  });

  it('trims whitespace from avatarUrl', () => {
    const { result } = renderHook(() => useAccount());
    act(() => {
      result.current[1]({ avatarUrl: '  https://example.com/img.png  ' });
    });
    expect(result.current[0].avatarUrl).toBe('https://example.com/img.png');
  });

  it('truncates name exceeding 500 characters', () => {
    const longName = 'a'.repeat(600);
    const { result } = renderHook(() => useAccount());
    act(() => {
      result.current[1]({ name: longName });
    });
    expect(result.current[0].name.length).toBe(500);
  });

  it('truncates email exceeding 500 characters', () => {
    const longEmail = 'b'.repeat(600);
    const { result } = renderHook(() => useAccount());
    act(() => {
      result.current[1]({ email: longEmail });
    });
    expect(result.current[0].email.length).toBe(500);
  });

  it('sanitizes values loaded from localStorage on mount', () => {
    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify({ name: '  Padded  ', email: 'trim@example.com', avatarUrl: '' }),
    );
    const { result } = renderHook(() => useAccount());
    expect(result.current[0].name).toBe('Padded');
  });

  it('persists sanitized value to localStorage', () => {
    const { result } = renderHook(() => useAccount());
    act(() => {
      result.current[1]({ name: '  Stored  ' });
    });
    const stored = JSON.parse(localStorage.getItem(STORAGE_KEY)!);
    expect(stored.name).toBe('Stored');
  });
});

describe('useAccount – return shape', () => {
  it('returns a tuple of [AccountInfo, function]', () => {
    const { result } = renderHook(() => useAccount());
    expect(Array.isArray(result.current)).toBe(true);
    expect(result.current).toHaveLength(2);
    expect(typeof result.current[0]).toBe('object');
    expect(typeof result.current[1]).toBe('function');
  });

  it('account object has name, email, avatarUrl keys', () => {
    const { result } = renderHook(() => useAccount());
    const [account] = result.current;
    expect(Object.keys(account)).toEqual(expect.arrayContaining(['name', 'email', 'avatarUrl']));
  });
});

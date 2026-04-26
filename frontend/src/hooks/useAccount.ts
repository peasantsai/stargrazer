import { useState } from 'react';
import type { AccountInfo } from '../types';

const STORAGE_KEY = 'stargrazer-account';
const DEFAULTS: AccountInfo = { name: 'User', email: '', avatarUrl: '' };
const MAX_FIELD_LENGTH = 500;

/** Trims and length-limits a string to prevent oversized or padded values in storage. */
function sanitizeField(value: string): string {
  return value.trim().slice(0, MAX_FIELD_LENGTH);
}

function loadAccount(): AccountInfo {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (!saved) return DEFAULTS;
    const parsed: unknown = JSON.parse(saved);
    if (typeof parsed !== 'object' || parsed === null) return DEFAULTS;
    const p = parsed as Record<string, unknown>;
    return {
      name:      typeof p.name      === 'string' ? sanitizeField(p.name)      : DEFAULTS.name,
      email:     typeof p.email     === 'string' ? sanitizeField(p.email)     : DEFAULTS.email,
      avatarUrl: typeof p.avatarUrl === 'string' ? sanitizeField(p.avatarUrl) : DEFAULTS.avatarUrl,
    };
  } catch {
    return DEFAULTS;
  }
}

export function useAccount(): [AccountInfo, (partial: Partial<AccountInfo>) => void] {
  const [account, setAccount] = useState<AccountInfo>(loadAccount);

  const updateAccount = (partial: Partial<AccountInfo>) => {
    const next: AccountInfo = {
      name:      typeof partial.name      === 'string' ? sanitizeField(partial.name)      : account.name,
      email:     typeof partial.email     === 'string' ? sanitizeField(partial.email)     : account.email,
      avatarUrl: typeof partial.avatarUrl === 'string' ? sanitizeField(partial.avatarUrl) : account.avatarUrl,
    };
    setAccount(next);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
  };

  return [account, updateAccount];
}

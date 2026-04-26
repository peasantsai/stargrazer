import { useState } from 'react';
import type { AccountInfo } from '../types';

const STORAGE_KEY = 'stargrazer-account';
const DEFAULTS: AccountInfo = { name: 'User', email: '', avatarUrl: '' };

function loadAccount(): AccountInfo {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (!saved) return DEFAULTS;
    const parsed: unknown = JSON.parse(saved);
    if (typeof parsed !== 'object' || parsed === null) return DEFAULTS;
    const p = parsed as Record<string, unknown>;
    return {
      name:      typeof p.name      === 'string' ? p.name      : DEFAULTS.name,
      email:     typeof p.email     === 'string' ? p.email     : DEFAULTS.email,
      avatarUrl: typeof p.avatarUrl === 'string' ? p.avatarUrl : DEFAULTS.avatarUrl,
    };
  } catch {
    return DEFAULTS;
  }
}

export function useAccount(): [AccountInfo, (partial: Partial<AccountInfo>) => void] {
  const [account, setAccountState] = useState<AccountInfo>(loadAccount);

  const updateAccount = (partial: Partial<AccountInfo>) => {
    const next: AccountInfo = {
      name:      typeof partial.name      === 'string' ? partial.name      : account.name,
      email:     typeof partial.email     === 'string' ? partial.email     : account.email,
      avatarUrl: typeof partial.avatarUrl === 'string' ? partial.avatarUrl : account.avatarUrl,
    };
    setAccountState(next);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
  };

  return [account, updateAccount];
}

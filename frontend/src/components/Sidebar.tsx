import { useState } from 'react';
import type { View, AccountInfo } from '../types';
import { PLATFORM_ICONS, PLATFORM_IDS } from '../constants/platforms';
import { AccountModal } from './modals/AccountModal';

interface Props {
  readonly view: View;
  readonly setView: (v: View) => void;
  readonly browserStatus: string;
  readonly open: boolean;
  readonly onToggle: () => void;
  readonly account: AccountInfo;
  readonly updateAccount: (partial: Partial<AccountInfo>) => void;
}

const PLATFORM_LABELS: Record<string, string> = {
  facebook: 'Facebook',
  instagram: 'Instagram',
  tiktok: 'TikTok',
  youtube: 'YouTube',
  linkedin: 'LinkedIn',
  x: 'X',
};

export function Sidebar({ view, setView, browserStatus, open, onToggle, account, updateAccount }: Props) {
  const [showAccountModal, setShowAccountModal] = useState(false);

  if (!open) return null;

  const initials = account.name
    .split(' ')
    .map(w => w[0])
    .join('')
    .toUpperCase()
    .slice(0, 2) || 'U';

  return (
    <aside className="sidebar">
      <div className="sidebar-header">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <circle cx="12" cy="12" r="10" />
          <path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20" />
          <path d="M2 12h20" />
        </svg>
        <h1>Stargrazer</h1>
        <button className="sidebar-close-btn" onClick={onToggle} aria-label="Close sidebar">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <rect x="3" y="3" width="18" height="18" rx="2" />
            <line x1="9" y1="3" x2="9" y2="21" />
            <polyline points="16 16 13 12 16 8" />
          </svg>
        </button>
      </div>

      <nav className="sidebar-nav">
        {/* Core navigation */}
        <button
          className={`nav-btn ${view === 'chat' ? 'active' : ''}`}
          onClick={() => setView('chat')}
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
          </svg>
          Chat
          <span className={`status-dot ${browserStatus}`} style={{ marginLeft: 'auto' }} />
        </button>

        <button
          className={`nav-btn ${view === 'config' ? 'active' : ''}`}
          onClick={() => setView('config')}
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="12" cy="12" r="3" />
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
          </svg>
          Settings
        </button>

        {/* Platform automation pages */}
        <div className="sidebar-section-label">Platforms</div>
        {PLATFORM_IDS.map(pid => {
          const platformView: View = `platform:${pid}`;
          return (
            <button
              key={pid}
              className={`nav-btn nav-btn-platform ${view === platformView ? 'active' : ''}`}
              onClick={() => setView(platformView)}
            >
              <span className="nav-platform-icon">{PLATFORM_ICONS[pid]}</span>
              {PLATFORM_LABELS[pid]}
            </button>
          );
        })}
      </nav>

      <div className="sidebar-account">
        <button className="account-card" onClick={() => setShowAccountModal(true)}>
          <div className="account-avatar">
            {account.avatarUrl
              ? <img src={account.avatarUrl} alt="" referrerPolicy="no-referrer" />
              : initials}
          </div>
          <div className="account-info">
            <span className="account-name">{account.name}</span>
            {account.email && <span className="account-email">{account.email}</span>}
          </div>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--text-secondary)" strokeWidth="2">
            <polyline points="9 18 15 12 9 6" />
          </svg>
        </button>
      </div>

      {showAccountModal && (
        <AccountModal
          account={account}
          updateAccount={updateAccount}
          onClose={() => setShowAccountModal(false)}
        />
      )}
    </aside>
  );
}

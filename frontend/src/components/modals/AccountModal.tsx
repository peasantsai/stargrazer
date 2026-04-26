import { useState } from 'react';
import type { AccountInfo } from '../../types';

interface Props {
  readonly account: AccountInfo;
  readonly updateAccount: (partial: Partial<AccountInfo>) => void;
  readonly onClose: () => void;
}

export function AccountModal({ account, updateAccount, onClose }: Props) {
  const [name, setName] = useState(account.name);
  const [email, setEmail] = useState(account.email);
  const [avatarUrl, setAvatarUrl] = useState(account.avatarUrl);

  const initials = name.split(' ').map(w => w[0]).join('').toUpperCase().slice(0, 2) || 'U';

  const handleSave = () => {
    updateAccount({ name, email, avatarUrl });
    onClose();
  };

  return (
    <dialog
      className="modal-overlay"
      open
      onClick={e => { if (e.target === e.currentTarget) onClose(); }}
      onKeyDown={e => { if (e.key === 'Escape') onClose(); }}
    >
      <div className="modal-content">
        <div className="modal-header">
          <h3>Account Settings</h3>
          <button className="modal-close" onClick={onClose} aria-label="Close">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
        <div className="modal-body account-modal-body">
          <div className="account-avatar-edit">
            <div className="account-avatar-large">
              {avatarUrl
                ? <img src={avatarUrl} alt="" referrerPolicy="no-referrer" />
                : initials}
            </div>
            <div className="config-field" style={{ flex: 1, marginBottom: 0 }}>
              <label htmlFor="account-avatar-url">Avatar URL</label>
              <input
                id="account-avatar-url"
                type="url"
                value={avatarUrl}
                onChange={e => setAvatarUrl(e.target.value)}
                placeholder="https://example.com/avatar.png"
              />
            </div>
          </div>
          <div className="config-field" style={{ marginBottom: 0 }}>
            <label htmlFor="account-display-name">Display Name</label>
            <input
              id="account-display-name"
              type="text"
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder="Your name"
            />
          </div>
          <div className="config-field" style={{ marginBottom: 0 }}>
            <label htmlFor="account-email">Email</label>
            <input
              id="account-email"
              type="email"
              value={email}
              onChange={e => setEmail(e.target.value)}
              placeholder="you@example.com"
            />
          </div>
        </div>
        <div className="modal-footer">
          <button className="btn-primary" onClick={handleSave}>Save</button>
          <button className="btn-secondary" onClick={onClose}>Cancel</button>
        </div>
      </div>
    </dialog>
  );
}

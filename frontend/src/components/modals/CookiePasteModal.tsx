import { useState } from 'react';
import type { PlatformResponse } from '../../types';
import { PLATFORM_ICONS, PLATFORM_COLORS } from '../../constants/platforms';

interface Props {
  readonly platform: PlatformResponse;
  readonly onImport: (text: string) => void;
  readonly onCancel: () => void;
}

export function CookiePasteModal({ platform, onImport, onCancel }: Props) {
  const [cookieText, setCookieText] = useState('');
  const colors = PLATFORM_COLORS[platform.id];
  const lineCount = cookieText.split('\n').filter(l => l.trim() && !l.startsWith('#')).length;

  return (
    <div
      className="modal-overlay"
      onClick={onCancel}
      onKeyDown={e => { if (e.key === 'Escape') onCancel(); }}
    >
      <div
        className="modal-content cookie-modal"
        role="dialog"
        aria-modal="true"
        onClick={e => e.stopPropagation()}
        onKeyDown={e => e.stopPropagation()}
      >
        <div className="modal-header">
          <div className="modal-title-row">
            <div
              className="social-card-icon"
              style={{ '--platform-bg': colors?.bg || '#333', '--platform-text': '#fff' } as React.CSSProperties}
            >
              {PLATFORM_ICONS[platform.id]}
            </div>
            <div>
              <h3>Import {platform.name} Cookies</h3>
              <span style={{ fontSize: 12, color: 'var(--text-secondary)' }}>
                Paste Netscape cookie.txt from the cookies extension
              </span>
            </div>
          </div>
          <button className="modal-close" onClick={onCancel} aria-label="Close">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
        <div className="modal-body" style={{ padding: '16px 20px' }}>
          <div className="cookie-instructions">
            <div className="cookie-step">
              <span className="cookie-step-num">1</span>{' '}
              Open {platform.name} in the browser and log in
            </div>
            <div className="cookie-step">
              <span className="cookie-step-num">2</span>{' '}
              Click the cookies extension icon (pinned in toolbar)
            </div>
            <div className="cookie-step">
              <span className="cookie-step-num">3</span>{' '}
              Export as Netscape format and paste below
            </div>
          </div>
          <textarea
            className="cookie-textarea"
            placeholder={`# Netscape HTTP Cookie File\n# Paste your cookies here...\n\n.${platform.id}.com\tTRUE\t/\tTRUE\t0\tsession_id\tabc123...`}
            value={cookieText}
            onChange={e => setCookieText(e.target.value)}
            rows={12}
            spellCheck={false}
          />
          {cookieText && (
            <div className="cookie-count">
              {lineCount} cookie{lineCount === 1 ? '' : 's'} detected
            </div>
          )}
        </div>
        <div className="modal-footer">
          <button
            className="btn-primary"
            onClick={() => onImport(cookieText)}
            disabled={lineCount === 0}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="7 10 12 15 17 10" />
              <line x1="12" y1="15" x2="12" y2="3" />
            </svg>
            Import Cookies
          </button>
          <button className="btn-secondary" onClick={onCancel}>Cancel</button>
        </div>
      </div>
    </div>
  );
}

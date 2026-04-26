import { useState, useEffect } from 'react';
import {
  GetPlatforms, OpenPlatform, CheckAllLoginStatus,
  ImportCookies, PurgeSession,
} from '../../../wailsjs/go/main/App';
import type { ChatMessage, PlatformResponse } from '../../types';
import { PLATFORM_ICONS, PLATFORM_COLORS } from '../../constants/platforms';
import { CookiePasteModal } from '../modals/CookiePasteModal';

interface Props {
  readonly onBrowserStatusChange: (s: string) => void;
  readonly addMessage: (type: ChatMessage['type'], text: string) => void;
  readonly refreshPlatforms: () => void;
}

export function SocialMediaSection({ onBrowserStatusChange, addMessage, refreshPlatforms }: Props) {
  const [platforms, setPlatforms] = useState<PlatformResponse[]>([]);
  const [loadingPlatform, setLoadingPlatform] = useState<string | null>(null);
  const [cookieModal, setCookieModal] = useState<string | null>(null);
  const [infoModal, setInfoModal] = useState<PlatformResponse | null>(null);

  useEffect(() => { GetPlatforms().then(setPlatforms); }, []);

  const handleConnect = async (platformId: string) => {
    setCookieModal(platformId);
    try {
      const res = await OpenPlatform(platformId);
      if (res.status === 'running') {
        onBrowserStatusChange('running');
      } else if (res.error) {
        addMessage('error', `Browser: ${res.error}`);
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      addMessage('error', `Failed to open browser: ${msg}`);
    }
  };

  const handleOpenInBrowser = async (platformId: string) => {
    setLoadingPlatform(platformId);
    try {
      const res = await OpenPlatform(platformId);
      if (res.status === 'running') onBrowserStatusChange('running');
      else addMessage('error', `Failed: ${res.error}`);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      addMessage('error', `Error: ${msg}`);
    }
    setLoadingPlatform(null);
  };

  const handleImportCookies = async (platformId: string, cookieText: string) => {
    const name = platforms.find(p => p.id === platformId)?.name ?? platformId;
    addMessage('system', `Importing ${name} cookies...`);
    try {
      const status = await ImportCookies(platformId, cookieText);
      setPlatforms(prev => prev.map(p => p.id === platformId ? status : p));
      if (status.loggedIn) {
        addMessage('success', `${name} cookies imported! Session saved.`);
        refreshPlatforms();
      } else {
        addMessage('error', `${name}: Cookie import failed. Check the format.`);
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      addMessage('error', `Import failed: ${msg}`);
    }
    setCookieModal(null);
  };

  const handlePurgeSession = async (platformId: string) => {
    const name = platforms.find(p => p.id === platformId)?.name ?? platformId;
    const status = await PurgeSession(platformId);
    setPlatforms(prev => prev.map(p => p.id === platformId ? status : p));
    refreshPlatforms();
    setInfoModal(null);
    addMessage('info', `${name} session purged. You can reconnect.`);
  };

  const handleRefreshAll = async () => {
    try {
      const all = await CheckAllLoginStatus();
      setPlatforms(all);
      refreshPlatforms();
    } catch { /* silent */ }
  };

  const openModal = (p: PlatformResponse) => {
    if (!loadingPlatform) {
      if (p.loggedIn) handleOpenInBrowser(p.id);
      else handleConnect(p.id);
    }
  };

  return (
    <div className="config-section social-section">
      <div className="social-header">
        <h3>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
            <circle cx="9" cy="7" r="4" />
            <path d="M23 21v-2a4 4 0 0 0-3-3.87" /><path d="M16 3.13a4 4 0 0 1 0 7.75" />
          </svg>
          Social Media Connections
        </h3>
        <button className="btn-icon" onClick={handleRefreshAll} title="Refresh all">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <polyline points="23 4 23 10 17 10" />
            <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10" />
          </svg>
        </button>
      </div>

      <div className="social-grid">
        {platforms.map(p => {
          const colors = PLATFORM_COLORS[p.id] ?? { bg: '#333', hover: '#444', text: '#fff' };
          return (
            <div
              key={p.id}
              className={`social-card ${p.loggedIn ? 'logged-in' : ''}`}
              style={{ '--platform-bg': colors.bg, '--platform-hover': colors.hover, '--platform-text': colors.text } as React.CSSProperties}
            >
              <button type="button" className="social-card-main" onClick={() => openModal(p)}>
                <div className="social-card-icon">{PLATFORM_ICONS[p.id]}</div>
                <div className="social-card-info">
                  <span className="social-card-name">{p.name}</span>
                  {p.loggedIn
                    ? <span className="social-card-status connected"><span className="status-dot running" />{p.username || 'Connected'}</span>
                    : <span className="social-card-status disconnected">Click to connect</span>}
                </div>
                {p.loggedIn && p.lastLogin && (
                  <div className="social-card-meta">Since {new Date(p.lastLogin).toLocaleDateString()}</div>
                )}
              </button>
              <div className="social-card-actions">
                <button
                  className="social-info-btn"
                  onClick={() => setInfoModal(p)}
                  title="Session info"
                  aria-label={`${p.name} session info`}
                >
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <circle cx="12" cy="12" r="10" />
                    <line x1="12" y1="16" x2="12" y2="12" /><line x1="12" y1="8" x2="12.01" y2="8" />
                  </svg>
                </button>
              </div>
            </div>
          );
        })}
      </div>

      {cookieModal && (
        <CookiePasteModal
          platform={platforms.find(p => p.id === cookieModal)!}
          onImport={text => handleImportCookies(cookieModal, text)}
          onCancel={() => setCookieModal(null)}
        />
      )}

      {infoModal && (
        <div
          className="modal-overlay"
          onClick={() => setInfoModal(null)}
          onKeyDown={e => { if (e.key === 'Escape') setInfoModal(null); }}
        >
          <div
            className="modal-content"
            role="dialog"
            aria-modal="true"
            onClick={e => e.stopPropagation()}
            onKeyDown={e => e.stopPropagation()}
          >
            <div className="modal-header">
              <div className="modal-title-row">
                <div
                  className="social-card-icon"
                  style={{ '--platform-bg': PLATFORM_COLORS[infoModal.id]?.bg || '#333', '--platform-text': '#fff' } as React.CSSProperties}
                >
                  {PLATFORM_ICONS[infoModal.id]}
                </div>
                <h3>{infoModal.name}</h3>
              </div>
              <button className="modal-close" onClick={() => setInfoModal(null)} aria-label="Close">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              </button>
            </div>
            <div className="modal-body">
              <div className="modal-field">
                <span className="modal-label">Status</span>
                <span className={`modal-value ${infoModal.loggedIn ? 'text-success' : 'text-muted'}`}>
                  <span className={`status-dot ${infoModal.loggedIn ? 'running' : 'stopped'}`} />
                  {infoModal.loggedIn ? 'Connected' : 'Not connected'}
                </span>
              </div>
              {infoModal.username && (
                <div className="modal-field">
                  <span className="modal-label">User / ID</span>
                  <span className="modal-value">{infoModal.username}</span>
                </div>
              )}
              <div className="modal-field">
                <span className="modal-label">URL</span>
                <span className="modal-value modal-url">{infoModal.url}</span>
              </div>
              <div className="modal-field">
                <span className="modal-label">Session Directory</span>
                <span className="modal-value modal-path">{infoModal.sessionDir}</span>
              </div>
              {infoModal.lastLogin && (
                <div className="modal-field">
                  <span className="modal-label">Logged In</span>
                  <span className="modal-value">{new Date(infoModal.lastLogin).toLocaleString()}</span>
                </div>
              )}
              {infoModal.lastCheck && (
                <div className="modal-field">
                  <span className="modal-label">Last Verified</span>
                  <span className="modal-value">{new Date(infoModal.lastCheck).toLocaleString()}</span>
                </div>
              )}
            </div>
            <div className="modal-footer">
              <button
                className="btn-primary"
                onClick={() => { handleOpenInBrowser(infoModal.id); setInfoModal(null); }}
              >
                {infoModal.loggedIn ? 'Open' : 'Connect'}
              </button>
              {infoModal.loggedIn && (
                <button className="btn-danger" onClick={() => handlePurgeSession(infoModal.id)}>
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <polyline points="3 6 5 6 21 6" />
                    <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
                  </svg>
                  Purge Session
                </button>
              )}
              <button className="btn-secondary" onClick={() => setInfoModal(null)}>Close</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

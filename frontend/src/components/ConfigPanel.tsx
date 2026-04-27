import { useState, useEffect } from 'react';
import {
  GetBrowserConfig, UpdateBrowserConfig, ResetBrowserConfig,
  GetBrowserStatus, RestartBrowser,
} from '../../wailsjs/go/main/App';
import type { ChatMessage, BrowserConfigResponse, PlatformResponse, Theme } from '../types';
import { CHROMIUM_FLAGS } from '../constants/chromiumFlags';
import { HamburgerBtn } from './HamburgerBtn';
import { LogsModal } from './modals/LogsModal';

function isKnownFlag(f: string): boolean {
  return CHROMIUM_FLAGS.some(g => g.flags.some(gf => gf.flag === f));
}

interface Props {
  readonly onSaved: (msg: string) => void;
  readonly sidebarOpen: boolean;
  readonly onToggleSidebar: () => void;
  readonly onBrowserStatusChange: (s: string) => void;
  readonly addMessage: (type: ChatMessage['type'], text: string) => void;
  readonly browserStatus: string;
  readonly loading: boolean;
  readonly onStartBrowser: () => void;
  readonly onStopBrowser: () => void;
  readonly theme: Theme;
  readonly setTheme: (t: Theme) => void;
  readonly platforms: PlatformResponse[];
}

export function ConfigPanel({
  onSaved, sidebarOpen, onToggleSidebar, onBrowserStatusChange, addMessage,
  browserStatus, loading, onStartBrowser, onStopBrowser, theme, setTheme, platforms,
}: Props) {
  const [config, setConfig] = useState<BrowserConfigResponse | null>(null);
  const [saving, setSaving] = useState(false);
  const [resetting, setResetting] = useState(false);
  const [expandedCategories, setExpandedCategories] = useState<Set<string>>(new Set());
  const [search, setSearch] = useState('');
  const [showLogs, setShowLogs] = useState(false);

  useEffect(() => { GetBrowserConfig().then(setConfig); }, []);

  const updateField = <K extends keyof BrowserConfigResponse>(key: K, value: BrowserConfigResponse[K]) => {
    if (config) setConfig({ ...config, [key]: value });
  };

  const toggleFlag = (flag: string) => {
    if (!config) return;
    const current = config.extraFlags ?? [];
    updateField('extraFlags', current.includes(flag) ? current.filter(f => f !== flag) : [...current, flag]);
  };

  const handleCustomFlagsChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const current = config?.extraFlags ?? [];
    const known = current.filter(isKnownFlag);
    updateField('extraFlags', [
      ...known,
      ...e.target.value.split(',').map(s => s.trim()).filter(Boolean),
    ]);
  };

  const toggleCategory = (cat: string) => {
    setExpandedCategories(prev => {
      const next = new Set(prev);
      if (next.has(cat)) { next.delete(cat); } else { next.add(cat); }
      return next;
    });
  };

  const handleSave = async () => {
    if (!config) return;
    setSaving(true);
    try {
      const updated = await UpdateBrowserConfig(config);
      setConfig(updated);
      const bStatus = await GetBrowserStatus();
      if (bStatus.status === 'running') {
        addMessage('system', 'Restarting browser with new settings...');
        const res = await RestartBrowser();
        onBrowserStatusChange(res.status);
        addMessage(
          res.status === 'running' ? 'success' : 'error',
          res.status === 'running' ? 'Browser restarted with new settings.' : `Restart failed: ${res.error}`,
        );
      } else {
        onSaved('Configuration saved.');
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      onSaved(`Error: ${msg}`);
    }
    setSaving(false);
  };

  const handleReset = async () => {
    setResetting(true);
    try {
      const d = await ResetBrowserConfig();
      setConfig(d);
      onSaved('Settings reset to defaults.');
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      onSaved(`Error: ${msg}`);
    }
    setResetting(false);
  };

  if (!config) return <div className="config-panel"><p>Loading...</p></div>;

  const isRunning = browserStatus === 'running';
  const connectedPlatforms = platforms.filter(p => p.loggedIn).length;
  const activeFlags = config.extraFlags ?? [];
  const activeCount = activeFlags.length;
  const q = search.toLowerCase().trim();
  const match = (...terms: string[]) => !q || terms.some(t => t.toLowerCase().includes(q));

  const showConnection = match('connection', 'cdp', 'port', 'chromium', 'path', 'user data', 'directory');
  const showDisplay = match('display', 'headless', 'window', 'width', 'height');
  const showCustomFlags = match('custom', 'flags', 'extra', 'additional');
  const showTheme = match('theme', 'dark', 'light', 'appearance');
  const filteredFlagGroups = CHROMIUM_FLAGS
    .map(g => {
      if (!q) return { ...g, filteredFlags: g.flags };
      const catMatch = g.category.toLowerCase().includes(q);
      return {
        ...g,
        filteredFlags: catMatch ? g.flags : g.flags.filter(
          f => f.label.toLowerCase().includes(q) || f.description.toLowerCase().includes(q) || f.flag.toLowerCase().includes(q),
        ),
      };
    })
    .filter(g => g.filteredFlags.length > 0);
  const showFlags = filteredFlagGroups.length > 0 || match('chromium', 'flags');
  const noResults = q && !showConnection && !showDisplay && !showFlags && !showCustomFlags && !showTheme;

  return (
    <div className="config-panel">
      <div className="config-header">
        <HamburgerBtn sidebarOpen={sidebarOpen} onToggle={onToggleSidebar} />
        <h2>Settings</h2>
        <button className="btn-icon" style={{ marginLeft: 'auto' }} onClick={() => setShowLogs(true)} title="View logs">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
            <line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/>
            <polyline points="10 9 9 9 8 9"/>
          </svg>
        </button>
      </div>

      {/* Stats bar */}
      <div className="settings-stats">
        <div className="settings-stat">
          <span className={`status-dot ${browserStatus}`} />
          <span className="settings-stat-label">Browser</span>
          <span className="settings-stat-value">{browserStatus}</span>
        </div>
        <div className="settings-stat">
          <span className="settings-stat-label">Platforms</span>
          <span className="settings-stat-value">{connectedPlatforms}/{platforms.length} connected</span>
        </div>
        <div className="settings-stat">
          <span className="settings-stat-label">Flags</span>
          <span className="settings-stat-value">{activeCount} active</span>
        </div>
        <div className="settings-stat-actions">
          {isRunning ? (
            <button className="btn-danger" onClick={onStopBrowser} disabled={loading} style={{ padding: '6px 14px', fontSize: 12 }}>
              {loading ? 'Stopping...' : 'Stop Browser'}
            </button>
          ) : (
            <button className="btn-primary" onClick={onStartBrowser} disabled={loading} style={{ padding: '6px 14px', fontSize: 12 }}>
              {loading ? 'Starting...' : 'Start Browser'}
            </button>
          )}
        </div>
      </div>

      <div className="config-search">
        <svg className="config-search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
        </svg>
        <input
          type="text"
          className="config-search-input"
          placeholder="Search settings..."
          value={search}
          onChange={e => setSearch(e.target.value)}
        />
        {search && (
          <button className="config-search-clear" onClick={() => setSearch('')}>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        )}
      </div>

      {noResults && <div className="config-no-results">No settings match "{search}"</div>}

      {showTheme && (
        <div className="config-section">
          <h3>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="12" cy="12" r="4" />
              <path d="M12 2v2m0 16v2M4.93 4.93l1.41 1.41m11.32 11.32l1.41 1.41M2 12h2m16 0h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41" />
            </svg>
            Appearance
          </h3>
          <div className="config-field">
            <span className="config-field-label">Theme</span>
            <div className="theme-switcher" style={{ alignSelf: 'flex-start' }}>
              <button className={theme === 'dark' ? 'active' : ''} onClick={() => setTheme('dark')} aria-label="Dark theme">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
                </svg>
                Dark
              </button>
              <button className={theme === 'light' ? 'active' : ''} onClick={() => setTheme('light')} aria-label="Light theme">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <circle cx="12" cy="12" r="5" />
                  <line x1="12" y1="1" x2="12" y2="3" /><line x1="12" y1="21" x2="12" y2="23" />
                  <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" /><line x1="18.36" y1="18.36" x2="19.78" y2="19.78" />
                  <line x1="1" y1="12" x2="3" y2="12" /><line x1="21" y1="12" x2="23" y2="12" />
                  <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" /><line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
                </svg>
                Light
              </button>
            </div>
          </div>
        </div>
      )}

      {showConnection && (
        <div className="config-section">
          <h3>Connection</h3>
          <div className="config-field">
            <label htmlFor="cfg-cdp-port">CDP Port</label>
            <input
              id="cfg-cdp-port"
              type="number"
              value={config.cdpPort}
              onChange={e => updateField('cdpPort', Number.parseInt(e.target.value) || 9222)}
            />
          </div>
          <div className="config-field">
            <label htmlFor="cfg-chromium-path">Chromium Path</label>
            <input
              id="cfg-chromium-path"
              type="text"
              value={config.chromiumPath}
              onChange={e => updateField('chromiumPath', e.target.value)}
              placeholder="Auto-detect"
            />
            <span className="config-hint">Auto-detected from bundled assets.</span>
          </div>
          <div className="config-field">
            <label htmlFor="cfg-user-data-dir">User Data Directory</label>
            <input
              id="cfg-user-data-dir"
              type="text"
              value={config.userDataDir}
              onChange={e => updateField('userDataDir', e.target.value)}
              placeholder="Default"
            />
          </div>
        </div>
      )}

      {showDisplay && (
        <div className="config-section">
          <h3>Display</h3>
          <div className="config-field">
            <span className="config-field-label">Headless Mode</span>
            <div className="config-field-row">
              <button className={`toggle ${config.headless ? 'active' : ''}`} onClick={() => updateField('headless', !config.headless)} />
              <span>{config.headless ? 'Enabled' : 'Disabled'}</span>
            </div>
          </div>
          <div className="config-field-inline">
            <div className="config-field">
              <label htmlFor="cfg-window-width">Width</label>
              <input
                id="cfg-window-width"
                type="number"
                value={config.windowWidth}
                onChange={e => updateField('windowWidth', Number.parseInt(e.target.value) || 1280)}
              />
            </div>
            <div className="config-field">
              <label htmlFor="cfg-window-height">Height</label>
              <input
                id="cfg-window-height"
                type="number"
                value={config.windowHeight}
                onChange={e => updateField('windowHeight', Number.parseInt(e.target.value) || 900)}
              />
            </div>
          </div>
        </div>
      )}

      {showFlags && (
        <div className="config-section">
          <h3>
            Chromium Flags
            {activeCount > 0 && <span className="flag-badge">{activeCount} active</span>}
          </h3>
          {filteredFlagGroups.map(group => {
            const expanded = q ? true : expandedCategories.has(group.category);
            const groupActive = group.filteredFlags.filter(f => activeFlags.includes(f.flag)).length;
            return (
              <div key={group.category} className="flag-group">
                <button className="flag-group-header" onClick={() => toggleCategory(group.category)}>
                  <svg className={`flag-chevron ${expanded ? 'expanded' : ''}`} width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <polyline points="9 18 15 12 9 6"/>
                  </svg>
                  <span>{group.category}</span>
                  {groupActive > 0 && <span className="flag-badge">{groupActive}</span>}
                </button>
                {expanded && (
                  <div className="flag-list">
                    {group.filteredFlags.map(({ flag, label, description, dangerous }) => (
                      <label key={flag} className={`flag-item ${dangerous ? 'flag-item-danger' : ''}`} title={description}>
                        <input
                          type="checkbox"
                          checked={activeFlags.includes(flag)}
                          onChange={() => toggleFlag(flag)}
                        />
                        <div className="flag-item-content">
                          <span className="flag-label">
                            {label}
                            {dangerous && <span className="flag-danger-badge" title="Security risk">⚠</span>}
                          </span>
                          <span className="flag-desc">{description}</span>
                        </div>
                        <code className="flag-code">{flag}</code>
                      </label>
                    ))}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}

      {showCustomFlags && (
        <div className="config-section">
          <h3>Custom Flags</h3>
          <div className="config-field">
            <label htmlFor="cfg-custom-flags">Additional flags (comma-separated)</label>
            <input
              id="cfg-custom-flags"
              type="text"
              value={activeFlags.filter(f => !isKnownFlag(f)).join(', ')}
              onChange={handleCustomFlagsChange}
              placeholder="--proxy-server=host:port"
            />
          </div>
        </div>
      )}

      <div className="config-actions">
        <button className="btn-primary" onClick={handleSave} disabled={saving}>
          {saving ? 'Saving...' : 'Save Settings'}
        </button>
        <button className="btn-secondary" onClick={handleReset} disabled={resetting}>
          {resetting ? 'Resetting...' : 'Reset to Defaults'}
        </button>
      </div>

      {showLogs && <LogsModal onClose={() => setShowLogs(false)} />}
    </div>
  );
}

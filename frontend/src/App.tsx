import { useState, useEffect, useRef, useCallback } from 'react';
import {
  StartBrowser, StopBrowser, GetBrowserStatus, GetBrowserConfig,
  UpdateBrowserConfig, ResetBrowserConfig, RestartBrowser,
  GetPlatforms, OpenPlatform, CheckAllLoginStatus,
  ImportCookies,
  GetLogs, ExportLogs, ClearLogs, TriggerUpload,
  GetSchedules, CreateSchedule, DeleteSchedule, PauseSchedule, ResumeSchedule,
  PurgeSession, SelectFile,
} from '../wailsjs/go/main/App';
import type { main } from '../wailsjs/go/models';

type View = 'chat' | 'schedules' | 'config';

interface ChatMessage {
  id: number;
  type: 'system' | 'info' | 'error' | 'success';
  text: string;
}

const CHROMIUM_FLAGS: { category: string; flags: { flag: string; label: string; description: string }[] }[] = [
  { category: 'Stealth & Anti-Detection', flags: [
    { flag: '--disable-blink-features=AutomationControlled', label: 'Hide WebDriver Flag', description: 'Remove navigator.webdriver from the JS runtime' },
    { flag: '--disable-features=AutomationControlled', label: 'No Automation Signal', description: 'Remove the automation-controlled feature flag' },
    { flag: '--disable-infobars', label: 'Disable Infobars', description: 'Suppress "Chrome is being controlled" bar' },
    { flag: '--disable-background-networking', label: 'No Background Network', description: 'Prevent background network requests' },
    { flag: '--disable-client-side-phishing-detection', label: 'No Phishing Detection', description: 'Disable phishing detection' },
    { flag: '--disable-component-update', label: 'No Component Updates', description: 'Prevent auto-updates' },
    { flag: '--disable-default-apps', label: 'No Default Apps', description: 'Skip default apps' },
    { flag: '--disable-domain-reliability', label: 'No Domain Reliability', description: 'Disable monitoring' },
    { flag: '--metrics-recording-only', label: 'Metrics Record Only', description: 'Never upload metrics' },
    { flag: '--no-pings', label: 'No Pings', description: 'Disable auditing pings' },
    { flag: '--safebrowsing-disable-auto-update', label: 'No SafeBrowsing Updates', description: 'No SB updates' },
  ]},
  { category: 'Privacy & Telemetry', flags: [
    { flag: '--disable-extensions', label: 'Disable Extensions', description: 'No extensions' },
    { flag: '--disable-sync', label: 'Disable Sync', description: 'No Google sync' },
    { flag: '--disable-translate', label: 'Disable Translate', description: 'No translation' },
    { flag: '--incognito', label: 'Incognito Mode', description: 'Private browsing' },
  ]},
  { category: 'Automation & CDP', flags: [
    { flag: '--disable-hang-monitor', label: 'Disable Hang Monitor', description: 'No unresponsive dialog' },
    { flag: '--disable-popup-blocking', label: 'Disable Popup Blocking', description: 'Allow popups' },
    { flag: '--disable-prompt-on-repost', label: 'No Repost Prompt', description: 'No resubmission dialog' },
    { flag: '--disable-ipc-flooding-protection', label: 'No IPC Flood Protection', description: 'Rapid CDP commands' },
    { flag: '--disable-renderer-backgrounding', label: 'No Renderer Backgrounding', description: 'Full CPU priority' },
    { flag: '--disable-background-timer-throttling', label: 'No Timer Throttling', description: 'No background throttle' },
    { flag: '--disable-backgrounding-occluded-windows', label: 'No Occluded Throttling', description: 'Keep hidden active' },
    { flag: '--enable-features=NetworkService,NetworkServiceInProcess', label: 'In-Process Network', description: 'Faster CDP' },
  ]},
  { category: 'Display & UI', flags: [
    { flag: '--force-dark-mode', label: 'Force Dark Mode', description: 'Dark browser chrome' },
    { flag: '--enable-features=WebUIDarkMode', label: 'WebUI Dark Mode', description: 'Dark internal pages' },
    { flag: '--disable-notifications', label: 'Disable Notifications', description: 'Block notifications' },
    { flag: '--start-maximized', label: 'Start Maximized', description: 'Maximized window' },
    { flag: '--start-fullscreen', label: 'Start Fullscreen', description: 'Fullscreen mode' },
    { flag: '--hide-scrollbars', label: 'Hide Scrollbars', description: 'No scrollbars' },
  ]},
  { category: 'Network', flags: [
    { flag: '--ignore-certificate-errors', label: 'Ignore Cert Errors', description: 'Skip SSL errors' },
    { flag: '--disable-web-security', label: 'Disable Web Security', description: 'No CORS' },
    { flag: '--allow-running-insecure-content', label: 'Allow Insecure Content', description: 'Mixed content' },
    { flag: '--disable-gpu', label: 'Disable GPU', description: 'No GPU acceleration' },
  ]},
];

const PLATFORM_ICONS: Record<string, JSX.Element> = {
  facebook: <svg viewBox="0 0 24 24" fill="currentColor"><path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"/></svg>,
  instagram: <svg viewBox="0 0 24 24" fill="currentColor"><path d="M12 2.163c3.204 0 3.584.012 4.85.07 3.252.148 4.771 1.691 4.919 4.919.058 1.265.069 1.645.069 4.849 0 3.205-.012 3.584-.069 4.849-.149 3.225-1.664 4.771-4.919 4.919-1.266.058-1.644.07-4.85.07-3.204 0-3.584-.012-4.849-.07-3.26-.149-4.771-1.699-4.919-4.92-.058-1.265-.07-1.644-.07-4.849 0-3.204.013-3.583.07-4.849.149-3.227 1.664-4.771 4.919-4.919 1.266-.057 1.645-.069 4.849-.069zM12 0C8.741 0 8.333.014 7.053.072 2.695.272.273 2.69.073 7.052.014 8.333 0 8.741 0 12c0 3.259.014 3.668.072 4.948.2 4.358 2.618 6.78 6.98 6.98C8.333 23.986 8.741 24 12 24c3.259 0 3.668-.014 4.948-.072 4.354-.2 6.782-2.618 6.979-6.98.059-1.28.073-1.689.073-4.948 0-3.259-.014-3.667-.072-4.947-.196-4.354-2.617-6.78-6.979-6.98C15.668.014 15.259 0 12 0zm0 5.838a6.162 6.162 0 100 12.324 6.162 6.162 0 000-12.324zM12 16a4 4 0 110-8 4 4 0 010 8zm6.406-11.845a1.44 1.44 0 100 2.881 1.44 1.44 0 000-2.881z"/></svg>,
  tiktok: <svg viewBox="0 0 24 24" fill="currentColor"><path d="M12.525.02c1.31-.02 2.61-.01 3.91-.02.08 1.53.63 3.09 1.75 4.17 1.12 1.11 2.7 1.62 4.24 1.79v4.03c-1.44-.05-2.89-.35-4.2-.97-.57-.26-1.1-.59-1.62-.93-.01 2.92.01 5.84-.02 8.75-.08 1.4-.54 2.79-1.35 3.94-1.31 1.92-3.58 3.17-5.91 3.21-1.43.08-2.86-.31-4.08-1.03-2.02-1.19-3.44-3.37-3.65-5.71-.02-.5-.03-1-.01-1.49.18-1.9 1.12-3.72 2.58-4.96 1.66-1.44 3.98-2.13 6.15-1.72.02 1.48-.04 2.96-.04 4.44-.99-.32-2.15-.23-3.02.37-.63.41-1.11 1.04-1.36 1.75-.21.51-.15 1.07-.14 1.61.24 1.64 1.82 3.02 3.5 2.87 1.12-.01 2.19-.66 2.77-1.61.19-.33.4-.67.41-1.06.1-1.79.06-3.57.07-5.36.01-4.03-.01-8.05.02-12.07z"/></svg>,
  youtube: <svg viewBox="0 0 24 24" fill="currentColor"><path d="M23.498 6.186a3.016 3.016 0 00-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 00.502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 002.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 002.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814zM9.545 15.568V8.432L15.818 12l-6.273 3.568z"/></svg>,
  linkedin: <svg viewBox="0 0 24 24" fill="currentColor"><path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433a2.062 2.062 0 01-2.063-2.065 2.064 2.064 0 112.063 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"/></svg>,
  x: <svg viewBox="0 0 24 24" fill="currentColor"><path d="M18.901 1.153h3.68l-8.04 9.19L24 22.846h-7.406l-5.8-7.584-6.638 7.584H.474l8.6-9.83L0 1.154h7.594l5.243 6.932L18.901 1.153zM17.61 20.644h2.039L6.486 3.24H4.298L17.61 20.644z"/></svg>,
};

const PLATFORM_COLORS: Record<string, { bg: string; hover: string; text: string }> = {
  facebook:  { bg: '#1877F2', hover: '#166FE5', text: '#fff' },
  instagram: { bg: 'linear-gradient(45deg, #f09433, #e6683c, #dc2743, #cc2366, #bc1888)', hover: 'linear-gradient(45deg, #e08529, #d55d35, #cc2040, #bb1f5c, #aa1580)', text: '#fff' },
  tiktok:    { bg: '#010101', hover: '#1a1a1a', text: '#fff' },
  youtube:   { bg: '#FF0000', hover: '#CC0000', text: '#fff' },
  linkedin:  { bg: '#0A66C2', hover: '#004182', text: '#fff' },
  x:         { bg: '#000000', hover: '#1a1a1a', text: '#fff' },
};

let msgId = 0;

type Theme = 'dark' | 'light';

interface AccountInfo {
  name: string;
  email: string;
  avatarUrl: string;
}

function useTheme(): [Theme, (t: Theme) => void] {
  const [theme, setThemeState] = useState<Theme>(() => (localStorage.getItem('stargrazer-theme') as Theme) || 'dark');
  const setTheme = (t: Theme) => { setThemeState(t); localStorage.setItem('stargrazer-theme', t); document.documentElement.dataset.theme = t; };
  useEffect(() => { document.documentElement.dataset.theme = theme; }, []);
  return [theme, setTheme];
}

function useAccount(): [AccountInfo, (a: Partial<AccountInfo>) => void] {
  const defaults: AccountInfo = { name: 'User', email: '', avatarUrl: '' };
  const [account, setAccountState] = useState<AccountInfo>(() => {
    try {
      const saved = localStorage.getItem('stargrazer-account');
      if (!saved) return defaults;
      const parsed = JSON.parse(saved);
      // Validate shape before trusting stored data
      return {
        name: typeof parsed.name === 'string' ? parsed.name : defaults.name,
        email: typeof parsed.email === 'string' ? parsed.email : defaults.email,
        avatarUrl: typeof parsed.avatarUrl === 'string' ? parsed.avatarUrl : defaults.avatarUrl,
      };
    } catch { return defaults; }
  });
  const updateAccount = (partial: Partial<AccountInfo>) => {
    const next: AccountInfo = {
      name: typeof partial.name === 'string' ? partial.name : account.name,
      email: typeof partial.email === 'string' ? partial.email : account.email,
      avatarUrl: typeof partial.avatarUrl === 'string' ? partial.avatarUrl : account.avatarUrl,
    };
    setAccountState(next);
    localStorage.setItem('stargrazer-account', JSON.stringify(next));
  };
  return [account, updateAccount];
}

function App() {
  const [view, setView] = useState<View>('chat');
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [browserStatus, setBrowserStatus] = useState('stopped');
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [loading, setLoading] = useState(false);
  const [platforms, setPlatforms] = useState<main.PlatformResponse[]>([]);
  const [theme, setTheme] = useTheme();
  const [account, updateAccount] = useAccount();
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const addMessage = useCallback((type: ChatMessage['type'], text: string) => {
    setMessages(prev => [...prev, { id: ++msgId, type, text }]);
  }, []);

  useEffect(() => { messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' }); }, [messages]);
  useEffect(() => { GetBrowserStatus().then(r => setBrowserStatus(r.status)); }, []);
  useEffect(() => { GetPlatforms().then(setPlatforms); }, []);
  // Refresh platforms whenever user switches to chat view (picks up new logins)
  useEffect(() => { if (view === 'chat') GetPlatforms().then(setPlatforms); }, [view]);

  const refreshPlatforms = useCallback(() => { GetPlatforms().then(setPlatforms); }, []);

  const handleStartBrowser = async () => {
    setLoading(true);
    addMessage('system', 'Starting browser...');
    const res = await StartBrowser();
    setBrowserStatus(res.status);
    addMessage(res.status === 'running' ? 'success' : 'error', res.status === 'running' ? 'Browser started. CDP active.' : `Failed: ${res.error}`);
    setLoading(false);
  };

  const handleStopBrowser = async () => {
    setLoading(true);
    addMessage('system', 'Stopping browser...');
    const res = await StopBrowser();
    setBrowserStatus(res.status);
    addMessage('info', 'Browser stopped.');
    setLoading(false);
  };

  return (
    <div className="app-layout">
      <Sidebar view={view} setView={setView} browserStatus={browserStatus} open={sidebarOpen} onToggle={() => setSidebarOpen(p => !p)} theme={theme} setTheme={setTheme} account={account} updateAccount={updateAccount} />
      <div className="main-content">
        {view === 'chat' && (
          <ChatPanel
            messages={messages} browserStatus={browserStatus} loading={loading}
            onStart={handleStartBrowser} onStop={handleStopBrowser}
            messagesEndRef={messagesEndRef} sidebarOpen={sidebarOpen}
            onToggleSidebar={() => setSidebarOpen(true)}
            platforms={platforms} addMessage={addMessage}
          />
        )}
        {view === 'schedules' && (
          <SchedulesPanel sidebarOpen={sidebarOpen} onToggleSidebar={() => setSidebarOpen(true)} addMessage={addMessage} platforms={platforms} />
        )}
        {view === 'config' && (
          <ConfigPanel
            onSaved={msg => addMessage('success', msg)} sidebarOpen={sidebarOpen}
            onToggleSidebar={() => setSidebarOpen(true)}
            onBrowserStatusChange={setBrowserStatus} addMessage={addMessage}
            refreshPlatforms={refreshPlatforms}
          />
        )}
      </div>
    </div>
  );
}

/* --- Sidebar --- */
function Sidebar({ view, setView, browserStatus, open, onToggle, theme, setTheme, account, updateAccount }: {
  view: View; setView: (v: View) => void; browserStatus: string; open: boolean; onToggle: () => void;
  theme: Theme; setTheme: (t: Theme) => void; account: AccountInfo; updateAccount: (a: Partial<AccountInfo>) => void;
}) {
  const [showAccountModal, setShowAccountModal] = useState(false);
  if (!open) return null;

  const initials = account.name.split(' ').map(w => w[0]).join('').toUpperCase().slice(0, 2) || 'U';

  return (
    <aside className="sidebar">
      <div className="sidebar-header">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"/><path d="M2 12h20"/></svg>
        <h1>Stargrazer</h1>
        <button className="sidebar-close-btn" onClick={onToggle}><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><rect x="3" y="3" width="18" height="18" rx="2"/><line x1="9" y1="3" x2="9" y2="21"/><polyline points="16 16 13 12 16 8"/></svg></button>
      </div>

      <nav className="sidebar-nav">
        <button className={`nav-btn ${view === 'chat' ? 'active' : ''}`} onClick={() => setView('chat')}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
          Chat <span className={`status-dot ${browserStatus}`} style={{ marginLeft: 'auto' }} />
        </button>
        <button className={`nav-btn ${view === 'schedules' ? 'active' : ''}`} onClick={() => setView('schedules')}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
          Schedules
        </button>
        <button className={`nav-btn ${view === 'config' ? 'active' : ''}`} onClick={() => setView('config')}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
          Settings
        </button>
      </nav>

      <div className="sidebar-account">
        <div className="theme-toggle-row">
          <span>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="4"/><path d="M12 2v2m0 16v2M4.93 4.93l1.41 1.41m11.32 11.32l1.41 1.41M2 12h2m16 0h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41"/></svg>
            Theme
          </span>
          <div className="theme-switcher">
            <button className={theme === 'dark' ? 'active' : ''} onClick={() => setTheme('dark')}>
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
            </button>
            <button className={theme === 'light' ? 'active' : ''} onClick={() => setTheme('light')}>
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
            </button>
          </div>
        </div>

        <button className="account-card" onClick={() => setShowAccountModal(true)}>
          <div className="account-avatar">
            {account.avatarUrl ? <img src={account.avatarUrl} alt="" /> : initials}
          </div>
          <div className="account-info">
            <span className="account-name">{account.name}</span>
            {account.email && <span className="account-email">{account.email}</span>}
          </div>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--text-secondary)" strokeWidth="2"><polyline points="9 18 15 12 9 6"/></svg>
        </button>
      </div>

      {showAccountModal && (
        <AccountModal account={account} updateAccount={updateAccount} onClose={() => setShowAccountModal(false)} />
      )}
    </aside>
  );
}

/* --- Account Settings Modal --- */
function AccountModal({ account, updateAccount, onClose }: { account: AccountInfo; updateAccount: (a: Partial<AccountInfo>) => void; onClose: () => void }) {
  const [name, setName] = useState(account.name);
  const [email, setEmail] = useState(account.email);
  const [avatarUrl, setAvatarUrl] = useState(account.avatarUrl);
  const initials = name.split(' ').map(w => w[0]).join('').toUpperCase().slice(0, 2) || 'U';

  const handleSave = () => { updateAccount({ name, email, avatarUrl }); onClose(); };

  return (
    <div className="modal-overlay" onClick={onClose} onKeyDown={e => e.key === 'Escape' && onClose()} role="presentation">
      <div className="modal-content" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h3>Account Settings</h3>
          <button className="modal-close" onClick={onClose}><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button>
        </div>
        <div className="modal-body account-modal-body">
          <div className="account-avatar-edit">
            <div className="account-avatar-large">
              {avatarUrl ? <img src={avatarUrl} alt="" /> : initials}
            </div>
            <div className="config-field" style={{ flex: 1, marginBottom: 0 }}>
              <label htmlFor="account-avatar-url">Avatar URL</label>
              <input id="account-avatar-url" type="text" value={avatarUrl} onChange={e => setAvatarUrl(e.target.value)} placeholder="https://example.com/avatar.png" />
            </div>
          </div>
          <div className="config-field" style={{ marginBottom: 0 }}>
            <label htmlFor="account-display-name">Display Name</label>
            <input id="account-display-name" type="text" value={name} onChange={e => setName(e.target.value)} placeholder="Your name" />
          </div>
          <div className="config-field" style={{ marginBottom: 0 }}>
            <label htmlFor="account-email">Email</label>
            <input id="account-email" type="text" value={email} onChange={e => setEmail(e.target.value)} placeholder="you@example.com" />
          </div>
        </div>
        <div className="modal-footer">
          <button className="btn-primary" onClick={handleSave}>Save</button>
          <button className="btn-secondary" onClick={onClose}>Cancel</button>
        </div>
      </div>
    </div>
  );
}

/* --- Hamburger --- */
function HamburgerBtn({ sidebarOpen, onToggle }: { sidebarOpen: boolean; onToggle: () => void }) {
  if (sidebarOpen) return null;
  return (
    <button className="sidebar-toggle-float" onClick={onToggle}>
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="18" x2="21" y2="18"/></svg>
    </button>
  );
}

/* --- Chat Panel with Upload Form --- */
function ChatPanel({ messages, browserStatus, loading, onStart, onStop, messagesEndRef, sidebarOpen, onToggleSidebar, platforms, addMessage }: {
  messages: ChatMessage[]; browserStatus: string; loading: boolean;
  onStart: () => void; onStop: () => void;
  messagesEndRef: React.RefObject<HTMLDivElement>;
  sidebarOpen: boolean; onToggleSidebar: () => void;
  platforms: main.PlatformResponse[]; addMessage: (t: ChatMessage['type'], m: string) => void;
}) {
  const isRunning = browserStatus === 'running';
  const [caption, setCaption] = useState('');
  const [tags, setTags] = useState<string[]>([]);
  const [tagInput, setTagInput] = useState('');
  const [selectedFile, setSelectedFile] = useState('');
  const [selectedFileName, setSelectedFileName] = useState('');
  const [selectedPlatforms, setSelectedPlatforms] = useState<Set<string>>(new Set());
  const [uploading, setUploading] = useState(false);

  const togglePlatform = (id: string) => {
    setSelectedPlatforms(prev => { const n = new Set(prev); n.has(id)?n.delete(id):n.add(id); return n; });
  };

  const handleTagKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if ((e.key === ' ' || e.key === 'Enter' || e.key === ',') && tagInput.trim()) {
      e.preventDefault();
      const raw = tagInput.trim().replace(/^#/, '');
      if (raw && !tags.includes(`#${raw}`)) {
        setTags(prev => [...prev, `#${raw}`]);
      }
      setTagInput('');
    } else if (e.key === 'Backspace' && !tagInput && tags.length > 0) {
      setTags(prev => prev.slice(0, -1));
    }
  };

  const removeTag = (tag: string) => setTags(prev => prev.filter(t => t !== tag));

  const handleSelectFile = async () => {
    const path = await SelectFile();
    if (path) {
      setSelectedFile(path);
      // Extract filename from full path
      const name = path.split(/[/\\]/).pop() || path;
      setSelectedFileName(name);
    }
  };

  const handleSend = async () => {
    if (!isRunning) { addMessage('error', 'Start the browser first.'); return; }
    if (selectedPlatforms.size === 0) { addMessage('error', 'Select at least one platform.'); return; }
    // Include any text still in the tag input
    const finalTags = [...tags];
    if (tagInput.trim()) {
      const raw = tagInput.trim().replace(/^#/, '');
      if (raw) finalTags.push(`#${raw}`);
    }
    if (!selectedFile && !caption.trim() && finalTags.length === 0) {
      addMessage('error', 'Provide at least a file, caption, or hashtags.'); return;
    }

    setUploading(true);
    const platformNames = [...selectedPlatforms].map(id => platforms.find(p => p.id === id)?.name || id).join(', ');
    addMessage('system', `Uploading to ${platformNames}...`);
    if (selectedFileName) addMessage('info', `File: ${selectedFileName}`);
    if (caption.trim()) addMessage('info', `Caption: ${caption.trim()}`);
    if (finalTags.length > 0) addMessage('info', `Tags: ${finalTags.join(' ')}`);

    try {
      const res = await TriggerUpload({
        platforms: [...selectedPlatforms],
        filePath: selectedFile,
        caption: caption.trim(),
        hashtags: finalTags,
      } as any);
      addMessage(res.success ? 'success' : 'error', res.message);
      if (res.success) {
        setCaption(''); setTags([]); setTagInput('');
        setSelectedFile(''); setSelectedFileName('');
      }
    } catch (err: any) {
      addMessage('error', `Upload error: ${err?.message || err}`);
    }
    setUploading(false);
  };

  return (
    <div className="chat-panel">
      <div className="chat-header">
        <div className="chat-header-left">
          <HamburgerBtn sidebarOpen={sidebarOpen} onToggle={onToggleSidebar} />
          <h2><span className={`status-dot ${browserStatus}`} />Browser: {browserStatus}</h2>
        </div>
        <div className="browser-actions">
          {isRunning ? (
            <button className="btn-danger" onClick={onStop} disabled={loading}>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><rect x="6" y="6" width="12" height="12" rx="2"/></svg>
              Stop Browser
            </button>
          ) : (
            <button className="btn-primary" onClick={onStart} disabled={loading}>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"/></svg>
              {loading ? 'Starting...' : 'Start Browser'}
            </button>
          )}
        </div>
      </div>

      <div className="chat-messages">
        {messages.length === 0 ? (
          <div className="chat-empty">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--text-secondary)" strokeWidth="1.5"><circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"/><path d="M2 12h20"/></svg>
            <h3>Welcome to Stargrazer</h3>
            <p>Start the browser, connect your social accounts in Settings, then upload content below.</p>
          </div>
        ) : messages.map(msg => (
          <div key={msg.id} className={`message ${msg.type}`}>{msg.text}</div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      <div className="chat-input-area">
        {/* Platform checkboxes */}
        <div className="upload-platforms">
          {platforms.map(p => {
            const icon = PLATFORM_ICONS[p.id];
            const colors = PLATFORM_COLORS[p.id];
            return (
              <label key={p.id} className={`upload-platform-chip ${selectedPlatforms.has(p.id) ? 'selected' : ''} ${p.loggedIn ? '' : 'disabled'}`}
                style={{ '--chip-bg': colors?.bg } as React.CSSProperties}>
                <input type="checkbox" checked={selectedPlatforms.has(p.id)} onChange={() => p.loggedIn && togglePlatform(p.id)} disabled={!p.loggedIn} />
                <span className="upload-platform-icon">{icon}</span>
                <span>{p.name}</span>
                {!p.loggedIn && <span className="upload-platform-lock">Not connected</span>}
              </label>
            );
          })}
        </div>

        {/* Upload form */}
        <div className="upload-form">
          <div className="upload-file-row">
            <button className="btn-secondary upload-file-btn" onClick={handleSelectFile}>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M21.44 11.05l-9.19 9.19a6 6 0 01-8.49-8.49l9.19-9.19a4 4 0 015.66 5.66l-9.2 9.19a2 2 0 01-2.83-2.83l8.49-8.48"/></svg>
              {selectedFileName || 'Attach file'}
            </button>
            {selectedFile && (
              <button className="upload-file-clear" onClick={() => { setSelectedFile(''); setSelectedFileName(''); }} title="Remove file">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            )}
            {/* Hashtag bubble input */}
            <div className="tag-input-wrapper">
              {tags.map(tag => (
                <span key={tag} className="tag-bubble">
                  {tag}
                  <button className="tag-remove" onClick={() => removeTag(tag)}>
                    <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                  </button>
                </span>
              ))}
              <input
                className="tag-input"
                type="text"
                placeholder={tags.length === 0 ? '#hashtags...' : ''}
                value={tagInput}
                onChange={e => setTagInput(e.target.value)}
                onKeyDown={handleTagKeyDown}
              />
            </div>
          </div>
          <div className="upload-caption-row">
            <textarea className="upload-caption" placeholder="Write your caption..." rows={2} value={caption} onChange={e => setCaption(e.target.value)} />
            <button className="btn-primary upload-send" onClick={handleSend} disabled={uploading || !isRunning || selectedPlatforms.size === 0}>
              {uploading ? 'Uploading...' : 'Send'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

/* --- Social Media Section --- */
function SocialMediaSection({ onBrowserStatusChange, addMessage, refreshPlatforms }: {
  onBrowserStatusChange: (s: string) => void;
  addMessage: (t: ChatMessage['type'], m: string) => void;
  refreshPlatforms: () => void;
}) {
  const [platforms, setPlatforms] = useState<main.PlatformResponse[]>([]);
  const [loadingPlatform, setLoadingPlatform] = useState<string | null>(null);
  const [cookieModal, setCookieModal] = useState<string | null>(null);
  const [infoModal, setInfoModal] = useState<main.PlatformResponse | null>(null);

  useEffect(() => { GetPlatforms().then(setPlatforms); }, []);

  const handleConnect = async (platformId: string) => {
    setCookieModal(platformId);
    // Also open the platform in the browser so user can log in and copy cookies
    try {
      const res = await OpenPlatform(platformId);
      if (res.status === 'running') {
        onBrowserStatusChange('running');
      } else if (res.error) {
        addMessage('error', `Browser: ${res.error}`);
      }
    } catch (err: any) {
      addMessage('error', `Failed to open browser: ${err?.message || err}`);
    }
  };

  const handleOpenInBrowser = async (platformId: string) => {
    setLoadingPlatform(platformId);
    try {
      const res = await OpenPlatform(platformId);
      if (res.status === 'running') onBrowserStatusChange('running');
      else addMessage('error', `Failed: ${res.error}`);
    } catch (err: any) { addMessage('error', `Error: ${err?.message || err}`); }
    setLoadingPlatform(null);
  };

  const handleImportCookies = async (platformId: string, cookieText: string) => {
    const name = platforms.find(p => p.id === platformId)?.name || platformId;
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
    } catch (err: any) { addMessage('error', `Import failed: ${err?.message || err}`); }
    setCookieModal(null);
  };

  const handlePurgeSession = async (platformId: string) => {
    const name = platforms.find(p => p.id === platformId)?.name || platformId;
    const status = await PurgeSession(platformId);
    setPlatforms(prev => prev.map(p => p.id === platformId ? status : p));
    refreshPlatforms();
    setInfoModal(null);
    addMessage('info', `${name} session purged. You can reconnect.`);
  };

  const handleRefreshAll = async () => { try { const all = await CheckAllLoginStatus(); setPlatforms(all); refreshPlatforms(); } catch {} };

  return (
    <div className="config-section social-section">
      <div className="social-header">
        <h3>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
          Social Media Connections
        </h3>
        <button className="btn-icon" onClick={handleRefreshAll} title="Refresh all"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg></button>
      </div>
      <div className="social-grid">
        {platforms.map(p => {
          const colors = PLATFORM_COLORS[p.id] || { bg: '#333', hover: '#444', text: '#fff' };
          return (
            <div key={p.id} className={`social-card ${p.loggedIn ? 'logged-in' : ''}`}
              onClick={() => !loadingPlatform && (p.loggedIn ? handleOpenInBrowser(p.id) : handleConnect(p.id))}
              onKeyDown={e => e.key === 'Enter' && !loadingPlatform && (p.loggedIn ? handleOpenInBrowser(p.id) : handleConnect(p.id))}
              role="button" tabIndex={0}
              style={{ '--platform-bg': colors.bg, '--platform-hover': colors.hover, '--platform-text': colors.text } as React.CSSProperties}>
              <div className="social-card-icon">{PLATFORM_ICONS[p.id]}</div>
              <div className="social-card-info">
                <span className="social-card-name">{p.name}</span>
                {p.loggedIn
                  ? <span className="social-card-status connected"><span className="status-dot running"/>{p.username || 'Connected'}</span>
                  : <span className="social-card-status disconnected">Click to connect</span>}
              </div>
              <div className="social-card-actions">
                <button className="social-info-btn" onClick={e => { e.stopPropagation(); setInfoModal(p); }} title="Session info">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
                </button>
              </div>
              {p.loggedIn && p.lastLogin && <div className="social-card-meta">Since {new Date(p.lastLogin).toLocaleDateString()}</div>}
            </div>
          );
        })}
      </div>

      {/* Cookie Paste Modal */}
      {cookieModal && <CookiePasteModal
        platform={platforms.find(p => p.id === cookieModal)!}
        onImport={(text) => handleImportCookies(cookieModal, text)}
        onCancel={() => setCookieModal(null)}
      />}

      {/* Info Modal */}
      {infoModal && (
        <div className="modal-overlay" onClick={() => setInfoModal(null)} onKeyDown={e => e.key === 'Escape' && setInfoModal(null)} role="presentation">
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <div className="modal-title-row">
                <div className="social-card-icon" style={{ '--platform-bg': PLATFORM_COLORS[infoModal.id]?.bg || '#333', '--platform-text': '#fff' } as React.CSSProperties}>{PLATFORM_ICONS[infoModal.id]}</div>
                <h3>{infoModal.name}</h3>
              </div>
              <button className="modal-close" onClick={() => setInfoModal(null)}><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button>
            </div>
            <div className="modal-body">
              <div className="modal-field"><span className="modal-label">Status</span><span className={`modal-value ${infoModal.loggedIn ? 'text-success' : 'text-muted'}`}><span className={`status-dot ${infoModal.loggedIn ? 'running' : 'stopped'}`}/>{infoModal.loggedIn ? 'Connected' : 'Not connected'}</span></div>
              {infoModal.username && <div className="modal-field"><span className="modal-label">User / ID</span><span className="modal-value">{infoModal.username}</span></div>}
              <div className="modal-field"><span className="modal-label">URL</span><span className="modal-value modal-url">{infoModal.url}</span></div>
              <div className="modal-field"><span className="modal-label">Session Directory</span><span className="modal-value modal-path">{infoModal.sessionDir}</span></div>
              {infoModal.lastLogin && <div className="modal-field"><span className="modal-label">Logged In</span><span className="modal-value">{new Date(infoModal.lastLogin).toLocaleString()}</span></div>}
              {infoModal.lastCheck && <div className="modal-field"><span className="modal-label">Last Verified</span><span className="modal-value">{new Date(infoModal.lastCheck).toLocaleString()}</span></div>}
            </div>
            <div className="modal-footer">
              <button className="btn-primary" onClick={() => { handleOpenInBrowser(infoModal.id); setInfoModal(null); }}>{infoModal.loggedIn ? 'Open' : 'Connect'}</button>
              {infoModal.loggedIn && (
                <button className="btn-danger" onClick={() => handlePurgeSession(infoModal.id)}>
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
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

/* --- Cookie Paste Modal --- */
function CookiePasteModal({ platform, onImport, onCancel }: {
  platform: main.PlatformResponse;
  onImport: (text: string) => void;
  onCancel: () => void;
}) {
  const [cookieText, setCookieText] = useState('');
  const colors = PLATFORM_COLORS[platform.id];
  const lineCount = cookieText.split('\n').filter(l => l.trim() && !l.startsWith('#')).length;

  return (
    <div className="modal-overlay" onClick={onCancel} onKeyDown={e => e.key === 'Escape' && onCancel()} role="presentation">
      <div className="modal-content cookie-modal" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <div className="modal-title-row">
            <div className="social-card-icon" style={{ '--platform-bg': colors?.bg || '#333', '--platform-text': '#fff' } as React.CSSProperties}>
              {PLATFORM_ICONS[platform.id]}
            </div>
            <div>
              <h3>Import {platform.name} Cookies</h3>
              <span style={{ fontSize: 12, color: 'var(--text-secondary)' }}>Paste Netscape cookie.txt format from the cookies extension</span>
            </div>
          </div>
          <button className="modal-close" onClick={onCancel}><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button>
        </div>
        <div className="modal-body" style={{ padding: '16px 20px' }}>
          <div className="cookie-instructions">
            <div className="cookie-step"><span className="cookie-step-num">1</span> Open {platform.name} in the browser and log in</div>
            <div className="cookie-step"><span className="cookie-step-num">2</span> Click the cookies extension icon (pinned in toolbar)</div>
            <div className="cookie-step"><span className="cookie-step-num">3</span> Export as Netscape format and paste below</div>
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
          <button className="btn-primary" onClick={() => onImport(cookieText)} disabled={lineCount === 0}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
            Import Cookies
          </button>
          <button className="btn-secondary" onClick={onCancel}>Cancel</button>
        </div>
      </div>
    </div>
  );
}

/* --- Logs Modal --- */
function LogsModal({ onClose }: { onClose: () => void }) {
  const [logs, setLogs] = useState<main.LogEntryResponse[]>([]);
  const [filter, setFilter] = useState('');
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => { GetLogs().then(setLogs); const i = setInterval(() => GetLogs().then(setLogs), 2000); return () => clearInterval(i); }, []);
  useEffect(() => { bottomRef.current?.scrollIntoView(); }, [logs]);

  const filtered = filter ? logs.filter(l => l.level === filter || l.source.includes(filter) || l.message.toLowerCase().includes(filter.toLowerCase())) : logs;

  const handleExport = async () => {
    const json = await ExportLogs();
    const blob = new Blob([json], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url; a.download = `stargrazer-logs-${new Date().toISOString().slice(0,10)}.json`;
    a.click(); URL.revokeObjectURL(url);
  };

  return (
    <div className="modal-overlay" onClick={onClose} onKeyDown={e => e.key === 'Escape' && onClose()} role="presentation">
      <div className="modal-content logs-modal" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h3>Application Logs</h3>
          <div style={{ display: 'flex', gap: 8 }}>
            <button className="btn-secondary" style={{ padding: '6px 12px', fontSize: 12 }} onClick={handleExport}>Export JSON</button>
            <button className="btn-secondary" style={{ padding: '6px 12px', fontSize: 12 }} onClick={async () => { await ClearLogs(); setLogs([]); }}>Clear</button>
            <button className="modal-close" onClick={onClose}><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button>
          </div>
        </div>
        <div style={{ padding: '0 20px 8px' }}>
          <input className="config-search-input" placeholder="Filter logs..." value={filter} onChange={e => setFilter(e.target.value)} style={{ width: '100%', fontSize: 12 }} />
        </div>
        <div className="logs-body">
          {filtered.map((l, i) => (
            <div key={`${l.timestamp}-${l.source}-${i}`} className={`log-entry log-${l.level}`}>
              <span className="log-time">{new Date(l.timestamp).toLocaleTimeString()}</span>
              <span className={`log-level log-level-${l.level}`}>{l.level.toUpperCase()}</span>
              <span className="log-source">[{l.source}]</span>
              <span className="log-msg">{l.message}</span>
            </div>
          ))}
          <div ref={bottomRef} />
        </div>
      </div>
    </div>
  );
}

function statusDotClass(status: string): string {
  if (status === 'active') return 'running';
  if (status === 'paused') return 'stopped';
  return 'error';
}

function statusTextClass(status: string): string {
  if (status === 'active') return 'text-success';
  if (status === 'paused') return 'text-muted';
  return 'text-error';
}

/* --- Schedules Panel --- */
function SchedulesPanel({ sidebarOpen, onToggleSidebar, addMessage, platforms }: {
  sidebarOpen: boolean; onToggleSidebar: () => void;
  addMessage: (t: ChatMessage['type'], m: string) => void;
  platforms: main.PlatformResponse[];
}) {
  const [schedules, setSchedules] = useState<main.ScheduleResponse[]>([]);
  const [showCreate, setShowCreate] = useState(false);
  const [selectedJob, setSelectedJob] = useState<main.ScheduleResponse | null>(null);

  const refresh = () => GetSchedules().then(setSchedules);
  useEffect(() => { refresh(); }, []);

  const handleDelete = async (id: string) => {
    await DeleteSchedule(id);
    addMessage('info', 'Schedule deleted.');
    refresh();
    setSelectedJob(null);
  };

  const handlePauseResume = async (j: main.ScheduleResponse) => {
    const res = j.status === 'active' ? await PauseSchedule(j.id) : await ResumeSchedule(j.id);
    addMessage('info', `Schedule ${res.status === 'active' ? 'resumed' : 'paused'}.`);
    refresh();
    setSelectedJob(res.id ? res : null);
  };

  return (
    <div className="config-panel">
      <div className="config-header">
        <HamburgerBtn sidebarOpen={sidebarOpen} onToggle={onToggleSidebar} />
        <h2>Schedules</h2>
        <button className="btn-primary" style={{ marginLeft: 'auto', padding: '8px 16px', fontSize: 13 }} onClick={() => setShowCreate(true)}>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          Create Schedule
        </button>
      </div>

      {schedules.length === 0 ? (
        <div className="chat-empty" style={{ paddingTop: 80 }}>
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--text-secondary)" strokeWidth="1.5"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
          <h3>No Scheduled Jobs</h3>
          <p>Create a schedule to automate uploads or keep sessions alive. Keep-alive jobs are created automatically when you connect a platform.</p>
        </div>
      ) : (
        <div className="schedule-list">
          {schedules.map(j => (
            <div key={j.id} className="schedule-card" onClick={() => setSelectedJob(j)} onKeyDown={e => e.key === 'Enter' && setSelectedJob(j)} role="button" tabIndex={0}>
              <span className={`status-dot ${statusDotClass(j.status)}`} />
              <div className="schedule-card-info">
                <span className="schedule-card-name">
                  {j.name}
                  {j.auto && <span className="schedule-badge auto">auto</span>}
                  <span className={`schedule-badge ${j.type === 'session_keepalive' ? 'keepalive' : 'upload'}`}>
                    {j.type === 'session_keepalive' ? 'keep-alive' : 'upload'}
                  </span>
                </span>
                <span className="schedule-card-meta">
                  {j.platforms.join(', ')} &middot; {j.cronExpr}
                  {j.nextRun && <> &middot; Next: {new Date(j.nextRun).toLocaleString()}</>}
                </span>
              </div>
              <div className="schedule-card-stats">
                <span>{j.runCount} runs</span>
                {j.lastResult && <span className={j.lastResult === 'success' ? 'text-success' : 'text-error'}>{j.lastResult}</span>}
              </div>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--text-secondary)" strokeWidth="2"><polyline points="9 18 15 12 9 6"/></svg>
            </div>
          ))}
        </div>
      )}

      {showCreate && <CreateScheduleModal platforms={platforms} onCreated={() => { refresh(); setShowCreate(false); addMessage('success', 'Schedule created.'); }} onCancel={() => setShowCreate(false)} />}
      {selectedJob && <ScheduleDetailModal job={selectedJob} onClose={() => { setSelectedJob(null); refresh(); }} onDelete={handleDelete} onPauseResume={handlePauseResume} />}
    </div>
  );
}

/* --- Create Schedule Modal --- */
function CreateScheduleModal({ platforms, onCreated, onCancel }: {
  platforms: main.PlatformResponse[];
  onCreated: () => void;
  onCancel: () => void;
}) {
  const [name, setName] = useState('');
  const [type, setType] = useState<'session_keepalive' | 'upload'>('upload');
  const [cronExpr, setCronExpr] = useState('0 */12 * * *');
  const [interval, setScheduleInterval] = useState('12h');
  const [selectedPlatforms, setSelectedPlatforms] = useState<Set<string>>(new Set());
  const [caption, setCaption] = useState('');
  const [hashtags, setHashtags] = useState('');
  const [saving, setSaving] = useState(false);

  const intervalMap: Record<string, string> = { '6h': '0 */6 * * *', '12h': '0 */12 * * *', '24h': '0 0 * * *', '3d': '0 0 */3 * *', 'custom': '' };
  const handleIntervalChange = (v: string) => { setScheduleInterval(v); if (v !== 'custom') setCronExpr(intervalMap[v]); };

  const togglePlatform = (id: string) => setSelectedPlatforms(prev => { const n = new Set(prev); n.has(id)?n.delete(id):n.add(id); return n; });

  const handleCreate = async () => {
    if (!name.trim() || selectedPlatforms.size === 0 || !cronExpr.trim()) return;
    setSaving(true);
    const tags = hashtags.split(/[\s,]+/).map(t => t.startsWith('#') ? t : `#${t}`).filter(t => t.length > 1);
    await CreateSchedule({
      name, type, platforms: [...selectedPlatforms], cronExpr,
      ...(type === 'upload' ? { caption, hashtags: tags } : {}),
    } as any);
    setSaving(false);
    onCreated();
  };

  return (
    <div className="modal-overlay" onClick={onCancel} onKeyDown={e => e.key === 'Escape' && onCancel()} role="presentation">
      <div className="modal-content" style={{ width: 500 }} onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h3>Create Schedule</h3>
          <button className="modal-close" onClick={onCancel}><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button>
        </div>
        <div className="modal-body">
          <div className="config-field" style={{marginBottom:12}}>
            <label htmlFor="sched-job-name">Job Name</label>
            <input id="sched-job-name" type="text" value={name} onChange={e => setName(e.target.value)} placeholder="My upload schedule" />
          </div>
          <div className="config-field" style={{marginBottom:12}}>
            <span className="config-field-label">Type</span>
            <div className="theme-switcher" style={{width:'100%'}}>
              <button className={type==='session_keepalive'?'active':''} onClick={()=>setType('session_keepalive')} style={{flex:1,justifyContent:'center'}}>Keep Alive</button>
              <button className={type==='upload'?'active':''} onClick={()=>setType('upload')} style={{flex:1,justifyContent:'center'}}>Upload</button>
            </div>
          </div>
          <div className="config-field" style={{marginBottom:12}}>
            <span className="config-field-label">Interval</span>
            <div style={{display:'flex',gap:6,flexWrap:'wrap'}}>
              {['6h','12h','24h','3d','custom'].map(v => (
                <button key={v} className={`btn-secondary ${interval===v?'active':''}`} style={{padding:'6px 12px',fontSize:12,...(interval===v?{background:'var(--accent)',color:'#fff',borderColor:'var(--accent)'}:{})}} onClick={()=>handleIntervalChange(v)}>
                  {v === 'custom' ? 'Custom' : `Every ${v}`}
                </button>
              ))}
            </div>
            {interval === 'custom' && <input type="text" value={cronExpr} onChange={e=>setCronExpr(e.target.value)} placeholder="0 */12 * * *" style={{marginTop:6}} aria-label="Custom cron expression" />}
          </div>
          <div className="config-field" style={{marginBottom:12}}>
            <span className="config-field-label">Platforms</span>
            <div className="upload-platforms">
              {platforms.map(p => (
                <label key={p.id} className={`upload-platform-chip ${selectedPlatforms.has(p.id)?'selected':''}`}>
                  <input type="checkbox" checked={selectedPlatforms.has(p.id)} onChange={()=>togglePlatform(p.id)} />
                  <span className="upload-platform-icon">{PLATFORM_ICONS[p.id]}</span>
                  <span>{p.name}</span>
                </label>
              ))}
            </div>
          </div>
          {type === 'upload' && <>
            <div className="config-field" style={{marginBottom:12}}>
              <label htmlFor="sched-caption">Caption</label>
              <textarea id="sched-caption" className="upload-caption" value={caption} onChange={e=>setCaption(e.target.value)} placeholder="Post caption..." rows={2} />
            </div>
            <div className="config-field" style={{marginBottom:0}}>
              <label htmlFor="sched-hashtags">Hashtags</label>
              <input id="sched-hashtags" type="text" value={hashtags} onChange={e=>setHashtags(e.target.value)} placeholder="#hashtag1 #hashtag2" />
            </div>
          </>}
        </div>
        <div className="modal-footer">
          <button className="btn-primary" onClick={handleCreate} disabled={saving || !name.trim() || selectedPlatforms.size===0}>{saving?'Creating...':'Create'}</button>
          <button className="btn-secondary" onClick={onCancel}>Cancel</button>
        </div>
      </div>
    </div>
  );
}

/* --- Schedule Detail Modal --- */
function ScheduleDetailModal({ job, onClose, onDelete, onPauseResume }: {
  job: main.ScheduleResponse;
  onClose: () => void;
  onDelete: (id: string) => void;
  onPauseResume: (j: main.ScheduleResponse) => void;
}) {
  return (
    <div className="modal-overlay" onClick={onClose} onKeyDown={e => e.key === 'Escape' && onClose()} role="presentation">
      <div className="modal-content" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h3>{job.name}</h3>
          <button className="modal-close" onClick={onClose}><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button>
        </div>
        <div className="modal-body">
          <div className="modal-field"><span className="modal-label">Status</span><span className={`modal-value ${statusTextClass(job.status)}`}><span className={`status-dot ${statusDotClass(job.status)}`}/>{job.status}</span></div>
          <div className="modal-field"><span className="modal-label">Type</span><span className="modal-value">{job.type === 'session_keepalive' ? 'Session Keep-Alive' : 'Content Upload'}{job.auto && ' (auto-created)'}</span></div>
          <div className="modal-field"><span className="modal-label">Platforms</span><span className="modal-value">{job.platforms.join(', ')}</span></div>
          <div className="modal-field"><span className="modal-label">Schedule</span><span className="modal-value modal-path">{job.cronExpr}</span></div>
          <div className="modal-field"><span className="modal-label">Runs</span><span className="modal-value">{job.runCount}</span></div>
          {job.lastRun && <div className="modal-field"><span className="modal-label">Last Run</span><span className="modal-value">{new Date(job.lastRun).toLocaleString()}</span></div>}
          {job.lastResult && <div className="modal-field"><span className="modal-label">Last Result</span><span className={`modal-value ${job.lastResult==='success'?'text-success':'text-error'}`}>{job.lastResult}</span></div>}
          {job.nextRun && <div className="modal-field"><span className="modal-label">Next Run</span><span className="modal-value">{new Date(job.nextRun).toLocaleString()}</span></div>}
          <div className="modal-field"><span className="modal-label">Created</span><span className="modal-value">{new Date(job.createdAt).toLocaleString()}</span></div>
          {job.caption && <div className="modal-field"><span className="modal-label">Caption</span><span className="modal-value">{job.caption}</span></div>}
          {job.hashtags && job.hashtags.length > 0 && <div className="modal-field"><span className="modal-label">Hashtags</span><span className="modal-value">{job.hashtags.join(' ')}</span></div>}
        </div>
        <div className="modal-footer">
          <button className={job.status==='active'?'btn-secondary':'btn-primary'} onClick={()=>onPauseResume(job)}>{job.status==='active'?'Pause':'Resume'}</button>
          <button className="btn-danger" onClick={()=>onDelete(job.id)}>Delete</button>
          <button className="btn-secondary" onClick={onClose}>Close</button>
        </div>
      </div>
    </div>
  );
}

/* --- Config Panel --- */
function ConfigPanel({ onSaved, sidebarOpen, onToggleSidebar, onBrowserStatusChange, addMessage, refreshPlatforms }: {
  onSaved: (msg: string) => void; sidebarOpen: boolean; onToggleSidebar: () => void;
  onBrowserStatusChange: (s: string) => void;
  addMessage: (t: ChatMessage['type'], m: string) => void;
  refreshPlatforms: () => void;
}) {
  const [config, setConfig] = useState<main.BrowserConfigResponse | null>(null);
  const [saving, setSaving] = useState(false);
  const [resetting, setResetting] = useState(false);
  const [expandedCategories, setExpandedCategories] = useState<Set<string>>(new Set());
  const [search, setSearch] = useState('');
  const [showLogs, setShowLogs] = useState(false);

  useEffect(() => { GetBrowserConfig().then(setConfig); }, []);

  const updateField = <K extends keyof main.BrowserConfigResponse>(key: K, value: main.BrowserConfigResponse[K]) => { if (config) setConfig({ ...config, [key]: value }); };
  const toggleFlag = (flag: string) => { if (!config) return; const c = config.extraFlags||[]; updateField('extraFlags', c.includes(flag) ? c.filter(f=>f!==flag) : [...c, flag]); };
  const toggleCategory = (cat: string) => setExpandedCategories(prev => { const n = new Set(prev); n.has(cat)?n.delete(cat):n.add(cat); return n; });

  const handleSave = async () => {
    if (!config) return;
    setSaving(true);
    try {
      const updated = await UpdateBrowserConfig(config);
      setConfig(updated);
      // Restart browser if it was running
      const bStatus = await GetBrowserStatus();
      if (bStatus.status === 'running') {
        addMessage('system', 'Restarting browser with new settings...');
        const res = await RestartBrowser();
        onBrowserStatusChange(res.status);
        addMessage(res.status === 'running' ? 'success' : 'error', res.status === 'running' ? 'Browser restarted with new settings.' : `Restart failed: ${res.error}`);
      } else {
        onSaved('Configuration saved.');
      }
    } catch (err: any) { onSaved(`Error: ${err?.message || err}`); }
    setSaving(false);
  };

  const handleReset = async () => {
    setResetting(true);
    try { const d = await ResetBrowserConfig(); setConfig(d); onSaved('Settings reset to defaults.'); }
    catch (err: any) { onSaved(`Error: ${err?.message || err}`); }
    setResetting(false);
  };

  if (!config) return <div className="config-panel"><p>Loading...</p></div>;
  const activeFlags = config.extraFlags||[]; const activeCount = activeFlags.length;
  const q = search.toLowerCase().trim();
  const match = (...terms: string[]) => !q || terms.some(t => t.toLowerCase().includes(q));
  const showSocial = match('social','media','connection','facebook','instagram','tiktok','youtube','linkedin','login','account');
  const showConnection = match('connection','cdp','port','chromium','path','user data','directory');
  const showDisplay = match('display','headless','window','width','height');
  const showCustomFlags = match('custom','flags','extra','additional');
  const filteredFlagGroups = CHROMIUM_FLAGS.map(g => { if (!q) return {...g,filteredFlags:g.flags}; const cm=g.category.toLowerCase().includes(q); return {...g,filteredFlags:cm?g.flags:g.flags.filter(f=>f.label.toLowerCase().includes(q)||f.description.toLowerCase().includes(q)||f.flag.toLowerCase().includes(q))}; }).filter(g=>g.filteredFlags.length>0);
  const showFlags = filteredFlagGroups.length>0 || match('chromium','flags');
  const noResults = q && !showSocial && !showConnection && !showDisplay && !showFlags && !showCustomFlags;

  return (
    <div className="config-panel">
      <div className="config-header">
        <HamburgerBtn sidebarOpen={sidebarOpen} onToggle={onToggleSidebar} />
        <h2>Settings</h2>
        <button className="btn-icon" style={{ marginLeft: 'auto' }} onClick={() => setShowLogs(true)} title="View logs">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
        </button>
      </div>

      <div className="config-search">
        <svg className="config-search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
        <input type="text" className="config-search-input" placeholder="Search settings..." value={search} onChange={e => setSearch(e.target.value)} />
        {search && <button className="config-search-clear" onClick={() => setSearch('')}><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button>}
      </div>

      {noResults && <div className="config-no-results">No settings match "{search}"</div>}
      {showSocial && <SocialMediaSection onBrowserStatusChange={onBrowserStatusChange} addMessage={addMessage} refreshPlatforms={refreshPlatforms} />}

      {showConnection && <div className="config-section"><h3>Connection</h3>
        <div className="config-field"><label htmlFor="cfg-cdp-port">CDP Port</label><input id="cfg-cdp-port" type="number" value={config.cdpPort} onChange={e=>updateField('cdpPort',Number.parseInt(e.target.value)||9222)}/></div>
        <div className="config-field"><label htmlFor="cfg-chromium-path">Chromium Path</label><input id="cfg-chromium-path" type="text" value={config.chromiumPath} onChange={e=>updateField('chromiumPath',e.target.value)} placeholder="Auto-detect"/><span className="config-hint">Auto-detected from bundled assets.</span></div>
        <div className="config-field"><label htmlFor="cfg-user-data-dir">User Data Directory</label><input id="cfg-user-data-dir" type="text" value={config.userDataDir} onChange={e=>updateField('userDataDir',e.target.value)} placeholder="Default"/></div>
      </div>}

      {showDisplay && <div className="config-section"><h3>Display</h3>
        <div className="config-field"><span className="config-field-label">Headless Mode</span><div className="config-field-row"><button className={`toggle ${config.headless?'active':''}`} onClick={()=>updateField('headless',!config.headless)}/><span>{config.headless?'Enabled':'Disabled'}</span></div></div>
        <div className="config-field-inline"><div className="config-field"><label htmlFor="cfg-window-width">Width</label><input id="cfg-window-width" type="number" value={config.windowWidth} onChange={e=>updateField('windowWidth',Number.parseInt(e.target.value)||1280)}/></div><div className="config-field"><label htmlFor="cfg-window-height">Height</label><input id="cfg-window-height" type="number" value={config.windowHeight} onChange={e=>updateField('windowHeight',Number.parseInt(e.target.value)||900)}/></div></div>
      </div>}

      {showFlags && <div className="config-section"><h3>Chromium Flags {activeCount>0&&<span className="flag-badge">{activeCount} active</span>}</h3>
        {filteredFlagGroups.map(group => {
          const exp = q?true:expandedCategories.has(group.category);
          const ga = group.filteredFlags.filter(f=>activeFlags.includes(f.flag)).length;
          return <div key={group.category} className="flag-group">
            <button className="flag-group-header" onClick={()=>toggleCategory(group.category)}>
              <svg className={`flag-chevron ${exp?'expanded':''}`} width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="9 18 15 12 9 6"/></svg>
              <span>{group.category}</span>{ga>0&&<span className="flag-badge">{ga}</span>}
            </button>
            {exp && <div className="flag-list">{group.filteredFlags.map(({flag,label,description})=>
              <label key={flag} className="flag-item" title={description}><input type="checkbox" checked={activeFlags.includes(flag)} onChange={()=>toggleFlag(flag)}/><div className="flag-item-content"><span className="flag-label">{label}</span><span className="flag-desc">{description}</span></div><code className="flag-code">{flag}</code></label>
            )}</div>}
          </div>;
        })}
      </div>}

      {showCustomFlags && <div className="config-section"><h3>Custom Flags</h3>
        <div className="config-field"><label htmlFor="cfg-custom-flags">Additional flags (comma-separated)</label><input id="cfg-custom-flags" type="text"
          value={activeFlags.filter(f=>!CHROMIUM_FLAGS.some(g=>g.flags.some(gf=>gf.flag===f))).join(', ')}
          onChange={e=>{const known=activeFlags.filter(f=>CHROMIUM_FLAGS.some(g=>g.flags.some(gf=>gf.flag===f)));updateField('extraFlags',[...known,...e.target.value.split(',').map(s=>s.trim()).filter(Boolean)]);}}
          placeholder="--proxy-server=host:port"/></div>
      </div>}

      <div className="config-actions">
        <button className="btn-primary" onClick={handleSave} disabled={saving}>{saving?'Saving...':'Save Settings'}</button>
        <button className="btn-secondary" onClick={handleReset} disabled={resetting}>{resetting?'Resetting...':'Reset to Defaults'}</button>
      </div>

      {showLogs && <LogsModal onClose={() => setShowLogs(false)} />}
    </div>
  );
}

export default App;

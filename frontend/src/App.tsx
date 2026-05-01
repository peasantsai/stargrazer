import { useState, useEffect, useRef, useCallback } from 'react';
import { StartBrowser, StopBrowser, GetBrowserStatus, GetPlatforms } from '../wailsjs/go/main/App';
import type { ChatMessage, View, PlatformResponse } from './types';
import { isPlatformView, platformIdFromView } from './types';
import { useTheme } from './hooks/useTheme';
import { useAccount } from './hooks/useAccount';
import { Sidebar } from './components/Sidebar';
import { ChatPanel } from './components/ChatPanel';
import { ConfigPanel } from './components/ConfigPanel';
import { PlatformPage } from './components/PlatformPage';
import { RunStrip } from './components/RunStrip';
import { useRunEvents } from './hooks/useRunEvents';

const PLATFORM_NAMES: Record<string, string> = {
  facebook: 'Facebook',
  instagram: 'Instagram',
  tiktok: 'TikTok',
  youtube: 'YouTube',
  linkedin: 'LinkedIn',
  x: 'X',
};

function App() {
  const [view, setView] = useState<View>('chat');
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [browserStatus, setBrowserStatus] = useState('stopped');
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [loading, setLoading] = useState(false);
  const [platforms, setPlatforms] = useState<PlatformResponse[]>([]);
  const [theme, setTheme] = useTheme();
  const [account, updateAccount] = useAccount();
  const runEvents = useRunEvents();
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const msgIdRef = useRef(0);

  const addMessage = useCallback((type: ChatMessage['type'], text: string) => {
    msgIdRef.current += 1;
    setMessages(prev => [...prev, { id: msgIdRef.current, type, text }]);
  }, []);

  useEffect(() => { messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' }); }, [messages]);
  useEffect(() => { GetBrowserStatus().then(r => setBrowserStatus(r.status)); }, []);
  // Poll browser status so the toggle button stays in sync (e.g. browser opened via platform page).
  useEffect(() => {
    const interval = setInterval(() => {
      GetBrowserStatus().then(r => setBrowserStatus(r.status));
    }, 3000);
    return () => clearInterval(interval);
  }, []);
  useEffect(() => { GetPlatforms().then(setPlatforms); }, []);
  useEffect(() => { if (view === 'chat') GetPlatforms().then(setPlatforms); }, [view]);

  const refreshPlatforms = useCallback(() => { GetPlatforms().then(setPlatforms); }, []);

  const handleStartBrowser = async () => {
    setLoading(true);
    const res = await StartBrowser();
    setBrowserStatus(res.status);
    setLoading(false);
  };

  const handleStopBrowser = async () => {
    setLoading(true);
    try {
      const res = await StopBrowser();
      setBrowserStatus(res.status);
    } catch {
      // logged on backend
    } finally {
      setLoading(false);
    }
  };

  const renderPanel = () => {
    if (view === 'chat') {
      return (
        <ChatPanel
          messages={messages}
          messagesEndRef={messagesEndRef}
          sidebarOpen={sidebarOpen}
          onToggleSidebar={() => setSidebarOpen(true)}
          platforms={platforms}
          addMessage={addMessage}
        />
      );
    }
    if (view === 'config') {
      return (
        <ConfigPanel
          onSaved={msg => addMessage('success', msg)}
          sidebarOpen={sidebarOpen}
          onToggleSidebar={() => setSidebarOpen(true)}
          onBrowserStatusChange={setBrowserStatus}
          addMessage={addMessage}
          browserStatus={browserStatus}
          theme={theme}
          setTheme={setTheme}
          platforms={platforms}
        />
      );
    }
    if (isPlatformView(view)) {
      const pid = platformIdFromView(view);
      return (
        <PlatformPage
          platformId={pid}
          platformName={PLATFORM_NAMES[pid] ?? pid}
          sidebarOpen={sidebarOpen}
          onToggleSidebar={() => setSidebarOpen(true)}
          addMessage={addMessage}
          onBrowserStatusChange={setBrowserStatus}
          refreshPlatforms={refreshPlatforms}
        />
      );
    }
    return null;
  };

  return (
    <div className="app-layout">
      <Sidebar
        view={view}
        setView={setView}
        browserStatus={browserStatus}
        browserLoading={loading}
        onStartBrowser={handleStartBrowser}
        onStopBrowser={handleStopBrowser}
        open={sidebarOpen}
        onToggle={() => setSidebarOpen(p => !p)}
        account={account}
        updateAccount={updateAccount}
      />
      <div className="main-content">
        <RunStrip current={runEvents.current} />
        {renderPanel()}
      </div>
    </div>
  );
}

export default App;

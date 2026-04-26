import { useState, useEffect, useRef, useCallback } from 'react';
import { StartBrowser, StopBrowser, GetBrowserStatus, GetPlatforms } from '../wailsjs/go/main/App';
import type { ChatMessage, View, PlatformResponse } from './types';
import { isPlatformView, platformIdFromView } from './types';
import { useTheme } from './hooks/useTheme';
import { useAccount } from './hooks/useAccount';
import { Sidebar } from './components/Sidebar';
import { ChatPanel } from './components/ChatPanel';
import { SchedulesPanel } from './components/SchedulesPanel';
import { ConfigPanel } from './components/ConfigPanel';
import { PlatformPage } from './components/PlatformPage';

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
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const msgIdRef = useRef(0);

  const addMessage = useCallback((type: ChatMessage['type'], text: string) => {
    msgIdRef.current += 1;
    setMessages(prev => [...prev, { id: msgIdRef.current, type, text }]);
  }, []);

  useEffect(() => { messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' }); }, [messages]);
  useEffect(() => { GetBrowserStatus().then(r => setBrowserStatus(r.status)); }, []);
  useEffect(() => { GetPlatforms().then(setPlatforms); }, []);
  useEffect(() => { if (view === 'chat') GetPlatforms().then(setPlatforms); }, [view]);

  const refreshPlatforms = useCallback(() => { GetPlatforms().then(setPlatforms); }, []);

  const handleStartBrowser = async () => {
    setLoading(true);
    addMessage('system', 'Starting browser...');
    const res = await StartBrowser();
    setBrowserStatus(res.status);
    addMessage(
      res.status === 'running' ? 'success' : 'error',
      res.status === 'running' ? 'Browser started. CDP active.' : `Failed: ${res.error}`,
    );
    setLoading(false);
  };

  const handleStopBrowser = async () => {
    setLoading(true);
    addMessage('system', 'Stopping browser...');
    try {
      const res = await StopBrowser();
      setBrowserStatus(res.status);
      addMessage(res.status === 'stopped' ? 'info' : 'error',
        res.status === 'stopped' ? 'Browser stopped.' : `Stop failed: ${res.error}`);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      addMessage('error', `Stop error: ${msg}`);
    } finally {
      setLoading(false);
    }
  };

  const renderPanel = () => {
    if (view === 'chat') {
      return (
        <ChatPanel
          messages={messages}
          browserStatus={browserStatus}
          loading={loading}
          onStart={handleStartBrowser}
          onStop={handleStopBrowser}
          messagesEndRef={messagesEndRef}
          sidebarOpen={sidebarOpen}
          onToggleSidebar={() => setSidebarOpen(true)}
          platforms={platforms}
          addMessage={addMessage}
        />
      );
    }
    if (view === 'schedules') {
      return (
        <SchedulesPanel
          sidebarOpen={sidebarOpen}
          onToggleSidebar={() => setSidebarOpen(true)}
          addMessage={addMessage}
          platforms={platforms}
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
          refreshPlatforms={refreshPlatforms}
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
        open={sidebarOpen}
        onToggle={() => setSidebarOpen(p => !p)}
        theme={theme}
        setTheme={setTheme}
        account={account}
        updateAccount={updateAccount}
      />
      <div className="main-content">
        {renderPanel()}
      </div>
    </div>
  );
}

export default App;

import { useState, useEffect, useRef } from 'react';
import { GetLogs, ExportLogs, ClearLogs } from '../../../wailsjs/go/main/App';
import type { LogEntryResponse } from '../../types';

interface Props {
  readonly onClose: () => void;
}

const LEVEL_FILTERS = ['all', 'info', 'warn', 'error', 'debug'] as const;

export function LogsModal({ onClose }: Props) {
  const [logs, setLogs] = useState<LogEntryResponse[]>([]);
  const [filter, setFilter] = useState('');
  const [levelFilter, setLevelFilter] = useState<string>('all');
  const [autoScroll, setAutoScroll] = useState(true);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    GetLogs().then(setLogs);
    const interval = setInterval(() => { GetLogs().then(setLogs); }, 1500);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    if (autoScroll) bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [logs, autoScroll]);

  const filtered = logs.filter(l => {
    if (levelFilter !== 'all' && l.level !== levelFilter) return false;
    if (!filter) return true;
    const q = filter.toLowerCase();
    return l.source.toLowerCase().includes(q) || l.message.toLowerCase().includes(q);
  });

  const handleExport = async () => {
    const json = await ExportLogs();
    const blob = new Blob([json], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `stargrazer-logs-${new Date().toISOString().slice(0, 10)}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const handleClear = async () => {
    await ClearLogs();
    setLogs([]);
  };

  return (
    <dialog
      className="modal-overlay"
      open
      onCancel={e => { e.preventDefault(); onClose(); }}
    >
      <button className="modal-backdrop" aria-label="Close" tabIndex={-1} onClick={onClose} />
      <div className="modal-content logs-modal">
        <div className="modal-header">
          <h3>Application Logs</h3>
          <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
            <span style={{ fontSize: 11, color: 'var(--text-secondary)' }}>{filtered.length} entries</span>
            <button className="btn-secondary" style={{ padding: '6px 12px', fontSize: 12 }} onClick={handleExport}>
              Export
            </button>
            <button className="btn-secondary" style={{ padding: '6px 12px', fontSize: 12 }} onClick={handleClear}>
              Clear
            </button>
            <button className="modal-close" onClick={onClose} aria-label="Close">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" />
              </svg>
            </button>
          </div>
        </div>
        <div className="logs-toolbar">
          <div className="logs-level-filters">
            {LEVEL_FILTERS.map(lv => (
              <button
                key={lv}
                className={`logs-level-btn ${levelFilter === lv ? 'active' : ''} ${lv !== 'all' ? `logs-level-btn-${lv}` : ''}`}
                onClick={() => setLevelFilter(lv)}
              >
                {lv.toUpperCase()}
              </button>
            ))}
          </div>
          <input
            className="config-search-input"
            placeholder="Filter by source or message..."
            value={filter}
            onChange={e => setFilter(e.target.value)}
            style={{ flex: 1, fontSize: 12, padding: '6px 12px' }}
          />
          <label style={{ display: 'flex', alignItems: 'center', gap: 4, fontSize: 11, color: 'var(--text-secondary)', cursor: 'pointer', whiteSpace: 'nowrap' }}>
            <input type="checkbox" checked={autoScroll} onChange={e => setAutoScroll(e.target.checked)} style={{ accentColor: 'var(--accent)' }} />
            Auto-scroll
          </label>
        </div>
        <div className="logs-body">
          {filtered.length === 0 ? (
            <div style={{ padding: 24, textAlign: 'center', color: 'var(--text-secondary)' }}>
              {logs.length === 0 ? 'No logs yet.' : 'No logs match the current filter.'}
            </div>
          ) : filtered.map((l, i) => (
            <div key={`${l.timestamp}-${l.source}-${i}`} className={`log-entry log-${l.level}`}>
              <span className="log-time">{new Date(l.timestamp).toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit', fractionalSecondDigits: 3 } as Intl.DateTimeFormatOptions)}</span>
              <span className={`log-level log-level-${l.level}`}>{l.level.toUpperCase().padEnd(5)}</span>
              <span className="log-source">[{l.source}]</span>
              <span className="log-msg">{l.message}</span>
            </div>
          ))}
          <div ref={bottomRef} />
        </div>
      </div>
    </dialog>
  );
}

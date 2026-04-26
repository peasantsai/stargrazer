import { useState, useEffect, useRef } from 'react';
import { GetLogs, ExportLogs, ClearLogs } from '../../../wailsjs/go/main/App';
import type { LogEntryResponse } from '../../types';

interface Props {
  readonly onClose: () => void;
}

export function LogsModal({ onClose }: Props) {
  const [logs, setLogs] = useState<LogEntryResponse[]>([]);
  const [filter, setFilter] = useState('');
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    GetLogs().then(setLogs);
    const interval = setInterval(() => { GetLogs().then(setLogs); }, 2000);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => { bottomRef.current?.scrollIntoView(); }, [logs]);

  const filtered = filter
    ? logs.filter(l =>
        l.level === filter ||
        l.source.includes(filter) ||
        l.message.toLowerCase().includes(filter.toLowerCase())
      )
    : logs;

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
      onClick={e => { if (e.target === e.currentTarget) onClose(); }}
      onKeyDown={e => { if (e.key === 'Escape') onClose(); }}
    >
      <div className="modal-content logs-modal">
        <div className="modal-header">
          <h3>Application Logs</h3>
          <div style={{ display: 'flex', gap: 8 }}>
            <button className="btn-secondary" style={{ padding: '6px 12px', fontSize: 12 }} onClick={handleExport}>
              Export JSON
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
        <div style={{ padding: '0 20px 8px' }}>
          <input
            className="config-search-input"
            placeholder="Filter logs..."
            value={filter}
            onChange={e => setFilter(e.target.value)}
            style={{ width: '100%', fontSize: 12 }}
          />
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
    </dialog>
  );
}

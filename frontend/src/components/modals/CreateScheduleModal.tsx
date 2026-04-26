import { useState } from 'react';
import { CreateSchedule } from '../../../wailsjs/go/main/App';
import type { PlatformResponse, CreateScheduleRequest } from '../../types';
import { PLATFORM_ICONS } from '../../constants/platforms';

interface Props {
  platforms: PlatformResponse[];
  onCreated: () => void;
  onCancel: () => void;
}

const INTERVAL_MAP: Record<string, string> = {
  '6h': '0 */6 * * *',
  '12h': '0 */12 * * *',
  '24h': '0 0 * * *',
  '3d': '0 0 */3 * *',
  custom: '',
};

export function CreateScheduleModal({ platforms, onCreated, onCancel }: Props) {
  const [name, setName] = useState('');
  const [type, setType] = useState<'session_keepalive' | 'upload'>('upload');
  const [cronExpr, setCronExpr] = useState('0 */12 * * *');
  const [interval, setInterval] = useState('12h');
  const [selectedPlatforms, setSelectedPlatforms] = useState<Set<string>>(new Set());
  const [caption, setCaption] = useState('');
  const [hashtags, setHashtags] = useState('');
  const [saving, setSaving] = useState(false);

  const handleIntervalChange = (v: string) => {
    setInterval(v);
    if (v !== 'custom') setCronExpr(INTERVAL_MAP[v]);
  };

  const togglePlatform = (id: string) => {
    setSelectedPlatforms(prev => {
      const next = new Set(prev);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });
  };

  const handleCreate = async () => {
    if (!name.trim() || selectedPlatforms.size === 0 || !cronExpr.trim()) return;
    setSaving(true);
    const tags = hashtags
      .split(/[\s,]+/)
      .map(t => (t.startsWith('#') ? t : `#${t}`))
      .filter(t => t.length > 1);
    const req: CreateScheduleRequest = {
      name,
      type,
      platforms: [...selectedPlatforms],
      cronExpr,
      ...(type === 'upload' ? { caption, hashtags: tags } : {}),
    };
    await CreateSchedule(req);
    setSaving(false);
    onCreated();
  };

  return (
    <div
      className="modal-overlay"
      onClick={onCancel}
      onKeyDown={e => { if (e.key === 'Escape') onCancel(); }}
      role="presentation"
    >
      <div className="modal-content" style={{ width: 500 }} onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h3>Create Schedule</h3>
          <button className="modal-close" onClick={onCancel} aria-label="Close">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
        <div className="modal-body">
          <div className="config-field" style={{ marginBottom: 12 }}>
            <label htmlFor="sched-job-name">Job Name</label>
            <input
              id="sched-job-name"
              type="text"
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder="My upload schedule"
            />
          </div>

          <div className="config-field" style={{ marginBottom: 12 }}>
            <span className="config-field-label">Type</span>
            <div className="theme-switcher" style={{ width: '100%' }}>
              <button
                className={type === 'session_keepalive' ? 'active' : ''}
                onClick={() => setType('session_keepalive')}
                style={{ flex: 1, justifyContent: 'center' }}
              >
                Keep Alive
              </button>
              <button
                className={type === 'upload' ? 'active' : ''}
                onClick={() => setType('upload')}
                style={{ flex: 1, justifyContent: 'center' }}
              >
                Upload
              </button>
            </div>
          </div>

          <div className="config-field" style={{ marginBottom: 12 }}>
            <span className="config-field-label">Interval</span>
            <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
              {(['6h', '12h', '24h', '3d', 'custom'] as const).map(v => (
                <button
                  key={v}
                  className={`btn-secondary ${interval === v ? 'active' : ''}`}
                  style={{
                    padding: '6px 12px', fontSize: 12,
                    ...(interval === v ? { background: 'var(--accent)', color: '#fff', borderColor: 'var(--accent)' } : {}),
                  }}
                  onClick={() => handleIntervalChange(v)}
                >
                  {v === 'custom' ? 'Custom' : `Every ${v}`}
                </button>
              ))}
            </div>
            {interval === 'custom' && (
              <input
                type="text"
                value={cronExpr}
                onChange={e => setCronExpr(e.target.value)}
                placeholder="0 */12 * * *"
                style={{ marginTop: 6 }}
                aria-label="Custom cron expression"
              />
            )}
          </div>

          <div className="config-field" style={{ marginBottom: 12 }}>
            <span className="config-field-label">Platforms</span>
            <div className="upload-platforms">
              {platforms.map(p => (
                <label
                  key={p.id}
                  className={`upload-platform-chip ${selectedPlatforms.has(p.id) ? 'selected' : ''}`}
                >
                  <input
                    type="checkbox"
                    checked={selectedPlatforms.has(p.id)}
                    onChange={() => togglePlatform(p.id)}
                  />
                  <span className="upload-platform-icon">{PLATFORM_ICONS[p.id]}</span>
                  <span>{p.name}</span>
                </label>
              ))}
            </div>
          </div>

          {type === 'upload' && (
            <>
              <div className="config-field" style={{ marginBottom: 12 }}>
                <label htmlFor="sched-caption">Caption</label>
                <textarea
                  id="sched-caption"
                  className="upload-caption"
                  value={caption}
                  onChange={e => setCaption(e.target.value)}
                  placeholder="Post caption..."
                  rows={2}
                />
              </div>
              <div className="config-field" style={{ marginBottom: 0 }}>
                <label htmlFor="sched-hashtags">Hashtags</label>
                <input
                  id="sched-hashtags"
                  type="text"
                  value={hashtags}
                  onChange={e => setHashtags(e.target.value)}
                  placeholder="#hashtag1 #hashtag2"
                />
              </div>
            </>
          )}
        </div>
        <div className="modal-footer">
          <button
            className="btn-primary"
            onClick={handleCreate}
            disabled={saving || !name.trim() || selectedPlatforms.size === 0}
          >
            {saving ? 'Creating...' : 'Create'}
          </button>
          <button className="btn-secondary" onClick={onCancel}>Cancel</button>
        </div>
      </div>
    </div>
  );
}

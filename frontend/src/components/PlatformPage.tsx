import { useState, useEffect, useCallback } from 'react';
import {
  GetAutomations, SaveAutomation, DeleteAutomation, RunAutomation, TestAutomation, CreateSchedule,
  GetSchedules, DeleteSchedule, PauseSchedule, ResumeSchedule,
  GetPlatforms, OpenPlatform, ImportCookies, PurgeSession, LogFromFrontend,
} from '../../wailsjs/go/main/App';
import { main } from '../../wailsjs/go/models';
import type {
  ChatMessage, PlatformResponse, ScheduleResponse,
  AutomationData, AutomationStepData, AutomationAction,
} from '../types';
import { AUTOMATION_ACTIONS } from '../types';
import { PLATFORM_ICONS, PLATFORM_COLORS } from '../constants/platforms';
import { HamburgerBtn } from './HamburgerBtn';
import { CookiePasteModal } from './modals/CookiePasteModal';

interface Props {
  readonly platformId: string;
  readonly platformName: string;
  readonly sidebarOpen: boolean;
  readonly onToggleSidebar: () => void;
  readonly addMessage: (type: ChatMessage['type'], text: string) => void;
  readonly onBrowserStatusChange: (s: string) => void;
  readonly refreshPlatforms: () => void;
}

type Tab = 'define' | 'execute' | 'schedule' | 'history';

const INTERVAL_MAP: Record<string, string> = {
  '6h': '0 */6 * * *',
  '12h': '0 */12 * * *',
  '24h': '0 0 * * *',
  '3d': '0 0 */3 * *',
  custom: '',
};

function emptyStep(): AutomationStepData {
  return { action: 'navigate', target: '', value: '', label: '' };
}

function emptyAutomation(platformId: string): Omit<AutomationData, 'id' | 'createdAt' | 'lastRun' | 'runCount'> {
  return { platformId, name: '', description: '', steps: [] };
}

/* ── Step editor row ── */
function StepRow({
  step, index, total, onChange, onRemove, onMoveUp, onMoveDown,
}: {
  readonly step: AutomationStepData;
  readonly index: number;
  readonly total: number;
  readonly onChange: (s: AutomationStepData) => void;
  readonly onRemove: () => void;
  readonly onMoveUp: () => void;
  readonly onMoveDown: () => void;
}) {
  const actionMeta = AUTOMATION_ACTIONS.find(a => a.value === step.action);

  return (
    <div className="automation-step">
      <div className="automation-step-num">{index + 1}</div>
      <div className="automation-step-fields">
        <div className="automation-step-row">
          <select
            value={step.action}
            onChange={e => onChange({ ...step, action: e.target.value as AutomationAction })}
            aria-label="Action"
          >
            {AUTOMATION_ACTIONS.map(a => (
              <option key={a.value} value={a.value}>{a.label}</option>
            ))}
          </select>
          <input
            type="text"
            value={step.label}
            onChange={e => onChange({ ...step, label: e.target.value })}
            placeholder="Step label (optional)"
            aria-label="Step label"
          />
        </div>
        {actionMeta?.hasTarget && (
          <input
            type="text"
            value={step.target}
            onChange={e => onChange({ ...step, target: e.target.value })}
            placeholder={actionMeta.targetLabel}
            aria-label={actionMeta.targetLabel}
          />
        )}
        {actionMeta?.hasValue && (
          <input
            type="text"
            value={step.value}
            onChange={e => onChange({ ...step, value: e.target.value })}
            placeholder={actionMeta.valueLabel}
            aria-label={actionMeta.valueLabel}
          />
        )}
      </div>
      <div className="automation-step-actions">
        <button onClick={onMoveUp} disabled={index === 0} title="Move up" aria-label="Move step up">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <polyline points="18 15 12 9 6 15"/>
          </svg>
        </button>
        <button onClick={onMoveDown} disabled={index === total - 1} title="Move down" aria-label="Move step down">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <polyline points="6 9 12 15 18 9"/>
          </svg>
        </button>
        <button onClick={onRemove} title="Remove step" aria-label="Remove step" className="btn-remove-step">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>
    </div>
  );
}

/* ── Chrome Recorder JSON → automation steps converter ── */
interface RecorderStep {
  type: string;
  url?: string;
  selectors?: string[][];
  value?: string;
  key?: string;
  offsetX?: number;
  offsetY?: number;
  target?: string;
  width?: number;
  height?: number;
  duration?: number;
}

interface RecorderJson {
  title?: string;
  steps?: RecorderStep[];
}

function convertRecorderJson(json: RecorderJson): AutomationStepData[] {
  if (!json.steps) return [];
  const steps: AutomationStepData[] = [];

  for (const s of json.steps) {
    const selector = s.selectors?.[0]?.[0] ?? '';

    switch (s.type) {
      case 'navigate':
        if (s.url) {
          steps.push({ action: 'navigate', target: s.url, value: '', label: `Navigate to ${s.url}` });
        }
        break;
      case 'click':
      case 'doubleClick':
        if (selector) {
          steps.push({ action: 'click', target: selector, value: '', label: s.type === 'doubleClick' ? `Double click ${selector}` : `Click ${selector}` });
        }
        break;
      case 'change':
        if (selector && s.value !== undefined) {
          steps.push({ action: 'type', target: selector, value: s.value, label: `Type into ${selector}` });
        }
        break;
      case 'keyDown':
      case 'keyUp':
        if (s.key && s.type === 'keyDown') {
          steps.push({ action: 'evaluate', target: '', value: `document.dispatchEvent(new KeyboardEvent('keydown', {key: '${s.key}', bubbles: true}))`, label: `Key down: ${s.key}` });
        }
        break;
      case 'scroll':
        if (selector) {
          steps.push({ action: 'scroll', target: selector, value: '', label: `Scroll to ${selector}` });
        }
        break;
      case 'setViewport':
        if (s.width && s.height) {
          steps.push({ action: 'evaluate', target: '', value: `window.resizeTo(${s.width}, ${s.height})`, label: `Set viewport ${s.width}x${s.height}` });
        }
        break;
      case 'waitForElement':
        if (selector) {
          steps.push({ action: 'wait', target: '', value: '1000', label: `Wait for ${selector}` });
        }
        break;
      default:
        break;
    }
  }

  return steps;
}

/* ── Define tab ── */
function DefineTab({ platformId, automations, onSaved, onDeleted, addMessage }: {
  readonly platformId: string;
  readonly automations: AutomationData[];
  readonly onSaved: () => void;
  readonly onDeleted: (id: string) => void;
  readonly addMessage: (type: ChatMessage['type'], text: string) => void;
}) {
  const [editing, setEditing] = useState<Partial<AutomationData> | null>(null);
  const [saving, setSaving] = useState(false);
  const [testingId, setTestingId] = useState<string | null>(null);
  const [showImportModal, setShowImportModal] = useState(false);
  const [importJson, setImportJson] = useState('');
  const [importError, setImportError] = useState('');

  const startNew = () => setEditing(emptyAutomation(platformId));
  const startEdit = (a: AutomationData) => setEditing({ ...a, steps: [...a.steps] });

  const updateStep = (i: number, step: AutomationStepData) => {
    if (!editing) return;
    const steps = [...(editing.steps ?? [])];
    steps[i] = step;
    setEditing({ ...editing, steps });
  };

  const removeStep = (i: number) => {
    if (!editing) return;
    const steps = (editing.steps ?? []).filter((_, idx) => idx !== i);
    setEditing({ ...editing, steps });
  };

  const moveStep = (i: number, dir: -1 | 1) => {
    if (!editing) return;
    const steps = [...(editing.steps ?? [])];
    const j = i + dir;
    if (j < 0 || j >= steps.length) return;
    [steps[i], steps[j]] = [steps[j], steps[i]];
    setEditing({ ...editing, steps });
  };

  const addStep = () => {
    if (!editing) return;
    setEditing({ ...editing, steps: [...(editing.steps ?? []), emptyStep()] });
  };

  const handleSave = async () => {
    if (!editing?.name?.trim()) return;
    setSaving(true);
    try {
      await SaveAutomation(platformId, main.AutomationPayload.createFrom(editing));
      onSaved();
      setEditing(null);
    } finally {
      setSaving(false);
    }
  };

  const handleImportJson = () => {
    setImportError('');
    try {
      const parsed = JSON.parse(importJson) as RecorderJson;
      if (!parsed.steps || !Array.isArray(parsed.steps)) {
        setImportError('Invalid JSON: missing "steps" array.');
        return;
      }
      const steps = convertRecorderJson(parsed);
      if (steps.length === 0) {
        setImportError('No convertible steps found in the recording.');
        return;
      }
      LogFromFrontend('info', 'platform', `Imported Chrome Recorder JSON: "${parsed.title}" — ${parsed.steps.length} raw steps → ${steps.length} automation steps`);
      setEditing({
        ...emptyAutomation(platformId),
        name: parsed.title || 'Imported Recording',
        description: `Imported from Chrome Recorder (${steps.length} steps)`,
        steps,
      });
      setShowImportModal(false);
      setImportJson('');
    } catch {
      setImportError('Invalid JSON. Please paste a valid Chrome Recorder JSON.');
    }
  };

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => {
      setImportJson(reader.result as string);
      setImportError('');
    };
    reader.readAsText(file);
    e.target.value = '';
  };

  const handleTest = async (a: AutomationData) => {
    setTestingId(a.id);
    LogFromFrontend('info', 'platform', `Test automation "${a.name}" (${a.steps.length} steps) on ${platformId}`);
    try {
      const res = await TestAutomation(platformId, a.id);
      LogFromFrontend(res.success ? 'info' : 'error', 'platform', `Test result: ${res.message}`);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      LogFromFrontend('error', 'platform', `Test exception: ${msg}`);
    }
    setTestingId(null);
    onSaved();
  };

  if (editing) {
    return (
      <div className="automation-editor">
        <div className="automation-editor-header">
          <h3>{editing.id ? 'Edit Automation' : 'New Automation'}</h3>
          <button className="btn-secondary" onClick={() => setEditing(null)}>Cancel</button>
        </div>
        <div className="config-field">
          <label htmlFor="auto-name">Name</label>
          <input id="auto-name" type="text" value={editing.name ?? ''} onChange={e => setEditing({ ...editing, name: e.target.value })} placeholder="Login flow, Post video..." />
        </div>
        <div className="config-field">
          <label htmlFor="auto-desc">Description</label>
          <input id="auto-desc" type="text" value={editing.description ?? ''} onChange={e => setEditing({ ...editing, description: e.target.value })} placeholder="Optional description" />
        </div>
        <div className="automation-steps-label">
          Steps{' '}
          <button className="btn-secondary" style={{ marginLeft: 8, padding: '4px 10px', fontSize: 12 }} onClick={addStep}>+ Add Step</button>
        </div>
        {(editing.steps ?? []).length === 0 ? (
          <div className="automation-empty-steps">No steps yet. Add a step to build your automation.</div>
        ) : (
          <div className="automation-steps">
            {(editing.steps ?? []).map((step, i) => (
              <StepRow key={`${step.action}:${step.target}:${step.label}:${step.value}`} step={step} index={i} total={(editing.steps ?? []).length} onChange={s => updateStep(i, s)} onRemove={() => removeStep(i)} onMoveUp={() => moveStep(i, -1)} onMoveDown={() => moveStep(i, 1)} />
            ))}
          </div>
        )}
        <div className="automation-editor-footer">
          <button className="btn-primary" onClick={handleSave} disabled={saving || !editing.name?.trim()}>{saving ? 'Saving...' : 'Save Automation'}</button>
          <button className="btn-secondary" onClick={() => setEditing(null)}>Cancel</button>
        </div>
      </div>
    );
  }

  return (
    <div className="automation-list-view">
      <div className="automation-list-header">
        <button className="btn-primary" onClick={startNew}>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          New Automation
        </button>
        <button className="btn-secondary" onClick={() => { setShowImportModal(true); setImportJson(''); setImportError(''); }}>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
          Import JSON
        </button>
      </div>
      {automations.length === 0 ? (
        <div className="automation-empty">
          <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="var(--text-secondary)" strokeWidth="1.5"><rect x="3" y="3" width="18" height="18" rx="2"/><line x1="9" y1="9" x2="15" y2="9"/><line x1="9" y1="13" x2="15" y2="13"/><line x1="9" y1="17" x2="12" y2="17"/></svg>
          <p>No automations yet. Create one or import a Chrome Recorder JSON.</p>
        </div>
      ) : (
        <div className="automation-cards">
          {automations.map(a => (
            <div key={a.id} className="automation-card">
              <div className="automation-card-info">
                <span className="automation-card-name">{a.name}</span>
                {a.description && <span className="automation-card-desc">{a.description}</span>}
                <span className="automation-card-meta">{a.steps.length} step{a.steps.length === 1 ? '' : 's'} &middot; {a.runCount} runs</span>
              </div>
              <div className="automation-card-actions">
                <button className="btn-primary" style={{ fontSize: 12, padding: '6px 12px' }} onClick={() => handleTest(a)} disabled={testingId === a.id}>{testingId === a.id ? 'Running...' : 'Test'}</button>
                <button className="btn-secondary" style={{ fontSize: 12 }} onClick={() => startEdit(a)}>Edit</button>
                <button className="btn-danger" style={{ fontSize: 12 }} onClick={() => onDeleted(a.id)}>Delete</button>
              </div>
            </div>
          ))}
        </div>
      )}

      {showImportModal && (
        <dialog className="modal-overlay" open onCancel={e => { e.preventDefault(); setShowImportModal(false); }}>
          <button className="modal-backdrop" aria-label="Close" tabIndex={-1} onClick={() => setShowImportModal(false)} />
          <div className="modal-content import-json-modal">
            <div className="modal-header">
              <h3>Import Chrome Recorder JSON</h3>
              <button className="modal-close" onClick={() => setShowImportModal(false)} aria-label="Close">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" /></svg>
              </button>
            </div>
            <div className="modal-body">
              <p style={{ fontSize: 13, color: 'var(--text-secondary)', marginBottom: 12 }}>Paste a Chrome DevTools Recorder JSON export or upload a .json file.</p>
              <label className="btn-secondary import-file-btn" style={{ alignSelf: 'flex-start', fontSize: 12, padding: '6px 12px', cursor: 'pointer' }}>
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                Upload .json file
                <input type="file" accept=".json" style={{ display: 'none' }} onChange={handleFileUpload} />
              </label>
              <textarea className="import-json-textarea" placeholder='{"title": "Recording...", "steps": [...]}' value={importJson} onChange={e => { setImportJson(e.target.value); setImportError(''); }} rows={14} />
              {importError && <div className="import-json-error">{importError}</div>}
            </div>
            <div className="modal-footer">
              <button className="btn-primary" onClick={handleImportJson} disabled={!importJson.trim()}>Import</button>
              <button className="btn-secondary" onClick={() => setShowImportModal(false)}>Cancel</button>
            </div>
          </div>
        </dialog>
      )}
    </div>
  );
}

/* ── Execute tab ── */
function ExecuteTab({ platformId, automations, addMessage }: {
  readonly platformId: string;
  readonly automations: AutomationData[];
  readonly addMessage: (type: ChatMessage['type'], text: string) => void;
}) {
  const [selectedId, setSelectedId] = useState('');
  const [running, setRunning] = useState(false);

  const handleRun = async () => {
    if (!selectedId) return;
    const auto = automations.find(a => a.id === selectedId);
    if (!auto) return;
    setRunning(true);
    LogFromFrontend('info', 'platform', `Executing automation "${auto.name}" on ${platformId}`);
    try {
      const res = await RunAutomation(platformId, selectedId);
      addMessage(res.success ? 'success' : 'error', res.message);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      addMessage('error', `Automation error: ${msg}`);
    }
    setRunning(false);
  };

  if (automations.length === 0) {
    return <div className="automation-empty"><p>No automations defined. Go to the Define tab to create one first.</p></div>;
  }

  return (
    <div className="automation-execute">
      <div className="config-field">
        <label htmlFor="exec-select">Select Automation</label>
        <select id="exec-select" value={selectedId} onChange={e => setSelectedId(e.target.value)}>
          <option value="">-- Choose --</option>
          {automations.map(a => <option key={a.id} value={a.id}>{a.name}</option>)}
        </select>
      </div>
      {selectedId && (() => {
        const auto = automations.find(a => a.id === selectedId);
        if (!auto) return null;
        return (
          <div className="automation-preview">
            {auto.description && <p className="automation-preview-desc">{auto.description}</p>}
            <div className="automation-preview-steps">
              {auto.steps.map((s, stepIdx) => (
                <div key={`${s.action}:${s.target}:${s.label}:${s.value}`} className="automation-preview-step">
                  <span className="automation-preview-num">{stepIdx + 1}</span>
                  <span className="automation-preview-action">{AUTOMATION_ACTIONS.find(a => a.value === s.action)?.label ?? s.action}</span>
                  {s.label && <span className="automation-preview-label">{s.label}</span>}
                  {s.target && <code className="automation-preview-target">{s.target}</code>}
                  {s.value && <span className="automation-preview-value">{s.value}</span>}
                </div>
              ))}
            </div>
            <div style={{ marginTop: 16 }}>
              <span className="automation-card-meta">{auto.runCount} runs &middot; Last run: {auto.lastRun ? new Date(auto.lastRun).toLocaleString() : 'Never'}</span>
            </div>
          </div>
        );
      })()}
      <div style={{ marginTop: 16 }}>
        <button className="btn-primary" onClick={handleRun} disabled={!selectedId || running}>{running ? 'Running...' : 'Run Now'}</button>
      </div>
    </div>
  );
}

/* ── Schedule tab (create + list schedules for this platform) ── */
function ScheduleTab({ platformId, automations, addMessage }: {
  readonly platformId: string;
  readonly automations: AutomationData[];
  readonly addMessage: (type: ChatMessage['type'], text: string) => void;
}) {
  const [schedules, setSchedules] = useState<ScheduleResponse[]>([]);
  const [selectedId, setSelectedId] = useState('');
  const [scheduleInterval, setScheduleInterval] = useState('12h');
  const [cronExpr, setCronExpr] = useState('0 */12 * * *');
  const [saving, setSaving] = useState(false);

  const loadSchedules = useCallback(() => {
    GetSchedules().then(all => {
      setSchedules(all.filter(s => s.platforms?.includes(platformId)));
    });
  }, [platformId]);

  useEffect(() => { loadSchedules(); }, [loadSchedules]);

  const handleIntervalChange = (v: string) => {
    setScheduleInterval(v);
    if (v !== 'custom') setCronExpr(INTERVAL_MAP[v]);
  };

  const handleSchedule = async () => {
    if (!selectedId || !cronExpr.trim()) return;
    const auto = automations.find(a => a.id === selectedId);
    if (!auto) return;
    setSaving(true);
    try {
      await CreateSchedule({ name: `${auto.name} (${platformId})`, type: 'upload', platforms: [platformId], cronExpr });
      addMessage('success', `Scheduled "${auto.name}" with cron: ${cronExpr}`);
      loadSchedules();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      addMessage('error', `Schedule error: ${msg}`);
    }
    setSaving(false);
  };

  const handleDelete = async (id: string) => {
    await DeleteSchedule(id);
    loadSchedules();
  };

  const handleTogglePause = async (job: ScheduleResponse) => {
    if (job.status === 'paused') await ResumeSchedule(job.id);
    else await PauseSchedule(job.id);
    loadSchedules();
  };

  return (
    <div className="automation-schedule">
      {/* Existing schedules for this platform */}
      {schedules.length > 0 && (
        <div style={{ marginBottom: 24 }}>
          <h4 style={{ fontSize: 13, fontWeight: 600, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.5px', marginBottom: 10 }}>Active Schedules</h4>
          <div className="schedule-list">
            {schedules.map(s => (
              <div key={s.id} className="schedule-card">
                <div className="schedule-card-info">
                  <span className="schedule-card-name">
                    {s.name}
                    <span className={`schedule-badge ${s.type === 'session_keepalive' ? 'keepalive' : 'upload'}`}>{s.type === 'session_keepalive' ? 'keep-alive' : s.type}</span>
                    {s.auto && <span className="schedule-badge auto">auto</span>}
                    <span className={`schedule-badge ${s.status === 'active' ? 'keepalive' : ''}`}>{s.status}</span>
                  </span>
                  <span className="schedule-card-meta">{s.cronExpr} &middot; {s.runCount} runs{s.lastResult ? ` · ${s.lastResult}` : ''}</span>
                </div>
                <div className="schedule-card-stats">
                  <button className="btn-secondary" style={{ fontSize: 11, padding: '4px 10px' }} onClick={() => handleTogglePause(s)}>{s.status === 'paused' ? 'Resume' : 'Pause'}</button>
                  <button className="btn-danger" style={{ fontSize: 11, padding: '4px 10px' }} onClick={() => handleDelete(s.id)}>Delete</button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Create new schedule */}
      <h4 style={{ fontSize: 13, fontWeight: 600, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.5px', marginBottom: 10 }}>New Schedule</h4>
      {automations.length === 0 ? (
        <div className="automation-empty"><p>No automations defined. Go to the Define tab to create one first.</p></div>
      ) : (
        <>
          <div className="config-field">
            <label htmlFor="sched-auto-select">Automation</label>
            <select id="sched-auto-select" value={selectedId} onChange={e => setSelectedId(e.target.value)}>
              <option value="">-- Choose --</option>
              {automations.map(a => <option key={a.id} value={a.id}>{a.name}</option>)}
            </select>
          </div>
          <div className="config-field">
            <span className="config-field-label">Interval</span>
            <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
              {(['6h', '12h', '24h', '3d', 'custom'] as const).map(v => (
                <button key={v} className={`btn-secondary ${scheduleInterval === v ? 'active' : ''}`} style={{ padding: '6px 12px', fontSize: 12, ...(scheduleInterval === v ? { background: 'var(--accent)', color: '#fff', borderColor: 'var(--accent)' } : {}) }} onClick={() => handleIntervalChange(v)}>
                  {v === 'custom' ? 'Custom' : `Every ${v}`}
                </button>
              ))}
            </div>
            {scheduleInterval === 'custom' && (
              <input type="text" value={cronExpr} onChange={e => setCronExpr(e.target.value)} placeholder="0 */12 * * *" style={{ marginTop: 6 }} aria-label="Custom cron expression" />
            )}
          </div>
          <button className="btn-primary" onClick={handleSchedule} disabled={!selectedId || !cronExpr.trim() || saving} style={{ marginTop: 8 }}>
            {saving ? 'Scheduling...' : 'Create Schedule'}
          </button>
        </>
      )}
    </div>
  );
}

/* ── History tab (execution history for this platform) ── */
function HistoryTab({ platformId }: { readonly platformId: string }) {
  const [schedules, setSchedules] = useState<ScheduleResponse[]>([]);

  useEffect(() => {
    GetSchedules().then(all => {
      setSchedules(all.filter(s => s.platforms?.includes(platformId)));
    });
    const interval = setInterval(() => {
      GetSchedules().then(all => {
        setSchedules(all.filter(s => s.platforms?.includes(platformId)));
      });
    }, 5000);
    return () => clearInterval(interval);
  }, [platformId]);

  const executed = schedules.filter(s => s.runCount > 0);

  if (executed.length === 0) {
    return (
      <div className="automation-empty">
        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="var(--text-secondary)" strokeWidth="1.5">
          <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
        </svg>
        <p>No execution history yet. Schedule or run automations to see results here.</p>
      </div>
    );
  }

  return (
    <div className="schedule-list">
      {executed.map(s => (
        <div key={s.id} className="schedule-card">
          <div className="schedule-card-info">
            <span className="schedule-card-name">
              {s.name}
              <span className={`schedule-badge ${s.type === 'session_keepalive' ? 'keepalive' : 'upload'}`}>{s.type === 'session_keepalive' ? 'keep-alive' : s.type}</span>
              {s.auto && <span className="schedule-badge auto">auto</span>}
            </span>
            <span className="schedule-card-meta">
              {s.cronExpr} &middot; {s.runCount} run{s.runCount === 1 ? '' : 's'}
              {s.lastRun && <> &middot; Last: {new Date(s.lastRun).toLocaleString()}</>}
              {s.nextRun && <> &middot; Next: {new Date(s.nextRun).toLocaleString()}</>}
            </span>
            {s.lastResult && (
              <span className="schedule-card-meta" style={{ marginTop: 2 }}>
                Result: {s.lastResult}
              </span>
            )}
          </div>
          <div className="schedule-card-stats">
            <span className={`schedule-badge ${s.status === 'active' ? 'keepalive' : s.status === 'paused' ? 'auto' : ''}`}>{s.status}</span>
          </div>
        </div>
      ))}
    </div>
  );
}

/* ── Platform page root ── */
export function PlatformPage({ platformId, platformName, sidebarOpen, onToggleSidebar, addMessage, onBrowserStatusChange, refreshPlatforms }: Props) {
  const [tab, setTab] = useState<Tab>('define');
  const [automations, setAutomations] = useState<AutomationData[]>([]);
  const [platform, setPlatform] = useState<PlatformResponse | null>(null);
  const [cookieModal, setCookieModal] = useState(false);
  const [infoModal, setInfoModal] = useState(false);
  const [loadingBrowser, setLoadingBrowser] = useState(false);

  const loadAutomations = useCallback(() => {
    GetAutomations(platformId).then(data => setAutomations(data as unknown as AutomationData[]));
  }, [platformId]);

  const loadPlatform = useCallback(() => {
    GetPlatforms().then(all => {
      const p = all.find(p => p.id === platformId) ?? null;
      setPlatform(p);
    });
  }, [platformId]);

  useEffect(() => { loadAutomations(); }, [loadAutomations]);
  useEffect(() => { loadPlatform(); }, [loadPlatform]);

  const handleDeleted = async (id: string) => {
    await DeleteAutomation(platformId, id);
    loadAutomations();
  };

  const handleConnect = async () => {
    setCookieModal(true);
    try {
      const res = await OpenPlatform(platformId);
      if (res.status === 'running') onBrowserStatusChange('running');
      else if (res.error) addMessage('error', `Browser: ${res.error}`);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      addMessage('error', `Failed to open browser: ${msg}`);
    }
  };

  const handleOpenInBrowser = async () => {
    setLoadingBrowser(true);
    LogFromFrontend('info', 'platform', `Opening ${platformName} in browser`);
    try {
      const res = await OpenPlatform(platformId);
      if (res.status === 'running') onBrowserStatusChange('running');
      else LogFromFrontend('error', 'platform', `Open ${platformName} failed: ${res.error}`);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      LogFromFrontend('error', 'platform', `Open ${platformName} error: ${msg}`);
    }
    setLoadingBrowser(false);
  };

  const handleImportCookies = async (cookieText: string) => {
    LogFromFrontend('info', 'platform', `Importing cookies for ${platformName} (${cookieText.split('\n').length} lines)`);
    try {
      const status = await ImportCookies(platformId, cookieText);
      setPlatform(status);
      if (status.loggedIn) {
        LogFromFrontend('info', 'platform', `${platformName} cookies imported successfully`);
        refreshPlatforms();
      } else {
        LogFromFrontend('warn', 'platform', `${platformName} cookie import failed — not logged in after import`);
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      LogFromFrontend('error', 'platform', `${platformName} cookie import exception: ${msg}`);
    }
    setCookieModal(false);
  };

  const handlePurgeSession = async () => {
    LogFromFrontend('info', 'platform', `Purging session for ${platformName}`);
    const status = await PurgeSession(platformId);
    setPlatform(status);
    refreshPlatforms();
    setInfoModal(false);
  };

  const colors = PLATFORM_COLORS[platformId] ?? { bg: '#333', hover: '#444', text: '#fff' };

  return (
    <div className="config-panel platform-page">
      <div className="config-header">
        <HamburgerBtn sidebarOpen={sidebarOpen} onToggle={onToggleSidebar} />
        <div className="platform-page-title">
          <span className="platform-page-icon">{PLATFORM_ICONS[platformId]}</span>
          <h2>{platformName}</h2>
        </div>
      </div>

      {/* Connection status bar */}
      <div className="platform-connection" style={{ '--platform-bg': colors.bg, '--platform-hover': colors.hover, '--platform-text': colors.text } as React.CSSProperties}>
        <div className="platform-connection-status">
          <span className={`status-dot ${platform?.loggedIn ? 'running' : 'stopped'}`} />
          <span>{platform?.loggedIn ? (platform.username || 'Connected') : 'Not connected'}</span>
          {platform?.loggedIn && platform.lastLogin && (
            <span className="platform-connection-meta">since {new Date(platform.lastLogin).toLocaleDateString()}</span>
          )}
        </div>
        <div className="platform-connection-actions">
          {platform?.loggedIn ? (
            <>
              <button className="btn-secondary" onClick={handleOpenInBrowser} disabled={loadingBrowser}>{loadingBrowser ? 'Opening...' : 'Open in Browser'}</button>
              <button className="btn-icon" onClick={() => setInfoModal(true)} title="Session info" aria-label="Session info">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="10" /><line x1="12" y1="16" x2="12" y2="12" /><line x1="12" y1="8" x2="12.01" y2="8" /></svg>
              </button>
            </>
          ) : (
            <button className="btn-primary" onClick={handleConnect}>Connect</button>
          )}
        </div>
      </div>

      <div className="platform-tabs">
        {(['define', 'execute', 'schedule', 'history'] as Tab[]).map(t => (
          <button key={t} className={`platform-tab ${tab === t ? 'active' : ''}`} onClick={() => setTab(t)}>
            {t.charAt(0).toUpperCase() + t.slice(1)}
          </button>
        ))}
      </div>

      <div className="platform-tab-content">
        {tab === 'define' && <DefineTab platformId={platformId} automations={automations} onSaved={loadAutomations} onDeleted={handleDeleted} addMessage={addMessage} />}
        {tab === 'execute' && <ExecuteTab platformId={platformId} automations={automations} addMessage={addMessage} />}
        {tab === 'schedule' && <ScheduleTab platformId={platformId} automations={automations} addMessage={addMessage} />}
        {tab === 'history' && <HistoryTab platformId={platformId} />}
      </div>

      {cookieModal && platform && <CookiePasteModal platform={platform} onImport={handleImportCookies} onCancel={() => setCookieModal(false)} />}

      {infoModal && platform && (
        <dialog className="modal-overlay" open onCancel={e => { e.preventDefault(); setInfoModal(false); }}>
          <button className="modal-backdrop" aria-label="Close" tabIndex={-1} onClick={() => setInfoModal(false)} />
          <div className="modal-content">
            <div className="modal-header">
              <div className="modal-title-row">
                <div className="social-card-icon" style={{ '--platform-bg': colors.bg, '--platform-text': '#fff' } as React.CSSProperties}>{PLATFORM_ICONS[platformId]}</div>
                <h3>{platformName}</h3>
              </div>
              <button className="modal-close" onClick={() => setInfoModal(false)} aria-label="Close">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" /></svg>
              </button>
            </div>
            <div className="modal-body">
              <div className="modal-field"><span className="modal-label">Status</span><span className={`modal-value ${platform.loggedIn ? 'text-success' : 'text-muted'}`}><span className={`status-dot ${platform.loggedIn ? 'running' : 'stopped'}`} />{platform.loggedIn ? 'Connected' : 'Not connected'}</span></div>
              {platform.username && <div className="modal-field"><span className="modal-label">User / ID</span><span className="modal-value">{platform.username}</span></div>}
              <div className="modal-field"><span className="modal-label">URL</span><span className="modal-value modal-url">{platform.url}</span></div>
              <div className="modal-field"><span className="modal-label">Session Directory</span><span className="modal-value modal-path">{platform.sessionDir}</span></div>
              {platform.lastLogin && <div className="modal-field"><span className="modal-label">Logged In</span><span className="modal-value">{new Date(platform.lastLogin).toLocaleString()}</span></div>}
              {platform.lastCheck && <div className="modal-field"><span className="modal-label">Last Verified</span><span className="modal-value">{new Date(platform.lastCheck).toLocaleString()}</span></div>}
            </div>
            <div className="modal-footer">
              <button className="btn-primary" onClick={() => { handleOpenInBrowser(); setInfoModal(false); }}>Open</button>
              {platform.loggedIn && (
                <button className="btn-danger" onClick={handlePurgeSession}>
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="3 6 5 6 21 6" /><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" /></svg>
                  Purge Session
                </button>
              )}
              <button className="btn-secondary" onClick={() => setInfoModal(false)}>Close</button>
            </div>
          </div>
        </dialog>
      )}
    </div>
  );
}

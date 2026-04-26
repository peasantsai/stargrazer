import { useState, useEffect } from 'react';
import {
  GetAutomations, SaveAutomation, DeleteAutomation, RunAutomation, CreateSchedule,
} from '../../wailsjs/go/main/App';
import { main } from '../../wailsjs/go/models';
import type {
  ChatMessage, AutomationData, AutomationStepData, AutomationAction,
} from '../types';
import { AUTOMATION_ACTIONS } from '../types';
import { PLATFORM_ICONS } from '../constants/platforms';
import { HamburgerBtn } from './HamburgerBtn';

interface Props {
  platformId: string;
  platformName: string;
  sidebarOpen: boolean;
  onToggleSidebar: () => void;
  addMessage: (type: ChatMessage['type'], text: string) => void;
}

type Tab = 'define' | 'execute' | 'schedule';

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
  step: AutomationStepData;
  index: number;
  total: number;
  onChange: (s: AutomationStepData) => void;
  onRemove: () => void;
  onMoveUp: () => void;
  onMoveDown: () => void;
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

/* ── Define tab ── */
function DefineTab({ platformId, automations, onSaved, onDeleted }: {
  platformId: string;
  automations: AutomationData[];
  onSaved: () => void;
  onDeleted: (id: string) => void;
}) {
  const [editing, setEditing] = useState<Partial<AutomationData> | null>(null);
  const [saving, setSaving] = useState(false);

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

  if (editing) {
    return (
      <div className="automation-editor">
        <div className="automation-editor-header">
          <h3>{editing.id ? 'Edit Automation' : 'New Automation'}</h3>
          <button className="btn-secondary" onClick={() => setEditing(null)}>Cancel</button>
        </div>
        <div className="config-field">
          <label htmlFor="auto-name">Name</label>
          <input
            id="auto-name"
            type="text"
            value={editing.name ?? ''}
            onChange={e => setEditing({ ...editing, name: e.target.value })}
            placeholder="Login flow, Post video..."
          />
        </div>
        <div className="config-field">
          <label htmlFor="auto-desc">Description</label>
          <input
            id="auto-desc"
            type="text"
            value={editing.description ?? ''}
            onChange={e => setEditing({ ...editing, description: e.target.value })}
            placeholder="Optional description"
          />
        </div>
        <div className="automation-steps-label">
          Steps
          <button className="btn-secondary" style={{ marginLeft: 8, padding: '4px 10px', fontSize: 12 }} onClick={addStep}>
            + Add Step
          </button>
        </div>
        {(editing.steps ?? []).length === 0 ? (
          <div className="automation-empty-steps">No steps yet. Add a step to build your automation.</div>
        ) : (
          <div className="automation-steps">
            {(editing.steps ?? []).map((step, i) => (
              <StepRow
                key={`step-${i}`}
                step={step}
                index={i}
                total={(editing.steps ?? []).length}
                onChange={s => updateStep(i, s)}
                onRemove={() => removeStep(i)}
                onMoveUp={() => moveStep(i, -1)}
                onMoveDown={() => moveStep(i, 1)}
              />
            ))}
          </div>
        )}
        <div className="automation-editor-footer">
          <button
            className="btn-primary"
            onClick={handleSave}
            disabled={saving || !editing.name?.trim()}
          >
            {saving ? 'Saving...' : 'Save Automation'}
          </button>
          <button className="btn-secondary" onClick={() => setEditing(null)}>Cancel</button>
        </div>
      </div>
    );
  }

  return (
    <div className="automation-list-view">
      <div className="automation-list-header">
        <button className="btn-primary" onClick={startNew}>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          New Automation
        </button>
      </div>
      {automations.length === 0 ? (
        <div className="automation-empty">
          <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="var(--text-secondary)" strokeWidth="1.5">
            <rect x="3" y="3" width="18" height="18" rx="2"/>
            <line x1="9" y1="9" x2="15" y2="9"/><line x1="9" y1="13" x2="15" y2="13"/><line x1="9" y1="17" x2="12" y2="17"/>
          </svg>
          <p>No automations yet. Create one to define browser workflows for this platform.</p>
        </div>
      ) : (
        <div className="automation-cards">
          {automations.map(a => (
            <div key={a.id} className="automation-card">
              <div className="automation-card-info">
                <span className="automation-card-name">{a.name}</span>
                {a.description && <span className="automation-card-desc">{a.description}</span>}
                <span className="automation-card-meta">{a.steps.length} step{a.steps.length !== 1 ? 's' : ''} &middot; {a.runCount} runs</span>
              </div>
              <div className="automation-card-actions">
                <button className="btn-secondary" style={{ fontSize: 12 }} onClick={() => startEdit(a)}>Edit</button>
                <button className="btn-danger" style={{ fontSize: 12 }} onClick={() => onDeleted(a.id)}>Delete</button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

/* ── Execute tab ── */
function ExecuteTab({ platformId, automations, addMessage }: {
  platformId: string;
  automations: AutomationData[];
  addMessage: (type: ChatMessage['type'], text: string) => void;
}) {
  const [selectedId, setSelectedId] = useState('');
  const [running, setRunning] = useState(false);

  const handleRun = async () => {
    if (!selectedId) return;
    const auto = automations.find(a => a.id === selectedId);
    if (!auto) return;
    setRunning(true);
    addMessage('system', `Running "${auto.name}"...`);
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
    return (
      <div className="automation-empty">
        <p>No automations defined. Go to the Define tab to create one first.</p>
      </div>
    );
  }

  return (
    <div className="automation-execute">
      <div className="config-field">
        <label htmlFor="exec-select">Select Automation</label>
        <select id="exec-select" value={selectedId} onChange={e => setSelectedId(e.target.value)}>
          <option value="">-- Choose --</option>
          {automations.map(a => (
            <option key={a.id} value={a.id}>{a.name}</option>
          ))}
        </select>
      </div>
      {selectedId && (() => {
        const auto = automations.find(a => a.id === selectedId);
        if (!auto) return null;
        return (
          <div className="automation-preview">
            {auto.description && <p className="automation-preview-desc">{auto.description}</p>}
            <div className="automation-preview-steps">
              {auto.steps.map((s, i) => (
                <div key={`prev-${i}`} className="automation-preview-step">
                  <span className="automation-preview-num">{i + 1}</span>
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
        <button
          className="btn-primary"
          onClick={handleRun}
          disabled={!selectedId || running}
        >
          {running ? 'Running...' : 'Run Now'}
        </button>
      </div>
    </div>
  );
}

/* ── Schedule tab ── */
function ScheduleTab({ platformId, automations, addMessage }: {
  platformId: string;
  automations: AutomationData[];
  addMessage: (type: ChatMessage['type'], text: string) => void;
}) {
  const [selectedId, setSelectedId] = useState('');
  const [scheduleInterval, setScheduleInterval] = useState('12h');
  const [cronExpr, setCronExpr] = useState('0 */12 * * *');
  const [saving, setSaving] = useState(false);

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
      await CreateSchedule({
        name: `${auto.name} (${platformId})`,
        type: 'upload',
        platforms: [platformId],
        cronExpr,
      });
      addMessage('success', `Scheduled "${auto.name}" with cron: ${cronExpr}`);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      addMessage('error', `Schedule error: ${msg}`);
    }
    setSaving(false);
  };

  if (automations.length === 0) {
    return (
      <div className="automation-empty">
        <p>No automations defined. Go to the Define tab to create one first.</p>
      </div>
    );
  }

  return (
    <div className="automation-schedule">
      <div className="config-field">
        <label htmlFor="sched-auto-select">Automation</label>
        <select id="sched-auto-select" value={selectedId} onChange={e => setSelectedId(e.target.value)}>
          <option value="">-- Choose --</option>
          {automations.map(a => (
            <option key={a.id} value={a.id}>{a.name}</option>
          ))}
        </select>
      </div>
      <div className="config-field">
        <span className="config-field-label">Interval</span>
        <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
          {(['6h', '12h', '24h', '3d', 'custom'] as const).map(v => (
            <button
              key={v}
              className={`btn-secondary ${scheduleInterval === v ? 'active' : ''}`}
              style={{
                padding: '6px 12px', fontSize: 12,
                ...(scheduleInterval === v ? { background: 'var(--accent)', color: '#fff', borderColor: 'var(--accent)' } : {}),
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
      <button
        className="btn-primary"
        onClick={handleSchedule}
        disabled={!selectedId || !cronExpr.trim() || saving}
        style={{ marginTop: 8 }}
      >
        {saving ? 'Scheduling...' : 'Create Schedule'}
      </button>
    </div>
  );
}

/* ── Platform page root ── */
export function PlatformPage({ platformId, platformName, sidebarOpen, onToggleSidebar, addMessage }: Props) {
  const [tab, setTab] = useState<Tab>('define');
  const [automations, setAutomations] = useState<AutomationData[]>([]);

  const loadAutomations = () => {
    GetAutomations(platformId).then(data => setAutomations(data as unknown as AutomationData[]));
  };

  useEffect(() => { loadAutomations(); }, [platformId]);

  const handleDeleted = async (id: string) => {
    await DeleteAutomation(platformId, id);
    addMessage('info', 'Automation deleted.');
    loadAutomations();
  };

  return (
    <div className="config-panel platform-page">
      <div className="config-header">
        <HamburgerBtn sidebarOpen={sidebarOpen} onToggle={onToggleSidebar} />
        <div className="platform-page-title">
          <span className="platform-page-icon">{PLATFORM_ICONS[platformId]}</span>
          <h2>{platformName}</h2>
        </div>
      </div>

      <div className="platform-tabs">
        {(['define', 'execute', 'schedule'] as Tab[]).map(t => (
          <button
            key={t}
            className={`platform-tab ${tab === t ? 'active' : ''}`}
            onClick={() => setTab(t)}
          >
            {t.charAt(0).toUpperCase() + t.slice(1)}
          </button>
        ))}
      </div>

      <div className="platform-tab-content">
        {tab === 'define' && (
          <DefineTab
            platformId={platformId}
            automations={automations}
            onSaved={loadAutomations}
            onDeleted={handleDeleted}
          />
        )}
        {tab === 'execute' && (
          <ExecuteTab
            platformId={platformId}
            automations={automations}
            addMessage={addMessage}
          />
        )}
        {tab === 'schedule' && (
          <ScheduleTab
            platformId={platformId}
            automations={automations}
            addMessage={addMessage}
          />
        )}
      </div>
    </div>
  );
}

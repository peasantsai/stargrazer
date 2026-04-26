import { useState, useEffect } from 'react';
import { GetSchedules, DeleteSchedule, PauseSchedule, ResumeSchedule } from '../../wailsjs/go/main/App';
import type { ChatMessage, PlatformResponse, ScheduleResponse } from '../types';
import { HamburgerBtn } from './HamburgerBtn';
import { CreateScheduleModal } from './modals/CreateScheduleModal';
import { ScheduleDetailModal } from './modals/ScheduleDetailModal';

interface Props {
  sidebarOpen: boolean;
  onToggleSidebar: () => void;
  addMessage: (type: ChatMessage['type'], text: string) => void;
  platforms: PlatformResponse[];
}

function statusDotClass(status: string): string {
  if (status === 'active') return 'running';
  if (status === 'paused') return 'stopped';
  return 'error';
}

export function SchedulesPanel({ sidebarOpen, onToggleSidebar, addMessage, platforms }: Props) {
  const [schedules, setSchedules] = useState<ScheduleResponse[]>([]);
  const [showCreate, setShowCreate] = useState(false);
  const [selectedJob, setSelectedJob] = useState<ScheduleResponse | null>(null);

  const refresh = () => { GetSchedules().then(setSchedules); };
  useEffect(() => { refresh(); }, []);

  const handleDelete = async (id: string) => {
    await DeleteSchedule(id);
    addMessage('info', 'Schedule deleted.');
    refresh();
    setSelectedJob(null);
  };

  const handlePauseResume = async (j: ScheduleResponse) => {
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
        <button
          className="btn-primary"
          style={{ marginLeft: 'auto', padding: '8px 16px', fontSize: 13 }}
          onClick={() => setShowCreate(true)}
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          Create Schedule
        </button>
      </div>

      {schedules.length === 0 ? (
        <div className="chat-empty" style={{ paddingTop: 80 }}>
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--text-secondary)" strokeWidth="1.5">
            <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
          </svg>
          <h3>No Scheduled Jobs</h3>
          <p>Create a schedule to automate uploads or keep sessions alive. Keep-alive jobs are created automatically when you connect a platform.</p>
        </div>
      ) : (
        <div className="schedule-list">
          {schedules.map(j => (
            <div
              key={j.id}
              className="schedule-card"
              onClick={() => setSelectedJob(j)}
              onKeyDown={e => { if (e.key === 'Enter') setSelectedJob(j); }}
              role="button"
              tabIndex={0}
            >
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
                {j.lastResult && (
                  <span className={j.lastResult === 'success' ? 'text-success' : 'text-error'}>
                    {j.lastResult}
                  </span>
                )}
              </div>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--text-secondary)" strokeWidth="2">
                <polyline points="9 18 15 12 9 6"/>
              </svg>
            </div>
          ))}
        </div>
      )}

      {showCreate && (
        <CreateScheduleModal
          platforms={platforms}
          onCreated={() => { refresh(); setShowCreate(false); addMessage('success', 'Schedule created.'); }}
          onCancel={() => setShowCreate(false)}
        />
      )}
      {selectedJob && (
        <ScheduleDetailModal
          job={selectedJob}
          onClose={() => { setSelectedJob(null); refresh(); }}
          onDelete={handleDelete}
          onPauseResume={handlePauseResume}
        />
      )}
    </div>
  );
}

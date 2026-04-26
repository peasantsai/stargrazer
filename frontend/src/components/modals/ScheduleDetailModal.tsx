import type { ScheduleResponse } from '../../types';

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

interface Props {
  readonly job: ScheduleResponse;
  readonly onClose: () => void;
  readonly onDelete: (id: string) => void;
  readonly onPauseResume: (j: ScheduleResponse) => void;
}

export function ScheduleDetailModal({ job, onClose, onDelete, onPauseResume }: Props) {
  return (
    <dialog
      className="modal-overlay"
      open
      onClick={e => { if (e.target === e.currentTarget) onClose(); }}
      onKeyDown={e => { if (e.key === 'Escape') onClose(); }}
    >
      <div className="modal-content">
        <div className="modal-header">
          <h3>{job.name}</h3>
          <button className="modal-close" onClick={onClose} aria-label="Close">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
        <div className="modal-body">
          <div className="modal-field">
            <span className="modal-label">Status</span>
            <span className={`modal-value ${statusTextClass(job.status)}`}>
              <span className={`status-dot ${statusDotClass(job.status)}`} />{job.status}
            </span>
          </div>
          <div className="modal-field">
            <span className="modal-label">Type</span>
            <span className="modal-value">
              {job.type === 'session_keepalive' ? 'Session Keep-Alive' : 'Content Upload'}
              {job.auto && ' (auto-created)'}
            </span>
          </div>
          <div className="modal-field">
            <span className="modal-label">Platforms</span>
            <span className="modal-value">{job.platforms.join(', ')}</span>
          </div>
          <div className="modal-field">
            <span className="modal-label">Schedule</span>
            <span className="modal-value modal-path">{job.cronExpr}</span>
          </div>
          <div className="modal-field">
            <span className="modal-label">Runs</span>
            <span className="modal-value">{job.runCount}</span>
          </div>
          {job.lastRun && (
            <div className="modal-field">
              <span className="modal-label">Last Run</span>
              <span className="modal-value">{new Date(job.lastRun).toLocaleString()}</span>
            </div>
          )}
          {job.lastResult && (
            <div className="modal-field">
              <span className="modal-label">Last Result</span>
              <span className={`modal-value ${job.lastResult === 'success' ? 'text-success' : 'text-error'}`}>
                {job.lastResult}
              </span>
            </div>
          )}
          {job.nextRun && (
            <div className="modal-field">
              <span className="modal-label">Next Run</span>
              <span className="modal-value">{new Date(job.nextRun).toLocaleString()}</span>
            </div>
          )}
          <div className="modal-field">
            <span className="modal-label">Created</span>
            <span className="modal-value">{new Date(job.createdAt).toLocaleString()}</span>
          </div>
          {job.caption && (
            <div className="modal-field">
              <span className="modal-label">Caption</span>
              <span className="modal-value">{job.caption}</span>
            </div>
          )}
          {job.hashtags && job.hashtags.length > 0 && (
            <div className="modal-field">
              <span className="modal-label">Hashtags</span>
              <span className="modal-value">{job.hashtags.join(' ')}</span>
            </div>
          )}
        </div>
        <div className="modal-footer">
          <button
            className={job.status === 'active' ? 'btn-secondary' : 'btn-primary'}
            onClick={() => onPauseResume(job)}
          >
            {job.status === 'active' ? 'Pause' : 'Resume'}
          </button>
          <button className="btn-danger" onClick={() => onDelete(job.id)}>Delete</button>
          <button className="btn-secondary" onClick={onClose}>Close</button>
        </div>
      </div>
    </dialog>
  );
}

import type { RunStepEvent } from '../hooks/useRunEvents';
import { CancelRun } from '../../wailsjs/go/main/App';

interface Props {
  readonly current: RunStepEvent | null;
}

export function RunStrip({ current }: Props) {
  if (!current) return null;

  const handleCancel = () => { CancelRun(); };

  return (
    <div className={`run-strip run-strip-${current.status}`}>
      <span className="run-strip-icon" aria-hidden="true">▶</span>
      <span className="run-strip-name">Running: {current.automationName}</span>
      <span className="run-strip-progress">step {current.stepIndex + 1}/{current.total}</span>
      <span className="run-strip-action">({current.action})</span>
      {current.error && <span className="run-strip-error">{current.error}</span>}
      <button
        type="button"
        className="run-strip-cancel"
        aria-label="Cancel run"
        onClick={handleCancel}
      >
        ✕
      </button>
    </div>
  );
}

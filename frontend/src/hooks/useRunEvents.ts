import { useEffect, useRef, useState } from 'react';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';

export interface RunStepEvent {
  runId?: string;
  automationId: string;
  automationName: string;
  stepIndex: number;
  total: number;
  action: string;
  target?: string;
  status: 'running' | 'success' | 'failed';
  startedAt?: string;
  finishedAt?: string;
  error?: string;
}

export interface RunEventsState {
  current: RunStepEvent | null;
  history: RunStepEvent[];
}

const HISTORY_CAP = 50;
const CLEAR_DELAY_MS = 5_000;

export function useRunEvents(): RunEventsState {
  const [state, setState] = useState<RunEventsState>({ current: null, history: [] });
  const clearTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    const handler = (payload: unknown) => {
      const ev = payload as RunStepEvent;
      setState(prev => {
        const history = [ev, ...prev.history].slice(0, HISTORY_CAP);
        const isFinalStepOfRun = (ev.status === 'success' || ev.status === 'failed')
          && ev.stepIndex === ev.total - 1;

        if (clearTimerRef.current) {
          clearTimeout(clearTimerRef.current);
          clearTimerRef.current = null;
        }
        if (isFinalStepOfRun) {
          clearTimerRef.current = setTimeout(() => {
            setState(s => ({ ...s, current: null }));
          }, CLEAR_DELAY_MS);
        }

        return { current: ev, history };
      });
    };

    EventsOn('run.step', handler);
    return () => {
      EventsOff('run.step');
      if (clearTimerRef.current) clearTimeout(clearTimerRef.current);
    };
  }, []);

  return state;
}

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useRunEvents } from './useRunEvents';

let lastTopic = '';
let lastHandler: ((payload: unknown) => void) | null = null;

vi.mock('../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn((topic: string, handler: (p: unknown) => void) => {
    lastTopic = topic;
    lastHandler = handler;
    return () => { lastHandler = null; };
  }),
  EventsOff: vi.fn(),
}));

beforeEach(() => {
  lastTopic = '';
  lastHandler = null;
});

describe('useRunEvents', () => {
  it('subscribes to run.step on mount', () => {
    renderHook(() => useRunEvents());
    expect(lastTopic).toBe('run.step');
    expect(lastHandler).toBeTypeOf('function');
  });

  it('exposes the latest running event as current', () => {
    const { result } = renderHook(() => useRunEvents());
    expect(result.current.current).toBeNull();

    act(() => {
      lastHandler?.({
        runId: '', automationId: 'a1', automationName: 'Job', stepIndex: 0, total: 3,
        action: 'click', status: 'running', startedAt: new Date().toISOString(),
      });
    });
    expect(result.current.current?.automationName).toBe('Job');
    expect(result.current.current?.stepIndex).toBe(0);
  });

  it('clears current 5s after a final success of the last step', () => {
    vi.useFakeTimers();
    const { result } = renderHook(() => useRunEvents());

    act(() => {
      lastHandler?.({ automationId: 'a1', automationName: 'Job', stepIndex: 0, total: 1, action: 'click', status: 'running' });
    });
    act(() => {
      lastHandler?.({ automationId: 'a1', automationName: 'Job', stepIndex: 0, total: 1, action: 'click', status: 'success' });
    });
    expect(result.current.current).not.toBeNull();

    act(() => { vi.advanceTimersByTime(5_000); });
    expect(result.current.current).toBeNull();
    vi.useRealTimers();
  });
});

import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { RunStrip } from './RunStrip';
import type { RunStepEvent } from '../hooks/useRunEvents';

vi.mock('../../wailsjs/go/main/App', () => ({
  CancelRun: vi.fn().mockResolvedValue(true),
}));

const sample: RunStepEvent = {
  automationId: 'a1', automationName: 'My Automation',
  stepIndex: 2, total: 5, action: 'click', status: 'running',
};

describe('RunStrip', () => {
  it('renders nothing when current is null', () => {
    const { container } = render(<RunStrip current={null} />);
    expect(container.firstChild).toBeNull();
  });

  it('shows automation name, current step, and action', () => {
    render(<RunStrip current={sample} />);
    expect(screen.getByText(/My Automation/)).toBeInTheDocument();
    expect(screen.getByText(/3\s*\/\s*5/)).toBeInTheDocument();
    expect(screen.getByText(/click/)).toBeInTheDocument();
  });

  it('calls CancelRun when ✕ is clicked', async () => {
    const { CancelRun } = await import('../../wailsjs/go/main/App');
    render(<RunStrip current={sample} />);
    fireEvent.click(screen.getByRole('button', { name: /cancel run/i }));
    expect(CancelRun).toHaveBeenCalled();
  });

  it('uses status colour classes', () => {
    const { container, rerender } = render(<RunStrip current={sample} />);
    expect(container.firstChild).toHaveClass('run-strip', 'run-strip-running');
    rerender(<RunStrip current={{ ...sample, status: 'success' }} />);
    expect(container.firstChild).toHaveClass('run-strip-success');
    rerender(<RunStrip current={{ ...sample, status: 'failed' }} />);
    expect(container.firstChild).toHaveClass('run-strip-failed');
  });
});

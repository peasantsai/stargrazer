import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { SchedulesPanel } from './SchedulesPanel';
import * as wailsMocks from '../test/wailsMock';
import type { PlatformResponse, ScheduleResponse } from '../types';

const platforms: PlatformResponse[] = [
  { id: 'instagram', name: 'Instagram', url: '', loggedIn: true, username: 'u', lastLogin: '', lastCheck: '', sessionDir: '' },
];

const activeJob: ScheduleResponse = {
  id: 'j1', name: 'Daily Upload', type: 'upload',
  platforms: ['instagram'], cronExpr: '0 0 * * *',
  nextRun: '2024-06-01T00:00:00Z', lastRun: '', status: 'active',
  createdAt: '2024-01-01T00:00:00Z', runCount: 3, lastResult: 'success', auto: false,
};

beforeEach(() => {
  vi.clearAllMocks();
  (wailsMocks.GetSchedules as ReturnType<typeof vi.fn>).mockResolvedValue([]);
  (wailsMocks.DeleteSchedule as ReturnType<typeof vi.fn>).mockResolvedValue(true);
  (wailsMocks.PauseSchedule as ReturnType<typeof vi.fn>).mockResolvedValue({ id: 'j1', status: 'paused', name: 'Daily Upload' });
  (wailsMocks.ResumeSchedule as ReturnType<typeof vi.fn>).mockResolvedValue({ id: 'j1', status: 'active', name: 'Daily Upload' });
});

describe('SchedulesPanel – empty state', () => {
  it('shows empty state when no schedules', async () => {
    render(<SchedulesPanel sidebarOpen={false} onToggleSidebar={vi.fn()} addMessage={vi.fn()} platforms={platforms} />);
    await waitFor(() => {
      expect(screen.getByText('No Scheduled Jobs')).toBeInTheDocument();
    });
  });

  it('shows Create Schedule button', () => {
    render(<SchedulesPanel sidebarOpen={false} onToggleSidebar={vi.fn()} addMessage={vi.fn()} platforms={platforms} />);
    expect(screen.getByRole('button', { name: /create schedule/i })).toBeInTheDocument();
  });

  it('fetches schedules on mount', async () => {
    render(<SchedulesPanel sidebarOpen={false} onToggleSidebar={vi.fn()} addMessage={vi.fn()} platforms={platforms} />);
    await waitFor(() => {
      expect(wailsMocks.GetSchedules).toHaveBeenCalled();
    });
  });
});

describe('SchedulesPanel – with schedules', () => {
  beforeEach(() => {
    (wailsMocks.GetSchedules as ReturnType<typeof vi.fn>).mockResolvedValue([activeJob]);
  });

  it('renders schedule name', async () => {
    render(<SchedulesPanel sidebarOpen={false} onToggleSidebar={vi.fn()} addMessage={vi.fn()} platforms={platforms} />);
    await waitFor(() => {
      expect(screen.getByText('Daily Upload')).toBeInTheDocument();
    });
  });

  it('renders upload badge for upload type', async () => {
    render(<SchedulesPanel sidebarOpen={false} onToggleSidebar={vi.fn()} addMessage={vi.fn()} platforms={platforms} />);
    await waitFor(() => {
      expect(screen.getByText('upload')).toBeInTheDocument();
    });
  });

  it('renders keep-alive badge for keepalive type', async () => {
    (wailsMocks.GetSchedules as ReturnType<typeof vi.fn>).mockResolvedValue([
      { ...activeJob, type: 'session_keepalive' },
    ]);
    render(<SchedulesPanel sidebarOpen={false} onToggleSidebar={vi.fn()} addMessage={vi.fn()} platforms={platforms} />);
    await waitFor(() => {
      expect(screen.getByText('keep-alive')).toBeInTheDocument();
    });
  });

  it('renders auto badge for auto jobs', async () => {
    (wailsMocks.GetSchedules as ReturnType<typeof vi.fn>).mockResolvedValue([
      { ...activeJob, auto: true },
    ]);
    render(<SchedulesPanel sidebarOpen={false} onToggleSidebar={vi.fn()} addMessage={vi.fn()} platforms={platforms} />);
    await waitFor(() => {
      expect(screen.getByText('auto')).toBeInTheDocument();
    });
  });

  it('renders run count', async () => {
    render(<SchedulesPanel sidebarOpen={false} onToggleSidebar={vi.fn()} addMessage={vi.fn()} platforms={platforms} />);
    await waitFor(() => {
      expect(screen.getByText('3 runs')).toBeInTheDocument();
    });
  });

  it('opens detail modal when schedule card is clicked', async () => {
    render(<SchedulesPanel sidebarOpen={false} onToggleSidebar={vi.fn()} addMessage={vi.fn()} platforms={platforms} />);
    await waitFor(() => screen.getByText('Daily Upload'));
    fireEvent.click(screen.getByRole('button', { name: /daily upload/i }));
    // Pause button only appears in the detail modal
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /^pause$/i })).toBeInTheDocument();
    });
  });
});

describe('SchedulesPanel – create modal', () => {
  it('opens CreateScheduleModal when Create Schedule is clicked', async () => {
    render(<SchedulesPanel sidebarOpen={false} onToggleSidebar={vi.fn()} addMessage={vi.fn()} platforms={platforms} />);
    fireEvent.click(screen.getByRole('button', { name: /create schedule/i }));
    // The modal has a Job Name label that only appears inside the modal
    await waitFor(() => {
      expect(screen.getByLabelText('Job Name')).toBeInTheDocument();
    });
  });

  it('closes CreateScheduleModal when Cancel is clicked', async () => {
    render(<SchedulesPanel sidebarOpen={false} onToggleSidebar={vi.fn()} addMessage={vi.fn()} platforms={platforms} />);
    fireEvent.click(screen.getByRole('button', { name: /create schedule/i }));
    await waitFor(() => screen.getByLabelText('Job Name'));
    fireEvent.click(screen.getByRole('button', { name: /^cancel$/i }));
    await waitFor(() => {
      expect(screen.queryByLabelText('Job Name')).not.toBeInTheDocument();
    });
  });
});

describe('SchedulesPanel – delete and pause', () => {
  beforeEach(() => {
    (wailsMocks.GetSchedules as ReturnType<typeof vi.fn>).mockResolvedValue([activeJob]);
  });

  it('calls DeleteSchedule and refreshes when Delete is confirmed', async () => {
    const addMessage = vi.fn();
    render(<SchedulesPanel sidebarOpen={false} onToggleSidebar={vi.fn()} addMessage={addMessage} platforms={platforms} />);
    await waitFor(() => screen.getByText('Daily Upload'));
    fireEvent.click(screen.getByRole('button', { name: /daily upload/i }));
    fireEvent.click(screen.getByRole('button', { name: /^delete$/i }));
    await waitFor(() => {
      expect(wailsMocks.DeleteSchedule).toHaveBeenCalledWith('j1');
      expect(addMessage).toHaveBeenCalledWith('info', 'Schedule deleted.');
    });
  });

  it('calls PauseSchedule when Pause is clicked', async () => {
    const addMessage = vi.fn();
    render(<SchedulesPanel sidebarOpen={false} onToggleSidebar={vi.fn()} addMessage={addMessage} platforms={platforms} />);
    await waitFor(() => screen.getByText('Daily Upload'));
    fireEvent.click(screen.getByRole('button', { name: /daily upload/i }));
    fireEvent.click(screen.getByRole('button', { name: /^pause$/i }));
    await waitFor(() => {
      expect(wailsMocks.PauseSchedule).toHaveBeenCalledWith('j1');
    });
  });
});

import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ScheduleDetailModal } from './ScheduleDetailModal';
import type { ScheduleResponse } from '../../types';

const baseJob: ScheduleResponse = {
  id: 'job-1',
  name: 'My Upload',
  type: 'upload',
  platforms: ['instagram', 'facebook'],
  cronExpr: '0 */12 * * *',
  nextRun: '2024-06-01T12:00:00Z',
  lastRun: '2024-05-31T12:00:00Z',
  status: 'active',
  createdAt: '2024-01-01T00:00:00Z',
  runCount: 5,
  lastResult: 'success',
  auto: false,
};

describe('ScheduleDetailModal – rendering', () => {
  it('renders job name as heading', () => {
    render(<ScheduleDetailModal job={baseJob} onClose={vi.fn()} onDelete={vi.fn()} onPauseResume={vi.fn()} />);
    expect(screen.getByText('My Upload')).toBeInTheDocument();
  });

  it('shows status text', () => {
    render(<ScheduleDetailModal job={baseJob} onClose={vi.fn()} onDelete={vi.fn()} onPauseResume={vi.fn()} />);
    expect(screen.getByText('active')).toBeInTheDocument();
  });

  it('shows type as Content Upload for upload type', () => {
    render(<ScheduleDetailModal job={baseJob} onClose={vi.fn()} onDelete={vi.fn()} onPauseResume={vi.fn()} />);
    expect(screen.getByText('Content Upload')).toBeInTheDocument();
  });

  it('shows type as Session Keep-Alive for keepalive type', () => {
    const job = { ...baseJob, type: 'session_keepalive' };
    render(<ScheduleDetailModal job={job} onClose={vi.fn()} onDelete={vi.fn()} onPauseResume={vi.fn()} />);
    expect(screen.getByText('Session Keep-Alive')).toBeInTheDocument();
  });

  it('shows (auto-created) suffix for auto jobs', () => {
    const job = { ...baseJob, auto: true };
    render(<ScheduleDetailModal job={job} onClose={vi.fn()} onDelete={vi.fn()} onPauseResume={vi.fn()} />);
    expect(screen.getByText(/auto-created/)).toBeInTheDocument();
  });

  it('shows platforms joined by comma', () => {
    render(<ScheduleDetailModal job={baseJob} onClose={vi.fn()} onDelete={vi.fn()} onPauseResume={vi.fn()} />);
    expect(screen.getByText('instagram, facebook')).toBeInTheDocument();
  });

  it('shows cron expression', () => {
    render(<ScheduleDetailModal job={baseJob} onClose={vi.fn()} onDelete={vi.fn()} onPauseResume={vi.fn()} />);
    expect(screen.getByText('0 */12 * * *')).toBeInTheDocument();
  });

  it('shows run count', () => {
    render(<ScheduleDetailModal job={baseJob} onClose={vi.fn()} onDelete={vi.fn()} onPauseResume={vi.fn()} />);
    expect(screen.getByText('5')).toBeInTheDocument();
  });

  it('shows last result when present', () => {
    render(<ScheduleDetailModal job={baseJob} onClose={vi.fn()} onDelete={vi.fn()} onPauseResume={vi.fn()} />);
    expect(screen.getByText('success')).toBeInTheDocument();
  });

  it('hides last result when not present', () => {
    const job = { ...baseJob, lastResult: '' };
    render(<ScheduleDetailModal job={job} onClose={vi.fn()} onDelete={vi.fn()} onPauseResume={vi.fn()} />);
    expect(screen.queryByText('Last Result')).not.toBeInTheDocument();
  });

  it('shows caption when present', () => {
    const job = { ...baseJob, caption: 'Hello world' };
    render(<ScheduleDetailModal job={job} onClose={vi.fn()} onDelete={vi.fn()} onPauseResume={vi.fn()} />);
    expect(screen.getByText('Hello world')).toBeInTheDocument();
  });

  it('shows hashtags when present', () => {
    const job = { ...baseJob, hashtags: ['#one', '#two'] };
    render(<ScheduleDetailModal job={job} onClose={vi.fn()} onDelete={vi.fn()} onPauseResume={vi.fn()} />);
    expect(screen.getByText('#one #two')).toBeInTheDocument();
  });

  it('shows Pause button when status is active', () => {
    render(<ScheduleDetailModal job={baseJob} onClose={vi.fn()} onDelete={vi.fn()} onPauseResume={vi.fn()} />);
    expect(screen.getByRole('button', { name: /pause/i })).toBeInTheDocument();
  });

  it('shows Resume button when status is paused', () => {
    const job = { ...baseJob, status: 'paused' };
    render(<ScheduleDetailModal job={job} onClose={vi.fn()} onDelete={vi.fn()} onPauseResume={vi.fn()} />);
    expect(screen.getByRole('button', { name: /resume/i })).toBeInTheDocument();
  });
});

describe('ScheduleDetailModal – interactions', () => {
  it('calls onClose when Close button is clicked', () => {
    const onClose = vi.fn();
    render(<ScheduleDetailModal job={baseJob} onClose={onClose} onDelete={vi.fn()} onPauseResume={vi.fn()} />);
    // 3 Close buttons: backdrop (0), × (1), footer "Close" (2)
    const closeBtns = screen.getAllByRole('button', { name: /^close$/i });
    fireEvent.click(closeBtns[2]);
    expect(onClose).toHaveBeenCalled();
  });

  it('calls onDelete with job id when Delete is clicked', () => {
    const onDelete = vi.fn();
    render(<ScheduleDetailModal job={baseJob} onClose={vi.fn()} onDelete={onDelete} onPauseResume={vi.fn()} />);
    fireEvent.click(screen.getByRole('button', { name: /delete/i }));
    expect(onDelete).toHaveBeenCalledWith('job-1');
  });

  it('calls onPauseResume with job when Pause is clicked', () => {
    const onPauseResume = vi.fn();
    render(<ScheduleDetailModal job={baseJob} onClose={vi.fn()} onDelete={vi.fn()} onPauseResume={onPauseResume} />);
    fireEvent.click(screen.getByRole('button', { name: /pause/i }));
    expect(onPauseResume).toHaveBeenCalledWith(baseJob);
  });

  it('calls onPauseResume with job when Resume is clicked', () => {
    const onPauseResume = vi.fn();
    const job = { ...baseJob, status: 'paused' };
    render(<ScheduleDetailModal job={job} onClose={vi.fn()} onDelete={vi.fn()} onPauseResume={onPauseResume} />);
    fireEvent.click(screen.getByRole('button', { name: /resume/i }));
    expect(onPauseResume).toHaveBeenCalledWith(job);
  });

  it('calls onClose when backdrop button is clicked', () => {
    const onClose = vi.fn();
    const { container } = render(<ScheduleDetailModal job={baseJob} onClose={onClose} onDelete={vi.fn()} onPauseResume={vi.fn()} />);
    const backdrop = container.querySelector('.modal-backdrop') as HTMLElement;
    fireEvent.click(backdrop);
    expect(onClose).toHaveBeenCalled();
  });
});

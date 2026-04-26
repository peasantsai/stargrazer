import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CreateScheduleModal } from './CreateScheduleModal';
import { CreateSchedule } from '../../test/wailsMock';
import type { PlatformResponse } from '../../types';

const platforms: PlatformResponse[] = [
  { id: 'instagram', name: 'Instagram', url: 'https://instagram.com', loggedIn: true, username: 'user', lastLogin: '', lastCheck: '', sessionDir: '' },
  { id: 'facebook', name: 'Facebook', url: 'https://facebook.com', loggedIn: false, username: '', lastLogin: '', lastCheck: '', sessionDir: '' },
];

beforeEach(() => {
  vi.clearAllMocks();
});

describe('CreateScheduleModal – rendering', () => {
  it('renders Create Schedule heading', () => {
    render(<CreateScheduleModal platforms={platforms} onCreated={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByText('Create Schedule')).toBeInTheDocument();
  });

  it('renders Job Name input', () => {
    render(<CreateScheduleModal platforms={platforms} onCreated={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByLabelText('Job Name')).toBeInTheDocument();
  });

  it('renders interval buttons', () => {
    render(<CreateScheduleModal platforms={platforms} onCreated={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByRole('button', { name: /every 6h/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /every 12h/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /every 24h/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /every 3d/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /custom/i })).toBeInTheDocument();
  });

  it('renders platform checkboxes', () => {
    render(<CreateScheduleModal platforms={platforms} onCreated={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByText('Instagram')).toBeInTheDocument();
    expect(screen.getByText('Facebook')).toBeInTheDocument();
  });

  it('renders Caption and Hashtags fields when type is upload', () => {
    render(<CreateScheduleModal platforms={platforms} onCreated={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByLabelText('Caption')).toBeInTheDocument();
    expect(screen.getByLabelText('Hashtags')).toBeInTheDocument();
  });

  it('hides Caption and Hashtags when type is Keep Alive', () => {
    render(<CreateScheduleModal platforms={platforms} onCreated={vi.fn()} onCancel={vi.fn()} />);
    fireEvent.click(screen.getByRole('button', { name: /keep alive/i }));
    expect(screen.queryByLabelText('Caption')).not.toBeInTheDocument();
    expect(screen.queryByLabelText('Hashtags')).not.toBeInTheDocument();
  });

  it('Create button is disabled when name is empty', () => {
    render(<CreateScheduleModal platforms={platforms} onCreated={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByRole('button', { name: /^create$/i })).toBeDisabled();
  });

  it('Create button is disabled when no platforms selected', () => {
    render(<CreateScheduleModal platforms={platforms} onCreated={vi.fn()} onCancel={vi.fn()} />);
    fireEvent.change(screen.getByLabelText('Job Name'), { target: { value: 'My Job' } });
    expect(screen.getByRole('button', { name: /^create$/i })).toBeDisabled();
  });
});

describe('CreateScheduleModal – interval selection', () => {
  it('shows custom cron input when Custom is selected', () => {
    render(<CreateScheduleModal platforms={platforms} onCreated={vi.fn()} onCancel={vi.fn()} />);
    fireEvent.click(screen.getByRole('button', { name: /custom/i }));
    expect(screen.getByLabelText('Custom cron expression')).toBeInTheDocument();
  });

  it('hides custom cron input when a preset is selected', () => {
    render(<CreateScheduleModal platforms={platforms} onCreated={vi.fn()} onCancel={vi.fn()} />);
    fireEvent.click(screen.getByRole('button', { name: /custom/i }));
    fireEvent.click(screen.getByRole('button', { name: /every 6h/i }));
    expect(screen.queryByLabelText('Custom cron expression')).not.toBeInTheDocument();
  });
});

describe('CreateScheduleModal – interactions', () => {
  it('calls onCancel when Cancel is clicked', () => {
    const onCancel = vi.fn();
    render(<CreateScheduleModal platforms={platforms} onCreated={vi.fn()} onCancel={onCancel} />);
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
    expect(onCancel).toHaveBeenCalled();
  });

  it('calls onCancel when close (×) button is clicked', () => {
    const onCancel = vi.fn();
    render(<CreateScheduleModal platforms={platforms} onCreated={vi.fn()} onCancel={onCancel} />);
    fireEvent.click(screen.getByRole('button', { name: /^close$/i }));
    expect(onCancel).toHaveBeenCalled();
  });

  it('enables Create button when name and platform are filled', () => {
    render(<CreateScheduleModal platforms={platforms} onCreated={vi.fn()} onCancel={vi.fn()} />);
    fireEvent.change(screen.getByLabelText('Job Name'), { target: { value: 'Test Job' } });
    const checkboxes = screen.getAllByRole('checkbox');
    fireEvent.click(checkboxes[0]);
    expect(screen.getByRole('button', { name: /^create$/i })).toBeEnabled();
  });

  it('calls CreateSchedule and onCreated when Create is clicked with valid data', async () => {
    const onCreated = vi.fn();
    render(<CreateScheduleModal platforms={platforms} onCreated={onCreated} onCancel={vi.fn()} />);
    fireEvent.change(screen.getByLabelText('Job Name'), { target: { value: 'Test Job' } });
    const checkboxes = screen.getAllByRole('checkbox');
    fireEvent.click(checkboxes[0]);
    fireEvent.click(screen.getByRole('button', { name: /^create$/i }));
    await waitFor(() => {
      expect(CreateSchedule).toHaveBeenCalled();
      expect(onCreated).toHaveBeenCalled();
    });
  });

  it('prepends # to hashtags without prefix', async () => {
    const onCreated = vi.fn();
    render(<CreateScheduleModal platforms={platforms} onCreated={onCreated} onCancel={vi.fn()} />);
    fireEvent.change(screen.getByLabelText('Job Name'), { target: { value: 'My Job' } });
    const checkboxes = screen.getAllByRole('checkbox');
    fireEvent.click(checkboxes[0]);
    fireEvent.change(screen.getByLabelText('Hashtags'), { target: { value: 'foo bar' } });
    fireEvent.click(screen.getByRole('button', { name: /^create$/i }));
    await waitFor(() => {
      expect(CreateSchedule).toHaveBeenCalledWith(
        expect.objectContaining({ hashtags: ['#foo', '#bar'] })
      );
    });
  });

  it('does not duplicate # prefix on hashtags that already have it', async () => {
    render(<CreateScheduleModal platforms={platforms} onCreated={vi.fn()} onCancel={vi.fn()} />);
    fireEvent.change(screen.getByLabelText('Job Name'), { target: { value: 'My Job' } });
    const checkboxes = screen.getAllByRole('checkbox');
    fireEvent.click(checkboxes[0]);
    fireEvent.change(screen.getByLabelText('Hashtags'), { target: { value: '#foo #bar' } });
    fireEvent.click(screen.getByRole('button', { name: /^create$/i }));
    await waitFor(() => {
      expect(CreateSchedule).toHaveBeenCalledWith(
        expect.objectContaining({ hashtags: ['#foo', '#bar'] })
      );
    });
  });
});

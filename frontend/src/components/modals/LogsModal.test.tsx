import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { LogsModal } from './LogsModal';
import { GetLogs, ExportLogs, ClearLogs } from '../../test/wailsMock';

// jsdom does not implement scrollIntoView
Element.prototype.scrollIntoView = vi.fn();

beforeEach(() => {
  vi.clearAllMocks();
  // Restore mocks cleared by clearAllMocks
  (GetLogs as ReturnType<typeof vi.fn>).mockResolvedValue([
    { timestamp: '2024-01-01T00:00:00Z', level: 'info', source: 'test', message: 'Test log' },
  ]);
  (ExportLogs as ReturnType<typeof vi.fn>).mockResolvedValue('[]');
  (ClearLogs as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);
});

describe('LogsModal – rendering', () => {
  it('renders the Application Logs heading', async () => {
    render(<LogsModal onClose={vi.fn()} />);
    expect(screen.getByText('Application Logs')).toBeInTheDocument();
  });

  it('shows the filter input', async () => {
    render(<LogsModal onClose={vi.fn()} />);
    expect(screen.getByPlaceholderText('Filter logs...')).toBeInTheDocument();
  });

  it('renders Export JSON button', () => {
    render(<LogsModal onClose={vi.fn()} />);
    expect(screen.getByRole('button', { name: /export json/i })).toBeInTheDocument();
  });

  it('renders Clear button', () => {
    render(<LogsModal onClose={vi.fn()} />);
    expect(screen.getByRole('button', { name: /^clear$/i })).toBeInTheDocument();
  });

  it('fetches logs on mount', async () => {
    render(<LogsModal onClose={vi.fn()} />);
    await waitFor(() => {
      expect(GetLogs).toHaveBeenCalled();
    });
  });

  it('displays log entries returned from GetLogs', async () => {
    render(<LogsModal onClose={vi.fn()} />);
    await waitFor(() => {
      expect(screen.getByText('Test log')).toBeInTheDocument();
    });
  });

  it('shows log level in uppercase', async () => {
    render(<LogsModal onClose={vi.fn()} />);
    await waitFor(() => {
      expect(screen.getByText('INFO')).toBeInTheDocument();
    });
  });

  it('shows log source in brackets', async () => {
    render(<LogsModal onClose={vi.fn()} />);
    await waitFor(() => {
      expect(screen.getByText('[test]')).toBeInTheDocument();
    });
  });
});

describe('LogsModal – filtering', () => {
  it('filters logs by message text', async () => {
    render(<LogsModal onClose={vi.fn()} />);
    await waitFor(() => screen.getByText('Test log'));
    fireEvent.change(screen.getByPlaceholderText('Filter logs...'), { target: { value: 'nonexistent' } });
    expect(screen.queryByText('Test log')).not.toBeInTheDocument();
  });

  it('shows log when filter matches message', async () => {
    render(<LogsModal onClose={vi.fn()} />);
    await waitFor(() => screen.getByText('Test log'));
    fireEvent.change(screen.getByPlaceholderText('Filter logs...'), { target: { value: 'Test' } });
    expect(screen.getByText('Test log')).toBeInTheDocument();
  });

  it('shows log when filter matches level', async () => {
    render(<LogsModal onClose={vi.fn()} />);
    await waitFor(() => screen.getByText('Test log'));
    fireEvent.change(screen.getByPlaceholderText('Filter logs...'), { target: { value: 'info' } });
    expect(screen.getByText('Test log')).toBeInTheDocument();
  });

  it('shows log when filter matches source', async () => {
    render(<LogsModal onClose={vi.fn()} />);
    await waitFor(() => screen.getByText('Test log'));
    fireEvent.change(screen.getByPlaceholderText('Filter logs...'), { target: { value: 'test' } });
    expect(screen.getByText('Test log')).toBeInTheDocument();
  });
});

describe('LogsModal – interactions', () => {
  it('calls onClose when Close (×) button is clicked', () => {
    const onClose = vi.fn();
    render(<LogsModal onClose={onClose} />);
    // Two close buttons: backdrop (0) and × button (1)
    const closeBtns = screen.getAllByRole('button', { name: /^close$/i });
    fireEvent.click(closeBtns[1]);
    expect(onClose).toHaveBeenCalled();
  });

  it('calls onClose when backdrop button is clicked', () => {
    const onClose = vi.fn();
    const { container } = render(<LogsModal onClose={onClose} />);
    const backdrop = container.querySelector('.modal-backdrop') as HTMLElement;
    fireEvent.click(backdrop);
    expect(onClose).toHaveBeenCalled();
  });

  it('calls ClearLogs and clears displayed logs when Clear is clicked', async () => {
    render(<LogsModal onClose={vi.fn()} />);
    await waitFor(() => screen.getByText('Test log'));
    fireEvent.click(screen.getByRole('button', { name: /^clear$/i }));
    await waitFor(() => {
      expect(ClearLogs).toHaveBeenCalled();
      expect(screen.queryByText('Test log')).not.toBeInTheDocument();
    });
  });

  it('calls ExportLogs when Export JSON is clicked', async () => {
    // Mock URL.createObjectURL / revokeObjectURL for jsdom
    const createObjectURL = vi.fn().mockReturnValue('blob:mock');
    const revokeObjectURL = vi.fn();
    vi.stubGlobal('URL', { createObjectURL, revokeObjectURL });

    render(<LogsModal onClose={vi.fn()} />);
    fireEvent.click(screen.getByRole('button', { name: /export json/i }));
    await waitFor(() => {
      expect(ExportLogs).toHaveBeenCalled();
    });
  });
});

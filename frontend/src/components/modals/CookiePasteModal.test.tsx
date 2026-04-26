import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { CookiePasteModal } from './CookiePasteModal';
import type { PlatformResponse } from '../../types';

const instagramPlatform: PlatformResponse = {
  id: 'instagram',
  name: 'Instagram',
  url: 'https://www.instagram.com',
  loggedIn: false,
  username: '',
  lastLogin: '',
  lastCheck: '',
  sessionDir: '/tmp/sessions',
};

describe('CookiePasteModal – rendering', () => {
  it('renders the heading with platform name', () => {
    render(<CookiePasteModal platform={instagramPlatform} onImport={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByText('Import Instagram Cookies')).toBeInTheDocument();
  });

  it('renders the textarea for cookie input', () => {
    render(<CookiePasteModal platform={instagramPlatform} onImport={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByRole('textbox')).toBeInTheDocument();
  });

  it('renders Import Cookies button as disabled when textarea is empty', () => {
    render(<CookiePasteModal platform={instagramPlatform} onImport={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByRole('button', { name: /import cookies/i })).toBeDisabled();
  });

  it('renders Cancel button', () => {
    render(<CookiePasteModal platform={instagramPlatform} onImport={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
  });

  it('shows instructions steps', () => {
    render(<CookiePasteModal platform={instagramPlatform} onImport={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByText(/Open Instagram in the browser/)).toBeInTheDocument();
  });
});

describe('CookiePasteModal – cookie detection', () => {
  it('enables Import button when valid cookie line is pasted', () => {
    render(<CookiePasteModal platform={instagramPlatform} onImport={vi.fn()} onCancel={vi.fn()} />);
    fireEvent.change(screen.getByRole('textbox'), {
      target: { value: '.instagram.com\tTRUE\t/\tTRUE\t0\tsession_id\tabc123' },
    });
    expect(screen.getByRole('button', { name: /import cookies/i })).toBeEnabled();
  });

  it('shows cookie count when cookies are pasted', () => {
    render(<CookiePasteModal platform={instagramPlatform} onImport={vi.fn()} onCancel={vi.fn()} />);
    fireEvent.change(screen.getByRole('textbox'), {
      target: { value: '.instagram.com\tTRUE\t/\tTRUE\t0\tsession_id\tabc123\n.instagram.com\tTRUE\t/\tTRUE\t0\tcsrf_token\txyz789' },
    });
    expect(screen.getByText(/2 cookies detected/)).toBeInTheDocument();
  });

  it('shows singular "cookie" for exactly one cookie', () => {
    render(<CookiePasteModal platform={instagramPlatform} onImport={vi.fn()} onCancel={vi.fn()} />);
    fireEvent.change(screen.getByRole('textbox'), {
      target: { value: '.instagram.com\tTRUE\t/\tTRUE\t0\tsession_id\tabc123' },
    });
    expect(screen.getByText(/1 cookie detected/)).toBeInTheDocument();
  });

  it('ignores comment lines starting with #', () => {
    render(<CookiePasteModal platform={instagramPlatform} onImport={vi.fn()} onCancel={vi.fn()} />);
    fireEvent.change(screen.getByRole('textbox'), {
      target: { value: '# Netscape HTTP Cookie File\n.instagram.com\tTRUE\t/\tTRUE\t0\tsession_id\tabc123' },
    });
    expect(screen.getByText(/1 cookie detected/)).toBeInTheDocument();
  });

  it('keeps Import button disabled for comment-only input', () => {
    render(<CookiePasteModal platform={instagramPlatform} onImport={vi.fn()} onCancel={vi.fn()} />);
    fireEvent.change(screen.getByRole('textbox'), {
      target: { value: '# Netscape HTTP Cookie File\n# comment line' },
    });
    expect(screen.getByRole('button', { name: /import cookies/i })).toBeDisabled();
  });
});

describe('CookiePasteModal – interactions', () => {
  it('calls onCancel when Cancel is clicked', () => {
    const onCancel = vi.fn();
    render(<CookiePasteModal platform={instagramPlatform} onImport={vi.fn()} onCancel={onCancel} />);
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
    expect(onCancel).toHaveBeenCalledTimes(1);
  });

  it('calls onCancel when modal close button is clicked', () => {
    const onCancel = vi.fn();
    render(<CookiePasteModal platform={instagramPlatform} onImport={vi.fn()} onCancel={onCancel} />);
    fireEvent.click(screen.getByRole('button', { name: /^close$/i }));
    expect(onCancel).toHaveBeenCalledTimes(1);
  });

  it('calls onImport with textarea text when Import is clicked', () => {
    const onImport = vi.fn();
    const cookieText = '.instagram.com\tTRUE\t/\tTRUE\t0\tsession_id\tabc123';
    render(<CookiePasteModal platform={instagramPlatform} onImport={onImport} onCancel={vi.fn()} />);
    fireEvent.change(screen.getByRole('textbox'), { target: { value: cookieText } });
    fireEvent.click(screen.getByRole('button', { name: /import cookies/i }));
    expect(onImport).toHaveBeenCalledWith(cookieText);
  });
});

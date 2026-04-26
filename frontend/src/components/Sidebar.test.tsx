import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { Sidebar } from './Sidebar';
import type { AccountInfo } from '../types';

const defaultAccount: AccountInfo = { name: 'Alice Smith', email: 'alice@example.com', avatarUrl: '' };

const defaultProps = {
  view: 'chat' as const,
  setView: vi.fn(),
  browserStatus: 'stopped',
  open: true,
  onToggle: vi.fn(),
  theme: 'dark' as const,
  setTheme: vi.fn(),
  account: defaultAccount,
  updateAccount: vi.fn(),
};

describe('Sidebar – rendering', () => {
  it('renders nothing when open is false', () => {
    const { container } = render(<Sidebar {...defaultProps} open={false} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders the Stargrazer heading when open', () => {
    render(<Sidebar {...defaultProps} />);
    expect(screen.getByText('Stargrazer')).toBeInTheDocument();
  });

  it('renders Chat nav button', () => {
    render(<Sidebar {...defaultProps} />);
    expect(screen.getByRole('button', { name: /chat/i })).toBeInTheDocument();
  });

  it('renders Schedules nav button', () => {
    render(<Sidebar {...defaultProps} />);
    expect(screen.getByRole('button', { name: /schedules/i })).toBeInTheDocument();
  });

  it('renders Settings nav button', () => {
    render(<Sidebar {...defaultProps} />);
    expect(screen.getByRole('button', { name: /settings/i })).toBeInTheDocument();
  });

  it('marks Chat button as active when view is chat', () => {
    render(<Sidebar {...defaultProps} view="chat" />);
    expect(screen.getByRole('button', { name: /chat/i })).toHaveClass('active');
  });

  it('marks Schedules button as active when view is schedules', () => {
    render(<Sidebar {...defaultProps} view="schedules" />);
    expect(screen.getByRole('button', { name: /schedules/i })).toHaveClass('active');
  });

  it('renders all 6 platform buttons', () => {
    render(<Sidebar {...defaultProps} />);
    expect(screen.getByRole('button', { name: /facebook/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /instagram/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /tiktok/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /youtube/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /linkedin/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /^x$/i })).toBeInTheDocument();
  });

  it('renders account name', () => {
    render(<Sidebar {...defaultProps} />);
    expect(screen.getByText('Alice Smith')).toBeInTheDocument();
  });

  it('renders account email', () => {
    render(<Sidebar {...defaultProps} />);
    expect(screen.getByText('alice@example.com')).toBeInTheDocument();
  });

  it('renders initials when no avatarUrl', () => {
    render(<Sidebar {...defaultProps} />);
    // 'Alice Smith' → two words → initials 'AS'
    expect(screen.getByText('AS')).toBeInTheDocument();
  });

  it('renders avatar image when avatarUrl is set', () => {
    const account = { ...defaultAccount, avatarUrl: 'https://example.com/avatar.png' };
    const { container } = render(<Sidebar {...defaultProps} account={account} />);
    const img = container.querySelector('img');
    expect(img).not.toBeNull();
    expect(img).toHaveAttribute('src', 'https://example.com/avatar.png');
  });

  it('renders dark theme button as active when theme is dark', () => {
    render(<Sidebar {...defaultProps} theme="dark" />);
    expect(screen.getByRole('button', { name: /dark theme/i })).toHaveClass('active');
  });

  it('renders light theme button as active when theme is light', () => {
    render(<Sidebar {...defaultProps} theme="light" />);
    expect(screen.getByRole('button', { name: /light theme/i })).toHaveClass('active');
  });
});

describe('Sidebar – interactions', () => {
  it('calls setView("chat") when Chat is clicked', () => {
    const setView = vi.fn();
    render(<Sidebar {...defaultProps} setView={setView} />);
    fireEvent.click(screen.getByRole('button', { name: /chat/i }));
    expect(setView).toHaveBeenCalledWith('chat');
  });

  it('calls setView("schedules") when Schedules is clicked', () => {
    const setView = vi.fn();
    render(<Sidebar {...defaultProps} setView={setView} />);
    fireEvent.click(screen.getByRole('button', { name: /schedules/i }));
    expect(setView).toHaveBeenCalledWith('schedules');
  });

  it('calls setView("config") when Settings is clicked', () => {
    const setView = vi.fn();
    render(<Sidebar {...defaultProps} setView={setView} />);
    fireEvent.click(screen.getByRole('button', { name: /settings/i }));
    expect(setView).toHaveBeenCalledWith('config');
  });

  it('calls setView("platform:instagram") when Instagram nav is clicked', () => {
    const setView = vi.fn();
    render(<Sidebar {...defaultProps} setView={setView} />);
    fireEvent.click(screen.getByRole('button', { name: /instagram/i }));
    expect(setView).toHaveBeenCalledWith('platform:instagram');
  });

  it('calls onToggle when Close sidebar button is clicked', () => {
    const onToggle = vi.fn();
    render(<Sidebar {...defaultProps} onToggle={onToggle} />);
    fireEvent.click(screen.getByRole('button', { name: /close sidebar/i }));
    expect(onToggle).toHaveBeenCalled();
  });

  it('calls setTheme("light") when light theme button is clicked', () => {
    const setTheme = vi.fn();
    render(<Sidebar {...defaultProps} setTheme={setTheme} />);
    fireEvent.click(screen.getByRole('button', { name: /light theme/i }));
    expect(setTheme).toHaveBeenCalledWith('light');
  });

  it('calls setTheme("dark") when dark theme button is clicked', () => {
    const setTheme = vi.fn();
    render(<Sidebar {...defaultProps} theme="light" setTheme={setTheme} />);
    fireEvent.click(screen.getByRole('button', { name: /dark theme/i }));
    expect(setTheme).toHaveBeenCalledWith('dark');
  });

  it('opens AccountModal when account card is clicked', () => {
    render(<Sidebar {...defaultProps} />);
    fireEvent.click(screen.getByRole('button', { name: /alice smith/i }));
    expect(screen.getByText('Account Settings')).toBeInTheDocument();
  });

  it('closes AccountModal when onClose is triggered', () => {
    render(<Sidebar {...defaultProps} />);
    fireEvent.click(screen.getByRole('button', { name: /alice smith/i }));
    expect(screen.getByText('Account Settings')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
    expect(screen.queryByText('Account Settings')).not.toBeInTheDocument();
  });
});

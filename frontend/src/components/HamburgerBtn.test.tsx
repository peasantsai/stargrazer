import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { HamburgerBtn } from './HamburgerBtn';

describe('HamburgerBtn', () => {
  it('renders the button when sidebarOpen is false', () => {
    render(<HamburgerBtn sidebarOpen={false} onToggle={vi.fn()} />);
    expect(screen.getByRole('button', { name: /open sidebar/i })).toBeInTheDocument();
  });

  it('renders nothing when sidebarOpen is true', () => {
    const { container } = render(<HamburgerBtn sidebarOpen={true} onToggle={vi.fn()} />);
    expect(container.firstChild).toBeNull();
  });

  it('calls onToggle when clicked', () => {
    const onToggle = vi.fn();
    render(<HamburgerBtn sidebarOpen={false} onToggle={onToggle} />);
    fireEvent.click(screen.getByRole('button', { name: /open sidebar/i }));
    expect(onToggle).toHaveBeenCalledTimes(1);
  });

  it('has aria-label "Open sidebar"', () => {
    render(<HamburgerBtn sidebarOpen={false} onToggle={vi.fn()} />);
    expect(screen.getByRole('button')).toHaveAttribute('aria-label', 'Open sidebar');
  });

  it('has the sidebar-toggle-float class', () => {
    render(<HamburgerBtn sidebarOpen={false} onToggle={vi.fn()} />);
    expect(screen.getByRole('button')).toHaveClass('sidebar-toggle-float');
  });
});

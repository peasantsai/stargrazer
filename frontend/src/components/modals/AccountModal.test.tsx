import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { AccountModal } from './AccountModal';
import type { AccountInfo } from '../../types';

const defaultAccount: AccountInfo = { name: 'Alice', email: 'alice@example.com', avatarUrl: '' };

describe('AccountModal – rendering', () => {
  it('renders the modal heading', () => {
    render(<AccountModal account={defaultAccount} updateAccount={vi.fn()} onClose={vi.fn()} />);
    expect(screen.getByText('Account Settings')).toBeInTheDocument();
  });

  it('pre-fills name input with account.name', () => {
    render(<AccountModal account={defaultAccount} updateAccount={vi.fn()} onClose={vi.fn()} />);
    expect(screen.getByLabelText('Display Name')).toHaveValue('Alice');
  });

  it('pre-fills email input with account.email', () => {
    render(<AccountModal account={defaultAccount} updateAccount={vi.fn()} onClose={vi.fn()} />);
    expect(screen.getByLabelText('Email')).toHaveValue('alice@example.com');
  });

  it('pre-fills avatar URL input', () => {
    const account = { ...defaultAccount, avatarUrl: 'https://example.com/avatar.png' };
    render(<AccountModal account={account} updateAccount={vi.fn()} onClose={vi.fn()} />);
    expect(screen.getByLabelText('Avatar URL')).toHaveValue('https://example.com/avatar.png');
  });

  it('shows initials when no avatarUrl is set', () => {
    render(<AccountModal account={defaultAccount} updateAccount={vi.fn()} onClose={vi.fn()} />);
    // 'Alice' is one word → initial is 'A'
    expect(screen.getByText('A')).toBeInTheDocument();
  });

  it('shows "U" as fallback initial when name is empty', () => {
    render(<AccountModal account={{ name: '', email: '', avatarUrl: '' }} updateAccount={vi.fn()} onClose={vi.fn()} />);
    expect(screen.getByText('U')).toBeInTheDocument();
  });

  it('renders an img tag when avatarUrl is provided', () => {
    const account = { ...defaultAccount, avatarUrl: 'https://example.com/avatar.png' };
    const { container } = render(<AccountModal account={account} updateAccount={vi.fn()} onClose={vi.fn()} />);
    const img = container.querySelector('img');
    expect(img).not.toBeNull();
    expect(img).toHaveAttribute('src', 'https://example.com/avatar.png');
  });

  it('renders Save and Cancel buttons', () => {
    render(<AccountModal account={defaultAccount} updateAccount={vi.fn()} onClose={vi.fn()} />);
    expect(screen.getByRole('button', { name: /save/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
  });
});

describe('AccountModal – interactions', () => {
  it('calls onClose when Cancel is clicked', () => {
    const onClose = vi.fn();
    render(<AccountModal account={defaultAccount} updateAccount={vi.fn()} onClose={onClose} />);
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('calls onClose when close (×) button is clicked', () => {
    const onClose = vi.fn();
    render(<AccountModal account={defaultAccount} updateAccount={vi.fn()} onClose={onClose} />);
    // Two buttons share aria-label="Close": backdrop (index 0) and × button (index 1)
    const closeBtns = screen.getAllByRole('button', { name: /^close$/i });
    fireEvent.click(closeBtns[1]);
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('calls updateAccount with current values on Save', () => {
    const updateAccount = vi.fn();
    const onClose = vi.fn();
    render(<AccountModal account={defaultAccount} updateAccount={updateAccount} onClose={onClose} />);
    fireEvent.click(screen.getByRole('button', { name: /save/i }));
    expect(updateAccount).toHaveBeenCalledWith({
      name: 'Alice',
      email: 'alice@example.com',
      avatarUrl: '',
    });
  });

  it('calls onClose after Save', () => {
    const onClose = vi.fn();
    render(<AccountModal account={defaultAccount} updateAccount={vi.fn()} onClose={onClose} />);
    fireEvent.click(screen.getByRole('button', { name: /save/i }));
    expect(onClose).toHaveBeenCalled();
  });

  it('updates name field on user input', () => {
    render(<AccountModal account={defaultAccount} updateAccount={vi.fn()} onClose={vi.fn()} />);
    const input = screen.getByLabelText('Display Name');
    fireEvent.change(input, { target: { value: 'Bob' } });
    expect(input).toHaveValue('Bob');
  });

  it('updates email field on user input', () => {
    render(<AccountModal account={defaultAccount} updateAccount={vi.fn()} onClose={vi.fn()} />);
    const input = screen.getByLabelText('Email');
    fireEvent.change(input, { target: { value: 'bob@example.com' } });
    expect(input).toHaveValue('bob@example.com');
  });

  it('saves updated values when Save is clicked after editing', () => {
    const updateAccount = vi.fn();
    render(<AccountModal account={defaultAccount} updateAccount={updateAccount} onClose={vi.fn()} />);
    fireEvent.change(screen.getByLabelText('Display Name'), { target: { value: 'Charlie' } });
    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'charlie@example.com' } });
    fireEvent.click(screen.getByRole('button', { name: /save/i }));
    expect(updateAccount).toHaveBeenCalledWith({
      name: 'Charlie',
      email: 'charlie@example.com',
      avatarUrl: '',
    });
  });

  it('calls onClose when backdrop button is clicked', () => {
    const onClose = vi.fn();
    const { container } = render(<AccountModal account={defaultAccount} updateAccount={vi.fn()} onClose={onClose} />);
    const backdrop = container.querySelector('.modal-backdrop') as HTMLElement;
    fireEvent.click(backdrop);
    expect(onClose).toHaveBeenCalled();
  });
});

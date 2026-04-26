import { render, screen, waitFor, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi, describe, it, expect, beforeEach } from 'vitest';

// The wailsjs imports are aliased in vitest.config.ts to our mock files.
// Import the mocks so we can inspect/reset them in tests.
import * as wailsMocks from './test/wailsMock';

import App from './App';

// Mock scrollIntoView
Element.prototype.scrollIntoView = vi.fn();

// Mock URL.createObjectURL and URL.revokeObjectURL
window.URL.createObjectURL = vi.fn(() => 'blob:mock');
window.URL.revokeObjectURL = vi.fn();

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  // Re-apply default mock return values after clearAllMocks
  wailsMocks.GetBrowserStatus.mockResolvedValue({ status: 'stopped', error: '' });
  wailsMocks.GetPlatforms.mockResolvedValue([
    { id: 'instagram', name: 'Instagram', url: 'https://www.instagram.com', loggedIn: true, username: 'testuser', lastLogin: '2024-01-01T00:00:00Z', lastCheck: '2024-01-01T00:00:00Z', sessionDir: '/tmp/sessions' },
    { id: 'facebook', name: 'Facebook', url: 'https://www.facebook.com', loggedIn: false, username: '', lastLogin: '', lastCheck: '', sessionDir: '/tmp/sessions' },
    { id: 'tiktok', name: 'TikTok', url: 'https://www.tiktok.com', loggedIn: false, username: '', lastLogin: '', lastCheck: '', sessionDir: '/tmp/sessions' },
    { id: 'youtube', name: 'YouTube', url: 'https://www.youtube.com', loggedIn: false, username: '', lastLogin: '', lastCheck: '', sessionDir: '/tmp/sessions' },
    { id: 'linkedin', name: 'LinkedIn', url: 'https://www.linkedin.com', loggedIn: false, username: '', lastLogin: '', lastCheck: '', sessionDir: '/tmp/sessions' },
    { id: 'x', name: 'X', url: 'https://x.com', loggedIn: false, username: '', lastLogin: '', lastCheck: '', sessionDir: '/tmp/sessions' },
  ]);
  wailsMocks.StartBrowser.mockResolvedValue({ status: 'running', error: '' });
  wailsMocks.StopBrowser.mockResolvedValue({ status: 'stopped', error: '' });
  wailsMocks.GetBrowserConfig.mockResolvedValue({
    chromiumPath: '/usr/bin/chromium', cdpPort: 9222, headless: false,
    userDataDir: '/tmp/data', windowWidth: 1280, windowHeight: 900, extraFlags: ['--disable-infobars'],
  });
  wailsMocks.UpdateBrowserConfig.mockResolvedValue({
    chromiumPath: '/usr/bin/chromium', cdpPort: 9222, headless: false,
    userDataDir: '/tmp/data', windowWidth: 1280, windowHeight: 900, extraFlags: ['--disable-infobars'],
  });
  wailsMocks.ResetBrowserConfig.mockResolvedValue({
    chromiumPath: '', cdpPort: 9222, headless: false, userDataDir: '', windowWidth: 1280, windowHeight: 900, extraFlags: [],
  });
  wailsMocks.RestartBrowser.mockResolvedValue({ status: 'running', error: '' });
  wailsMocks.GetSchedules.mockResolvedValue([]);
  wailsMocks.SelectFile.mockResolvedValue('/tmp/photo.jpg');
  wailsMocks.TriggerUpload.mockResolvedValue({ success: true, message: 'Upload queued for 1 platform(s)' });
  wailsMocks.GetLogs.mockResolvedValue([{ timestamp: '2024-01-01T00:00:00Z', level: 'info', source: 'test', message: 'Test log' }]);
  wailsMocks.ExportLogs.mockResolvedValue('[]');
  wailsMocks.ClearLogs.mockResolvedValue(undefined);
  wailsMocks.CreateSchedule.mockResolvedValue({
    id: '1', name: 'Test', type: 'upload', platforms: ['instagram'], cronExpr: '0 */12 * * *',
    status: 'active', runCount: 0, auto: false, createdAt: '2024-01-01T00:00:00Z',
  });
  wailsMocks.DeleteSchedule.mockResolvedValue(true);
  wailsMocks.PauseSchedule.mockResolvedValue({ id: '1', status: 'paused', name: 'Test' });
  wailsMocks.ResumeSchedule.mockResolvedValue({ id: '1', status: 'active', name: 'Test' });
  wailsMocks.OpenPlatform.mockResolvedValue({ status: 'running', error: '' });
  wailsMocks.CheckAllLoginStatus.mockResolvedValue([]);
  wailsMocks.PurgeSession.mockResolvedValue({ id: 'instagram', name: 'Instagram', loggedIn: false, username: '', sessionDir: '/tmp' });
  wailsMocks.ImportCookies.mockResolvedValue({ id: 'instagram', name: 'Instagram', loggedIn: true, username: 'testuser' });
  wailsMocks.GetAutomations.mockResolvedValue([]);
  wailsMocks.SaveAutomation.mockResolvedValue({
    id: 'auto-1', platformId: 'instagram', name: 'Test Automation', description: '',
    steps: [], createdAt: '2024-01-01T00:00:00Z', lastRun: '', runCount: 0,
  });
  wailsMocks.DeleteAutomation.mockResolvedValue(true);
  wailsMocks.RunAutomation.mockResolvedValue({ success: true, message: 'Automation completed successfully' });
});

describe('App', () => {
  it('renders the app with Stargrazer title', async () => {
    await act(async () => { render(<App />); });
    expect(screen.getByText('Stargrazer')).toBeInTheDocument();
  });

  it('shows sidebar navigation with Chat, Schedules, Settings', async () => {
    await act(async () => { render(<App />); });
    expect(screen.getByText('Chat')).toBeInTheDocument();
    expect(screen.getByText('Schedules')).toBeInTheDocument();
    expect(screen.getByText('Settings')).toBeInTheDocument();
  });

  it('starts in Chat view with welcome message', async () => {
    await act(async () => { render(<App />); });
    expect(screen.getByText('Welcome to Stargrazer')).toBeInTheDocument();
  });

  it('shows browser status as stopped initially', async () => {
    await act(async () => { render(<App />); });
    expect(screen.getByText(/Browser: stopped/i)).toBeInTheDocument();
  });

  it('shows Start Browser button when stopped', async () => {
    await act(async () => { render(<App />); });
    expect(screen.getByText('Start Browser')).toBeInTheDocument();
  });

  it('calls StartBrowser when Start is clicked', async () => {
    const user = userEvent.setup();
    await act(async () => { render(<App />); });
    await user.click(screen.getByText('Start Browser'));
    await waitFor(() => {
      expect(wailsMocks.StartBrowser).toHaveBeenCalled();
    });
  });

  it('switches to Schedules view', async () => {
    const user = userEvent.setup();
    await act(async () => { render(<App />); });
    await user.click(screen.getByText('Schedules'));
    await waitFor(() => {
      expect(screen.getByText('No Scheduled Jobs')).toBeInTheDocument();
    });
  });

  it('switches to Settings view', async () => {
    const user = userEvent.setup();
    await act(async () => { render(<App />); });
    await user.click(screen.getByText('Settings'));
    await waitFor(() => {
      expect(wailsMocks.GetBrowserConfig).toHaveBeenCalled();
    });
  });

  it('shows theme toggle in sidebar', async () => {
    await act(async () => { render(<App />); });
    expect(screen.getByText('Theme')).toBeInTheDocument();
  });

  it('displays platform chips in chat', async () => {
    await act(async () => { render(<App />); });
    await waitFor(() => {
      expect(screen.getByText('Instagram')).toBeInTheDocument();
    });
  });

  it('shows upload form elements', async () => {
    await act(async () => { render(<App />); });
    await waitFor(() => {
      expect(screen.getByText('Attach file')).toBeInTheDocument();
      expect(screen.getByPlaceholderText('Write your caption...')).toBeInTheDocument();
      expect(screen.getByPlaceholderText('#hashtags...')).toBeInTheDocument();
    });
  });

  it('shows Send button disabled when no platform selected', async () => {
    await act(async () => { render(<App />); });
    await waitFor(() => {
      expect(screen.getByText('Send')).toBeDisabled();
    });
  });

  it('calls SelectFile when attach button clicked', async () => {
    const user = userEvent.setup();
    await act(async () => { render(<App />); });
    await waitFor(() => screen.getByText('Attach file'));
    await user.click(screen.getByText('Attach file'));
    expect(wailsMocks.SelectFile).toHaveBeenCalled();
  });

  it('shows caption input and allows typing', async () => {
    const user = userEvent.setup();
    await act(async () => { render(<App />); });
    const textarea = await waitFor(() => screen.getByPlaceholderText('Write your caption...'));
    await user.type(textarea, 'Hello world');
    expect(textarea).toHaveValue('Hello world');
  });

  it('can add hashtags with space key', async () => {
    const user = userEvent.setup();
    await act(async () => { render(<App />); });
    const tagInput = await waitFor(() => screen.getByPlaceholderText('#hashtags...'));
    await user.type(tagInput, 'test ');
    await waitFor(() => {
      expect(screen.getByText('#test')).toBeInTheDocument();
    });
  });

  it('sets dark theme by default', async () => {
    await act(async () => { render(<App />); });
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  });

  it('shows user account info in sidebar', async () => {
    await act(async () => { render(<App />); });
    expect(screen.getByText('User')).toBeInTheDocument();
  });

  it('loads platforms on mount', async () => {
    await act(async () => { render(<App />); });
    await waitFor(() => {
      expect(wailsMocks.GetPlatforms).toHaveBeenCalled();
    });
  });

  it('loads browser status on mount', async () => {
    await act(async () => { render(<App />); });
    expect(wailsMocks.GetBrowserStatus).toHaveBeenCalled();
  });
});

describe('Schedules Panel', () => {
  it('shows create schedule button', async () => {
    const user = userEvent.setup();
    await act(async () => { render(<App />); });
    await user.click(screen.getByText('Schedules'));
    await waitFor(() => {
      expect(screen.getByText('Create Schedule')).toBeInTheDocument();
    });
  });

  it('opens create schedule modal', async () => {
    const user = userEvent.setup();
    await act(async () => { render(<App />); });
    await user.click(screen.getByText('Schedules'));
    await waitFor(() => screen.getByText('Create Schedule'));
    await user.click(screen.getByText('Create Schedule'));
    await waitFor(() => {
      expect(screen.getByText('Job Name')).toBeInTheDocument();
    });
  });

  it('shows schedule list when schedules exist', async () => {
    wailsMocks.GetSchedules.mockResolvedValueOnce([{
      id: '1', name: 'Instagram Keep-Alive', type: 'session_keepalive',
      platforms: ['instagram'], cronExpr: '0 */12 * * *',
      nextRun: '2024-01-02T00:00:00Z', lastRun: '',
      status: 'active', createdAt: '2024-01-01T00:00:00Z',
      runCount: 5, lastResult: '', auto: true,
      hashtags: [], filePath: '', caption: '',
    }]);
    const user = userEvent.setup();
    await act(async () => { render(<App />); });
    await user.click(screen.getByText('Schedules'));
    await waitFor(() => {
      expect(screen.getByText('Instagram Keep-Alive')).toBeInTheDocument();
    });
  });
});

describe('Settings Panel', () => {
  it('loads and displays browser config', async () => {
    const user = userEvent.setup();
    await act(async () => { render(<App />); });
    await user.click(screen.getByText('Settings'));
    await waitFor(() => {
      expect(screen.getByText('Connection')).toBeInTheDocument();
    });
  });

  it('shows social media connections section', async () => {
    const user = userEvent.setup();
    await act(async () => { render(<App />); });
    await user.click(screen.getByText('Settings'));
    await waitFor(() => {
      expect(screen.getByText('Social Media Connections')).toBeInTheDocument();
    });
  });

  it('shows search input for filtering settings', async () => {
    const user = userEvent.setup();
    await act(async () => { render(<App />); });
    await user.click(screen.getByText('Settings'));
    await waitFor(() => {
      expect(screen.getByPlaceholderText('Search settings...')).toBeInTheDocument();
    });
  });

  it('shows save and reset buttons', async () => {
    const user = userEvent.setup();
    await act(async () => { render(<App />); });
    await user.click(screen.getByText('Settings'));
    await waitFor(() => {
      expect(screen.getByText('Save Settings')).toBeInTheDocument();
      expect(screen.getByText('Reset to Defaults')).toBeInTheDocument();
    });
  });

  it('calls UpdateBrowserConfig on save', async () => {
    const user = userEvent.setup();
    await act(async () => { render(<App />); });
    await user.click(screen.getByText('Settings'));
    await waitFor(() => screen.getByText('Save Settings'));
    await user.click(screen.getByText('Save Settings'));
    await waitFor(() => {
      expect(wailsMocks.UpdateBrowserConfig).toHaveBeenCalled();
    });
  });

  it('calls ResetBrowserConfig on reset', async () => {
    const user = userEvent.setup();
    await act(async () => { render(<App />); });
    await user.click(screen.getByText('Settings'));
    await waitFor(() => screen.getByText('Reset to Defaults'));
    await user.click(screen.getByText('Reset to Defaults'));
    await waitFor(() => {
      expect(wailsMocks.ResetBrowserConfig).toHaveBeenCalled();
    });
  });

  it('can filter settings by search', async () => {
    const user = userEvent.setup();
    await act(async () => { render(<App />); });
    await user.click(screen.getByText('Settings'));
    const searchInput = await waitFor(() => screen.getByPlaceholderText('Search settings...'));
    await user.type(searchInput, 'display');
    await waitFor(() => {
      expect(screen.getByText('Display')).toBeInTheDocument();
    });
  });

  it('shows no results message for unmatched search', async () => {
    const user = userEvent.setup();
    await act(async () => { render(<App />); });
    await user.click(screen.getByText('Settings'));
    const searchInput = await waitFor(() => screen.getByPlaceholderText('Search settings...'));
    await user.type(searchInput, 'zzzznonexistent');
    await waitFor(() => {
      expect(screen.getByText(/No settings match/)).toBeInTheDocument();
    });
  });
});

describe('Account Modal', () => {
  it('opens when clicking account card', async () => {
    const user = userEvent.setup();
    await act(async () => { render(<App />); });
    const accountCard = screen.getByText('User').closest('button');
    if (accountCard) {
      await user.click(accountCard);
      await waitFor(() => {
        expect(screen.getByText('Account Settings')).toBeInTheDocument();
      });
    }
  });
});

describe('Error handling', () => {
  it('shows error when StartBrowser fails', async () => {
    wailsMocks.StartBrowser.mockResolvedValueOnce({ status: 'error', error: 'Chromium not found' });
    const user = userEvent.setup();
    await act(async () => { render(<App />); });
    await user.click(screen.getByText('Start Browser'));
    await waitFor(() => {
      expect(screen.getByText(/Chromium not found/)).toBeInTheDocument();
    });
  });
});

describe('localStorage integration', () => {
  it('loads theme from localStorage', async () => {
    localStorage.setItem('stargrazer-theme', 'light');
    await act(async () => { render(<App />); });
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
  });

  it('loads account from localStorage', async () => {
    localStorage.setItem('stargrazer-account', JSON.stringify({ name: 'Test User', email: 'test@test.com', avatarUrl: '' }));
    await act(async () => { render(<App />); });
    expect(screen.getByText('Test User')).toBeInTheDocument();
  });

  it('handles corrupted localStorage gracefully', async () => {
    localStorage.setItem('stargrazer-account', 'not valid json');
    await act(async () => { render(<App />); });
    expect(screen.getByText('User')).toBeInTheDocument();
  });
});

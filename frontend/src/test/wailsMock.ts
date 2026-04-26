import { vi } from 'vitest';

// These are exported as named exports to match the wailsjs module structure.
// vitest.config.ts aliases '../wailsjs/go/main/App' to this file.

export const StartBrowser = vi.fn().mockResolvedValue({ status: 'running', error: '' });
export const StopBrowser = vi.fn().mockResolvedValue({ status: 'stopped', error: '' });
export const GetBrowserStatus = vi.fn().mockResolvedValue({ status: 'stopped', error: '' });
export const GetBrowserConfig = vi.fn().mockResolvedValue({
  chromiumPath: '/usr/bin/chromium',
  cdpPort: 9222,
  headless: false,
  userDataDir: '/tmp/data',
  windowWidth: 1280,
  windowHeight: 900,
  extraFlags: ['--disable-infobars'],
});
export const UpdateBrowserConfig = vi.fn().mockResolvedValue({
  chromiumPath: '/usr/bin/chromium',
  cdpPort: 9222,
  headless: false,
  userDataDir: '/tmp/data',
  windowWidth: 1280,
  windowHeight: 900,
  extraFlags: ['--disable-infobars'],
});
export const ResetBrowserConfig = vi.fn().mockResolvedValue({
  chromiumPath: '',
  cdpPort: 9222,
  headless: false,
  userDataDir: '',
  windowWidth: 1280,
  windowHeight: 900,
  extraFlags: [],
});
export const RestartBrowser = vi.fn().mockResolvedValue({ status: 'running', error: '' });
export const GetPlatforms = vi.fn().mockResolvedValue([
  { id: 'instagram', name: 'Instagram', url: 'https://www.instagram.com', loggedIn: true, username: 'testuser', lastLogin: '2024-01-01T00:00:00Z', lastCheck: '2024-01-01T00:00:00Z', sessionDir: '/tmp/sessions' },
  { id: 'facebook', name: 'Facebook', url: 'https://www.facebook.com', loggedIn: false, username: '', lastLogin: '', lastCheck: '', sessionDir: '/tmp/sessions' },
  { id: 'tiktok', name: 'TikTok', url: 'https://www.tiktok.com', loggedIn: false, username: '', lastLogin: '', lastCheck: '', sessionDir: '/tmp/sessions' },
  { id: 'youtube', name: 'YouTube', url: 'https://www.youtube.com', loggedIn: false, username: '', lastLogin: '', lastCheck: '', sessionDir: '/tmp/sessions' },
  { id: 'linkedin', name: 'LinkedIn', url: 'https://www.linkedin.com', loggedIn: false, username: '', lastLogin: '', lastCheck: '', sessionDir: '/tmp/sessions' },
  { id: 'x', name: 'X', url: 'https://x.com', loggedIn: false, username: '', lastLogin: '', lastCheck: '', sessionDir: '/tmp/sessions' },
]);
export const OpenPlatform = vi.fn().mockResolvedValue({ status: 'running', error: '' });
export const CheckLoginStatus = vi.fn().mockResolvedValue({ id: 'instagram', name: 'Instagram', loggedIn: true, username: 'testuser' });
export const CheckAllLoginStatus = vi.fn().mockResolvedValue([]);
export const ImportCookies = vi.fn().mockResolvedValue({ id: 'instagram', name: 'Instagram', loggedIn: true, username: 'testuser' });
export const GetLogs = vi.fn().mockResolvedValue([
  { timestamp: '2024-01-01T00:00:00Z', level: 'info', source: 'test', message: 'Test log' },
]);
export const ExportLogs = vi.fn().mockResolvedValue('[]');
export const ClearLogs = vi.fn().mockResolvedValue(undefined);
export const TriggerUpload = vi.fn().mockResolvedValue({ success: true, message: 'Upload queued for 1 platform(s)' });
export const GetSchedules = vi.fn().mockResolvedValue([]);
export const CreateSchedule = vi.fn().mockResolvedValue({
  id: '1', name: 'Test', type: 'upload', platforms: ['instagram'], cronExpr: '0 */12 * * *',
  status: 'active', runCount: 0, auto: false, createdAt: '2024-01-01T00:00:00Z',
});
export const DeleteSchedule = vi.fn().mockResolvedValue(true);
export const PauseSchedule = vi.fn().mockResolvedValue({ id: '1', status: 'paused', name: 'Test' });
export const ResumeSchedule = vi.fn().mockResolvedValue({ id: '1', status: 'active', name: 'Test' });
export const PurgeSession = vi.fn().mockResolvedValue({ id: 'instagram', name: 'Instagram', loggedIn: false, username: '', sessionDir: '/tmp' });
export const SelectFile = vi.fn().mockResolvedValue('/tmp/photo.jpg');

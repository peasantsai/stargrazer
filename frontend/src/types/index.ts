// Shared TypeScript types for the Stargrazer frontend.
// Wire-format types mirror Go struct definitions in app.go.

export type Theme = 'dark' | 'light';

/** Discriminated union of all navigable views. */
export type View = 'chat' | 'schedules' | 'config' | `platform:${string}`;

export function isPlatformView(v: View): v is `platform:${string}` {
  return v.startsWith('platform:');
}

export function platformIdFromView(v: View): string {
  return (v as string).replace('platform:', '');
}

export interface ChatMessage {
  id: number;
  type: 'system' | 'info' | 'error' | 'success';
  text: string;
}

export interface AccountInfo {
  name: string;
  email: string;
  avatarUrl: string;
}

// --- Wire types (mirror Go structs) ---

export interface BrowserStatusResponse {
  status: string;
  error: string;
}

export interface BrowserConfigResponse {
  chromiumPath: string;
  cdpPort: number;
  headless: boolean;
  userDataDir: string;
  windowWidth: number;
  windowHeight: number;
  extraFlags: string[];
}

export interface PlatformResponse {
  id: string;
  name: string;
  url: string;
  loggedIn: boolean;
  username: string;
  lastLogin: string;
  lastCheck: string;
  sessionDir: string;
}

export interface ScheduleResponse {
  id: string;
  name: string;
  type: string;
  platforms: string[];
  cronExpr: string;
  nextRun: string;
  lastRun: string;
  status: string;
  createdAt: string;
  runCount: number;
  lastResult: string;
  auto: boolean;
  filePath?: string;
  caption?: string;
  hashtags?: string[];
}

export interface CreateScheduleRequest {
  name: string;
  type: string;
  platforms: string[];
  cronExpr: string;
  filePath?: string;
  caption?: string;
  hashtags?: string[];
}

export interface LogEntryResponse {
  timestamp: string;
  level: string;
  source: string;
  message: string;
}

export interface UploadRequest {
  platforms: string[];
  filePath: string;
  caption: string;
  hashtags: string[];
}

export interface UploadResponse {
  success: boolean;
  message: string;
}

// --- Automation types ---

export type AutomationAction = 'navigate' | 'click' | 'type' | 'wait' | 'evaluate' | 'scroll';

export const AUTOMATION_ACTIONS: { value: AutomationAction; label: string; hasTarget: boolean; hasValue: boolean; targetLabel: string; valueLabel: string }[] = [
  { value: 'navigate', label: 'Navigate',  hasTarget: true,  hasValue: false, targetLabel: 'URL',          valueLabel: '' },
  { value: 'click',    label: 'Click',     hasTarget: true,  hasValue: false, targetLabel: 'CSS Selector', valueLabel: '' },
  { value: 'type',     label: 'Type',      hasTarget: true,  hasValue: true,  targetLabel: 'CSS Selector', valueLabel: 'Text' },
  { value: 'wait',     label: 'Wait',      hasTarget: false, hasValue: true,  targetLabel: '',              valueLabel: 'Milliseconds' },
  { value: 'evaluate', label: 'Evaluate',  hasTarget: false, hasValue: true,  targetLabel: '',              valueLabel: 'JS Expression' },
  { value: 'scroll',   label: 'Scroll',    hasTarget: true,  hasValue: false, targetLabel: 'CSS Selector', valueLabel: '' },
];

export interface AutomationStepData {
  action: AutomationAction;
  target: string;
  value: string;
  label: string;
}

export interface AutomationData {
  id: string;
  platformId: string;
  name: string;
  description: string;
  steps: AutomationStepData[];
  createdAt: string;
  lastRun: string;
  runCount: number;
}

export interface RunAutomationResponse {
  success: boolean;
  message: string;
}

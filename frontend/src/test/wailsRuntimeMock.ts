import { vi } from 'vitest';

// Aliased by vitest.config.ts to substitute for wailsjs/runtime/runtime in tests.
// jsdom has no window.runtime so the real binding throws on EventsOnMultiple.

export const EventsOn = vi.fn((_event: string, _cb: (...data: unknown[]) => void) => () => {});
export const EventsOnMultiple = vi.fn((_event: string, _cb: (...data: unknown[]) => void, _max: number) => () => {});
export const EventsOnce = vi.fn((_event: string, _cb: (...data: unknown[]) => void) => () => {});
export const EventsOff = vi.fn((_event: string, ..._others: string[]) => {});
export const EventsOffAll = vi.fn(() => {});
export const EventsEmit = vi.fn((_event: string, ..._data: unknown[]) => {});
export const LogPrint = vi.fn(() => {});
export const LogTrace = vi.fn(() => {});
export const LogDebug = vi.fn(() => {});
export const LogInfo = vi.fn(() => {});
export const LogWarning = vi.fn(() => {});
export const LogError = vi.fn(() => {});
export const LogFatal = vi.fn(() => {});
export const Quit = vi.fn(() => {});
export const Hide = vi.fn(() => {});
export const Show = vi.fn(() => {});
export const Environment = vi.fn().mockResolvedValue({ buildType: 'dev', platform: 'test', arch: 'x64' });

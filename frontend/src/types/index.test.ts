import { describe, it, expect } from 'vitest';
import { isPlatformView, platformIdFromView } from './index';
import type { View } from './index';

describe('isPlatformView', () => {
  it('returns true for platform:* views', () => {
    expect(isPlatformView('platform:instagram')).toBe(true);
    expect(isPlatformView('platform:facebook')).toBe(true);
    expect(isPlatformView('platform:tiktok')).toBe(true);
  });

  it('returns false for chat', () => {
    expect(isPlatformView('chat')).toBe(false);
  });

  it('returns false for config', () => {
    expect(isPlatformView('config')).toBe(false);
  });

  it('works as a type narrowing guard', () => {
    const v: View = 'platform:youtube';
    if (isPlatformView(v)) {
      // TypeScript would confirm v is `platform:${string}` here
      expect(v.startsWith('platform:')).toBe(true);
    }
  });
});

describe('platformIdFromView', () => {
  it('extracts platform id from platform:instagram', () => {
    expect(platformIdFromView('platform:instagram')).toBe('instagram');
  });

  it('extracts platform id from platform:facebook', () => {
    expect(platformIdFromView('platform:facebook')).toBe('facebook');
  });

  it('extracts platform id from platform:tiktok', () => {
    expect(platformIdFromView('platform:tiktok')).toBe('tiktok');
  });

  it('extracts platform id from platform:youtube', () => {
    expect(platformIdFromView('platform:youtube')).toBe('youtube');
  });

  it('extracts platform id from platform:x', () => {
    expect(platformIdFromView('platform:x')).toBe('x');
  });

  it('returns full string for non-platform views (strips prefix)', () => {
    // The function simply replaces 'platform:' prefix, so non-platform views return unchanged
    expect(platformIdFromView('chat')).toBe('chat');
  });
});

import { useState } from 'react';
import { TriggerUpload, SelectFile, LogFromFrontend } from '../../wailsjs/go/main/App';
import type { ChatMessage, PlatformResponse, UploadRequest } from '../types';
import { PLATFORM_ICONS, PLATFORM_COLORS } from '../constants/platforms';
import { HamburgerBtn } from './HamburgerBtn';

interface Props {
  readonly messages: ChatMessage[];
  readonly messagesEndRef: React.RefObject<HTMLDivElement>;
  readonly sidebarOpen: boolean;
  readonly onToggleSidebar: () => void;
  readonly platforms: PlatformResponse[];
  readonly addMessage: (type: ChatMessage['type'], text: string) => void;
}

export function ChatPanel({
  messages,
  messagesEndRef, sidebarOpen, onToggleSidebar, platforms, addMessage,
}: Props) {
  const [caption, setCaption] = useState('');
  const [tags, setTags] = useState<string[]>([]);
  const [tagInput, setTagInput] = useState('');
  const [selectedFile, setSelectedFile] = useState('');
  const [selectedFileName, setSelectedFileName] = useState('');
  const [selectedPlatforms, setSelectedPlatforms] = useState<Set<string>>(new Set());
  const [uploading, setUploading] = useState(false);

  const togglePlatform = (id: string) => {
    setSelectedPlatforms(prev => {
      const next = new Set(prev);
      if (next.has(id)) { next.delete(id); } else { next.add(id); }
      return next;
    });
  };

  const handleTagKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if ((e.key === ' ' || e.key === 'Enter' || e.key === ',') && tagInput.trim()) {
      e.preventDefault();
      const raw = tagInput.trim().replace(/^#/, '');
      if (raw && !tags.includes(`#${raw}`)) {
        setTags(prev => [...prev, `#${raw}`]);
      }
      setTagInput('');
    } else if (e.key === 'Backspace' && !tagInput && tags.length > 0) {
      setTags(prev => prev.slice(0, -1));
    }
  };

  const removeTag = (tag: string) => setTags(prev => prev.filter(t => t !== tag));

  const handleSelectFile = async () => {
    const path = await SelectFile();
    if (path) {
      setSelectedFile(path);
      setSelectedFileName(path.split(/[/\\]/).pop() || path);
    }
  };

  const handleSend = async () => {
    if (selectedPlatforms.size === 0) { addMessage('error', 'Select at least one platform.'); return; }
    const finalTags = [...tags];
    if (tagInput.trim()) {
      const raw = tagInput.trim().replace(/^#/, '');
      if (raw) finalTags.push(`#${raw}`);
    }
    if (!selectedFile && !caption.trim() && finalTags.length === 0) {
      addMessage('error', 'Provide at least a file, caption, or hashtags.');
      return;
    }

    // Show user's request as a right-aligned bubble
    const platformNames = [...selectedPlatforms]
      .map(id => platforms.find(p => p.id === id)?.name ?? id)
      .join(', ');
    const parts: string[] = [`Upload to ${platformNames}`];
    if (selectedFileName) parts.push(`File: ${selectedFileName}`);
    if (caption.trim()) parts.push(`Caption: ${caption.trim()}`);
    if (finalTags.length > 0) parts.push(`Tags: ${finalTags.join(' ')}`);
    addMessage('user', parts.join('\n'));

    LogFromFrontend('info', 'chat', `Upload request: ${platformNames} | file=${selectedFileName} | caption=${caption.trim()} | tags=${finalTags.join(',')}`);

    setUploading(true);

    try {
      const req: UploadRequest = {
        platforms: [...selectedPlatforms],
        filePath: selectedFile,
        caption: caption.trim(),
        hashtags: finalTags,
      };
      const res = await TriggerUpload(req);
      if (res.success) {
        addMessage('success', res.message);
        LogFromFrontend('info', 'chat', `Upload success: ${res.message}`);
        setCaption('');
        setTags([]);
        setTagInput('');
        setSelectedFile('');
        setSelectedFileName('');
      } else {
        addMessage('error', res.message);
        LogFromFrontend('error', 'chat', `Upload failed: ${res.message}`);
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      addMessage('error', `Upload error: ${msg}`);
      LogFromFrontend('error', 'chat', `Upload exception: ${msg}`);
    }
    setUploading(false);
  };

  return (
    <div className="chat-panel">
      <div className="chat-header">
        <div className="chat-header-left">
          <HamburgerBtn sidebarOpen={sidebarOpen} onToggle={onToggleSidebar} />
          <h2>Chat</h2>
        </div>
      </div>

      <div className="chat-messages">
        {messages.length === 0 ? (
          <div className="chat-empty">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--text-secondary)" strokeWidth="1.5">
              <circle cx="12" cy="12" r="10"/>
              <path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"/>
              <path d="M2 12h20"/>
            </svg>
            <h3>Welcome to Stargrazer</h3>
            <p>Connect your social accounts on each platform page, then upload content below.</p>
          </div>
        ) : messages.map(msg => (
          <div key={msg.id} className={`message ${msg.type}`}>
            {msg.text.split('\n').map((line, i) => (
              <span key={i}>{line}{i < msg.text.split('\n').length - 1 && <br />}</span>
            ))}
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      <div className="chat-input-area">
        <div className="upload-platforms">
          {platforms.map(p => {
            const colors = PLATFORM_COLORS[p.id];
            return (
              <label
                key={p.id}
                className={`upload-platform-chip ${selectedPlatforms.has(p.id) ? 'selected' : ''} ${p.loggedIn ? '' : 'disabled'}`}
                style={{ '--chip-bg': colors?.bg } as React.CSSProperties}
              >
                <input
                  type="checkbox"
                  checked={selectedPlatforms.has(p.id)}
                  onChange={() => p.loggedIn && togglePlatform(p.id)}
                  disabled={!p.loggedIn}
                />
                <span className="upload-platform-icon">{PLATFORM_ICONS[p.id]}</span>
                <span>{p.name}</span>
                {!p.loggedIn && <span className="upload-platform-lock">Not connected</span>}
              </label>
            );
          })}
        </div>

        <div className="upload-form">
          <div className="upload-file-row">
            <button className="btn-secondary upload-file-btn" onClick={handleSelectFile}>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M21.44 11.05l-9.19 9.19a6 6 0 01-8.49-8.49l9.19-9.19a4 4 0 015.66 5.66l-9.2 9.19a2 2 0 01-2.83-2.83l8.49-8.48"/>
              </svg>
              {selectedFileName || 'Attach file'}
            </button>
            {selectedFile && (
              <button
                className="upload-file-clear"
                onClick={() => { setSelectedFile(''); setSelectedFileName(''); }}
                title="Remove file"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
              </button>
            )}
            <div className="tag-input-wrapper">
              {tags.map(tag => (
                <span key={tag} className="tag-bubble">
                  {tag}
                  <button className="tag-remove" onClick={() => removeTag(tag)}>
                    <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3">
                      <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
                    </svg>
                  </button>
                </span>
              ))}
              <input
                className="tag-input"
                type="text"
                placeholder={tags.length === 0 ? '#hashtags...' : ''}
                value={tagInput}
                onChange={e => setTagInput(e.target.value)}
                onKeyDown={handleTagKeyDown}
              />
            </div>
          </div>
          <div className="upload-caption-row">
            <textarea
              className="upload-caption"
              placeholder="Write your caption..."
              rows={2}
              value={caption}
              onChange={e => setCaption(e.target.value)}
            />
            <button
              className="btn-primary upload-send"
              onClick={handleSend}
              disabled={uploading || selectedPlatforms.size === 0}
            >
              {uploading ? 'Uploading...' : 'Send'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

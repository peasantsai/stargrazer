interface Props {
  sidebarOpen: boolean;
  onToggle: () => void;
}

export function HamburgerBtn({ sidebarOpen, onToggle }: Props) {
  if (sidebarOpen) return null;
  return (
    <button className="sidebar-toggle-float" onClick={onToggle} aria-label="Open sidebar">
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <line x1="3" y1="12" x2="21" y2="12" />
        <line x1="3" y1="6" x2="21" y2="6" />
        <line x1="3" y1="18" x2="21" y2="18" />
      </svg>
    </button>
  );
}

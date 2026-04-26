export interface ChromiumFlag {
  flag: string;
  label: string;
  description: string;
  /** Flags that are dangerous should show a warning in the UI. */
  dangerous?: boolean;
}

export interface ChromiumFlagGroup {
  category: string;
  flags: ChromiumFlag[];
}

export const CHROMIUM_FLAGS: ChromiumFlagGroup[] = [
  {
    category: 'Stealth & Anti-Detection',
    flags: [
      { flag: '--disable-blink-features=AutomationControlled', label: 'Hide WebDriver Flag', description: 'Remove navigator.webdriver from the JS runtime' },
      { flag: '--disable-features=AutomationControlled', label: 'No Automation Signal', description: 'Remove the automation-controlled feature flag' },
      { flag: '--disable-infobars', label: 'Disable Infobars', description: 'Suppress "Chrome is being controlled" bar' },
      { flag: '--disable-background-networking', label: 'No Background Network', description: 'Prevent background network requests' },
      { flag: '--disable-client-side-phishing-detection', label: 'No Phishing Detection', description: 'Disable phishing detection' },
      { flag: '--disable-component-update', label: 'No Component Updates', description: 'Prevent auto-updates' },
      { flag: '--disable-default-apps', label: 'No Default Apps', description: 'Skip default apps' },
      { flag: '--disable-domain-reliability', label: 'No Domain Reliability', description: 'Disable monitoring' },
      { flag: '--metrics-recording-only', label: 'Metrics Record Only', description: 'Never upload metrics' },
      { flag: '--no-pings', label: 'No Pings', description: 'Disable auditing pings' },
      { flag: '--safebrowsing-disable-auto-update', label: 'No SafeBrowsing Updates', description: 'No SB updates' },
    ],
  },
  {
    category: 'Privacy & Telemetry',
    flags: [
      { flag: '--disable-extensions', label: 'Disable Extensions', description: 'No extensions' },
      { flag: '--disable-sync', label: 'Disable Sync', description: 'No Google sync' },
      { flag: '--disable-translate', label: 'Disable Translate', description: 'No translation' },
      { flag: '--incognito', label: 'Incognito Mode', description: 'Private browsing' },
    ],
  },
  {
    category: 'Automation & CDP',
    flags: [
      { flag: '--disable-hang-monitor', label: 'Disable Hang Monitor', description: 'No unresponsive dialog' },
      { flag: '--disable-popup-blocking', label: 'Disable Popup Blocking', description: 'Allow popups' },
      { flag: '--disable-prompt-on-repost', label: 'No Repost Prompt', description: 'No resubmission dialog' },
      { flag: '--disable-ipc-flooding-protection', label: 'No IPC Flood Protection', description: 'Rapid CDP commands' },
      { flag: '--disable-renderer-backgrounding', label: 'No Renderer Backgrounding', description: 'Full CPU priority' },
      { flag: '--disable-background-timer-throttling', label: 'No Timer Throttling', description: 'No background throttle' },
      { flag: '--disable-backgrounding-occluded-windows', label: 'No Occluded Throttling', description: 'Keep hidden active' },
      { flag: '--enable-features=NetworkService,NetworkServiceInProcess', label: 'In-Process Network', description: 'Faster CDP' },
    ],
  },
  {
    category: 'Display & UI',
    flags: [
      { flag: '--force-dark-mode', label: 'Force Dark Mode', description: 'Dark browser chrome' },
      { flag: '--enable-features=WebUIDarkMode', label: 'WebUI Dark Mode', description: 'Dark internal pages' },
      { flag: '--disable-notifications', label: 'Disable Notifications', description: 'Block notifications' },
      { flag: '--start-maximized', label: 'Start Maximized', description: 'Maximized window' },
      { flag: '--start-fullscreen', label: 'Start Fullscreen', description: 'Fullscreen mode' },
      { flag: '--hide-scrollbars', label: 'Hide Scrollbars', description: 'No scrollbars' },
    ],
  },
  {
    category: 'Network',
    flags: [
      { flag: '--ignore-certificate-errors', label: 'Ignore Cert Errors', description: 'Skip SSL certificate validation', dangerous: true },
      { flag: '--disable-web-security', label: 'Disable Web Security', description: 'Disable CORS and same-origin policy', dangerous: true },
      { flag: '--allow-running-insecure-content', label: 'Allow Insecure Content', description: 'Allow mixed HTTP/HTTPS content', dangerous: true },
      { flag: '--disable-gpu', label: 'Disable GPU', description: 'No GPU acceleration' },
    ],
  },
];

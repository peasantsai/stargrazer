package config

import (
	"fmt"
	"sync"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// BrowserConfig holds Chromium and CDP settings.
type BrowserConfig struct {
	ChromiumPath   string `mapstructure:"chromium_path" yaml:"chromium_path"`
	CDPPort        int    `mapstructure:"cdp_port" yaml:"cdp_port"`
	Headless       bool   `mapstructure:"headless" yaml:"headless"`
	UserDataDir    string `mapstructure:"user_data_dir" yaml:"user_data_dir"`
	WindowWidth    int    `mapstructure:"window_width" yaml:"window_width"`
	WindowHeight   int    `mapstructure:"window_height" yaml:"window_height"`
	ExtraFlags     []string `mapstructure:"extra_flags" yaml:"extra_flags"`
}

// WindowConfig holds the Wails app window settings.
type WindowConfig struct {
	Title    string `mapstructure:"title" yaml:"title"`
	Width    int    `mapstructure:"width" yaml:"width"`
	Height   int    `mapstructure:"height" yaml:"height"`
	MinWidth int    `mapstructure:"min_width" yaml:"min_width"`
	MinHeight int   `mapstructure:"min_height" yaml:"min_height"`
}

// SchedulerConfig holds job scheduler settings.
type SchedulerConfig struct {
	Enabled                  bool   `mapstructure:"enabled" yaml:"enabled"`
	DefaultKeepAliveInterval string `mapstructure:"default_keepalive_interval" yaml:"default_keepalive_interval"`
}

// AppConfig is the top-level configuration.
type AppConfig struct {
	Window    WindowConfig    `mapstructure:"window" yaml:"window"`
	Browser   BrowserConfig   `mapstructure:"browser" yaml:"browser"`
	Scheduler SchedulerConfig `mapstructure:"scheduler" yaml:"scheduler"`
}

var (
	instance *AppConfig
	mu       sync.RWMutex
	v        *viper.Viper
)

// GetWindow is a convenience accessor.
func GetWindow() WindowConfig {
	return Get().Window
}

// defaults returns a config with stealth-optimized defaults for automation.
func defaults() AppConfig {
	return AppConfig{
		Window: WindowConfig{
			Title:     "Stargrazer",
			Width:     1024,
			Height:    768,
			MinWidth:  800,
			MinHeight: 600,
		},
		Scheduler: SchedulerConfig{
			Enabled:                  true,
			DefaultKeepAliveInterval: "0 */12 * * *",
		},
		Browser: BrowserConfig{
			ChromiumPath: "",
			CDPPort:      9222,
			Headless:     false,
			UserDataDir:  "",
			WindowWidth:  1280,
			WindowHeight: 900,
			ExtraFlags: []string{
				// Stealth & anti-detection
				"--disable-blink-features=AutomationControlled",
				"--disable-features=AutomationControlled",
				"--disable-infobars",
				"--disable-background-networking",
				"--disable-client-side-phishing-detection",
				"--disable-component-update",
				"--disable-default-apps",
				"--disable-domain-reliability",
				"--disable-hang-monitor",
				"--disable-sync",
				"--disable-translate",
				"--metrics-recording-only",
				"--no-pings",
				"--safebrowsing-disable-auto-update",
				// Automation-friendly
				"--disable-popup-blocking",
				"--disable-prompt-on-repost",
				"--disable-ipc-flooding-protection",
				"--disable-renderer-backgrounding",
				"--disable-background-timer-throttling",
				"--disable-backgrounding-occluded-windows",
				// Display
				"--force-dark-mode",
				"--enable-features=WebUIDarkMode",
				"--disable-notifications",
			},
		},
	}
}

// Init loads configuration from file, env, and CLI flags.
// Call this once at startup before accessing Get().
func Init() error {
	v = viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.stargrazer")

	setDefaults(v)
	bindFlags(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("reading config: %w", err)
		}
	}

	v.AutomaticEnv()
	v.SetEnvPrefix("STARGRAZER")

	cfg := defaults()
	if err := v.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("unmarshaling config: %w", err)
	}

	mu.Lock()
	instance = &cfg
	mu.Unlock()

	return nil
}

// Get returns the current config snapshot.
func Get() AppConfig {
	mu.RLock()
	defer mu.RUnlock()
	if instance == nil {
		return defaults()
	}
	return *instance
}

// Update applies a partial config change at runtime and persists the singleton.
func Update(fn func(*AppConfig)) AppConfig {
	mu.Lock()
	defer mu.Unlock()
	if instance == nil {
		d := defaults()
		instance = &d
	}
	fn(instance)
	return *instance
}

// Reset restores all config to defaults.
func Reset() AppConfig {
	mu.Lock()
	defer mu.Unlock()
	d := defaults()
	instance = &d
	return *instance
}

// GetBrowser is a convenience accessor.
func GetBrowser() BrowserConfig {
	return Get().Browser
}

// GetScheduler is a convenience accessor.
func GetScheduler() SchedulerConfig {
	return Get().Scheduler
}

func setDefaults(v *viper.Viper) {
	d := defaults()
	v.SetDefault("window.title", d.Window.Title)
	v.SetDefault("window.width", d.Window.Width)
	v.SetDefault("window.height", d.Window.Height)
	v.SetDefault("window.min_width", d.Window.MinWidth)
	v.SetDefault("window.min_height", d.Window.MinHeight)
	v.SetDefault("scheduler.enabled", d.Scheduler.Enabled)
	v.SetDefault("scheduler.default_keepalive_interval", d.Scheduler.DefaultKeepAliveInterval)
	v.SetDefault("browser.chromium_path", d.Browser.ChromiumPath)
	v.SetDefault("browser.cdp_port", d.Browser.CDPPort)
	v.SetDefault("browser.headless", d.Browser.Headless)
	v.SetDefault("browser.user_data_dir", d.Browser.UserDataDir)
	v.SetDefault("browser.window_width", d.Browser.WindowWidth)
	v.SetDefault("browser.window_height", d.Browser.WindowHeight)
	v.SetDefault("browser.extra_flags", d.Browser.ExtraFlags)
}

func bindFlags(v *viper.Viper) {
	pflag.Int("cdp-port", 9222, "Chrome DevTools Protocol port")
	pflag.Bool("headless", false, "Run browser in headless mode")
	pflag.String("chromium-path", "", "Path to Chromium executable")
	pflag.Parse()
	v.BindPFlag("browser.cdp_port", pflag.Lookup("cdp-port"))
	v.BindPFlag("browser.headless", pflag.Lookup("headless"))
	v.BindPFlag("browser.chromium_path", pflag.Lookup("chromium-path"))
}

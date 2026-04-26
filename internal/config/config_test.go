package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestDefaults(t *testing.T) {
	d := defaults()

	t.Run("CDPPort is 9222", func(t *testing.T) {
		if d.Browser.CDPPort != 9222 {
			t.Errorf("expected CDPPort 9222, got %d", d.Browser.CDPPort)
		}
	})

	t.Run("BrowserWindowWidth is 1280", func(t *testing.T) {
		if d.Browser.WindowWidth != 1280 {
			t.Errorf("expected WindowWidth 1280, got %d", d.Browser.WindowWidth)
		}
	})

	t.Run("BrowserWindowHeight is 900", func(t *testing.T) {
		if d.Browser.WindowHeight != 900 {
			t.Errorf("expected WindowHeight 900, got %d", d.Browser.WindowHeight)
		}
	})

	t.Run("WindowTitle is Stargrazer", func(t *testing.T) {
		if d.Window.Title != "Stargrazer" {
			t.Errorf("expected Title 'Stargrazer', got %q", d.Window.Title)
		}
	})

	t.Run("AppWindowWidth is 1024", func(t *testing.T) {
		if d.Window.Width != 1024 {
			t.Errorf("expected Window.Width 1024, got %d", d.Window.Width)
		}
	})

	t.Run("AppWindowHeight is 768", func(t *testing.T) {
		if d.Window.Height != 768 {
			t.Errorf("expected Window.Height 768, got %d", d.Window.Height)
		}
	})

	t.Run("MinWidth is 800", func(t *testing.T) {
		if d.Window.MinWidth != 800 {
			t.Errorf("expected MinWidth 800, got %d", d.Window.MinWidth)
		}
	})

	t.Run("MinHeight is 600", func(t *testing.T) {
		if d.Window.MinHeight != 600 {
			t.Errorf("expected MinHeight 600, got %d", d.Window.MinHeight)
		}
	})

	t.Run("Headless is false", func(t *testing.T) {
		if d.Browser.Headless {
			t.Error("expected Headless false")
		}
	})

	t.Run("SchedulerEnabled is true", func(t *testing.T) {
		if !d.Scheduler.Enabled {
			t.Error("expected Scheduler.Enabled true")
		}
	})

	t.Run("DefaultKeepAliveInterval", func(t *testing.T) {
		if d.Scheduler.DefaultKeepAliveInterval != "0 */12 * * *" {
			t.Errorf("expected '0 */12 * * *', got %q", d.Scheduler.DefaultKeepAliveInterval)
		}
	})

	t.Run("ExtraFlags is non-empty", func(t *testing.T) {
		if len(d.Browser.ExtraFlags) == 0 {
			t.Error("expected non-empty ExtraFlags")
		}
	})

	t.Run("ChromiumPath is empty by default", func(t *testing.T) {
		if d.Browser.ChromiumPath != "" {
			t.Errorf("expected empty ChromiumPath, got %q", d.Browser.ChromiumPath)
		}
	})
}

func TestGetReturnsDefaultsWhenNoInit(t *testing.T) {
	// Reset the singleton so Get() falls back to defaults.
	mu.Lock()
	saved := instance
	instance = nil
	mu.Unlock()
	defer func() {
		mu.Lock()
		instance = saved
		mu.Unlock()
	}()

	cfg := Get()
	d := defaults()

	if cfg.Browser.CDPPort != d.Browser.CDPPort {
		t.Errorf("Get() CDPPort: expected %d, got %d", d.Browser.CDPPort, cfg.Browser.CDPPort)
	}
	if cfg.Window.Title != d.Window.Title {
		t.Errorf("Get() Title: expected %q, got %q", d.Window.Title, cfg.Window.Title)
	}
	if cfg.Scheduler.Enabled != d.Scheduler.Enabled {
		t.Errorf("Get() Scheduler.Enabled: expected %v, got %v", d.Scheduler.Enabled, cfg.Scheduler.Enabled)
	}
}

func TestInitLoadsWithoutError(t *testing.T) {
	// Init should succeed even when no config.yaml exists (ConfigFileNotFoundError is tolerated).
	// Save and restore the singleton after test.
	mu.Lock()
	saved := instance
	savedV := v
	mu.Unlock()
	defer func() {
		mu.Lock()
		instance = saved
		v = savedV
		mu.Unlock()
	}()

	// Change to a temp dir so no config.yaml is found.
	err := Init()
	if err != nil {
		t.Fatalf("Init() returned unexpected error: %v", err)
	}

	cfg := Get()
	if cfg.Browser.CDPPort != 9222 {
		t.Errorf("after Init() CDPPort: expected 9222, got %d", cfg.Browser.CDPPort)
	}
}

func TestUpdate(t *testing.T) {
	mu.Lock()
	saved := instance
	mu.Unlock()
	defer func() {
		mu.Lock()
		instance = saved
		mu.Unlock()
	}()

	// Reset to defaults first.
	Reset()

	updated := Update(func(c *AppConfig) {
		c.Browser.CDPPort = 5555
		c.Window.Title = "Modified"
	})

	if updated.Browser.CDPPort != 5555 {
		t.Errorf("Update CDPPort: expected 5555, got %d", updated.Browser.CDPPort)
	}
	if updated.Window.Title != "Modified" {
		t.Errorf("Update Title: expected 'Modified', got %q", updated.Window.Title)
	}

	// Get should also reflect the change.
	got := Get()
	if got.Browser.CDPPort != 5555 {
		t.Errorf("Get after Update CDPPort: expected 5555, got %d", got.Browser.CDPPort)
	}
}

func TestUpdateWhenInstanceNil(t *testing.T) {
	mu.Lock()
	saved := instance
	instance = nil
	mu.Unlock()
	defer func() {
		mu.Lock()
		instance = saved
		mu.Unlock()
	}()

	updated := Update(func(c *AppConfig) {
		c.Browser.CDPPort = 7777
	})

	if updated.Browser.CDPPort != 7777 {
		t.Errorf("expected CDPPort 7777, got %d", updated.Browser.CDPPort)
	}
	// Other fields should be defaults.
	if updated.Window.Title != "Stargrazer" {
		t.Errorf("expected default title, got %q", updated.Window.Title)
	}
}

func TestReset(t *testing.T) {
	mu.Lock()
	saved := instance
	mu.Unlock()
	defer func() {
		mu.Lock()
		instance = saved
		mu.Unlock()
	}()

	// Modify, then reset.
	Update(func(c *AppConfig) {
		c.Browser.CDPPort = 1111
		c.Window.Title = "Changed"
	})

	resetCfg := Reset()
	d := defaults()

	if resetCfg.Browser.CDPPort != d.Browser.CDPPort {
		t.Errorf("Reset CDPPort: expected %d, got %d", d.Browser.CDPPort, resetCfg.Browser.CDPPort)
	}
	if resetCfg.Window.Title != d.Window.Title {
		t.Errorf("Reset Title: expected %q, got %q", d.Window.Title, resetCfg.Window.Title)
	}
}

func TestGetBrowser(t *testing.T) {
	mu.Lock()
	saved := instance
	mu.Unlock()
	defer func() {
		mu.Lock()
		instance = saved
		mu.Unlock()
	}()

	Reset()

	bc := GetBrowser()
	if bc.CDPPort != 9222 {
		t.Errorf("GetBrowser CDPPort: expected 9222, got %d", bc.CDPPort)
	}
	if bc.WindowWidth != 1280 {
		t.Errorf("GetBrowser WindowWidth: expected 1280, got %d", bc.WindowWidth)
	}
}

func TestGetWindow(t *testing.T) {
	mu.Lock()
	saved := instance
	mu.Unlock()
	defer func() {
		mu.Lock()
		instance = saved
		mu.Unlock()
	}()

	Reset()

	wc := GetWindow()
	if wc.Title != "Stargrazer" {
		t.Errorf("GetWindow Title: expected 'Stargrazer', got %q", wc.Title)
	}
	if wc.Width != 1024 {
		t.Errorf("GetWindow Width: expected 1024, got %d", wc.Width)
	}
	if wc.Height != 768 {
		t.Errorf("GetWindow Height: expected 768, got %d", wc.Height)
	}
}

func TestGetScheduler(t *testing.T) {
	mu.Lock()
	saved := instance
	mu.Unlock()
	defer func() {
		mu.Lock()
		instance = saved
		mu.Unlock()
	}()

	Reset()

	sc := GetScheduler()
	if !sc.Enabled {
		t.Error("GetScheduler Enabled: expected true")
	}
	if sc.DefaultKeepAliveInterval != "0 */12 * * *" {
		t.Errorf("GetScheduler DefaultKeepAliveInterval: expected '0 */12 * * *', got %q", sc.DefaultKeepAliveInterval)
	}
}

// TestSetDefaultsSetsAllKeys verifies setDefaults covers all config keys.
func TestSetDefaultsSetsAllKeys(t *testing.T) {
	testV := viper.New()
	setDefaults(testV)

	if testV.GetString("window.title") != "Stargrazer" {
		t.Errorf("expected window.title 'Stargrazer', got %q", testV.GetString("window.title"))
	}
	if testV.GetInt("browser.cdp_port") != 9222 {
		t.Errorf("expected browser.cdp_port 9222, got %d", testV.GetInt("browser.cdp_port"))
	}
	if testV.GetBool("scheduler.enabled") != true {
		t.Error("expected scheduler.enabled true")
	}
	if testV.GetString("scheduler.default_keepalive_interval") != "0 */12 * * *" {
		t.Errorf("expected default interval, got %q", testV.GetString("scheduler.default_keepalive_interval"))
	}
}

func TestDefaultExtraFlagsContainStealthFlags(t *testing.T) {
	d := defaults()
	flags := d.Browser.ExtraFlags

	expected := []string{
		"--disable-blink-features=AutomationControlled",
		"--disable-infobars",
		"--force-dark-mode",
	}
	for _, e := range expected {
		found := false
		for _, f := range flags {
			if f == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %q in ExtraFlags", e)
		}
	}
}

func TestUpdateMultipleTimes(t *testing.T) {
	mu.Lock()
	saved := instance
	mu.Unlock()
	defer func() {
		mu.Lock()
		instance = saved
		mu.Unlock()
	}()

	Reset()

	Update(func(c *AppConfig) {
		c.Browser.CDPPort = 1111
	})
	Update(func(c *AppConfig) {
		c.Browser.CDPPort = 2222
	})

	cfg := Get()
	if cfg.Browser.CDPPort != 2222 {
		t.Errorf("expected CDPPort 2222, got %d", cfg.Browser.CDPPort)
	}
	// Other fields should still be defaults
	if cfg.Window.Title != "Stargrazer" {
		t.Errorf("expected default title, got %q", cfg.Window.Title)
	}
}

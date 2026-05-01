package browser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"stargrazer/internal/config"
	"stargrazer/internal/logger"

	"github.com/chromedp/chromedp"
)

// stepTimeout is how long each individual step is allowed to take.
const stepTimeout = 30 * time.Second

// ConnectChromedp returns a chromedp context connected to the running browser.
// The caller must call the returned cancel function when done.
func (m *Manager) ConnectChromedp(parentCtx context.Context) (context.Context, context.CancelFunc, error) {
	if !m.IsRunning() {
		return nil, nil, fmt.Errorf("browser is not running")
	}

	port := config.GetBrowser().CDPPort
	wsURL := fmt.Sprintf("ws://127.0.0.1:%d", port)

	allocCtx, allocCancel := chromedp.NewRemoteAllocator(parentCtx, wsURL)
	ctx, ctxCancel := chromedp.NewContext(allocCtx)

	cancel := func() {
		ctxCancel()
		allocCancel()
	}

	// Bind the browser + target to the long-lived ctx. chromedp ties the
	// websocket lifecycle to whatever context is passed to the FIRST Run call;
	// without this warm-up the next Exec* (which uses a WithTimeout child)
	// would own that lifecycle and tear down the connection on its defer cancel.
	if err := chromedp.Run(ctx); err != nil {
		cancel()
		return nil, nil, fmt.Errorf("attach target: %w", err)
	}

	return ctx, cancel, nil
}

// ExecNavigate navigates the current page and waits for DOMContentLoaded.
func (m *Manager) ExecNavigate(ctx context.Context, url string) error {
	logger.Info("cdpexec", fmt.Sprintf("navigate → %s", url))
	tctx, cancel := context.WithTimeout(ctx, stepTimeout)
	defer cancel()
	err := chromedp.Run(tctx, chromedp.Navigate(url))
	if err != nil {
		return fmt.Errorf("navigate %s: %w", url, err)
	}
	// Wait for page to be interactive.
	_ = chromedp.Run(tctx, chromedp.WaitReady("body"))
	logger.Debug("cdpexec", "navigate OK — page ready")
	return nil
}

// ExecClick finds an element using fallback selectors and clicks it.
func (m *Manager) ExecClick(ctx context.Context, target string, selectors [][]string) error {
	return m.execWithSelectors(ctx, target, selectors, "click", func(tctx context.Context, sel string) error {
		return chromedp.Run(tctx,
			chromedp.WaitVisible(sel, selectorOpt(sel)),
			chromedp.ScrollIntoView(sel, selectorOpt(sel)),
			chromedp.Click(sel, selectorOpt(sel)),
		)
	})
}

// ExecType finds an element and types text into it using keyboard events.
func (m *Manager) ExecType(ctx context.Context, target, value string, selectors [][]string) error {
	return m.execWithSelectors(ctx, target, selectors, "type", func(tctx context.Context, sel string) error {
		opt := selectorOpt(sel)
		return chromedp.Run(tctx,
			chromedp.WaitVisible(sel, opt),
			chromedp.ScrollIntoView(sel, opt),
			chromedp.Click(sel, opt),              // Focus the element
			chromedp.SetValue(sel, "", opt),        // Clear existing value
			chromedp.SendKeys(sel, value, opt),     // Type character by character
		)
	})
}

// ExecScroll finds an element and scrolls it into view.
func (m *Manager) ExecScroll(ctx context.Context, target string, selectors [][]string) error {
	return m.execWithSelectors(ctx, target, selectors, "scroll", func(tctx context.Context, sel string) error {
		return chromedp.Run(tctx,
			chromedp.ScrollIntoView(sel, selectorOpt(sel)),
		)
	})
}

// ExecEvaluate runs arbitrary JS in the page.
func (m *Manager) ExecEvaluate(ctx context.Context, expression string) error {
	logger.Info("cdpexec", fmt.Sprintf("evaluate JS (%d chars)", len(expression)))
	tctx, cancel := context.WithTimeout(ctx, stepTimeout)
	defer cancel()
	var res interface{}
	err := chromedp.Run(tctx, chromedp.Evaluate(expression, &res))
	if err != nil {
		return fmt.Errorf("evaluate: %w", err)
	}
	logger.Debug("cdpexec", "evaluate OK")
	return nil
}

// execWithSelectors tries the primary target, then each selector group,
// until the action succeeds or all options are exhausted.
func (m *Manager) execWithSelectors(
	ctx context.Context,
	target string,
	selectors [][]string,
	actionName string,
	action func(ctx context.Context, sel string) error,
) error {
	// Build ordered list of selectors to try.
	candidates := buildCandidates(target, selectors)
	if len(candidates) == 0 {
		return fmt.Errorf("%s: no selectors available", actionName)
	}

	var lastErr error
	for i, sel := range candidates {
		tctx, cancel := context.WithTimeout(ctx, 8*time.Second)
		err := action(tctx, sel)
		cancel()
		if err == nil {
			if i > 0 {
				logger.Info("cdpexec", fmt.Sprintf("%s OK with fallback selector #%d: %s", actionName, i+1, truncate(sel, 80)))
			} else {
				logger.Debug("cdpexec", fmt.Sprintf("%s OK: %s", actionName, truncate(sel, 80)))
			}
			return nil
		}
		lastErr = err
		logger.Debug("cdpexec", fmt.Sprintf("%s: selector %d/%d failed (%s): %v", actionName, i+1, len(candidates), truncate(sel, 60), err))
	}

	return fmt.Errorf("%s failed after %d selectors: %w", actionName, len(candidates), lastErr)
}

// buildCandidates produces a flat, prioritized list of selectors to try.
// It prefers ARIA and text selectors (most stable), then CSS, then XPath.
func buildCandidates(target string, selectors [][]string) []string {
	if len(selectors) == 0 {
		if target != "" {
			return []string{target}
		}
		return nil
	}

	var aria, text, css, xpath []string
	for _, group := range selectors {
		for _, sel := range group {
			sel = strings.TrimSpace(sel)
			if sel == "" {
				continue
			}
			switch {
			case strings.HasPrefix(sel, "aria/"):
				aria = append(aria, sel)
			case strings.HasPrefix(sel, "text/"):
				text = append(text, sel)
			case strings.HasPrefix(sel, "xpath/"):
				xpath = append(xpath, sel)
			case strings.HasPrefix(sel, "pierce/"):
				// Treat pierce as CSS (chromedp doesn't cross shadow DOM, but it's worth trying)
				css = append(css, strings.TrimPrefix(sel, "pierce/"))
			default:
				css = append(css, sel)
			}
		}
	}

	// Priority: text > aria > css > xpath.
	// Text and ARIA are most stable across renders.
	var out []string
	out = append(out, text...)
	out = append(out, aria...)
	out = append(out, css...)
	out = append(out, xpath...)

	// Also include the original target if not already present.
	if target != "" {
		found := false
		for _, s := range out {
			if s == target {
				found = true
				break
			}
		}
		if !found {
			out = append(out, target)
		}
	}

	return dedup(out)
}

// selectorOpt returns the appropriate chromedp query option for a selector string.
func selectorOpt(sel string) chromedp.QueryOption {
	switch {
	case strings.HasPrefix(sel, "xpath/"):
		return chromedp.BySearch
	case strings.HasPrefix(sel, "aria/"):
		return chromedp.BySearch
	case strings.HasPrefix(sel, "text/"):
		return chromedp.BySearch
	default:
		return chromedp.ByQuery
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func dedup(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

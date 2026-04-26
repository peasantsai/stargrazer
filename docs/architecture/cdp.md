# CDP Integration

Stargrazer communicates with Chromium via the Chrome DevTools Protocol over WebSocket.

## Connection Flow

1. Browser starts with `--remote-debugging-port=9222`
2. App polls `http://127.0.0.1:9222/json` for available targets
3. Connects to page or service worker targets via WebSocket
4. Sends CDP commands as JSON messages with incrementing IDs

## Cookie Operations

### Reading (via extension)

The cookies extension's service worker has access to `chrome.cookies.getAll()`:

```javascript
// Executed via Runtime.evaluate on the extension's service worker
const cookies = await chrome.cookies.getAll({domain: ".facebook.com"});
return JSON.stringify(cookies);
```

### Writing

Cookies are injected via `Network.setCookie`:

```json
{
  "method": "Network.setCookie",
  "params": {
    "name": "c_user",
    "value": "100000856037530",
    "domain": ".facebook.com",
    "path": "/",
    "secure": true,
    "url": "https://www.facebook.com/"
  }
}
```

### Netscape Format Parsing

Users paste cookies in Netscape/curl format:

```
.facebook.com  TRUE  /  TRUE  1811714556  datr  8TntafXk2CQB7RuROx54yyz9
```

Fields: domain, includeSubdomains, path, secure, expiry, name, value (tab-separated).

## Navigation

New tabs are created via the HTTP endpoint:

```
GET http://127.0.0.1:9222/json/new?https://www.instagram.com
```

Page navigation uses `Page.navigate`:

```json
{"method": "Page.navigate", "params": {"url": "https://www.instagram.com"}}
```

## Stealth Flags

Key flags applied to hide automation:

- `--disable-blink-features=AutomationControlled` — Removes `navigator.webdriver`
- `--disable-infobars` — No "controlled by automation" bar
- `--disable-features=AutomationControlled` — Removes feature flag

# Connecting Accounts

Stargrazer supports 6 platforms: **Facebook**, **Instagram**, **TikTok**, **YouTube**, **LinkedIn**, and **X**.

## Login Flow

1. Go to **Settings** and find the **Social Media Connections** section
2. Click a platform card (e.g., Instagram)
3. The browser opens the platform's login page AND a cookie paste modal appears
4. Log in to the platform in the browser
5. Click the **cookies extension** icon (pinned in the toolbar)
6. Export cookies in **Netscape format**
7. Paste the cookies into the modal and click **Import Cookies**

## What Happens After Import

- Cookies are parsed and saved to disk (`%APPDATA%/stargrazer/sessions/browser_profile/cookies/`)
- If the browser is running, cookies are injected immediately via CDP
- An **auto keep-alive schedule** is created to refresh the session before cookies expire
- The platform card turns green showing "Connected"

## Session Persistence

All platforms share a single browser profile. Cookies persist across:

- Browser restarts
- App restarts
- System reboots

## Purging a Session

To disconnect and re-login:

1. Click the **(i)** info button on the platform card
2. Click **Purge Session**
3. The stored cookies are deleted
4. Click the card again to reconnect

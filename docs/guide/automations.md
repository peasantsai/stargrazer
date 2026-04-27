# Automations

The **Automations** section on each platform page lets you record, store, and replay named sequences of browser actions without writing code.

## What Is an Automation?

An automation is a named list of **steps**. Each step performs one CDP-level action in the currently open browser tab. Steps run in order and stop immediately if any step fails.

## Supported Actions

| Action | Required fields | Description |
|--------|-----------------|-------------|
| `navigate` | Target (URL) | Navigate the active tab to a URL |
| `click` | Target (CSS selector) | Click the first matching element |
| `type` | Target (CSS selector), Value (text) | Set element value and fire `input`/`change` events |
| `wait` | Value (milliseconds) | Pause execution for the given duration |
| `evaluate` | Value (JS expression) | Run arbitrary JavaScript in the page context |
| `scroll` | Target (CSS selector) | Scroll the first matching element into view |

## Creating an Automation

1. Navigate to a platform page (e.g., **Instagram** in the sidebar)
2. Click **+ New Automation**
3. Enter a name and optional description
4. Add steps using the **+ Add Step** button
5. For each step: choose an action, fill in Target and/or Value
6. Click **Save**

## Running an Automation

1. Find the automation card under the platform page
2. Click **▶ Run**
3. The browser must be running — start it first from the Chat view if needed
4. Results appear in the application logs (click the log icon top-right)

> **Tip**: Keep automations short and focused. Long automations are harder to debug when a step fails.

## Editing and Deleting

- Click the **pencil icon** on an automation card to open the editor
- Reorder steps by dragging (or delete individual steps with the trash icon)
- Click the **trash icon** on the card to delete the entire automation

## Persistence

Automations are stored as JSON files under the application data directory:

```
sessions/automations/<platformID>.json
```

Each file holds all automations for that platform. The `RunCount` and `LastRun` fields are updated automatically after each successful run.

## Run Statistics

The automation card displays:

- **Run count** — total times successfully completed
- **Last run** — timestamp of the last execution
- **Created** — when the automation was first saved

## Limitations

- Automations run **synchronously** in the current browser tab
- Only one automation can run at a time
- The `wait` action uses wall-clock sleep; it does not wait for a specific DOM condition
- No loop or conditional logic — steps are always executed top-to-bottom

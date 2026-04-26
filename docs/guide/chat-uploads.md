# Chat & Uploads

The Chat view is the main interface for uploading content to social media platforms.

## Upload Form

The bottom of the Chat view contains:

1. **Platform checkboxes** — Select target platforms (only connected ones are enabled)
2. **File picker** — Click "Attach file" to open a native file dialog (images and videos)
3. **Hashtag input** — Type hashtags, press Space/Enter/Comma to create bubbles
4. **Caption textarea** — Write your post caption
5. **Send button** — Triggers the upload workflow

## How It Works

1. Select one or more connected platforms
2. Attach a file and/or write a caption with hashtags
3. Click **Send**
4. The app loads the platform-specific workflow (`workflows/<platform>_upload.json`)
5. Template variables (`{{file}}`, `{{caption}}`, `{{hashtags}}`) are substituted
6. The upload is queued and status messages appear in the chat log

## Upload Workflows

Each platform has a JSON workflow file defining CDP steps:

```json
{
  "id": "instagram_upload",
  "platform": "instagram",
  "steps": [
    { "type": "navigate", "value": "https://www.instagram.com" },
    { "type": "click", "selector": "svg[aria-label='New post']" },
    { "type": "upload_file", "selector": "input[type='file']", "value": "{{file}}" },
    { "type": "type", "selector": "textarea", "value": "{{caption}}\n\n{{hashtags}}" }
  ]
}
```

Workflows are stored in the `workflows/` directory and can be customized.

## Message Types

| Color | Meaning |
|-------|---------|
| Gray (center) | System status |
| Purple (left) | Info/progress |
| Green (left) | Success |
| Red (left) | Error |

# Schedules

The Schedules view manages automated jobs for session keep-alive and content uploads.

## Job Types

### Session Keep-Alive

Automatically created when you import cookies for a platform. Periodically visits the platform to refresh session tokens.

- **Auto-created**: Yes, on successful cookie import
- **Interval**: 75% of shortest cookie expiry (min 12h, max 7d)
- **Execution**: Opens platform URL, waits, re-exports cookies to disk

### Upload

User-created jobs that upload content on a schedule.

- **Auto-created**: No, manual only
- **Interval**: User-defined cron expression
- **Execution**: Runs the platform's upload workflow with stored content

## Creating a Schedule

1. Click **Create Schedule** (top right)
2. Enter a job name
3. Select type: **Keep Alive** or **Upload**
4. Choose an interval (6h, 12h, 24h, 3d, or custom cron)
5. Select target platforms
6. For uploads: add caption and hashtags
7. Click **Create**

## Managing Schedules

Click any schedule card to see details:

- **Status**: Active, Paused, or Failed
- **Stats**: Run count, last run time, last result
- **Actions**: Pause/Resume, Delete

## Cron Expressions

| Expression | Meaning |
|-----------|---------|
| `0 */6 * * *` | Every 6 hours |
| `0 */12 * * *` | Every 12 hours |
| `0 0 * * *` | Daily at midnight |
| `0 0 */3 * *` | Every 3 days |
| `0 9 * * 1-5` | Weekdays at 9 AM |

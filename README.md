# CalmDrafts

*** Alpha - Use at your own risk! ***

A Go application that helps you manage your Gmail drafts by:
- Periodically notifying you about drafts in your Gmail
- Automatically deleting empty drafts that haven't been touched for a week (configurable)

## Features

- **Desktop Notifications**: Get notified about your draft count with details about empty drafts
- **Automatic Cleanup**: Removes old empty drafts (default: 7 days old)
- **Configurable**: Customize check intervals and cleanup thresholds
- **OAuth2 Authentication**: Secure authentication with Gmail API
- **Background Operation**: Runs continuously in the background

## Prerequisites

- Go 1.21 or later
- A Google Cloud Project with Gmail API enabled
- OAuth 2.0 credentials from Google Cloud Console

## Setup

### 1. Enable Gmail API and Get Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Gmail API:
   - Navigate to "APIs & Services" > "Library"
   - Search for "Gmail API"
   - Click "Enable"
4. Create OAuth 2.0 credentials:
   - Go to "APIs & Services" > "Credentials"
   - Click "Create Credentials" > "OAuth client ID"
   - Select "Desktop app" as the application type
   - Download the credentials JSON file
   - Rename it to `credentials.json` and place it in the project root

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure the Application

Copy the example configuration file:

```bash
cp config.json.example config.json
```

Edit `config.json` to customize settings:

```json
{
  "check_interval": "1h",        // How often to check drafts (e.g., "30m", "2h")
  "cleanup_age": "168h",          // Age threshold for deleting empty drafts (168h = 7 days)
  "credentials_path": "credentials.json",
  "token_path": "token.json"
}
```

Time format examples:
- `"30m"` = 30 minutes
- `"1h"` = 1 hour
- `"24h"` = 1 day
- `"168h"` = 7 days

### 4. Build the Application

```bash
go build -o calmdrafts ./cmd/calmdrafts
```

### 5. Run the Application

First run will prompt you to authorize the application:

```bash
./calmdrafts
```

You'll see a URL - open it in your browser, authorize the application, and paste the authorization code back into the terminal.

## Usage

### Run continuously (default mode)

```bash
./calmdrafts
```

The application will:
1. Check your drafts immediately
2. Send a desktop notification with draft count
3. Delete empty drafts older than the configured threshold
4. Repeat the check at the configured interval

### Run a single check

```bash
./calmdrafts -check
```

This performs one check and exits - useful for testing or running via cron.

### Custom configuration file

```bash
./calmdrafts -config /path/to/config.json
```

## Notifications

The application sends desktop notifications for:
- **Draft count**: "You have X draft(s) in your Gmail (Y empty)"
- **Cleanup actions**: "Deleted X old empty draft(s)"
- **Errors**: Notification when an error occurs

## What are "Empty Drafts"?

Empty drafts are draft emails with:
- No subject line
- No recipient (To field)
- No body content

These are typically created accidentally and can clutter your drafts folder.

## Security Notes

- `credentials.json` and `token.json` contain sensitive authentication data
- These files are excluded from git via `.gitignore`
- Keep these files secure and never share them
- The application only requests the minimum required Gmail API scopes:
  - `gmail.readonly`: To read draft information
  - `gmail.modify`: To delete empty drafts

## Troubleshooting

### "Error creating Gmail client"

Make sure `credentials.json` is in the correct location and is valid.

### "Unable to retrieve drafts"

Check that:
1. The Gmail API is enabled in your Google Cloud project
2. Your OAuth token hasn't expired (delete `token.json` and re-authorize)
3. You have an internet connection

### Notifications not appearing

On macOS, ensure the application has notification permissions:
- System Preferences > Notifications
- Find the terminal or application you're running from
- Enable notifications

## Development

Project structure:

```
calmdrafts/
├── cmd/calmdrafts/          # Main application
│   └── main.go
├── internal/
│   ├── config/              # Configuration management
│   │   └── config.go
│   ├── gmail/               # Gmail API client
│   │   └── client.go
│   └── notifier/            # Desktop notifications
│       └── notifier.go
├── config.json.example      # Example configuration
├── credentials.json         # OAuth credentials (not in git)
├── token.json              # OAuth token (not in git)
└── README.md
```

## License

MIT License - feel free to use and modify as needed.

## Contributing

Contributions welcome! Please feel free to submit issues or pull requests.

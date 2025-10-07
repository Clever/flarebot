# FlareBot

Slack bot that assists in the creation of FLARE support documents and incident management.

# Flarebot - Slack Bolt App Development Guide

This guide covers how to develop, edit, run, and debug the Flarebot Slack Bolt application.

## Overview

Flarebot is a Slack bot built with the Slack Bolt framework for controlling flares and incident management. It uses TypeScript for the main Slack app logic and Go for the flarebot-slack-cleanup Lambda function.

## Prerequisites

- Node.js 24+ (configured via `@tsconfig/node24`)
- npm
- Go 1.24+
- Slack App credentials (Bot Token, Signing Secret, App Token)
- Jira credentials
- Google API credentials
- PagerDuty API key

## Project Structure

```
.
├── src/                    # TypeScript source files
│   ├── app.ts             # Main Slack Bolt app entry point
│   ├── clients/           # External service clients
│   ├── lib/               # Utility libraries and configuration
│   │   ├── config.ts      # Environment configuration
│   │   ├── usersCache.ts  # Slack users caching
│   │   ├── channelsCache.ts # Slack channels caching
│   │   ├── jira.ts        # Jira integration utilities
│   │   ├── googleDocs.ts  # Google Docs integration
│   │   └── ...            # Other utility modules
│   ├── listeners/         # Slack event listeners
│   │   ├── actions/       # Interactive component actions
│   │   └── messages/      # Message event handlers
│   ├── middleware/        # Custom middleware
│   └── types/             # TypeScript type definitions
├── dist/                  # Compiled JavaScript output
├── cmd/                   # Go Lambda functions
│   └── flarebot-slack-cleanup/ # Slack channel cleanup Lambda
├── jira/                  # Go Jira integration
└── launch/                # Deployment configurations
```

## Installation

1. Clone the repository
2. Install Node.js dependencies:
   ```bash
   npm install
   ```
3. Install Go dependencies:
   ```bash
   make install_deps
   ```

## Configuration

The app requires the following environment variables (see `src/lib/config.ts`):

### Slack Configuration
- `SLACK_BOT_TOKEN` - Bot User OAuth Token
- `SLACK_SIGNING_SECRET` - Signing Secret for request verification
- `SLACK_APP_TOKEN` - App-level token for Socket Mode
- `SLACK_ORIGIN` - Slack instance URL

### Channel Configuration
- `FLARES_CHANNEL_ID` - Main flares channel ID
- `FLARES_CHANNEL_NAME` - Main flares channel name
- `FLARE_CHANNEL_PREFIX` - Prefix for flare-specific channels

### Jira Configuration
- `JIRA_ORIGIN` - Jira instance URL
- `JIRA_USERNAME` - Jira username
- `JIRA_PASSWORD` - Jira API token/password
- `JIRA_PROJECT_ID` - Jira project ID for flares
- `JIRA_SLACK_CHANNEL_FIELD_ID` - ID of the custom slack channel url field

### Google Configuration
- `GOOGLE_FLAREBOT_SERVICE_ACCOUNT_CONF` - Service account JSON
- `GOOGLE_DOMAIN` - Google Workspace domain
- `GOOGLE_TEMPLATE_DOC_ID` - Template document ID
- `GOOGLE_SLACK_HISTORY_DOC_ID` - Slack history spreadsheet ID
- `GOOGLE_FLARE_FOLDER_ID` - Google Drive folder for flare documents
- `GOOGLE_SHARED_DRIVE_ID` - Google Shared Drive ID

### PagerDuty Configuration
- `PAGERDUTY_API_KEY` - PagerDuty API key for alert integration

### Other Configuration
- `USERS_TO_INVITE` - Comma-separated list of users to auto-invite

## Development

### Building the TypeScript Code

```bash
make build-ts
```

This compiles TypeScript files from `src/` to JavaScript in `dist/`.

### Running the App Locally

```bash
ark start -l
```

This will:
1. Build the TypeScript code
2. Start the Slack Bolt app with Socket Mode
3. Connect to production Catapult service (when `_IS_LOCAL=true`)

### Running the Go Lambda Function Locally

```bash
ark start -l flarebot-slack-cleanup
```

This will:
1. Build the Go Lambda function
2. Run it locally with `IS_LOCAL=true`

### Code Formatting

The project uses Prettier for code formatting:

```bash
# Format only modified files
make format

# Format all TypeScript files
make format-all

# Check formatting without making changes
make format-check
```

### Linting

The project uses ESLint for code quality:

```bash
# Run ESLint
make lint-es

# Run ESLint with auto-fix
make lint-fix

# Run both format check and ESLint
make lint
```

### Testing

#### Manual testing in Slack

Run the command `ark start -l` to start an instance on your local machine.

**Note:** If there is an instance running in clever-dev you will need to stop that instance by running a command like:

```
ark stop -e clever-dev flarebot
```

After startup, your app may spin for a few moments as it establishes a connection to Slack. (See Slack bolt app documentation to learn more about the app startup flow.)

After spinning for a few moments, you should see a message like the below:

```
Thu 18:02:51.854 flarebot/local> [INFO]  bolt-app ⚡️ Bolt app is running!
```

After your app has started successfully, you should be able to message your local bot from Slack by:

1. Going to a Slack channel that has @flarebot-dev as a member.
2. Message a simple message like `@flarebot-dev help`

Confirm you get a response from the bot. Your local console will aslo show the request with log lines like the below:

```
Thu 17:54:21.828 flarebot/local> users 1821
Thu 17:54:21.828 flarebot/local> channels {}
Thu 17:54:22.195 flarebot/local> {
	source.title=flarebot.request-finished
	channel-id:"C09RZS9QV"
	payload:map[blocks:[map[block_id:GqDWy elements:[map[elements:[map[type:user user_id:U090FAF1X9S] map[text: help type:text]] type:rich_text_section]] type:rich_text]] channel:C09RZS9QV channel_type:channel client_msg_id:e1fa4d88-b587-4464-a772-051af8e7564c event_ts:1754009660.250419 team:T027Y5L9T text:<@U090FAF1X9S> help ts:1754009660.250419 type:message user:U07EQUB5C2X]
	pod-account:"589690932525"
	pod-id:"local"
	pod-region:"us-west-2"
	pod-shortname:"local"
	response-time-ms:221
	status-code:200
	team:"eng-infra"
	user-id:"U07EQUB5C2X"
}
```

**Note:** If there are other dev instances of flarebot connected, you may not receive messages.

#### Run Jest tests

```bash
# Run all JavaScript/TypeScript tests
make test-js

# Run a specific test file
make src/listeners/messages/fireFlare.test.ts
```

#### Run Go tests

```bash
make test
```

## Debugging

### Debug Middleware

The app includes debug middleware in `src/app.ts:37-45` that logs incoming payloads. Uncomment this section to enable request debugging:

```typescript
app.use(async ({ next, payload, body, context }) => {
  console.log("payload", payload);
  console.log("body", body);
  console.log("users", context.usersCache.users.length);
  console.log("channels", context.channelsCache.channels);
  await next();
});
```

### Socket Mode

The app runs in Socket Mode, which is ideal for development as it doesn't require a public URL. The connection is established using the `SLACK_APP_TOKEN`.

### Test Commands

Test individual components using the provided CLI tools:

```bash
# Build test CLIs
make build

# Test Jira integration
./bin/jira-cli

# Test Slack integration
./bin/slack-cli
```

### Common Debugging Tips

1. **Check Environment Variables**: Ensure all required environment variables are set. The app will throw errors on startup if any are missing (except in test mode).

2. **Enable Debug Logging**: Uncomment the debug middleware to see all incoming Slack events and payloads.

3. **Test in Isolation**: Use the component-specific test CLIs to debug individual integrations.

4. **Monitor Caches**: The app caches Slack users (updated every 24 hours) and channels. Check cache state in the debug logs.

5. **Socket Mode Connection**: If the app isn't receiving events, verify:
   - `SLACK_APP_TOKEN` is valid and starts with `xapp-`
   - Socket Mode is enabled in your Slack app configuration
   - The app has the necessary OAuth scopes

## Deployment

The app is configured for deployment with:
- `Dockerfile` for containerization
- `launch/flarebot.yml` for main Slack bot deployment
- `launch/flarebot-slack-cleanup.yml` for Lambda function deployment
- `cmd/flarebot-slack-cleanup/` contains the Go Lambda function

## Key Features

- **Fire Flare**: Create new incident channels and Jira tickets
- **Flare Transitions**: Manage flare states (mitigate, not a flare, unmitigate)
- **Role Management**: Assign incident leads and communication leads
- **Recent Deploys**: Track and display recent deployments
- **PagerDuty Integration**: Show open alerts and recent critical alerts
- **Google Docs Integration**: Create incident documents from templates
- **Slack History Tracking**: Log messages to Google Sheets
- **Channel Cleanup**: Automated archiving of old flare channels via Lambda

## Architecture Notes

- The app uses Slack's Bolt framework with Socket Mode for real-time events
- Middleware pattern for request processing and context injection
- Caching layer for Slack users and channels to reduce API calls
- TypeScript for type safety and better developer experience
- Integration with external services (Jira, Google, PagerDuty) via dedicated modules
- Background tasks for file uploads and cache updates
- Go Lambda function for automated channel cleanup

## Contributing

1. Create a feature branch
2. Make your changes
3. Run `make lint` to ensure code quality
4. Run `make test-js` for TypeScript tests and `make test` for Go tests
5. Submit a pull request

## Troubleshooting

### App Not Responding to Messages
- Check if the app is running: Look for "⚡️ Bolt app is running!" in logs
- Verify the app is invited to the channel
- Check Socket Mode connection status

### Environment Variable Errors
- All environment variables are required except in test mode (`NODE_ENV=test`)
- Double-check the variable names match exactly as defined in `src/lib/config.ts`

### Build Errors
- Ensure you're using Node.js 24+
- Run `npm install` to update dependencies
- Clear the `dist/` directory and rebuild: `rm -rf dist && make build-ts`
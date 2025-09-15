# flarebot-slack-cleanup

cron job for cleaning up inactive flare slack channels

Owned by eng-infra

## Deploying

```
ark start flarebot-slack-cleanup -e production 
```

### Running locally

#### Option 1: Using ark
```bash
ark start flarebot-slack-cleanup -l
```

#### Option 2: Without ark
1. Set up environment variables
- `JIRA_ORIGIN` - Jira instance URL
- `JIRA_USERNAME` - Jira username
- `JIRA_PASSWORD` - Jira API token/password
- `JIRA_PROJECT_ID` - Jira project ID for flares
- `SLACK_BOT_TOKEN` - Bot User OAuth Token
- `FLARE_CHANNEL_PREFIX` - [optional] Prefix for flare-specific channels. Defaults to flaretest-
- `CHANNEL_AGE_THRESHOLD` - [optional] Channels older than this threshold will be archived. Defaults to 180 days
- `DRY_RUN` - [optional] Flag to run the script in dry-run mode. The script will output all channels that would be archived but does not perform the action. Defaults to true
2. Install dependencies and build
```bash
make install_deps
make build
```
2. Run the script locally 
```bash
go run cmd/flarebot-slack-cleanup/main.go  
```

# Flarebot TypeScript Slack Bolt App Architecture

## Overview

The app is built using Slack's Bolt framework for TypeScript and follows a modular architecture with clear separation of concerns. The main Slack bot functionality is implemented in TypeScript, while Go is used for the flarebot-slack-cleanup Lambda function.

## App Entry Point

- Main file: src/app.ts
- Initializes a Slack Bolt app with socket mode enabled
- Configures global middleware for:
  - Injecting shared clients (Jira, Google Drive/Sheets, PagerDuty, Catapult)
  - Managing user and channel caches
  - Logging with Kayvee
- Runs background tasks for file uploads and user cache updates

## Command Architecture

## Message Listeners

Commands are triggered via regex patterns when users mention the bot. Located in src/listeners/messages/:

1. Fire Flare Command (fireFlare.ts)
  - Pattern: /fire\s+(?:a\s+)?(?:flare\s+)?...
  - Triggers when: @flarebot fire a flare p0 Database is down
  - Creates Jira ticket, Google Docs, Slack channel
  - Supports priorities (p0/p1/p2) and types (preemptive/retroactive)
2. Flare Transition Command (flareTransition.ts)
  - Pattern: /(?:flare )?(?:is )?(mitigat(?:ed|e)|not (?:a )?flare|unmitigat(?:ed|e))/i
  - Triggers when: @flarebot mitigated, @flarebot not a flare
  - Updates Jira ticket status and notifies channels
3. Help Command (help.ts)
  - Pattern: /help\s*(all)?/i
  - Triggers when: @flarebot help
  - Provides context-aware help based on channel type
4. Comms Lead Command (commsLead.ts)
  - Manages communication lead assignments for flares
5. Incident Lead Command (incidentLead.ts)
  - Manages incident lead assignments for flares

## Action Listeners

Interactive components are handled in src/listeners/actions/:

1. Debugging101 (debugging101.ts) - Provides debugging resources
2. Flare Followups (flareFollowups.ts) - Manages follow-up actions
3. Flare Roles (flareRoles.ts) - Manages role assignments
4. How To Page (howToPage.ts) - Provides help documentation
5. Open Alerts (openAlerts.ts) - Shows open PagerDuty alerts
6. Recent Critical Alerts (recentCriticalAlerts.ts) - Shows recent critical alerts
7. Recent Deploys (recentDeploys.ts) - Shows recent deployments
8. What's A Flare (whatsAFlare.ts) - Explains what a flare is

## Middleware Pipeline

1. Message Middleware (src/middleware/message.ts):
  - Records all messages to Google Sheets for flare channels
  - Validates bot mentions and channel permissions
  - Enriches context with user/channel data
2. Block Action Middleware (src/middleware/blockAction.ts):
  - Handles interactive button clicks

## External Service Clients

Located in src/clients/index.ts:
- Jira Client (jira.js) - Creates and manages Jira tickets
- Google Drive Client - Creates and manages Google Docs
- Google Sheets Client - Records Slack message history
- PagerDuty Client - Fetches alert information
- Catapult Client - Service discovery and communication

## Go Lambda Function

The flarebot-slack-cleanup Lambda function (cmd/flarebot-slack-cleanup/):
- Archives old flare channels based on age threshold
- Updates Jira tickets with archived labels
- Runs as a scheduled Lambda function
- Uses Go with AWS Lambda runtime

## Command Flow

1. User mentions bot with command in allowed channel
2. Message middleware validates and enriches context
3. Regex patterns match to specific handlers
4. Handler executes business logic (create tickets, channels, etc.)
5. Bot responds in thread or channel

## Key Design Patterns

- Regex-based routing: Each command has a regex pattern for flexible matching
- Context enrichment: Middleware adds user/channel data before handlers
- Error handling: Try-catch blocks with fallback messages
- Caching: User and channel data cached to reduce API calls
- Integration-heavy: Integrates with Jira, Google Drive/Sheets, PagerDuty, and Slack APIs
- Background tasks: File uploads and cache updates run asynchronously
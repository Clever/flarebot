# FlareBot

Slack bot that assists in the creation of FLARE support documents.

## Issues & Feature Requests

Flarebot issues and feature requests should be added to this Github project's issues list.

## Configuration

Flarebot is an OAuth client into Slack, into Google Docs, and an
HTTP-auth client into JIRA. It needs a lot of configuration.

### Slack

Flarebot needs to be a full user (not a bot), which means you need to
have an OAuth app set up to at least get a token. The following
environment variables are expected.

* `SLACK_DOMAIN`: Your team's Slack domain, e.g. `https://<team>.slack.com`
* `SLACK_USERNAME`: The username you configured the user for in Slack, e.g. `flarebot`
* `SLACK_CLIENT_ID`: Slack OAuth App client ID
* `SLACK_CLIENT_SECRET`: Slack OAuth App client secret
* `SLACK_FLAREBOT_ACCESS_TOKEN`: Slack OAuth access token for the Flarebot user
* `SLACK_CHANNEL`: the Channel ID where Flarebot should be listening

### Google

Flarebot needs to be a full user (not a service app), which means you
need to have an OAuth app set up to at least get a token. The
following environment variables are expected:

* `GOOGLE_CLIENT_ID`: Google OAuth app client ID
* `GOOGLE_CLIENT_SECRET`: Google OAuth app client secret
* `GOOGLE_FLAREBOT_ACCESS_TOKEN`: Google OAuth access token for the Flarebot user
* `GOOGLE_TEMPLATE_DOC_ID`: the Google Doc ID for the template to copy as the Facts Doc.

### JIRA

JIRA is accessed using HTTP Basic Auth, which means you need a JIRA
user and you really should run JIRA over SSL. The following
environment variables are expected:

* `JIRA_ORIGIN`: the web origin where JIRA lives, e.g. `https://<company>.atlassian.net`
* `JIRA_USERNAME` and `JIRA_PASSWORD`: login for JIRA for Flarebot
* `JIRA_PRIORITIES`: a comma-separated list of IDs for the priorities P0, P1, P2, in that order.
* `JIRA_PROJECT_ID`: the JIRA project ID where the ticket should be added
* `JIRA_ISSUETYPE_ID`: the JIRA issue type ID for the ticket, usually the one that corresponds to `Bug`.


## Usage

```
@flarebot: fire a flare p2 District 9 users cannot log in 
@channel: OK, go chat in #flare-4242
```

In #flare-4242, @flarebot will:
* set the topic
* post a link to the JIRA ticket it created
* post a link to the Facts Google Doc it created

## Future Features (Maybe)

In the specific channel:

```
@flarebot: incident lead is @z
OK, got that

@flarebot: comms lead is @ben
OK, got that

@flarebot: at 10:45am, we see an increase in error rates in oauth service
OK, logged that to the Facts Doc

@flarebot: right now, we see a decrease in error rates
OK, logged that at 10:48am to the Facts Doc
```

## Trickiness

Initially we thought we would use a new "slash" command in Slack,
e.g. `/fire_flare`, but those integrations are enabled in all rooms,
which doesn't make sense, and they require webhooks, which makes
development quite a bit harder, so this isn't worthwhile for now.

Both Google and Slack APIs require full users, not just a Slack bot
user for example, to do the things we want to do.

# Deployment

On Heroku

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

If you set up Flarebot as a normal Google account, the best you can get is an OAuth token
that expires after a year. To get a forever-token, to match best-practices, and to not have
to do an OAuth dance, you should set up Flarebot as a [Google Service Account](https://developers.google.com/identity/protocols/OAuth2ServiceAccount#creatinganaccount).

When you generate such an account, Google gives you a JSON-formatted set of service account
configuration parameters. You'll need this JSON blob as a configuration parameter to Flarebot.

The following environment variables are expected:

* `GOOGLE_DOMAIN`: the domain name of your organization that Flarebot documents will be shared with, e.g. `clever.com`
* `GOOGLE_CLIENT_ID`: Google OAuth app client ID
* `GOOGLE_CLIENT_SECRET`: Google OAuth app client secret
* `GOOGLE_FLAREBOT_SERVICE_ACCOUNT_CONF`: Google Service Account JSON configuration blob
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

You can test the jira library in isolation by setting the above environment variables and then running:
```
make build
./bin/jira-cli --help
```


### Documentation

Flarebot provides links to documentation when a Flare is fired. This link is configured as

* `FLARE_RESOURCES_URL`: a URL of Flare-handling resources, checklists, etc.

## Usage

### Help

```
@flarebot: help
```

Lists all commands

### Fire a Flare

```
@flarebot: fire a flare p2 District 9 users cannot log in 
@channel: OK, go chat in #flare-4242
```

In #flare-4242, @flarebot will:
* set the topic
* post a link to the JIRA ticket it created, assigning the reporter.
* post a link to the Facts Google Doc it created
* post a link to the Flare Resources page.


### Declaring Incident Lead

Within the Flare-specific channel:

```
@flarebot: I am incident lead
OK, @ben is incident lead
```

### Declaring not a Flare or Flare mitigated

Within the Flare-specific channel:

```
@flarebot: not a flare
```

```
@flarebot: flare is mitigated
```


## Future Features (Maybe)

In the specific channel:

```
@flarebot: I am comms lead
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

## Tech Design

Ideally, the Flarebot process is stateless, looking up state in JIRA
and Slack. This is relatively easy for interactions in the main Flare
channel, which is stable and can be referenced by a config
parameter. It gets a little bit harder for:

* having Flarebot know when to respond in single-Flare-specific channels
* accumulating state during a Flare and transferring it to a Google doc.

For the first problem, the approach we'll take is:

* ensure that flare channels are named the same as JIRA ticket id.
* if Flarebot receives a command it recognizes, it will check the channel name against JIRA and ensure it is a ticket in the Flares Project.

# Deployment

Procfile provided for deployment on e.g. Heroku.

# FlareBot

Slack bot that assists in the creation of FLARE support documents.

## Configuration

* `SLACK_TOKEN`: Your Slack token, found by making a Hubot integration in Slack
* `SLACK_DOMAIN`: Your team's Slack domain, e.g. `https://<team>.slack.com`
* `SLACK_USERNAME`: The username you configured the bot for in Slack, e.g. `redbull`

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
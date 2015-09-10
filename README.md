# redbullbot

Slack bot that can be informed + track redbull consumption, for an individual or a team.

It requires three environment variables:

* `SLACK_TOKEN`: Your Slack token, found by making a Hubot integration in Slack
* `SLACK_DOMAIN`: Your team's Slack domain, e.g. `https://<team>.slack.com`
* `SLACK_USERNAME`: The username you configured the bot for in Slack, e.g. `redbull`

Then you can message it with or without your team:

```
@redbull rack me up
@redbull rack one up
@redbull rack me up for the api team
@redbull rack me up for the API team
@redbull rack one up for the API team
```

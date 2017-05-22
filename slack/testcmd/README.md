# slack-cli
To more easily test the slack subpackage, this directory creates an executable called `slack-cli` that you can use to access some of the slack functions directly.

## usage

First set the following environment variables:
- **SLACK_LEGACY_TOKEN** (omit of you are using an oauth account)
- **SLACK_DOMAIN**
- **SLACK_USERNAME**
- **SLACK_FLAREBOT_ACCESS_TOKEN**

```
> slack-cli --help
Usage:
  slack-cli [command]

Available Commands:
  createChannel create a channel
  help          Help about any command
  pinMessage    pin the message matching the text
  postMessage   post a message to a channel

Use "slack-cli [command] --help" for more information about a command.
```

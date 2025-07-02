const { helpAll } = require("../lib/help");
const kayvee = require("kayvee");
const { Version3Client } = require("jira.js");

var logger = new kayvee.logger("flarebot");

const jiraClient = new Version3Client({
  host: process.env.JIRA_ORIGIN,
  authentication: {
    basic: {
      email: process.env.JIRA_USERNAME,
      apiToken: process.env.JIRA_PASSWORD,
    },
  },
});

async function middleware({ payload, client, context, next, say }) {
  if (!payload.text.includes(`<@${context.botUserId}>`)) {
    return;
  }

  context.jiraClient = jiraClient;
  context.logger = logger;

  const now = new Date();
  try {
    const userInfo = await client.users.info({
      user: payload.user,
    });
    context.user = userInfo.user;

    const channelInfo = await client.conversations.info({
      channel: payload.channel,
    });
    context.channel = channelInfo.channel;

    if (
      channelInfo.channel.name === process.env.FLARES_CHANNEL_NAME ||
      channelInfo.channel.name.startsWith(process.env.FLARE_CHANNEL_PREFIX)
    ) {
      await next();
      context.logger.infoD("request-finished", {
        payload: payload,
        "response-time-ms": new Date() - now,
        "channel-name": context.channel.name,
        "user-name": context.user.real_name,
        "status-code": 200,
      });
    } else {
      await say(
        `Sorry! I can't help you with that. I am only allowed to reply to messages in the <#${process.env.FLARES_CHANNEL_ID}> channel or a flare channel. ${helpAll()}`,
      );
      context.logger.infoD("request-finished", {
        payload: payload,
        "response-time-ms": new Date() - now,
        "channel-name": context.channel.name,
        "user-name": context.user.real_name,
        "status-code": 400,
      });
    }
  } catch (error) {
    context.logger.errorD("request-finished", {
      payload: payload,
      "response-time-ms": new Date() - now,
      "channel-name": context.channel ? context.channel.name : "unknown",
      "user-name": context.user ? context.user.real_name : "unknown",
      "status-code": 500,
      error: error,
    });
    await say(`Sorry! I'm having trouble processing your request. ${error}`);
  }
}

module.exports = { middleware };

import { helpAll } from "../lib/help";
import config from "../lib/config";
import { AllMiddlewareArgs, SlackEventMiddlewareArgs } from "@slack/bolt";

const messageMiddleware = async ({
  payload,
  client,
  context,
  next,
}: AllMiddlewareArgs & SlackEventMiddlewareArgs<"message">) => {
  // we don't care about all the subtypes. We only care about generic message events.
  if (payload.type !== "message" || payload.subtype !== undefined) {
    await next();
    return;
  }

  // this middleware is only interested in messages that mention the bot.
  if (payload.text && !payload.text.includes(`<@${context.botUserId}>`)) {
    await next();
    return;
  }

  const now = new Date();
  try {
    if (!payload.user || !payload.channel) {
      await client.chat.postMessage({
        channel: payload.channel ?? "",
        text: `Sorry! Missing user or channel information in the event payload.`,
      });
      return;
    }
    const userInfo = await client.users.info({
      user: payload.user,
    });
    context.user = userInfo.user;

    const channelInfo = await client.conversations.info({
      channel: payload.channel,
    });
    if (!channelInfo.channel || !channelInfo.channel.name) {
      await client.chat.postMessage({
        channel: payload.channel,
        text: `Sorry! Missing channel information.`,
      });
      return;
    }
    context.channel = channelInfo.channel;

    if (
      channelInfo.channel.name === config.FLARES_CHANNEL_NAME ||
      channelInfo.channel.name.startsWith(config.FLARE_CHANNEL_PREFIX)
    ) {
      await next();
      context.logger.infoD("request-finished", {
        payload: payload,
        "response-time-ms": new Date().getTime() - now.getTime(),
        "channel-id": context.channel.id,
        "user-id": context.user.id,
        "status-code": 200,
      });
    } else {
      await client.chat.postMessage({
        channel: payload.channel,
        text: `Sorry! I can't help you with that. I am only allowed to reply to messages in the <#${config.FLARES_CHANNEL_ID}> channel or a flare channel. ${helpAll(context.botUserId)}`,
      });
      context.logger.infoD("request-finished", {
        payload: payload,
        "response-time-ms": new Date().getTime() - now.getTime(),
        "channel-id": context.channel.id,
        "user-id": context.user.id,
        "status-code": 400,
      });
    }
  } catch (error) {
    context.logger.errorD("request-finished", {
      payload: payload,
      "response-time-ms": new Date().getTime() - now.getTime(),
      "channel-id": context.channel && context.channel.id ? context.channel.id : "unknown",
      "user-id": context.user && context.user.id ? context.user.id : "unknown",
      "status-code": 500,
      error: error,
    });
    await client.chat.postMessage({
      channel: payload.channel ?? "",
      text: `Sorry! I'm having trouble processing your request. ${error}`,
    });
  }
};

export { messageMiddleware };

import { helpAll } from "../lib/help";
import config from "../lib/config";
import { AllMiddlewareArgs, Context, SlackEventMiddlewareArgs } from "@slack/bolt";
import { AllMessageEvents } from "@slack/types";
import { WebClient } from "@slack/web-api";
import { ChannelsCache } from "../lib/channelsCache";

const messageMiddleware = async ({
  payload,
  client,
  context,
  next,
}: AllMiddlewareArgs & SlackEventMiddlewareArgs<"message">) => {
  if (payload.type !== "message") {
    await next();
    return;
  }

  try {
    await recordMessage(payload, context, client);
  } catch (error) {
    // we don't want to block the main flow if we fail to record message history.
    context.logger.errorD("record-message-error", {
      payload: payload,
      error: error,
    });
  }

  // we don't care about all the subtypes. We only care about generic message events
  // in future we could consider adding support for message_changed if requested
  if (payload.subtype !== undefined) {
    return;
  }

  // flarebot only cares about messages that mention the bot.
  if (payload.text && !payload.text.includes(`<@${context.botUserId}>`)) {
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

    context.user = await context.usersCache.getUser(client, payload.user);
    context.channel = await context.channelsCache.getChannel(client, payload.channel);

    if (
      context.channel.name === config.FLARES_CHANNEL_NAME ||
      context.channel.name.startsWith(config.FLARE_CHANNEL_PREFIX)
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

async function recordMessage(payload: AllMessageEvents, context: Context, client: WebClient) {
  const channelsCache = context.channelsCache as ChannelsCache;

  const channelHistoryDocId = await channelsCache.getChannelHistoryDocId(
    client,
    payload.channel,
    context.botUserId ?? "",
  );

  if (!channelHistoryDocId) {
    return;
  }

  // there is no point in tracking every single message event. Lets do our best to track the most important ones.
  let message = "";
  let user = "";
  if (payload.subtype === undefined) {
    message = payload.text || "";
    // this handles a bug in the slack api described here https://api.slack.com/events/message/message_replied
    if (payload.thread_ts) {
      message = `(message_replied ${payload.thread_ts}) ${message}`;
    }
    user = payload.user;
  } else if (payload.subtype === "message_replied") {
    if ("text" in payload.message && payload.message.text) {
      message = `(message_replied ${payload.message.thread_ts}): ${payload.message.text}`;
    }
    if ("user" in payload.message && payload.message.user) {
      user = payload.message.user;
    }
  } else if (payload.subtype === "message_changed") {
    if (
      "text" in payload.message &&
      payload.message.text &&
      "text" in payload.previous_message &&
      payload.previous_message.text &&
      payload.message.text !== payload.previous_message.text
    ) {
      message = `(message_changed ${payload.message.ts}): ${payload.message.text}`;
    }
    if ("user" in payload.message && payload.message.user) {
      user = payload.message.user;
    }
  } else if (payload.subtype === "message_deleted") {
    message = `(message_deleted ${payload.previous_message.ts}): Message deleted`;
    if ("user" in payload.previous_message && payload.previous_message.user) {
      user = payload.previous_message.user;
    }
  } else if (payload.subtype === "channel_join" || payload.subtype === "channel_leave") {
    message = payload.text || "";
    user = payload.user;
  }
  if (message === "" || user === "") {
    context.logger.debugD("message-not-recorded", { payload: payload });
    return;
  }

  // Replace Slack user mentions with actual names in the message
  const mentionMatches = message.match(/<@([A-Z0-9]+)>/g);
  if (mentionMatches) {
    for (const match of mentionMatches) {
      const userId = match.match(/<@([A-Z0-9]+)>/)?.[1];
      if (userId) {
        const user = await context.usersCache.getUser(client, userId);
        const userName = user ? `@${user.real_name || user.name || userId}` : match;
        message = message.replace(match, userName);
      }
    }
  }

  const author = await context.usersCache.getUser(client, user);

  await context.clients.googleSheetsClient.spreadsheets.values.append({
    spreadsheetId: channelHistoryDocId,
    range: "Sheet1",
    valueInputOption: "USER_ENTERED",
    insertDataOption: "INSERT_ROWS",
    requestBody: {
      values: [
        [
          payload.ts,
          new Date(parseInt(payload.ts) * 1000).toLocaleString("en-US", { timeZone: "US/Pacific" }),
          author ? `${author.real_name || author.name || author.id}` : user,
          message,
        ],
      ],
      majorDimension: "ROWS",
    },
  });
}

export { messageMiddleware };

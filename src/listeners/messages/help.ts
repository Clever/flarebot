import { AllMiddlewareArgs, SlackEventMiddlewareArgs } from "@slack/bolt";
import config from "../../lib/config";
import { helpAll, helpFlaresChannel, helpFlareChannel } from "../../lib/help";
import { setListenerMatch } from "../../lib/listenerMatch";

const helpRegex = /help\s*(all)?/i;

async function help({
  client,
  message,
  context,
}: AllMiddlewareArgs & SlackEventMiddlewareArgs<"message">) {
  setListenerMatch(context);
  if (message.subtype !== undefined && message.subtype !== "bot_message") {
    return;
  }

  const isAll = message.text?.match(helpRegex)?.[1] === "all";
  let helpText;
  if (isAll) {
    helpText = helpAll(context.botUserId);
  } else if (context.channel.name === config.FLARES_CHANNEL_NAME) {
    helpText = helpFlaresChannel(context.botUserId);
  } else if (context.channel.name.startsWith(config.FLARE_CHANNEL_PREFIX)) {
    helpText = helpFlareChannel(context.botUserId);
  } else {
    helpText = helpAll(context.botUserId);
  }

  await client.chat.postMessage({
    channel: context.channel.id,
    thread_ts: message.ts,
    text: helpText,
  });
}

export { helpRegex, help };

import { AllMiddlewareArgs, SlackEventMiddlewareArgs } from "@slack/bolt";
import config from "../../lib/config";
import { helpAll, helpFlaresChannel, helpFlareChannel } from "../../lib/help";

const helpRegex = /help\s*(all)?/i;

async function help({
  message,
  context,
  say,
}: AllMiddlewareArgs & SlackEventMiddlewareArgs<"message">) {
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
  await say({
    text: helpText,
  });
}

export { helpRegex, help };

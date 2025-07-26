import { AllMiddlewareArgs, SlackEventMiddlewareArgs } from "@slack/bolt";
import { Version3Client } from "jira.js";
import { helpFlareChannel } from "../../lib/help";
import { doJiraTransition } from "../../lib/jira";
import config from "../../lib/config";

const flareTransitionRegex =
  /(?:flare )?(?:is )?(mitigat(?:ed|e)|not (?:a )?flare|unmitigat(?:ed|e))/i;

async function flareTransition({
  client,
  message,
  say,
  context,
}: AllMiddlewareArgs & SlackEventMiddlewareArgs<"message">) {
  const jiraClient = context.clients.jiraClient as Version3Client;

  if (message.subtype !== undefined) {
    return;
  }

  const transition = extractFlareTransition(message.text ?? "");

  if (!transition) {
    await say({
      text: `Sorry! I couldn't extract the flare transition from your message. ${helpFlareChannel(context.botUserId)}`,
    });
    return;
  }

  const jiraTicket = context.channel.name;
  const channelId = context.channel.id;

  try {
    if (transition === "mitigated" || transition === "mitigate") {
      await doJiraTransition(jiraClient, jiraTicket, "Mitigated");
    } else if (transition === "not a flare" || transition === "not flare") {
      await doJiraTransition(jiraClient, jiraTicket, "NotAFlare");
    } else if (transition === "unmitigated" || transition === "unmitigate") {
      await doJiraTransition(jiraClient, jiraTicket, "In Progress");
    }
  } catch (error) {
    throw new Error(
      "Error transitioning flare: " + error + ". You can manually transition the flare in Jira.",
    );
  }

  let responseText = "";
  let flareChannelText = "";
  if (transition === "mitigated" || transition === "mitigate") {
    responseText = "The Flare was mitigated and there was much rejoicing throughout the land.";
    flareChannelText = `<#${channelId}> has been mitigated`;
  } else if (transition === "not a flare" || transition === "not flare") {
    responseText = "The Flare was not a Flare and there was much rejoicing throughout the land.";
    flareChannelText = `turns out <#${channelId}> is not a Flare`;
  } else if (transition === "unmitigated" || transition === "unmitigate") {
    responseText = "UhOh! The Flare was unmitigated and the land is in chaos.";
    flareChannelText = `<!channel> <#${channelId}> has been unmitigated and is back in progress.`;
  }
  await client.chat.postMessage({
    channel: channelId,
    thread_ts: message.ts,
    text: responseText,
  });
  await client.chat.postMessage({
    channel: config.FLARES_CHANNEL_ID,
    text: flareChannelText,
  });
}

function extractFlareTransition(text: string) {
  const matches = text.toLowerCase().match(flareTransitionRegex);
  if (!matches) return null;
  return matches[1];
}

export { flareTransition, flareTransitionRegex, extractFlareTransition };

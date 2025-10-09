import { AllMiddlewareArgs, SlackEventMiddlewareArgs } from "@slack/bolt";
import { Version3Client } from "jira.js";
import { helpFlareChannel } from "../../lib/help";
import { doJiraTransition } from "../../lib/jira";
import config from "../../lib/config";
import { setListenerMatch } from "../../lib/listenerMatch";

const flareTransitionRegex =
  /(?:flare )?(?:is )?(mitigat(?:ed|e)|not (?:a )?flare|unmitigat(?:ed|e))/i;

async function flareTransition({
  client,
  message,
  say,
  context,
}: AllMiddlewareArgs & SlackEventMiddlewareArgs<"message">) {
  setListenerMatch(context);
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
    const jiraUser = await jiraClient.userSearch.findUsers({
      query: context.user.profile.email,
    });

    if (!jiraUser || jiraUser.length === 0) {
      throw new Error("Could not find your JIRA user account");
    }

    if (transition === "mitigated" || transition === "mitigate") {
      await doJiraTransition(jiraClient, jiraTicket, "Mitigated", jiraUser[0].accountId);
    } else if (transition === "not a flare" || transition === "not flare") {
      await doJiraTransition(jiraClient, jiraTicket, "NotAFlare", jiraUser[0].accountId);
    } else if (transition === "unmitigated" || transition === "unmitigate") {
      await doJiraTransition(jiraClient, jiraTicket, "In Progress", jiraUser[0].accountId);
    }
  } catch (error) {
    throw new Error(
      "Error transitioning flare: " + error + ". You can manually transition the flare in Jira.",
    );
  }

  let responseText = "";
  let flareChannelText = "";
  let mitigated = false;
  let unmitigated = false;
  if (transition === "mitigated" || transition === "mitigate") {
    responseText = "The Flare was mitigated and there was much rejoicing throughout the land.";
    flareChannelText = `<#${channelId}> has been mitigated`;
    mitigated = true;
  } else if (transition === "not a flare" || transition === "not flare") {
    responseText = "The Flare was not a Flare and there was much rejoicing throughout the land.";
    flareChannelText = `turns out <#${channelId}> is not a Flare`;
  } else if (transition === "unmitigated" || transition === "unmitigate") {
    responseText = "UhOh! The Flare was unmitigated and the land is in chaos.";
    flareChannelText = `<!channel> <#${channelId}> has been unmitigated and is back in progress.`;
    unmitigated = true;
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

  const followupMessage = `can you please <https://clever.atlassian.net/wiki/spaces/ENG/pages/108210465/Flare+Followups#Step-1%3A-Signing-up-for-flare-followup|sign up> for followup for tomorrow? Fill out the <https://clever.atlassian.net/browse/${jiraTicket}| jira ticket> to capture what we know, following instructions <https://clever.atlassian.net/wiki/spaces/ENG/pages/108210465/Flare+Followups#Step-2%3A-Updating-the-flare-ticket|here>.`;

  if (mitigated) {
    const mitigationTime = new Date(parseInt(message.ts) * 1000);
    const day = mitigationTime.getUTCDay();
    const hour = mitigationTime.getUTCHours();

    if (day == 4 && hour >= 15) {
      // respond immediately in thread for mitigation between thursday 3 pm UTC and 11:59 pm UTC. (i.e approximately US business hours)
      await client.chat.postMessage({
        channel: channelId,
        thread_ts: message.ts,
        text: `<@${context.user.id}> ` + followupMessage,
      });
    } else {
      // Schedule a message for Thursday 3 pm UTC
      const nextThursday8am = new Date(mitigationTime);
      nextThursday8am.setUTCDate(
        nextThursday8am.getUTCDate() + ((4 - nextThursday8am.getUTCDay() + 7) % 7),
      );
      nextThursday8am.setHours(15, 0, 0, 0);

      await client.chat.scheduleMessage({
        channel: channelId,
        post_at: nextThursday8am.getTime() / 1000,
        text: `<@${context.user.id}> if you haven't already, ` + followupMessage,
      });
    }
  } else if (unmitigated) {
    const msgList = await client.chat.scheduledMessages.list({
      channel: channelId,
    });
    for (const msg of msgList.scheduled_messages ?? []) {
      await client.chat.deleteScheduledMessage({
        channel: channelId,
        scheduled_message_id: msg.id ?? "",
      });
    }
  }
}

function extractFlareTransition(text: string) {
  const matches = text.toLowerCase().match(flareTransitionRegex);
  if (!matches) return null;
  return matches[1];
}

export { flareTransition, flareTransitionRegex, extractFlareTransition };

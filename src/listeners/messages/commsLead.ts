import { AllMiddlewareArgs, SlackEventMiddlewareArgs } from "@slack/bolt";
import { helpFlareChannel } from "../../lib/help";
import config from "../../lib/config";
import { Version3Client } from "jira.js";

const commsLeadRegex = /(?:i am|i'm)(?:\s+(?:the|a))?\s+comms lead/i;

async function commsLead({
  client,
  message,
  say,
  context,
}: AllMiddlewareArgs & SlackEventMiddlewareArgs<"message">) {
  if (message.subtype !== undefined) {
    return;
  }

  const channelId = context.channel.id;
  const userId = message.user;

  try {
    const jiraUser = await jiraClient.userSearch.findUsers({
      query: userId,
    });

    if (jiraUser.length === 0) {
      throw new Error("User not found in JIRA");
    }

    const jiraClient = context.clients.jiraClient as Version3Client;
    await jiraClient.issues.updateIssue({
      issueIdOrKey: issueKey,
      fields: {
        assignee: { id: jiraUser[0].accountId },
      },
    });

    console.log(`User ${userId} declared themselves as comms lead in channel ${channelId}`);
  } catch (error) {
    throw new Error(
      "Error setting comms lead: " +
      error +
      ". You can manually set the comms lead in the channel.",
    );
  }

  const responseText = "You are now the comms lead for this flare.";
  const flareChannelText = `<@${userId}> is now the comms lead for <#${channelId}>`;

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

export { commsLead, commsLeadRegex };

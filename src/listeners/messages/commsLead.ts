import { AllMiddlewareArgs, SlackEventMiddlewareArgs } from "@slack/bolt";
import { Version3Client } from "jira.js";
import config from "../../lib/config";
import { setListenerMatch } from "../../lib/listenerMatch";

const commsLeadRegex = /^(?:comms lead\b|.*\bi(?:'m| am)(?: the)? comms lead)|^\S+\s+comms lead$/i;

async function commsLead({
  client,
  message,
  say,
  context,
}: AllMiddlewareArgs & SlackEventMiddlewareArgs<"message">) {
  setListenerMatch(context);
  const jiraClient = context.clients.jiraClient as Version3Client;
  const channelId = context.channel.id;

  if (message.subtype !== undefined) {
    return;
  }

  if (!context.channel.name.startsWith(config.FLARE_CHANNEL_PREFIX)) {
    await say({
      text: "Sorry, I can only assign comms leads in a channel that corresponds to a Flare issue in JIRA.",
    });
    return;
  }

  const jiraTicket = context.channel.name.toUpperCase();

  await client.chat.postMessage({
    channel: channelId,
    thread_ts: message.ts,
    text: "working on assigning comms lead....",
  });

  try {
    const jiraUser = await jiraClient.userSearch.findUsers({
      query: context.user.profile.email,
    });

    if (!jiraUser || jiraUser.length === 0) {
      throw new Error("Could not find your JIRA user account");
    }

    // customfield_11405 is the ID of the Comms Lead custom field in our Flare project.
    // it was determined by calling jiraClient.issues.getEditIssueMeta() and looking at the response
    await jiraClient.issues.editIssue({
      issueIdOrKey: jiraTicket,
      fields: {
        customfield_11405: { id: jiraUser[0].accountId },
      },
    });

    await client.chat.postMessage({
      channel: channelId,
      thread_ts: message.ts,
      text: `Comms lead assigned! <@${context.user.id}> is now responsible for external communications.

If needed please make a copy of <https://docs.google.com/document/d/1KX1BlOq3eAvAnSdvzUg1AhbYamZtCSxmWGyMo0qYJa4|comms template doc> and fill in the details. When in doubt follow the instructions in this <https://app.getguru.com/card/ipR9GzET/Flares-Comms-Leads-Processes-and-RR|guru card>.
`,
    });
  } catch (error) {
    context.logger.errorD("comms-lead-assignment-error", { error: error });
    await say({
      text: `Sorry, I couldn't assign you as comms lead. Error: ${error}`,
    });
  }
}

export { commsLead, commsLeadRegex };

import { AllMiddlewareArgs, SlackEventMiddlewareArgs } from "@slack/bolt";
import { Version3Client } from "jira.js";
import config from "../../lib/config";

const incidentLeadRegex = /i(?:'m| am)(?: the)? incident lead/i;

async function incidentLead({
  client,
  message,
  say,
  context,
}: AllMiddlewareArgs & SlackEventMiddlewareArgs<"message">) {
  const jiraClient = context.clients.jiraClient as Version3Client;

  if (message.subtype !== undefined) {
    return;
  }

  if (!context.channel.name.startsWith(config.FLARE_CHANNEL_PREFIX)) {
    await say({
      text: "Sorry, I can only assign incident leads in a channel that corresponds to a Flare issue in JIRA.",
    });
    return;
  }

  const jiraTicket = context.channel.name.toUpperCase();

  await say({
    text: "working on assigning incident lead....",
  });

  try {
    const jiraUser = await jiraClient.userSearch.findUsers({
      query: context.user.profile.email,
    });

    if (!jiraUser || jiraUser.length === 0) {
      throw new Error("Could not find your JIRA user account");
    }

    await jiraClient.issues.editIssue({
      issueIdOrKey: jiraTicket,
      fields: {
        assignee: { id: jiraUser[0].accountId },
      },
    });

    await say({
      text: `Oh Captain My Captain! <@${context.user.id}> is now incident lead. Please confirm all actions with them.`,
    });
  } catch (error) {
    context.logger.errorD("incident-lead-assignment-error", { error: error });
    await say({
      text: `Sorry, I couldn't assign you as incident lead. Error: ${error}`,
    });
  }
}

export { incidentLead, incidentLeadRegex };
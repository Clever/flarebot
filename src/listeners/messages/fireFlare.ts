import config from "../../lib/config";
import { helpFlaresChannel } from "../../lib/help";
import { doJiraTransition, jiraDescription } from "../../lib/jira";
import { AllMiddlewareArgs, SlackEventMiddlewareArgs } from "@slack/bolt";
import introMessage from "../../lib/introMessage";
import { SectionBlock } from "@slack/types";
import { Version3Client } from "jira.js";
import { drive_v3 } from "@googleapis/drive";
import { setListenerMatch } from "../../lib/listenerMatch";

const specialTypeRetroactive = "retroactive";

// Regex that matches a flare command. The rough explanation is as follows:
// - Starts with "fire" or "fire a"
// - "flare" "<preemptive|retroactive>" "<p0|p1|p2>" can come in any order.
// - At least one of "<preemptive|retroactive>" or "<p0|p1|p2>" must be present.
//   But requiring priority is validated in the extractPriorityAndTitle function.
// - Title is always the last argument.
const fireAFlareRegex =
  /fire\s+(?:a\s+)?(?:flare\s+)?(?:(pre[- ]?emptive|retroactive|p0|p1|p2)\s+)?(?:flare\s+)?(?:(pre[- ]?emptive|retroactive|p0|p1|p2)\s+)(?:flare\s+)?(.+)/i;

const jiraErrorMessage =
  "I can't seem to connect to Jira right now. So I can't make a ticket or determine the next flare number. If you need to make a new channel to discuss, please don't use the next flare-number channel, that'll confuse me later on.";

async function fireFlare({
  client,
  message,
  say,
  context,
}: AllMiddlewareArgs & SlackEventMiddlewareArgs<"message">) {
  setListenerMatch(context);
  const jiraClient = context.clients.jiraClient as Version3Client;
  const googleDriveClient = context.clients.googleDriveClient as drive_v3.Drive;

  if (message.subtype !== undefined) {
    return;
  }

  const result = extractPriorityAndTitle(message.text ?? "");
  if (!result) {
    await say({
      text: `Sorry! I couldn't extract the priority and title from your message. ${helpFlaresChannel(context.botUserId)}`,
    });
    return;
  }
  const { specialType, priority, title } = result;

  if (specialType) {
    await say({
      text: `OK, let me quietly set up the Flare documents. Nobody freak out, this is ${specialType}.`,
    });
  } else {
    await say({
      text: `OK, let me get my flaregun`,
    });
  }

  let issueKey = "";

  try {
    const jiraUser = await jiraClient.userSearch.findUsers({
      query: context.user.profile.email,
    });

    const newIssue = await jiraClient.issues.createIssue({
      fields: {
        summary: title,
        issuetype: { name: "Bug" },
        project: { id: config.JIRA_PROJECT_ID },
        priority: { id: String(Number(priority) + 1) }, // P0 matches 1 and so on
        assignee: { id: jiraUser[0].accountId },
      },
    });

    issueKey = newIssue.key;
  } catch (error) {
    throw new Error(`${jiraErrorMessage}. Error: ${error}.`);
  }

  try {
    await doJiraTransition(jiraClient, issueKey, "In Progress");
    if (specialType === specialTypeRetroactive) {
      await doJiraTransition(jiraClient, issueKey, "Mitigated");
    }
  } catch (error) {
    await say({
      text: "JIRA ticket created but couldn't mark it 'started'. Continuing anyway...",
    });
    context.logger.errorD("jira-transition-error", { error: error });
  }

  let flareDocID = "";
  let slackHistoryDocID = "";
  // create a google doc for the flare using the template
  try {
    let flaredocTitle = `${issueKey}: ${title}`;
    if (specialType) {
      flaredocTitle = `${issueKey}: ${title} (${specialType})`;
    }

    let slackHistoryDocTitle = `${issueKey}: ${title} (Slack History)`;
    if (specialType) {
      slackHistoryDocTitle = `${issueKey}: ${title} (${specialType}) (Slack History)`;
    }

    const flaredoc = await googleDriveClient.files.copy({
      requestBody: {
        name: flaredocTitle,
      },
      fileId: config.GOOGLE_TEMPLATE_DOC_ID,
    });
    flareDocID = flaredoc.data.id ?? "";

    const slackHistoryDoc = await googleDriveClient.files.copy({
      requestBody: {
        name: slackHistoryDocTitle,
      },
      fileId: config.GOOGLE_SLACK_HISTORY_DOC_ID,
    });
    slackHistoryDocID = slackHistoryDoc.data.id ?? "";

    await googleDriveClient.permissions.create({
      fileId: flareDocID,
      requestBody: {
        role: "writer",
        type: "domain",
        domain: config.GOOGLE_DOMAIN,
      },
    });

    await googleDriveClient.permissions.create({
      fileId: slackHistoryDocID,
      requestBody: {
        role: "writer",
        type: "domain",
        domain: config.GOOGLE_DOMAIN,
      },
    });

    const flaredocHTML = await googleDriveClient.files.export({
      fileId: flareDocID,
      mimeType: "text/html",
    });

    let html = flaredocHTML.data as string;

    html = html.replace("[FLARE-KEY]", issueKey);
    html = html.replace(
      "[START-DATE]",
      new Date().toLocaleString("en-US", { timeZone: "US/Pacific" }) + " PT",
    );
    html = html.replace("[SUMMARY]", title);
    html = html.replace(
      "[HISTORY-DOC]",
      `<a href="https://docs.google.com/spreadsheets/d/${slackHistoryDocID}">${slackHistoryDocTitle}</a>`,
    );

    await googleDriveClient.files.update({
      fileId: flareDocID,
      media: {
        mimeType: "text/html",
        body: html,
      },
    });
  } catch (error) {
    await say({
      text: "I'm having trouble connecting to google drive right now, so I can't make a flare doc for tracking. Continuing anyway...",
    });
    context.logger.errorD("google-drive-error", { error: error });
  }

  try {
    await jiraClient.issues.editIssue({
      issueIdOrKey: issueKey,
      fields: {
        description: jiraDescription(flareDocID, slackHistoryDocID),
      },
    });
  } catch (error) {
    await say({
      text: "JIRA ticket created but couldn't set the description. Continuing anyway...",
    });
    context.logger.errorD("jira-description-error", { error: error });
  }

  let flareChannelId = "";
  try {
    const flareChannel = await client.conversations.create({
      name: issueKey.toLowerCase(),
    });

    flareChannelId = flareChannel.channel?.id ?? "";
    context.channelsCache.setChannel(flareChannelId, flareChannel.channel, slackHistoryDocID);
  } catch (error) {
    throw new Error(
      `Error creating flare channel ${error}. If you need to make a new channel to discuss, please create a channel with name ${issueKey.toLowerCase()}.`,
    );
  }

  try {
    await client.conversations.setTopic({
      channel: flareChannelId,
      topic: title,
    });

    const introMessageBlock = introMessage(
      issueKey,
      flareDocID,
      slackHistoryDocID,
      context.botUserId ?? "",
    );

    const introMessageResponse = await client.chat.postMessage({
      channel: flareChannelId,
      blocks: introMessageBlock,
      text: (introMessageBlock[0] as SectionBlock).text?.text ?? "", // this is used for notifications and bolt logs a warning if the text is not set.
    });

    if (!introMessageResponse.ts) {
      throw new Error("Intro message response has no timestamp. This should never happen.");
    }

    await client.pins.add({
      channel: flareChannelId,
      timestamp: introMessageResponse.ts,
    });

    const usersToInvite = [context.user.id, ...config.USERS_TO_INVITE.split(",")];
    await client.conversations.invite({
      channel: flareChannelId,
      users: usersToInvite.join(","),
    });
  } catch (error) {
    await say({
      text: `I'm having trouble setting up the new channel. Continuing anyway...`,
    });
    context.logger.errorD("flare-channel-invite-error", { error: error });
  }

  let audience = "<!channel>";
  if (specialType) {
    audience = `<@${context.user.id}>`;
  }
  await say({
    text: `${audience}: Flare fired. Please visit <#${flareChannelId}>`,
  });
}

function extractPriorityAndTitle(text: string) {
  const matches = text.match(fireAFlareRegex);
  if (!matches) return null;
  let priority = "";
  let specialType = "";
  if (matches[1] && matches[1].length == 2) {
    priority = matches[1].substring(1);
    specialType = matches[2];
  } else if (matches[2] && matches[2].length == 2) {
    specialType = matches[1] ? matches[1] : "";
    priority = matches[2].substring(1);
  }
  if (!priority && !specialType) {
    return null;
  }

  specialType = specialType ? specialType.replace("-", "").replace(" ", "").toLowerCase() : "";
  priority = priority ? priority.toLowerCase() : "";

  const title = matches[3] ? matches[3].trim() : "";
  if (!title) return null;
  return { specialType, priority, title };
}

export { fireFlare, fireAFlareRegex, extractPriorityAndTitle };

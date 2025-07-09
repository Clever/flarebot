import config from "../../lib/config";
import { helpFlaresChannel } from "../../lib/help";
import { doJiraTransition } from "../../lib/jira";
import { AllMiddlewareArgs, SlackEventMiddlewareArgs } from "@slack/bolt";
import introMessage from "../../lib/introMessage";
import { SectionBlock } from "@slack/types";
import { Version3Client } from "jira.js";
import { drive_v3 } from "@googleapis/drive";

const specialTypeRetroactive = "retroactive";

// Regex that matches a flare command. The rough explanation is as follows:
// - Starts with "fire" or "fire a"
// - "flare" "<preemptive|retroactive>" "<p0|p1|p2>" can come in any order.
// - At least one of "<preemptive|retroactive>" or "<p0|p1|p2>" must be present.
//   But requiring priority is validated in the extractPriorityAndTitle function.
// - Title is always the last argument.
const fireAFlareRegex =
  /fire\s+(?:a\s+)?(?:flare\s+)?(?:(pre[- ]?emptive|retroactive|p0|p1|p2)\s+)?(?:flare\s+)?(?:(pre[- ]?emptive|retroactive|p0|p1|p2)\s+)(?:flare\s+)?(.+)/i;

async function fireFlare({
  client,
  message,
  say,
  context,
}: AllMiddlewareArgs & SlackEventMiddlewareArgs<"message">) {
  const jiraClient = context.clients.jiraClient as Version3Client;
  const googleDriveClient = context.clients.googleDriveClient as drive_v3.Drive;

  if (!jiraClient || !googleDriveClient) {
    throw new Error("Jira or Google Drive client not found");
  }

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

    await doJiraTransition(jiraClient, issueKey, "Start Progress");
    if (specialType === specialTypeRetroactive) {
      await doJiraTransition(jiraClient, issueKey, "Mitigate");
    }
  } catch (error) {
    throw new Error("Error creating Jira issue", { cause: error });
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
    throw new Error("Error creating Google Doc", { cause: error });
  }

  let flareChannelId = "";
  try {
    const flareChannel = await client.conversations.create({
      name: issueKey.toLowerCase(),
    });

    flareChannelId = flareChannel.channel?.id ?? "";

    await client.conversations.setTopic({
      channel: flareChannelId,
      topic: title,
    });

    const introMessageBlock = introMessage(issueKey, flareDocID, slackHistoryDocID);

    const introMessageResponse = await client.chat.postMessage({
      channel: flareChannelId,
      blocks: introMessageBlock,
      text: (introMessageBlock[0] as SectionBlock).text?.text ?? "", // this is used for notifications and bolt logs a warning if the text is not set.
    });

    if (!introMessageResponse.ts) {
      throw new Error("Unexpected error - intro message response has no timestamp");
    }

    await client.pins.add({
      channel: flareChannelId,
      timestamp: introMessageResponse.ts,
    });
  } catch (error) {
    throw new Error("Error creating flare channel", { cause: error });
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
    specialType = matches[2].toLowerCase();
  } else if (matches[2] && matches[2].length == 2) {
    specialType = matches[1] ? matches[1].toLowerCase() : "";
    priority = matches[2].substring(1);
  }
  if (!priority && !specialType) {
    return null;
  }

  specialType = specialType ? specialType.replace("-", "").replace(" ", "") : "";

  const title = matches[3] ? matches[3].trim() : "";
  if (!title) return null;
  return { specialType, priority, title };
}

export { fireFlare, fireAFlareRegex, extractPriorityAndTitle };

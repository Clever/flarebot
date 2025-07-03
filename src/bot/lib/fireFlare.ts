import config from "./config";
import { helpFlaresChannel } from "./help";
import { doJiraTransition } from "./jira";
import { WebClient } from "@slack/web-api";
import { Context, KnownEventFromType, SayFn } from "@slack/bolt";
import introMessage from "./introMessage";
import { SectionBlock } from "@slack/types";

const specialTypeRetroactive = "retroactive";

// Regex that matches a flare command. The rough explanation is as follows:
// - Starts with "fire" or "fire a"
// - "flare" "<preemptive|retroactive>" "<p0|p1|p2>" can come in any order.
// - Accoding to regex at least one of "<preemptive|retroactive>" or "<p0|p1|p2>" must be present.
//   but requiring priority is validated in the extractPriorityAndTitle function.
// - Title is always the last argument.
const fireAFlareRegex =
  /fire\s+(?:a\s+)?(?:flare\s+)?(?:(pre[- ]?emptive|retroactive|p0|p1|p2)\s+)?(?:flare\s+)?(?:(pre[- ]?emptive|retroactive|p0|p1|p2)\s+)(?:flare\s+)?(.+)/i;

async function fireFlare({
  client,
  message,
  say,
  context,
}: {
  client: WebClient;
  message: KnownEventFromType<"message">;
  say: SayFn;
  context: Context;
}) {
  if (message.subtype !== undefined && message.subtype !== "bot_message") {
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
    const jiraUser = await context.jiraClient.userSearch.findUsers({
      query: context.user.profile.email,
    });

    const newIssue = await context.jiraClient.issues.createIssue({
      fields: {
        summary: title,
        issuetype: { name: "Bug" },
        project: { id: config.JIRA_PROJECT_ID },
        priority: { id: String(Number(priority) + 1) }, // P0 matches 1 and so on
        assignee: { id: jiraUser[0].accountId },
      },
    });

    issueKey = newIssue.key;

    await doJiraTransition(context.jiraClient, issueKey, "Start Progress");
    if (specialType === specialTypeRetroactive) {
      await doJiraTransition(context.jiraClient, issueKey, "Mitigate");
    }
  } catch (error) {
    throw new Error(`Error creating Jira issue: ${error}`);
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

    const introMessageResponse = await client.chat.postMessage({
      channel: flareChannelId,
      blocks: introMessage,
      text: (introMessage[0] as SectionBlock).text?.text ?? "", // this is used for notifications. Bolt logs a warning if the text is not set.
    });

    if (!introMessageResponse.ts) {
      throw new Error("Unexpected error - intro message response has no timestamp");
    }

    await client.pins.add({
      channel: flareChannelId,
      timestamp: introMessageResponse.ts,
    });

  } catch (error) {
    throw new Error(`Error creating flare channel: ${error}`);
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

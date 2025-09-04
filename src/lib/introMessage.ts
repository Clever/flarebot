import { AnyBlock } from "@slack/types";
import { recentDeploysActionID } from "./recentDeploys";
import { helpFlareChannel } from "./help";
import { flareRolesActionID } from "./flareRoles";
import { flareFollowupsActionID } from "./flareFollowups";
import { whatsAFlareActionID } from "./whatsAFlare";
import { howToPageActionID } from "./howToPage";
import { debugging101ActionID } from "./debugging101";
import { recentCriticalAlertsActionID } from "./recentCriticalAlerts";
import { openAlertsActionID } from "./openAlerts";

const introMessage = (
  issueKey: string,
  flaredoc: string,
  slackHistoryDoc: string,
  botUserId: string,
): AnyBlock[] => [
  {
    type: "section",
    text: {
      type: "mrkdwn",
      text: "Thank you for firing a flare! I am here to help you manage this flare and help solve it as soon as possible. This message is pinned so you can always access it easily from the top of the channel. To get started you can reference <https://app.getguru.com/card/TkXnd6ac/Engineering-Flare-Resources| Engineering Flare Resources> which is a comprehensive list of links (sample log queries, metrics dashboards, oncall guides, etc.) to help you debug this flare faster.",
    },
  },
  {
    type: "section",
    text: {
      type: "mrkdwn",
      text: helpFlareChannel(botUserId),
    },
  },
  {
    type: "section",
    text: {
      type: "mrkdwn",
      text: "I can also help you quickly look up a few things. Click on any of the buttons below to get some information that might help debug this flare faster.",
    },
  },
  {
    type: "actions",
    elements: [
      {
        type: "button",
        text: {
          type: "plain_text",
          text: "Recent Deploys",
          emoji: true,
        },
        action_id: recentDeploysActionID,
      },
      {
        type: "button",
        text: {
          type: "plain_text",
          text: "Recent Critical Alerts",
          emoji: true,
        },
        action_id: recentCriticalAlertsActionID,
      },
      {
        type: "button",
        text: {
          type: "plain_text",
          text: "Open Alerts",
          emoji: true,
        },
        action_id: openAlertsActionID,
      },
    ],
  },
  {
    type: "section",
    text: {
      type: "mrkdwn",
      text: "If you are unfamilar with our tooling, press these buttons below for some quick links to start debugging. *Do not hesitate to ask for help and page others if needed*. If there is high level of customer impact, like downtime, please page CS Managers",
    },
  },
  {
    type: "actions",
    elements: [
      {
        type: "button",
        text: {
          type: "plain_text",
          text: "how to page",
          emoji: true,
        },
        action_id: howToPageActionID,
      },
      {
        type: "button",
        text: {
          type: "plain_text",
          text: "debugging 101",
          emoji: true,
        },
        action_id: debugging101ActionID,
      },
    ],
  },
  {
    type: "section",
    text: {
      type: "mrkdwn",
      text: "If you are new to flares, then the following buttons are a good place to start.",
    },
  },
  {
    type: "actions",
    elements: [
      {
        type: "button",
        text: {
          type: "plain_text",
          text: "What is a flare?",
          emoji: true,
        },
        action_id: whatsAFlareActionID,
      },
      {
        type: "button",
        text: {
          type: "plain_text",
          text: "Flare Roles",
          emoji: true,
        },
        action_id: flareRolesActionID,
      },
      {
        type: "button",
        text: {
          type: "plain_text",
          text: "Flare Followups",
          emoji: true,
        },
        action_id: flareFollowupsActionID,
      },
    ],
  },
  // if intro message is updated so that this block is not the last one
  // then update recordMessage to account for the changes
  {
    type: "section",
    text: {
      type: "mrkdwn",
      text: `Finally, once the flare is mitigated, fill out the <https://clever.atlassian.net/browse/${issueKey}| jira ticket> to capture what we know, following instructions <https://clever.atlassian.net/wiki/spaces/ENG/pages/108210465/Flare+Followups#Updating-the-flare-ticket|here> and <https://clever.atlassian.net/wiki/spaces/ENG/pages/108210465/Flare+Followups#Signing-up-for-flare-followup|sign up> for the retro. Optionally use the <https://docs.google.com/document/d/${flaredoc}|Flare doc> and schedule a full post-mortem if there was high impact. You can also find the slack history stored <https://docs.google.com/spreadsheets/d/${slackHistoryDoc}|here>`,
    },
  },
];

export default introMessage;

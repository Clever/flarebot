import { AnyBlock } from "@slack/types";
import { recentDeploysActionID } from "./recentDeploys";
import { helpFlareChannel } from "./help";

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
      text: "Thank you for firing a flare! I am here to help you manage this flare and help solve it as soon as possible. This message is pinned so you can always access it easily from the top of the channel.",
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
        value: "click_me_123",
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
          text: "Log Queries",
          emoji: true,
        },
        value: "click_me_123",
      },
      {
        type: "button",
        text: {
          type: "plain_text",
          text: "Metrics Dashboard",
          emoji: true,
        },
        value: "click_me_123",
      },
      {
        type: "button",
        text: {
          type: "plain_text",
          text: "ark cheatsheet",
          emoji: true,
        },
        value: "click_me_123",
      },
      {
        type: "button",
        text: {
          type: "plain_text",
          text: "how to page",
          emoji: true,
        },
        value: "click_me_123",
      },
      {
        type: "button",
        text: {
          type: "plain_text",
          text: "debugging 101",
          emoji: true,
        },
        value: "click_me_123",
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
          text: "Flare Investigation Steps",
          emoji: true,
        },
        value: "click_me_123",
      },
      {
        type: "button",
        text: {
          type: "plain_text",
          text: "Flare Roles Definition",
          emoji: true,
        },
        value: "click_me_123",
      },
    ],
  },
  {
    type: "section",
    text: {
      type: "mrkdwn",
      text: `Finally, once the flare is mitigated, fill out the <https://clever.atlassian.net/browse/${issueKey}| jira ticket> to capture what we know, following instructions <https://clever.atlassian.net/wiki/spaces/ENG/pages/108210465/Flare+Followups#Updating-the-flare-ticket|here> and <https://clever.atlassian.net/wiki/spaces/ENG/pages/108210465/Flare+Followups#Signing-up-for-flare-followup|sign up> for the retro. Optionally use the <https://docs.google.com/document/d/${flaredoc}|Flare doc> and schedule a full post-mortem if there was high impact. You can also find the slack history stored <https://docs.google.com/spreadsheets/d/${slackHistoryDoc}|here>`,
    },
  },
];

export default introMessage;

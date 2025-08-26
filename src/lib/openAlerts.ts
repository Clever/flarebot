import { KnownBlock, ModalView } from "@slack/types";

const openAlertsActionID = "open_alerts";

const baseBlocks: KnownBlock[] = [
  {
    type: "header",
    text: {
      type: "plain_text",
      text: "All pagerduty alerts that haven't been resolved",
    },
  },
  {
    type: "context",
    elements: [
      {
        type: "mrkdwn",
        text: "<https://getclever.pagerduty.com/incidents|View all alerts in PagerDuty>",
      },
    ],
  },
];

/* eslint-disable-next-line  @typescript-eslint/no-explicit-any */
const openAlertsModalView = (alerts?: any[]): ModalView => {
  const blocks = [...baseBlocks];
  if (alerts === undefined) {
    blocks.push({
      type: "divider",
    });
    blocks.push({
      type: "section",
      text: {
        type: "mrkdwn",
        text: ":loading:",
      },
    });
  } else {
    if (alerts.length === 0) {
      blocks.push({
        type: "divider",
      });
      blocks.push({
        type: "section",
        text: {
          type: "mrkdwn",
          text: "No open alerts found.",
        },
      });
    } else {
      if (alerts.length > 30) {
        blocks.push({
          type: "context",
          elements: [
            {
              type: "mrkdwn",
              text: `:information_source: Showing only the most recent 30 alerts. There are ${alerts.length} total open alerts.`,
            },
          ],
        });
      }

      /* eslint-disable-next-line  @typescript-eslint/no-explicit-any */
      alerts.slice(0, 30).forEach((alert: any) => {
        blocks.push({
          type: "divider",
        });
        blocks.push({
          type: "section",
          text: {
            type: "mrkdwn",
            text: `<${alert.html_url}|${alert.title.trim().replaceAll("\n", " ")}>`,
          },
          fields: [
            {
              type: "mrkdwn",
              text: `*Status:*\n${alert.status}`,
            },
            {
              type: "mrkdwn",
              text: `*Assignee:*\n${alert.assignments?.[0]?.assignee?.summary || "Unassigned"}`,
            },
            {
              type: "mrkdwn",
              text: `*Team:*\n<${alert.service.html_url}|${alert.service.summary}>`,
            },
            {
              type: "mrkdwn",
              text: `*Created:*\n${new Date(alert.created_at).toLocaleString("en-US", { timeZone: "US/Pacific" }) + " PT"}`,
            },
          ],
        });
      });
    }
  }

  return {
    type: "modal",
    title: {
      type: "plain_text",
      text: "Open Alerts",
    },
    close: {
      type: "plain_text",
      text: "Close",
    },
    blocks: blocks,
  };
};

const errorModalView = (error: string): ModalView => {
  const blocks = [...baseBlocks];
  blocks.push({
    type: "divider",
  });
  blocks.push({
    type: "section",
    text: {
      type: "mrkdwn",
      text: `:emergency: error getting open alerts:\n\`\`\`${error}\`\`\``,
    },
  });
  return {
    type: "modal",
    title: {
      type: "plain_text",
      text: "Open Incidents",
    },
    close: {
      type: "plain_text",
      text: "Close",
    },
    blocks: blocks,
  };
};

export { openAlertsActionID, openAlertsModalView, errorModalView };

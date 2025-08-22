import { KnownBlock, ModalView } from "@slack/types";

const recentCriticalAlertsActionID = "recent_critical_alerts";

const baseBlocks: KnownBlock[] = [
  {
    type: "header",
    text: {
      type: "plain_text",
      text: "Most recent 30 critical alerts",
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
const recentCriticalAlertsModalView = (alerts?: any[]): ModalView => {
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
    /* eslint-disable-next-line  @typescript-eslint/no-explicit-any */
    alerts.forEach((alert: any) => {
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
            text: `*Triggered:*\n${
              new Date(alert.created_at).toLocaleString("en-US", { timeZone: "US/Pacific" }) + " PT"
            }`,
          },
          {
            type: "mrkdwn",
            text: `*Last Updated:*\n${
              new Date(alert.updated_at).toLocaleString("en-US", { timeZone: "US/Pacific" }) + " PT"
            }`,
          },
          {
            type: "mrkdwn",
            text: `*Status*: ${alert.status}`,
          },
          {
            type: "mrkdwn",
            text: `*Team*: <${alert.service.html_url}|${alert.service.summary}>`,
          },
        ],
      });
    });
  }

  return {
    type: "modal",
    title: {
      type: "plain_text",
      text: "Recent Critical Alerts",
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
      text: `:emergency: error getting recent critical alerts:\n\`\`\`${error}\`\`\``,
    },
  });
  return {
    type: "modal",
    title: {
      type: "plain_text",
      text: "Recent Critical Alerts",
    },
    close: {
      type: "plain_text",
      text: "Close",
    },
    blocks: blocks,
  };
};

export { recentCriticalAlertsActionID, recentCriticalAlertsModalView, errorModalView };

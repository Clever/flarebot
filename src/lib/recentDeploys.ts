import Catapult from "@clever/catapult";
import { KnownBlock, ModalView } from "@slack/types";

const recentDeploysActionID = "recent_deploys";

const baseBlocks: KnownBlock[] = [
  {
    type: "header",
    text: {
      type: "plain_text",
      text: "Last 15 deployments to production",
    },
  },
  {
    type: "context",
    elements: [
      {
        type: "mrkdwn",
        text: "`ark info -e production` for detailed info and `ark rollback --help` on how go to a previous deployment",
      },
    ],
  },
];

const recentDeploysModalView = (deployments?: Catapult.Models.DeploymentV2[]): ModalView => {
  const blocks = [...baseBlocks];
  if (deployments === undefined) {
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
    deployments.forEach((deployment) => {
      blocks.push({
        type: "divider",
      });
      blocks.push({
        type: "section",
        text: {
          type: "mrkdwn",
          text: `*${deployment.id}*`,
        },
        fields: [
          {
            type: "plain_text",
            text: deployment.createdAt
              ? new Date(deployment.createdAt).toLocaleString("en-US", { timeZone: "US/Pacific" }) +
                " PT"
              : "unknown",
            emoji: true,
          },
          {
            type: "plain_text",
            text: deployment.owner?.replace("via dapple", "") || "unknown",
            emoji: true,
          },
          {
            type: "mrkdwn",
            text: deployment.build
              ? `<https://github.com/Clever/catapult/commit/${deployment.build}|build - ${deployment.build}>`
              : "unknown",
          },
          {
            type: "mrkdwn",
            text: deployment.envProvider
              ? `<https://github.com/Clever/ark-config/tree/${deployment.build}/apps/${deployment.envProvider.repo}|config - ${deployment.build}>`
              : "unknown",
          },
        ],
      });
    });
  }

  return {
    type: "modal",
    title: {
      type: "plain_text",
      text: "Recent Deploys",
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
      text: `:emergency: error getting recent deploys:\n\`\`\`${error}\`\`\``,
    },
  });
  return {
    type: "modal",
    title: {
      type: "plain_text",
      text: "Recent Deploys",
    },
    close: {
      type: "plain_text",
      text: "Close",
    },
    blocks: blocks,
  };
};

export { recentDeploysActionID, recentDeploysModalView, errorModalView };

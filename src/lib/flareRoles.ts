import { ModalView } from "@slack/types";

const flareRolesActionID = "flare_roles";

const flareRolesModalView = (): ModalView => {
  return {
    type: "modal",
    title: {
      type: "plain_text",
      text: "Flare Roles",
    },
    blocks: [
      {
        type: "header",
        text: {
          type: "plain_text",
          text: "Incident Lead",
        },
      },
      {
        type: "section",
        text: {
          type: "mrkdwn",
          text:
            "Directs the flare response.\n" +
            "Selects comms lead and participants.\n" +
            "Declares flare state and makes all key decisions.\n" +
            "Keeps the big picture in mind and delegates deep dives.\n" +
            "Should not handle comms or get lost in details."
        },
      },
      {
        type: "divider"
      },
      {
        type: "header",
        text: {
          type: "plain_text",
          text: "Comms Lead",
        },
      },
      {
        type: "section",
        text: {
          type: "mrkdwn",
          text:
            "Manages all communications for the flare.\n" +
            "Updates Statuspage and Facts Doc as needed.\n" +
            "Keeps relevant Slack channels informed.\n" +
            "Coordinates with Incident Lead and brings in support for complex comms.\n" +
            "Usually a CS Manager."
        },
      },
      {
        type: "divider"
      },
      {
        type: "header",
        text: {
          type: "plain_text",
          text: "Participant",
        },
      },
      {
        type: "section",
        text: {
          type: "mrkdwn",
          text:
            "Follows Incident Lead’s direction.\n" +
            "Communicates relevant info and proposes actions.\n" +
            "Focuses on assigned tasks and investigations.\n" +
            "Waits for approval before making changes.\n" +
            "Does not act independently."
        },
      },
      {
        type: "divider"
      },
      {
        type: "header",
        text: {
          type: "plain_text",
          text: "Non-Participant",
        },
      },
      {
        type: "section",
        text: {
          type: "mrkdwn",
          text:
            "Stays silent in the channel unless sharing unique, important info.\n" +
            "Requests to join if they have relevant expertise.\n" +
            "Becomes a participant only if asked by the Incident Lead.\n" +
            "Avoids unnecessary suggestions or commentary."
        },
      },
    ],
  };
};

export { flareRolesActionID, flareRolesModalView };


import { ModalView } from "@slack/types";

const whatsAFlareActionID = "whats_a_flare";

const whatsAFlareModalView = (): ModalView => {
  return {
    type: "modal",
    title: {
      type: "plain_text",
      text: "What is a flare?",
    },
    blocks: [
      {
        type: "header",
        text: {
          type: "plain_text",
          text: "What is a flare?",
        },
      },
      {
        type: "section",
        text: {
          type: "mrkdwn",
          text: `
A flare is an outage, regression, or other large issues affecting many end-users (external or internal).

Each individual flare has an associated priority. This helps us know how to respond to the flare.

- P0 is an existential threat to Clever if it isn't rapidly addressed. It wakes everyone up.
- P1 affects more than 10% of our schools or apps urgently. It wakes someone up.
- P2 affects less than 10% of our schools or apps. While time-sensitive, this can usually wait until the beginning of the business day (9am).`,
        },
      },
    ],
  };
};

export { whatsAFlareActionID, whatsAFlareModalView };

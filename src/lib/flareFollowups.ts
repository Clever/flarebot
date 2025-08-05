import { ModalView } from "@slack/types";

const flareFollowupsActionID = "flare_followups";

const flareFollowupsModalView = (): ModalView => {
  return {
    type: "modal",
    title: {
      type: "plain_text",
      text: "Flare Followups",
    },
    blocks: [
      {
        type: "header",
        text: {
          type: "plain_text",
          text: "Flare Followups",
        },
      },
      {
        type: "section",
        text: {
          type: "mrkdwn",
          text: `
Once a flare has been called mitigated, the incident lead will sign up for flare followups following the <https://clever.atlassian.net/wiki/spaces/ENG/pages/108210465/Flare+Followups|instructions> here.

The purpose of this meeting is to:

- Document a summary of the flare (what happened and why)
- Discuss any potential follow-ups that haven't already occurred (e.g. creating an alert)
- Decide whether a full retro is required

Before the meeting the incident lead should prepare by:

- gathering the relevant facts like customer impact and timeline
- update the FLARE JIRA ticket with details of the flare impact`,
        },
      },
    ],
  };
};

export { flareFollowupsActionID, flareFollowupsModalView };

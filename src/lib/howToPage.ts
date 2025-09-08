import { ModalView } from "@slack/types";

const howToPageActionID = "how_to_page";

const howToPageModalView = (fileId: string): ModalView => {
  return {
    type: "modal",
    title: {
      type: "plain_text",
      text: "How to Page?",
    },
    blocks: [
      {
        type: "header",
        text: {
          type: "plain_text",
          text: "How to Page?",
        },
      },
      {
        type: "section",
        text: {
          type: "mrkdwn",
          text: `
To page someone you simply type \`/pd trigger\` in slack and hit return. This will open a modal to create a new incident.

You can then fill in the details of the incident and hit "Create".

For "Impacted Service" - search for the team name that you want to page e.g. Eng ..., Eng Managers or CS Managers.

You can also page a specific person by clicking "Add details" and then searching for an individual in the "Assign to" field.
        `,
        },
      },
      {
        type: "image",
        slack_file: {
          id: fileId,
        },
        alt_text: "/pd trigger",
      },
    ],
  };
};

export { howToPageActionID, howToPageModalView };

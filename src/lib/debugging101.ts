import { ModalView } from "@slack/types";

const debugging101ActionID = "debugging101";

const debugging101ModalView = (): ModalView => {
  return {
    type: "modal",
    title: {
      type: "plain_text",
      text: "Debuging 101",
    },
    blocks: [
      {
        type: "header",
        text: {
          type: "plain_text",
          text: "Debuging 101",
        },
      },
      {
        type: "section",
        text: {
          type: "mrkdwn",
          text: "This is a very complicated topic and it is often something you learn from experience instead of following a guide but the flow chart below is a good starting point for how to approach incidents.",
        },
      },
      {
        type: "image",
        slack_file: {
          id: "F0990KPQFRA", // Hard code this for now but we should look it up from slack based on file name
        },
        alt_text: "Debuging 101",
      },
    ],
  };
};

export { debugging101ActionID, debugging101ModalView };

import { Version3Client } from "jira.js";

async function doJiraTransition(
  jiraclient: Version3Client,
  ticket: string,
  transitionName: string,
) {
  const jiraIssue = await jiraclient.issues.getIssue({
    issueIdOrKey: ticket,
    expand: "transitions",
  });

  const transition = jiraIssue.transitions?.find((t) => t.to?.name === transitionName);

  if (!transition) {
    throw new Error(
      `Jira transition "${transitionName}" not found. Allowed transitions for current status: [${jiraIssue.transitions?.map((t) => t.to?.name).join(", ")}]`,
    );
  }

  await jiraclient.issues.doTransition({
    issueIdOrKey: ticket,
    transition: transition,
  });
}

const jiraDescription = (flaredoc: string, slackHistoryDoc: string) => ({
  version: 1,
  type: "doc",
  content: [
    {
      type: "heading",
      attrs: { level: 2 },
      content: [{ type: "text", text: "Customer Impact" }],
    },
    {
      type: "paragraph",
      content: [
        {
          type: "text",
          text: "[example: All users attempting to use Help center documentation could not see body content for about 45 minutes from approximately Wed 4/7 at 4:30pm to 5:15pm.]",
          marks: [{ type: "em" }],
        },
      ],
    },
    {
      type: "heading",
      attrs: { level: 2 },
      content: [{ type: "text", text: "Description" }],
    },
    {
      type: "paragraph",
      content: [
        {
          type: "text",
          text: "[example: The body content of Help Center articles were not visible on desktop browsers. After testing, body content was still visible for mobile browsers and desktop in developer “mobile view”. No indication of an outage was surfaced at ",
          marks: [{ type: "em" }],
        },
        {
          type: "text",
          text: "status.salesforce.com",
          marks: [
            { type: "link", attrs: { href: "http://status.salesforce.com/" } },
            { type: "em" },
          ],
        },
        {
          type: "text",
          text: " and no admission of error has been secured from the Salesforce team. Around 5:17pm PT, all content was again visible without any action taken by the Clever team in our Salesforce instance.\n\n4/7 4:33pm PT, first post in #oncall-solutions\n4/7 4:34pm PT, flare fired\n4/7 4:37pm PT, messages in motion to contractor, account manager, & Salesforce support\n4/7 4:54pm PT, status page updated\n4/7 5:17pm PT, HC live again\n4/7 5:22pm PT, status page - mitigated\n4/7 5:22pm PT, flare mitigated]",
          marks: [{ type: "em" }],
        },
      ],
    },
    {
      type: "heading",
      attrs: { level: 2 },
      content: [{ type: "text", text: "Followup" }],
    },
    {
      type: "paragraph",
      content: [
        {
          type: "text",
          text: "[example: Webex meeting with Salesforce support on Monday, 4/12 to continue investigation into root issue]\n\n",
          marks: [{ type: "em" }],
        },
        {
          type: "text",
          text: "Flare Doc",
          marks: [
            { type: "link", attrs: { href: `https://docs.google.com/document/d/${flaredoc}` } },
            { type: "em" },
          ],
        },
        {
          type: "text",
          text: " | ",
        },
        {
          type: "text",
          text: "Slack History",
          marks: [
            {
              type: "link",
              attrs: { href: `https://docs.google.com/spreadsheets/d/${slackHistoryDoc}` },
            },
            { type: "em" },
          ],
        },
      ],
    },
  ],
});

export { doJiraTransition, jiraDescription };

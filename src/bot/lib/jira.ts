import { Version3Client } from "jira.js";

async function doJiraTransition(
  jiraclient: Version3Client,
  ticket: string,
  transitionName: string,
) {
  const resp = await jiraclient.issues.getTransitions({
    issueIdOrKey: ticket,
  });

  const transition = resp.transitions?.find((t) => t.name === transitionName);

  if (!transition) {
    throw new Error(`Transition ${transitionName} not found`);
  }

  await jiraclient.issues.doTransition({
    issueIdOrKey: ticket,
    transition: transition,
  });
}

export { doJiraTransition };

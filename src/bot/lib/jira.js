async function doTransition(jiraclient, ticket, transitionName) {
  const resp = await jiraclient.issues.getTransitions({
    issueIdOrKey: ticket,
  });

  const transition = resp.transitions.find((t) => t.name === transitionName);

  if (!transition) {
    throw new Error(`Transition ${transitionName} not found`);
  }

  await jiraclient.issues.doTransition({
    issueIdOrKey: ticket,
    transition: transition,
  });
}

module.exports = { doTransition };

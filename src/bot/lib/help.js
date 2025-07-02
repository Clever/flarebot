const helpRegex = /help\s*(all)?/i;

async function help({ message, context, say }) {
  const isAll = isHelpAll(message.text);
  let helpText;
  if (isAll) {
    helpText = helpAll(context.botUserId);
  } else if (context.channel.name === process.env.FLARES_CHANNEL_NAME) {
    helpText = helpFlaresChannel(context.botUserId);
  } else if (context.channel.name.startsWith(process.env.FLARE_CHANNEL_PREFIX)) {
    helpText = helpFlareChannel(context.botUserId);
  }
  await say({
    text: helpText,
  });
}

function isHelpAll(text) {
  return text.match(helpRegex)[1] === "all";
}

function helpFlaresChannel(botUserId) {
  return `
Commands available in the <#${process.env.FLARES_CHANNEL_ID}> channel:

<@${botUserId}> help [all] - Display the list of commands available in this/all channel.
<@${botUserId}> fire a flare <p0|p1|p2> [preemptive|retroactive] <title> - Fire a new Flare with the given priority and description. Optionally specify preemptive or retroactive. Ordering is not important but title should be last.
`;
}

function helpFlareChannel(botUserId) {
  return `
Commands available in a single Flare channel:

<@${botUserId}> help [all] - Display the list of commands available in this/all channel.
<@${botUserId}> i am incident lead - Declare yourself incident lead.
<@${botUserId}> i am comms lead - Declare yourself comms lead.
<@${botUserId}> flare mitigated - Mark the Flare mitigated.
<@${botUserId}> not a flare - Mark the Flare not-a-flare.
`;
}

function helpAll(botUserId) {
  return `
${helpFlaresChannel(botUserId)}${helpFlareChannel(botUserId)}
`;
}

module.exports = { helpRegex, help, helpAll, helpFlaresChannel };

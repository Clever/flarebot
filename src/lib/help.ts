import config from "./config";

function helpFlaresChannel(botUserId: string | undefined) {
  return `
Commands available in the <#${config.FLARES_CHANNEL_ID}> channel:

<@${botUserId}> help [all] - Display the list of commands available in this/all channel.
<@${botUserId}> fire a flare <p0|p1|p2> [preemptive|retroactive] <title> - Fire a new Flare with the given priority and description. Optionally specify preemptive or retroactive. Ordering is not important but title should be last.
`;
}

function helpFlareChannel(botUserId: string | undefined) {
  return `
Commands available in a single Flare channel:

<@${botUserId}> help [all] - Display the list of commands available in this/all channel.
<@${botUserId}> i am incident lead - Declare yourself incident lead.
<@${botUserId}> i am comms lead - Declare yourself comms lead.
<@${botUserId}> flare mitigated - Mark the Flare mitigated.
<@${botUserId}> not a flare - Mark the Flare not-a-flare.
`;
}

function helpAll(botUserId: string | undefined) {
  return `
${helpFlaresChannel(botUserId)}${helpFlareChannel(botUserId)}
`;
}

export { helpAll, helpFlaresChannel, helpFlareChannel };

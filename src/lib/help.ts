import config from "./config";

function helpFlaresChannel(botUserId: string | undefined) {
  return `
Commands available in the <#${config.FLARES_CHANNEL_ID}> channel: (Text in [ ] is optional)

<@${botUserId}> help [all] - Display the list of commands available in this/all channel.
<@${botUserId}> fire a flare <p0|p1|p2> [preemptive|retroactive] <title> - Fire a new Flare with the given priority and description. Optionally specify preemptive or retroactive. Ordering is not important but title should be last.
`;
}

function helpFlareChannel(botUserId: string | undefined) {
  return `
Commands available in a flare channel: (Text in [ ] is optional)

<@${botUserId}> help [all] - Display the list of commands available in this/all channel.
<@${botUserId}> [i am] incident lead - Declare yourself incident lead.
<@${botUserId}> [i am] comms lead - Declare yourself comms lead.
<@${botUserId}> [flare is] mitigate[d] - Mark the Flare as mitigated.
<@${botUserId}> [flare is] not a flare - Mark the Flare as not-a-flare.
<@${botUserId}> [flare is] unmitigate[d] - Mark the Flare as in-progress.
`;
}

function helpAll(botUserId: string | undefined) {
  return `
${helpFlaresChannel(botUserId)}${helpFlareChannel(botUserId)}
`;
}

export { helpAll, helpFlaresChannel, helpFlareChannel };

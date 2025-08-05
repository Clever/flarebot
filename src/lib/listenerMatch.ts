import { Context } from "@slack/bolt";

// https://github.com/slackapi/bolt-js/issues/1164

function setListenerMatch(context: Context) {
  context.listenerMatch = true;
}
function isListenerMatch(context: Context) {
  return context.listenerMatch == true;
}

export { setListenerMatch, isListenerMatch };

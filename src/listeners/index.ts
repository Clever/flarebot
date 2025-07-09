import type { App } from "@slack/bolt";
import messages from "./messages";
import actions from "./actions";

const register = (app: App) => {
  messages.register(app);
  actions.register(app);
};

export default { register };

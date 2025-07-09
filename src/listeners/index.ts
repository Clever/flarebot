import type { App } from "@slack/bolt";
import messages from "./messages";

const register = (app: App) => {
  messages.register(app);
};

export default { register };

import type { App } from "@slack/bolt";
import { help, helpRegex } from "./help";
import { fireAFlareRegex, fireFlare } from "./fireFlare";

const register = (app: App) => {
  app.message(fireAFlareRegex, fireFlare);
  app.message(helpRegex, help);
};

export default { register };

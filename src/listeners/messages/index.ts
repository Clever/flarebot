import type { App } from "@slack/bolt";
import { help, helpRegex } from "./help";
import { fireAFlareRegex, fireFlare } from "./fireFlare";
import { flareTransitionRegex, flareTransition } from "./flareTransition";
import { commsLeadRegex, commsLead } from "./commsLead";

const register = (app: App) => {
  app.message(fireAFlareRegex, fireFlare);
  app.message(helpRegex, help);
  app.message(flareTransitionRegex, flareTransition);
  app.message(commsLeadRegex, commsLead);
};

export default { register };

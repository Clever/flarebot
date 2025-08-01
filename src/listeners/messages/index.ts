import type { App } from "@slack/bolt";
import { help, helpRegex } from "./help";
import { fireAFlareRegex, fireFlare } from "./fireFlare";
import { flareTransitionRegex, flareTransition } from "./flareTransition";
import { incidentLeadRegex, incidentLead } from "./incidentLead";

const register = (app: App) => {
  app.message(fireAFlareRegex, fireFlare);
  app.message(helpRegex, help);
  app.message(flareTransitionRegex, flareTransition);
  app.message(incidentLeadRegex, incidentLead);
};

export default { register };

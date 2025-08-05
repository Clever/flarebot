import type { App } from "@slack/bolt";
import { recentDeploys } from "./recentDeploys";
import { recentDeploysActionID } from "../../lib/recentDeploys";
import { flareRoles } from "./flareRoles";
import { flareRolesActionID } from "../../lib/flareRoles";
import { flareFollowupsActionID } from "../../lib/flareFollowups";
import { howToPageActionID } from "../../lib/howToPage";
import { howToPage } from "./howToPage";
import { whatsAFlareActionID } from "../../lib/whatsAFlare";
import { whatsAFlare } from "./whatsAFlare";
import { flareFollowups } from "./flareFollowups";
import { debugging101ActionID } from "../../lib/debugging101";
import { debugging101 } from "./debugging101";

const register = (app: App) => {
  app.action(recentDeploysActionID, recentDeploys);
  app.action(flareRolesActionID, flareRoles);
  app.action(flareFollowupsActionID, flareFollowups);
  app.action(howToPageActionID, howToPage);
  app.action(whatsAFlareActionID, whatsAFlare);
  app.action(debugging101ActionID, debugging101);
};

export default { register };

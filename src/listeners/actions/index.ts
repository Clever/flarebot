import type { App } from "@slack/bolt";
import { recentDeploys } from "./recentDeploys";
import { recentDeploysActionID } from "../../lib/recentDeploys";
import { flareRoles } from "./flareRoles";
import { flareRolesActionID } from "../../lib/flareRoles";

const register = (app: App) => {
  app.action(recentDeploysActionID, recentDeploys);
  app.action(flareRolesActionID, flareRoles);
};

export default { register };

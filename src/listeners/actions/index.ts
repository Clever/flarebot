import type { App } from "@slack/bolt";
import { recentDeploys } from "./recentDeploys";
import { recentDeploysActionID } from "../../lib/recentDeploys";

const register = (app: App) => {
  app.action(recentDeploysActionID, recentDeploys);
};

export default { register };

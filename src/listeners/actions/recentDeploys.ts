import { AllMiddlewareArgs, SlackActionMiddlewareArgs } from "@slack/bolt";
import { BlockAction } from "@slack/bolt/dist/types/actions";
import { errorModalView, recentDeploysModalView } from "../../lib/recentDeploys";
import Catapult from "@clever/catapult";

const recentDeploys = async ({
  client,
  body,
  ack,
  context,
}: AllMiddlewareArgs & SlackActionMiddlewareArgs<BlockAction>) => {
  await ack();

  let viewID = "";
  try {
    // trigger_id is only valid for 3 seconds so open the modal immediately and then later fetch the catapult api for the recent deploys
    const view = await client.views.open({
      trigger_id: body.trigger_id,
      view: recentDeploysModalView(),
    });
    viewID = view.view?.id ?? "";
  } catch (error) {
    throw new Error("Error opening modal", { cause: error });
  }

  try {
    const catapultClient = context.clients.catapultClient as Catapult;
    const deployments = await catapultClient.getDeploymentsV2({
      env: "production",
      limit: 15,
    });

    await client.views.update({
      view_id: viewID,
      view: recentDeploysModalView(deployments),
    });
  } catch (error) {
    const errorMessage = error instanceof Error ? error.toString() : "Unknown error";
    await client.views.update({
      view_id: viewID,
      view: errorModalView(errorMessage),
    });
    throw error;
  }
};

export { recentDeploys };

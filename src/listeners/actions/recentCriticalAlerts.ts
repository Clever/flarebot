import { AllMiddlewareArgs, SlackActionMiddlewareArgs } from "@slack/bolt";
import { BlockAction } from "@slack/bolt/dist/types/actions";
import { errorModalView, recentCriticalAlertsModalView } from "../../lib/recentCriticalAlerts";

const recentCriticalAlerts = async ({
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
      view: recentCriticalAlertsModalView(),
    });
    viewID = view.view?.id ?? "";
  } catch (error) {
    throw new Error("Error opening modal", { cause: error });
  }

  try {
    const pagerDutyClient = context.clients.pagerDutyClient;
    const data = await pagerDutyClient.get("/incidents", {
      queryParameters: {
        limit: 30,
        sort_by: "created_at:desc",
        time_zone: "US/Pacific",
      },
    });

    await client.views.update({
      view_id: viewID,
      view: recentCriticalAlertsModalView(data.resource),
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

export { recentCriticalAlerts };

import { AllMiddlewareArgs, SlackActionMiddlewareArgs } from "@slack/bolt";
import { BlockAction } from "@slack/bolt/dist/types/actions";
import { errorModalView, openAlertsModalView } from "../../lib/openAlerts";

const openAlerts = async ({
  client,
  body,
  ack,
  context,
}: AllMiddlewareArgs & SlackActionMiddlewareArgs<BlockAction>) => {
  await ack();

  let viewID = "";
  try {
    // trigger_id is only valid for 3 seconds so open the modal immediately and then later fetch the pagerduty api for the open incidents
    const view = await client.views.open({
      trigger_id: body.trigger_id,
      view: openAlertsModalView(),
    });
    viewID = view.view?.id ?? "";
  } catch (error) {
    throw new Error("Error opening modal", { cause: error });
  }

  try {
    const pagerDutyClient = context.clients.pagerDutyClient;
    const data = await pagerDutyClient.get("/incidents", {
      queryParameters: {
        limit: 50,
        sort_by: "created_at:desc",
        time_zone: "US/Pacific",
        "statuses[]": ["acknowledged", "triggered"],
      },
    });
    if (data.response.status !== 200) {
      throw new Error(
        "Error getting open alerts: " +
          data.response.status +
          " - " +
          data.response.statusText +
          " - " +
          data.data.error.message +
          " - " +
          data.data.error.code +
          " - " +
          data.data.error.errors.join(", "),
      );
    }

    await client.views.update({
      view_id: viewID,
      view: openAlertsModalView(data.resource),
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

export { openAlerts };

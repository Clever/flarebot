import { AllMiddlewareArgs, SlackActionMiddlewareArgs } from "@slack/bolt";
import { BlockAction } from "@slack/bolt/dist/types/actions";
import { flareFollowupsModalView } from "../../lib/flareFollowups";

const flareFollowups = async ({
  client,
  body,
  ack,
}: AllMiddlewareArgs & SlackActionMiddlewareArgs<BlockAction>) => {
  await ack();
  await client.views.open({
    trigger_id: body.trigger_id,
    view: flareFollowupsModalView(),
  });
};

export { flareFollowups };

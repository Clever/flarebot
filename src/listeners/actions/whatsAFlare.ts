import { AllMiddlewareArgs, SlackActionMiddlewareArgs } from "@slack/bolt";
import { BlockAction } from "@slack/bolt/dist/types/actions";
import { whatsAFlareModalView } from "../../lib/whatsAFlare";

const whatsAFlare = async ({
  client,
  body,
  ack,
}: AllMiddlewareArgs & SlackActionMiddlewareArgs<BlockAction>) => {
  await ack();
  await client.views.open({
    trigger_id: body.trigger_id,
    view: whatsAFlareModalView(),
  });
};

export { whatsAFlare };

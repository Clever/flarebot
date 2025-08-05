import { AllMiddlewareArgs, SlackActionMiddlewareArgs } from "@slack/bolt";
import { BlockAction } from "@slack/bolt/dist/types/actions";
import { debugging101ModalView } from "../../lib/debugging101";

const debugging101 = async ({
  client,
  body,
  ack,
}: AllMiddlewareArgs & SlackActionMiddlewareArgs<BlockAction>) => {
  await ack();
  await client.views.open({
    trigger_id: body.trigger_id,
    view: debugging101ModalView(),
  });
};

export { debugging101 };

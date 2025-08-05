import { AllMiddlewareArgs, SlackActionMiddlewareArgs } from "@slack/bolt";
import { BlockAction } from "@slack/bolt/dist/types/actions";
import { howToPageModalView } from "../../lib/howToPage";

const howToPage = async ({
  client,
  body,
  ack,
}: AllMiddlewareArgs & SlackActionMiddlewareArgs<BlockAction>) => {
  await ack();
  await client.views.open({
    trigger_id: body.trigger_id,
    view: howToPageModalView(),
  });
};

export { howToPage };

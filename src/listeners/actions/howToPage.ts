import { AllMiddlewareArgs, SlackActionMiddlewareArgs } from "@slack/bolt";
import { BlockAction } from "@slack/bolt/dist/types/actions";
import { howToPageModalView } from "../../lib/howToPage";
import { uploadFiles } from "../../lib/uploadFiles";

const howToPage = async ({
  client,
  body,
  context,
  ack,
}: AllMiddlewareArgs & SlackActionMiddlewareArgs<BlockAction>) => {
  await ack();

  const files = await client.files.list({
    user: context.botUserId,
    types: "images",
  });

  const filesMap = new Map(files.files?.map((file) => [file.name, file]) ?? []);

  if (!filesMap.has("pd-trigger.png")) {
    await uploadFiles(client, context.botUserId ?? "");
  }

  await client.views.open({
    trigger_id: body.trigger_id,
    view: howToPageModalView(filesMap.get("pd-trigger.png")?.id ?? ""),
  });
};

export { howToPage };

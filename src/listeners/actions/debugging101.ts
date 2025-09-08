import { AllMiddlewareArgs, SlackActionMiddlewareArgs } from "@slack/bolt";
import { BlockAction } from "@slack/bolt/dist/types/actions";
import { debugging101ModalView } from "../../lib/debugging101";
import { uploadFiles } from "../../lib/uploadFiles";

const debugging101 = async ({
  client,
  body,
  ack,
  context,
}: AllMiddlewareArgs & SlackActionMiddlewareArgs<BlockAction>) => {
  await ack();

  const files = await client.files.list({
    user: context.botUserId,
    types: "images",
  });

  const filesMap = new Map(files.files?.map((file) => [file.name, file]) ?? []);

  if (!filesMap.has("debugging-101.png")) {
    await uploadFiles(client, context.botUserId ?? "");
  }

  await client.views.open({
    trigger_id: body.trigger_id,
    view: debugging101ModalView(filesMap.get("debugging-101.png")?.id ?? ""),
  });
};

export { debugging101 };

import { AllMiddlewareArgs, SlackActionMiddlewareArgs } from "@slack/bolt";
import { BlockAction } from "@slack/bolt/dist/types/actions";

const blockActionMiddleware = async ({
  payload,
  client,
  context,
  body,
  next,
}: AllMiddlewareArgs & SlackActionMiddlewareArgs<BlockAction>) => {
  if (body.type !== "block_actions") {
    await next();
    return;
  }

  const now = new Date();
  try {
    const userInfo = await client.users.info({
      user: body.user.id,
    });
    context.user = userInfo.user;

    const channelInfo = await client.conversations.info({
      channel: body.channel?.id ?? "",
    });
    context.channel = channelInfo.channel;

    await next();
    context.logger.infoD("request-finished", {
      payload: payload,
      "response-time-ms": new Date().getTime() - now.getTime(),
      "channel-id": context.channel.id,
      "user-id": context.user.id,
      "status-code": 200,
    });
  } catch (error) {
    context.logger.infoD("request-finished", {
      payload: payload,
      "response-time-ms": new Date().getTime() - now.getTime(),
      "channel-id": context.channel && context.channel.id ? context.channel.id : "unknown",
      "user-id": context.user && context.user.id ? context.user.id : "unknown",
      "status-code": 500,
      error: error,
    });
  }
};

export { blockActionMiddleware };

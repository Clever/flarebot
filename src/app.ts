import { App } from "@slack/bolt";
import config from "./lib/config";
import kayvee from "kayvee";
import middleware from "./middleware";
import listeners from "./listeners";
import clients from "./clients";
import { UsersCache } from "./lib/usersCache";
import { ChannelsCache } from "./lib/channelsCache";
import { uploadFiles } from "./lib/uploadFiles";

const logger = new kayvee.logger("flarebot");
const usersCache = new UsersCache();
const channelsCache = new ChannelsCache();

const app = new App({
  token: config.SLACK_BOT_TOKEN,
  signingSecret: config.SLACK_SIGNING_SECRET,
  socketMode: true,
  appToken: config.SLACK_APP_TOKEN,
});

app.use(async ({ next, context }) => {
  context.clients = clients;
  context.logger = logger;
  context.usersCache = usersCache;
  // channelsCache is updated as new channels are seen or created by flarebot
  context.channelsCache = channelsCache;
  await next();
});

// This middleware is used to debug the incoming requests
// uncomment this only for debugging
// app.use(async ({ next, payload, body, context }) => {
//   console.log("payload", payload);
//   console.log("body", body);

//   console.log("users", context.usersCache.users.length);
//   console.log("channels", context.channelsCache.channels);

//   await next();
// });

middleware.register(app);
listeners.register(app);

(async () => {
  await app.start();
  app.logger.info("⚡️ Bolt app is running!");
})();

// background tasks
(async () => {
  while (true) {
    // files used in modal views expire every 90 days so lets just check and upload once a day if they are missing
    try {
      const self = await app.client.auth.test();
      // we don't cache file ids because the buttons are rarely clicked and its not worth the extra complexity for just 2 files
      await uploadFiles(app.client, self.user_id ?? "");
    } catch (error) {
      logger.errorD("upload-files-error", { error: error });
    }

    // We add new users to slack once a day and they probably don't interact with flares on their first day
    // so its okay to update cache just once every 24 hours.
    try {
      await usersCache.update(app.client);
    } catch (error) {
      logger.errorD("update-users-cache-error", { error: error });
    }

    await new Promise((resolve) => setTimeout(resolve, 24 * 60 * 60 * 1000));
  }
})();

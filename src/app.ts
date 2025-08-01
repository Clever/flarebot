import { App } from "@slack/bolt";
import config from "./lib/config";
import kayvee from "kayvee";
import middleware from "./middleware";
import listeners from "./listeners";
import clients from "./clients";
import { UsersCache } from "./lib/usersCache";
import { ChannelsCache } from "./lib/channelsCache";

const logger = new kayvee.logger("flarebot");
const usersCache = new UsersCache();
const channelsCache = new ChannelsCache();

console.log("starting flarebot...");
console.log("config", config);
console.log("SLACK_BOT_TOKEN", config.SLACK_BOT_TOKEN);
console.log("SLACK_SIGNING_SECRET", config.SLACK_SIGNING_SECRET);
console.log("SLACK_APP_TOKEN", config.SLACK_APP_TOKEN);

const app = new App({
  token: config.SLACK_BOT_TOKEN,
  signingSecret: config.SLACK_SIGNING_SECRET,
  socketMode: true,
  appToken: config.SLACK_APP_TOKEN,
});

app.use(async ({ next, context, client }) => {
  context.clients = clients;
  context.logger = logger;
  // We add new users to slack once a day and they probably don't interact with flares on their first day
  // so its okay to update cache just once every 24 hours.
  if (usersCache.users.length === 0 || Date.now() - usersCache.lastUpdated > 1000 * 60 * 60 * 24) {
    await usersCache.update(client);
  }
  context.usersCache = usersCache;
  // channelsCache is updated as new channels are seen or created by flarebot
  context.channelsCache = channelsCache;
  await next();
});

// This middleware is used to debug the incoming requests
// uncomment this only for debugging
app.use(async ({ next, payload, body, context }) => {
  console.log("payload", payload);
  console.log("body", body);

  console.log("users", context.usersCache.users.length);
  console.log("channels", context.channelsCache.channels);

  await next();
});

middleware.register(app);
listeners.register(app);

(async () => {
  await app.start();
  app.logger.info("⚡️ Bolt app is running!");
})();

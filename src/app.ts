import { App } from "@slack/bolt";
import config from "./lib/config";
import kayvee from "kayvee";
import middleware from "./middleware";
import listeners from "./listeners";
import clients from "./clients";

const logger = new kayvee.logger("flarebot");

const app = new App({
  token: config.SLACK_BOT_TOKEN,
  signingSecret: config.SLACK_SIGNING_SECRET,
  socketMode: true,
  appToken: config.SLACK_APP_TOKEN,
});

app.use(async ({ next, context }) => {
  context.clients = clients;
  context.logger = logger;
  await next();
});

// This middleware is used to debug the incoming requests
// uncomment this only for debugging
// app.use(async ({ next, payload, body, context }) => {
//   console.log("payload", payload);
//   console.log("body", body);

//   await next();
// });

middleware.register(app);
listeners.register(app);

(async () => {
  await app.start();
  app.logger.info("⚡️ Bolt app is running!");
})();

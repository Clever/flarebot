import { App, Middleware, AnyMiddlewareArgs } from "@slack/bolt";
import { middleware } from "./middleware/global";
import { fireAFlareRegex, fireFlare } from "./lib/fireFlare";
import { helpRegex, help } from "./lib/help";
import config from "./lib/config";

const app = new App({
  token: config.SLACK_BOT_TOKEN,
  signingSecret: config.SLACK_SIGNING_SECRET,
  socketMode: true,
  appToken: config.SLACK_APP_TOKEN,
});

app.use(middleware as Middleware<AnyMiddlewareArgs>);

app.message(fireAFlareRegex, fireFlare);
app.message(helpRegex, help);

export default app;

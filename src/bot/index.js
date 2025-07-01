const { App } = require("@slack/bolt");
const { middleware } = require("./middleware/global");
const { fireAFlareRegex, fireFlare } = require("./lib/fireFlare");
const { helpRegex, help } = require("./lib/help");

const app = new App({
  token: process.env.SLACK_BOT_TOKEN,
  signingSecret: process.env.SLACK_SIGNING_SECRET,
  socketMode: true,
  appToken: process.env.SLACK_APP_TOKEN,
});

app.use(middleware);

app.message(fireAFlareRegex, fireFlare);
app.message(helpRegex, help);

module.exports = app;

const app = require("./bot");

(async () => {
  await app.start();
  app.logger.info("⚡️ Bolt app is running!");
})();

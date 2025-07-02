import app from "./bot";

(async () => {
  await app.start();
  app.logger.info("⚡️ Bolt app is running!");
})();

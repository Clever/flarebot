import type { AnyMiddlewareArgs, App, Middleware } from "@slack/bolt";
import { messageMiddleware } from "./message";

const register = (app: App) => {
  app.use(messageMiddleware as Middleware<AnyMiddlewareArgs>);
};

export default { register };

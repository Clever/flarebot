import type { AnyMiddlewareArgs, App, Middleware } from "@slack/bolt";
import { messageMiddleware } from "./message";
import { blockActionMiddleware } from "./blockAction";

const register = (app: App) => {
  app.use(messageMiddleware as Middleware<AnyMiddlewareArgs>);
  app.use(blockActionMiddleware as Middleware<AnyMiddlewareArgs>);
};

export default { register };

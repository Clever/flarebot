// eslint-disable-next-line @typescript-eslint/no-explicit-any
type TodoType = any;

declare module "kayvee" {
  export const logger: TodoType;
  export type logger = TodoType;
  export const Logger: TodoType;
  export type Logger = TodoType;
  export const middleware: TodoType;
  export const setGlobalRouting: TodoType;
}

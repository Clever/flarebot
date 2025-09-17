const config = {
  SLACK_BOT_TOKEN: getEnvVar("SLACK_BOT_TOKEN"),
  SLACK_SIGNING_SECRET: getEnvVar("SLACK_SIGNING_SECRET"),
  SLACK_APP_TOKEN: getEnvVar("SLACK_APP_TOKEN"),

  FLARES_CHANNEL_ID: getEnvVar("FLARES_CHANNEL_ID"),
  FLARES_CHANNEL_NAME: getEnvVar("FLARES_CHANNEL_NAME"),
  FLARE_CHANNEL_PREFIX: getEnvVar("FLARE_CHANNEL_PREFIX"),

  JIRA_ORIGIN: getEnvVar("JIRA_ORIGIN"),
  JIRA_USERNAME: getEnvVar("JIRA_USERNAME"),
  JIRA_PASSWORD: getEnvVar("JIRA_PASSWORD"),
  JIRA_PROJECT_ID: getEnvVar("JIRA_PROJECT_ID"),

  GOOGLE_FLAREBOT_SERVICE_ACCOUNT_CONF: getEnvVar("GOOGLE_FLAREBOT_SERVICE_ACCOUNT_CONF"),
  GOOGLE_DOMAIN: getEnvVar("GOOGLE_DOMAIN"),
  GOOGLE_TEMPLATE_DOC_ID: getEnvVar("GOOGLE_TEMPLATE_DOC_ID"),
  GOOGLE_SLACK_HISTORY_DOC_ID: getEnvVar("GOOGLE_SLACK_HISTORY_DOC_ID"),

  USERS_TO_INVITE: getEnvVar("USERS_TO_INVITE"),
  PAGERDUTY_API_KEY: getEnvVar("PAGERDUTY_API_KEY"),

  GOOGLE_FLARE_FOLDER_ID: getEnvVar("GOOGLE_FLARE_FOLDER_ID"),
  GOOGLE_SHARED_DRIVE_ID: getEnvVar("GOOGLE_SHARED_DRIVE_ID"),
};

export default config;

function getEnvVar(name: string) {
  const value = process.env[name];
  if (!value) {
    if (process.env.NODE_ENV === "test") {
      return "";
    }
    throw new Error(`${name} must be set in environment variables.`);
  }
  return value;
}

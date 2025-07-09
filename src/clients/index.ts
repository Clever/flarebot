import config from "../lib/config";
import { drive } from "@googleapis/drive";
import { GoogleAuth } from "google-auth-library";
import { Version3Client } from "jira.js";

const jiraClient = new Version3Client({
  host: config.JIRA_ORIGIN,
  authentication: {
    basic: {
      email: config.JIRA_USERNAME,
      apiToken: config.JIRA_PASSWORD,
    },
  },
});

const googleAuth = new GoogleAuth({
  credentials: JSON.parse(config.GOOGLE_FLAREBOT_SERVICE_ACCOUNT_CONF),
  scopes: ["https://www.googleapis.com/auth/drive"],
});

const googleDriveClient = drive({
  version: "v3",
  auth: googleAuth,
});

export default class clients {
  static jiraClient = jiraClient;
  static googleDriveClient = googleDriveClient;
}

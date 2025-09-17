import { drive_v3 } from "@googleapis/drive";
import config from "./config";

async function createFlareDoc(
  googleDriveClient: drive_v3.Drive,
  issueKey: string,
  title: string,
  specialType: string,
): Promise<{ docId: string; historyDocId: string }> {
  let docId = "";
  let historyDocId = "";
  let flaredocTitle = `${issueKey}: ${title}`;
  if (specialType) {
    flaredocTitle = `${issueKey}: ${title} (${specialType})`;
  }

  let slackHistoryDocTitle = `${issueKey}: ${title} (Slack History)`;
  if (specialType) {
    slackHistoryDocTitle = `${issueKey}: ${title} (${specialType}) (Slack History)`;
  }

  const year = new Date().getFullYear();

  const folders = await googleDriveClient.files.list({
    driveId: config.GOOGLE_SHARED_DRIVE_ID,
    supportsAllDrives: true,
    includeItemsFromAllDrives: true,
    corpora: "drive",
    q: `'${config.GOOGLE_FLARE_FOLDER_ID}' in parents and trashed=false`,
  });

  const matchingYearFolder = folders.data.files?.find((file) => file.name === year.toString());
  let yearFolderID = matchingYearFolder?.id ?? "";

  if (!matchingYearFolder) {
    const yearFolder = await googleDriveClient.files.create({
      requestBody: {
        name: year.toString(),
        parents: [config.GOOGLE_FLARE_FOLDER_ID],
        mimeType: "application/vnd.google-apps.folder",
      },
      supportsAllDrives: true,
    });
    yearFolderID = yearFolder.data.id ?? "";
  }

  const flareFolder = await googleDriveClient.files.create({
    requestBody: {
      name: issueKey,
      parents: [yearFolderID],
      mimeType: "application/vnd.google-apps.folder",
    },
    supportsAllDrives: true,
  });

  const flareFolderID = flareFolder.data.id ?? "";
  if (!flareFolderID) {
    throw new Error("Failed to create flare folder.");
  }

  const flaredoc = await googleDriveClient.files.copy({
    requestBody: {
      name: flaredocTitle,
      parents: [flareFolderID],
    },
    fileId: config.GOOGLE_TEMPLATE_DOC_ID,
    supportsAllDrives: true,
  });
  docId = flaredoc.data.id ?? "";

  const slackHistoryDoc = await googleDriveClient.files.copy({
    requestBody: {
      name: slackHistoryDocTitle,
      parents: [flareFolderID],
    },
    fileId: config.GOOGLE_SLACK_HISTORY_DOC_ID,
    supportsAllDrives: true,
  });
  historyDocId = slackHistoryDoc.data.id ?? "";

  const flaredocHTML = await googleDriveClient.files.export({
    fileId: docId,
    mimeType: "text/html",
  });

  let html = flaredocHTML.data as string;

  html = html.replace("[FLARE-KEY]", issueKey);
  html = html.replace(
    "[START-DATE]",
    new Date().toLocaleString("en-US", { timeZone: "US/Pacific" }) + " PT",
  );
  html = html.replace("[SUMMARY]", title);
  html = html.replace(
    "[HISTORY-DOC]",
    `<a href="https://docs.google.com/spreadsheets/d/${historyDocId}">${slackHistoryDocTitle}</a>`,
  );

  await googleDriveClient.files.update({
    fileId: docId,
    media: {
      mimeType: "text/html",
      body: html,
    },
    supportsAllDrives: true,
  });

  return { docId, historyDocId };
}

export { createFlareDoc };

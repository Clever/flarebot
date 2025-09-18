import { GoogleAuth } from "google-auth-library";
import { drive } from "@googleapis/drive";
import { Version3Client } from "jira.js";


const CONF = process.env["CONF"] ?? "";
const JIRA_ORIGIN = "https://clever.atlassian.net"
const JIRA_USERNAME = "flarebot@clever.com"
const JIRA_PASSWORD = process.env["JIRA_PASSWORD"] ?? "";

const sharedDriveId = "0AEIK1eR4cewyUk9PVA"

// hardcode to reduce google api calls
const yearFolderMap = {
  "FLARETEST": {
    "2025": "1_FZdvcENxcqDH5y_3FQeFPcZAeaf9zZW",
  },
  "FLARE": {
    "2025": "15hLrR7MPVudWmytJKz_ctSngdw6c_lSJ",
    "2024": "17VVfQH1JBPAWjXBuN-mrn2wL8XiJU6QA",
    "2023": "13TaKCgfxkOtERSFE0GPFoQYq0wGFIhWR",
    "2022": "1XratNuCR8CJdBF4Zp9l4kSS0Esz9kecp",
    "2021": "1kB7ozhQSITVXmjzbb_8qf2Dipyt02AKK",
    "2020": "1kfwuqdAJTam1oHZQmITXkqnP47Oz4m4l",
    "2019": "1l4Dwy8MliyDRjiqw7SUhXWZk63_Hx-mB",
  },
}

const flareFolderMap = {}

const googleAuth = new GoogleAuth({
  credentials: JSON.parse(CONF),
  scopes: ["https://www.googleapis.com/auth/drive"],
});

const googleDriveClient = drive({
  version: "v3",
  auth: googleAuth,
});

const jiraClient = new Version3Client({
  host: JIRA_ORIGIN,
  authentication: {
    basic: {
      email: JIRA_USERNAME,
      apiToken: JIRA_PASSWORD,
    },
  },
});

async function getAllFiles(query) {
  const files = [];
  let pageToken = null;
  while (true) {
    query.pageToken = pageToken;
    const f = await googleDriveClient.files.list(query); 

    f.data.files?.forEach((file) => {
      files.push(file);
    });

    if (f.data.nextPageToken) {
      pageToken = f.data.nextPageToken;
    } else {
      break;
    }
  }

  return files;
}

async function getAllInFolder(folderID) {
  return getAllFiles({
    driveId: sharedDriveId,
    supportsAllDrives: true,
    includeItemsFromAllDrives: true,
    corpora: "drive",
    q: `'${folderID}' in parents and trashed=false`,
  });
}

const currentFolders = {
  "FLARETEST": {
    "2025": await getAllInFolder(yearFolderMap["FLARETEST"]["2025"])
  },
  "FLARE": {
    "2025": await getAllInFolder(yearFolderMap["FLARE"]["2025"]),
    "2024": await getAllInFolder(yearFolderMap["FLARE"]["2024"]),
    "2023": await getAllInFolder(yearFolderMap["FLARE"]["2023"]),
    "2022": await getAllInFolder(yearFolderMap["FLARE"]["2022"]),
    "2021": await getAllInFolder(yearFolderMap["FLARE"]["2021"]),
    "2020": await getAllInFolder(yearFolderMap["FLARE"]["2020"]),
    "2019": await getAllInFolder(yearFolderMap["FLARE"]["2019"]),
  }
}

const files = await getAllFiles({
  q: `trashed=false`,
});

const filteredFiles = files.filter((file) => file.name.startsWith("FLARETEST-") || file.name.startsWith("FLARE-"));

const sortedFiles = filteredFiles.sort((a, b) => a.name.localeCompare(b.name))

// Process files sequentially to avoid race conditions
for (const file of sortedFiles) {  
  const issueKey = file.name.split(":")[0];
  const issueType = issueKey.split("-")[0];

  let issueDate;
  try {
    const issue = await jiraClient.issues.getIssue({issueIdOrKey: issueKey});
    issueDate = new Date(issue.fields.created);
  } catch (error) {
    console.log(`Issue ${issueKey} not found: ${error.message}`);
    continue;
  }
  
  const year = issueDate.getFullYear();
  const yearFolderId = yearFolderMap[issueType][year];

  if (file.name.startsWith("FLARETEST-") && year !== 2025) {
    continue;
  }

  // Check if folder already exists in our cache
  let flareFolderId = flareFolderMap[issueKey];
  if (!flareFolderId) {
    // Check if folder already exists in the year folder
    const folders = currentFolders[issueType][year];
    flareFolderId = folders.find((folder) => folder.name === issueKey)?.id;
    
    if (!flareFolderId) {
      try {
        const flareFolder = await googleDriveClient.files.create({
          requestBody: {
            name: issueKey,
            parents: [yearFolderId],
            mimeType: "application/vnd.google-apps.folder",
          },
          supportsAllDrives: true,
        });
        flareFolderId = flareFolder.data.id;
      } catch (error) {
        console.error(`Failed to create folder for ${issueKey}: ${error.message}`);
        continue;
      }
    }
    // Cache the folder ID to prevent duplicate creation
    flareFolderMap[issueKey] = flareFolderId;
  }

  try {
    await googleDriveClient.files.update({
      fileId: file.id,
      supportsAllDrives: true,
      addParents: flareFolderId,
    });
    console.log(`Moved ${file.name} to ${issueType} folder in year ${year}`);
  } catch (error) {
    console.error(`Failed to move ${file.name}: ${error.message}`);
  }
}
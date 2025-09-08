import { WebClient } from "@slack/web-api";
import fs from "fs";

async function uploadFiles(client: WebClient, botId: string) {
  // get all files in the images directory
  const files = fs.readdirSync("src/lib/images");

  const userfiles = await client.files.list({
    user: botId,
    types: "images",
  });

  const userfilesMap = new Map(userfiles.files?.map((file) => [file.name, file]) ?? []);

  for (const file of files) {
    if (file === "README.md") {
      continue;
    }

    if (userfilesMap.has(file)) {
      continue;
    }

    const response = await client.filesUploadV2({
      file: `src/lib/images/${file}`,
      filename: file,
      title: file,
    });

    if (!response.ok) {
      throw new Error(`Failed to upload file ${file}: ${response.error}`);
    }
  }
}

export { uploadFiles };

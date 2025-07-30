import { GenericMessageEvent, KnownBlock, WebClient } from "@slack/web-api";
import { Channel } from "@slack/web-api/dist/types/response/ConversationsInfoResponse";
import config from "./config";

class ChannelsCache {
  channels: Record<string, { info: Channel; historyDocId?: string }>;

  constructor() {
    this.channels = {};
  }

  async getChannel(client: WebClient, channelId: string): Promise<Channel> {
    let channel = this.channels[channelId];
    if (!channel) {
      const response = await client.conversations.info({
        channel: channelId,
      });
      if (!response.channel) {
        throw new Error(`Channel ${channelId} not found`);
      }
      channel = {
        info: response.channel,
      };
      this.channels[channelId] = channel;
    }
    return channel.info;
  }

  async getChannelHistoryDocId(
    client: WebClient,
    channelId: string,
    botUserId: string,
  ): Promise<string | undefined> {
    let channel = this.channels[channelId];
    if (!channel || !channel.historyDocId) {
      const info = await this.getChannel(client, channelId);
      // only flare channels have a flare doc id
      if (!info.name?.startsWith(config.FLARE_CHANNEL_PREFIX)) {
        return undefined;
      }
      const pinedMessages = await client.pins.list({
        channel: channelId,
      });
      // this can be simplified after https://github.com/slackapi/node-slack-sdk/issues/2316 is resolved
      // currently the pins.list response type doesn't include the message
      for (const pin of pinedMessages.items ?? []) {
        if (pin.created_by === botUserId && pin.type === "message") {
          const message = "message" in pin ? (pin.message as GenericMessageEvent) : undefined;
          if (message && message.text?.startsWith("Thank you for firing a flare")) {
            const lastBlock = message.blocks?.[message.blocks.length - 1] as KnownBlock;
            if (lastBlock && lastBlock.type === "section") {
              const lastBlockText = lastBlock.text?.text;
              if (lastBlockText) {
                const docIdMatch = lastBlockText.match(
                  /docs\.google\.com\/spreadsheets\/d\/([a-zA-Z0-9_-]+)/,
                );
                if (docIdMatch) {
                  channel = {
                    info: info,
                    historyDocId: docIdMatch[1],
                  };
                  this.channels[channelId] = channel;
                }
              }
            }
          }
        }
      }
    }
    return channel.historyDocId;
  }

  setChannel(channelId: string, channel: Channel, historyDocId?: string) {
    this.channels[channelId] = {
      info: channel,
      historyDocId: historyDocId,
    };
  }
}

export { ChannelsCache };

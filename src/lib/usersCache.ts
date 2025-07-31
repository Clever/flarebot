import { WebClient } from "@slack/web-api";
import { Member } from "@slack/web-api/dist/types/response/UsersListResponse";

class UsersCache {
  users: Member[];
  lastUpdated: number;

  constructor() {
    this.users = [];
    this.lastUpdated = 0;
  }

  async update(client: WebClient) {
    const allUsers: Member[] = [];
    let cursor = undefined;
    while (true) {
      const response = await client.users.list({
        cursor: cursor,
      });
      if (response.members) {
        allUsers.push(...response.members);
      }
      cursor = response.response_metadata?.next_cursor;
      if (!cursor) {
        break;
      }
    }
    this.users = allUsers;
    this.lastUpdated = Date.now();
  }

  async getUser(client: WebClient, userId: string): Promise<Member | undefined> {
    if (userId === "") {
      return undefined;
    }
    let user = this.users.find((user) => user.id === userId);
    if (!user) {
      const userInfo = await client.users.info({
        user: userId,
      });
      if (userInfo.user) {
        user = userInfo.user;
        this.users.push(user);
      }
    }
    return user;
  }
}

export { UsersCache };

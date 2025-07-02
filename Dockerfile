FROM node:24-slim

WORKDIR /flarebot
# for now we are copying only the relevant files but in the future we could simplify this
# by copying the entire repo when old flarebot is removed
COPY src /flarebot/src
COPY tsconfig.json /flarebot
COPY package.json /flarebot
COPY package-lock.json /flarebot

RUN npm ci

RUN npx tsc -p .

CMD ["node", "dist/app.js"]

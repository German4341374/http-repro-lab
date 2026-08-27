# syntax=docker/dockerfile:1.7
FROM node:24.14.1-alpine3.23 AS dependencies
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --ignore-scripts

FROM dependencies AS build
COPY tsconfig.json tsconfig.build.json ./
COPY src ./src
RUN npm run build && npm prune --omit=dev

FROM node:24.14.1-alpine3.23 AS runtime
ENV NODE_ENV=production HTTP_REPRO_BIND=0.0.0.0 HTTP_REPRO_PORT=8080
WORKDIR /app
COPY --from=build --chown=node:node /app/node_modules ./node_modules
COPY --from=build --chown=node:node /app/dist ./dist
COPY --chown=node:node package.json ./
USER node
EXPOSE 8080
CMD ["node", "dist/server.js"]


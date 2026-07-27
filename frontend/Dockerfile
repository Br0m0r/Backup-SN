FROM node:20.20.2-alpine3.22 AS builder

WORKDIR /app

# Copy package files
COPY frontend/package.json frontend/package-lock.json* ./

RUN npm ci

# Copy frontend source code
COPY frontend/ ./

# Build the production bundle
RUN npm run build

FROM node:20.20.2-alpine3.22

WORKDIR /app
COPY --from=builder --chown=node:node /app/dist ./dist
COPY --chown=node:node frontend/server.mjs ./server.mjs

USER node
EXPOSE 3000
STOPSIGNAL SIGTERM

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -q -O /dev/null http://127.0.0.1:3000/ || exit 1

ENTRYPOINT ["node", "server.mjs"]

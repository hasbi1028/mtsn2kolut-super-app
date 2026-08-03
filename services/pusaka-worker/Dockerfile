# PUSAKA Worker — image mandiri untuk deploy ke PaaS (DomCloud, dll).
#
# Cara pakai di DomCloud:
#   1. Push folder ini sebagai repo (atau git subtree split dari monorepo).
#   2. Di DomCloud pilih deploy dengan Dockerfile.
#   3. Set env vars (lihat DEPLOY_DOMCLOUD.md).
#   4. Health check: GET /health (port 8091 default, atau PORT dari platform).

FROM node:22-slim

WORKDIR /app

# Install dependencies + browser Chromium untuk Playwright (termasuk deps OS)
COPY package.json package-lock.json ./
RUN npm ci && npx playwright install --with-deps chromium

# Build TypeScript
COPY tsconfig.json tsconfig.build.json ./
COPY src ./src
RUN npm run build

ENV NODE_ENV=production
# Port HTTP health/metrics — DomCloud biasanya meng-inject PORT sendiri;
# kalau tidak, default 8091.
ENV PORT=8091
EXPOSE 8091

CMD ["node", "dist/index.js"]

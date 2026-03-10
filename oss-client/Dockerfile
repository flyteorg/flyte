# syntax=docker/dockerfile:1

# Base image setup
FROM node:23-alpine AS base
RUN apk add --no-cache libc6-compat rsync git

LABEL org.opencontainers.image.source=https://github.com/unionai/consolev2

ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production
ENV PORT=8080

# -------------------------------
# Build stage
FROM base AS deps
ARG BUILDKITE_COMMIT

# Enable and pin pnpm
RUN corepack enable && corepack prepare pnpm@10.20.0 --activate

WORKDIR /app

# Copy only files needed to install dependencies first (better cache)
COPY package.json pnpm-lock.yaml ./

# Install dependencies (cached if lockfile unchanged)
RUN \
  --mount=type=cache,target=/root/.local/share/pnpm/store \
  pnpm install --frozen-lockfile

# Now copy the full app code
COPY . .

RUN apk add --no-cache aws-cli curl jq

# Inject commit SHA
RUN GIT_SHA=${BUILDKITE_COMMIT} node ./scripts/prodsha.mjs

# Build: fetch faro sourcemaps token + run Next build
RUN export FARO_SOURCEMAP_TOKEN="$(./scripts/get-grafana-token.sh)" \
  && GIT_SHA=${BUILDKITE_COMMIT} pnpm build:prod

# Prune dev dependencies to keep image lean
RUN pnpm prune --prod

# -------------------------------
# Runtime stage
FROM base AS runner
ARG BUILDKITE_COMMIT

ENV GIT_SHA=${BUILDKITE_COMMIT}
ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production
ENV PORT=8080

WORKDIR /app

# Create non-root user
RUN addgroup -g 1001 -S nodejs && \
  adduser -S nextjs -u 1001

# Copy standalone server and dependencies
COPY --from=deps --chown=nextjs:nodejs /app/.next/standalone ./
# Copy static files for Next.js to serve
COPY --from=deps --chown=nextjs:nodejs /app/.next/static ./.next/static
# Copy public directory if it exists
COPY --from=deps --chown=nextjs:nodejs /app/public ./public

USER nextjs

EXPOSE 8080 8081

# Run the standalone Next.js server (no pnpm required)
CMD ["node", "server.js"]
